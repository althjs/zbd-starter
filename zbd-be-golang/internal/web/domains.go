package web

// 도메인 — 사이트의 사용자 주소·관리자 주소는 SITE 설정 domains 한 곳에 둔다(2026-10-06 대표 「하나로만」).
//
//	SITE.domains = {"user": ["example.com", "www.example.com"], "admin": ["admin.example.com"]} — 각 목록의 첫 번째가 대표 주소.
//
// 홈 주소(homepage)·쿠키 범위(cookieHost)·검색엔진 사이트맵 주소는 이 두 목록에서 계산해 같이 저장한다(관리자 「환경 설정 › 도메인」,
// API /admin/domain). 서버(관리자 주소 판단·쿠키)·화면(<meta name="zbd:admin-hosts">)·소켓 서버(/socket/domains)는 이 값을 읽기만 한다.
// 같은 주소가 두 목록에 다 있으면(개발 localhost:5000) 「한 주소」 — 관리자 주소로 보지 않는다(쿠키 없이 헤더 토큰, 화면 주소 검사 끔).
// 예전 설정(domains 없음)은 홈 주소 + .env ADMIN_HOSTS 로 읽는다 — 옮기는 기간(배포·저장하면 domains 가 생긴다).

import (
	"errors"
	"net"
	"regexp"
	"strings"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
)

// Domains — 한 사이트의 주소 목록(호스트[:포트], 소문자)
type Domains struct {
	User  []string `json:"user"`
	Admin []string `json:"admin"`
}

