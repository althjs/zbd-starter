package content

// 게시판(공개) — /board/[hostingId] 목록·조회수·여러 글, /board/[hostingId]/[postId] 상세,
// 댓글 쓰기·재문의·답글 목록·주간 인기 글.

import (
	"context"
	"database/sql"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// jsParseInt — JS 의 parseInt(s, 10). 숫자로 시작하지 않으면 NaN.
func jsParseInt(s string) float64 {
	s = trimJS(s)
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	j := i
	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	if j == i {
		return math.NaN()
	}
	f, _ := strconv.ParseFloat(s[:j], 64)
	return f
}

// jsParseIntAny — parseInt(v + "", 10) (값이 없으면 "undefined" → NaN)
func jsParseIntAny(v any) float64 {
	if v == nil {
		return math.NaN()
	}
	return jsParseInt(jsx.Str(v))
}

// utf16Len — JS 의 s.length
func utf16Len(s string) int { return len(utf16.Encode([]rune(s))) }

// findGroup — 게시판 목록에서 groupKey 가 같은 첫 게시판
func findGroup(groups []any, groupKey any) *jsx.OrderedMap {
	for _, g := range groups {
		if gm, ok := g.(*jsx.OrderedMap); ok && looseEq(gm.V("groupKey"), groupKey) {
			return gm
		}
	}
	return nil
}

var categoryPathRe = regexp.MustCompile(`^[A-Za-z0-9_]+(\.[A-Za-z0-9_]+)*$`)

var listKeepKeys = []string{"processStatus", "commentCount", "viewCount", "isFixTop", "thumbnail", "name"}

func pickForList(src any) *jsx.OrderedMap {
	out := jsx.NewOrderedMap()
	m, _ := src.(*jsx.OrderedMap)
	for _, k := range listKeepKeys {
		if m != nil && m.Has(k) {
			out.Set(k, m.V(k))
		}
	}
	return out
}

// maskField — json·jsonUser 를 목록에 필요한 칸만 남긴다(문자열이면 문자열로)
func maskField(o *jsx.OrderedMap, k string) error {
	v := o.V(k)
	if s, ok := v.(string); ok {
		p, err := parseJSONField(s)
		if err != nil {
			return err
		}
		o.Set(k, jsx.Stringify(pickForList(p)))
		return nil
	}
	if !jsx.Truthy(v) {
		v = nil
	}
	o.Set(k, pickForList(v))
	return nil
}

func boardList(c *web.Ctx) {
	hostingID := c.Q("hostingId")
	groupKey := jsx.StrOrEmpty(c.QV("groupKey"))
	typ := jsx.StrOrEmpty(c.QV("type"))
	keyword := trimJS(jsx.StrOrEmpty(c.QV("keyword")))
	var postIDs any = c.QV("postIds")
	if !jsx.Truthy(postIDs) {
		postIDs = ""
	}
	keywordType := jsx.StrOrEmpty(c.QV("keywordType"))
	categoryFilter := jsx.NewOrderedMap()
	if cf, err := jsx.ParseString(orStr(c.QV("categoryFilter"), "{}")); err == nil {
		if m, ok := cf.(*jsx.OrderedMap); ok {
			for _, k := range m.Keys() {
				if s, ok := m.V(k).(string); ok && categoryPathRe.MatchString(k) && utf16Len(s) <= 100 {
					categoryFilter.Set(k, s)
				}
			}
		}
	}
	userID := jsParseIntAny(tokenKey(c))
	if math.IsNaN(userID) {
		userID = 0
	}
	withConn(c, func(ctx context.Context, conn *sql.Conn) {
		if keywordType == "COMMENT" && keyword != "" {
			ids, err := doCommentSearchPostIDs(ctx, conn, hostingID, keyword)
			if err != nil {
				ecbErr(c, err)
				return
			}
			postIDs = ids
		}
		meta := c.App.Settings.Config(hostingID)
		pg := meta.Map("POST_GROUPS")
		if pg == nil {
			c.Ecb("TypeError: Cannot read properties of undefined (reading 'adminPostGroups')")
			return
		}
		adminGroups, userGroups := pg.Arr("adminPostGroups"), pg.Arr("userPostGroups")
		var groupKeys []any
		switch {
		case groupKey != "":
			groupKeys = []any{groupKey}
		case userID != 0 && typ == "my":
			groupKeys = []any{}
			for _, g := range userGroups {
				gm, _ := g.(*jsx.OrderedMap)
				if jsx.Str(gm.V("groupKey")) == "CS" {
					continue
				}
				groupKeys = append(groupKeys, gm.V("groupKey"))
			}
		case jsx.Truthy(c.QV("userId")):
			// 한 회원의 글(관리자 회원 상세) — 게시판으로 등록된 그룹만(라이프·앱 데이터 같은 비공개 기록은 게시판이 아니다)
			groupKeys = boardGroupKeys(adminGroups, userGroups)
		default:
			groupKeys = []any{}
			for _, g := range adminGroups {
				gm, _ := g.(*jsx.OrderedMap)
				if jsx.Truthy(gm.V("isBoardStore")) && !jsx.Truthy(gm.V("isGroupKeyRequiredForList")) && !jsx.Truthy(gm.Path("feature", "useES")) {
					groupKeys = append(groupKeys, gm.V("groupKey"))
				}
			}
		}
		postGroup := findGroup(adminGroups, groupKey)
		if groupKey != "" && postGroup != nil && !jsx.Truthy(postGroup.Path("feature", "useES")) && !jsx.Truthy(postGroup.V("isBoardStore")) {
			c.Ecb("잘못된 요청")
			return
		} else if postGroup == nil {
			if !pg.Has("userPostGroups") {
				c.Ecb("TypeError: Cannot read properties of undefined (reading 'find')")
				return
			}
			postGroup = findGroup(userGroups, groupKey)
		}
		pageNo := jsParseInt(orStr(c.QV("pageNo"), "1"))
		defSize := "20"
		if len(groupKeys) > 0 {
			defSize = "1000"
		}
		pageSize := jsParseInt(orStr(c.QV("pageSize"), defSize))
		orderBy := orStr(c.QV("orderBy"), "postId")
		orderByDesc := orStr(c.QV("orderByDesc"), "desc")
		searchGroupKeys := groupKeys

		assignPostNo := func(list []*jsx.OrderedMap, total any) []any {
			offset := (pageNo - 1) * pageSize
			isDesc := strings.ToLower(orderByDesc) != "asc"
			out := []any{}
			for i, o := range list {
				x := jsx.Assign(nil, o)
				if isDesc {
					t := jsx.Number(orVal(total, 0.0))
					x.Set("postNo", math.Max(t-offset-float64(i), 1))
				} else {
					x.Set("postNo", offset+float64(i)+1)
				}
				out = append(out, x)
			}
			return out
		}
		ownerOnly := jsx.Truthy(postGroup.Path("feature", "useOwnerOnly"))
		maskOthers := func(o *jsx.OrderedMap) (*jsx.OrderedMap, error) {
			if !ownerOnly || (userID != 0 && looseEq(o.V("userId"), userID)) {
				return o, nil
			}
			if err := maskField(o, "json"); err != nil {
				return nil, err
			}
			if err := maskField(o, "jsonUser"); err != nil {
				return nil, err
			}
			return o, nil
		}
		maskAll := func(list []*jsx.OrderedMap) ([]*jsx.OrderedMap, error) {
			out := []*jsx.OrderedMap{}
			for _, o := range list {
				m, err := maskOthers(o)
				if err != nil {
					return nil, err
				}
				out = append(out, m)
			}
			return out, nil
		}
		fillNick := func(posts []*jsx.OrderedMap) error {
			var uids []any
			for _, o := range posts {
				if jsx.Truthy(o.V("userId")) {
					uids = append(uids, o.V("userId"))
				}
			}
			nm, err := getUserIDToNicknameMap(ctx, conn, uids)
			if err != nil {
				return err
			}
			for _, o := range posts {
				n := nm[jsx.Str(o.V("userId"))]
				o.Set("nickname", orVal(n.V("nickname"), ""))
				o.Set("userImage", orVal(n.V("image"), ""))
			}
			return nil
		}

		switch c.Method {
		case "GET":
			useEsGroupKeys := false
			qUserID := c.QV("userId")
			_, postIDsIsString := c.QV("postIds").(string)
			switch {
			case postGroup == nil && strings.Contains(groupKey, ","):
				useEsGroupKeys = true
				groupKeys = []any{}
				for _, x := range strings.Split(groupKey, ",") {
					groupKeys = append(groupKeys, trimJS(x))
				}
			case keyword != "":
				groupKeys, useEsGroupKeys = []any{}, true
			case jsx.Truthy(postIDs) || postIDsIsString:
				groupKeys, useEsGroupKeys = []any{}, true
			case jsx.Truthy(qUserID):
				groupKeys, useEsGroupKeys = []any{}, true
			case typ == "my":
				qUserID = jsx.NumStr(userID)
				groupKeys = []any{}
			}
			// 회원 번호로 찾을 땐(?userId=·내 글) ES 를 쓰지 않고 DB 에서 — ES 색인에 회원 게시판 글이 빠져 0건이 나왔다(2026-10-04)
			if jsx.Truthy(meta.Path("SITE", "feature", "useESPost")) && (jsx.Truthy(postGroup.Path("feature", "useES")) || useEsGroupKeys) && !jsx.Truthy(qUserID) {
				if jsx.Truthy(postGroup.V("useSort")) {
					orderBy, orderByDesc = "sortNo", "asc"
				}
				var status any = jsParseInt(orStr(c.QV("status"), ""))
				if math.IsNaN(status.(float64)) {
					status = ""
				}
				s := esSearch{
					hostingID: hostingID, status: status, userID: orVal(qUserID, ""),
					orderBy: orderBy, orderByDesc: orderByDesc, pageSize: pageSize, pageNo: pageNo,
					groupKeys: groupKeys, hasGroupKeys: true, groupKey: orVal(postGroup.V("groupKey"), ""),
					from: orStr(c.QV("from"), ""), to: orStr(c.QV("to"), ""),
					keyword: keyword, dataColumn: orStr(postGroup.V("dataColumn"), ""),
					nickname: orStr(c.QV("nickname"), ""), groupSubKey: orStr(c.QV("groupSubKey"), ""),
					categoryFilter: categoryFilter,
				}
				if ps, ok := c.QV("postIds").(string); ok {
					s.hasPostIDs = true
					for _, x := range strings.Split(ps, ",") {
						s.postIDs = append(s.postIDs, x)
					}
				} else if arr, ok := postIDs.([]any); ok {
					s.hasPostIDs, s.postIDs = true, arr
				}
				res, err := getEsPosts(ctx, c, s, c.Token.MemberType)
				if err != nil {
					ecbErr(c, err)
					return
				}
				posts := res.posts
				if len(groupKeys) == 0 && groupKey != "" {
					groupKeys = []any{groupKey}
				}
				if keyword != "" {
					noneEs := jsx.NewOrderedMap()
					for _, gk := range groupKeys {
						g := c.App.Settings.PostGroup(hostingID, jsx.Str(gk))
						if g == nil {
							c.Ecb("TypeError: Cannot read properties of undefined (reading 'groupKey')")
							return
						}
						if !jsx.Truthy(g.Path("feature", "useES")) {
							noneEs.Set(jsx.Str(g.V("groupKey")), g)
						}
					}
					if noneEs.Len() > 0 {
						keys := []any{}
						for _, k := range noneEs.Keys() {
							keys = append(keys, k)
						}
						zero := 0.0
						rows, _, err := getAllGroupPost(ctx, conn, hostingID, postSearch{groupKeys: keys, status: &zero})
						if err != nil {
							ecbErr(c, err)
							return
						}
						for _, o := range rows {
							col := jsx.StrOrEmpty(noneEs.Map(jsx.Str(o.V("groupKey"))).V("dataColumn"))
							if strings.Contains(jsx.Str(o.V("title")), keyword) || strings.Contains(jsx.StrOrEmpty(o.V(col)), keyword) {
								j, err := parseJSONField(o.V("json"))
								if err != nil {
									ecbErr(c, err)
									return
								}
								ju, err := parseJSONField(o.V("jsonUser"))
								if err != nil {
									ecbErr(c, err)
									return
								}
								o.Set("json", j)
								o.Set("jsonUser", ju)
								posts = append(posts, o)
							}
						}
					}
				}
				if err := fillNick(posts); err != nil {
					ecbErr(c, err)
					return
				}
				var summary any = jsx.NewOrderedMap()
				if len(posts) > 0 {
					ids := append(postIDsOf(posts), postIDsOf(res.fixedPosts)...)
					sm, err := getPostSummaryMap(ctx, conn, hostingID, ids, false)
					if err != nil {
						ecbErr(c, err)
						return
					}
					summary = sm
				}
				masked, err := maskAll(posts)
				if err != nil {
					ecbErr(c, err)
					return
				}
				fixed, err := maskAll(res.fixedPosts)
				if err != nil {
					ecbErr(c, err)
					return
				}
				c.Scb(assignPostNo(masked, res.totalCount), jsx.Obj("totalCount", res.totalCount, "summary", summary, "useES", true, "fixedPosts", rowsAny(fixed)))
				return
			}

			s := postSearch{groupKeys: searchGroupKeys, pageNo: pageNo, pageSize: pageSize, orderBy: orderBy, orderByDesc: orderByDesc}
			if jsx.Truthy(qUserID) {
				// 그 회원의 글만 — 예전엔 이 조건이 빠져 공개 관리자 게시판(공지·약관) 전체가 나왔다(2026-10-04)
				s.userID = jsx.NumStr(jsParseIntAny(qUserID))
				s.groupSubKey = orStr(c.QV("groupSubKey"), "")
			}
			if !(typ == "my" && userID != 0) {
				zero := 0.0
				s.status = &zero
			}
			rows, total, err := getAllGroupPost(ctx, conn, hostingID, s)
			if err != nil {
				ecbErr(c, err)
				return
			}
			ps := []*jsx.OrderedMap{}
			for _, o := range rows {
				if jsx.Number(o.V("status")) == 99 && !(userID != 0 && looseEq(o.V("userId"), userID)) {
					continue
				}
				if g := findGroup(adminGroups, o.V("groupKey")); g != nil {
					blocked := false
					for _, b := range g.Arr("userBlockGroupSubKeys") {
						blocked = blocked || strictEq(b, o.V("groupSubKey"))
					}
					if blocked {
						continue
					}
				}
				j, err := parseJSONField(o.V("json"))
				if err != nil {
					ecbErr(c, err)
					return
				}
				if jm, ok := j.(*jsx.OrderedMap); ok {
					jm.Delete("memo")
				}
				o.Set("json", j)
				m, err := maskOthers(o)
				if err != nil {
					ecbErr(c, err)
					return
				}
				ps = append(ps, m)
			}
			if userID > 0 && typ == "my" {
				mine := []*jsx.OrderedMap{}
				for _, o := range ps {
					if looseEq(o.V("userId"), userID) {
						mine = append(mine, o)
					}
				}
				ps = mine
			}
			summary, err := getPostSummaryMap(ctx, conn, hostingID, postIDsOf(ps), false)
			if err != nil {
				ecbErr(c, err)
				return
			}
			c.Scb(assignPostNo(ps, total), jsx.Obj("totalCount", total, "summary", summary))
		case "POST":
			bd, ok := c.Body().(*jsx.OrderedMap)
			if !ok || bd == nil || (!jsx.Truthy(bd.V("postId")) && !jsx.Truthy(bd.V("postIds"))) {
				c.Ecb("잘못된 요청")
				return
			}
			if jsx.Truthy(bd.V("postId")) {
				o, err := countUpPost(ctx, conn, jsx.Param(bd.V("postId")), hostingID)
				if err != nil {
					ecbErr(c, err)
					return
				}
				c.Scb(o)
				return
			}
			var ids []any
			if arr, ok := bd.V("postIds").([]any); ok {
				ids = arr
			} else {
				ids = []any{bd.V("postIds")}
			}
			items, err := getPostByPostIDs(ctx, conn, hostingID, ids)
			if err != nil {
				ecbErr(c, err)
				return
			}
			if err := fillNick(items); err != nil {
				ecbErr(c, err)
				return
			}
			summary, err := getPostSummaryMap(ctx, conn, hostingID, postIDsOf(items), false)
			if err != nil {
				ecbErr(c, err)
				return
			}
			c.Scb(rowsAny(items), jsx.Obj("totalCount", float64(len(items)), "summary", summary))
		default:
			c.Ecb("잘못된 요청")
		}
	})
}

// GET /board/[hostingId]/[postId] — 글 상세(비공개·작성자만 보기 검사)
func boardPost(c *web.Ctx) {
	hostingID := c.Q("hostingId")
	t := c.Token
	if jsx.StrOrEmpty(c.QV("type")) == "my" && tokenKey(c) == "" {
		c.Ecb("로그인 필요")
		return
	}
	withConn(c, func(ctx context.Context, conn *sql.Conn) {
		if c.Method != "GET" {
			c.Ecb("잘못된 요청")
			return
		}
		postID := c.QV("postId")
		if !jsx.Truthy(postID) {
			c.Ecb("잘못된 요청")
			return
		}
		o, err := getPostAdmin(ctx, conn, postID, hostingID)
		if err != nil {
			ecbErr(c, err)
			return
		}
		if o == nil {
			c.Ecb("등록된 글이 없습니다.")
			return
		}
		if jsx.Number(o.V("status")) == 99 && t.MemberType != "A" {
			ecbCode(c, "비공개 글입니다.", "99")
			return
		}
		meta := c.App.Settings.Config(hostingID)
		postGroup := findGroup(meta.Map("POST_GROUPS").Arr("adminPostGroups"), o.V("groupKey"))
		if postGroup == nil {
			postGroup = findGroup(meta.Map("POST_GROUPS").Arr("userPostGroups"), o.V("groupKey"))
		}
		if postGroup == nil {
			if t.MemberType == "A" && looseEq(o.V("hostingId"), t.HostingID) {
				c.Scb(o)
				return
			}
			c.Ecb("등록되지 않은 게시판입니다.")
			return
		}
		if jsx.Truthy(postGroup.Path("feature", "useOwnerOnly")) && t.MemberType != "A" {
			login := jsParseIntAny(tokenKey(c))
			if math.IsNaN(login) || login == 0 || !looseEq(o.V("userId"), login) {
				ecbCode(c, "작성자만 볼 수 있는 글입니다.", "99")
				return
			}
		}
		if jsx.Str(postGroup.V("dataColumn")) == "jsonUser" {
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
		ju, err := jsx.ParseString(jsx.Str(o.V("jsonUser")))
		if err != nil {
			ecbErr(c, err)
			return
		}
		o.Set("jsonUser", ju)
		var nm map[string]*jsx.OrderedMap
		if jsx.Truthy(o.V("userId")) {
			nm, err = getUserIDToNicknameMap(ctx, conn, []any{o.V("userId")})
			if err != nil {
				ecbErr(c, err)
				return
			}
		}
		n := nm[jsx.Str(o.V("userId"))]
		nickname := jsx.StrOrEmpty(n.V("nickname"))
		image := n.V("image")
		if nickname != "" {
			if jm, ok := ju.(*jsx.OrderedMap); ok && jsx.Truthy(jm.V("name")) {
				jm.Set("name", nickname)
			}
		}
		if jsx.Truthy(image) {
			o.Set("userImage", image)
		}
		summary, err := getPostSummaryMap(ctx, conn, hostingID, []any{o.V("postId")}, t.MemberType == "A")
		if err != nil {
			ecbErr(c, err)
			return
		}
		c.Scb(o, jsx.Obj("summary", summary))
	})
}

// boardGroupKeys — 게시판으로 등록된 그룹(관리자·사용자, 고객 문의 CS 제외)
func boardGroupKeys(adminGroups, userGroups []any) []any {
	out := []any{}
	for _, list := range [][]any{adminGroups, userGroups} {
		for _, g := range list {
			gm, _ := g.(*jsx.OrderedMap)
			if gm == nil || jsx.Str(gm.V("groupKey")) == "" || jsx.Str(gm.V("groupKey")) == "CS" {
				continue
			}
			out = append(out, gm.V("groupKey"))
		}
	}
	return out
}
