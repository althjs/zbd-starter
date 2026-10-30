// Package pub — /api/v1 공개 화면 목록(게시판·인기글·스토어·평가·글 댓글·미니홈피 댓글). 쪽·검색·정렬을 서버에서.
//
//	GET /api/v1/boards/[groupKey]/posts?page&size&q&groupSubKey&cat.{경로}=값&sort   (감사표 1-13·1-14 — 위젯 「최근 글 N개」 는 size=N)
//	GET /api/v1/boards/trending?tab=1h|6h|24h|3d|7d&groupKey&page&size               (1-15)
//	GET /api/v1/store/apps?page&size&q&category&sort=-updatedAt                        (1-16, 3-2)
//	GET /api/v1/store/apps/[postId]/reviews?page&size&hasText=1&sort=-updatedAt|-score (2-11)
//	GET /api/v1/posts/[postId]/comments?size&before                                   (2-8, 커서)
//	GET /api/v1/minihome/[handle]/comments?target=app:{앱}|{ymd}:{memoId}&size&before  (2-7, 주인·친구만)
package pub

import (
	"regexp"
	"strings"

	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

func Register(r *web.Router) {
	r.Handle("/boards/[groupKey]/posts", web.Cors, v1x.Public(boardPosts))
}

// viewer — 로그인한 회원이면 그 번호(토큰을 발급받은 브라우저일 때만), 아니면 ""
func viewer(c *web.Ctx) string {
	t := c.Token
	if !t.IsUser() {
		return ""
	}
	uid := c.BrowserUID()
	if uid == "" {
		uid = "UNKNOWN"
	}
	if uid != t.BrowserUID {
		return ""
	}
	return t.Key
}

// boardInfo — 공개 목록을 줄 수 있는 게시판인가. 관리자 게시판은 isBoardStore(공지·FAQ·약관 …),
// 회원 게시판은 글쓰기 게시판(feature.useWrite)만 — 라이프·AI 대화·문의 같은 개인 기록은 공개 목록이 없다.
type boardInfo struct {
	ok, ownerOnly, useSort bool
	dataColumn             string
}

func board(c *web.Ctx, cid, groupKey string) boardInfo {
	pg, _ := c.App.Settings.Meta(cid, "POST_GROUPS").(*jsx.OrderedMap)
	for _, x := range pg.Arr("adminPostGroups") {
		g, _ := x.(*jsx.OrderedMap)
		if g != nil && jsx.Str(g.V("groupKey")) == groupKey && jsx.Truthy(g.V("isBoardStore")) {
			return boardInfo{ok: true, ownerOnly: jsx.Truthy(g.Path("feature", "useOwnerOnly")), useSort: jsx.Truthy(g.V("useSort")), dataColumn: dataCol(g)}
		}
	}
	for _, x := range pg.Arr("userPostGroups") {
		g, _ := x.(*jsx.OrderedMap)
		if g != nil && jsx.Str(g.V("groupKey")) == groupKey && jsx.Truthy(g.Path("feature", "useWrite")) {
			return boardInfo{ok: true, ownerOnly: jsx.Truthy(g.Path("feature", "useOwnerOnly")), useSort: jsx.Truthy(g.V("useSort")), dataColumn: dataCol(g)}
		}
	}
	return boardInfo{}
}

func dataCol(g *jsx.OrderedMap) string {
	if jsx.Str(g.V("dataColumn")) == "jsonUser" {
		return "jsonUser"
	}
	return "json"
}

var reGroupKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,20}$`)
var reCatPath = regexp.MustCompile(`^[A-Za-z0-9_]{1,30}(\.[A-Za-z0-9_]{1,30}){0,3}$`)
var reTags = regexp.MustCompile(`<[^>]*>|&nbsp;|[#*>\x60_~\[\]()]+`)
var reSpaces = regexp.MustCompile(`\s+`)

// reLang — 목록 ?lang= (사이트 언어 키 en·ja·cn·fr 같은 두 글자)
var reLang = regexp.MustCompile(`^[a-z]{2}$`)

// postLang — 글의 그 언어 제목·본문(관리자 글 편집의 언어 탭 = 데이터 칸 i18n.{언어}). 번역이 빈 칸은 기본 글 — 화면 shared/post-i18n.ts 와 같은 규칙
func postLang(title, body, i18nJSON string) (string, string) {
	if i18nJSON == "" {
		return title, body
	}
	t, _ := jsx.ParseOr(i18nJSON, nil).(*jsx.OrderedMap)
	if t == nil {
		return title, body
	}
	if s := t.S("title"); strings.TrimSpace(s) != "" {
		title = s
	}
	if s := t.S("body"); strings.TrimSpace(s) != "" {
		body = s
	}
	return title, body
}

// viewsSQL — 글 조회수. post_views 는 글·시각(시간 단위)마다 한 줄이라 글마다 합친다
// (예전 join 은 groupKey 칸을 게시판으로 맞춰 늘 0 이었다 — 그 칸엔 「2026-09-25 10:00」 같은 시각이 들어 있다)
const viewsSQL = "(SELECT IFNULL(SUM(x.pv), 0) FROM post_views x WHERE x.postId = p.postId)"

func boardSort(col string) map[string]string {
	return map[string]string{
		"createdAt": "p.createdAt",
		"updatedAt": "p.updatedAt",
		"postId":    "p.postId",
		"title":     "p.title",
		"sortNo":    "COALESCE(CAST(JSON_UNQUOTE(JSON_EXTRACT(p.json, '$.sortNo')) AS DECIMAL(20,4)), 999999999)",
		"fixTop":    "IF(JSON_UNQUOTE(JSON_EXTRACT(p.json, '$.isFixTop')) = 'true', 1, 0)",
		"views":     viewsSQL,
	}
}

func boardPosts(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	gk := c.Q("groupKey")
	cid := v1x.PublicHosting(c)
	if !reGroupKey.MatchString(gk) {
		c.Fail(404, web.ErrNotFound, "없는 게시판입니다")
		return
	}
	b := board(c, cid, gk)
	if !b.ok {
		c.Fail(404, web.ErrNotFound, "없는 게시판입니다")
		return
	}
	def := "-fixTop,-createdAt"
	if b.useSort {
		def = "sortNo,-createdAt"
	}
	lq, ok := v1x.List(c, boardSort(b.dataColumn), def, "p.postId")
	if !ok {
		return
	}
	me := viewer(c)
	w := &v1x.Where{}
	w.Add("p.hostingId = ?", cid)
	w.Add("p.groupKey = ?", gk)
	w.Add("p.status = 0")
	if b.ownerOnly {
		// 글쓴이만 보는 게시판(1:1 문의 같은 것) — 로그인한 회원의 글만
		if me == "" {
			c.Fail(401, web.ErrUnauthorized, "로그인이 필요합니다")
			return
		}
		w.Add("p.userId = ?", me)
	}
	if s := c.Q("groupSubKey"); s != "" {
		w.Add("p.groupSubKey IN (?)", v1x.Strs(s))
	}
	// 분류 필터: cat.{경로}=값 → 본문 JSON 의 그 경로(값 100자 이내)
	for _, k := range c.Query.Keys() {
		if !strings.HasPrefix(k, "cat.") {
			continue
		}
		path, val := strings.TrimPrefix(k, "cat."), c.Q(k)
		if !reCatPath.MatchString(path) || len([]rune(val)) > 100 {
			c.Fail(400, web.ErrBadRequest, "분류 필터가 올바르지 않습니다: "+k)
			return
		}
		w.Add("JSON_UNQUOTE(JSON_EXTRACT(p."+b.dataColumn+", ?)) = ?", "$."+path, val)
	}
	if lq.Q != "" {
		like := web.LikeContains(lq.Q)
		w.Add("(p.title LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(p."+b.dataColumn+", '$.body')) LIKE ?)", like, like)
	}
	ctx := c.Context()
	// ?lang= 이면 그 언어 제목·본문도(번역이 없으면 기본 글) — 언어 키는 위 정규식만 통과해 SQL 경로에 그대로 넣는다
	langSQL := ""
	if l := c.Q("lang"); l != "" && l != "ko" && reLang.MatchString(l) {
		langSQL = "JSON_EXTRACT(p." + b.dataColumn + ", '$.i18n." + l + "') AS i18n,\n  "
	}
	rows, total, err := v1x.Page(ctx, c.App.DB, `SELECT p.postId, p.groupKey, p.groupSubKey, p.title, p.userId, `+langSQL+`
  UNIX_TIMESTAMP(p.createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(p.updatedAt)*1000 AS updatedAt,
  JSON_EXTRACT(p.json, '$.isFixTop') AS isFixTop,
  JSON_EXTRACT(p.`+b.dataColumn+`, '$.image') AS image,
  LEFT(JSON_UNQUOTE(JSON_EXTRACT(p.`+b.dataColumn+`, '$.body')), 600) AS body,
  `+viewsSQL+` AS viewCount,
  0 AS commentCount,
  0 AS likeCount`,
		" FROM posts p", w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	items, err := postCards(c, cid, rows)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.SendList(items, total, lq)
}

// postCards — 목록 한 줄: 본문은 앞 100자 글자만(excerpt), 글쓴이는 닉네임·사진
func postCards(c *web.Ctx, cid string, rows []*jsx.OrderedMap) ([]any, error) {
	var uids []any
	for _, r := range rows {
		if jsx.Truthy(r.V("userId")) {
			uids = append(uids, r.V("userId"))
		}
	}
	people := map[string]*jsx.OrderedMap{}
	if len(uids) > 0 {
		ms, err := db.Query(c.Context(), c.App.DB, `SELECT userId, nickname, image FROM members WHERE hostingId = ? AND userId IN (?)`, cid, uids)
		if err != nil {
			return nil, err
		}
		for _, m := range ms {
			people[jsx.Str(m.V("userId"))] = m
		}
	}
	items := make([]any, 0, len(rows))
	for _, r := range rows {
		title, body := postLang(jsx.StrOrEmpty(r.V("title")), jsx.StrOrEmpty(r.V("body")), jsx.StrOrEmpty(r.V("i18n")))
		ex := strings.TrimSpace(reSpaces.ReplaceAllString(reTags.ReplaceAllString(body, " "), " "))
		var writer any
		if jsx.Truthy(r.V("userId")) {
			p := people[jsx.Str(r.V("userId"))]
			writer = jsx.Obj("userId", r.V("userId"), "nickname", v1x.Nick(p.V("nickname"), r.V("userId")), "image", v1x.OrEmpty(p.V("image")))
		}
		var image any
		if s := jsx.StrOrEmpty(r.V("image")); s != "" {
			image = jsx.ParseOr(s, nil)
		}
		items = append(items, jsx.Obj(
			"postId", r.V("postId"), "groupKey", r.V("groupKey"), "groupSubKey", r.V("groupSubKey"), "title", title,
			"excerpt", v1x.Cut(ex, 100), "image", image, "isFixTop", jsx.StrOrEmpty(r.V("isFixTop")) == "true",
			"writer", writer, "viewCount", jsx.Number(r.V("viewCount")), "commentCount", jsx.Number(r.V("commentCount")),
			"likeCount", jsx.Number(r.V("likeCount")),
			"createdAt", r.V("createdAt"), "updatedAt", r.V("updatedAt"),
		))
	}
	return items, nil
}

var trendTabs = map[string]int{"1h": 1, "6h": 6, "24h": 24, "3d": 72, "7d": 168}
