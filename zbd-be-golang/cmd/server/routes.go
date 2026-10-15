package main

// 경로 등록 — 기능별 패키지가 자기 라우트를 등록한다(internal/api/<기능>/).
//   경로는 지금 서비스와 같은 모양("/admin/design/[designId]")이고 /service · /api/v1 · /api 세 주소로 받는다.

import (
	"zbd.kr/be/internal/api/admincontent"
	"zbd.kr/be/internal/api/apps"
	"zbd.kr/be/internal/api/collab"
	"zbd.kr/be/internal/api/content"
	"zbd.kr/be/internal/api/core"
	"zbd.kr/be/internal/api/site"
	v1admin "zbd.kr/be/internal/api/v1/admin"
	v1me "zbd.kr/be/internal/api/v1/me"
	v1pub "zbd.kr/be/internal/api/v1/pub"
	"zbd.kr/be/internal/web"
)

func registerRoutes(r *web.Router) {
	content.Register(r)
	apps.Register(r)
	core.Register(r)
	admincontent.Register(r)
	collab.Register(r)
	site.Register(r)
	v1admin.Register(r)
	v1me.Register(r)
	v1pub.Register(r)
	r.Handle("/healthz", web.Plain, func(c *web.Ctx) { c.Send("ok") })
}
