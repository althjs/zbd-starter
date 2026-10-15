package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/auth"
)

// Ctx — 요청 하나. 지금 서비스의 req·res 를 한데 모은 모양이다.
//
//	c.Query   req.query — 주소의 ?a=1 과 경로 변수([postId]) 를 합친 것. 값은 string, 같은 키가 여럿이면 []any
//	c.Body()  req.body  — JSON 이면 객체(*jsx.OrderedMap)/배열, 폼이면 객체, 그 밖은 문자열, 없으면 nil
//	c.Token   로그인 토큰(없거나 잘못되면 Valid()=false)
//	c.Send(v) res.send(v) — 객체면 JSON(+ meta.requestId), 문자열이면 그대로
type Ctx struct {
	W      http.ResponseWriter
	R      *http.Request
	App    *app.App
	Method string
	// Route — 경로 정의(지금 서비스의 "/service/admin/design/[designId]" 모양). 관리자 권한 검사에 쓴다.
	Route string
	Query *jsx.OrderedMap
	Token *auth.Token

	guard Guard
	opts  Options
	// V1 — `/api/v1/…` 로 들어온 요청. 응답이 새 형식이다(v1.go).
	V1        bool
	reqID     string
	status    int
	failed    bool // Ecb·Fail 로 실패를 돌려줬다(Options.Changes 판단)
	wrote     bool
	body      any
	bodyRead  bool
	bodyErr   error
	ctxCancel context.CancelFunc
	// fromCookie — 세션 쿠키로 로그인한 요청(session.go)
	fromCookie bool
}

// Ctx 는 context.Context 처럼도 쓴다(DB 조회 등)
func (c *Ctx) Context() context.Context { return c.R.Context() }

func newRequestID() string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 3)
	for i := range b {
		b[i] = digits[rand.Intn(36)]
	}
	return strconv.FormatInt(time.Now().UnixMilli(), 36) + string(b)
}

func unescape(s string) (string, error) { return url.PathUnescape(s) }

// buildQuery — 주소 쿼리 + 경로 변수(경로 변수가 이긴다 — Next 와 같다)
func buildQuery(req *http.Request, params map[string]any) *jsx.OrderedMap {
	q := jsx.NewOrderedMap()
	vals := req.URL.Query()
	// url.Values 는 순서가 없다 → 원문에서 순서대로
	for _, kv := range strings.Split(req.URL.RawQuery, "&") {
		if kv == "" {
			continue
		}
		k, _, _ := strings.Cut(kv, "=")
		key, err := url.QueryUnescape(k)
		if err != nil || q.Has(key) {
			continue
		}
		vs := vals[key]
		if len(vs) == 1 {
			q.Set(key, vs[0])
		} else if len(vs) > 1 {
			arr := make([]any, len(vs))
			for i, v := range vs {
				arr[i] = v
			}
			q.Set(key, arr)
		}
	}
	renameOldKeys(q, 0)
	for k, v := range params {
		q.Set(k, v)
	}
	return q
}

func renameOldKeys(v any, depth int) {
	if depth > 1 {
		return
	}
	switch x := v.(type) {
	case *jsx.OrderedMap:
		if x == nil {
			return
		}
		if x.Has("companyId") {
			if x.Has("hostingId") {
				x.Delete("companyId")
			} else {
				x.RenameKey("companyId", "hostingId")
			}
		}
		for _, k := range x.Keys() {
			renameOldKeys(x.V(k), depth+1)
		}
	case []any: // 배열은 단계로 세지 않는다({items:[{…}]} 의 항목까지)
		for _, it := range x {
			renameOldKeys(it, depth)
		}
	}
}

const corsAllowHeaders = "content-type, Content-Type, Content-Encoding, Origin, X-Requested-With, Accept, authorization, adminid, imageid, zbduid"

func addCors(w http.ResponseWriter, req *http.Request) {
	h := w.Header()
	origin := req.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	h.Set("Access-Control-Allow-Origin", origin)
	h.Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE")
	h.Set("Access-Control-Allow-Credentials", "true")
	h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
}

