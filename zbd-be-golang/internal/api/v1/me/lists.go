package me

import (
	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

var commentSort = map[string]string{"createdAt": "c.createdAt", "commentId": "c.commentId"}

// userGroups — 회원 게시판(POST_GROUPS.userPostGroups). writable = 글쓰기 게시판만(기본 목록)
func userGroups(c *web.Ctx, writable bool) map[string]bool {
	out := map[string]bool{}
	pg, _ := c.App.Settings.Meta(c.Token.HostingID, "POST_GROUPS").(*jsx.OrderedMap)
	for _, x := range pg.Arr("userPostGroups") {
		g, _ := x.(*jsx.OrderedMap)
		k := jsx.Str(g.V("groupKey"))
		if g == nil || k == "CS" {
			continue
		}
		if writable && !jsx.Truthy(g.Path("feature", "useWrite")) {
			continue
		}
		out[k] = true
	}
	return out
}

var postSort = map[string]string{"createdAt": "createdAt", "updatedAt": "updatedAt", "postId": "postId", "title": "title"}

// posts — 내가 쓴 글. groupKey 를 안 주면 글쓰기 게시판 전체(라이프 같은 개인 기록은 groupKey 로만).
func posts(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, postSort, "-createdAt", "postId")
	if !ok {
		return
	}
	var keys []any
	if s := c.Q("groupKey"); s != "" {
		all := userGroups(c, false)
		for _, k := range v1x.Strs(s) {
			if !all[jsx.Str(k)] {
				c.Fail(400, web.ErrBadRequest, "groupKey 가 올바르지 않습니다: "+jsx.Str(k))
				return
			}
			keys = append(keys, k)
		}
	} else {
		for k := range userGroups(c, true) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		c.SendList(nil, 0, lq)
		return
	}
	w := &v1x.Where{}
	w.Add("hostingId = ?", c.Token.HostingID)
	w.Add("userId = ?", c.Token.Key)
	w.Add("groupKey IN (?)", keys)
	w.Add("status <> -1")
	if lq.Q != "" {
		w.Add("title LIKE ?", web.LikeContains(lq.Q))
	}
	rows, total, err := v1x.Page(c.Context(), c.App.DB, `SELECT postId, groupKey, groupSubKey, title, status,
  UNIX_TIMESTAMP(createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(updatedAt)*1000 AS updatedAt,
  0 AS commentCount`, " FROM posts", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.SendList(v1x.Any(rows), total, lq)
}

// bookmarks — 담은 글(회원 json.postIds — 배열 뒤가 최근). 공개 글(status 0)·내 글만. 순서·쪽은 SQL 의 FIELD() 로.
func bookmarks(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := c.ParseList(map[string]string{}, "")
	if !ok {
		return
	}
	ctx := c.Context()
	cid := c.Token.HostingID
	m, err := db.QueryOne(ctx, c.App.DB, `SELECT JSON_EXTRACT(json, '$.postIds') AS ids FROM members WHERE hostingId = ? AND userId = ?`, cid, c.Token.Key)
	if err != nil {
		c.Ecb(err)
		return
	}
	var ids []any
	if m != nil {
		arr, _ := jsx.ParseOr(jsx.StrOrEmpty(m.V("ids")), nil).([]any)
		seen := map[string]bool{}
		for i := len(arr) - 1; i >= 0; i-- { // 최근 담은 것부터
			s := jsx.Str(arr[i])
			if s == "" || seen[s] || len(s) > 16 {
				continue
			}
			seen[s] = true
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		c.SendList(nil, 0, lq)
		return
	}
	w := &v1x.Where{}
	w.Add("p.hostingId = ?", cid)
	w.Add("p.postId IN (?)", ids)
	w.Add("(p.status = 0 OR p.userId = ?)", c.Token.Key)
	from := " FROM posts p"
	cnt, err := db.QueryOne(ctx, c.App.DB, "SELECT COUNT(*) AS n"+from+w.SQL(), w.Args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	total := jsx.Int(cnt.V("n"))
	if total == 0 || int64(lq.Offset()) >= total {
		c.SendList(nil, total, lq)
		return
	}
	rows, err := db.Query(ctx, c.App.DB, `SELECT p.postId, p.groupKey, p.groupSubKey, p.title, p.userId,
  UNIX_TIMESTAMP(p.createdAt)*1000 AS createdAt, m.nickname`+from+
		` LEFT JOIN members m ON m.userId = p.userId AND m.hostingId = p.hostingId`+w.SQL()+
		` ORDER BY FIELD(p.postId, ?) LIMIT ?, ?`, append(append([]any{}, w.Args...), ids, lq.Offset(), lq.Size)...)
	if err != nil {
		c.Ecb(err)
		return
	}
	for _, r := range rows {
		r.Set("writer", jsx.Obj("userId", r.V("userId"), "nickname", v1x.Nick(r.V("nickname"), r.V("userId"))))
		r.Delete("nickname")
	}
	c.SendList(v1x.Any(rows), total, lq)
}

var todoSort = map[string]string{"ts": "t.ts", "ymd": "t.ymd", "doneTS": "t.doneTS"}

// 프리셀 결과: v2 는 isWon, 옛 v1 은 첫·마지막 스냅숏의 isWon
const fcWonSQL = `(CASE WHEN JSON_EXTRACT(jsonUser, '$.v') = 2 THEN JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.isWon')) IN ('true', '1')
  ELSE (JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.history[0].isWon')) IN ('true', '1')
     OR JSON_UNQUOTE(JSON_EXTRACT(jsonUser, CONCAT('$.history[', GREATEST(IFNULL(JSON_LENGTH(jsonUser, '$.history'), 1) - 1, 0), '].isWon'))) IN ('true', '1')) END)`

const fcGaveupSQL = `IFNULL(JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.result')), '') = 'gaveup'`

var fcSort = map[string]string{"createdAt": "createdAt", "postId": "postId"}

var shareSort = map[string]string{"updatedAt": "updatedAt", "createdAt": "createdAt", "shareId": "shareId"}
