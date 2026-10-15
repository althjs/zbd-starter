package collab

// 브라우저에 묶이지 않는 토큰(ANY_BROWSER) 발급 기록 집계 — GET /admin/system-monitor/token-log?days=7|30|90
// 서버 전체 기록이라 개발자·최고 관리자만(canViewSystemSecurity).

import (
	"regexp"
	"sort"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/auth"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// canViewSystemSecurity — 시스템 보안 기록(공격 통계·차단·토큰 발급 기록)을 볼 수 있는가 — 판정 systemSecurityAllowed
func canViewSystemSecurity(c *web.Ctx, t *auth.Token) bool {
	if !t.Valid() {
		return false
	}
	return systemSecurityAllowed(c.App.Admin.Member(t.Key), jsx.Number(t.GroupID))
}

// systemMenuAuth — 서버 관리 화면(배포 이력·E2E·공격 감지·CPU 등)의 권한 "" | "R" | "RW"(2026-10-09 대표 — 「메뉴 권한 있는 사람은 보고, RW 면 실행까지」).
// 개발자·최고 관리자·1그룹은 늘 RW, 나머지는 그 화면 메뉴(path)에 준 권한.
func systemMenuAuth(c *web.Ctx, t *auth.Token, path string) string {
	if !t.Valid() {
		return ""
	}
	if canViewSystemSecurity(c, t) {
		return "RW"
	}
	return c.App.Admin.MenuAuth(t.HostingID, t.Key, path)
}

// adminSystemMenuAuth — GET ?path=/admin/system-monitor/… → {auth: "" | "R" | "RW"}. 서버 관리 화면 주소만 받는다
func adminSystemMenuAuth(c *web.Ctx) {
	path := jsx.StrOrEmpty(c.QV("path"))
	if c.Method != "GET" || !strings.HasPrefix(path, "/admin/system-monitor/") {
		c.Ecb("잘못된 요청")
		return
	}
	c.Scb(jsx.Obj("auth", systemMenuAuth(c, c.Token, path)))
}

// systemSecurityAllowed — 개발자 · 최고 관리자(isSuperAdmin) · 1그룹(최고 관리자 그룹). 테스트 token_log_policy_test.go(운영은 nginxattack_policy_test.go)
// (2026-10-05 — 예전엔 「1그룹 + 본사(회사 없음)」 였는데 스타터 사이트 주인(1그룹·자기 호스팅)이 못 봤다. 운영 즈비드에서 늘어난 사람은 없다)
func systemSecurityAllowed(m *jsx.OrderedMap, groupID float64) bool {
	if m != nil && (m.V("isDeveloper") == true || m.V("isSuperAdmin") == true) {
		return true
	}
	return groupID == 1
}

var reNoTable = regexp.MustCompile(`(?i)doesn't exist|ER_NO_SUCH_TABLE`)

func adminTokenLog(c *web.Ctx) {
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	if !canViewSystemSecurity(c, c.Token) {
		c.Ecb("개발자·최고 관리자만 볼 수 있습니다.")
		return
	}
	days := 7.0
	if d := jsx.Number(c.QV("days")); d == 7 || d == 30 || d == 90 {
		days = d
	}
	rows, err := db.Query(c.Context(), c.App.DB, `SELECT logId, kind, result, reason, issuerAdminId, issuerName, issuerHostingId,
              targetAdminId, targetName, targetHostingId, ip, ipCandidates, isOffice, host, ua,
              UNIX_TIMESTAMP(createdAt) * 1000 AS createdAt
         FROM admin_token_logs
        WHERE createdAt >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
        ORDER BY logId DESC
        LIMIT 5000`, int64(days-1))
	if err != nil {
		if reNoTable.MatchString(err.Error()) {
			c.Ecb("기록 테이블(admin_token_logs)이 아직 없습니다 — docs/policies/2026-10-02_admin_token_logs.sql 실행 필요")
			return
		}
		c.Ecb(err.Error())
		return
	}
	type grp struct {
		m       *jsx.OrderedMap
		kinds   *jsx.OrderedMap
		targets *jsx.OrderedMap
		issuers *jsx.OrderedMap
		hosts   []any
		hostSet map[string]bool
	}
	var order []string
	byIP := map[string]*grp{}
	okN, denied := 0.0, 0.0
	person := func(set *jsx.OrderedMap, id any, name, cid any) {
		k := jsx.Str(id)
		p, _ := set.V(k).(*jsx.OrderedMap)
		if p == nil {
			p = jsx.Obj("adminId", id, "name", name, "hostingId", cid, "count", float64(0))
			set.Set(k, p)
		}
		p.Set("count", p.N("count")+1)
	}
	for _, r := range rows {
		ip := jsx.Str(r.V("ip"))
		g := byIP[ip]
		if g == nil {
			g = &grp{
				m: jsx.Obj("ip", r.V("ip"), "isOffice", jsx.Truthy(r.V("isOffice")), "total", float64(0), "ok", float64(0), "denied", float64(0),
					"kinds", nil, "targets", nil, "issuers", nil, "hosts", nil, "first", r.V("createdAt"), "last", r.V("createdAt")),
				kinds: jsx.NewOrderedMap(), targets: jsx.NewOrderedMap(), issuers: jsx.NewOrderedMap(), hostSet: map[string]bool{},
			}
			byIP[ip] = g
			order = append(order, ip)
		}
		g.m.Set("total", g.m.N("total")+1)
		if r.V("result") == "OK" {
			g.m.Set("ok", g.m.N("ok")+1)
			okN++
		} else {
			g.m.Set("denied", g.m.N("denied")+1)
			denied++
		}
		kind := jsx.Str(r.V("kind"))
		g.kinds.Set(kind, g.kinds.N(kind)+1)
		if jsx.Truthy(r.V("targetAdminId")) {
			person(g.targets, r.V("targetAdminId"), r.V("targetName"), r.V("targetHostingId"))
		}
		if jsx.Truthy(r.V("issuerAdminId")) {
			person(g.issuers, r.V("issuerAdminId"), r.V("issuerName"), r.V("issuerHostingId"))
		}
		if h := r.V("host"); jsx.Truthy(h) && !g.hostSet[jsx.Str(h)] {
			g.hostSet[jsx.Str(h)] = true
			g.hosts = append(g.hosts, h)
		}
		ca := jsx.Number(r.V("createdAt"))
		if ca < jsx.Number(g.m.V("first")) {
			g.m.Set("first", ca)
		} else {
			g.m.Set("first", jsx.Number(g.m.V("first")))
		}
		if ca > jsx.Number(g.m.V("last")) {
			g.m.Set("last", ca)
		} else {
			g.m.Set("last", jsx.Number(g.m.V("last")))
		}
		if jsx.Truthy(r.V("isOffice")) {
			g.m.Set("isOffice", true)
		}
	}
	byCount := func(set *jsx.OrderedMap) []any {
		var out []any
		for _, k := range set.Keys() {
			out = append(out, set.V(k))
		}
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].(*jsx.OrderedMap).N("count") > out[j].(*jsx.OrderedMap).N("count")
		})
		if out == nil {
			out = []any{}
		}
		return out
	}
	var ips []*jsx.OrderedMap
	for _, ip := range order {
		g := byIP[ip]
		x := jsx.Assign(nil, g.m)
		x.Set("kinds", g.kinds)
		x.Set("targets", byCount(g.targets))
		x.Set("issuers", byCount(g.issuers))
		h := g.hosts
		if h == nil {
			h = []any{}
		}
		x.Set("hosts", h)
		ips = append(ips, x)
	}
	b2n := func(b any) float64 {
		if v, _ := b.(bool); v {
			return 1
		}
		return 0
	}
	sort.SliceStable(ips, func(i, j int) bool {
		a, b := b2n(ips[i].V("isOffice")), b2n(ips[j].V("isOffice"))
		if a != b {
			return a < b
		}
		return ips[i].N("total") > ips[j].N("total")
	})
	office, external := 0.0, 0.0
	ipsAny := []any{}
	for _, g := range ips {
		if g.V("isOffice") == true {
			office++
		} else {
			external++
		}
		ipsAny = append(ipsAny, g)
	}
	recent := []any{}
	for i, r := range rows {
		if i >= 300 {
			break
		}
		recent = append(recent, r)
	}
	c.Scb(jsx.Obj("days", days, "total", float64(len(rows)), "ok", okN, "denied", denied,
		"officeIps", office, "externalIps", external, "ips", ipsAny, "recent", recent))
}
