package site

import (
	"html"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// sendClosedPage — 사이트 비공개일 때 사용자 주소의 응답(2026-10-08 대표 — 즈비드 호스팅 「사이트 공개」 스위치)
func sendClosedPage(c *web.Ctx, site *jsx.OrderedMap, p string) {
	switch p {
	case "/robots.txt":
		c.SetHeader("Content-Type", "text/plain; charset=utf-8")
		c.Send("User-agent: *\nDisallow: /")
		return
	case "/sitemap.xml":
		c.Status(404).End()
		return
	}
	name := ""
	if site != nil {
		name = orStr(site.Path("design", "title"))
		if name == "" {
			name = orStr(site.V("title"))
		}
	}
	title := "준비 중입니다"
	if name != "" {
		title = name + " — " + title
	}
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.SetHeader("Cache-Control", "no-store")
	c.SetHeader("X-Robots-Tag", "noindex")
	c.Status(503)
	c.Send(`<!doctype html><html lang="ko"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="robots" content="noindex"><title>` +
		html.EscapeString(title) + `</title><style>body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;font-family:-apple-system,BlinkMacSystemFont,"Apple SD Gothic Neo","Malgun Gothic",sans-serif;background:#f6f7f9;color:#333}main{text-align:center;padding:24px}h1{font-size:22px;margin:0 0 10px}p{margin:0;color:#777;font-size:15px}</style></head><body><main><h1>` +
		html.EscapeString(siteNameOr(name)) + `</h1><p>지금은 준비 중이에요. 곧 열게요.</p></main></body></html>`)
}

func siteNameOr(s string) string {
	if s == "" {
		return "사이트"
	}
	return s
}
