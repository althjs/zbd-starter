package web

// 세션 쿠키 — 로그인을 대표 도메인(호스팅 SITE 설정 cookieHost, 즈비드 = .zbd.kr) 쿠키로 나눠 쓴다(2026-10-03 대표).
//
// 예전엔 토큰을 출처마다 localStorage 에 두어 chat·mail·talk·{주소}.zbd.kr 에서 매번 다시 로그인해야 했다.
// 이제 서버가 HttpOnly 쿠키를 심는다 — 화면 스크립트는 토큰을 읽지 못한다(앱·미니홈피 스크립트 사고가 토큰으로 번지지 않게).
//
//   - 회원 쿠키 zbd_user_session  : 사용자 주소(zbd.kr·talk·{주소}.zbd.kr) 요청에서만 받는다
//   - 관리자 쿠키 zbd_admin_session : 관리자 주소(SITE 설정 domains.admin — domains.go) 요청에서만 받는다
//   - 쿠키로 들어온 요청은 Origin 이 같은 쪽(사용자/관리자) 의 zbd.kr 주소여야 한다 — 서브도메인끼리는 같은 사이트라
//     SameSite 가 막아 주지 않는다(미니홈피 페이지가 backyard API 를 부르는 것을 여기서 끊는다). GET 이 아닌데 Origin 이 없으면 거절.
//   - 유지 기간: 회원 = SITE 설정 userSessionDays · 관리자 = SITE adminSessionDays(둘 다 없으면 30일, 최대 400일). 마지막 활동(토큰 갱신 시각)부터 센다.
//     쿠키로 들어온 요청이 한 시간 넘게 갱신되지 않았으면 갱신 시각을 지금으로 바꿔 다시 심는다(접속하면 기간이 다시 시작).
//   - 헤더 토큰(Authorization) 은 그대로 받는다 — MCP·중계·스크립트·개발(localhost). 기간 검사는 쿠키에만.
//   - zbd_user_in / zbd_admin_in = 「로그인돼 있음」 표시(값 1, HttpOnly 아님) — 화면이 손님일 때 프로필 요청을 건너뛴다.

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/auth"
)

const (
	UserSessionCookie  = "zbd_user_session"
	AdminSessionCookie = "zbd_admin_session"
	userMarkCookie     = "zbd_user_in"
	adminMarkCookie    = "zbd_admin_in"
	defaultSessionDays = 30
	maxSessionDays     = 400 // 브라우저가 쿠키 기간을 400일로 자른다
	sessionTouchMs     = int64(time.Hour / time.Millisecond)
)

func hostOnly(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if i := strings.LastIndexByte(h, ':'); i >= 0 && !strings.Contains(h[i:], "]") {
		h = h[:i]
	}
	return h
}

// isAdminHost — 관리자 주소인가(SITE 설정 domains.admin — domains.go)
func (c *Ctx) isAdminHost(h string) bool {
	if base, _ := tunnelHosts(c.R); base != "" && tunnelUnder(h, base) { // 외부 접속 주소 — 앱 주소가 관리자 쪽(session_tunnel.go)
		return hostOnly(h) != base
	}
	return c.AdminHostsFor(h) != nil
}

// IsAdminHost — 관리자 주소인가. 페이지 문서가 관리자 주소의 호스팅을 정할 때
func (c *Ctx) IsAdminHost(h string) bool { return c.isAdminHost(h) }

// domainFor — 이 주소가 속한 호스팅의 쿠키 도메인(SITE 설정 cookieHost, 예 ".zbd.kr" → "zbd.kr"). 없으면 ""(localhost 등 — 쿠키를 쓰지 않는다).
func (c *Ctx) domainFor(h string) string {
	h = hostOnly(h)
	if base, _ := tunnelHosts(c.R); base != "" && tunnelUnder(h, base) { // 외부 접속 주소 — 그 고객 사이트 주소 아래만(session_tunnel.go)
		return base
	}
	for _, x := range c.App.Settings.Hostings() {
		m, _ := x.(*jsx.OrderedMap)
		if m == nil {
			continue
		}
		site := c.App.Settings.Site(jsx.StrOrEmpty(m.V("hostingId")))
		// 주소 목록(domains)이 있으면 거기서 계산 — 저장된 cookieHost 는 그 계산이 바뀌기 전 값일 수 있다(domains.go CookieHost)
		d := strings.TrimPrefix(SiteDomains(site, c.App.Cfg.Get("ADMIN_HOSTS")).CookieHost(), ".")
		if site == nil || site.Map("domains") == nil {
			d = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(jsx.StrOrEmpty(site.V("cookieHost")))), ".")
		}
		if d != "" && (h == d || strings.HasSuffix(h, "."+d)) {
			return d
		}
	}
	return AdminCookieParent(h, c.AdminHostsFor(h))
}

