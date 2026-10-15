// Package apps — 회원 가입·탈퇴·내 정보·내 댓글·로그인 기기·회원 파일 (/user/member, /user/my, /user/file …)
package apps

import "zbd.kr/be/internal/web"

// Register — 이 기능의 라우트를 등록한다.
func Register(r *web.Router) {
	r.Handle("/user/my/settings", web.Auth, mySettings)
	r.Handle("/user/my/device", web.Cors, myDevices)
	r.Handle("/user/my/device/[browserUid]", web.Cors, myDeviceDelete)
	r.Handle("/user/member/join", web.Cors, memberJoin)
	r.Handle("/user/member/leave", web.Cors, memberLeave)
	r.Handle("/user/my/account", web.Auth, myAccount)
	r.Handle("/user/file/[folder]", web.Auth, userFolder)
	r.Handle("/user/file/download", web.Cors, userDownload, web.Options{RawBody: true})
	r.Handle("/user/file/upload", web.Auth, userUpload, web.Options{RawBody: true, BodyLimit: 1 << 40})
	r.Handle("/user/file/secure-upload", web.Cors, secureUpload, web.Options{RawBody: true, BodyLimit: 1 << 40})
	r.Handle("/user/file/secure-download", web.Cors, secureDownload)
	r.Handle("/user/file/secure-move", web.Auth, secureMove)
}
