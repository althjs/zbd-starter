package collab

// 시스템 보안 기록(토큰 발급 기록 · 외부 공격 감지 플러그인) 보는 권한 — 개발자 · 최고 관리자(isSuperAdmin) · 1그룹

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

func TestSystemSecurityStarterPolicy(t *testing.T) {
	cases := []struct {
		js    string
		group float64
		ok    bool
	}{
		{`{"isDeveloper":false,"isSuperAdmin":true}`, 1, true}, // 첫 관리자(create-admin)
		{`{"isDeveloper":true}`, 5, true},
		{`{}`, 1, true},
		{`{}`, 5, false},
	}
	for i, c := range cases {
		if systemSecurityAllowed(jsx.ParseMap(c.js), c.group) != c.ok {
			t.Errorf("%d: 기대 %v", i, c.ok)
		}
	}
}
