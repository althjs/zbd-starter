package web

// 외부 접속 주소(zbd tunnel, starter-N.zbd.co.kr) 의 세션 쿠키 — 2026-10-16 대표 「운영처럼 서버 쿠키로」.
//
// 터널은 고객 PC 의 개발 서버(run_dev)로 들어온다: 방문자 → zbd.kr 터널 서버 → zbd tunnel → 개발 화면(vite) → 이 서버.
// 이 서버가 보는 Host 는 localhost 라 원래는 쿠키를 쓰지 않았고(개발 = 헤더 토큰), 토큰이 주소마다 localStorage 에 있어
// 앱 주소(admin·chat·mail·office·photo.{사이트 주소})로 갈 때마다 다시 로그인했다.
// 이제 zbd tunnel 이 붙여 주는 방문자 주소(X-Zbd-Public-Host)·사이트 주소(X-Zbd-Tunnel-Base — 앱 주소가 준비됐을 때만)로
// 운영과 같은 HttpOnly 쿠키를 사이트 주소(Domain=starter-N.zbd.co.kr)에 심는다 — 그 고객 주소 아래만이라 다른 고객으로 가지 않는다.
//   - 사이트 주소 = 사용자 쪽, 앱 주소 = 관리자 쪽(관리자 주소 목록과 같은 이름 — 개발 화면 hooks.server.ts tunnelHead)
//   - 두 헤더는 zbd.kr 터널 서버·zbd tunnel 이 방문자가 보낸 같은 이름 헤더를 버리고 붙인다. 그래도 이 서버는
//     자기 PC 안(Host·보낸 곳 모두 루프백 — 개발 화면의 프록시)에서 온 요청에서만 본다. 배포한 사이트(nginx 뒤, Host = 실제 주소)는 늘 무시.
//   - zbd tunnel 은 방문자 주소의 Origin 을 localhost 로 바꿔 보낸다 — 그런 Origin 은 그 방문자 주소에서 온 것으로 본다.
// 테스트 session_tunnel_policy_test.go

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var tunnelBaseRe = regexp.MustCompile(`^starter-[0-9]{1,10}\.zbd\.co\.kr$`)

// tunnelAdminApps — 사이트 주소 아래 관리자 쪽 앱 주소
var tunnelAdminApps = map[string]bool{"admin": true, "chat": true, "mail": true, "office": true, "photo": true}

// tunnelHosts — 터널로 온 요청이면 사이트 주소·방문자가 연 주소. 아니면 둘 다 "".
func tunnelHosts(r *http.Request) (base, pub string) {
	if r == nil || !loopbackRequest(r) {
		return "", ""
	}
	return tunnelPair(r.Header.Get("X-Zbd-Tunnel-Base"), r.Header.Get("X-Zbd-Public-Host"))
}

// tunnelPair — 정책: 사이트 주소가 starter-N.zbd.co.kr 이고 방문자 주소가 그 주소 또는 정해진 앱 주소일 때만
func tunnelPair(baseH, pubH string) (string, string) {
	base, pub := hostOnly(baseH), hostOnly(pubH)
	if !tunnelBaseRe.MatchString(base) {
		return "", ""
	}
	if pub == base || tunnelAdminApps[strings.TrimSuffix(pub, "."+base)] && strings.HasSuffix(pub, "."+base) {
		return base, pub
	}
	return "", ""
}

// loopbackRequest — 이 PC 안에서 온 요청(개발 화면의 프록시)인가: Host 와 보낸 곳 모두 루프백
func loopbackRequest(r *http.Request) bool {
	if !isLoopbackName(hostOnly(r.Host)) {
		return false
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	p := net.ParseIP(strings.Trim(ip, "[]"))
	return p != nil && p.IsLoopback()
}

func isLoopbackName(h string) bool {
	h = strings.Trim(h, "[]")
	if h == "localhost" {
		return true
	}
	p := net.ParseIP(h)
	return p != nil && p.IsLoopback()
}

// tunnelUnder — h 가 사이트 주소이거나 그 아래 주소인가
func tunnelUnder(h, base string) bool {
	h = hostOnly(h)
	return h == base || strings.HasSuffix(h, "."+base)
}

// sessHost — 세션 쿠키를 고를 때의 주소: 터널이면 방문자가 연 주소, 아니면 요청 Host
func (c *Ctx) sessHost() string {
	if _, pub := tunnelHosts(c.R); pub != "" {
		return pub
	}
	return requestHost(c.R)
}

// tunnelOrigin — zbd tunnel 이 localhost 로 바꾼 Origin 을 방문자 주소(https)로 되돌린다. 터널이 아니거나 다른 Origin 이면 그대로.
func (c *Ctx) tunnelOrigin(u *url.URL) *url.URL {
	if _, pub := tunnelHosts(c.R); pub != "" && u.Scheme == "http" && isLoopbackName(hostOnly(u.Host)) {
		return &url.URL{Scheme: "https", Host: pub}
	}
	return u
}
