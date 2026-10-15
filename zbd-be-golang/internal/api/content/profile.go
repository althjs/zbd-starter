package content

// 내 프로필 — GET /user/profile (화면이 로그인 상태를 확인할 때마다 부른다)
//
// 토큰이 이 브라우저(zbduid)·이 사이트(referer)·등록된 기기의 것인지 본 뒤 회원 정보 + 새 토큰을 준다.
// 세션 시간 제한(회사 server.sessionTimeoutHourUser)이 있으면 처음 로그인 시각 기준으로 끊는다.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

var envHostPrefixes = []string{"staging-", "alpha-", "dev-"}

func stripEnvHostPrefix(h string) string {
	for _, p := range envHostPrefixes {
		if strings.HasPrefix(h, p) {
			return h[len(p):]
		}
	}
	return h
}

// hostingIDByHost — 주소로 회사 찾기(사이트빌더 주소는 옮기지 않았다 — 없으면 "")
func hostingIDByHost(c *web.Ctx, host string) string {
	if host == "" {
		return ""
	}
	homepage := "https://" + stripEnvHostPrefix(host) + "/"
	if homepage == "https://talk.zbd.kr/" {
		homepage = "https://zbd.kr/"
	}
	for _, x := range c.App.Settings.Hostings() {
		if m, _ := x.(*jsx.OrderedMap); m != nil && jsx.Str(m.V("homepage")) == homepage {
			return jsx.StrOrEmpty(m.V("hostingId"))
		}
	}
	return ""
}

// checkRequiredUserInfo — 가입 때 꼭 받아야 할 정보가 빠졌나(회사 설정에 따라)
func checkRequiredUserInfo(c *web.Ctx, m *jsx.OrderedMap, js *jsx.OrderedMap) bool {
	f := c.App.Settings.Config(jsx.StrOrEmpty(m.V("hostingId"))).Path("SITE", "feature")
	feature, _ := f.(*jsx.OrderedMap)
	if !jsx.Truthy(feature.V("userMemberJoinForceAdditionalInfo")) {
		return false
	}
	switch {
	case !jsx.Truthy(m.V("name")):
		return true
	case jsx.Truthy(feature.V("userMemberNicknameRequired")) && !jsx.Truthy(m.V("nickname")):
		return true
	case jsx.Truthy(feature.V("useMemberDemographic")) && !jsx.Truthy(js.V("birthDate")):
		return true
	case jsx.Truthy(feature.V("isRequiredMemberAddress")) && !jsx.Truthy(js.Path("address", "address1")):
		return true
	case jsx.Truthy(feature.V("useMemberJoinMobileVerify")) && !jsx.Truthy(js.V("me")) && !jsx.Truthy(js.V("isMobileVerified")):
		return true
	}
	return false
}

// getUserProfile — 화면이 쓰는 내 정보 모양(+ 새 토큰). 탈퇴·정지 회원이면 nil.
func getUserProfile(ctx context.Context, c *web.Ctx, q db.Querier, userID string) (*jsx.OrderedMap, error) {
	m, err := member.Get(ctx, c.App, q, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errString("TypeError: Cannot read properties of null (reading 'status')")
	}
	if st, ok := m.V("status").(float64); !ok || st != 0 {
		return nil, nil
	}
	t := c.Token
	js := jsx.ParseMap(orStr(m.V("json"), "{}"))
	jj := jsx.Obj("agreeMap", orVal(js.V("agreeMap"), jsx.NewOrderedMap()))
	for _, k := range []string{"shopId", "cars", "address", "school"} {
		if jsx.Truthy(js.V(k)) {
			jj.Set(k, js.V(k))
		}
	}
	p := jsx.Obj(
		"userId", m.V("userId"),
		"groupId", m.V("groupId"),
		"email", orVal(m.V("email"), ""),
		"name", orVal(m.V("name"), ""),
		"mobile", orVal(m.V("mobile"), ""),
		"nickname", orVal(m.V("nickname"), ""),
		"image", orVal(m.V("image"), ""),
		"description", orVal(m.V("description"), ""),
		"hostingId", orVal(m.V("hostingId"), t.HostingID),
		"productIds", orVal(js.V("productIds"), []any{}),
		"postIds", orVal(js.V("postIds"), []any{}),
		"commerceProductIds", orVal(js.V("commerceProductIds"), []any{}),
	)
	if jsx.Truthy(js.V("isDeveloper")) {
		p.Set("isDeveloper", true)
	}
	p.Set("auth", orVal(js.V("auth"), jsx.NewOrderedMap()))
	p.Set("commerceCartCount", 0.0)
	p.Set("Json", jj)
	if isOffice(c) {
		p.Set("isOffice", true)
	}
	if c.QV("check") != "1" { // 토큰 유효성만 볼 때는 새 토큰을 주지 않는다
		{
			now := time.Now().UnixMilli()
			ts, rts := t.TS, int64(0)
			if ts == 0 {
				ts = now
			} else {
				rts = now
			}
			mt := t.MemberType
			if mt == "" {
				mt = "U"
			}
			group := jsx.Str(m.V("groupId"))
			if group == "" {
				group = "-"
			}
			parts := []string{
				t.Key, t.Email, strconv.FormatInt(ts, 10), mt, t.BrowserUID, t.HostingID,
				jsx.StrOrEmpty(m.V("name")), group, "undefined", "0", jsx.StrOrEmpty(m.V("nickname")),
				strconv.FormatInt(rts, 10), "0",
			}
			if tok := (util.AES{Key: c.App.Cfg.AESKey, IV: c.App.Cfg.AESIV}).Encrypt(strings.Join(parts, "|")); !c.IssueSession(tok) {
				p.Set("token", tok)
			}
		}
	}
	if jsx.Truthy(c.App.Settings.Config(t.HostingID).Path("SITE", "feature", "userMemberJoinForceAdditionalInfo")) && checkRequiredUserInfo(c, m, js) {
		p.Set("isAdditionalInfoRequired", true)
	}
	return p, nil
}

