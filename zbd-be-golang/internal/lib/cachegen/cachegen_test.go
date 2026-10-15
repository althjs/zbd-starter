package cachegen

import "testing"

// 정책: 설정이 바뀌면(Bump) 세대가 올라 — 세대를 적어 둔 작은 캐시가 다음 조회 때 새로 읽는다
func TestBump(t *testing.T) {
	g := Gen()
	Bump()
	if Gen() != g+1 {
		t.Fatalf("Bump 뒤 세대 %d, 기대 %d", Gen(), g+1)
	}
}
