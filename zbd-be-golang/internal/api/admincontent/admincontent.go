// Package admincontent — 사이트 바탕·관리자 게시물·파일·웹하드·업무관리 (/design/{호스팅}/base, /admin/post, /admin/file, /admin/work …)
package admincontent

import "zbd.kr/be/internal/web"

// Register — 이 기능의 라우트를 등록한다.
func Register(r *web.Router) {
	// 사이트 바탕(사용자 화면이 처음 받는 설정 — 주소의 design 은 옛 이름)
	r.Handle("/design/[hostingId]/base", web.Cors, designBase)

	// 업무관리(work)
	r.Handle("/admin/work", web.Auth, workIndex)
	r.Handle("/admin/work/quick-update", web.Auth, workQuickUpdate)
	r.Handle("/admin/work/[workId]", web.Auth, workOne)
	r.Handle("/admin/work/[workId]/[shapeId]", web.Auth, workShape)

	// 게시물·고객센터
	r.Handle("/admin/post", web.Auth, adminPostIndex)
	r.Handle("/admin/post/[postId]", web.Auth, adminPostOne)
	r.Handle("/admin/post/group/[groupKey]", web.Auth, adminPostGroup)
	r.Handle("/admin/post/group/multi-update", web.Auth, adminPostMultiUpdate)
	r.Handle("/admin/post/new-count", web.Auth, adminPostNewCount)
	r.Handle("/admin/post/summary", web.Auth, adminPostSummary)
	r.Handle("/admin/cs/board/[groupKey]", web.Auth, csBoard)
	r.Handle("/admin/cs/board/[groupKey]/[postId]", web.Auth, csBoardPost)

	// 파일·웹하드
	r.Handle("/admin/file/[folder]", web.Auth, adminFileFolder)
	r.Handle("/admin/file/download", web.Auth, adminFileDownload, web.Options{RawBody: true})
	r.Handle("/admin/file/secure-download", web.Auth, adminFileSecureDownload)
	r.Handle("/admin/file/unlink-file", web.Auth, adminFileUnlink, web.Options{RawBody: true})
	r.Handle("/admin/file/upload", web.Auth, adminFileUpload, web.Options{RawBody: true})
	r.Handle("/admin/file-explorer/[rootFolder]", web.Auth, adminFileExplorer)
	r.Handle("/admin/file-explorer/[rootFolder]/move", web.Auth, adminFileExplorerMove)
	r.Handle("/admin/file-explorer/[rootFolder]/presence", web.Auth, adminFileExplorerPresence)
	r.Handle("/admin/file-explorer/[rootFolder]/upload", web.Auth, adminFileExplorerUpload, web.Options{RawBody: true})
}
