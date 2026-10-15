package admincontent

// 관리자 게시물·고객센터 — /admin/post/*, /admin/cs/*
// 원본: pages/api/admin/post/*, pages/api/admin/cs/*

import (
	"bytes"
	"context"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

const postCols = `postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt,  UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId`

func getPost(ctx context.Context, q db.Querier, postID any, hostingID string) (*jsx.OrderedMap, error) {
	query := `select ` + postCols + `
          FROM posts where postId = ?`
	args := []any{jsx.Param(postID)}
	if hostingID != "" {
		query += ` AND hostingId = ?`
		args = append(args, hostingID)
	}
	return db.QueryOne(ctx, q, query, args...)
}

// postTok — setPost 가 쓰는 토큰 값(관리자 화면 · 고객센터의 { key } 흉내)
type postTok struct {
	key       any
	hostingID string
	uid       any // nil 이면 신호에 브라우저 id 칸(zbduid)이 빠진다
}

func tokOf(c *web.Ctx) postTok {
	return postTok{key: c.Token.Key, hostingID: c.Token.HostingID, uid: browserUID(c.Token)}
}

// mysqlDate — Date 를 드라이버가 SQL 에 넣는 모양(한국 시각, 밀리초까지)
func mysqlDate(ms float64) any {
	if math.IsNaN(ms) {
		return nil
	}
	return time.UnixMilli(int64(ms)).In(util.KST).Format("2006-01-02 15:04:05.000")
}

// setPost — 글 저장(있으면 고침) + 시스템 알림 확인 처리 · ES 색인 · 새 글 알림 · 소켓 신호. post 를 고쳐 돌려준다.
func setPost(ctx context.Context, c *web.Ctx, q dbq, tok postTok, post *jsx.OrderedMap, pg *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	now := float64(time.Now().UnixMilli())
	if pg == nil {
		pg = c.App.Settings.PostGroup(orStr(post.V("hostingId")), orStr(post.V("groupKey")))
	}
	if pg == nil {
		pg = jsx.NewOrderedMap()
	}
	var adminID any
	if jsx.Str(pg.V("dataColumn")) == "jsonUser" || !jsx.Truthy(post.V("adminId")) {
		adminID = tok.key
		if !jsx.Truthy(adminID) {
			adminID = float64(0)
		}
	} else {
		adminID = post.V("adminId")
	}
	jsonUser := post.V("jsonUser")
	if !jsx.Truthy(jsonUser) {
		jsonUser = ""
	}
	created := mysqlDate(now)
	if jsx.Truthy(post.V("createdAt")) {
		created = mysqlDate(jsDateMs(post.V("createdAt")))
	}
	P := jsx.Param
	ok, err := db.Exec(ctx, q, `INSERT INTO posts (postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser, status, hostingId) VALUES (?,?,?,?,?,?,?,?,?,?)
        ON DUPLICATE KEY UPDATE adminId = ?, title = ?, json = ?, jsonUser = ?,
        groupKey = ?, groupSubKey = ?, status = ?, createdAt = ?,
  updatedAt = current_timestamp()`,
		P(post.V("postId")), P(adminID), float64(0), P(post.V("groupKey")), P(post.V("groupSubKey")), P(post.V("title")), P(post.V("json")), P(jsonUser), P(post.V("status")), P(post.V("hostingId")),
		P(adminID), P(post.V("title")), P(post.V("json")), P(jsonUser), P(post.V("groupKey")), P(post.V("groupSubKey")), P(post.V("status")), created)
	if err != nil {
		return nil, err
	}
	if jsx.Truthy(post.V("createdAt")) {
		post.Set("updatedAt", now)
	} else {
		post.Set("createdAt", now)
	}
	isNew := db.Affected(ok) == 1

	// 후속 작업 — 실패해도 저장은 그대로(지금 서비스도 로그만 남긴다)
	func() {
		xv, err := parseJSONStr(post.V("json"))
		if err != nil {
			log.Printf("@ 게시글 시스템 메일노티 에러: %v", err)
			return
		}
		x := asMap(xv)
		if jsx.Truthy(pg.V("isSystemNoti")) && jsx.Truthy(x.V("processStatus")) {
			admin, err := getAdminRow(ctx, q, post.V("adminId"))
			if err != nil || admin == nil {
				log.Printf("@ 게시글 시스템 메일노티 에러: %v", err)
				return
			}
			aj := jsx.ParseMap(orStrDef(admin.V("json"), "{}"))
			ids, _ := aj.V("checkedNotiIds").([]any)
			var next []any
			changed := false
			for _, id := range ids {
				if looseEq(id, post.V("postId")) {
					changed = true
				} else {
					next = append(next, id)
				}
			}
			if next == nil {
				next = []any{}
			}
			aj.Set("checkedNotiIds", next)
			if changed {
				admin.Set("json", jsx.Stringify(aj))
				if _, err := updateAdmin(ctx, q, admin); err != nil {
					log.Printf("@ 게시글 시스템 메일노티 에러: %v", err)
					return
				}
			}
		}
		if jsx.Truthy(pg.Path("feature", "useES")) {
			doc := jsx.Assign(nil, post)
			doc.Set("json", x)
			cid := pg.V("customHostingId")
			if !jsx.Truthy(cid) {
				cid = tok.hostingID
			}
			if !jsx.Truthy(cid) {
				cid = post.V("hostingId")
			}
			doc.Set("hostingId", cid)
			doc.Set("adminId", adminID)
			esPutPost(c, doc)
		}
		replies, _ := x.V("replies").([]any)
		lastQuestion := len(replies) > 0 && jsx.Str(asMap(replies[len(replies)-1]).V("type")) == "QUESTION"
		isReQuestion := !jsx.Truthy(post.V("ignoreNoti")) && jsx.Truthy(pg.V("doolayNotiUrl")) && looseEq(x.V("processStatus"), float64(10)) && lastQuestion
		if !isNew && !isReQuestion {
			return
		}
		sub := ""
		if jsx.Truthy(post.V("groupSubKey")) {
			sub = "[" + jsx.Str(post.V("groupSubKey")) + "]"
		}
		title := "[" + tplU(pg, "name") + "]" + sub + " " + tplU(post, "title")
		if jsx.Truthy(pg.V("doolayNotiUrl")) {
			if isReQuestion {
				title = "[" + tplU(pg, "name") + "]" + sub + " [재문의] " + tplU(post, "title")
			}
			msg := "```\n" + title + "\n" + tplU(pg, "notiLinkPrefix") + tplU(post, "postId") + "```"
			sendDooray(msg, jsx.Str(pg.V("doolayNotiUrl")))
		}
	}()

	ev := jsx.NewOrderedMap()
	cid := pg.V("customHostingId")
	if !jsx.Truthy(cid) {
		cid = post.V("hostingId")
	}
	ev.Set("hostingId", cid)
	ev.Set("postId", post.V("postId"))
	if tok.uid != nil {
		ev.Set("zbduid", tok.uid)
		ev.Set("oxuid", tok.uid) // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
	}
	c.App.Emit("doPostUpdate", ev)
	return post, nil
}

// orStrDef — (v || d) 문자열
func orStrDef(v any, d string) string {
	if jsx.Truthy(v) {
		return jsx.Str(v)
	}
	return d
}

// sendDooray — 두레이 웹훅 알림(2초 제한, 실패는 로그만)
func sendDooray(message, url string) {
	if url == "" {
		return
	}
	cl := &http.Client{Timeout: 2 * time.Second}
	body := jsx.Stringify(jsx.Obj("botName", "즈비드", "text", message))
	resp, err := cl.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		log.Printf("@ 두레이 알림 실패: %v", err)
		return
	}
	resp.Body.Close()
}

