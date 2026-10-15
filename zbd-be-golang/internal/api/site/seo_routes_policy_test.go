package site

import "testing"

// 없는 주소 = 404 판정(routes.json — 화면 라우트 폴더 그대로)
func TestRouteKnownPolicy(t *testing.T) {
	routes := [][]string{}
	for _, r := range []string{"/", "/board/[groupKey]", "/board/[groupKey]/[postId]", "/cs/faq/[[keyword]]", "/admin/explorer/[rootFolder]/[...path]", "/user/my"} {
		routes = append(routes, routeSegs(r))
	}
	cases := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"", true},
		{"/board/notice", true},
		{"/board/notice/abc12345", true},
		{"/board/notice/abc12345/x", false},
		{"/board", false},
		{"/cs/faq", true},
		{"/cs/faq/환불", true},
		{"/admin/explorer/WEB_HDD", true},
		{"/admin/explorer/WEB_HDD/a/b/c.md", true},
		{"/user/my", true},
		{"/user/my/", true},
		{"/user/mine", false},
		{"/no-such-page-1791000000", false},
	}
	for _, c := range cases {
		if got := routeKnown(routes, c.path); got != c.want {
			t.Errorf("routeKnown(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
