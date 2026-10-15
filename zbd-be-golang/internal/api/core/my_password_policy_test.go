package core

import "testing"

// 관리자 내 정보 비밀번호 바꾸기 — 확인한 계정 = 로그인한 본인(2026-10-06)
func TestAdminSameAdminPolicy(t *testing.T) {
	cases := []struct {
		name string
		key  any
		me   string
		want bool
	}{
		{"본인", float64(352), "352", true},
		{"본인(문자)", "352", "352", true},
		{"다른 관리자 비밀번호로 확인 — 거절", float64(340), "352", false},
		{"토큰 없음 — 거절", float64(352), "", false},
		{"번호 없음 — 거절", nil, "352", false},
	}
	for _, c := range cases {
		if got := sameAdmin(c.key, c.me); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
