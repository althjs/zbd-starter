package content

// 작은 조회·저장 라우트 — QR 코드 · 동영상 목록 · 약관 · 타자연습 기록 · 회원 중복 확인 · 요약 · 북마크

import (
	"context"
	"database/sql"
	"net"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// isOffice — 사무실·서버 자신에서 온 요청인가.
// Next 는 x-forwarded-for 가 없으면 접속 주소로 채운다 — 같게 하려고 헤더가 없을 때 접속 주소를 넣어 본다.
func isOffice(c *web.Ctx) bool {
	h := c.R.Header
	if h.Get("x-forwarded-for") == "" {
		h = h.Clone()
		host, _, err := net.SplitHostPort(c.R.RemoteAddr)
		if err != nil {
			host = c.R.RemoteAddr
		}
		h.Set("x-forwarded-for", host)
	}
	return util.IsOffice(h, util.OfficeIPs(c.App.Cfg.Get("OFFICE_IPS")))
}

// GET /user/terms?hostingId= — 약관(공개 글만, 이력·메모 빼고)
func userTerms(c *web.Ctx) {
	hostingID := jsx.StrOrEmpty(c.QV("hostingId"))
	if hostingID == "" {
		c.Ecb("잘못된 요청")
		return
	}
	rows, err := db.Query(c.Context(), c.App.DB, `select postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
           FROM posts WHERE 1  AND (hostingId = ?  )   AND groupKey = ?`, hostingID, "TERMS")
	if err != nil {
		c.Ecb(err)
		return
	}
	out := []any{}
	for _, o := range rows {
		if jsx.Number(o.V("status")) != 0 {
			continue
		}
		js, _ := jsx.ParseString(jsx.Str(o.V("json")))
		x := jsx.Obj("postId", o.V("postId"), "title", o.V("title"), "groupSubKey", o.V("groupSubKey"))
		if m, ok := js.(*jsx.OrderedMap); ok {
			m.Delete("history")
			m.Delete("memo")
			jsx.Assign(x, m)
		}
		out = append(out, x)
	}
	c.Scb(out)
}

// nilIfNoMap — (*OrderedMap)(nil) 을 any(nil) 로(JSON null)
func nilIfNoMap(m *jsx.OrderedMap) any {
	if m == nil {
		return nil
	}
	return m
}

// POST /user/[hostingId]/check-duplicated — 이메일·닉네임·휴대폰 중복 확인
func userCheckDuplicated(c *web.Ctx) {
	bd := c.BodyMap()
	hostingID := c.Q("hostingId")
	email, nick, mobile := bd.V("email"), bd.V("nickname"), bd.V("mobileNo")
	if !jsx.Truthy(email) && !jsx.Truthy(nick) && !jsx.Truthy(mobile) {
		c.Ecb("잘못된 요청")
		return
	}
	ctx := c.Context()
	count := func(col string, v any) (bool, error) {
		r, err := db.QueryOne(ctx, c.App.DB, `SELECT count(*) as cnt FROM members WHERE hostingId = ? AND groupId = 0 AND `+col+` = ?`, hostingID, v)
		if err != nil {
			return false, err
		}
		return jsx.Number(r.V("cnt")) > 0, nil
	}
	var (
		dup      bool
		err      error
		bad, okm string
	)
	switch {
	case jsx.Truthy(email):
		dup, err = count("email_hash", c.App.PII.Hash("email", email))
		bad, okm = "이미 사용중인 이메일입니다.", "사용 가능한 이메일입니다."
	case jsx.Truthy(nick):
		dup, err = count("nickname", jsx.Param(nick))
		bad, okm = "이미 사용중인 닉네임입니다.", "사용 가능한 닉네임입니다."
	default:
		dup, err = count("mobile_hash", c.App.PII.Hash("mobile", mobile))
		bad, okm = "이미 사용중인 휴대폰 번호입니다.", "사용 가능한 휴대폰 번호입니다."
	}
	if err != nil {
		c.Ecb(err)
		return
	}
	if dup {
		c.Ecb(bad)
	} else {
		c.Scb(okm)
	}
}

// GET /user/[hostingId]/summary — 마이페이지 요약(글·댓글 수)
func userSummary(c *web.Ctx) {
	hostingID := c.Q("hostingId")
	summary := jsx.Obj(
		"purchangeCount", 0.0, "reserveCount", 0.0, "reviewCount", 0.0, "likeCount", 0.0,
		"couponCount", 0.0, "couponAssignedCount", 0.0, "commerceOrderCount", 0.0, "siteCount", 0.0,
		"postCount", 0.0, "postBookmarkCount", 0.0, "commentCount", 0.0,
	)
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	userID := tokenKey(c)
	if userID == "" || c.Token.MemberType != "U" {
		c.Ecb("잘못된 요청")
		return
	}
	{
		ctx := c.Context()
		r, err := db.QueryOne(ctx, c.App.DB, `SELECT count(*) as cnt from posts WHERE hostingId = ? AND groupKey != 'CS'  AND userId = ?`, hostingID, userID)
		if err != nil {
			c.Ecb(err)
			return
		}
		summary.Set("postCount", orZero(r.V("cnt")))
		cc, err := userCommentCount(ctx, c.App.DB, hostingID, userID)
		if err != nil {
			c.Ecb(err)
			return
		}
		summary.Set("commentCount", cc)
	}
	c.Scb(summary)
}

// orZero — x || 0
func orZero(v any) any {
	if !jsx.Truthy(v) {
		return 0.0
	}
	return v
}

// PUT /user/bookmark — 상품·글 북마크 넣기/빼기(회원 json 의 목록)
func userBookmark(c *web.Ctx) {
	bd := c.BodyMap()
	withConn(c, func(ctx context.Context, conn *sql.Conn) {
		if c.Method != "PUT" {
			c.Ecb("잘못된 요청")
			return
		}
		userID := tokenKey(c)
		typ := jsx.StrOrEmpty(bd.V("type"))
		keyOf := map[string]string{"TOUR": "productIds", "COMMERCE": "commerceProductIds", "POST": "postIds"}
		key, okType := keyOf[typ]
		if userID == "" || c.Token.MemberType != "U" || !jsx.Truthy(bd.V("id")) || !okType || !jsx.Truthy(bd.V("action")) {
			c.Ecb("잘못된 요청")
			return
		}
		m, err := member.Get(ctx, c.App, conn, userID)
		if err != nil {
			c.Ecb(err)
			return
		}
		if m == nil {
			c.Ecb("Cannot read properties of null (reading 'json')")
			return
		}
		js := jsx.ParseMap(jsx.StrOrEmpty(m.V("json")))
		ids := []any{}
		for _, id := range js.Arr(key) {
			if !looseEq(id, bd.V("id")) {
				ids = append(ids, id)
			}
		}
		if jsx.StrOrEmpty(bd.V("action")) == "ADD" {
			ids = append(ids, bd.V("id"))
		}
		js.Set(key, ids)
		m.Set("json", jsx.Stringify(js))
		if _, err := member.Update(ctx, c.App, conn, m); err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(ids)
	})
}

// looseEq — JS 의 a == b (숫자·문자열·null 만)
func looseEq(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	_, as := a.(string)
	_, bs := b.(string)
	if as && bs {
		return a.(string) == b.(string)
	}
	if ab, ok := a.(bool); ok {
		a = map[bool]float64{true: 1, false: 0}[ab]
	}
	if bb, ok := b.(bool); ok {
		b = map[bool]float64{true: 1, false: 0}[bb]
	}
	switch a.(type) {
	case string, float64:
	default:
		return false
	}
	switch b.(type) {
	case string, float64:
	default:
		return false
	}
	return jsx.Number(a) == jsx.Number(b)
}
