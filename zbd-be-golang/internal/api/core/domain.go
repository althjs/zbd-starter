package core

// 도메인 관리(2026-10-06 대표 「하나로만」) — 사용자 주소·관리자 주소를 SITE 설정 domains 한 곳에(규칙 web/domains.go).
//
//	GET /admin/domain                    지금 주소 목록 + 계산한 홈 주소·쿠키 범위 + 지금 접속 주소
//	PUT /admin/domain {user[], admin[]}  저장 — 개발자·최고 관리자만. 홈 주소·쿠키 범위·사이트맵 주소·호스팅 목록 homepage 를 같이 바꾼다
//	GET /socket/domains                  소켓 서버 전용(헤더 X-Socket-Secret) — 관리자 주소·쿠키 범위
//
// nginx·인증서는 서버 권한이 필요해 여기서 바꾸지 않는다 — 저장 뒤 `zbd deploy` 가 DB 주소대로 맞춘다.

import (
	"crypto/subtle"
	"regexp"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

func domainView(c *web.Ctx, d web.Domains) *jsx.OrderedMap {
	return jsx.Obj("user", strList(d.User), "admin", strList(d.Admin), "homepage", d.Homepage(), "cookieHost", d.CookieHost(),
		"current", strings.ToLower(c.R.Host))
}

func strList(a []string) []any {
	out := make([]any, 0, len(a))
	for _, s := range a {
		out = append(out, s)
	}
	return out
}

var reSitemapLine = regexp.MustCompile(`(?m)^Sitemap:.*$`)

// adminDomain — GET · PUT
func adminDomain(c *web.Ctx) {
	ctx := c.Context()
	hostingID := c.Token.HostingID
	switch c.Method {
	case "GET":
		c.Scb(domainView(c, web.SiteDomains(c.App.Settings.Site(hostingID), c.App.Cfg.Get("ADMIN_HOSTS"))))
	case "PUT":
		if isDev, isSuper := callerFlags(c); !isDev && !isSuper {
			c.Ecb("개발자·최고 관리자만 바꿀 수 있습니다")
			return
		}
		bd := c.BodyMap()
		var d web.Domains
		for _, side := range []struct {
			key string
			out *[]string
		}{{"user", &d.User}, {"admin", &d.Admin}} {
			for _, x := range bd.Arr(side.key) {
				raw := jsx.StrOrEmpty(x)
				if strings.TrimSpace(raw) == "" {
					continue
				}
				h, ok := web.NormalizeHost(raw)
				if !ok {
					c.Ecb("주소 형식이 맞지 않습니다: " + raw)
					return
				}
				*side.out = append(*side.out, h)
			}
		}
		if err := web.ValidateDomains(d, c.R.Host); err != nil {
			c.Ecb(err.Error())
			return
		}
		conn, err := c.App.DB.Conn(ctx)
		if err != nil {
			c.Ecb(err)
			return
		}
		defer conn.Close()
		// 사이트 설정 — 읽고 바꿀 칸만
		site, _ := func() (*jsx.OrderedMap, error) {
			v, err := metaJSONGet(ctx, conn, "SITE", hostingID)
			m, _ := v.(*jsx.OrderedMap)
			return m, err
		}()
		if site == nil {
			c.Ecb("사이트 설정(SITE)을 읽지 못했습니다")
			return
		}
		hp := d.Homepage()
		site.Set("domains", jsx.Obj("user", strList(d.User), "admin", strList(d.Admin)))
		site.Set("homepage", hp)
		site.Set("cookieHost", d.CookieHost())
		design := site.Map("design")
		if design == nil {
			design = jsx.NewOrderedMap()
			site.Set("design", design)
		}
		design.Set("homepage", hp)
		if rb := jsx.StrOrEmpty(design.V("robots")); reSitemapLine.MatchString(rb) {
			design.Set("robots", reSitemapLine.ReplaceAllString(rb, "Sitemap: "+hp+"sitemap.xml"))
		}
		if err := metaJSONSet(c, conn, "SITE", site, hostingID); err != nil {
			c.Ecb(err)
			return
		}
		// 호스팅 목록의 homepage(옛 방식으로 주소를 찾는 곳) 도 맞춘다
		if v, err := metaJSONGet(ctx, conn, "HOSTINGS", ""); err == nil {
			if hs, _ := v.(*jsx.OrderedMap); hs != nil {
				changed := false
				for _, x := range hs.Arr("items") {
					if it, _ := x.(*jsx.OrderedMap); it != nil && jsx.StrOrEmpty(it.V("hostingId")) == hostingID && jsx.StrOrEmpty(it.V("homepage")) != hp {
						it.Set("homepage", hp)
						changed = true
					}
				}
				if changed {
					if err := metaJSONSet(c, conn, "HOSTINGS", hs, ""); err != nil {
						c.Ecb(err)
						return
					}
				}
			}
		}
		c.Scb(domainView(c, d))
	default:
		c.Ecb("잘못된 요청")
	}
}

// socketDomains — 소켓 서버 전용. 공유 비밀이 맞아야 답한다(관리자 주소는 공개하지 않는다)
func socketDomains(c *web.Ctx) {
	want := strings.TrimSpace(c.App.Cfg.Get("SOCKET_SECRET"))
	if want == "" {
		want = strings.TrimSpace(c.App.Cfg.Get("ADMIN_SOCKET_SECRET"))
	}
	if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(c.R.Header.Get("X-Socket-Secret"))) != 1 {
		c.Status(403).Ecb("거절")
		return
	}
	var cookie []any
	seen := map[string]bool{}
	for _, x := range c.App.Settings.Hostings() {
		m, _ := x.(*jsx.OrderedMap)
		if m == nil {
			continue
		}
		d := web.SiteDomains(c.App.Settings.Site(jsx.StrOrEmpty(m.V("hostingId"))), c.App.Cfg.Get("ADMIN_HOSTS"))
		if ch := strings.TrimPrefix(d.CookieHost(), "."); ch != "" && !seen[ch] {
			seen[ch] = true
			cookie = append(cookie, ch)
		}
	}
	c.Scb(jsx.Obj("admin", strList(c.AdminHostsAll()), "cookie", cookie))
}
