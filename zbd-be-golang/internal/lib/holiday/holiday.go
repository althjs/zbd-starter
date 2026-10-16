package holiday

// 공휴일 한 곳(2026-10-16 대표 — 「달력뷰에 공휴일 연동이 안 되어 있네」).
//
//	코드 표(아래, 대체공휴일·선거일 포함 — 해마다 더한다) + 운영자가 넣은 설정 COMMON.holydays([{ymd, name}]) 를 합친다.
//	같은 날이면 설정이 이긴다(임시공휴일·이름 고치기). 설정 캐시가 다시 읽힐 때 settings 가 SetExtra 로 넘긴다.
//	쓰는 곳: 근무 관리(근무일 판정)·공유 일정·업무 달력·라이프 달력(COMMON.holydays 로 내려가는 합친 목록).

import (
	"sort"
	"sync"

	"zbd.kr/be/internal/jsx"
)

// Day — 공휴일 하루
type Day struct{ Ymd, Name string }

var code = map[string]string{
	// 2025
	"2025-01-01": "신정", "2025-01-27": "임시공휴일", "2025-01-28": "설날 연휴", "2025-01-29": "설날", "2025-01-30": "설날 연휴",
	"2025-03-01": "삼일절", "2025-03-03": "대체공휴일(삼일절)", "2025-05-05": "어린이날·부처님오신날", "2025-05-06": "대체공휴일",
	"2025-06-03": "대통령 선거일", "2025-06-06": "현충일", "2025-08-15": "광복절", "2025-10-03": "개천절",
	"2025-10-05": "추석 연휴", "2025-10-06": "추석", "2025-10-07": "추석 연휴", "2025-10-08": "대체공휴일(추석)", "2025-10-09": "한글날", "2025-12-25": "성탄절",
	// 2026
	"2026-01-01": "신정", "2026-02-16": "설날 연휴", "2026-02-17": "설날", "2026-02-18": "설날 연휴",
	"2026-03-01": "삼일절", "2026-03-02": "대체공휴일(삼일절)", "2026-05-05": "어린이날", "2026-05-24": "부처님오신날", "2026-05-25": "대체공휴일(부처님오신날)",
	"2026-06-03": "지방선거일", "2026-06-06": "현충일", "2026-08-15": "광복절", "2026-08-17": "대체공휴일(광복절)",
	"2026-09-24": "추석 연휴", "2026-09-25": "추석", "2026-09-26": "추석 연휴", "2026-10-03": "개천절", "2026-10-05": "대체공휴일(개천절)",
	"2026-10-09": "한글날", "2026-12-25": "성탄절",
	// 2027
	"2027-01-01": "신정", "2027-02-06": "설날 연휴", "2027-02-07": "설날", "2027-02-08": "설날 연휴", "2027-02-09": "대체공휴일(설날)",
	"2027-03-01": "삼일절", "2027-05-05": "어린이날", "2027-05-13": "부처님오신날", "2027-06-06": "현충일",
	"2027-08-15": "광복절", "2027-08-16": "대체공휴일(광복절)", "2027-09-14": "추석 연휴", "2027-09-15": "추석", "2027-09-16": "추석 연휴",
	"2027-10-03": "개천절", "2027-10-04": "대체공휴일(개천절)", "2027-10-09": "한글날", "2027-10-11": "대체공휴일(한글날)",
	"2027-12-25": "성탄절", "2027-12-27": "대체공휴일(성탄절)",
}

var (
	mu    sync.RWMutex
	extra = map[string]string{}
)

// SetExtra — 설정 COMMON.holydays 를 넘긴다([{ymd, name}] — 모양이 틀린 줄은 버린다)
func SetExtra(list any) {
	m := map[string]string{}
	if arr, ok := list.([]any); ok {
		for _, x := range arr {
			o, _ := x.(*jsx.OrderedMap)
			if o == nil {
				continue
			}
			ymd, name := jsx.StrOrEmpty(o.V("ymd")), jsx.StrOrEmpty(o.V("name"))
			if len(ymd) == 10 && ymd[4] == '-' && ymd[7] == '-' {
				m[ymd] = name
			}
		}
	}
	mu.Lock()
	extra = m
	mu.Unlock()
}

// Name — 그날 공휴일 이름(아니면 "")
func Name(ymd string) string {
	mu.RLock()
	n, ok := extra[ymd]
	mu.RUnlock()
	if ok {
		if n == "" {
			n = "공휴일"
		}
		return n
	}
	return code[ymd]
}

// Is — 공휴일인가
func Is(ymd string) bool { return Name(ymd) != "" }

// Between — from~to(YYYY-MM-DD, 둘 다 포함) 사이 공휴일, 날짜순
func Between(from, to string) []Day {
	out := []Day{}
	for _, d := range All() {
		if d.Ymd >= from && d.Ymd <= to {
			out = append(out, d)
		}
	}
	return out
}

// All — 합친 전체, 날짜순
func All() []Day {
	seen := map[string]bool{}
	out := []Day{}
	mu.RLock()
	for y := range extra {
		seen[y] = true
	}
	mu.RUnlock()
	for y := range code {
		seen[y] = true
	}
	for y := range seen {
		out = append(out, Day{y, Name(y)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ymd < out[j].Ymd })
	return out
}

// List — COMMON.holydays 모양([{ymd, name}])
func List() []any {
	out := []any{}
	for _, d := range All() {
		out = append(out, jsx.Obj("ymd", d.Ymd, "name", d.Name))
	}
	return out
}
