// Package v1x — `/api/v1` 목록 API 들이 같이 쓰는 입구·쪽 나누기·검색 도우미.
//
// 새 목록 API 는 /api/v1 에서만 쓴다(같은 경로라도 /service·/api 로 오면 404).
// 입구는 web.Cors 로 등록하고 권한은 여기서 본다:
//
//	Admin(원래경로, h) — 관리자 토큰 + 원래 API(/service/…)에 묶인 메뉴 권한(GET = R 이상). 새 경로는 메뉴 연결이 없어서 원래 경로로 확인한다.
//	User(h)            — 회원 토큰 + 발급받은 브라우저(헤더 zbduid)
//	Public(h)          — 로그인 없이(공개 화면)
package v1x

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/pii"
	"zbd.kr/be/internal/web"
)

// pre — /api/v1 이 아니면 404, 미리 보내는 OPTIONS 는 200. 계속할지 돌려준다.
func pre(c *web.Ctx) bool {
	if !c.V1 {
		c.Fail(http.StatusNotFound, web.ErrNotFound, "없는 주소입니다")
		return false
	}
	if c.Method == http.MethodOptions {
		c.Status(200).End()
		return false
	}
	return true
}

// Admin — 관리자만. orig = 같은 데이터를 주던 원래 API 경로(지금 서비스 모양, 예 "/admin/user") — 그 메뉴 권한으로 확인한다.
func Admin(orig string, h web.Handler) web.Handler {
	return func(c *web.Ctx) {
		if !pre(c) {
			return
		}
		t := c.Token
		if !t.IsAdmin() {
			c.Fail(http.StatusUnauthorized, web.ErrUnauthorized, "로그인이 필요합니다")
			return
		}
		if t.HostingID == "" {
			c.Fail(http.StatusForbidden, web.ErrForbidden, "호스팅 관리자만 볼 수 있습니다")
			return
		}
		method := c.Method
		if method == http.MethodHead {
			method = http.MethodGet
		}
		if !c.App.Admin.Allowed(t.HostingID, t.Key, method, "/service"+orig) {
			c.Fail(http.StatusForbidden, web.ErrForbidden, "권한이 없습니다")
			return
		}
		h(c)
	}
}

// User — 로그인한 회원만(토큰을 발급받은 브라우저에서만).
func User(h web.Handler) web.Handler {
	return func(c *web.Ctx) {
		if !pre(c) {
			return
		}
		t := c.Token
		if !t.IsUser() || t.HostingID == "" {
			c.Fail(http.StatusUnauthorized, web.ErrUnauthorized, "로그인이 필요합니다")
			return
		}
		uid := c.BrowserUID()
		if uid == "" {
			uid = "UNKNOWN"
		}
		if uid != t.BrowserUID {
			c.Fail(http.StatusUnauthorized, web.ErrUnauthorized, "로그인이 필요합니다")
			return
		}
		h(c)
	}
}

// Public — 로그인 없이(공개). /api/v1 만.
func Public(h web.Handler) web.Handler {
	return func(c *web.Ctx) {
		if !pre(c) {
			return
		}
		h(c)
	}
}

// OnlyGET — GET 이 아니면 405
func OnlyGET(c *web.Ctx) bool {
	if c.Method != http.MethodGet && c.Method != http.MethodHead {
		c.Fail(http.StatusMethodNotAllowed, web.ErrBadRequest, "GET 만 받습니다")
		return false
	}
	return true
}

// PublicHosting — 공개 화면의 호스팅(토큰이 있으면 그 호스팅, 없으면 즈비드)
func PublicHosting(c *web.Ctx) string {
	if c.Token.Valid() && c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return "ZBD"
}

// Where — WHERE 절을 조각으로 모은다(값은 늘 ? 인자).
type Where struct {
	parts []string
	Args  []any
}

// Add — 조건 하나(sql 안의 ? 개수와 args 가 같아야 한다)
func (w *Where) Add(sql string, args ...any) {
	w.parts = append(w.parts, sql)
	w.Args = append(w.Args, args...)
}

// SQL — " WHERE a AND b" (조건이 없으면 "")
func (w *Where) SQL() string {
	if len(w.parts) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.parts, " AND ")
}

// Page — 같은 조건으로 개수와 한 쪽을 읽는다. sel = "SELECT …", from = " FROM … [JOIN …]" (WHERE 는 w 로).
func Page(ctx context.Context, q db.Querier, sel, from string, w *Where, lq web.ListQuery) ([]*jsx.OrderedMap, int64, error) {
	cnt, err := db.Query(ctx, q, "SELECT COUNT(*) AS n"+from+w.SQL(), w.Args...)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if len(cnt) > 0 {
		total = jsx.Int(cnt[0].V("n"))
	}
	if total == 0 || int64(lq.Offset()) >= total {
		return nil, total, nil
	}
	ol, la := lq.OrderLimit()
	rows, err := db.Query(ctx, q, sel+from+w.SQL()+ol, append(append([]any{}, w.Args...), la...)...)
	return rows, total, err
}

// Any — []*OrderedMap → []any
func Any(rows []*jsx.OrderedMap) []any {
	out := make([]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, r)
	}
	return out
}

