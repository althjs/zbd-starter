package util

import (
	"reflect"
	"testing"
)

// 정책: 사무실 IP = .env OFFICE_IPS + 관리자 설정(SITE.officeIps). 코드 기본값 없음(2026-10-16).
func TestOfficeIPsPolicy(t *testing.T) {
	defer func() { OfficeIPsFromSettings = nil }()
	OfficeIPsFromSettings = nil
	if got := OfficeIPs(""); len(got) != 0 {
		t.Fatalf("설정이 없으면 비어야 한다: %v", got)
	}
	OfficeIPsFromSettings = func() []string { return []string{" 1.2.3.4", "5.6.7.8", ""} }
	got := OfficeIPs("5.6.7.8, 9.9.9.9")
	want := []string{"5.6.7.8", "9.9.9.9", "1.2.3.4"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("합치기: got %v want %v", got, want)
	}
}