func adminPostOne(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		cid := hostingID(c)
		postID := c.QV("postId")
		switch c.Method {
		case "GET":
			p, err := getPost(ctx, q, postID, "")
			if err != nil {
				return err
			}
			scbOne(c, p)
		case "POST":
			bd := bodyMap(c)
			rv, err := parseJSONStr(bd.V("json"))
			if err != nil {
				return err
			}
			reqJSON := asMap(rv)
			if !bd.Has("customHostingId") && bd.Has("customCompanyId") { // 옛 이름 companyId — 2026-10 옮기는 기간(본문의 customCompanyId)
				bd.RenameKey("customCompanyId", "customHostingId")
			}
			if !bd.Has("customHostingId") {
				bd.Set("hostingId", cid)
			} else {
				bd.Set("hostingId", bd.V("customHostingId"))
				if cid != "" && reqJSON != nil {
					reqJSON.Set("hostingId", cid)
				}
			}
			pg := c.App.Settings.PostGroup(orStr(bd.V("hostingId")), orStr(bd.V("groupKey")))
			if pg == nil {
				pg = jsx.NewOrderedMap()
			}
			if jsx.Truthy(pg.V("customHostingId")) {
				bd.Set("hostingId", pg.V("customHostingId"))
			}
			dc := pg.V("dataColumn")
			keepMemo := (!jsx.Truthy(dc) || jsx.Str(dc) == "json") && !looseEq(c.Token.HostingID, bd.V("hostingId")) && !jsx.Truthy(reqJSON.V("memo"))
			if keepMemo {
				p, err := getPost(ctx, q, postID, orStr(bd.V("hostingId")))
				if err != nil {
					return err
				}
				var pj any = jsx.NewOrderedMap()
				if p != nil && jsx.Truthy(p.V("json")) {
					if pj, err = parseJSONStr(p.V("json")); err != nil {
						return err
					}
				}
				if m := asMap(pj).V("memo"); jsx.Truthy(m) {
					reqJSON.Set("memo", m)
				}
			}
			cp := jsx.Assign(nil, bd)
			cp.Set("json", jsx.Stringify(rv))
			if _, err := setPost(ctx, c, q, tokOf(c), cp, pg); err != nil {
				return err
			}
			if keepMemo {
				reqJSON.Delete("memo")
			}
			bd.Set("json", jsx.Stringify(rv))
			c.Scb(bd)
		case "DELETE":
			p, err := getPost(ctx, q, postID, cid)
			if err != nil {
				return err
			}
			if p == nil {
				return errMsg("TypeError: Cannot read properties of null (reading 'groupKey')")
			}
			pg := c.App.Settings.PostGroup(cid, orStr(p.V("groupKey")))
			if jsx.Truthy(pg.Path("feature", "useES")) {
				esDelete(c, "posts", jsx.Str(postID))
			}
			ok, err := db.Exec(ctx, q, `DELETE FROM posts WHERE postId = ? AND hostingId = ?`, jsx.Param(postID), cid)
			if err != nil {
				return err
			}
			c.App.Emit("doPostUpdate", jsx.Obj("hostingId", cid, "postId", postID, "zbduid", browserUID(c.Token), "oxuid", browserUID(c.Token) /* 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다 */))
			c.Scb(ok)
		default:
			c.Ecb("잘못된 요청")
		}
		return nil
	})
}

