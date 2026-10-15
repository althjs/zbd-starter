package site

// 이 패키지 안에서만 쓰는 작은 도우미 — 지금 서비스(Node)의 값 모양·문자열 규칙을 맞추기 위한 것들.

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

// asMap — 객체면 그것(아니면 nil). OrderedMap 의 읽기 함수는 nil 에도 안전하다.
func asMap(v any) *jsx.OrderedMap {
	m, _ := v.(*jsx.OrderedMap)
	return m
}

// arrOf — Array.isArray(v) ? v : []
func arrOf(v any) []any {
	if a, ok := v.([]any); ok {
		return a
	}
	return []any{}
}

// tmpl — 템플릿 문자열 `${v ?? ""}` (nil → "")
func tmpl(v any) string {
	if v == nil {
		return ""
	}
	return jsx.Str(v)
}

// orStr — (v || "") 를 문자열로
func orStr(v any) string { return jsx.StrOrEmpty(v) }

// notFalse — v !== false
func notFalse(v any) bool {
	b, ok := v.(bool)
	return !(ok && !b)
}

func isNum(v any) bool {
	switch v.(type) {
	case float64, int, int64, json.Number:
		return true
	}
	return false
}

// looseEq — JS 의 a == b (원시값만)
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
	return false
}

// jsLen — s.length (UTF-16 단위)
func jsLen(s string) int { return len(utf16.Encode([]rune(s))) }

// jsSlice — s.slice(a, b) / substring (UTF-16 단위)
func jsSlice(s string, a, b int) string {
	u := utf16.Encode([]rune(s))
	if a < 0 {
		a = 0
	}
	if b > len(u) {
		b = len(u)
	}
	if a >= b {
		return ""
	}
	if a == 0 && b == len(u) {
		return s
	}
	return string(utf16.Decode(u[a:b]))
}

// jsSub — s.substring(0, n)
func jsSub(s string, n int) string { return jsSlice(s, 0, n) }

// jsTrim — String.prototype.trim
func jsTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return unicode.IsSpace(r) || r == 0xfeff })
}

// JS 정규식의 \s 와 같은 공백
const jsSpace = `\t\n\v\f\r \x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}`

var reSpaces = regexp.MustCompile(`[` + jsSpace + `]+`)

// collapse — .replace(/\s+/g, " ")
func collapse(s string) string { return reSpaces.ReplaceAllString(s, " ") }

// escAttr — 따옴표·꺾쇠·& (seo.utils escHtml: ' 까지)
func escAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", `"`, "&quot;", "'", "&#39;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// ltEscape — JSON.stringify(...).replace(/</g, "\\u003c")
func ltEscape(s string) string { return strings.ReplaceAll(s, "<", `<`) }

// encodeURIComponent — JS 와 같은 규칙(A-Z a-z 0-9 - _ . ! ~ * ' ( ) 만 그대로)
func encodeURIComponent(s string) string {
	const keep = "-_.!~*'()"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || strings.IndexByte(keep, ch) >= 0 {
			b.WriteByte(ch)
			continue
		}
		b.WriteString("%" + strings.ToUpper(strconv.FormatInt(int64(ch)>>4, 16)) + strings.ToUpper(strconv.FormatInt(int64(ch)&15, 16)))
	}
	return b.String()
}

// isoMs — new Date(ms).toISOString() (NaN 이면 ok=false — JS 는 RangeError)
func isoMs(ms float64) (string, bool) {
	if math.IsNaN(ms) || math.IsInf(ms, 0) {
		return "", false
	}
	return time.UnixMilli(int64(ms)).UTC().Format("2006-01-02T15:04:05.000Z"), true
}

// jsDateMs — new Date(v).getTime() (숫자 ms · 숫자 문자열 아님 · ISO · DB 날짜 문자열). 못 읽으면 NaN.
func jsDateMs(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case float64:
		return x
	case json.Number:
		return jsx.Number(x)
	case bool:
		return math.NaN()
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return math.NaN()
		}
		for _, l := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.000Z", "2006-01-02", "2006-01", "2006"} {
			if t, err := time.Parse(l, s); err == nil {
				return float64(t.UnixMilli())
			}
		}
		for _, l := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04"} {
			if t, err := time.ParseInLocation(l, s, db.KST); err == nil {
				return float64(t.UnixMilli())
			}
		}
	}
	return math.NaN()
}

// parseJSONStr — JSON.parse(v) (문자열이 아니면 String(v))
func parseJSONStr(v any) (any, error) {
	s, ok := v.(string)
	if !ok {
		if v == nil {
			return nil, nil
		}
		s = jsx.Str(v)
	}
	return jsx.ParseString(s)
}

// trimRightSlash — .replace(/\/+$/, "")
func trimRightSlash(s string) string { return strings.TrimRight(s, "/") }

// errText — 오류 문장
type errText string

func (e errText) Error() string { return string(e) }
