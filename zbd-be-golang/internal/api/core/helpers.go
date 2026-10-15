package core

import (
	"errors"
	"strings"

	"fmt"
	"time"

	"zbd.kr/be/internal/lib/util"

	"zbd.kr/be/internal/jsx"
)

// parseJSON — JSON.parse(v || def). 실패하면 err.
func parseJSON(v any, def string) (any, error) {
	s := jsx.StrOrEmpty(v)
	if s == "" {
		s = def
	}
	return jsx.ParseString(s)
}

// durationSec — getDurationSecond: "0.12s"
func durationSec(start time.Time) string {
	return fmt.Sprintf("%.2fs", float64(time.Since(start).Milliseconds())/1000)
}

// nowMs — Date.now()
func nowMs() float64 { return float64(time.Now().UnixMilli()) }

// looseEq — JS 의 a == b (숫자·문자열 섞임)
func looseEq(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	_, as := a.(string)
	_, bs := b.(string)
	if as && bs {
		return a.(string) == b.(string)
	}
	if ab, ok := a.(bool); ok {
		a = map[bool]float64{true: 1, false: 0}[ab]
	}
	if bb, ok := b.(bool); ok {
		b = map[bool]float64{true: 1, false: 0}[bb]
	}
	return jsx.Number(a) == jsx.Number(b)
}

// rowsAny — 행 목록을 응답용 배열로(없으면 [])
func rowsAny(rows []*jsx.OrderedMap) []any {
	out := make([]any, len(rows))
	for i, r := range rows {
		out[i] = r
	}
	return out
}

// bcryptAny — 지금 서비스의 bcrypt.hash(v) — 값이 없거나 문자열이 아니면 그 라이브러리와 같은 오류
func bcryptAny(v any) (string, error) {
	if v == nil {
		return "", errors.New("data and salt arguments required")
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.New("data must be a string or Buffer and salt must either be a salt string or a number of rounds")
	}
	return util.BcryptHash(s)
}

func trimStr(s string) string { return strings.TrimSpace(s) }

type strError string

func (e strError) Error() string { return string(e) }

func errFmt(s string) error { return strError(s) }