// 회사 글 목록(작성 관리자 이름 포함)
func adminPostIndex(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		cid := hostingID(c)
		groupKey, groupSubKey := c.QV("groupKey"), c.QV("groupSubKey")
		query := `select p.postId, p.adminId, p.userId, p.groupKey, p.groupSubKey, p.title, p.json,
          p.status, UNIX_TIMESTAMP(p.createdAt)*1000 as createdAt, UNIX_TIMESTAMP(p.updatedAt)*1000 as updatedAt, p.hostingId,
          am.name
          FROM posts p
          LEFT JOIN admins am ON p.adminId = am.adminId
          WHERE 1 `
		var args []any
		if cid != "" {
			query += " AND p.hostingId = ? "
			args = append(args, cid)
		}
		if jsx.Truthy(groupKey) {
			query += " AND p.groupKey = ?"
			args = append(args, groupKey)
		}
		if jsx.Truthy(groupSubKey) {
			query += " AND p.groupSubKey = ?"
			args = append(args, groupSubKey)
		}
		rows, err := db.Query(ctx, q, query, args...)
		if err != nil {
			return err
		}
		c.Scb(rows)
		return nil
	})
}

type groupPostSearch struct {
	groupKey  string
	hostingID *string
	postID    any
}

// getAllGroupPost — 게시판 글(페이지 나눔 없이 쓰는 경우만)
func getAllGroupPost(ctx context.Context, q db.Querier, hostingID string, s groupPostSearch) ([]*jsx.OrderedMap, error) {
	where := ` FROM posts WHERE 1 `
	var args []any
	if hostingID != "" {
		// TIP 은 공통("") 팁도 같이 — 공통 팁이 없으면 `IN ()` 문법 오류로 목록 전체가 실패했다(2026-10-03)
		var tipIDs []any
		if s.groupKey == "TIP" {
			tips, err := db.Query(ctx, q, `select postId from posts where hostingId = "" AND groupKey = "TIP"`)
			if err != nil {
				return nil, err
			}
			for _, r := range tips {
				tipIDs = append(tipIDs, r.V("postId"))
			}
		}
		where += ` AND (hostingId = ? `
		args = append(args, hostingID)
		if len(tipIDs) > 0 {
			where += `OR postId in (?)`
			args = append(args, tipIDs)
		}
		where += `) `
	}
	where += " "
	if jsx.Truthy(s.postID) && s.groupKey != "" {
		where += ` AND postId = ? AND groupKey = ?`
		args = append(args, jsx.Param(s.postID), s.groupKey)
	} else if s.groupKey != "" {
		where += ` AND groupKey = ?`
		args = append(args, s.groupKey)
	}
	if s.hostingID != nil {
		where += ` AND hostingId = ?`
		args = append(args, *s.hostingID)
	}
	return db.Query(ctx, q, `select postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
          `+where+`
          `, args...)
}

