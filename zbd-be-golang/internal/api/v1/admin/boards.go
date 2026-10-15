package admin

// 관리자 앱(위젯) 관리 · 업무관리 카드
//
//	GET /api/v1/admin/widgets?page&size&q&status=public|private&category=&sort=-updatedAt   (감사표 1-16)
//	GET /api/v1/admin/work-items/[groupKey]?page&size&q&status=10,20&adminId=&category=c1:1&parent=0&doneStatus=30&doneSince=YYYY-MM-DD&sort  (2-1)
//
// 앱 q = 제목 부분 일치 · postId 정확 · 제작자(닉네임 앞부분, 이메일·전화·실명은 정확 일치) — 회원을 통째로 풀어 비교하지 않는다.
// 업무관리 경로는 원래 /admin/work/[workId]/[shapeId] 와 겹쳐서 /admin/work-items/[groupKey] 로 했다.

import (
	"regexp"
	"strings"

	"zbd.kr/be/internal/api/admincontent"
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

var widgetSort = map[string]string{"updatedAt": "p.updatedAt", "createdAt": "p.createdAt", "title": "p.title", "status": "p.status"}

var workSort = map[string]string{
	"workId":     "workId",
	"createdAt":  "createdAt",
	"updatedAt":  "updatedAt",
	"status":     "CAST(JSON_UNQUOTE(JSON_EXTRACT(Json, '$.status')) AS SIGNED)",
	"title":      "JSON_UNQUOTE(JSON_EXTRACT(Json, '$.title'))",
	"scheduleTo": "JSON_UNQUOTE(JSON_EXTRACT(Json, '$.schedule.to'))",
}

var reCatFilter = regexp.MustCompile(`^([A-Za-z0-9_]{1,20}):(-?\d{1,9})$`)

func workItems(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, workSort, "-workId", "workId")
	if !ok {
		return
	}
	gk := c.Q("groupKey")
	if !reGroupKey.MatchString(gk) {
		c.Fail(400, web.ErrBadRequest, "groupKey 가 필요합니다")
		return
	}
	cids := admincontent.WorkHostingIDs(c, gk)
	ca := make([]any, len(cids))
	for i, x := range cids {
		ca[i] = x
	}
	w := &v1x.Where{}
	w.Add("hostingId IN (?)", ca)
	w.Add("groupKey = ?", gk)
	if s := c.Q("parent"); s != "" {
		n, ok := v1x.IntParam(c, "parent", 0)
		if !ok {
			return
		}
		w.Add("parentWorkId = ?", n)
	}
	status := "CAST(JSON_UNQUOTE(JSON_EXTRACT(Json, '$.status')) AS SIGNED)"
	if !v1x.IntIn(c, w, "status", status) {
		return
	}
	// 칸반: 「완료」 처럼 끝없이 쌓이는 상태는 기간으로 자른다(doneStatus 의 카드만 doneSince 이후 고친 것)
	ds, ok := v1x.YMDParam(c, "doneSince")
	if !ok {
		return
	}
	if ds != "" {
		done, ok := v1x.IntParam(c, "doneStatus", 30)
		if !ok {
			return
		}
		w.Add("(IFNULL("+status+", 0) <> ? OR updatedAt >= ?)", done, ds+" 00:00:00")
	}
	if s := c.Q("adminId"); s != "" {
		ids := v1x.Ints(s)
		if len(ids) == 0 {
			c.Fail(400, web.ErrBadRequest, "adminId 는 정수")
			return
		}
		var parts []string
		var args []any
		for _, id := range ids {
			parts = append(parts, "JSON_CONTAINS(IFNULL(JSON_EXTRACT(Json, '$.adminIds'), '[]'), ?)")
			args = append(args, jsx.Str(id))
		}
		w.Add("("+strings.Join(parts, " OR ")+")", args...)
	}
	for _, f := range v1x.Strs(c.Q("category")) {
		m := reCatFilter.FindStringSubmatch(jsx.Str(f))
		if m == nil {
			c.Fail(400, web.ErrBadRequest, "category 는 키:값(예 c1:1)")
			return
		}
		w.Add("CAST(JSON_UNQUOTE(JSON_EXTRACT(Json, ?)) AS SIGNED) = ?", `$.category."`+m[1]+`"`, m[2])
	}
	if lq.Q != "" {
		like := web.LikeContains(lq.Q)
		w.Add("(JSON_UNQUOTE(JSON_EXTRACT(Json, '$.title')) LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(Json, '$.html')) LIKE ? OR workId = ?)", like, like, lq.Q)
	}
	rows, total, err := v1x.Page(c.Context(), c.App.DB, `SELECT workId, parentWorkId, adminId, hostingId, groupKey,
  UNIX_TIMESTAMP(createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(updatedAt)*1000 AS updatedAt,
  JSON_UNQUOTE(JSON_EXTRACT(Json, '$.title')) AS title,
  JSON_EXTRACT(Json, '$.status') AS status,
  JSON_EXTRACT(Json, '$.adminIds') AS adminIds,
  JSON_EXTRACT(Json, '$.category') AS category,
  JSON_EXTRACT(Json, '$.schedule') AS schedule,
  JSON_EXTRACT(Json, '$.claudeChat') AS claudeChat,
  IFNULL(JSON_LENGTH(Json, '$.shapes'), 0) AS shapeCount`, " FROM work_items", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	for _, r := range rows {
		r.Set("status", jsonVal(r.V("status")))
		for _, k := range []string{"adminIds", "category", "schedule", "claudeChat"} {
			r.Set(k, jsonVal(r.V(k)))
		}
		if r.V("adminIds") == nil {
			r.Set("adminIds", []any{})
		}
		r.Set("shapeCount", jsx.Number(r.V("shapeCount")))
	}
	c.SendList(v1x.Any(rows), total, lq)
}
