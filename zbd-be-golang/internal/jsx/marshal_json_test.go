package jsx

import (
	"encoding/json"
	"testing"
)

// encoding/json 으로 보내도(소켓 이벤트) 칸이 그대로 — 예전엔 `{}` 로 나갔다
func TestOrderedMapEncodingJSON(t *testing.T) {
	data := any(Obj("hostingId", "ZBD", "groupKey", "TALK_BOT", "key", float64(14), "to", []any{float64(11)}, "n", Obj("a", "<b>")))
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var back map[string]any
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back["hostingId"] != "ZBD" || back["groupKey"] != "TALK_BOT" || back["key"] != float64(14) {
		t.Fatalf("칸이 빠짐: %s", b)
	}
	if n, _ := back["n"].(map[string]any); n["a"] != "<b>" {
		t.Fatalf("안쪽 객체가 빠짐: %s", b)
	}
}
