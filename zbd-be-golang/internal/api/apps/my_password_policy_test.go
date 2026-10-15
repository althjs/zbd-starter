package apps

import "testing"

// 내 정보 비밀번호 바꾸기 — 기존 비밀번호 확인(2026-10-06 E2E K8: 회원은 passwordOld 없이 보내면 그냥 바뀌었다)
func TestMyPasswordChangePolicy(t *testing.T) {
	cmp := func(plain, hash string) bool { return "h:"+plain == hash }
	cases := []struct {
		name               string
		newPw, oldPw, hash string
		want               bool
	}{
		{"비밀번호를 안 바꾸면 통과", "", "", "h:old", true},
		{"기존 비밀번호가 맞으면 바꿈", "New!1234", "old", "h:old", true},
		{"기존 비밀번호 없이 새 비밀번호만 — 거절", "New!1234", "", "h:old", false},
		{"기존 비밀번호가 틀리면 거절", "New!1234", "wrong", "h:old", false},
		{"비밀번호 없는 회원(소셜 가입) — 거절", "New!1234", "x", "", false},
		{"기존만 보내도 확인한다", "", "wrong", "h:old", false},
	}
	for _, c := range cases {
		if got := myPasswordChangeOK(c.newPw, c.oldPw, c.hash, cmp); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
