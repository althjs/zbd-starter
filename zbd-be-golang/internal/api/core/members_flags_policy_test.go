package core

import (
	"reflect"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 관리자 저장 때 바꿀 수 없는 칸 — isDeveloper 는 개발자만, claudeManager 는 개발자·최고 관리자만 (members.go adminFlagKeys·keepFlags, 2026-10-05)
func TestAdminFlagsPolicy(t *testing.T) {
	cases := []struct {
		dev, super bool
		want       []string
	}{
		{true, false, nil},
		{true, true, nil},
		{false, true, []string{"isDeveloper"}},
		{false, false, []string{"isDeveloper", "claudeManager"}},
	}
	for _, c := range cases {
		if got := adminFlagKeys(c.dev, c.super); !reflect.DeepEqual(got, c.want) {
			t.Errorf("dev=%v super=%v: %v, 기대 %v", c.dev, c.super, got, c.want)
		}
	}
	// 일반 관리자가 isDeveloper·claudeManager 를 켜서 보내도 이전 값으로(없으면 뺀다), 다른 칸은 그대로
	next := jsx.ParseMap(`{"isDeveloper":true,"claudeManager":true,"nickname":"a"}`)
	prev := jsx.ParseMap(`{"claudeManager":true}`)
	out := keepFlags(next, prev, adminFlagKeys(false, false)...).(*jsx.OrderedMap)
	if out.Has("isDeveloper") || !jsx.Truthy(out.V("claudeManager")) || jsx.StrOrEmpty(out.V("nickname")) != "a" {
		t.Errorf("keepFlags: %s", jsx.Stringify(out))
	}
	// 문자열 json 은 문자열로 돌려준다
	if _, ok := keepFlags(`{"isDeveloper":true}`, nil, "isDeveloper").(string); !ok {
		t.Error("문자열 json 은 문자열로")
	}
}