func adminPostGroup(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		rows, err := getAllGroupPost(ctx, q, hostingID(c), groupPostSearch{groupKey: c.Q("groupKey")})
		if err != nil {
			return err
		}
		c.Scb(rows)
		return nil
	})
}

// 여러 글의 정렬 번호·상태를 한 번에
func adminPostMultiUpdate(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		cid := hostingID(c)
		if c.Method != "POST" {
			c.Ecb("잘못된 요청")
			return nil
		}
		bd := bodyMap(c)
		typ := jsx.StrOrEmpty(bd.V("type"))
		if bd == nil || typ == "" || (typ != "sort" && typ != "status") || !jsx.Truthy(bd.V("postIds")) {
			c.Ecb("잘못된 요청")
			return nil
		}
		ids, _ := bd.V("postIds").([]any)
		posts := []any{}
		for i, x := range ids {
			var postID any = x
			if typ == "status" {
				postID = asMap(x).V("postId")
			}
			post, err := getPost(ctx, q, postID, cid)
			if err != nil {
				return err
			}
			if post == nil {
				return errMsg("TypeError: Cannot read properties of null (reading 'json')")
			}
			if typ == "sort" {
				j, err := parseJSONStr(post.V("json"))
				if err != nil {
					return err
				}
				jm := asMap(j)
				if jm == nil {
					return errMsg("TypeError: Cannot set properties of null (setting 'sortNo')")
				}
				jm.Set("sortNo", float64(i))
				cp := jsx.Assign(nil, post)
				cp.Set("json", jsx.Stringify(jm))
				r, err := setPost(ctx, c, q, tokOf(c), cp, nil)
				if err != nil {
					return err
				}
				posts = append(posts, r)
			} else {
				post.Set("status", asMap(x).V("status"))
				r, err := setPost(ctx, c, q, tokOf(c), post, nil)
				if err != nil {
					return err
				}
				posts = append(posts, r)
			}
		}
		c.Scb(posts)
		return nil
	})
}

// 내 회사 게시판의 since 이후 새 글 수(메뉴 배지)
func adminPostNewCount(c *web.Ctx) {
	groupKey := strings.TrimSpace(orStr(c.QV("groupKey")))
	since := jsNum(c.QV("since"))
	if groupKey == "" || since == 0 {
		c.Ecb("groupKey, since 가 필요합니다.")
		return
	}
	cid := c.Token.HostingID
	if cid == "" {
		c.Scb(jsx.Obj("count", float64(0), "since", since))
		return
	}
	withConn(c, func(ctx context.Context, q dbqConn) error {
		r, err := db.QueryOne(ctx, q, `SELECT COUNT(*) AS cnt FROM posts
        WHERE hostingId = ? AND groupKey = ? AND status >= 0
          AND createdAt > FROM_UNIXTIME(? / 1000)`, cid, groupKey, jsx.Param(since))
		if err != nil {
			return err
		}
		c.Scb(jsx.Obj("count", jsNum(r.V("cnt")), "since", since))
		return nil
	})
}

func adminPostSummary(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		cid := c.Token.HostingID
		query := `select hostingId, groupKey, count(*) as cnt from posts
  `
		var args []any
		if cid != "" {
			query += ` WHERE hostingId = ? `
			args = append(args, cid)
		}
		query += `
  group by hostingId, groupKey`
		rows, err := db.Query(ctx, q, query, args...)
		if err != nil {
			return err
		}
		c.Scb(rows)
		return nil
	})
}

// ─── 댓글 ───

var commentOrderCols = []string{"c.commentId", "c.createdAt", "c.userId", "c.userNickname", "p.title", "p.groupKey", "p.groupSubKey"}

// ─── 고객센터(공지·1:1 문의 — 공통 게시판) ───

var csGroupKeys = []string{"NOTICE", "INQUIRY"}

func csPostGroups(c *web.Ctx) ([]any, error) {
	pgs := asMap(c.App.Settings.Config("").V("POST_GROUPS"))
	if pgs == nil {
		return nil, errMsg("TypeError: Cannot read properties of undefined (reading 'adminPostGroups')")
	}
	out := []any{}
	for _, x := range pgs.Arr("adminPostGroups") {
		g := jsx.Str(asMap(x).V("groupKey"))
		if g == "INQUIRY" || g == "NOTICE" {
			out = append(out, x)
		}
	}
	return out, nil
}

