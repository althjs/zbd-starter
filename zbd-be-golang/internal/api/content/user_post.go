package content

// 사용자 글 — /user/post (내 글) · /user/post/[postId] (읽기·쓰기·지우기) · /user/post/list · /user/post/detail · /user/post/guest/[postId]

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// GET /user/post?groupKey= — 내가 쓴 글
func userPost(c *web.Ctx) {
	query := `select postId, adminId, userId, groupKey, groupSubKey, title,
          status, jsonUser, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt
          FROM posts WHERE userId = ?`
	var key any = tokenKey(c)
	if key == "" {
		key = 0.0 // {key: 0}
	}
	args := []any{key}
	if gk := jsx.StrOrEmpty(c.QV("groupKey")); gk != "" {
		query += " AND groupKey = ?"
		args = append(args, gk)
	}
	rows, err := db.Query(c.Context(), c.App.DB, query, args...)
	if err != nil {
		ecbErr(c, err)
		return
	}
	c.Scb(rowsAny(rows))
}

// ── 공개 목록 개인정보 가리기 (server-libs/mask-personal.ts 와 같은 규칙) ──

var sensitiveKeys = []string{"mobile", "phone", "tel", "email", "mail", "name", "이름", "성명", "birth", "zipcode", "address", "addr", "휴대", "전화", "연락처", "주소", "생년"}
var keepKeys = []string{"emailverified", "addresscount", "filename", "imagename", "companyname", "productname", "hotelname", "bankname", "sitename", "groupname", "username"}
var fileParents = []string{"image", "images", "file", "files", "img", "photo", "attach", "thumbnail"}
var sensitiveHint = regexp.MustCompile(`(?i)mobile|phone|tel|mail|name|birth|zip|addr|휴대|전화|연락|주소|생년|이름|성명`)
var nonDigit = regexp.MustCompile(`\D`)

func isNameKey(k string) bool {
	return k == "name" || strings.HasSuffix(k, "name") || strings.Contains(k, "이름") || strings.Contains(k, "성명")
}

