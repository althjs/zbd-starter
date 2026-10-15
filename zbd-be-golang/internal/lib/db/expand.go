package db

import (
	"math"
	"strings"

	"zbd.kr/be/internal/jsx"
)

// expand — 지금 서비스의 mysql 드라이버(sqlstring.format) 와 같은 자리표 규칙.
//
//	?   값 하나. 배열([]any·[]int64·[]string)이면 "a, b, c" 로 펼친다(IN (?)). 배열 안의 배열은 "(a, b), (c, d)"(여러 행 INSERT).
//	    객체(*jsx.OrderedMap)면 "`k1` = v1, `k2` = v2" (SET ?).
//	??  식별자(칸·표 이름) — `이름`
//	⚠ 문자열 안의 ? 도 바꾼다 — 드라이버와 같다(쿼리에 ? 를 글자로 쓰지 말 것).
//	숫자 float64 가 정수면 int64 로(드라이버가 1e+06 처럼 쓰지 않게).
func expand(query string, args []any) (string, []any) {
	if len(args) == 0 {
		return query, args
	}
	var b strings.Builder
	out := make([]any, 0, len(args))
	ai := 0
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch != '?' {
			b.WriteByte(ch)
			continue
		}
		if ai >= len(args) {
			b.WriteByte(ch)
			continue
		}
		if i+1 < len(query) && query[i+1] == '?' { // ??
			i++
			b.WriteString(ident(jsx.Str(args[ai])))
			ai++
			continue
		}
		writeValue(&b, &out, args[ai], true)
		ai++
	}
	return b.String(), out
}

func ident(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = "`" + strings.ReplaceAll(p, "`", "``") + "`"
	}
	return strings.Join(parts, ".")
}

func writeValue(b *strings.Builder, out *[]any, v any, top bool) {
	switch x := v.(type) {
	case []any:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			if inner, ok := e.([]any); ok {
				b.WriteByte('(')
				writeValue(b, out, inner, false)
				b.WriteByte(')')
				continue
			}
			writeValue(b, out, e, false)
		}
	case []int64:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
			*out = append(*out, e)
		}
	case []string:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
			*out = append(*out, e)
		}
	case *jsx.OrderedMap:
		if x == nil {
			b.WriteString("NULL")
			return
		}
		for i, k := range x.Keys() {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(ident(k))
			b.WriteString(" = ")
			writeValue(b, out, x.V(k), false)
		}
	default:
		b.WriteByte('?')
		*out = append(*out, scalar(v))
	}
}

func scalar(v any) any {
	switch x := v.(type) {
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1<<53 {
			return int64(x)
		}
		return x
	case *jsx.OrderedMap, []any:
		return jsx.Stringify(x) // 드라이버에 객체를 넘기면 안 된다 — JSON 문자열로
	}
	return v
}