func serve(a *app.App, rt *route, prefix string, params map[string]any, w http.ResponseWriter, req *http.Request) {
	c := &Ctx{
		W: w, R: req, App: a, Method: req.Method,
		Route: "/service" + rt.def, Query: buildQuery(req, params),
		guard: rt.guard, opts: rt.opts, status: 200, V1: prefix == "/api/v1",
	}
	c.Token = auth.Parse(req, a.Cfg.AESKey, a.Cfg.AESIV, "")
	c.loadSession()

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("@@ %s %s 처리 중 오류: %v\n%s", req.Method, req.URL.Path, rec, debug.Stack())
			if !c.wrote {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
				c.wrote = true
			}
		}
	}()

	if rt.guard == Cors || rt.guard == Auth {
		addCors(w, req)
		if rt.guard == Auth && req.Method == http.MethodOptions {
			c.wrote = true
			w.Header().Set("Content-Type", "application/json; charset=utf-8") // res.json() — 본문 없음
			w.WriteHeader(200)
			return
		}
		rid := strings.TrimSpace(req.Header.Get("x-request-id"))
		if rid == "" {
			rid = newRequestID()
		}
		req.Header.Set("x-request-id", rid)
		w.Header().Set("x-request-id", rid)
		c.reqID = rid
	} else if c.V1 {
		c.reqID = newRequestID()
		w.Header().Set("x-request-id", c.reqID)
	}
	if !rt.opts.RawBody {
		if err := c.readBody(); err != nil {
			if errors.Is(err, errTooLarge) {
				c.Status(413).Send("Body exceeded " + humanSize(rt.opts.BodyLimit) + " limit")
				return
			}
			if c.V1 {
				c.Fail(400, ErrBadRequest, "요청 본문을 읽을 수 없습니다")
				return
			}
			c.Status(400).JSON(jsx.Obj("code", "FAIL", "message", "Invalid request body"))
			return
		}
	}
	if rt.guard == Auth {
		// 회원 토큰은 발급받은 브라우저에서만(헤더 zbduid)
		if c.Token.MemberType == "U" {
			uid := c.BrowserUID()
			if uid == "" {
				uid = "UNKNOWN"
			}
			if uid != c.Token.BrowserUID {
				c.unauthorized()
				return
			}
		}
		if c.Token.MemberType != "A" {
			if !strings.HasPrefix(rt.def, "/user") {
				c.unauthorized()
				return
			}
		} else if !a.Admin.Allowed(c.Token.HostingID, c.Token.Key, req.Method, c.Route) {
			c.unauthorized()
			return
		}
	}
	rt.h(c)
	if changeNotifies(rt.opts, req.Method, c) {
		onChanged(a, req.Context(), c.Token.HostingID)
	}
}

// onChanged — 「설정이 바뀌었다」 알림(캐시 다시 읽기 + 소켓). 테스트가 바꿔 끼운다
var onChanged = func(a *app.App, ctx context.Context, hostingID string) { a.Changed(ctx, hostingID) }

// changeNotifies — 정책: Options.Changes 라우트에서 GET 이 아닌 요청이 성공(실패 응답 아님 · 상태 400 미만)했고 로그인한 요청일 때만 알린다.
// 테스트 changes_policy_test.go
func changeNotifies(o Options, method string, c *Ctx) bool {
	return o.Changes && method != http.MethodGet && !c.failed && c.status < 400 && c.Token.Valid()
}

func humanSize(n int64) string {
	if n%(1<<20) == 0 {
		return strconv.FormatInt(n>>20, 10) + "mb"
	}
	if n%(1<<10) == 0 {
		return strconv.FormatInt(n>>10, 10) + "kb"
	}
	return strconv.FormatInt(n, 10) + "b"
}

var errTooLarge = errors.New("본문이 너무 크다")

// RawBodyReader — 압축을 푼 본문(RawBody 라우트에서 직접 읽을 때)
func (c *Ctx) RawBodyReader() (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(c.R.Header.Get("Content-Encoding"))) {
	case "", "identity":
		return c.R.Body, nil
	case "gzip":
		return gzip.NewReader(c.R.Body)
	case "deflate":
		return flate.NewReader(c.R.Body), nil
	case "br":
		return brotli.NewReader(c.R.Body), nil
	}
	return nil, fmt.Errorf("지원하지 않는 Content-Encoding")
}