func csFind(list []any, groupKey string) *jsx.OrderedMap {
	for _, x := range list {
		if jsx.Str(asMap(x).V("groupKey")) == groupKey {
			return asMap(x)
		}
	}
	return nil
}

func csBoard(c *web.Ctx) {
	groupKey := c.Q("groupKey")
	if !contains(csGroupKeys, groupKey) {
		c.Ecb("잘못된 요청")
		return
	}
	withConn(c, func(ctx context.Context, q dbqConn) error {
		t := c.Token
		pgs, err := csPostGroups(c)
		if err != nil {
			return err
		}
		empty := ""
		os, err := getAllGroupPost(ctx, q, "", groupPostSearch{groupKey: groupKey, hostingID: &empty})
		if err != nil {
			return err
		}
		items := []any{}
		if groupKey == "NOTICE" {
			for _, o := range os {
				if !looseEq(o.V("status"), float64(0)) {
					continue
				}
				if t.HostingID != "" {
					j, err := parseJSONStr(o.V("json"))
					if err != nil {
						return err
					}
					asMap(j).Delete("memo")
					o.Set("json", j)
					o.Delete("jsonUser")
				}
				items = append(items, o)
			}
		} else {
			pg := csFind(pgs, groupKey)
			gid, gok := parseIntJS(t.GroupID)
			master := false
			if gok {
				for _, x := range pg.Arr("masterGroupIds") {
					if f, ok := x.(float64); ok && f == float64(gid) {
						master = true
					}
				}
			}
			for _, o := range os {
				if !(looseEq(o.V("adminId"), t.Key) || master) {
					continue
				}
				if !master {
					j, err := parseJSONStr(o.V("json"))
					if err != nil {
						return err
					}
					if m := asMap(j); m != nil {
						m.Set("memo", "")
					}
					o.Set("json", j)
				}
				items = append(items, o)
			}
		}
		c.Scb(jsx.Obj("postGroups", pgs, "items", items))
		return nil
	})
}

func csBoardPost(c *web.Ctx) {
	groupKey := c.Q("groupKey")
	if !contains(csGroupKeys, groupKey) {
		c.Ecb("잘못된 요청")
		return
	}
	withConn(c, func(ctx context.Context, q dbqConn) error {
		t := c.Token
		pgs, err := csPostGroups(c)
		if err != nil {
			return err
		}
		empty := ""
		os, err := getAllGroupPost(ctx, q, "", groupPostSearch{groupKey: groupKey, hostingID: &empty, postID: c.QV("postId")})
		if err != nil {
			return err
		}
		if len(os) == 0 {
			c.Ecb("등록된 post 가 없습니다.")
			return nil
		}
		o := os[0]
		jv, err := parseJSONStr(o.V("json"))
		if err != nil {
			return err
		}
		j := asMap(jv)
		o.Set("json", jv)
		if jsx.Truthy(j.V("hostingId")) && t.HostingID != "" && !looseEq(t.HostingID, j.V("hostingId")) {
			c.Ecb("등록된 post 가 없습니다. (호스팅ID 불일치)")
			return nil
		}
		if t.HostingID != "" {
			pv := j.V("pageView")
			if !jsx.Truthy(pv) {
				pv = float64(0)
			}
			j.Set("pageView", jsx.Number(pv)+1)
			cp := jsx.Assign(nil, o)
			cp.Set("json", jsx.Stringify(j))
			cp.Set("ignoreNoti", true)
			if _, err := setPost(ctx, c, q, postTok{key: o.V("adminId")}, cp, nil); err != nil {
				return err
			}
			me, err := getAdminRow(ctx, q, t.Key)
			if err != nil {
				return err
			}
			if me == nil {
				return errMsg("TypeError: Cannot read properties of null (reading 'json')")
			}
			mj := jsx.ParseMap(orStrDef(me.V("json"), "{}"))
			ids, _ := mj.V("checkedNotiIds").([]any)
			if !mj.Has("checkedNotiIds") || !jsx.Truthy(mj.V("checkedNotiIds")) {
				ids = []any{}
			}
			found := false
			for _, x := range ids {
				if strictEq(x, o.V("postId")) {
					found = true
				}
			}
			if !found {
				mj.Set("checkedNotiIds", append(ids, o.V("postId")))
				me.Set("json", jsx.Stringify(mj))
				if _, err := updateAdmin(ctx, q, me); err != nil {
					return err
				}
			}
		}
		o.Delete("jsonUser")
		j.Delete("memo")
		res := jsx.NewOrderedMap()
		if pg := csFind(pgs, groupKey); pg != nil {
			res.Set("postGroup", pg)
		}
		res.Set("item", o)
		c.Scb(res)
		return nil
	})
}
