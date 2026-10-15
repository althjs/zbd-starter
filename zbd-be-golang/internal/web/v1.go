package web

import (
	"net/http"
	"strconv"
	"strings"

	"zbd.kr/be/internal/jsx"
)

// 새 응답 형식(`/api/v1/…`, 대표 결정 2026-10-03)
//
//	성공  HTTP 200 {data, meta:{requestId, …}}
//	실패  HTTP 4xx/5xx {error:{code, message}, meta:{requestId}}
//
// 화면은 HTTP 상태로 성공·실패를 가르고, 실패는 error.code(뜻이 읽히는 문자열)로 나눈다 — 문구로 나누지 않는다.
// 옛 핸들러도 /api/v1 로 들어오면 c.Scb·c.Ecb 가 이 모양으로 낸다.

// 오류 코드 — 한 곳에 모은다. 화면이 이 값으로 분기한다.
const (
	ErrBadRequest   = "BAD_REQUEST"  // 요청 값이 잘못됨
	ErrInvalidSort  = "INVALID_SORT" // 허락하지 않은 정렬 칸
	ErrUnauthorized = "UNAUTHORIZED" // 로그인 필요·토큰 만료
	ErrForbidden    = "FORBIDDEN"    // 권한 없음
	ErrNotFound     = "NOT_FOUND"    // 없음
	ErrConflict     = "CONFLICT"     // 그사이 바뀜·이미 있음
	ErrFailed       = "FAILED"       // 그 밖의 처리 실패(옛 ecb)
)

// Fail — 새 형식 실패 응답
func (c *Ctx) Fail(status int, code, message string) {
	c.failed = true
	if status < 400 {
		status = 400
	}
	if message == "" {
		message = "실패"
	}
	c.status = status
	c.Send(jsx.Obj("error", jsx.Obj("code", code, "message", message)))
}

func (c *Ctx) sendV1(data any, meta *jsx.OrderedMap) {
	m := jsx.NewOrderedMap()
	m.Set("data", data)
	if meta != nil {
		m.Set("meta", meta)
	}
	c.Send(m)
}

// failStatus — 옛 ecb 는 HTTP 200 이었다. 새 형식에선 이미 정한 상태가 없으면 400.
func (c *Ctx) failStatus() int {
	if c.status >= 400 {
		return c.status
	}
	return 400
}

func (c *Ctx) unauthorized() {
	if c.V1 {
		c.Fail(http.StatusUnauthorized, ErrUnauthorized, "로그인이 필요합니다")
		return
	}
	c.Status(401).Send("401 Unauthorized")
}

// ListQuery — 목록 요청 값(규약: ~/zbd-port/LIST_AUDIT.md).
//
//	?page=1&size=20&q=검색어&sort=createdAt,-title
//
// 정렬은 허락한 칸만(이름 → SQL 식). 목록 전체를 내려 화면이 거르게 하지 않는다.
type ListQuery struct {
	Page, Size int
	Q          string
	orderSQL   string
}

const (
	ListDefaultSize = 20
	ListMaxSize     = 100
)

// ParseList — page·size·q·sort 를 읽는다. allowed = 정렬 이름 → SQL 식, def = 기본 정렬(예 "-createdAt").
// 허락하지 않은 정렬이면 400 을 보내고 ok=false.
func (c *Ctx) ParseList(allowed map[string]string, def string) (ListQuery, bool) {
	lq := ListQuery{Page: 1, Size: ListDefaultSize, Q: strings.TrimSpace(c.Q("q"))}
	if v := c.Q("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			c.Fail(400, ErrBadRequest, "page 는 1 이상의 정수")
			return lq, false
		}
		lq.Page = n
	}
	if v := c.Q("size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			c.Fail(400, ErrBadRequest, "size 는 1 이상의 정수")
			return lq, false
		}
		if n > ListMaxSize {
			n = ListMaxSize
		}
		lq.Size = n
	}
	sort := c.Q("sort")
	if sort == "" {
		sort = def
	}
	var parts []string
	for _, f := range strings.Split(sort, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		dir := "ASC"
		if strings.HasPrefix(f, "-") {
			dir, f = "DESC", f[1:]
		}
		expr, ok := allowed[f]
		if !ok {
			c.Fail(400, ErrInvalidSort, "정렬할 수 없는 칸: "+f)
			return lq, false
		}
		parts = append(parts, expr+" "+dir)
	}
	if len(parts) > 0 {
		lq.orderSQL = " ORDER BY " + strings.Join(parts, ", ")
	}
	return lq, true
}

// Offset — LIMIT 의 시작
func (lq ListQuery) Offset() int { return (lq.Page - 1) * lq.Size }

// OrderLimit — " ORDER BY … LIMIT ?, ?" 와 그 인자. 정렬 식은 허락 목록에서 온 것만 들어간다.
func (lq ListQuery) OrderLimit() (string, []any) {
	return lq.orderSQL + " LIMIT ?, ?", []any{lq.Offset(), lq.Size}
}

// LikePrefix — 검색어를 LIKE '검색어%' 인자로(%·_ 는 글자로).
func LikePrefix(q string) string { return escLike(q) + "%" }

// LikeContains — LIKE '%검색어%' 인자로.
func LikeContains(q string) string { return "%" + escLike(q) + "%" }

func escLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// SendList — 목록 응답 {data:[…], meta:{requestId, total, page, size}}
func (c *Ctx) SendList(items []any, total int64, lq ListQuery) {
	if items == nil {
		items = []any{}
	}
	c.sendV1(items, jsx.Obj("total", float64(total), "page", float64(lq.Page), "size", float64(lq.Size)))
}

// SendListSummary — SendList + meta.summary(목록과 같이 필요한 합계·개수 — 화면이 전체를 받아 세지 않게)
func (c *Ctx) SendListSummary(items []any, total int64, lq ListQuery, summary any) {
	if items == nil {
		items = []any{}
	}
	c.sendV1(items, jsx.Obj("total", float64(total), "page", float64(lq.Page), "size", float64(lq.Size), "summary", summary))
}

// SendCursor — 커서형 목록 {data:[…], meta:{requestId, size, next}} — 끝없이 쌓이는 것(대화·알림·로그).
// next = 다음 쪽을 받을 커서(before=next), 더 없으면 null.
func (c *Ctx) SendCursor(items []any, size int, next any) {
	if items == nil {
		items = []any{}
	}
	c.sendV1(items, jsx.Obj("size", float64(size), "next", next))
}

// WithTie — 정렬 끝에 유일한 칸(기본키)을 더한다. 같은 값이 많은 칸으로 정렬하면 쪽마다 순서가 흔들려
// 1쪽·2쪽에 같은 행이 겹치거나 빠지기 때문이다. 이미 그 칸으로 정렬하면 그대로.
func (lq ListQuery) WithTie(expr string) ListQuery {
	if strings.Contains(lq.orderSQL, " "+expr+" ") {
		return lq
	}
	if lq.orderSQL == "" {
		lq.orderSQL = " ORDER BY " + expr + " DESC"
	} else {
		lq.orderSQL += ", " + expr + " DESC"
	}
	return lq
}
