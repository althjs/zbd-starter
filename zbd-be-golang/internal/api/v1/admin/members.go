package admin

// 회원 목록 — GET /api/v1/admin/members (감사표 1-1·1-2)
//
//	?page&size&q&sort=-userId|createdAt|nickname|status&status=0,-1&groupId=&type=GENERAL|…&authGroupId=&agree=MARKETING:EMAIL,TERMS&exclude=1,2
//
// q 는 개인정보 규칙(v1x.MemberSearch): 이메일·전화·실명은 전체 정확 일치, 닉네임은 앞부분.
// GET /api/v1/admin/members/ids — 같은 조건의 회원 번호만(「검색 결과 전체 선택」, 최대 5000).
// GET /api/v1/admin/members/[userId]/posts · /comments — 그 회원이 쓴 글·댓글(쪽).

import (
	"regexp"
	"strings"

	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

const memberOrig = "/admin/user"

var memberSort = map[string]string{
	"userId":    "userId",
	"createdAt": "createdAt",
	"updatedAt": "updatedAt",
	"nickname":  "nickname",
	"status":    "status",
}

var reAgreeKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

// memberWhere — 목록·번호 목록이 같이 쓰는 조건. 잘못된 값이면 400 을 보내고 false.
func memberWhere(c *web.Ctx, cid, q string) (*v1x.Where, bool) {
	w := &v1x.Where{}
	w.Add("hostingId = ?", cid)
	if s := c.Q("status"); s != "" {
		ids := v1x.Ints(s)
		if len(ids) == 0 {
			c.Fail(400, web.ErrBadRequest, "status 는 정수(쉼표로 여럿)")
			return nil, false
		}
		w.Add("status IN (?)", ids)
	}
	if s := c.Q("groupId"); s != "" {
		ids := v1x.Ints(s)
		if len(ids) == 0 {
			c.Fail(400, web.ErrBadRequest, "groupId 는 정수")
			return nil, false
		}
		w.Add("groupId IN (?)", ids)
	}
	switch t := c.Q("type"); {
	case t == "":
	case t == "GENERAL":
		w.Add("JSON_EXTRACT(json, '$.type') IS NULL")
	default:
		w.Add("JSON_UNQUOTE(JSON_EXTRACT(json, '$.type')) = ?", t)
	}
	if s := c.Q("authGroupId"); s != "" {
		w.Add("COALESCE(JSON_UNQUOTE(JSON_EXTRACT(json, '$.auth.authGroupId')), '1') = ?", s)
	}
	// 약관 동의: K = 그 약관에 동의(참이거나, 하위 항목 중 하나라도 참) · K:S = 하위 항목 S 에 동의. 여럿은 모두(AND).
	for _, a := range v1x.Strs(c.Q("agree")) {
		k, sub, hasSub := strings.Cut(jsx.Str(a), ":")
		if !reAgreeKey.MatchString(k) || (hasSub && !reAgreeKey.MatchString(sub)) {
			c.Fail(400, web.ErrBadRequest, "agree 는 약관키 또는 약관키:항목키")
			return nil, false
		}
		path := `$.agreeMap."` + k + `"`
		if hasSub {
			w.Add("JSON_UNQUOTE(JSON_EXTRACT(json, ?)) = 'true'", path+`."`+sub+`"`)
		} else {
			w.Add("(JSON_UNQUOTE(JSON_EXTRACT(json, ?)) = 'true' OR JSON_CONTAINS(JSON_EXTRACT(json, ?), 'true'))", path, path+".*")
		}
	}
	if s := c.Q("exclude"); s != "" {
		if ids := v1x.Ints(s); len(ids) > 0 {
			w.Add("userId NOT IN (?)", ids)
		}
	}
	v1x.MemberSearch(w, c.App.PII, "", q)
	return w, true
}

func members(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, memberSort, "-userId", "userId")
	if !ok {
		return
	}
	cid := c.Token.HostingID
	w, ok := memberWhere(c, cid, lq.Q)
	if !ok {
		return
	}
	rows, total, err := v1x.Page(c.Context(), c.App.DB, `SELECT userId, groupId, name, email, mobile, nickname, image, description, status, socialId,
  UNIX_TIMESTAMP(createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(updatedAt)*1000 AS updatedAt,
  JSON_UNQUOTE(JSON_EXTRACT(json, '$.type')) AS type,
  JSON_UNQUOTE(JSON_EXTRACT(json, '$.auth.authGroupId')) AS authGroupId,
  JSON_EXTRACT(json, '$.agreeMap') AS agreeMap`, " FROM members", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.App.PII.DecMembers(rows)
	for _, r := range rows {
		am := jsx.ParseMap(jsx.StrOrEmpty(r.V("agreeMap")))
		if am == nil {
			r.Set("agreeMap", jsx.NewOrderedMap())
		} else {
			r.Set("agreeMap", am)
		}
		if !jsx.Truthy(r.V("authGroupId")) {
			r.Set("authGroupId", float64(1))
		} else {
			r.Set("authGroupId", jsx.Number(r.V("authGroupId")))
		}
	}
	c.SendList(v1x.Any(rows), total, lq)
}

const memberIDsMax = 5000

func memberIDs(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	cid := c.Token.HostingID
	w, ok := memberWhere(c, cid, strings.TrimSpace(c.Q("q")))
	if !ok {
		return
	}
	cnt, err := db.Query(c.Context(), c.App.DB, "SELECT COUNT(*) AS n FROM members"+w.SQL(), w.Args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	rows, err := db.Query(c.Context(), c.App.DB, "SELECT userId FROM members"+w.SQL()+" ORDER BY userId DESC LIMIT ?", append(w.Args, memberIDsMax)...)
	if err != nil {
		c.Ecb(err)
		return
	}
	ids := make([]any, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.V("userId"))
	}
	total := jsx.Number(cnt[0].V("n"))
	c.Scb(ids, jsx.Obj("total", total, "truncated", total > memberIDsMax))
}

var memberPostSort = map[string]string{"createdAt": "createdAt", "updatedAt": "updatedAt", "postId": "postId", "groupKey": "groupKey"}

// memberPosts — 그 회원이 쓴 글(라이프·앱 데이터 같은 개인 기록도 groupKey 로 고를 수 있다 — 관리자 화면)
func memberPosts(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, memberPostSort, "-createdAt", "postId")
	if !ok {
		return
	}
	uid, ok := v1x.IntParam(c, "userId", 0)
	if !ok {
		return
	}
	w := &v1x.Where{}
	w.Add("hostingId = ?", c.Token.HostingID)
	w.Add("userId = ?", uid)
	if gk := v1x.Strs(c.Q("groupKey")); len(gk) > 0 {
		w.Add("groupKey IN (?)", gk)
	}
	if lq.Q != "" {
		w.Add("title LIKE ?", web.LikeContains(lq.Q))
	}
	rows, total, err := v1x.Page(c.Context(), c.App.DB, `SELECT postId, groupKey, groupSubKey, title, status,
  UNIX_TIMESTAMP(createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(updatedAt)*1000 AS updatedAt`, " FROM posts", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.SendList(v1x.Any(rows), total, lq)
}

var memberCommentSort = map[string]string{"createdAt": "c.createdAt", "commentId": "c.commentId"}
