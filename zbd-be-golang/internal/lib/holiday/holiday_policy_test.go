package holiday

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 공휴일 = 코드 표 + 설정 COMMON.holydays, 같은 날은 설정이 이긴다 (2026-10-16)

func TestHolidayMergePolicy(t *testing.T) {
	defer SetExtra(nil)
	if Name("2026-10-09") != "한글날" || !Is("2026-12-25") || Is("2026-10-16") {
		t.Fatal("코드 표")
	}
	SetExtra([]any{
		jsx.Obj("ymd", "2026-10-16", "name", "임시공휴일"),
		jsx.Obj("ymd", "2026-10-09", "name", "한글날(설정)"),
		jsx.Obj("ymd", "2026/10/20", "name", "모양 틀림"),
		"문자열",
	})
	if Name("2026-10-16") != "임시공휴일" || Name("2026-10-09") != "한글날(설정)" {
		t.Error("설정이 이기지 않음")
	}
	if Is("2026/10/20") {
		t.Error("모양 틀린 줄이 들어감")
	}
	b := Between("2026-10-01", "2026-10-31")
	want := []string{"2026-10-03", "2026-10-05", "2026-10-09", "2026-10-16"}
	if len(b) != len(want) {
		t.Fatalf("10월 %v", b)
	}
	for i, d := range b {
		if d.Ymd != want[i] {
			t.Errorf("%d: %s", i, d.Ymd)
		}
	}
	SetExtra(nil)
	if Is("2026-10-16") {
		t.Error("설정을 비워도 남음")
	}
}
