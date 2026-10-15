package adminauth

import "testing"

// 관리자 API 권한 — 메뉴 ↔ API 묶음 판정(E2E K5: 웹하드만 가진 관리자가 관리자 계정을 만들 수 있었다)
func TestAllowedByPolicy(t *testing.T) {
	binds := map[string][]string{
		"W:/service/admin/member":             {"32"},
		"/service/admin/member/login-devices": {"34"},
		"/service/admin/group":                {"33"},
		"/service/admin/user":                 {"31"},
		"/service/admin/file-explorer":        {"11"},
		"/service/admin/meta":                 {"41", "43"},
	}
	webhardOnly := func(m string) string { return map[string]string{"11": "RW"}[m] }
	memberR := func(m string) string { return map[string]string{"32": "R", "11": "RW"}[m] }
	admin := func(m string) string {
		return map[string]string{"32": "RW", "34": "R", "33": "RW", "31": "RW", "43": "R"}[m]
	}
	cases := []struct {
		name   string
		auth   func(string) string
		method string
		route  string
		want   bool
	}{
		{"묶이지 않은 API 는 통과", webhardOnly, "GET", "/service/admin/version", true},
		{"자기 메뉴 API", webhardOnly, "PUT", "/service/admin/file-explorer/[rootFolder]", true},
		{"관리자 목록 읽기는 누구나(W: 는 쓰기만)", webhardOnly, "GET", "/service/admin/member", true},
		{"관리자 만들기는 관리자 회원 RW 만", webhardOnly, "POST", "/service/admin/member", false},
		{"관리자 회원 R 은 만들기 불가", memberR, "POST", "/service/admin/member", false},
		{"관리자 회원 RW 는 만들기", admin, "POST", "/service/admin/member", true},
		{"관리자 고치기(하위 경로)도 쓰기 묶음", webhardOnly, "PUT", "/service/admin/member/[adminId]", false},
		{"더 긴 묶음이 이긴다 — 로그인 기기 읽기는 34", webhardOnly, "GET", "/service/admin/member/login-devices", false},
		{"더 긴 묶음이 이긴다 — 34 R 이면 읽기 됨", admin, "GET", "/service/admin/member/login-devices", true},
		{"권한 그룹 읽기도 막힘", webhardOnly, "GET", "/service/admin/group", false},
		{"권한 올리기(그룹 메뉴 권한) 막힘", webhardOnly, "PUT", "/service/admin/group/auth/[groupId]", false},
		{"사용자 회원 목록 막힘", webhardOnly, "GET", "/service/admin/user", false},
		{"설정 — 메뉴 둘 중 하나 R 이면 읽기", admin, "GET", "/service/admin/meta/[metaKey]", true},
		{"설정 — R 만이면 쓰기 불가", admin, "PUT", "/service/admin/meta/[metaKey]", false},
		{"설정 읽기도 막힘(비밀값이 들 수 있다)", webhardOnly, "GET", "/service/admin/meta/[metaKey]", false},
	}
	// CanRead — W: 묶음도 읽기 권한으로
	if allowedBy(binds, webhardOnly, "GET", "/service/admin/member", true) {
		t.Error("CanRead: 웹하드만 가진 사람이 관리자 회원 읽기 권한으로 잡힘")
	}
	if !allowedBy(binds, memberR, "GET", "/service/admin/member", true) {
		t.Error("CanRead: 관리자 회원 R 인 사람이 읽기 권한으로 안 잡힘")
	}
	for _, c := range cases {
		if got := AllowedBy(binds, c.auth, c.method, c.route); got != c.want {
			t.Errorf("%s: %s %s = %v, want %v", c.name, c.method, c.route, got, c.want)
		}
	}
}
