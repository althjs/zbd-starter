package apps

// 이 패키지 안에서 여러 라우트가 같이 쓰는 작은 도우미.

import (
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// undef — JS 의 undefined 자리(객체 칸으로 넣으면 그 칸을 뺀다 — objU 참고)
type undefT struct{}

var undef = undefT{}

// objU — jsx.Obj 와 같되 값이 undef 인 칸은 넣지 않는다(JSON.stringify 가 undefined 칸을 빼는 것과 같게)
func objU(kv ...any) *jsx.OrderedMap {
	m := jsx.NewOrderedMap()
	for i := 0; i+1 < len(kv); i += 2 {
		if _, u := kv[i+1].(undefT); u {
			continue
		}
		m.Set(kv[i].(string), kv[i+1])
	}
	return m
}

// ecbCode — res.send(ecb(msg, code))
func ecbCode(c *web.Ctx, msg any, code string) {
	m := web.Ecb(msg)
	m.Set("code", code)
	c.Send(m)
}

// errMsg — catch (e) { ecb(e.stack || e.message) } 자리: 스택 없이 첫 줄만
func errMsg(err error) string {
	s := err.Error()
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	return s
}

// tokenKey — token.key 그대로(문자열). 없으면 ""
func tokenKey(c *web.Ctx) string {
	if c.Token == nil {
		return ""
	}
	return c.Token.Key
}

type errString string

// rowsAny — []*OrderedMap → []any (빈 결과도 [])
func rowsAny(rows []*jsx.OrderedMap) []any {
	out := make([]any, len(rows))
	for i, r := range rows {
		out[i] = r
	}
	return out
}

// mapOrNil — nil 객체는 JSON null 로(타입 있는 nil 을 그대로 넘기면 직렬화가 멈춘다)
func mapOrNil(m *jsx.OrderedMap) any {
	if m == nil {
		return nil
	}
	return m
}