// cookieHostOK — 이 요청 주소에서 세션 쿠키를 쓸 수 있는가
func (c *Ctx) cookieHostOK(h string) bool {
	_, ok := cookieScope(hostOnly(h), c.domainFor(h), c.isAdminHost(h))
	return ok
}

// cookieScope — 세션 쿠키의 Domain. 대표 도메인 아래 주소면 그 도메인(여러 주소가 같이 쓴다),
// 아니어도 관리자 주소(SITE domains.admin)면 그 주소 전용 쿠키(Domain 없음) — 스타터는 관리자 주소가 사이트 주소 아래가 아닐 수 있다
// (admin-starter-test.zbd.kr 은 .starter-test.zbd.kr 밖이라 쿠키를 못 심어 로그인 뒤 바로 튕겼다, 2026-10-06). localhost 는 쿠키를 쓰지 않는다(개발은 헤더 토큰).
func cookieScope(host, domain string, admin bool) (string, bool) {
	if domain != "" {
		return domain, true
	}
	if !admin || host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "127.0.0.1" {
		return "", false
	}
	return "", true
}

// cookieAttr — Set-Cookie 의 Domain 부분(그 주소 전용이면 빈 값)
func (c *Ctx) cookieAttr() string {
	if d := c.domainFor(c.sessHost()); d != "" {
		return "; Domain=" + d
	}
	return ""
}

// sessionCookie — 이 이름의 세션 쿠키 중 이 사이트 키로 풀리는 첫 값. 같은 상위 도메인의 다른 즈비드 사이트(운영 backyard 의 .zbd.kr 쿠키 등)가
// 같은 이름 쿠키를 함께 보내도 남의 것을 읽지 않는다 — 브라우저는 오래된 쿠키를 앞에 보내 r.Cookie 가 남의 값을 골랐다.
func (c *Ctx) sessionCookie(name string) (string, *auth.Token) {
	var vals []string
	for _, ck := range c.R.Cookies() {
		if ck.Name == name && ck.Value != "" {
			vals = append(vals, ck.Value)
		}
	}
	v := pickSession(vals, func(s string) bool { return auth.ParseString(s, c.App.Cfg.AESKey, c.App.Cfg.AESIV).Valid() })
	if v == "" {
		return "", nil
	}
	return v, auth.ParseString(v, c.App.Cfg.AESKey, c.App.Cfg.AESIV)
}

// pickSession — 풀리는 첫 값(없으면 "")
func pickSession(vals []string, valid func(string) bool) string {
	for _, v := range vals {
		if valid(v) {
			return v
		}
	}
	return ""
}

// requestHost — 요청 Host(nginx 가 그대로 넘긴다). X-Forwarded-Host 는 보내는 쪽이 꾸밀 수 있어 보지 않는다.
func requestHost(r *http.Request) string { return r.Host }

func (c *Ctx) adminRealm() bool { return c.isAdminHost(c.sessHost()) }

// SessionDays — 이 요청 쪽(사용자/관리자) 의 로그인 유지 기간(일)
func (c *Ctx) SessionDays(hostingID string) int {
	return sessionDaysFrom(c.App.Settings.Site(hostingID), c.adminRealm())
}

