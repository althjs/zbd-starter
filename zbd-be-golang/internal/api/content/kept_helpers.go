package content

import (
	"regexp"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// 지운 기능 파일에 있던 공용 도우미 — 남은 코드가 쓴다(스타터 정리 2026-10-04)

// orStr — JS 의 (s || d) 를 문자열로
func orStr(v any, d string) string {
	if !jsx.Truthy(v) {
		return d
	}
	return jsx.Str(v)
}

// orVal — x || d
func orVal(v any, d any) any {
	if !jsx.Truthy(v) {
		return d
	}
	return v
}

// containsStrict — JS 의 arr.includes(v)(원시값만)
func containsStrict(arr []any, v any) bool {
	for _, x := range arr {
		if strictEq(x, v) {
			return true
		}
	}
	return false
}

type errString string

// strictEq — JS 의 a === b (원시값만)
func strictEq(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case string:
		y, ok := b.(string)
		return ok && x == y
	case float64:
		y, ok := b.(float64)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	}
	return false // 객체는 같은 참조여야 — 여기선 늘 다르다
}

// notifyLifeChanged — 라이프 하루 기록이 바뀌면 그 사람의 다른 화면에 알린다(보낸 브라우저는 zbduid 로 거른다)
func notifyLifeChanged(c *web.Ctx, userID float64, ymd any, action string) {
	y := ""
	if ymd != nil {
		y = jsx.StrOrEmpty(ymd)
	}
	if userID == 0 || !ymdRe.MatchString(y) {
		return
	}
	uid := c.BrowserUID()
	if r := []rune(uid); len(r) > 64 {
		uid = string(r[:64])
	}
	c.App.Emit("doUserDataUpdate", jsx.Obj("userId", userID, "groupKey", "LIFE", "key", y, "action", action, "zbduid", uid,
		"oxuid", uid)) // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
}

var shuffleKeys = []int{1, 6, 3, 8, 2}

var ymdRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (e errString) Error() string { return string(e) }
