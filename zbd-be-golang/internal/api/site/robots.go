package site

// GET /seo/robots — robots.txt(원본: pages/api/seo/robots.ts). 내용은 관리자 「robots.txt 설정」(SITE design.robots), 없으면 전부 막는다.

import (
	"zbd.kr/be/internal/web"
)

func robotsHandler(c *web.Ctx) {
	cid := c.Q("hostingId")
	if cid == "" {
		cid = getHostingID(c.App, c.R.Host)
	}
	if cid == "" {
		sendText(c, "hostingId NOT FOUND")
		return
	}
	// text/plain 이어야 한다(구글 규격) — 예전엔 text/html 로 나갔다
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	body := orStr(c.App.Settings.Config(cid).Path("SITE", "design", "robots"))
	if body == "" {
		body = "User-agent: *\nDisallow: /"
	}
	c.Send(body)
}
