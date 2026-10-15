// Package collab — 관리자 토큰 발급 기록 (/admin/system-monitor/token-log)
package collab

import "zbd.kr/be/internal/web"

// Register — 이 기능의 라우트를 등록한다.
func Register(r *web.Router) {
	r.Handle("/admin/system-monitor/token-log", web.Auth, adminTokenLog)
	r.Handle("/admin/system-monitor/menu-auth", web.Auth, adminSystemMenuAuth) // 서버 관리 화면 메뉴 권한(token_log.go)
}
