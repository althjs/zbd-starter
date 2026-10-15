// Package me — /api/v1/me/… 로그인한 회원 본인의 목록(쪽·검색·정렬을 서버에서). 늘 토큰의 회원 것만.
//
//	GET /api/v1/me/comments?page&size&sort=-createdAt&groupKey                 (감사표 1-10)
//	GET /api/v1/me/posts?page&size&groupKey&q&sort=-createdAt                  (1-11 — 1000개 자른 뒤 거르던 버그 없음)
//	GET /api/v1/me/bookmarks?page&size                                          (1-12 — 회원 json.postIds, 최근 담은 것부터)
//	GET /api/v1/me/todos?done=0|1&page&size&sort=-ts                            (2-9)
//	GET /api/v1/me/freecell/records?result=won|gaveup|lost&page&size&sort       (2-10)
//	GET /api/v1/me/shares?page&size&sort=-updatedAt                             (2-15)
//	GET /api/v1/me/notifications?size&before                                    (3-2, 커서)
//	GET /api/v1/me/rooms?page&size&q                                            (2-12, 친구 채팅 방)
//	GET /api/v1/me/rooms/[roomId]/messages?size&before                          (3-2, 커서)
package me

import (
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/web"
)

func Register(r *web.Router) {
	r.Handle("/me/posts", web.Cors, v1x.User(posts))
	r.Handle("/me/bookmarks", web.Cors, v1x.User(bookmarks))
}
