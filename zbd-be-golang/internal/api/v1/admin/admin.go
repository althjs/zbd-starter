// Package admin — /api/v1 관리자 목록(쪽·검색·정렬·필터를 서버에서). 감사표 ~/zbd-port/LIST_AUDIT.md.
//
// 권한은 같은 데이터를 주던 원래 API 의 메뉴 권한으로 본다(v1x.Admin) — 새 경로는 메뉴 API 연결이 없다.
package admin

import (
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/web"
)

func Register(r *web.Router) {
	r.Handle("/admin/members", web.Cors, v1x.Admin(memberOrig, members))
	r.Handle("/admin/members/ids", web.Cors, v1x.Admin(memberOrig, memberIDs))
	r.Handle("/admin/members/[userId]/posts", web.Cors, v1x.Admin(memberOrig, memberPosts))
	r.Handle("/admin/posts", web.Cors, v1x.Admin("/admin/post/group/[groupKey]", posts))
	r.Handle("/admin/work-items/[groupKey]", web.Cors, v1x.Admin("/admin/work", workItems))
	r.Handle("/admin/token-logs", web.Cors, v1x.Admin("/admin/system-monitor/token-log", tokenLogs))
	r.Handle("/admin/files/[folder]", web.Cors, v1x.Admin("/admin/file/[folder]", files))
	r.Handle("/admin/login-devices", web.Cors, v1x.Admin("/admin/member/login-devices", loginDevices))
}
