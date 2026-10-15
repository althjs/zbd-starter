// Package jsx — 지금 서비스(Node BE)와 **같은 응답**을 내기 위한 JavaScript 의미론 도우미.
//
// API 를 옮기는 동안 화면이 응답을 그대로 쓰므로, 값 해석(Number() · truthy · String())과
// JSON 직렬화(JSON.stringify)가 **글자 단위로 같아야** 한다(scripts/compare.sh 로 비교한다).
//
//	· 요청 본문은 map[string]any(json.Number) 로 받는다 — 숫자/문자열 구분을 잃지 않기 위해.
//	· 응답은 OrderedMap 으로 만든다 — Go map 은 키를 정렬해 버려 키 순서가 달라진다.
package jsx

import (
	"bytes"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ─── 값 해석 ─────────────────────────────────────

// Truthy — JS 의 `!!v`.
func Truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		return err == nil && f != 0 && !math.IsNaN(f)
	case float64:
		return x != 0 && !math.IsNaN(x)
	case int:
		return x != 0
	case int64:
		return x != 0
	default:
		return true // 객체·배열은 비어 있어도 truthy
	}
}

// Number — JS 의 `Number(v)`. 변환 불가면 NaN.
func Number(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case bool:
		if x {
			return 1
		}
		return 0
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			return math.NaN()
		}
		return f
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	case []any:
		// Number([]) = 0, Number([x]) = Number(x), 그 외 NaN
		if len(x) == 0 {
			return 0
		}
		if len(x) == 1 {
			return Number(x[0])
		}
		return math.NaN()
	default:
		return math.NaN()
	}
}

// NumOr — JS 의 `Number(v) || d`.
func NumOr(v any, d float64) float64 {
	n := Number(v)
	if n == 0 || math.IsNaN(n) {
		return d
	}
	return n
}

// Int — JS 의 `Number(v) || 0` 을 정수로(소수부 버림 없이 그대로 쓰는 곳은 쓰지 말 것).
func Int(v any) int64 {
	n := NumOr(v, 0)
	return int64(n)
}

// Str — JS 의 `String(v)`. 누락(undefined)은 호출자가 따로 구분한다(여기선 nil = "null").
func Str(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case json.Number:
		return NumStr(Number(x))
	case float64:
		return NumStr(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			if e == nil {
				parts[i] = ""
			} else {
				parts[i] = Str(e)
			}
		}
		return strings.Join(parts, ",")
	default:
		return "[object Object]"
	}
}

// StrOrEmpty — JS 의 `String(v || "")`.
func StrOrEmpty(v any) string {
	if !Truthy(v) {
		return ""
	}
	return Str(v)
}

// NumStr — JS 의 Number → String (ECMAScript Number::toString 규칙).
func NumStr(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	if f == 0 {
		return "0" // -0 도 "0"
	}
	neg := f < 0
	if neg {
		f = -f
	}
	// 가장 짧은 자릿수(digits)와 지수: f = 0.d1d2…dk × 10^n
	e := strconv.FormatFloat(f, 'e', -1, 64) // d.ddddde±xx
	mant, expStr, _ := strings.Cut(e, "e")
	digits := strings.Replace(mant, ".", "", 1)
	exp, _ := strconv.Atoi(expStr)
	k := len(digits)
	n := exp + 1
	var out string
	switch {
	case k <= n && n <= 21:
		out = digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		out = digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		out = "0." + strings.Repeat("0", -n) + digits
	default:
		sign := "+"
		if n-1 < 0 {
			sign = "-"
		}
		ex := n - 1
		if ex < 0 {
			ex = -ex
		}
		if k == 1 {
			out = digits + "e" + sign + strconv.Itoa(ex)
		} else {
			out = digits[:1] + "." + digits[1:] + "e" + sign + strconv.Itoa(ex)
		}
	}
	if neg {
		return "-" + out
	}
	return out
}

// Param — 요청 본문 값을 SQL 바인딩 값으로. 정수 json.Number 는 int64(드라이버가 1e+06 처럼 쓰지 않게).
func Param(v any) any {
	switch x := v.(type) {
	case json.Number:
		if i, err := strconv.ParseInt(string(x), 10, 64); err == nil {
			return i
		}
		f, _ := strconv.ParseFloat(string(x), 64)
		return f
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1<<53 {
			return int64(x)
		}
		return x
	default:
		return v
	}
}

// ─── 순서 보존 객체 ──────────────────────────────

// OrderedMap — 키 순서를 지키는 객체(Node 의 삽입 순서와 같게).
type OrderedMap struct {
	keys []string
	vals map[string]any
}

