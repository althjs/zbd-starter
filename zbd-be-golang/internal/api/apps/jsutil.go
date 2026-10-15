package apps

// JS 의 문자열·비교 규칙을 그대로 흉내 내는 작은 도우미(응답이 글자 단위로 같아야 해서)

import (
	"encoding/json"
	"strings"
	"unicode"

	"zbd.kr/be/internal/jsx"
)

// tmpl — 템플릿 문자열 `${v ?? ""}` (nil → "")
func tmpl(v any) string {
	if v == nil {
		return ""
	}
	return jsx.Str(v)
}

// jsTrim — String.prototype.trim (BOM 포함)
func jsTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return unicode.IsSpace(r) || r == 0xfeff })
}

func isNum(v any) bool {
	switch v.(type) {
	case float64, int, int64, json.Number:
		return true
	}
	return false
}

// looseEq — JS 의 a == b (원시값만. 객체끼리는 false)
func looseEq(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			return sa == sb
		}
	}
	if ba, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			return ba == bb
		}
	}
	prim := func(v any) bool {
		switch v.(type) {
		case string, bool:
			return true
		}
		return isNum(v)
	}
	if prim(a) && prim(b) {
		return jsx.Number(a) == jsx.Number(b)
	}
	if sa, ok := a.([]any); ok && prim(b) { // [x] == "x"
		return looseEq(jsx.Str(sa), b)
	}
	if sb, ok := b.([]any); ok && prim(a) {
		return looseEq(a, jsx.Str(sb))
	}
	return false
}

// parseJSONObj — JSON.parse(s || "{}") 가 객체면 그 객체, 실패·객체 아님이면 빈 객체
func parseJSONObj(s any) *jsx.OrderedMap {
	str := jsx.StrOrEmpty(s)
	if str == "" {
		return jsx.NewOrderedMap()
	}
	if m := jsx.ParseMap(str); m != nil {
		return m
	}
	return jsx.NewOrderedMap()
}
