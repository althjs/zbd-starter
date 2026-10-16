package adminauth

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 세부 권한 — 개발자·최고 관리자는 늘, 그 밖은 자기 그룹의 json.perms 에 있을 때만. 계정 없음·빈 키는 거절
func TestHasPermPolicy(t *testing.T) {
	m := func(s string) *jsx.OrderedMap { return jsx.ParseMap(s) }
	roles := map[string]map[string]bool{
		"59": RolePermsOf(m(`{"perms":["attendance.ip"," x ", 3, ""]}`)),
		"60": RolePermsOf(m(`{"memos":[]}`)),
	}
	cases := []struct {
		name   string
		member *jsx.OrderedMap
		key    string
		want   bool
	}{
		{"그룹에 있음", m(`{"groupId":59}`), "attendance.ip", true},
		{"공백 정리", m(`{"groupId":59}`), "x", true},
		{"그룹에 없음", m(`{"groupId":60}`), "attendance.ip", false},
		{"다른 권한", m(`{"groupId":59}`), "approval.form", false},
		{"개발자는 늘", m(`{"groupId":60,"isDeveloper":true}`), "attendance.ip", true},
		{"최고 관리자는 늘", m(`{"groupId":60,"isSuperAdmin":true}`), "attendance.ip", true},
		{"계정 없음", nil, "attendance.ip", false},
		{"빈 키", m(`{"groupId":59,"isDeveloper":true}`), "", false},
		{"그룹 없음", m(`{"groupId":99}`), "attendance.ip", false},
	}
	for _, c := range cases {
		if got := HasPermBy(c.member, roles, c.key); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
	if len(RolePermsOf(nil)) != 0 {
		t.Error("nil json")
	}
}
