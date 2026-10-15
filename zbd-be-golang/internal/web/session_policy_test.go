package web

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 로그인 유지 기간 — 회원 userSessionDays · 관리자 adminSessionDays(SITE 설정), 없거나 0 이하 30일, 최대 400일(2026-10-04)
func TestSessionDaysPolicy(t *testing.T) {
	site := func(s string) *jsx.OrderedMap {
		m, _ := jsx.Parse([]byte(s))
		return m.(*jsx.OrderedMap)
	}
	cases := []struct {
		name  string
		site  *jsx.OrderedMap
		admin bool
		want  int
	}{
		{"회원 설정값", site(`{"userSessionDays":3,"adminSessionDays":5}`), false, 3},
		{"관리자 설정값", site(`{"userSessionDays":3,"adminSessionDays":5}`), true, 5},
		{"관리자는 회원 값을 쓰지 않는다", site(`{"userSessionDays":3}`), true, 30},
		{"회원은 관리자 값을 쓰지 않는다", site(`{"adminSessionDays":5}`), false, 30},
		{"없으면 30일", site(`{}`), false, 30},
		{"사이트 설정이 없어도 30일", nil, true, 30},
		{"0 이하면 30일", site(`{"userSessionDays":0,"adminSessionDays":-2}`), false, 30},
		{"400일 넘으면 400", site(`{"adminSessionDays":9999}`), true, 400},
		{"문자열 숫자도", site(`{"userSessionDays":"7"}`), false, 7},
	}
	for _, c := range cases {
		if got := sessionDaysFrom(c.site, c.admin); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

// 정책: 세션 쿠키 범위 — 대표 도메인 아래면 그 도메인, 밖이어도 관리자 주소면 그 주소 전용, localhost·사용자 주소는 쿠키 없음(2026-10-06)
func TestCookieScopePolicy(t *testing.T) {
	cases := []struct {
		name, host, domain string
		admin              bool
		wantD              string
		wantOK             bool
	}{
		{"대표 도메인 아래 관리자 주소", "backyard.zbd.kr", "zbd.kr", true, "zbd.kr", true},
		{"대표 도메인 아래 사용자 주소", "zbd.kr", "zbd.kr", false, "zbd.kr", true},
		{"도메인 밖 관리자 주소 = 그 주소 전용", "admin-starter-test.zbd.kr", "", true, "", true},
		{"도메인 밖 사용자 주소 = 쿠키 없음", "other.example.com", "", false, "", false},
		{"localhost 관리자 = 쿠키 없음(헤더 토큰)", "localhost", "", true, "", false},
		{"*.localhost = 쿠키 없음", "chat.localhost", "", true, "", false},
		{"127.0.0.1 = 쿠키 없음", "127.0.0.1", "", true, "", false},
	}
	for _, c := range cases {
		d, ok := cookieScope(c.host, c.domain, c.admin)
		if d != c.wantD || ok != c.wantOK {
			t.Errorf("%s: got (%q,%v) want (%q,%v)", c.name, d, ok, c.wantD, c.wantOK)
		}
	}
}

// 정책: 같은 이름 세션 쿠키가 여럿이면(상위 도메인의 다른 사이트 쿠키) 이 사이트 키로 풀리는 첫 값(2026-10-06)
func TestPickSessionPolicy(t *testing.T) {
	valid := func(s string) bool { return s == "mine" }
	cases := []struct {
		name string
		vals []string
		want string
	}{
		{"남의 것이 앞에 와도 내 것", []string{"other", "mine"}, "mine"},
		{"내 것만", []string{"mine"}, "mine"},
		{"풀리는 것이 없으면 없음", []string{"other"}, ""},
		{"쿠키 없음", nil, ""},
	}
	for _, c := range cases {
		if got := pickSession(c.vals, valid); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}