var hostRe = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*(:[0-9]{1,5})?$`)

// NormalizeHost — 사람이 넣은 주소를 「호스트[:포트]」 로(https://·끝 / 는 떼고, 경로·공백·잘못된 글자면 false)
func NormalizeHost(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(strings.TrimPrefix(s, "https://"), "http://")
	s = strings.TrimSuffix(s, "/")
	if s == "" || len(s) > 253 || !hostRe.MatchString(s) {
		return "", false
	}
	return s, true
}

// IsLocalHost — 개발용 주소(localhost·*.localhost·127.0.0.1 · 같은 네트워크의 사설 IP 192.168.x·10.x·172.16~31.x)
// 사설 IP 는 개발 화면을 휴대폰·다른 PC 에서 열 때 — 운영 주소로 쓸 수 없는 값이라 개발로 본다(2026-10-06)
func IsLocalHost(h string) bool {
	h = hostOnly(h)
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate()
	}
	return h == "localhost" || strings.HasSuffix(h, ".localhost")
}

func hostList(v []any) []string {
	var out []string
	for _, x := range v {
		if h, ok := NormalizeHost(jsx.StrOrEmpty(x)); ok {
			out = append(out, h)
		}
	}
	return out
}

// SiteDomains — SITE 설정의 주소 목록. domains 가 없으면 옛 설정(홈 주소 · .env ADMIN_HOSTS)으로
func SiteDomains(site *jsx.OrderedMap, envAdmin string) Domains {
	if site != nil {
		if dm := site.Map("domains"); dm != nil {
			return Domains{User: hostList(dm.Arr("user")), Admin: hostList(dm.Arr("admin"))}
		}
	}
	var d Domains
	if site != nil {
		hp := jsx.StrOrEmpty(site.V("homepage"))
		if hp == "" {
			hp = jsx.StrOrEmpty(site.Path("design", "homepage"))
		}
		if h, ok := NormalizeHost(hp); ok {
			d.User = []string{h}
		}
	}
	for _, h := range strings.Split(envAdmin, ",") {
		if h, ok := NormalizeHost(h); ok {
			d.Admin = append(d.Admin, h)
		}
	}
	// 개발(관리자 주소를 따로 두지 않던 localhost) = 한 주소
	if len(d.Admin) == 0 && len(d.User) > 0 && IsLocalHost(d.User[0]) {
		d.Admin = []string{d.User[0]}
	}
	return d
}

func hasHost(list []string, h string) bool {
	h = hostOnly(h)
	for _, x := range list {
		if hostOnly(x) == h {
			return true
		}
	}
	return false
}

// AdminOnly — 관리자 전용 주소(사용자 목록에도 있는 주소 = 한 주소 모드라 뺀다)
func (d Domains) AdminOnly() []string {
	var out []string
	for _, h := range d.Admin {
		if !hasHost(d.User, h) {
			out = append(out, h)
		}
	}
	return out
}

// Homepage — 대표 사용자 주소(개발 주소는 http)
func (d Domains) Homepage() string {
	if len(d.User) == 0 {
		return ""
	}
	if IsLocalHost(d.User[0]) {
		return "http://" + d.User[0] + "/"
	}
	return "https://" + d.User[0] + "/"
}

// CookieHost — 세션 쿠키 범위 = "." + 대표 사용자 주소(개발 주소면 ""). 그 아래 관리자 주소는 같이 쓰고,
// 밖의 관리자 주소는 그 주소 전용 쿠키(cookieScope)라 따로 정할 것이 없다.
func (d Domains) CookieHost() string {
	if len(d.User) == 0 || IsLocalHost(d.User[0]) {
		return ""
	}
	return "." + hostOnly(d.User[0])
}

// AdminCookieParent — 사용자 쿠키 범위 밖의 관리자 주소가 쓸 쿠키 범위: 관리자 목록(순서대로)에서 이 주소이거나 이 주소의 상위인 첫 주소.
// 단독 앱 주소를 관리자 주소 아래에 두면(chat.admin.example.com) 관리자·앱이 한 로그인을 같이 쓴다(2026-10-20 오토카지 — 예전엔 주소마다 전용 쿠키라
// 앱마다 다시 로그인했다). 해당 없으면 ""(그 주소 전용). 화면 config.ts cookieHostHere 와 같은 규칙.
func AdminCookieParent(h string, admins []string) string {
	h = hostOnly(h)
	for _, a := range admins {
		a = hostOnly(a)
		if a == "" || IsLocalHost(a) {
			continue
		}
		if h == a || strings.HasSuffix(h, "."+a) {
			return a
		}
	}
	return ""
}

// ValidateDomains — 저장 전에 검사. current = 지금 이 화면을 연 주소(그 주소를 관리자 목록에서 빼면 잠긴다)
func ValidateDomains(d Domains, current string) error {
	if len(d.User) == 0 {
		return errors.New("사용자 주소를 하나 이상 넣어 주세요")
	}
	if len(d.Admin) == 0 {
		return errors.New("관리자 주소를 하나 이상 넣어 주세요")
	}
	seen := map[string]string{}
	for _, side := range []struct {
		name string
		list []string
	}{{"사용자", d.User}, {"관리자", d.Admin}} {
		for _, h := range side.list {
			k := hostOnly(h)
			if prev, ok := seen[k]; ok {
				if prev == side.name {
					return errors.New(side.name + " 주소가 겹칩니다: " + h)
				}
				// 두 목록에 같은 주소 = 한 주소(개발)만
				if !IsLocalHost(k) {
					return errors.New("같은 주소를 사용자·관리자에 함께 쓸 수 없습니다(개발 주소 localhost 만 됩니다): " + h)
				}
			}
			seen[k] = side.name
		}
	}
	// 개발 주소(localhost — 개발 화면은 vite 가 127.0.0.1 로 넘긴다)는 검사하지 않는다: 관리자 목록에 없어도 개발 주소에선 화면 주소 검사가 꺼져 잠기지 않는다
	if current != "" && !IsLocalHost(current) && !hasHost(d.Admin, current) {
		return errors.New("지금 접속한 관리자 주소(" + hostOnly(current) + ")를 빼면 관리자 화면에 못 들어옵니다 — 새 주소에서 접속해 확인한 뒤 빼 주세요")
	}
	return nil
}

// siteDomainsAll — 모든 호스팅의 주소 목록
func (c *Ctx) siteDomainsAll() []Domains {
	env := c.App.Cfg.Get("ADMIN_HOSTS") // 옛 설정 — 옮기는 기간
	var out []Domains
	for _, x := range c.App.Settings.Hostings() {
		if m, _ := x.(*jsx.OrderedMap); m != nil {
			out = append(out, SiteDomains(c.App.Settings.Site(jsx.StrOrEmpty(m.V("hostingId"))), env))
		}
	}
	return out
}

// AdminHostsFor — 요청 주소가 관리자 주소면 그 사이트의 관리자 전용 주소 목록(첫 번째가 대표), 아니면 nil
func (c *Ctx) AdminHostsFor(h string) []string {
	for _, d := range c.siteDomainsAll() {
		if a := d.AdminOnly(); hasHost(a, h) {
			return a
		}
	}
	return nil
}

// AdminHostsOfUserHost — 이 사용자 주소가 속한 사이트의 관리자 전용 주소. 디자인 미리보기 창(사용자 주소)이
// 관리자 편집 화면(다른 주소)이 보내는 내용을 받을 때 그 주소를 알아야 한다(2026-10-08 — 모르면 미리보기가 비었다)
func (c *Ctx) AdminHostsOfUserHost(h string) []string {
	for _, d := range c.siteDomainsAll() {
		if hasHost(d.User, h) {
			return d.AdminOnly()
		}
	}
	return nil
}

// AdminHostsAll — 모든 사이트의 관리자 전용 주소(소켓 서버용)
func (c *Ctx) AdminHostsAll() []string {
	var out []string
	for _, d := range c.siteDomainsAll() {
		out = append(out, d.AdminOnly()...)
	}
	return out
}

// HostingIDFor — 이 주소(사용자·관리자)가 속한 호스팅. 못 찾으면 ""
func HostingIDFor(a *app.App, host string) string {
	env := a.Cfg.Get("ADMIN_HOSTS")
	for _, x := range a.Settings.Hostings() {
		m, _ := x.(*jsx.OrderedMap)
		if m == nil {
			continue
		}
		id := jsx.StrOrEmpty(m.V("hostingId"))
		if d := SiteDomains(a.Settings.Site(id), env); hasHost(d.User, host) || hasHost(d.Admin, host) {
			return id
		}
	}
	return ""
}
