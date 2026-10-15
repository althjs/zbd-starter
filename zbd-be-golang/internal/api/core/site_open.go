package core

// 사이트 공개 여부(2026-10-08 대표 — 즈비드 호스팅 고객만, 규칙 web/site_open.go)
//
//	GET /admin/site-open          {hosted, open} — hosted 가 false 면 화면이 칸을 숨긴다
//	PUT /admin/site-open {open}   저장(SITE siteClosed) — 개발자·최고 관리자, 호스팅 사이트만

import (
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

func adminSiteOpen(c *web.Ctx) {
	hostingID := c.Token.HostingID
	hosted := web.IsHostedSite(c.App.Cfg.Get)
	switch c.Method {
	case "GET":
		c.Scb(jsx.Obj("hosted", hosted, "open", web.SiteOpen(c.App.Settings.Site(hostingID))))
	case "PUT":
		if !hosted {
			c.Ecb("즈비드 호스팅 사이트에서만 쓸 수 있습니다")
			return
		}
		if isDev, isSuper := callerFlags(c); !isDev && !isSuper {
			c.Ecb("개발자·최고 관리자만 바꿀 수 있습니다")
			return
		}
		open := jsx.Truthy(c.BodyMap().V("open"))
		v, err := metaJSONGet(c.Context(), c.App.DB, "SITE", hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		site, _ := v.(*jsx.OrderedMap)
		if site == nil {
			c.Ecb("사이트 설정이 없습니다")
			return
		}
		if open {
			site.Delete("siteClosed")
		} else {
			site.Set("siteClosed", true)
		}
		if err := metaJSONSet(c, c.App.DB, "SITE", site, hostingID); err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(jsx.Obj("hosted", true, "open", open))
	default:
		c.Ecb("잘못된 요청")
	}
}