func isSensitiveKey(key, parent string) bool {
	k := strings.ToLower(key)
	for _, x := range keepKeys {
		if x == k {
			return false
		}
	}
	if isNameKey(k) {
		pl := strings.ToLower(parent)
		for _, p := range fileParents {
			if strings.Contains(pl, p) {
				return false
			}
		}
	}
	for _, s := range sensitiveKeys {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

// maskFirst — s[0] + "*" * (s.length - 1) (UTF-16 단위)
func maskFirst(s string) string {
	u := utf16.Encode([]rune(s))
	if len(u) <= 1 {
		return s
	}
	return string(utf16.Decode(u[:1])) + strings.Repeat("*", len(u)-1)
}

func isObj(v any) bool {
	switch v.(type) {
	case *jsx.OrderedMap, []any:
		return true
	}
	return false
}

func maskValue(key string, v any) any {
	if v == nil || isObj(v) {
		return v
	}
	s := jsx.Str(v)
	if trimJS(s) == "" {
		return v
	}
	k := strings.ToLower(key)
	if strings.Contains(k, "email") || strings.Contains(k, "mail") {
		parts := strings.Split(s, "@")
		if len(parts) > 1 && parts[1] != "" {
			u := utf16.Encode([]rune(parts[0]))
			if len(u) > 1 {
				u = u[:1]
			}
			return string(utf16.Decode(u)) + "***@" + parts[1]
		}
		return "***"
	}
	for _, x := range []string{"mobile", "phone", "tel", "휴대", "전화", "연락"} {
		if strings.Contains(k, x) {
			d := nonDigit.ReplaceAllString(s, "")
			if len(d) >= 4 {
				return "***" + d[len(d)-2:]
			}
			return "***"
		}
	}
	if isNameKey(k) {
		return maskFirst(s)
	}
	return "***"
}

func maskPersonalDeep(in any, depth int, parent string) any {
	if depth > 8 || in == nil {
		return in
	}
	switch x := in.(type) {
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = maskPersonalDeep(e, depth+1, parent)
		}
		return out
	case *jsx.OrderedMap:
		out := jsx.NewOrderedMap()
		for _, k := range x.Keys() {
			v := x.V(k)
			if isSensitiveKey(k, parent) && (v == nil || !isObj(v)) {
				out.Set(k, maskValue(k, v))
			} else {
				out.Set(k, maskPersonalDeep(v, depth+1, k))
			}
		}
		return out
	}
	return in
}

// maskPublicPostRow — 공개 목록 한 줄: 작성자 실명·jsonUser 안 연락처를 가린다
func maskPublicPostRow(row *jsx.OrderedMap) *jsx.OrderedMap {
	ju := row.V("jsonUser")
	hasName := jsx.Truthy(row.V("name"))
	jus, isStr := ju.(string)
	if !hasName && (ju == nil || (isStr && !sensitiveHint.MatchString(jus))) {
		return row
	}
	out := jsx.Assign(nil, row)
	if hasName {
		out.Set("name", maskFirst(jsx.Str(out.V("name"))))
	}
	if ju != nil && (!isStr || sensitiveHint.MatchString(jus)) {
		obj := ju
		if isStr {
			v, err := jsx.ParseString(jus)
			if err != nil {
				return out
			}
			obj = v
		}
		masked := maskPersonalDeep(obj, 0, "")
		if isStr {
			out.Set("jsonUser", jsx.Stringify(masked))
		} else {
			out.Set("jsonUser", masked)
		}
	}
	return out
}

// GET /user/post/list?hostingId=&groupKey=&groupSubKey=&mine=1 — 공개 글 목록(남의 글은 개인정보 가림)
func userPostList(c *web.Ctx) {
	t := c.Token
	hostingID := tokenHosting(c)
	if hostingID == "" && t.MemberType != "A" {
		c.Ecb("잘못된 요청")
		return
	}
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	mine := jsx.StrOrEmpty(c.QV("mine")) == "1"
	where := ""
	args := []any{}
	if gk := c.QV("groupKey"); jsx.Truthy(gk) {
		where += `AND p.groupKey = ? `
		args = append(args, gk)
	}
	if gs := c.QV("groupSubKey"); jsx.Truthy(gs) {
		where += `AND p.groupSubKey = ? `
		args = append(args, gs)
	}
	var uid float64
	if mine && tokenKey(c) != "" {
		uid = jsParseIntAny(t.Key)
	}
	if uid != 0 && uid == uid {
		where += `AND p.userId = ? `
		args = append(args, uid)
	}
	where += `AND p.hostingId = ? `
	args = append(args, hostingID)
	status := `p.status = 0`
	if uid != 0 && uid == uid {
		status = `p.status IN (0, 99)`
	}
	rows, err := db.Query(c.Context(), c.App.DB, `SELECT 
      p.postId, p.adminId, p.userId, p.groupKey, p.groupSubKey, 
      p.title, p.json, p.jsonUser, p.status, 
      UNIX_TIMESTAMP(p.createdAt) * 1000 AS createdAt, 
      UNIX_TIMESTAMP(p.updatedAt) * 1000 AS updatedAt, 
      p.hostingId,
      um.name AS name, um.nickname
    FROM posts AS p 
    LEFT JOIN members AS um ON p.userId = um.userId
    WHERE `+status+`
      `+where, args...)
	if err != nil {
		c.Ecb(strings.TrimPrefix(errText(err), "Error: "))
		return
	}
	owner := (mine && tokenKey(c) != "") || t.MemberType == "A"
	out := []any{}
	for _, r := range rows {
		x := jsx.Assign(nil, r)
		x.Set("name", c.App.PII.Dec(r.V("name")))
		if !owner {
			x = maskPublicPostRow(x)
		}
		out = append(out, x)
	}
	c.Scb(out)
}

// GET /user/post/detail?hostingId=&groupKey=&groupSubKey= — 공개 글 하나
func userPostDetail(c *web.Ctx) {
	t := c.Token
	hostingID := tokenHosting(c)
	if hostingID == "" && t.MemberType != "A" {
		c.Ecb("잘못된 요청")
		return
	}
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	where := ""
	args := []any{}
	if gk := c.QV("groupKey"); jsx.Truthy(gk) {
		where += `AND p.groupKey = ? `
		args = append(args, gk)
	}
	if gs := c.QV("groupSubKey"); jsx.Truthy(gs) {
		where += `AND p.groupSubKey = ? `
		args = append(args, gs)
	}
	where += `AND p.hostingId = ? `
	args = append(args, hostingID)
	o, err := db.QueryOne(c.Context(), c.App.DB, `SELECT 
      p.postId, p.adminId, p.userId, p.groupKey, p.groupSubKey, 
      p.title, p.json, p.jsonUser, p.status, 
      UNIX_TIMESTAMP(p.createdAt) * 1000 AS createdAt, 
      UNIX_TIMESTAMP(p.updatedAt) * 1000 AS updatedAt, 
      hostingId 
    FROM posts AS p 
    WHERE p.status = 0 
      `+where+`
    LIMIT 1`, args...)
	if err != nil {
		ecbErr(c, err)
		return
	}
	c.Scb(nilIfNoMap(o))
}

// upsertPost — 글 넣기/고치기(postId 가 같으면 제목·본문·게시판·상태만)
func upsertPost(ctx context.Context, q db.Execer, tokenKeyV any, tokenHostingV string, bd *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	p := func(k string) any { return jsx.Param(bd.V(k)) }
	cid := tokenHostingV
	if cid == "" {
		cid = jsx.StrOrEmpty(bd.V("hostingId"))
	}
	return db.Exec(ctx, q, `INSERT INTO posts (postId, adminId, userId, groupKey, groupSubKey, title, jsonUser, status, hostingId) VALUES (?,?,?,?,?,?,?,?,?) 
  ON DUPLICATE KEY UPDATE title = ?, jsonUser = ?,
  groupKey = ?, groupSubKey = ?, status = ?,
updatedAt = current_timestamp()`,
		p("postId"), 0, tokenKeyV, p("groupKey"), p("groupSubKey"), p("title"), p("jsonUser"), p("status"), cid,
		p("title"), p("jsonUser"), p("groupKey"), p("groupSubKey"), p("status"))
}

// POST /user/post/guest/[postId] — 비회원 글쓰기(지금 서비스 그대로 — 로그인 확인 없음)
func userPostGuest(c *web.Ctx) {
	// 보안(2026-10-03) — 막음: 로그인 없이 어떤 글이든(라이프 기록 포함) 덮어쓸 수 있었다. 부르는 화면이 없다.
	c.Ecb("지원하지 않는 요청입니다.")
	if true {
		return
	}
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	bd, _ := c.Body().(*jsx.OrderedMap)
	if bd == nil {
		c.Ecb("TypeError: Cannot read properties of undefined (reading 'postId')")
		return
	}
	var key any = tokenKey(c)
	if key == "" {
		key = 0
	}
	if _, err := upsertPost(c.Context(), c.App.DB, key, c.Token.HostingID, bd); err != nil {
		ecbErr(c, err)
		return
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	if jsx.Truthy(bd.V("createdAt")) {
		bd.Set("updatedAt", now)
	} else {
		bd.Set("createdAt", now)
	}
	c.Scb(bd)
}

var widgetGrants = []string{"usePostGroup", "postGroupKey", "useSocket", "needLogin", "noShare", "noNotify", "adminMonitor", "monitorEmails", "pageEmbed", "records"}
var storeCategories = []string{"life", "health", "money", "time", "game", "social"}

// pinWidgetGrants — 위젯의 운영자 승인값은 제작자가 못 바꾼다(이전 값으로 되돌림)
func pinWidgetGrants(next, prev any) any {
	s, ok := next.(string)
	if !ok {
		return next
	}
	nv, err := jsx.ParseString(s)
	nm, isMap := nv.(*jsx.OrderedMap)
	if err != nil || !isMap {
		return next
	}
	pm := jsx.NewOrderedMap()
	if pv, err := jsx.ParseString(orStr(prev, "{}")); err == nil {
		if m, ok := pv.(*jsx.OrderedMap); ok {
			pm = m
		}
	}
	for _, k := range widgetGrants {
		if !pm.Has(k) {
			nm.Delete(k)
		} else {
			nm.Set(k, pm.V(k))
		}
	}
	return jsx.Stringify(nm)
}

// cleanWidgetCategory — 스토어 분류는 정해진 key 만(아니면 비운다)
func cleanWidgetCategory(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	j, err := jsx.ParseString(s)
	m, isMap := j.(*jsx.OrderedMap)
	if err != nil || !isMap || !m.Has("category") {
		return v
	}
	for _, k := range storeCategories {
		if looseEq(k, m.V("category")) {
			return v
		}
	}
	m.Set("category", "")
	return jsx.Stringify(m)
}

func getUserPost(ctx context.Context, q db.Querier, postID any) (*jsx.OrderedMap, error) {
	return db.QueryOne(ctx, q, `select postId, adminId, userId, groupKey, groupSubKey, title, jsonUser, json,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt,  UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt
          FROM posts where postId = ?`, postID)
}

// GET·POST·DELETE /user/post/[postId] — 사용자 글(라이프 기록은 주인만)
func userPostByID(c *web.Ctx) {
	t := c.Token
	postID := c.QV("postId")
	bd, _ := c.Body().(*jsx.OrderedMap)
	withConn(c, func(ctx context.Context, conn *sql.Conn) {
		cid := t.HostingID
		if cid == "" {
			cid = jsx.StrOrEmpty(bd.V("hostingId"))
		}
		groups := c.App.Settings.Config(cid).Map("POST_GROUPS")
		// findBoth — adminFirst 면 관리자 게시판부터(쓰기·지우기), 아니면 사용자 게시판부터(읽기)
		findBoth := func(gk any, adminFirst bool) *jsx.OrderedMap {
			a, u := groups.Arr("adminPostGroups"), groups.Arr("userPostGroups")
			if !adminFirst {
				a, u = u, a
			}
			if g := findGroup(a, gk); g != nil {
				return g
			}
			return findGroup(u, gk)
		}
		switch c.Method {
		case "GET":
			o, err := getUserPost(ctx, conn, postID)
			if err != nil {
				ecbErr(c, err)
				return
			}
			if o == nil {
				c.Ecb("등록된 글이 없습니다.")
				return
			}
			if looseEq(o.V("groupKey"), "LIFE") && (tokenKey(c) == "" || jsx.Str(o.V("userId")) != t.Key) {
				c.Ecb("등록된 글이 없습니다.")
				return
			}
			pg := findBoth(o.V("groupKey"), false)
			if pg == nil {
				c.Ecb("등록되지 않은 게시판입니다.")
				return
			}
			if looseEq(pg.V("dataColumn"), "jsonUser") {
				if !jsx.Truthy(pg.V("isAdminReply")) {
					o.Delete("json")
				} else {
					j, err := parseJSONField(o.V("json"))
					if err != nil {
						ecbErr(c, err)
						return
					}
					if jm, ok := j.(*jsx.OrderedMap); ok {
						jm.Delete("memo")
					}
					o.Set("json", j)
				}
			}
			ju, err := jsx.ParseString(jsx.Str(o.V("jsonUser")))
			if err != nil {
				ecbErr(c, err)
				return
			}
			o.Set("jsonUser", ju)
			c.Scb(o)
		case "POST":
			if bd == nil {
				c.Ecb("TypeError: Cannot read properties of undefined (reading 'groupKey')")
				return
			}
			pg := findBoth(bd.V("groupKey"), true)
			if tokenKey(c) == "" && !jsx.Truthy(pg.Path("feature", "useWrite")) {
				c.Ecb("잘못된 요청입니다.")
				return
			}
			if looseEq(pg.V("dataColumn"), "jsonUser") && jsx.Truthy(bd.V("createdAt")) {
				p, err := getUserPost(ctx, conn, postID)
				if err != nil {
					ecbErr(c, err)
					return
				}
				if p == nil {
					c.Ecb("TypeError: Cannot read properties of null (reading 'json')")
					return
				}
				if jsx.Truthy(p.V("json")) {
					bd.Set("json", p.V("json"))
				}
			}
			uid := jsParseInt(orStr(tokenKey(c), "0"))
			bd.Set("userId", uid)
			post, err := getUserPost(ctx, conn, postID)
			if err != nil {
				ecbErr(c, err)
				return
			}
			// 지금 서비스는 로그인 없이도(userId 0) 작성자가 0 인 글(관리자 글)을 덮을 수 있었다 — 막는다
			if post != nil && (!looseEq(post.V("userId"), uid) || tokenKey(c) == "") {
				c.Ecb("본인이 작성한 글이 아닙니다.")
				return
			}
			// 지금 서비스는 저장을 본문의 postId 로 해서, 주소의 글만 확인하고 남의 글을 덮을 수 있었다 — 본문 글도 확인한다
			if bpid := bd.V("postId"); jsx.Truthy(bpid) && !looseEq(bpid, postID) {
				other, err := getUserPost(ctx, conn, jsx.Param(bpid))
				if err != nil {
					ecbErr(c, err)
					return
				}
				if other != nil && (!looseEq(other.V("userId"), uid) || tokenKey(c) == "") {
					c.Ecb("본인이 작성한 글이 아닙니다.")
					return
				}
			}
			if looseEq(bd.V("groupKey"), "WIDGET") {
				bd.Set("jsonUser", pinWidgetGrants(bd.V("jsonUser"), post.V("jsonUser")))
				bd.Set("jsonUser", cleanWidgetCategory(bd.V("jsonUser")))
			}
			var key any = tokenKey(c)
			if key == "" {
				key = 0
			}
			if _, err := upsertPost(ctx, conn, key, t.HostingID, bd); err != nil {
				ecbErr(c, err)
				return
			}
			if looseEq(pg.V("dataColumn"), "jsonUser") {
				bd.Delete("json")
			}
			if jsx.Truthy(bd.V("createdAt")) {
				bd.Set("updatedAt", nowMs())
			} else {
				bd.Set("createdAt", nowMs())
			}
			if jsx.Truthy(pg.Path("feature", "useES")) {
				x := jsx.Assign(nil, bd)
				x.Set("hostingId", t.HostingID)
				putEsPost(ctx, c, x)
			}
			if looseEq(bd.V("groupKey"), "LIFE") {
				notifyLifeChanged(c, uid, bd.V("groupSubKey"), "update")
			}
			c.Scb(bd)
		case "DELETE":
			post, err := getUserPost(ctx, conn, postID)
			if err != nil {
				ecbErr(c, err)
				return
			}
			var owner any
			if post != nil {
				owner = post.V("userId")
			}
			var key any
			if tokenKey(c) != "" {
				key = t.Key
			} else {
				key = 0.0
			}
			// 지금 서비스는 로그인 없이도 작성자가 0 인 글(관리자 글)을 지울 수 있었다 — 로그인을 요구한다
			if !looseEq(owner, key) || tokenKey(c) == "" {
				c.Ecb("삭제 권한이 없습니다.")
				return
			}
			o, err := db.Exec(ctx, conn, `DELETE FROM posts WHERE postId = ?`, postID)
			if err != nil {
				ecbErr(c, err)
				return
			}
			pg := findBoth(post.V("groupKey"), true)
			if jsx.Truthy(pg.Path("feature", "useES")) {
				deleteEsPost(ctx, c, jsx.Str(postID))
			}
			if looseEq(post.V("groupKey"), "LIFE") {
				notifyLifeChanged(c, jsx.Number(post.V("userId")), post.V("groupSubKey"), "delete")
			}
			c.Scb(o)
		default:
			c.Ecb("잘못된 요청")
		}
	})
}
