package core

import (
	"testing"
	"time"
)

// 플러그인 라이선스 확인 — 다시 묻는 간격(성공 = 즈비드가 준 분, 30분~1일 · 실패 = 30분) · 답 해석(실패 답은 믿지 않음 · 문제 20개·글 300자까지)
func TestPluginLicNext(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		ok   bool
		min  int
		want time.Duration
	}{
		{true, 360, 6 * time.Hour}, {true, 0, 6 * time.Hour}, {true, 5, 30 * time.Minute}, {true, 99999, 24 * time.Hour}, {false, 360, 30 * time.Minute},
	} {
		if got := pluginLicNext(now, c.ok, c.min).Sub(now); got != c.want {
			t.Errorf("ok=%v min=%d → %v, 기대 %v", c.ok, c.min, got, c.want)
		}
	}
}

func TestPluginLicParse(t *testing.T) {
	ps, m, ok := pluginLicParse([]byte(`{"code":"0000","data":{"ok":false,"problems":[{"code":"unlicensed","plugin":"webmail","message":"없음"},{"code":"","message":"빈 코드"}],"checkMin":120}}`))
	if !ok || m != 120 || len(ps) != 1 || ps[0].Plugin != "webmail" {
		t.Fatalf("해석 %v %d %v", ps, m, ok)
	}
	if ps, _, ok := pluginLicParse([]byte(`{"code":"0000","data":{"ok":true,"problems":[]}}`)); !ok || len(ps) != 0 {
		t.Errorf("정상 답 %v %v", ps, ok)
	}
	for _, bad := range []string{`{"code":"9999","message":"x"}`, `<html>`, ``} {
		if _, _, ok := pluginLicParse([]byte(bad)); ok {
			t.Errorf("%q 를 믿음", bad)
		}
	}
}