func userProfile(c *web.Ctx) {
	t := c.Token
	withConn(c, func(ctx context.Context, conn *sql.Conn) {
		hosting := c.App.Settings.Hosting(t.HostingID)
		if hours := jsx.Number(hosting.Path("server", "sessionTimeoutHourUser")); hours > 0 {
			ttl := int64(1000 * 60 * 60 * hours)
			uts := t.RTS
			if uts == 0 {
				uts = t.TS
			}
			if uts+ttl < time.Now().UnixMilli() {
				c.Ecb("세션이 만료되었습니다. 다시 로그인해 주세요.")
				return
			}
		}
		ref := c.Header("Referer")
		host := ""
		if i := strings.Index(ref, "//"); i >= 0 {
			host, _, _ = strings.Cut(ref[i+2:], "/")
		}
		hostingID := hostingIDByHost(c, host)

		// token.browserUid != 브라우저 id — 토큰이 없으면 undefined, 브라우저 id 가 없어도 undefined
		var tokUID, hdrUID any
		if t.MemberType != "" {
			tokUID = t.BrowserUID
		}
		if uid := c.BrowserUID(); uid != "" {
			hdrUID = uid
		}
		if !looseEq(tokUID, hdrUID) || (hostingID != "" && t.HostingID != hostingID) {
			c.Ecb("유효하지 않은 토큰입니다.")
			return
		}
		r, err := db.QueryOne(ctx, conn, `SELECT count(*) as cnt FROM login_devices WHERE userId = ? AND memberType = ? AND email = ? AND browserUid = ? AND hostingId = ?`,
			nilIfEmpty(t.Key, t), nilIfEmpty(t.MemberType, t), c.App.PII.Hash("email", nilIfEmpty(t.Email, t)), nilIfEmpty(t.BrowserUID, t), nilIfEmpty(t.HostingID, t))
		if err != nil {
			c.Ecb(err)
			return
		}
		if jsx.Number(r.V("cnt")) <= 0 {
			c.Ecb("유효하지 않은 토큰입니다..")
			return
		}
		if c.Method != "GET" {
			c.Ecb("잘못된 요청")
			return
		}
		userID := t.Key
		if userID == "" || t.MemberType != "U" {
			c.Ecb("잘못된 요청")
			return
		}
		devs, err := db.Query(ctx, conn, `select browserUid, 
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt, 
      UNIX_TIMESTAMP(lastLoginedAt)*1000 as lastLoginedAt
    FROM login_devices WHERE userId = ? AND memberType = ? AND email = ? AND hostingId = ?`, t.Key, "U", c.App.PII.Hash("email", t.Email), t.HostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		site := c.App.Settings.Site(t.HostingID)
		found := false
		for _, d := range devs {
			if looseEq(d.V("browserUid"), t.BrowserUID) {
				found = true
				break
			}
		}
		if t.BrowserUID == "" || !found {
			host := jsx.StrOrEmpty(site.V("cookieHost"))
			if c.App.Cfg.Get("NODE_ENV") == "development" {
				host = ""
			}
			dom := ""
			if host != "" {
				dom = "Domain=" + host + ";"
			}
			c.W.Header().Add("Set-Cookie", "zbd_uid=deleted;"+dom+" Max-Age=0; path=/")
			c.W.Header().Add("Set-Cookie", "ox_uid=deleted;"+dom+" Max-Age=0; path=/") // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
			ecbCode(c, "토큰정보가 유효하지 않습니다.", "1000")
			return
		}
		m, err := member.Get(ctx, c.App, conn, userID)
		if err != nil {
			c.Ecb(err)
			return
		}
		if m == nil {
			c.Ecb("TypeError: Cannot read properties of null (reading 'status')")
			return
		}
		if st, ok := m.V("status").(float64); !ok || st != 0 {
			c.Ecb("사용자 정보가 없습니다.")
			return
		}
		p, err := getUserProfile(ctx, c, conn, userID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(nilIfNoMap(p))
	})
}

// nilIfEmpty — 토큰이 없을 때(지금 서비스는 {key:0} 만 — 나머지 칸은 undefined → NULL)
func nilIfEmpty(v string, t interface{ Valid() bool }) any {
	if !t.Valid() {
		return nil
	}
	return v
}
