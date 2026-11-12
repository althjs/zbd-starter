package web

import "testing"

func TestApiTokenUsable(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		ra, ta   int64
		rh, th   string
		exp, now int64
		want     bool
	}{
		{"정상", 0, 5, 5, "ZBD", "ZBD", 0, 100, true},
		{"기간 안", 0, 5, 5, "ZBD", "ZBD", 200, 100, true},
		{"만료", 0, 5, 5, "ZBD", "ZBD", 50, 100, false},
		{"폐기", 1, 5, 5, "ZBD", "ZBD", 0, 100, false},
		{"다른 관리자", 0, 5, 6, "ZBD", "ZBD", 0, 100, false},
		{"다른 사이트", 0, 5, 5, "ZBD", "X", 0, 100, false},
	}
	for _, c := range cases {
		if got := ApiTokenUsable(c.status, c.ra, c.ta, c.rh, c.th, c.exp, c.now); got != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
}
