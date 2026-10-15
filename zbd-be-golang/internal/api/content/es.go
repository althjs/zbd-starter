package content

// 글 검색(ES 인덱스 posts) — 게시판 목록·통합검색·답글 목록, 글 저장 때 색인.
// 사본: pages/api/user/post/post-es-utils.ts · elastic/es-utils.ts 의 일부 — 전환 뒤 합칠 것
//
// 설정: ES_HOST(기본 http://localhost:9202) · ES_PASSWORD(사용자 elastic)

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

const esIndex = "posts"

var esHTTP = &http.Client{Timeout: 30 * time.Second}

// esCall — ES 요청 하나. 2xx 가 아니면 오류(지금 서비스 클라이언트의 ResponseError 모양).
func esCall(ctx context.Context, c *web.Ctx, method, path string, body any) (*jsx.OrderedMap, error) {
	host := strings.TrimRight(c.App.Cfg.GetOr("ES_HOST", "http://localhost:9202"), "/")
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(jsx.Marshal(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, host+path, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("elastic", c.App.Cfg.Get("ES_PASSWORD"))
	res, err := esHTTP.Do(req)
	if err != nil {
		return nil, esConnErr(host, err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	out, _ := jsx.ParseString(string(b))
	m, _ := out.(*jsx.OrderedMap)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		typ := jsx.StrOrEmpty(m.Path("error", "type"))
		if typ == "" {
			typ = res.Status
		}
		return m, errString("ResponseError: " + typ)
	}
	if m == nil {
		m = jsx.NewOrderedMap()
	}
	return m, nil
}

// esConnErr — 접속 실패를 지금 서비스와 같은 문구로("ConnectionError: connect ECONNREFUSED 127.0.0.1:9")
func esConnErr(host string, err error) error {
	if errors.Is(err, syscall.ECONNREFUSED) {
		addr := host
		if u, e := url.Parse(host); e == nil {
			h, p := u.Hostname(), u.Port()
			if h == "localhost" {
				h = "127.0.0.1"
			}
			addr = net.JoinHostPort(h, p)
		}
		return errString("ConnectionError: connect ECONNREFUSED " + addr)
	}
	return errString("ConnectionError: " + err.Error())
}

// putEsPost — 글 색인(고치기, 없으면 만들기). 실패는 조용히 넘긴다(지금 서비스와 같다).
func putEsPost(ctx context.Context, c *web.Ctx, post *jsx.OrderedMap) {
	p := jsx.Assign(nil, post)
	if s, ok := p.V("json").(string); ok {
		if v, err := parseJSONField(s); err == nil {
			p.Set("json", v)
		}
	}
	if s, ok := p.V("jsonUser").(string); ok {
		if v, err := parseJSONField(s); err == nil {
			p.Set("jsonUser", v)
		}
	}
	id := url.PathEscape(jsx.StrOrEmpty(p.V("postId")))
	if _, err := esCall(ctx, c, "POST", "/"+esIndex+"/_update/"+id, jsx.Obj("doc", p)); err != nil {
		p.Set("createdAt", float64(time.Now().UnixMilli()))
		_, _ = esCall(ctx, c, "PUT", "/"+esIndex+"/_create/"+id, p)
	}
}

// deleteEsPost — 색인에서 지우기(없어도 괜찮다)
func deleteEsPost(ctx context.Context, c *web.Ctx, postID string) {
	_, _ = esCall(ctx, c, "DELETE", "/"+esIndex+"/_doc/"+url.PathEscape(postID), nil)
}

// esSearch — getEsPosts 의 조건
type esSearch struct {
	hostingID      any
	status         any // 숫자 또는 ""
	userID         any
	parentPostIDs  []any
	orderBy        string
	orderByDesc    string
	pageSize       float64
	pageNo         float64
	groupKeys      []any
	hasGroupKeys   bool
	groupKey       any
	from, to       string
	postIDs        []any
	hasPostIDs     bool
	keyword        string
	dataColumn     string
	nickname       string
	groupSubKey    string
	categoryFilter *jsx.OrderedMap
}

// esResult — {posts, totalCount, fixedPosts?}
type esResult struct {
	posts      []*jsx.OrderedMap
	totalCount any
	fixedPosts []*jsx.OrderedMap
	hasFixed   bool
}

func term(k string, v any) *jsx.OrderedMap { return jsx.Obj("term", jsx.Obj(k, v)) }

// kstMs — "YYYY-MM-DD HH:MM:SS"(서버 시각 = 한국) → ms. 안 되면 NaN(JSON null)
func kstMs(s string) float64 {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, kst())
	if err != nil {
		return jsx.Number("x")
	}
	return float64(t.UnixMilli())
}

func kst() *time.Location {
	l, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return time.FixedZone("KST", 9*3600)
	}
	return l
}

// getEsPosts — ES 에서 글 목록(고정글은 따로)
func getEsPosts(ctx context.Context, c *web.Ctx, s esSearch, memberType string) (*esResult, error) {
	base := []any{term("hostingId", s.hostingID)}
	_, statusNum := s.status.(float64)
	if !(memberType == "A" || jsx.Truthy(s.userID)) {
		base = append(base, term("status", 0.0))
	}
	must := append([]any{}, base...)
	must = append(must, jsx.Obj("bool", jsx.Obj("should", []any{
		jsx.Obj("bool", jsx.Obj("must_not", jsx.Obj("exists", jsx.Obj("field", "json.isFixTop")))),
		term("json.isFixTop", false),
	})))
	var mustNot []any
	if len(s.parentPostIDs) == 0 {
		mustNot = []any{jsx.Obj("exists", jsx.Obj("field", "json.parentPostIds"))}
	} else {
		must = append(must, jsx.Obj("terms", jsx.Obj("json.parentPostIds", s.parentPostIDs)))
	}
	sort := []any{}
	if s.orderBy == "sortNo" {
		sort = append(sort, jsx.Obj("json.sortNo", jsx.Obj("order", s.orderByDesc, "missing", "_first")))
	}
	sort = append(sort, jsx.Obj("createdAt", jsx.Obj("order", s.orderByDesc)))
	highlight := jsx.Obj("pre_tags", []any{"<mark>"}, "post_tags", []any{"</mark>"}, "fields", jsx.Obj(
		"title", jsx.NewOrderedMap(), "json.body", jsx.NewOrderedMap(), "jsonUser.body", jsx.NewOrderedMap(), "jsonUser.name", jsx.NewOrderedMap()))

	if memberType == "A" && statusNum {
		must = append(must, jsx.Obj("match", jsx.Obj("status", s.status)))
	}
	if len(s.groupKeys) > 0 {
		up := []any{}
		for _, g := range s.groupKeys {
			up = append(up, strings.ToUpper(jsx.Str(g)))
		}
		must = append(must, jsx.Obj("terms", jsx.Obj("groupKey", up)))
	} else if jsx.Truthy(s.groupKey) {
		must = append(must, term("groupKey", s.groupKey))
	}
	if jsx.Truthy(s.userID) {
		must = append(must, term("userId", s.userID))
	}
	if s.from != "" && s.to != "" {
		must = append(must, jsx.Obj("range", jsx.Obj("createdAt", jsx.Obj("gte", kstMs(s.from+" 00:00:00"), "lte", kstMs(s.to+" 23:59:59")))))
	}
	if s.hasPostIDs && len(s.postIDs) == 1 && !jsx.Truthy(s.postIDs[0]) {
		return &esResult{posts: []*jsx.OrderedMap{}, totalCount: 0.0}, nil
	}
	if s.hasPostIDs && len(s.postIDs) > 0 {
		must = append(must, jsx.Obj("ids", jsx.Obj("values", s.postIDs)))
	} else {
		if s.keyword != "" {
			fields := []any{"title", "json.body", "jsonUser.body", "jsonUser.name"}
			if s.dataColumn == "json" {
				fields = []any{"title", "json.body"}
			} else if s.dataColumn == "jsonUser" {
				fields = []any{"title", "jsonUser.body", "jsonUser.name"}
			}
			must = append(must, jsx.Obj("multi_match", jsx.Obj("query", s.keyword, "minimum_should_match", "90%", "fields", fields)))
		}
		if s.nickname != "" {
			must = append(must, jsx.Obj("multi_match", jsx.Obj("query", s.nickname, "minimum_should_match", "90%", "fields", []any{"jsonUser.name"})))
		}
	}
	if s.groupSubKey != "" {
		must = append(must, term("groupSubKey", s.groupSubKey))
	}
	if s.categoryFilter == nil {
		s.categoryFilter = jsx.NewOrderedMap()
	}
	for _, path := range s.categoryFilter.Keys() {
		v := s.categoryFilter.V(path)
		if path == "" || !jsx.Truthy(v) {
			continue
		}
		dc := s.dataColumn
		if dc == "" {
			dc = "json"
		}
		cols := []string{}
		for _, x := range []string{dc, "jsonUser", "json"} { // new Set — 순서대로 한 번씩
			dup := false
			for _, y := range cols {
				dup = dup || y == x
			}
			if !dup {
				cols = append(cols, x)
			}
		}
		should := []any{}
		for _, col := range cols {
			should = append(should, term(col+"."+path+".keyword", v))
		}
		must = append(must, jsx.Obj("bool", jsx.Obj("should", should, "minimum_should_match", 1.0)))
	}
	boolQ := jsx.Obj("must", must)
	if len(mustNot) > 0 {
		boolQ.Set("must_not", mustNot)
	}
	body := jsx.Obj(
		"highlight", highlight,
		"query", jsx.Obj("bool", boolQ),
		"size", s.pageSize,
		"from", s.pageSize*(s.pageNo-1),
		"sort", sort,
		"track_total_hits", true,
	)
	res, err := esCall(ctx, c, "POST", "/"+esIndex+"/_search", body)
	if err != nil {
		return nil, err
	}
	out := &esResult{totalCount: res.Path("hits", "total", "value")}
	for _, h := range res.Map("hits").Arr("hits") {
		hm, _ := h.(*jsx.OrderedMap)
		post, _ := hm.V("_source").(*jsx.OrderedMap)
		if post == nil {
			post = jsx.NewOrderedMap()
		}
		post.Set("postId", orVal(hm.V("_id"), ""))
		if hl := hm.Map("highlight"); hl != nil {
			if t := hl.Arr("title"); t != nil && post.Has("title") {
				post.Set("title", t[0])
			}
			if n := hl.Arr("jsonUser.name"); n != nil {
				if ju, ok := post.V("jsonUser").(*jsx.OrderedMap); ok {
					ju.Set("name", n[0])
				}
			}
		}
		out.posts = append(out.posts, post)
	}
	if out.posts == nil {
		out.posts = []*jsx.OrderedMap{}
	}
	if s.keyword == "" && s.hasGroupKeys && len(s.groupKeys) < 2 {
		fm := append([]any{}, base...)
		fm = append(fm, term("json.isFixTop", true))
		fm = append(fm, term("groupKey", s.groupKey))
		fb := jsx.Assign(nil, body)
		fb.Set("query", jsx.Obj("bool", jsx.Obj("must", fm)))
		fb.Set("size", 100.0)
		fb.Set("from", 0.0)
		fr, err := esCall(ctx, c, "POST", "/"+esIndex+"/_search", fb)
		if err != nil {
			return nil, err
		}
		out.hasFixed = true
		out.fixedPosts = []*jsx.OrderedMap{}
		for _, h := range fr.Map("hits").Arr("hits") {
			hm, _ := h.(*jsx.OrderedMap)
			post, _ := hm.V("_source").(*jsx.OrderedMap)
			if post == nil {
				post = jsx.NewOrderedMap()
			}
			post.Set("postId", orVal(hm.V("_id"), ""))
			out.fixedPosts = append(out.fixedPosts, post)
		}
	}
	return out, nil
}
