package web

import (
	"net/http/httptest"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 옛 이름 companyId — 2026-10 옮기는 기간: 바깥 입력의 companyId 를 hostingId 로 읽는가
func TestRenameOldKeys(t *testing.T) {
	q := buildQuery(httptest.NewRequest("GET", "/x?companyId=ZBD&a=1", nil), nil)
	if q.S("hostingId") != "ZBD" || q.Has("companyId") || jsx.Stringify(q) != `{"hostingId":"ZBD","a":"1"}` {
		t.Fatalf("query: %s", jsx.Stringify(q))
	}
	// 경로 변수가 이긴다
	q = buildQuery(httptest.NewRequest("GET", "/x?companyId=OLD", nil), map[string]any{"hostingId": "ZBD"})
	if q.S("hostingId") != "ZBD" {
		t.Fatalf("param: %s", jsx.Stringify(q))
	}
	// 둘 다 오면 새 이름
	q = buildQuery(httptest.NewRequest("GET", "/x?companyId=OLD&hostingId=ZBD", nil), nil)
	if jsx.Stringify(q) != `{"hostingId":"ZBD"}` {
		t.Fatalf("both: %s", jsx.Stringify(q))
	}
	b, _ := jsx.Parse([]byte(`{"companyId":"ZBD","post":{"companyId":"ZBD","x":1},"items":[{"companyId":"A","deep":{"companyId":"keep"}}]}`))
	renameOldKeys(b, 0)
	want := `{"hostingId":"ZBD","post":{"hostingId":"ZBD","x":1},"items":[{"hostingId":"A","deep":{"companyId":"keep"}}]}`
	if got := jsx.Stringify(b); got != want {
		t.Fatalf("body:\n got %s\nwant %s", got, want)
	}
}
