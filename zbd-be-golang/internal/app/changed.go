package app

// 설정·권한을 바꾼 뒤 캐시 다시 읽기 — 플러그인 공용 묶음(zbd-plugins/common)도 같은 파일을 들고 있다(옛 스타터에도 붙게). 운영은 app.go 에 같은 함수.

import (
	"context"

	"zbd.kr/be/internal/lib/cachegen"
)

// Changed — 설정·권한을 바꾼 뒤: 이 프로세스 캐시를 바로 다시 읽고(시간 캐시는 cachegen 세대가 바뀌면 버린다) 열린 화면에 알린다(웹소켓이 있을 때)
func (a *App) Changed(ctx context.Context, hostingID string) {
	a.ReloadCaches(ctx)
	cachegen.Bump()
	a.Emit("doUpdateCodes", map[string]any{"hostingId": hostingID})
}
