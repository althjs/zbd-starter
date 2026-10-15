package admincontent

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 정책: 업무관리 canvas → work 이름 바꾸기(2026-10-04) 옮기는 기간 — 옛 칸 이름도 받되, 새 이름이 같이 오면 새 이름이 이긴다
func TestLegacyWorkKeysPolicy(t *testing.T) {
	m := jsx.Obj("canvasId", float64(7), "parentCanvasId", float64(3), "canvasIds", []any{float64(1)})
	renameLegacyWorkKeys(m)
	if m.V("workId") != float64(7) || m.V("parentWorkId") != float64(3) || jsx.Stringify(m.V("workIds")) != "[1]" {
		t.Errorf("옛 이름을 새 이름으로: %s", jsx.Stringify(m))
	}
	both := jsx.Obj("canvasId", float64(7), "workId", float64(9))
	renameLegacyWorkKeys(both)
	if both.V("workId") != float64(9) {
		t.Errorf("새 이름이 이긴다: %s", jsx.Stringify(both))
	}
	renameLegacyWorkKeys(nil) // nil 이어도 멈추지 않는다
}