func (c *Ctx) readBody() error {
	if c.bodyRead {
		return c.bodyErr
	}
	c.bodyRead = true
	if c.R.Body == nil || c.R.Method == http.MethodGet || c.R.Method == http.MethodHead {
		return nil
	}
	rd, err := c.RawBodyReader()
	if err != nil {
		c.bodyErr = err
		return err
	}
	b, err := io.ReadAll(io.LimitReader(rd, c.opts.BodyLimit+1))
	if err != nil {
		c.bodyErr = err
		return err
	}
	if int64(len(b)) > c.opts.BodyLimit {
		c.bodyErr = errTooLarge
		return errTooLarge
	}
	if len(b) == 0 {
		return nil
	}
	ct := strings.ToLower(c.R.Header.Get("Content-Type"))
	switch {
	case strings.Contains(ct, "application/json") || strings.Contains(ct, "+json"):
		v, err := jsx.Parse(bytes.TrimSpace(b))
		if err != nil {
			c.bodyErr = err
			return err
		}
		renameOldKeys(v, 0)
		c.body = v
	case strings.Contains(ct, "application/x-www-form-urlencoded"):
		vals, err := url.ParseQuery(string(b))
		if err != nil {
			c.bodyErr = err
			return err
		}
		m := jsx.NewOrderedMap()
		for _, kv := range strings.Split(string(b), "&") {
			k, _, _ := strings.Cut(kv, "=")
			key, _ := url.QueryUnescape(strings.ReplaceAll(k, "+", " "))
			if key == "" || m.Has(key) {
				continue
			}
			vs := vals[key]
			if len(vs) == 1 {
				m.Set(key, vs[0])
			} else {
				arr := make([]any, len(vs))
				for i, v := range vs {
					arr[i] = v
				}
				m.Set(key, arr)
			}
		}
		renameOldKeys(m, 0)
		c.body = m
	default:
		c.body = string(b)
	}
	return nil
}

// Body — req.body (없으면 nil)
func (c *Ctx) Body() any {
	_ = c.readBody()
	return c.body
}

// BodyMap — 본문이 객체면 그 객체, 아니면 빈 객체(req.body || {} 자리)
func (c *Ctx) BodyMap() *jsx.OrderedMap {
	if m, ok := c.Body().(*jsx.OrderedMap); ok && m != nil {
		return m
	}
	return jsx.NewOrderedMap()
}

// Q — req.query[k] 를 문자열로(없으면 ""). 같은 키가 여럿이면 JS 의 String(배열) 처럼 쉼표로.
func (c *Ctx) Q(k string) string {
	v := c.Query.V(k)
	if v == nil {
		return ""
	}
	return jsx.Str(v)
}

// QV — req.query[k] 그대로(string / []any / nil)
func (c *Ctx) QV(k string) any { return c.Query.V(k) }

// Header — 요청 헤더
func (c *Ctx) Header(k string) string { return c.R.Header.Get(k) }

// BrowserUID — 브라우저 id. 화면은 쿠키 zbd_uid 값을 헤더 zbduid 로 보낸다(회원 토큰이 이 값을 품는다).
// 헤더가 없는 요청(소셜 로그인 콜백처럼 브라우저가 바로 여는 주소)은 쿠키에서 읽는다. 없으면 "".
func (c *Ctx) BrowserUID() string {
	if v := c.R.Header.Get("zbduid"); v != "" {
		return v
	}
	// 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
	if v := c.R.Header.Get("oxuid"); v != "" {
		return v
	}
	for _, name := range []string{"zbd_uid", "ox_uid"} { // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
		if ck, err := c.R.Cookie(name); err == nil && ck.Value != "" {
			if v, err := url.QueryUnescape(ck.Value); err == nil {
				return v
			}
			return ck.Value
		}
	}
	return ""
}

// SetHeader — 응답 헤더
func (c *Ctx) SetHeader(k, v string) *Ctx {
	c.W.Header().Set(k, v)
	return c
}

// Status — res.status(n)
func (c *Ctx) Status(n int) *Ctx {
	c.status = n
	return c
}

// Wrote — 이미 응답했는가
func (c *Ctx) Wrote() bool { return c.wrote }

// injectMeta — 입구가 Cors/Auth 면 객체 응답에 meta.requestId 를 얹는다(배열·문자열은 그대로)
func (c *Ctx) injectMeta(v any) any {
	if c.reqID == "" {
		return v
	}
	m, ok := v.(*jsx.OrderedMap)
	if !ok || m == nil {
		return v
	}
	next := jsx.NewOrderedMap()
	if mm, isMap := m.V("meta").(*jsx.OrderedMap); isMap {
		for _, k := range mm.Keys() {
			next.Set(k, mm.V(k))
		}
	}
	next.Set("requestId", c.reqID)
	m.Set("meta", next)
	return m
}

