// Package content — 사용자 글·게시판·댓글·약관·회원 요약 (/board, /post, /user/post, /user/profile …)
package content

import "zbd.kr/be/internal/web"

// Register — 이 기능의 라우트를 등록한다.
func Register(r *web.Router) {
	r.Handle("/user/terms", web.Cors, userTerms)
	r.Handle("/user/[hostingId]/check-duplicated", web.Cors, userCheckDuplicated)
	r.Handle("/user/[hostingId]/summary", web.Auth, userSummary)
	r.Handle("/user/bookmark", web.Cors, userBookmark)
	r.Handle("/user/profile", web.Cors, userProfile)
	r.Handle("/board/[hostingId]", web.Cors, boardList)
	r.Handle("/board/[hostingId]/[postId]", web.Cors, boardPost)
	r.Handle("/user/post", web.Auth, userPost)
	r.Handle("/user/post/[postId]", web.Auth, userPostByID)
	r.Handle("/user/post/list", web.Cors, userPostList)
	r.Handle("/user/post/detail", web.Cors, userPostDetail)
	r.Handle("/user/post/guest/[postId]", web.Cors, userPostGuest)
}
