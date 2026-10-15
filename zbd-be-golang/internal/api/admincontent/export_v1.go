package admincontent

// /api/v1 목록(internal/api/v1/…)이 같이 쓰는 것 — 이 패키지의 화면 모양·규칙을 그대로 쓰게 내보낸다.

import (
	"zbd.kr/be/internal/web"
)

// WorkHostingIDs — 그 보드를 볼 수 있는 회사들(공유 보드면 "" 도)
func WorkHostingIDs(c *web.Ctx, groupKey string) []string {
	return workAccessibleHostingIDs(c, c.Token, groupKey)
}