// sessionDaysFrom — 정책: 회원 = SITE.userSessionDays · 관리자 = SITE.adminSessionDays(관리자 「환경 설정」, 2026-10-04 — 그 전엔 30일 고정).
// 없거나 0 이하면 30일, 400일 넘으면 400(브라우저가 쿠키 기간을 400일로 자른다). 테스트 session_policy_test.go
func sessionDaysFrom(site *jsx.OrderedMap, admin bool) int {
	key := "userSessionDays"
	if admin {
		key = "adminSessionDays"
	}
	n := int(jsx.Number(site.V(key)))
	if n <= 0 {
		return defaultSessionDays
	}
	if n > maxSessionDays {
		return maxSessionDays
	}
	return n
}

// originOK — 쿠키 로그인 요청의 출처 검사(같은 쪽 zbd.kr 주소만)
func (c *Ctx) originOK() bool {
	o := c.R.Header.Get("Origin")
	if o == "" {
		return c.Method == http.MethodGet || c.Method == http.MethodHead
	}
	u, err := url.Parse(o)
	if err == nil {
		u = c.tunnelOrigin(u)
	}
	if err != nil || u.Scheme != "https" || c.domainFor(u.Host) != c.domainFor(c.sessHost()) {
		return false
	}
	// 그 주소 전용 쿠키(대표 도메인 밖 관리자 주소)면 같은 주소에서 온 요청만
	if c.domainFor(c.sessHost()) == "" && hostOnly(u.Host) != hostOnly(c.sessHost()) {
		return false
	}
	return c.isAdminHost(u.Host) == c.adminRealm()
}

// loadSession — 헤더 토큰이 없으면 이쪽 세션 쿠키를 읽는다. 기간이 지났거나 출처가 맞지 않으면 손님.
func (c *Ctx) loadSession() {
	if bearer(c.R.Header.Get("Authorization")) != "" || !c.cookieHostOK(c.sessHost()) {
		return
	}
	name := UserSessionCookie
	if c.adminRealm() {
		name = AdminSessionCookie
	}
	raw, t := c.sessionCookie(name)
	if raw == "" {
		return
	}
	if !t.Valid() || (t.MemberType == "A") != c.adminRealm() || !c.originOK() {
		return
	}
	now := time.Now().UnixMilli()
	if now-t.LastActive() > int64(c.SessionDays(t.HostingID))*86400000 {
		c.ClearSession()
		return
	}
	if t.BrowserUID == auth.AnyBrowser {
		t.HeaderUID = c.BrowserUID()
	}
	c.Token = t
	c.fromCookie = true
	if now-t.LastActive() > sessionTouchMs {
		c.setSessionCookie(t.Refreshed(now, c.App.Cfg.AESKey, c.App.Cfg.AESIV), c.SessionDays(t.HostingID))
		uid := t.BrowserUID
		if uid == auth.AnyBrowser {
			uid = t.HeaderUID
		}
		c.App.TouchLoginActivity(t.Key, t.MemberType, uid, t.HostingID) // 접속 통계 「마지막 활동일」
	}
}

// bearer — 헤더 값에서 토큰만("Bearer " 만 온 것은 없음으로 본다 — 쿠키를 쓴다)
func bearer(h string) string {
	h = strings.TrimSpace(h)
	if len(h) >= 6 && strings.EqualFold(h[:6], "bearer") {
		h = strings.TrimSpace(h[6:])
	}
	return h
}

// FromSessionCookie — 이 요청이 세션 쿠키로 로그인했는가
func (c *Ctx) FromSessionCookie() bool { return c.fromCookie }

// SessionToken — 지금 요청의 로그인 토큰 원문(헤더 또는 쿠키). 없으면 "".
func (c *Ctx) SessionToken() string {
	if raw := bearer(c.R.Header.Get("Authorization")); raw != "" {
		return raw
	}
	if !c.fromCookie {
		return ""
	}
	name := UserSessionCookie
	if c.adminRealm() {
		name = AdminSessionCookie
	}
	raw, _ := c.sessionCookie(name)
	return raw
}

