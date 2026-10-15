package admincontent

import "zbd.kr/be/internal/jsx"

// 지운 기능 파일에 있던 공용 도우미 — 남은 코드가 쓴다(스타터 정리 2026-10-04)

func contains(arr []string, s string) bool {
	for _, x := range arr {
		if x == s {
			return true
		}
	}
	return false
}

// nullishStr — `${v ?? ""}`
func nullishStr(v any) string {
	if v == nil {
		return ""
	}
	return jsx.Str(v)
}

// orStr — (v || "") 를 문자열로 이어 붙일 때
func orStr(v any) string { return jsx.StrOrEmpty(v) }

// tplStr — 템플릿 문자열 안의 ${v}
func tplStr(v any) string {
	if v == nil {
		return "null"
	}
	return jsx.Str(v)
}
