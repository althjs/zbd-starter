package admin

// 관리자 글·댓글·신고·알림 목록
//
//	GET /api/v1/admin/posts?groupKey=(필수)&groupSubKey&status&from&to&schedule=wait|live|end&q&sort=sortNo,-createdAt  (감사표 1-3·1-4)
//	GET /api/v1/admin/comments?groupKey&status&from&to&q&sort                                                        (3-2)
//	GET /api/v1/admin/reports?status&target=post|comment&q&sort=-reportedAt                                            (1-6)
//	GET /api/v1/admin/notifications?size&before=&unread=1                                                              (1-7, 커서)

import (
	"regexp"

	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

var reGroupKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,20}$`)

// 일정 시각(문자열 — 「2026-09-01T10:00:00.000Z」 또는 「2026-09-01 10:00」) → 초. Z 로 끝나면 UTC 로 본다.
func schedSec(path string) string {
	v := "JSON_UNQUOTE(JSON_EXTRACT(p.json, '" + path + "'))"
	return "COALESCE(UNIX_TIMESTAMP(STR_TO_DATE(LEFT(REPLACE(" + v + ", 'T', ' '), 19), '%Y-%m-%d %H:%i:%s'))" +
		" + IF(RIGHT(" + v + ", 1) = 'Z', TIMESTAMPDIFF(SECOND, UTC_TIMESTAMP(), NOW()), 0), 0)"
}

// sortNo 가 없는 글은 뒤로(화면이 lodash orderBy 로 하던 것과 같게)
const sortNoExpr = "COALESCE(CAST(JSON_UNQUOTE(JSON_EXTRACT(p.json, '$.sortNo')) AS DECIMAL(20,4)), 999999999)"

var postSort = map[string]string{
	"sortNo":    sortNoExpr,
	"createdAt": "p.createdAt",
	"updatedAt": "p.updatedAt",
	"title":     "p.title",
	"status":    "p.status",
	"postId":    "p.postId",
}

func posts(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, postSort, "sortNo,-createdAt", "p.postId")
	if !ok {
		return
	}
	gk := c.Q("groupKey")
	if !reGroupKey.MatchString(gk) {
		c.Fail(400, web.ErrBadRequest, "groupKey 가 필요합니다")
		return
	}
	cid := c.Token.HostingID
	w := &v1x.Where{}
	w.Add("p.hostingId = ?", cid)
	w.Add("p.groupKey = ?", gk)
	if s := c.Q("groupSubKey"); s != "" {
		w.Add("p.groupSubKey IN (?)", v1x.Strs(s))
	}
	if !v1x.IntIn(c, w, "status", "p.status") {
		return
	}
	if !v1x.DateRange(c, w, "p.createdAt") {
		return
	}
	from, to := schedSec("$.schedule.from"), schedSec("$.schedule.to")
	switch c.Q("schedule") {
	case "":
	case "wait":
		w.Add("UNIX_TIMESTAMP() < " + from)
	case "live":
		w.Add("UNIX_TIMESTAMP() >= " + from + " AND UNIX_TIMESTAMP() < " + to)
	case "end":
		w.Add("UNIX_TIMESTAMP() >= " + from + " AND UNIX_TIMESTAMP() >= " + to)
	default:
		c.Fail(400, web.ErrBadRequest, "schedule 은 wait|live|end")
		return
	}
	if lq.Q != "" {
		// 본문은 json 안이라 색인이 없다 — 그룹 안에서만 훑는다(LIST_DDL.sql 의 searchText 전문 색인 제안)
		like := web.LikeContains(lq.Q)
		w.Add(`(p.title LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(p.json, '$.body')) LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(p.jsonUser, '$.body')) LIKE ?)`, like, like, like)
	}
	rows, total, err := v1x.Page(c.Context(), c.App.DB, `SELECT p.postId, p.groupKey, p.groupSubKey, p.title, p.status, p.adminId, p.userId,
  UNIX_TIMESTAMP(p.createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(p.updatedAt)*1000 AS updatedAt,
  JSON_EXTRACT(p.json, '$.sortNo') AS sortNo,
  JSON_EXTRACT(p.json, '$.isFixTop') AS isFixTop,
  JSON_EXTRACT(p.json, '$.schedule') AS schedule,
  0 AS commentCount`, " FROM posts p", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	var uids []any
	for _, r := range rows {
		uids = append(uids, r.V("userId"))
	}
	names, err := v1x.Writers(c.Context(), c.App.DB, cid, uids)
	if err != nil {
		c.Ecb(err)
		return
	}
	for _, r := range rows {
		r.Set("sortNo", jsonVal(r.V("sortNo")))
		r.Set("isFixTop", jsx.Truthy(jsonVal(r.V("isFixTop"))))
		r.Set("schedule", jsonVal(r.V("schedule")))
		var writer any
		if jsx.Truthy(r.V("userId")) {
			writer = jsx.Obj("userId", r.V("userId"), "nickname", names[jsx.Str(r.V("userId"))])
		}
		r.Set("writer", writer)
	}
	c.SendList(v1x.Any(rows), total, lq)
}

// jsonVal — JSON_EXTRACT 결과(문자열) → 값(없으면 null)
func jsonVal(v any) any {
	s := jsx.StrOrEmpty(v)
	if s == "" {
		return nil
	}
	return jsx.ParseOr(s, nil)
}

var commentSort = map[string]string{
	"createdAt": "c.createdAt",
	"commentId": "c.commentId",
	"userId":    "c.userId",
	"title":     "p.title",
	"groupKey":  "p.groupKey",
}

var reportSort = map[string]string{"reportedAt": "r.createdAt", "reactionId": "r.reactionId", "status": "r.status"}
