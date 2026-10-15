package web

import (
	"strings"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 도메인 정책(2026-10-06) — 사용자·관리자 주소는 SITE 설정 domains 한 곳, 나머지는 계산
func TestNormalizeHostPolicy(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		ok       bool
	}{
		{"Example.com", "example.com", true},
		{"https://admin.example.com/", "admin.example.com", true},
		{"localhost:5000", "localhost:5000", true},
		{"example.com/path", "", false},
		{"exa mple.com", "", false},
		{"", "", false},
		{"-bad.com", "", false},
	} {
		got, ok := NormalizeHost(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("NormalizeHost(%q) = %q,%v want %q,%v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestDomainsDerivedPolicy(t *testing.T) {
	d := Domains{User: []string{"starter.qa.zbd.kr", "www.starter.qa.zbd.kr"}, Admin: []string{"admin.starter.qa.zbd.kr"}}
	if d.Homepage() != "https://starter.qa.zbd.kr/" || d.CookieHost() != ".starter.qa.zbd.kr" {
		t.Errorf("운영 주소 계산: %q %q", d.Homepage(), d.CookieHost())
	}
	if a := d.AdminOnly(); len(a) != 1 || a[0] != "admin.starter.qa.zbd.kr" {
		t.Errorf("관리자 전용: %v", a)
	}
	// 개발 = 한 주소 — 관리자 전용 주소 없음(쿠키 없이 헤더 토큰), 홈은 http, 쿠키 범위 없음
	dev := Domains{User: []string{"localhost:5000"}, Admin: []string{"localhost:5000"}}
	if dev.Homepage() != "http://localhost:5000/" || dev.CookieHost() != "" || len(dev.AdminOnly()) != 0 {
		t.Errorf("개발 주소 계산: %q %q %v", dev.Homepage(), dev.CookieHost(), dev.AdminOnly())
	}
}

func TestValidateDomainsPolicy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		d       Domains
		current string
		bad     string // 오류 문구에 들어갈 말(비면 통과)
	}{
		{"정상", Domains{[]string{"a.com"}, []string{"admin.a.com"}}, "admin.a.com", ""},
		{"사용자 없음", Domains{nil, []string{"admin.a.com"}}, "admin.a.com", "사용자 주소"},
		{"관리자 없음", Domains{[]string{"a.com"}, nil}, "", "관리자 주소를"},
		{"같은 주소 운영", Domains{[]string{"a.com"}, []string{"a.com"}}, "a.com", "함께 쓸 수 없습니다"},
		{"같은 주소 개발", Domains{[]string{"localhost:5000"}, []string{"localhost:5000"}}, "localhost:5000", ""},
		{"겹침", Domains{[]string{"a.com", "A.com"}, []string{"admin.a.com"}}, "admin.a.com", "겹칩니다"},
		{"잠김", Domains{[]string{"a.com"}, []string{"admin2.a.com"}}, "admin.a.com", "못 들어옵니다"},
		{"개발 주소에서는 잠김 검사 안 함", Domains{[]string{"a.com"}, []string{"admin.a.com"}}, "127.0.0.1:5001", ""},
		{"포트만 다름", Domains{[]string{"a.com"}, []string{"admin.a.com"}}, "admin.a.com:443", ""},
	} {
		err := ValidateDomains(tc.d, tc.current)
		if tc.bad == "" && err != nil {
			t.Errorf("%s: 통과해야 하는데 %v", tc.name, err)
		}
		if tc.bad != "" && (err == nil || !strings.Contains(err.Error(), tc.bad)) {
			t.Errorf("%s: %q 오류여야 하는데 %v", tc.name, tc.bad, err)
		}
	}
}

func TestSiteDomainsFallbackPolicy(t *testing.T) {
	// domains 가 있으면 그것만
	s := jsx.ParseMap(`{"homepage":"https://old.com/","domains":{"user":["new.com"],"admin":["admin.new.com"]}}`)
	if d := SiteDomains(s, "env-admin.com"); d.User[0] != "new.com" || d.Admin[0] != "admin.new.com" || len(d.Admin) != 1 {
		t.Errorf("domains 우선: %+v", d)
	}
	// 옛 설정 — 홈 주소 + .env ADMIN_HOSTS
	o := jsx.ParseMap(`{"homepage":"https://old.com/"}`)
	if d := SiteDomains(o, "admin.old.com, chat.old.com"); d.User[0] != "old.com" || len(d.Admin) != 2 {
		t.Errorf("옛 설정: %+v", d)
	}
	// 홈 주소가 design 안에만 있는 옛 설정
	if d := SiteDomains(jsx.ParseMap(`{"design":{"homepage":"http://localhost:5000/"}}`), ""); len(d.User) != 1 || d.User[0] != "localhost:5000" || len(d.Admin) != 1 || d.Admin[0] != "localhost:5000" {
		t.Errorf("design.homepage: %+v", d)
	}
}

// 개발 주소 — localhost 와 같은 네트워크 사설 IP(휴대폰·다른 PC 에서 개발 화면 열기)는 개발, 공인 IP·도메인은 아님
func TestIsLocalHostPolicy(t *testing.T) {
	for h, want := range map[string]bool{
		"localhost:5000": true, "admin.localhost": true, "127.0.0.1:5000": true,
		"192.168.0.19:5000": true, "10.1.2.3": true, "172.20.0.5:5000": true,
		"172.32.0.1": false, "8.8.8.8": false, "example.com": false, "localhost.example.com": false,
	} {
		if got := IsLocalHost(h); got != want {
			t.Errorf("IsLocalHost(%q) = %v, want %v", h, got, want)
		}
	}
}
