// Package site — 서버 렌더링 문서·검색엔진(SEO·사이트맵·robots) (/seo, /sitemap …)
package site

import "zbd.kr/be/internal/web"

// Register — 이 기능의 라우트를 등록한다.
func Register(r *web.Router) {
	r.Handle("/seo", web.Cors, seoHandler)
	r.Handle("/seo/robots", web.Cors, robotsHandler)
	r.Handle("/seo/parse-open-graph", web.Cors, parseOpenGraph)
	r.Handle("/sitemap/[hostingId]", web.Cors, sitemapHandler)
}
