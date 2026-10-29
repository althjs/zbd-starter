package core

import (
	"context"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"zbd.kr/be/internal/lib/auth"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// reqHeaders — 요청 헤더(프록시 헤더가 없으면 접속 주소를 x-forwarded-for 로 — 지금 서비스 웹서버가 하던 것)
func reqHeaders(c *web.Ctx) http.Header {
	h := c.R.Header.Clone()
	if h.Get("x-forwarded-for") == "" {
		if host, _, err := net.SplitHostPort(c.R.RemoteAddr); err == nil {
			h.Set("x-forwarded-for", host)
		}
	}
	return h
}

// ipOf — getIpAddressFromHeader: x-forwarded-for 첫 값 → x-real-ip
func ipOf(h http.Header) string {
	if x := h.Get("x-forwarded-for"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	return h.Get("x-real-ip")
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// jsParseInt — parseInt(s, 10) (실패면 NaN)
func jsParseInt(s string) float64 {
	s = strings.TrimSpace(s)
	end := 0
	for end < len(s) && ((s[end] >= '0' && s[end] <= '9') || (end == 0 && (s[end] == '-' || s[end] == '+'))) {
		end++
	}
	if end == 0 || (end == 1 && (s[0] == '-' || s[0] == '+')) {
		return math.NaN()
	}
	return jsx.Number(s[:end])
}

// orNum — (v || d) 숫자
func orNum(v, d float64) float64 {
	if v == 0 || math.IsNaN(v) {
		return d
	}
	return v
}

// localMs — new Date("YYYY-MM-DD HH:MM:SS").getTime() (한국 시각). 잘못되면 NaN.
func localMs(s string) float64 {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, util.KST)
	if err != nil {
		return math.NaN()
	}
	return float64(t.UnixMilli())
}

// adminLoginDevices — 로그인 기기 목록(관리자 A / 회원 U 탭, 기간·쪽수)
func adminLoginDevices(c *web.Ctx) {
	ctx := c.Context()
	qs := func(k, d string) string {
		if v := c.Q(k); v != "" {
			return v
		}
		return d
	}
	pageNo := orNum(jsParseInt(qs("pageNo", "1")), 1)
	pageSize := orNum(jsParseInt(qs("pageSize", "20")), 20)
	orderBy := qs("orderBy", "lastLoginedAt")
	switch orderBy {
	case "lastLoginedAt", "lastActivedAt", "createdAt", "email", "memberType":
	default:
		orderBy = "lastLoginedAt"
	}
	desc := "DESC"
	if qs("orderByDesc", "desc") == "asc" {
		desc = "ASC"
	}
	tab := qs("tab", "U")
	from := localMs(qs("from", "2027-01-01") + " 00:00:00")
	to := localMs(qs("to", util.Today()) + " 23:59:59")
	hostingID := c.Token.HostingID

	where := "WHERE 1"
	var args []any
	if hostingID != "" {
		where += " AND d.hostingId = ?"
		args = append(args, hostingID)
	}
	if tab != "" {
		where += " AND d.memberType = ?"
		args = append(args, tab)
	}
	if !math.IsNaN(from) && from != 0 {
		where += " AND UNIX_TIMESTAMP(d.lastLoginedAt)*1000 >= ?"
		args = append(args, from)
	}
	if !math.IsNaN(to) && to != 0 {
		where += " AND UNIX_TIMESTAMP(d.lastLoginedAt)*1000 <= ?"
		args = append(args, to)
	}
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	cnt, err := query(ctx, conn, `SELECT COUNT(*) as cnt FROM login_devices d `+where, args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	total := 0.0
	if len(cnt) > 0 {
		total = jsx.NumOr(cnt[0].V("cnt"), 0)
	}
	sqlText := `SELECT d.userId, d.memberType, d.browserUid, d.ip, d.hostingId
    , UNIX_TIMESTAMP(d.createdAt)*1000 as createdAt
    , UNIX_TIMESTAMP(d.lastLoginedAt)*1000 as lastLoginedAt
    , UNIX_TIMESTAMP(d.lastActivedAt)*1000 as lastActivedAt
    , u.email as uEmail`
	if tab == "U" {
		sqlText += `, u.nickname, u.name as name, u.groupId`
	}
	sqlText += ` FROM login_devices d LEFT JOIN members u ON d.memberType = 'U' AND d.userId = u.userId AND d.hostingId = u.hostingId ` + where +
		` ORDER BY ` + orderBy + ` ` + desc + ` LIMIT ?, ?`
	rows, err := query(ctx, conn, sqlText, append(append([]any{}, args...), (pageNo-1)*pageSize, pageSize)...)
	if err != nil {
		c.Ecb(err)
		return
	}
	items := []any{}
	for _, o := range rows {
		if jsx.Str(o.V("memberType")) == "A" {
			m := c.App.Admin.Member(jsx.Str(o.V("userId")))
			var email, name any = "", "-"
			if m != nil {
				if jsx.Truthy(m.V("email")) {
					email = m.V("email")
				}
				if jsx.Truthy(m.V("name")) {
					name = m.V("name")
				}
			}
			o.Set("email", email)
			if tab == "A" {
				o.Set("name", name)
			}
		} else {
			o.Set("email", c.App.PII.Dec(o.V("uEmail")))
			if o.Has("name") {
				o.Set("name", c.App.PII.Dec(o.V("name")))
			}
		}
		o.Delete("uEmail")
		items = append(items, o)
	}
	u, err := queryOne(ctx, conn, `SELECT count(distinct(d.userId)) as cnt FROM login_devices d `+where, args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	uniq := 0.0
	if u != nil {
		uniq = jsx.NumOr(u.V("cnt"), 0)
	}
	c.Scb(jsx.Obj("totalCount", total, "uniqUserIdCount", uniq, "items", items))
}

var hostRe = regexp.MustCompile(`^https?://([^/]+)`)

// logTokenIssue — 브라우저 확인을 건너뛰는 토큰(ANY_BROWSER) 발급 기록. 실패해도 막지 않는다.
func logTokenIssue(c *web.Ctx, kind, result, reason string, issuer, target *jsx.OrderedMap) {
	h := reqHeaders(c)
	ip := clip(ipOf(h), 50)
	cands := clip(strings.Join(util.ClientIPs(h), ","), 300)
	office := 0
	if util.IsOffice(h, util.OfficeIPs(c.App.Cfg.Get("OFFICE_IPS"))) {
		office = 1
	}
	src := c.Header("origin")
	if src == "" {
		src = c.Header("referer")
	}
	host := src
	if m := hostRe.FindStringSubmatch(src); m != nil {
		host = m[1]
	}
	s := func(m *jsx.OrderedMap, k string, n int) string { return clip(jsx.StrOrEmpty(m.V(k)), n) }
	num := func(m *jsx.OrderedMap, k string) float64 { return jsx.NumOr(m.V(k), 0) }
	_, err := db.Exec(context.Background(), c.App.DB, `INSERT INTO admin_token_logs
        (kind, result, reason, issuerAdminId, issuerName, issuerHostingId,
         targetAdminId, targetName, targetHostingId, ip, ipCandidates, isOffice, host, ua)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		kind, result, clip(reason, 200), num(issuer, "adminId"), s(issuer, "name", 100), s(issuer, "hostingId", 50),
		num(target, "adminId"), s(target, "name", 100), s(target, "hostingId", 50), ip, cands, office, clip(host, 200), clip(c.Header("user-agent"), 500))
	if err != nil {
		log.Printf("[token-log] 기록 실패 %v", err)
	}
}

// adminMemberToken — 다른 관리자로 접속하는 토큰(ANY_BROWSER). 본사 마스터이거나 같은 회사 마스터, 또는 사이트 마스터만.
func adminMemberToken(c *web.Ctx) {
	ctx := c.Context()
	adminID := c.Q("adminId")
	if adminID == "" {
		c.Ecb("잘못된 요청")
		return
	}
	t := c.Token
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	admin, err := adminGet(c, conn, adminID, false, false)
	if err != nil {
		c.Ecb(err)
		return
	}
	if admin == nil {
		c.Ecb("Cannot read properties of null (reading 'groupId')")
		return
	}
	site := c.App.Settings.Site(t.HostingID)
	isSiteMaster := site != nil && jsx.Truthy(site.V("masterAdminId")) && jsx.Number(site.V("masterAdminId")) == jsx.Number(t.Key)
	me, err := queryOne(ctx, conn, `SELECT groupId FROM admins WHERE adminId = ?`, t.Key)
	if err != nil {
		c.Ecb(err)
		return
	}
	if me == nil {
		c.Ecb("Cannot read properties of null (reading 'groupId')")
		return
	}
	isHqMaster := t.HostingID == ""
	masterAllowed := jsx.Number(me.V("groupId")) == 1 && (isHqMaster || jsx.StrOrEmpty(admin.V("hostingId")) == t.HostingID)
	issuer := jsx.Obj("adminId", t.Key, "name", t.Name, "hostingId", t.HostingID)
	target := jsx.Obj("adminId", adminID, "name", admin.V("name"), "hostingId", admin.V("hostingId"))
	if !isSiteMaster && !masterAllowed {
		logTokenIssue(c, "DEV_LOGIN", "DENIED", "권한 없음", issuer, target)
		c.Ecb("잘못된 접근")
		return
	}
	o, err := queryOne(ctx, conn, `SELECT name, groupId, email, hostingId FROM admins where adminId = ?`, adminID)
	if err != nil || o == nil {
		c.Ecb(err)
		return
	}
	cid := jsx.StrOrEmpty(o.V("hostingId"))
	hosting := c.App.Settings.Hosting(cid)
	siteAdmin := hosting.V("siteAdminId")
	if !jsx.Truthy(siteAdmin) {
		siteAdmin = hosting.Path("server", "associationAdminId")
	}
	sub := hosting.V("subAdminId")
	ns := func(v any) string {
		if !jsx.Truthy(v) {
			return "0"
		}
		return jsx.Str(v)
	}
	log.Printf("@ 관리자 토큰 강제접속: %s %s %s => %s %v 으로 접속함", t.Name, t.HostingID, t.Key, cid, o.V("name"))
	logTokenIssue(c, "DEV_LOGIN", "OK", "", issuer, jsx.Obj("adminId", adminID, "name", o.V("name"), "hostingId", cid))
	tok := makeToken(c, adminID, jsx.StrOrEmpty(o.V("email")), "A", auth.AnyBrowser, cid, jsx.StrOrEmpty(o.V("name")), jsx.StrOrEmpty(o.V("groupId")), "0", ns(siteAdmin), "", 0, ns(sub))
	data := jsx.Obj("token", tok, "uid", t.BrowserUID, "name", o.V("name"))
	if hosting != nil {
		data.Set("config", hosting)
	}
	c.Scb(data)
}

// dbq — 읽기·쓰기 연결(여러 기능이 같이 쓴다)

type dbq interface {
	db.Querier
	db.Execer
}