func NewOrderedMap() *OrderedMap { return &OrderedMap{vals: map[string]any{}} }

// MarshalJSON — encoding/json 으로 보낼 때도 Marshal 과 같은 내용(키 순서 포함).
// 없으면 칸이 비공개라 `{}` 로 나간다 — 소켓 이벤트(lib/sock 이 encoding/json 을 쓴다)가 전부 빈 값이 되어
// 데몬이 버리고 있었다(2026-10-03 Go 전환 뒤 메신저·친구 채팅 Claude·라이프 동기화 등 실시간 전부).
func (m *OrderedMap) MarshalJSON() ([]byte, error) { return Marshal(m), nil }

func (m *OrderedMap) Set(k string, v any) {
	if _, ok := m.vals[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.vals[k] = v
}

func (m *OrderedMap) Get(k string) (any, bool) { v, ok := m.vals[k]; return v, ok }

// Keys — JS 객체의 키 순서(복사본): 배열 번호 같은 키(0 ~ 2^32-2 의 정규 정수) 가 먼저 오름차순, 나머지는 넣은 순서.
func (m *OrderedMap) Keys() []string {
	if m == nil {
		return nil
	}
	return jsKeyOrder(m.keys)
}

// isArrayIndex — JS 가 정수 키로 보는 문자열인가("0", "12" — "01"·"-1"·"1.0" 은 아니다)
func isArrayIndex(k string) bool {
	if k == "" || len(k) > 10 || (len(k) > 1 && k[0] == '0') {
		return false
	}
	var n uint64
	for i := 0; i < len(k); i++ {
		c := k[i]
		if c < '0' || c > '9' {
			return false
		}
		n = n*10 + uint64(c-'0')
	}
	return n < 4294967295
}

func jsKeyOrder(keys []string) []string {
	var idx, rest []string
	for _, k := range keys {
		if isArrayIndex(k) {
			idx = append(idx, k)
		} else {
			rest = append(rest, k)
		}
	}
	if len(idx) == 0 {
		return append([]string(nil), keys...)
	}
	sort.Slice(idx, func(i, j int) bool {
		if len(idx[i]) != len(idx[j]) {
			return len(idx[i]) < len(idx[j])
		}
		return idx[i] < idx[j]
	})
	return append(idx, rest...)
}

// Raw — 이미 JSON 으로 인코딩된 값(숫자 표기를 드라이버 문자열 그대로 싣는 등).
type Raw string

// ─── JSON.stringify 호환 인코더 ──────────────────

// Marshal — JSON.stringify 와 같은 출력. (Go encoding/json 은 < > & U+2028/2029 를 escape 해 달라진다)
func Marshal(v any) []byte {
	var b bytes.Buffer
	encode(&b, v)
	return b.Bytes()
}

func encode(b *bytes.Buffer, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case Raw:
		b.WriteString(string(x))
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		writeString(b, x)
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			b.WriteString("null")
		} else {
			b.WriteString(NumStr(f))
		}
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			b.WriteString("null")
		} else {
			b.WriteString(NumStr(x))
		}
	case int:
		b.WriteString(strconv.Itoa(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case *OrderedMap:
		b.WriteByte('{')
		for i, k := range jsKeyOrder(x.keys) {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, k)
			b.WriteByte(':')
			encode(b, x.vals[k])
		}
		b.WriteByte('}')
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			encode(b, e)
		}
		b.WriteByte(']')
	case []string:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, e)
		}
		b.WriteByte(']')
	case []*OrderedMap:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			encode(b, e)
		}
		b.WriteByte(']')
	case map[string]any:
		// 순서가 필요 없는 곳만(요청 본문 되돌려 주기 등). 키는 정렬된다.
		bs, _ := json.Marshal(x)
		b.Write(bs)
	default:
		bs, _ := json.Marshal(x)
		b.Write(bs)
	}
}

// writeString — JSON.stringify 의 문자열 규칙: " \ 와 제어문자만 escape, 나머지는 그대로(한글·<>& 포함).
func writeString(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			switch c {
			case '"':
				b.WriteString(`\"`)
			case '\\':
				b.WriteString(`\\`)
			case '\b':
				b.WriteString(`\b`)
			case '\f':
				b.WriteString(`\f`)
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			default:
				if c < 0x20 {
					b.WriteString(`\u00`)
					b.WriteByte("0123456789abcdef"[c>>4])
					b.WriteByte("0123456789abcdef"[c&0xf])
				} else {
					b.WriteByte(c)
				}
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			b.WriteString(`�`)
		} else {
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	b.WriteByte('"')
}