func (c *Ctx) write(ct string, b []byte) {
	if c.wrote {
		log.Printf("⚠ %s %s 응답을 두 번 보내려 했다", c.Method, c.R.URL.Path)
		return
	}
	c.wrote = true
	if ct != "" && c.W.Header().Get("Content-Type") == "" {
		c.W.Header().Set("Content-Type", ct)
	}
	c.W.Header().Set("Content-Length", strconv.Itoa(len(b)))
	c.W.WriteHeader(c.status)
	if c.Method != http.MethodHead {
		_, _ = c.W.Write(b)
	}
}

// Send — res.send(v): 문자열·[]byte 는 그대로, 그 밖(객체·배열·숫자·참거짓)은 JSON. nil 이면 빈 본문.
func (c *Ctx) Send(v any) {
	switch x := v.(type) {
	case nil:
		c.write("", nil)
	case string:
		c.write("", []byte(x)) // 지금 서비스도 문자열엔 Content-Type 을 붙이지 않는다
	case []byte:
		c.write("application/octet-stream", x)
	default:
		c.write("application/json; charset=utf-8", jsx.Marshal(c.injectMeta(x)))
	}
}

// JSON — res.json(v): 늘 JSON(문자열도 "…" 로)
func (c *Ctx) JSON(v any) {
	c.write("application/json; charset=utf-8", jsx.Marshal(c.injectMeta(v)))
}

// Redirect — res.redirect(url) 는 307, Redirect(302, url) 처럼 코드를 줄 수 있다
func (c *Ctx) Redirect(code int, to string) {
	if c.wrote {
		return
	}
	c.wrote = true
	c.W.Header().Set("Location", to)
	c.W.WriteHeader(code)
}

// End — res.end() / res.status(n).end()
func (c *Ctx) End() { c.write("", nil) }

// Scb — res.send(scb(data[, meta]))
func (c *Ctx) Scb(data any, meta ...*jsx.OrderedMap) {
	var m *jsx.OrderedMap
	if len(meta) > 0 {
		m = meta[0]
	}
	if c.V1 {
		c.sendV1(data, m)
		return
	}
	c.Send(Scb(data, m))
}

// Ecb — res.send(ecb(msg))
func (c *Ctx) Ecb(msg any) {
	c.failed = true
	if c.V1 {
		c.Fail(c.failStatus(), ErrFailed, jsx.StrOrEmpty(Ecb(msg).V("message")))
		return
	}
	c.Send(Ecb(msg))
}

// Scb — { code:"0000", status:"SUCCESS", message:"성공", data, meta?, count? }
//
//	count 는 data 가 문자열이 아니고 길이(배열 길이 / length 칸)가 0 이 아닐 때.
func Scb(data any, meta *jsx.OrderedMap) *jsx.OrderedMap {
	m := jsx.NewOrderedMap()
	m.Set("code", "0000")
	m.Set("status", "SUCCESS")
	m.Set("message", "성공")
	m.Set("data", data)
	if meta != nil {
		m.Set("meta", meta)
	}
	n := 0
	switch x := data.(type) {
	case []any:
		n = len(x)
	case []*jsx.OrderedMap:
		n = len(x)
	case []string:
		n = len(x)
	case *jsx.OrderedMap:
		if l := x.V("length"); l != nil {
			if f := jsx.Number(l); f > 0 {
				m.Set("count", l)
			}
		}
	}
	if n > 0 {
		m.Set("count", float64(n))
	}
	return m
}

// Ecb — { code:"9999", status:"FAIL", message }. 메시지가 비면 "실패".
//
//	error 를 넣으면 그 문구만(서버 내부 정보 — 스택·명령·SQL — 는 싣지 않는다).
func Ecb(msg any) *jsx.OrderedMap {
	m := jsx.NewOrderedMap()
	m.Set("code", "9999")
	m.Set("status", "FAIL")
	switch x := msg.(type) {
	case nil:
		m.Set("message", "실패")
	case error:
		s := x.Error()
		if s == "" {
			s = "실패"
		}
		m.Set("message", s)
	case string:
		if x == "" {
			m.Set("message", "실패")
		} else {
			m.Set("message", x)
		}
	default:
		if !jsx.Truthy(x) {
			m.Set("message", "실패")
		} else {
			m.Set("message", x)
		}
	}
	return m
}
