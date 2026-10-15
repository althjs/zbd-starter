package core

import (
	"context"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// adminDashboard — 관리자 첫 화면: 최근 글·댓글·신고 · 회원 수. pages/api/admin/dashboard
// adminDashboard — 관리자 첫 화면: 회원 현황 · 게시판별 글 수 · 업무 상태 · 최근 글·회원·업무
func adminDashboard(c *web.Ctx) {
	ctx := c.Context()
	cid := c.Token.HostingID
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	fail := func(err error) { c.Ecb(err) }
	summary, err := dashMemberSummary(ctx, conn, cid)
	if err != nil {
		fail(err)
		return
	}
	posts, err := dashRecentPosts(c, conn, cid, 6)
	if err != nil {
		fail(err)
		return
	}
	members, err := dashRecentMembers(c, conn, cid, 6)
	if err != nil {
		fail(err)
		return
	}
	boards, err := dashBoardCounts(c, conn, cid)
	if err != nil {
		fail(err)
		return
	}
	work, err := dashWork(ctx, conn, cid)
	if err != nil {
		fail(err)
		return
	}
	c.Scb(jsx.Obj("memberSummary", summary, "recentPosts", posts, "recentMembers", members, "boards", boards,
		"work", work))
}

// dashRecentMembers — 최근 가입 회원(닉네임·가입일만 — 이메일·이름은 싣지 않는다)
func dashRecentMembers(c *web.Ctx, q db.Querier, cid string, limit int) ([]any, error) {
	rows, err := query(c.Context(), q, `SELECT userId, nickname, status, UNIX_TIMESTAMP(createdAt)*1000 AS createdAt FROM members
WHERE hostingId = ? AND status <> -1 ORDER BY userId DESC LIMIT ?`, cid, limit)
	if err != nil {
		return nil, err
	}
	out := []any{}
	for _, r := range rows {
		out = append(out, jsx.Obj("userId", r.V("userId"), "nickname", jsx.StrOrEmpty(r.V("nickname")), "status", r.V("status"), "createdAt", r.V("createdAt")))
	}
	return out, nil
}

// dashBoardCounts — 게시판(설정의 관리자·사용자 게시판)별 공개 글 수
func dashBoardCounts(c *web.Ctx, q db.Querier, cid string) ([]any, error) {
	pg, _ := c.App.Settings.Meta(cid, "POST_GROUPS").(*jsx.OrderedMap)
	type bd struct{ key, name, kind string }
	var list []bd
	for _, k := range []string{"adminPostGroups", "userPostGroups"} {
		for _, x := range pg.Arr(k) {
			if g, _ := x.(*jsx.OrderedMap); g != nil && jsx.Truthy(g.V("groupKey")) {
				kind := "admin"
				if k == "userPostGroups" {
					kind = "user"
				}
				list = append(list, bd{jsx.Str(g.V("groupKey")), jsx.StrOrEmpty(g.V("name")), kind})
			}
		}
	}
	out := []any{}
	if len(list) == 0 {
		return out, nil
	}
	keys := make([]any, len(list))
	for i, b := range list {
		keys[i] = b.key
	}
	rows, err := query(c.Context(), q, `SELECT groupKey, COUNT(*) AS cnt, SUM(createdAt >= DATE_SUB(NOW(), INTERVAL 7 DAY)) AS week FROM posts
WHERE hostingId = ? AND status = 0 AND groupKey IN (?) GROUP BY groupKey`, cid, keys)
	if err != nil {
		return nil, err
	}
	cnt := map[string][2]float64{}
	for _, r := range rows {
		cnt[jsx.Str(r.V("groupKey"))] = [2]float64{jsx.NumOr(r.V("cnt"), 0), jsx.NumOr(decWeek(r.V("week")), 0)}
	}
	for _, b := range list {
		v := cnt[b.key]
		out = append(out, jsx.Obj("groupKey", b.key, "name", b.name, "kind", b.kind, "count", v[0], "week", v[1]))
	}
	return out, nil
}

// decWeek — SUM() 이 DECIMAL 문자열로 올 때 숫자로
func decWeek(v any) any {
	if s, ok := v.(string); ok {
		f := 0.0
		for _, ch := range s {
			if ch >= '0' && ch <= '9' {
				f = f*10 + float64(ch-'0')
			} else {
				break
			}
		}
		return f
	}
	return v
}

// dashWork — 업무관리(TASK) 상태별 카드 수 + 최근 고친 카드
func dashWork(ctx context.Context, q db.Querier, cid string) (*jsx.OrderedMap, error) {
	rows, err := query(ctx, q, `SELECT JSON_EXTRACT(Json, '$.status') AS status, COUNT(*) AS cnt FROM work_items
WHERE hostingId = ? AND groupKey = 'TASK' AND parentWorkId = 0 GROUP BY status`, cid)
	if err != nil {
		return nil, err
	}
	by := []any{}
	for _, r := range rows {
		by = append(by, jsx.Obj("status", jsx.NumOr(r.V("status"), 0), "count", jsx.NumOr(r.V("cnt"), 0)))
	}
	rec, err := query(ctx, q, `SELECT workId, JSON_UNQUOTE(JSON_EXTRACT(Json, '$.title')) AS title, JSON_EXTRACT(Json, '$.status') AS status,
  UNIX_TIMESTAMP(COALESCE(updatedAt, createdAt))*1000 AS updatedAt FROM work_items
WHERE hostingId = ? AND groupKey = 'TASK' AND parentWorkId = 0 ORDER BY COALESCE(updatedAt, createdAt) DESC, workId DESC LIMIT 6`, cid)
	if err != nil {
		return nil, err
	}
	recent := []any{}
	for _, r := range rec {
		recent = append(recent, jsx.Obj("workId", r.V("workId"), "title", jsx.StrOrEmpty(r.V("title")), "status", jsx.NumOr(r.V("status"), 0), "updatedAt", r.V("updatedAt")))
	}
	return jsx.Obj("byStatus", by, "recent", recent), nil
}

// 대시보드 「최근 게시글」 에서 빼는 묶음 — 회원 비공개 기록(라이프·AI 대화)과 앱 문서는 게시판 글이 아니다.
var dashPrivateGroups = []any{"LIFE", "AI_CHAT", "WIDGET"}

// dashRecentPosts — 공개 게시판 글만(status 0 · 등록된 게시판 · 비공개 묶음 제외), 작성자는 회원 닉네임.
// 지금 서비스와 일부러 다르다(대표 2026-10-03): 예전엔 라이프 기록까지 섞였고 본문 json 을 통째로 내렸다.
// 화면이 jsonUser.name 을 작성자로 읽으므로 jsonUser 는 {name} 만 싣는다.
func dashRecentPosts(c *web.Ctx, q db.Querier, cid string, limit int) ([]any, error) {
	ctx := c.Context()
	rows, err := query(ctx, q, `SELECT postId, groupKey, title, userId, status, createdAt FROM posts
WHERE hostingId = ? AND userId <> 0 AND status = 0 AND groupKey NOT IN (?)
ORDER BY createdAt DESC LIMIT ?`, cid, dashPrivateGroups, limit*5)
	if err != nil {
		return nil, err
	}
	picked := []*jsx.OrderedMap{}
	var ids []any
	seen := map[string]bool{}
	for _, o := range rows {
		gk := jsx.StrOrEmpty(o.V("groupKey"))
		if strings.HasPrefix(gk, "WIDGET_") || c.App.Settings.PostGroup(cid, gk) == nil {
			continue
		}
		picked = append(picked, o)
		if k := jsx.Str(o.V("userId")); !seen[k] {
			seen[k] = true
			ids = append(ids, o.V("userId"))
		}
		if len(picked) >= limit {
			break
		}
	}
	names := map[string]string{}
	if len(ids) > 0 {
		ms, err := member.ByIDs(ctx, c.App, q, cid, ids, false)
		if err != nil {
			return nil, err
		}
		for _, m := range ms {
			names[jsx.Str(m.V("userId"))] = jsx.StrOrEmpty(member.PublicCreator(m).V("nickname"))
		}
	}
	out := []any{}
	for _, o := range picked {
		name := names[jsx.Str(o.V("userId"))]
		if name == "" {
			name = "사용자"
		}
		o.Set("jsonUser", jsx.Obj("name", name))
		out = append(out, o)
	}
	return out, nil
}

func dashMemberSummary(ctx context.Context, q db.Querier, cid string) (*jsx.OrderedMap, error) {
	os, err := query(ctx, q, `SELECT count(*) AS cnt, status FROM members 
  WHERE hostingId = ? 
GROUP BY status`, cid)
	if err != nil {
		return nil, err
	}
	day, _ := time.Parse("2006-01-02", util.Today())
	o, err := queryOne(ctx, q, `SELECT count(*) AS cnt FROM members 
  WHERE hostingId = ? AND UNIX_TIMESTAMP(createdAt)*1000 >= ?`, cid, day.UnixMilli())
	if err != nil {
		return nil, err
	}
	total := 0.0
	leave, leaveReq := 0.0, 0.0
	gotLeave, gotReq := false, false
	for _, r := range os {
		total += jsx.NumOr(r.V("cnt"), 0)
		st := jsx.Number(r.V("status"))
		if st == -1 && !gotLeave {
			gotLeave = true
			leave = jsx.NumOr(r.V("cnt"), 0)
		}
		if st == 99 && !gotReq {
			gotReq = true
			leaveReq = jsx.NumOr(r.V("cnt"), 0)
		}
	}
	today := 0.0
	if o != nil {
		today = jsx.NumOr(o.V("cnt"), 0)
	}
	return jsx.Obj("totalCount", total, "todayCount", today, "leaveCount", leave, "leaveRequestCount", leaveReq), nil
}
