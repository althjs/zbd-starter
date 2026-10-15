package settings

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 옛 이름 companyId — 2026-10 옮기는 기간: COMPANIES·항목 companyId 를 HOSTINGS·hostingId 로 읽는가
func TestMigrateOldNames(t *testing.T) {
	m, _ := jsx.Parse([]byte(`{"":{"COMPANIES":{"items":[{"companyId":"ZBD","name":"즈비드"}]}}}`))
	mm := m.(*jsx.OrderedMap)
	migrateOldNames(mm)
	if got := jsx.Stringify(mm); got != `{"":{"HOSTINGS":{"items":[{"hostingId":"ZBD","name":"즈비드"}]}}}` {
		t.Fatal(got)
	}
	// 새 키가 있으면 그대로
	m, _ = jsx.Parse([]byte(`{"":{"HOSTINGS":{"items":[{"hostingId":"ZBD"}]},"COMPANIES":{"items":[]}}}`))
	mm = m.(*jsx.OrderedMap)
	migrateOldNames(mm)
	if n := len(mm.Map("").Map(HostingsKey).Arr("items")); n != 1 {
		t.Fatal(jsx.Stringify(mm))
	}
}
