package admin

// 운영 화면 목록
//
//	GET /api/v1/admin/support/threads?page&size&q(회원)&unread=1&sort=-lastId        (감사표 2-2)
//	GET /api/v1/admin/plugin-licenses?page&size&q(회원)&state=trial|active|canceling|paused|expired&pluginKey&sort  (2-3)
//	GET /api/v1/admin/token-logs?page&size&from&to&ip&result&sort=-logId              (2-4, 개발자·최고 관리자)
//	GET /api/v1/admin/security/attacks?days=1|3|7|14&cat=&page&size&sort=-susp        (2-5, 외부 공격 감지 플러그인 — attacks.go)
//	GET /api/v1/admin/tag-log/summary?from&to&interval=day|5m&hosting=                  (1-9, ES 집계)

import (
	"net/http"
	"regexp"

	"zbd.kr/be/internal/api/collab"
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

var supportSort = map[string]string{"lastId": "t.lastId", "unread": "t.unread", "userId": "t.userId"}

var pluginSort = map[string]string{"id": "l.id", "createdAt": "l.createdAt", "expiresAt": "l.expiresAt", "updatedAt": "l.updatedAt"}

var rePluginKey = regexp.MustCompile(`^[a-z0-9_-]{1,40}$`)

// licenseStateSQL — core.licenseState 와 같은 판정(정지 > 만료 > 체험 > 자동연장 > 해지 예정)
const licenseStateSQL = `CASE WHEN l.status = 1 THEN 'paused'
  WHEN IFNULL(l.expiresAt, 0) > 0 AND l.expiresAt <= ? THEN 'expired'
  WHEN l.kind = 'trial' THEN 'trial'
  WHEN IFNULL(l.autoRenew, 0) <> 0 THEN 'active'
  ELSE 'canceling' END`

func securityGuard(c *web.Ctx) bool {
	if !collab.CanViewSystemSecurity(c) {
		c.Fail(http.StatusForbidden, web.ErrForbidden, "개발자·최고 관리자만 볼 수 있습니다")
		return false
	}
	return true
}

var tokenLogSort = map[string]string{"logId": "logId", "createdAt": "createdAt", "ip": "ip"}

func tokenLogs(c *web.Ctx) {
	if !v1x.OnlyGET(c) || !securityGuard(c) {
		return
	}
	lq, ok := v1x.List(c, tokenLogSort, "-logId", "logId")
	if !ok {
		return
	}
	w := &v1x.Where{}
	if !v1x.DateRange(c, w, "createdAt") {
		return
	}
	if s := c.Q("ip"); s != "" {
		w.Add("ip = ?", s)
	}
	if s := c.Q("result"); s != "" {
		w.Add("result = ?", s)
	}
	if s := c.Q("kind"); s != "" {
		w.Add("kind = ?", s)
	}
	if lq.Q != "" {
		like := web.LikePrefix(lq.Q)
		w.Add("(issuerName LIKE ? OR targetName LIKE ? OR ip LIKE ?)", like, like, like)
	}
	ctx := c.Context()
	if c.Q("groupBy") == "ip" {
		tokenLogsByIP(c, w, lq)
		return
	}
	rows, total, err := v1x.Page(ctx, c.App.DB, `SELECT logId, kind, result, reason, issuerAdminId, issuerName, issuerHostingId,
  targetAdminId, targetName, targetHostingId, ip, ipCandidates, isOffice, host, ua, UNIX_TIMESTAMP(createdAt)*1000 AS createdAt`,
		" FROM admin_token_logs", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.SendList(v1x.Any(rows), total, lq)
}

// tokenLogsByIP — IP 별 묶음(발급 수·성공·거절·마지막 시각) 한 쪽
func tokenLogsByIP(c *web.Ctx, w *v1x.Where, lq web.ListQuery) {
	ctx := c.Context()
	inner := " FROM (SELECT ip, COUNT(*) AS count, SUM(result = 'ok') AS ok, SUM(result <> 'ok') AS denied, MAX(isOffice) AS isOffice," +
		" UNIX_TIMESTAMP(MAX(createdAt))*1000 AS lastAt, MAX(logId) AS logId FROM admin_token_logs" + w.SQL() + " GROUP BY ip) g"
	cnt, err := db.QueryOne(ctx, c.App.DB, "SELECT COUNT(*) AS n"+inner, w.Args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	total := jsx.Int(cnt.V("n"))
	if total == 0 || int64(lq.Offset()) >= total {
		c.SendList(nil, total, lq)
		return
	}
	// 묶음에서는 정렬 칸이 logId(최근)·ip 만 뜻이 있다 — createdAt 은 마지막 시각으로
	order := " ORDER BY g.logId DESC"
	switch c.Q("sort") {
	case "ip":
		order = " ORDER BY g.ip ASC"
	case "-ip":
		order = " ORDER BY g.ip DESC"
	case "logId", "createdAt":
		order = " ORDER BY g.logId ASC"
	}
	rows, err := db.Query(ctx, c.App.DB, "SELECT g.ip, g.count, g.ok, g.denied, g.isOffice, g.lastAt"+inner+order+" LIMIT ?, ?",
		append(append([]any{}, w.Args...), lq.Offset(), lq.Size)...)
	if err != nil {
		c.Ecb(err)
		return
	}
	for _, r := range rows {
		for _, k := range []string{"count", "ok", "denied"} {
			r.Set(k, jsx.Number(r.V(k)))
		}
	}
	c.SendList(v1x.Any(rows), total, lq)
}

var reHostingKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,50}$`)