func (c *Ctx) setSessionCookie(tok string, days int) {
	name, mark := UserSessionCookie, userMarkCookie
	if c.adminRealm() {
		name, mark = AdminSessionCookie, adminMarkCookie
	}
	age := strconv.Itoa(days * 86400)
	tail := c.cookieAttr() + "; Path=/; Max-Age=" + age + "; Secure; SameSite=Lax"
	h := c.W.Header()
	h.Add("Set-Cookie", name+"="+tok+tail+"; HttpOnly")
	h.Add("Set-Cookie", mark+"=1"+tail)
	h.Add("Cache-Control", "private")
}

// IssueSession — 새로 발급한 토큰을 이쪽 세션 쿠키로 심는다. 심었으면 true — 그때는 응답 본문에 토큰을 싣지 않는다
// (화면 스크립트가 토큰을 보지 못하게). 대표 도메인이 아니거나(개발) 토큰의 쪽이 주소와 다르면 false.
func (c *Ctx) IssueSession(tok string) bool {
	// 스크립트(맥미니 중계·board.mjs 등)는 헤더 X-Token-Delivery: body 로 「본문으로 달라」 고 한다 — 쿠키를 못 쓰는 쪽
	if !c.cookieHostOK(c.sessHost()) || strings.EqualFold(c.R.Header.Get("X-Token-Delivery"), "body") {
		return false
	}
	t := auth.ParseString(tok, c.App.Cfg.AESKey, c.App.Cfg.AESIV)
	if !t.Valid() || (t.MemberType == "A") != c.adminRealm() {
		return false
	}
	c.setSessionCookie(tok, c.SessionDays(t.HostingID))
	return true
}

// ClearSession — 이쪽 세션 쿠키를 지운다(로그아웃·기간 지남)
func (c *Ctx) ClearSession() {
	if !c.cookieHostOK(c.sessHost()) {
		return
	}
	name, mark := UserSessionCookie, userMarkCookie
	if c.adminRealm() {
		name, mark = AdminSessionCookie, adminMarkCookie
	}
	// 지금 범위 + 그 주소 전용 + 이 주소의 상위들 — 쿠키 범위를 넓히기 전(2026-10-20)에 심은 것도 같이 지운다(남으면 로그아웃해도 로그인 그대로)
	attrs := []string{c.cookieAttr(), ""}
	h := hostOnly(c.sessHost())
	for p := h; strings.Count(p, ".") >= 1; p = p[strings.IndexByte(p, '.')+1:] {
		if a := "; Domain=" + p; a != attrs[0] {
			attrs = append(attrs, a)
		}
	}
	for _, a := range attrs {
		tail := "=" + a + "; Path=/; Max-Age=0; Secure; SameSite=Lax"
		c.W.Header().Add("Set-Cookie", name+tail+"; HttpOnly")
		c.W.Header().Add("Set-Cookie", mark+tail)
	}
}

// PasskeyRP — 패스키(WebAuthn) 의 RP ID·출처. 요청 Origin 기준: 대표 도메인 주소면 RP = 그 도메인(zbd.kr — backyard·chat·mail 이 한 패스키를 같이 쓴다),
// 개발(localhost) 이면 RP = localhost 이고 한 출처에 두 화면이 있어 주소의 ?realm=user 로 회원 쪽을 고른다(운영 주소에선 이 값을 보지 않는다).
// admin = 관리자 주소. 그 밖의 출처면 ok=false. 테스트 realm_policy_test.go
func (c *Ctx) PasskeyRP() (rpID, origin string, admin, ok bool) {
	origin = c.R.Header.Get("Origin")
	u, err := url.Parse(origin)
	if origin == "" || err != nil {
		return "", "", false, false
	}
	u = c.tunnelOrigin(u)
	if u.Scheme == "https" && origin != u.String() { // 외부 접속 주소 — 브라우저가 본 출처
		origin = u.String()
	}
	if hostOnly(u.Host) == "localhost" { // 127.0.0.1 은 RP 가 될 수 없다(IP 금지)
		return "localhost", origin, c.R.URL.Query().Get("realm") != "user", u.Scheme == "http" || u.Scheme == "https"
	}
	d := c.domainFor(u.Host)
	if d == "" || u.Scheme != "https" {
		return "", "", false, false
	}
	return d, origin, c.isAdminHost(u.Host), true
}
