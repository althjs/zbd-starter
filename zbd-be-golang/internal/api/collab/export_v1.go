package collab

// /api/v1 목록(internal/api/v1/…)이 같이 쓰는 것 — 보안 기록 권한·ES 조회를 그대로 쓰게 내보낸다.

import (
	"zbd.kr/be/internal/web"
)

// CanViewSystemSecurity — 시스템 보안 기록(공격 통계·토큰 발급 기록)을 볼 수 있는가(개발자 또는 본사 마스터)
func CanViewSystemSecurity(c *web.Ctx) bool { return canViewSystemSecurity(c, c.Token) }