// Ints — "1,2,3" → 정수 목록(숫자가 아닌 것은 버린다)
func Ints(s string) []any {
	var out []any
	for _, p := range strings.Split(s, ",") {
		if n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// Strs — "a,b" → 빈 값 뺀 목록
func Strs(s string) []any {
	var out []any
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IntParam — 정수 칸(없으면 def, 숫자가 아니면 ok=false 와 400)
func IntParam(c *web.Ctx, k string, def int64) (int64, bool) {
	v := c.Q(k)
	if v == "" {
		return def, true
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		c.Fail(400, web.ErrBadRequest, k+" 는 정수")
		return 0, false
	}
	return n, true
}

var reYMD = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// DateRange — from·to(YYYY-MM-DD, 한국 날짜)를 col(DB 시각 칸) 조건으로. 형식이 틀리면 400.
func DateRange(c *web.Ctx, w *Where, col string) bool {
	for _, k := range []string{"from", "to"} {
		v := c.Q(k)
		if v == "" {
			continue
		}
		if !reYMD.MatchString(v) {
			c.Fail(400, web.ErrBadRequest, k+" 는 YYYY-MM-DD")
			return false
		}
		if _, err := time.Parse("2006-01-02", v); err != nil {
			c.Fail(400, web.ErrBadRequest, k+" 는 YYYY-MM-DD")
			return false
		}
		if k == "from" {
			w.Add(col+" >= ?", v+" 00:00:00")
		} else {
			w.Add(col+" < DATE_ADD(?, INTERVAL 1 DAY)", v+" 00:00:00")
		}
	}
	return true
}

var reDigits = regexp.MustCompile(`^[0-9][0-9 -]*$`)

// MemberSearch — 회원 검색어 q 를 조건으로(이름·이메일·전화는 암호문이라 LIKE 불가 — 정확 일치만).
//
//	이메일 모양(@)      → email_hash 정확
//	숫자·하이픈만        → mobile_hash 정확 또는 userId
//	그 밖               → nickname 앞부분 LIKE 또는 name_hash 정확(실명 전체)
//
// p = 회원 표 별칭 앞부분("" 또는 "m.").
func MemberSearch(w *Where, k *pii.Keys, p, q string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return
	}
	hash := func(kind pii.Kind) string {
		h := k.Hash(kind, q)
		if h == "" {
			h = "-" // 키가 없거나 빈 값 — 아무것도 맞지 않게
		}
		return h
	}
	switch {
	case strings.Contains(q, "@"):
		w.Add(p+"email_hash = ?", hash(pii.Email))
	case reDigits.MatchString(q):
		digits := strings.NewReplacer("-", "", " ", "").Replace(q)
		w.Add("("+p+"mobile_hash = ? OR "+p+"userId = ?)", hash(pii.Mobile), digits)
	default:
		w.Add("("+p+"nickname LIKE ? OR "+p+"name_hash = ?)", web.LikePrefix(q), hash(pii.Name))
	}
}

// Nick — 남에게 보일 이름: 닉네임 → 없으면 "회원 {번호}"
func Nick(nick any, userID any) string {
	if s := strings.TrimSpace(jsx.StrOrEmpty(nick)); s != "" {
		return s
	}
	if jsx.Truthy(userID) {
		return "회원 " + jsx.Str(userID)
	}
	return ""
}

// Writers — userId 목록 → {userId: 닉네임}(회원 표 한 번)
func Writers(ctx context.Context, q db.Querier, hostingID string, ids []any) (map[string]string, error) {
	out := map[string]string{}
	seen := map[string]bool{}
	var uniq []any
	for _, id := range ids {
		s := jsx.Str(id)
		if !jsx.Truthy(id) || seen[s] {
			continue
		}
		seen[s] = true
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return out, nil
	}
	rows, err := db.Query(ctx, q, `SELECT userId, nickname FROM members WHERE hostingId = ? AND userId IN (?)`, hostingID, uniq)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[jsx.Str(r.V("userId"))] = Nick(r.V("nickname"), r.V("userId"))
	}
	return out, nil
}

// Cut — 글자(rune) 기준 앞 n 자
func Cut(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// IntIn — 주소 값 k(정수, 쉼표로 여럿)를 col IN (…) 조건으로. 값이 있는데 정수가 하나도 없으면 400.
func IntIn(c *web.Ctx, w *Where, k, col string) bool {
	s := c.Q(k)
	if s == "" {
		return true
	}
	ids := Ints(s)
	if len(ids) == 0 {
		c.Fail(400, web.ErrBadRequest, k+" 는 정수(쉼표로 여럿)")
		return false
	}
	w.Add(col+" IN (?)", ids)
	return true
}

// OrEmpty — 거짓 값이면 ""
func OrEmpty(v any) any {
	if !jsx.Truthy(v) {
		return ""
	}
	return v
}

// YMDParam — 날짜 칸(YYYY-MM-DD). 없으면 "", 형식이 틀리면 400 과 ok=false.
func YMDParam(c *web.Ctx, k string) (string, bool) {
	v := c.Q(k)
	if v == "" {
		return "", true
	}
	if _, err := time.Parse("2006-01-02", v); err != nil || !reYMD.MatchString(v) {
		c.Fail(400, web.ErrBadRequest, k+" 는 YYYY-MM-DD")
		return "", false
	}
	return v, true
}

// List — ParseList + 정렬 끝에 유일한 칸(tie, 보통 기본키) — 쪽 사이 순서가 흔들리지 않게
func List(c *web.Ctx, allowed map[string]string, def, tie string) (web.ListQuery, bool) {
	lq, ok := c.ParseList(allowed, def)
	if !ok {
		return lq, false
	}
	if tie != "" {
		lq = lq.WithTie(tie)
	}
	return lq, true
}
