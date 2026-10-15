package pub

import (
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/web"
)

var storeSort = map[string]string{"updatedAt": "p.updatedAt", "createdAt": "p.createdAt", "title": "p.title"}

const reviewScore = "CAST(JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.score')) AS DECIMAL(4,2))"
const reviewText = "IFNULL(JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.text')), '')"

var reviewSort = map[string]string{
	"updatedAt": "updatedAt",
	"score":     reviewScore,
	"hasText":   "IF(" + reviewText + " <> '', 1, 0)",
}

// cursorParams — 커서 목록의 size·before(글 댓글 등 — 고급 게시판 플러그인이 쓴다)
func cursorParams(c *web.Ctx) (size, before int64, ok bool) {
	size, ok = v1x.IntParam(c, "size", web.ListDefaultSize)
	if !ok {
		return
	}
	if size < 1 {
		size = web.ListDefaultSize
	}
	if size > web.ListMaxSize {
		size = web.ListMaxSize
	}
	before, ok = v1x.IntParam(c, "before", 0)
	return
}
