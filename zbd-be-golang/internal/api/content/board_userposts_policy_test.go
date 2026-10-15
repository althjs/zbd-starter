package content

import (
	"reflect"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 회원 번호로 글을 찾을 때(관리자 회원 상세 ?userId=) 대상은 게시판으로 등록된 그룹만 — 관리자·사용자 게시판, 고객 문의 CS 는 뺀다(2026-10-04)
func TestBoardGroupKeysPolicy(t *testing.T) {
	g := func(k string) any { return jsx.Obj("groupKey", k) }
	cases := []struct {
		name        string
		admin, user []any
		want        []any
	}{
		{"관리자·사용자 게시판 모두", []any{g("NOTICE"), g("TERMS")}, []any{g("FREE")}, []any{"NOTICE", "TERMS", "FREE"}},
		{"CS 는 뺀다", []any{g("NOTICE")}, []any{g("CS"), g("FREE")}, []any{"NOTICE", "FREE"}},
		{"빈 키·이상한 값은 뺀다", []any{g(""), "x", nil}, nil, []any{}},
	}
	for _, c := range cases {
		if got := boardGroupKeys(c.admin, c.user); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
