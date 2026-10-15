package web

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

func TestSiteOpenPolicy(t *testing.T) {
	env := func(v string) func(string) string { return func(string) string { return v } }
	if IsHostedSite(env("")) || IsHostedSite(env("  ")) || !IsHostedSite(env("1.zbd.co.kr")) {
		t.Fatal("호스팅 판정")
	}
	if !SiteOpen(nil) || !SiteOpen(jsx.Obj()) || !SiteOpen(jsx.Obj("siteClosed", false)) || SiteOpen(jsx.Obj("siteClosed", true)) {
		t.Fatal("값 없음·false = 공개, true = 닫힘")
	}
	cases := []struct{ hosted, open, admin, want bool }{
		{false, false, false, false}, // 호스팅 아닌 스타터는 설정이 있어도 무시
		{true, true, false, false},
		{true, false, false, true},
		{true, false, true, false}, // 관리자 주소는 늘 열림
	}
	for _, k := range cases {
		if got := ShowClosedPage(k.hosted, k.open, k.admin); got != k.want {
			t.Errorf("%+v → %v", k, got)
		}
	}
}
