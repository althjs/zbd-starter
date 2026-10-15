package jsx

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

// Parse — JSON.parse 와 같은 결과(객체 키 순서 보존).
//
//	객체 → *OrderedMap · 배열 → []any · 숫자 → float64(JS 숫자) · 문자열 → string · true/false → bool · null → nil
//	같은 키가 두 번 나오면 JS 처럼 **값은 마지막 것, 자리는 처음 것**.
func Parse(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("JSON 뒤에 남은 글자가 있다")
	}
	return v, nil
}

// ParseString — 문자열 JSON. 빈 문자열·잘못된 JSON 은 오류.
func ParseString(s string) (any, error) { return Parse([]byte(s)) }

// ParseOr — 실패하면 d (JS 의 try { JSON.parse(s) } catch { d }).
func ParseOr(s string, d any) any {
	v, err := Parse([]byte(s))
	if err != nil {
		return d
	}
	return v
}

// ParseMap — 객체 JSON 을 *OrderedMap 으로. 객체가 아니거나 실패하면 빈 객체(JSON.parse(x || "{}") 자리).
func ParseMap(s string) *OrderedMap {
	if s == "" {
		return NewOrderedMap()
	}
	if m, ok := ParseOr(s, nil).(*OrderedMap); ok {
		return m
	}
	return NewOrderedMap()
}

func parseValue(dec *json.Decoder) (any, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return fromToken(dec, t)
}

func fromToken(dec *json.Decoder, t json.Token) (any, error) {
	switch x := t.(type) {
	case json.Delim:
		switch x {
		case '{':
			m := NewOrderedMap()
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k, ok := kt.(string)
				if !ok {
					return nil, errors.New("객체 키가 문자열이 아니다")
				}
				v, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				m.Set(k, v)
			}
			if _, err := dec.Token(); err != nil { // '}'
				return nil, err
			}
			return m, nil
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			if _, err := dec.Token(); err != nil { // ']'
				return nil, err
			}
			return arr, nil
		}
		return nil, errors.New("잘못된 JSON")
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case string, bool, nil:
		return x, nil
	}
	return nil, errors.New("잘못된 JSON 토큰")
}

// ─── OrderedMap 도우미 (JS 객체 다루듯) ───────────────

// Len — 키 수
func (m *OrderedMap) Len() int {
	if m == nil {
		return 0
	}
	return len(m.keys)
}

// Has — 키가 있는가 (JS 의 `k in o`)
func (m *OrderedMap) Has(k string) bool {
	if m == nil {
		return false
	}
	_, ok := m.vals[k]
	return ok
}

// V — 값(없으면 nil). nil 맵에도 안전하다.
func (m *OrderedMap) V(k string) any {
	if m == nil {
		return nil
	}
	return m.vals[k]
}

// Delete — 키 지우기(JS 의 delete o[k]).
func (m *OrderedMap) Delete(k string) {
	if m == nil {
		return
	}
	if _, ok := m.vals[k]; !ok {
		return
	}
	delete(m.vals, k)
	for i, kk := range m.keys {
		if kk == k {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

// Map — 하위 객체(없거나 객체가 아니면 nil).
func (m *OrderedMap) Map(k string) *OrderedMap {
	x, _ := m.V(k).(*OrderedMap)
	return x
}

// Arr — 하위 배열(없거나 배열이 아니면 nil).
func (m *OrderedMap) Arr(k string) []any {
	x, _ := m.V(k).([]any)
	return x
}

// S — 문자열 값(JS 의 `o[k] || ""` 가 아니라 String 변환 — 없으면 "").
func (m *OrderedMap) S(k string) string {
	v := m.V(k)
	if v == nil {
		return ""
	}
	return Str(v)
}

// N — 숫자 값(JS 의 Number(o[k]), 없으면 0).
func (m *OrderedMap) N(k string) float64 {
	v := m.V(k)
	if v == nil {
		return 0
	}
	return Number(v)
}

// Get2 — 경로 값 o?.a?.b?.c (중간이 객체가 아니면 nil).
func (m *OrderedMap) Path(keys ...string) any {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(*OrderedMap)
		if !ok || mm == nil {
			return nil
		}
		cur = mm.vals[k]
	}
	return cur
}

// Clone — 깊은 복사(JSON.parse(JSON.stringify(x)) 자리, 값 모양 그대로).
func Clone(v any) any {
	switch x := v.(type) {
	case *OrderedMap:
		if x == nil {
			return x
		}
		n := NewOrderedMap()
		for _, k := range x.keys {
			n.Set(k, Clone(x.vals[k]))
		}
		return n
	case []any:
		n := make([]any, len(x))
		for i, e := range x {
			n[i] = Clone(e)
		}
		return n
	case []*OrderedMap:
		n := make([]any, len(x))
		for i, e := range x {
			n[i] = Clone(e)
		}
		return n
	}
	return v
}

// Assign — Object.assign(dst, src) / {...dst, ...src} (dst 를 고친다, nil 이면 새로).
func Assign(dst *OrderedMap, srcs ...*OrderedMap) *OrderedMap {
	if dst == nil {
		dst = NewOrderedMap()
	}
	for _, s := range srcs {
		if s == nil {
			continue
		}
		for _, k := range s.keys {
			dst.Set(k, s.vals[k])
		}
	}
	return dst
}

// Stringify — JSON.stringify(v) 문자열.
func Stringify(v any) string { return string(Marshal(v)) }

// Obj — 순서 있는 객체를 한 줄로: Obj("code", "0000", "data", x)
func Obj(kv ...any) *OrderedMap {
	m := NewOrderedMap()
	for i := 0; i+1 < len(kv); i += 2 {
		m.Set(kv[i].(string), kv[i+1])
	}
	return m
}

// RenameKey — 키 이름만 바꾼다(자리·값 그대로). 새 키가 이미 있거나 옛 키가 없으면 아무것도 안 한다.
func (m *OrderedMap) RenameKey(oldK, newK string) {
	if m == nil || oldK == newK {
		return
	}
	v, ok := m.vals[oldK]
	if !ok {
		return
	}
	if _, dup := m.vals[newK]; dup {
		return
	}
	delete(m.vals, oldK)
	m.vals[newK] = v
	for i, k := range m.keys {
		if k == oldK {
			m.keys[i] = newK
			break
		}
	}
}
