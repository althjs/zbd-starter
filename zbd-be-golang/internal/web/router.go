// Package web — 경로 처리와 요청 입구.
//
// 경로는 지금 서비스(Node BE)와 **같은 모양으로 적는다**: "/admin/design/[designId]", "/user/file/[...path]".
// 같은 핸들러가 세 주소로 받는다 — `/service/…`(지금 화면), `/api/v1/…`(새 주소), `/api/…`(서버 안에서 직접 부르던 주소).
//
// 고르는 규칙(Next.js 와 같다): 구간마다 고정 이름 > [변수] > [...나머지] > [[...나머지(없어도 됨)]].
// 고정 구간이 맞았다가 그 아래에서 막히면 되돌아가 다음 후보를 본다.
package web

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"zbd.kr/be/internal/app"
)

// Guard — 입구 종류
type Guard int

const (
	// Plain — 그대로(CORS·요청 번호·meta 주입 없음). 지금 서비스에서 withAuth·withCors 로 감싸지 않은 라우트.
	Plain Guard = iota
	// Cors — withCors: CORS + 요청 번호 + 응답 meta.requestId. 로그인 확인 없음.
	Cors
	// Auth — withAuth: Cors + 회원 토큰-브라우저 대조 + 관리자 메뉴 권한. /user/… 가 아닌 주소는 관리자만.
	Auth
)

// Options — 라우트별 설정(지금 서비스의 `export const config = { api: { bodyParser … } }`)
type Options struct {
	// RawBody — 본문을 읽지 않는다(파일 올리기 등 — 핸들러가 c.R.Body / multipart 를 직접 읽는다)
	RawBody bool
	// BodyLimit — 본문 최대 바이트(기본 1MB, 지금 서비스와 같다)
	BodyLimit int64
	// Changes — 설정·권한을 바꾸는 라우트: GET 이 아닌 요청이 성공하면 app.Changed(캐시 즉시 갱신 + 소켓 알림).
	// 없으면 1분 주기 재읽기까지 옛 값이다 — 새로 만든 관리자가 1분 동안 로그인되지 않았다(2026-10-06 E2E 로 발견, 운영 판에서 옮김)
	Changes bool
}

type Handler func(c *Ctx)

type route struct {
	def   string // "/admin/design/[designId]"
	segs  []string
	guard Guard
	opts  Options
	h     Handler
}

type Router struct {
	App *app.App
	mu  sync.RWMutex
	all []*route
	// 다른 라우트에 안 맞으면(아직 안 옮긴 주소) — 개발 중 확인용. nil 이면 404.
	NotFound http.Handler
}

func NewRouter(a *app.App) *Router { return &Router{App: a} }

// Handle — 라우트 등록. def 는 "/" 로 시작하는 지금 서비스의 경로 모양.
func (r *Router) Handle(def string, g Guard, h Handler, opts ...Options) {
	o := Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.BodyLimit == 0 {
		o.BodyLimit = 1 << 20
	}
	rt := &route{def: def, segs: split(def), guard: g, opts: o, h: h}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.all {
		if x.def == def {
			panic("같은 경로를 두 번 등록했다: " + def)
		}
	}
	r.all = append(r.all, rt)
}

// Routes — 등록된 경로 정의(정렬)
func (r *Router) Routes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.all))
	for _, x := range r.all {
		out = append(out, x.def)
	}
	sort.Strings(out)
	return out
}

func split(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// 경로 앞부분 — 세 주소 모두 같은 핸들러로
var prefixes = []string{"/service", "/api/v1", "/api"}

// stripPrefix — "/service/user/x" → "/user/x". 맞는 앞부분이 없으면 ok=false.
func stripPrefix(p string) (string, string, bool) {
	for _, pre := range prefixes {
		if p == pre || strings.HasPrefix(p, pre+"/") {
			return p[len(pre):], pre, true
		}
	}
	return "", "", false
}

type segKind int

const (
	kStatic segKind = iota
	kParam
	kCatchAll
	kOptionalCatchAll
)

func kindOf(seg string) (segKind, string) {
	switch {
	case strings.HasPrefix(seg, "[[...") && strings.HasSuffix(seg, "]]"):
		return kOptionalCatchAll, seg[5 : len(seg)-2]
	case strings.HasPrefix(seg, "[...") && strings.HasSuffix(seg, "]"):
		return kCatchAll, seg[4 : len(seg)-1]
	case strings.HasPrefix(seg, "[") && strings.HasSuffix(seg, "]"):
		return kParam, seg[1 : len(seg)-1]
	}
	return kStatic, seg
}

// match — 경로 구간들 → 라우트 + 경로 변수(문자열, 나머지는 []any). Next.js 규칙대로 고른다.
func (r *Router) match(parts []string) (*route, map[string]any) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cands := r.all
	return matchAt(cands, parts, 0, map[string]any{})
}

func matchAt(cands []*route, parts []string, depth int, params map[string]any) (*route, map[string]any) {
	// 이 깊이에서 끝난 경우
	if depth == len(parts) {
		for _, rt := range cands {
			if len(rt.segs) == depth {
				return rt, params
			}
		}
		for _, rt := range cands { // [[...x]] 가 0개를 받는 경우
			if len(rt.segs) == depth+1 {
				if k, name := kindOf(rt.segs[depth]); k == kOptionalCatchAll {
					p := copyParams(params)
					_ = name // 없으면 query 에 넣지 않는다(Next 와 같다)
					return rt, p
				}
			}
		}
		return nil, nil
	}
	seg := parts[depth]
	// 1) 고정 이름
	var next []*route
	for _, rt := range cands {
		if len(rt.segs) > depth {
			if k, name := kindOf(rt.segs[depth]); k == kStatic && name == seg {
				next = append(next, rt)
			}
		}
	}
	if len(next) > 0 {
		if rt, p := matchAt(next, parts, depth+1, params); rt != nil {
			return rt, p
		}
	}
	// 2) [변수]
	byName := map[string][]*route{}
	var names []string
	for _, rt := range cands {
		if len(rt.segs) > depth {
			if k, name := kindOf(rt.segs[depth]); k == kParam {
				if _, ok := byName[name]; !ok {
					names = append(names, name)
				}
				byName[name] = append(byName[name], rt)
			}
		}
	}
	for _, name := range names {
		p := copyParams(params)
		p[name] = seg
		if rt, pp := matchAt(byName[name], parts, depth+1, p); rt != nil {
			return rt, pp
		}
	}
	// 3) [...나머지] / [[...나머지]] — 이 구간부터 끝까지
	for _, want := range []segKind{kCatchAll, kOptionalCatchAll} {
		for _, rt := range cands {
			if len(rt.segs) == depth+1 {
				if k, name := kindOf(rt.segs[depth]); k == want {
					p := copyParams(params)
					rest := make([]any, 0, len(parts)-depth)
					for _, s := range parts[depth:] {
						rest = append(rest, s)
					}
					p[name] = rest
					return rt, p
				}
			}
		}
	}
	return nil, nil
}

func copyParams(p map[string]any) map[string]any {
	n := make(map[string]any, len(p)+1)
	for k, v := range p {
		n[k] = v
	}
	return n
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rest, prefix, ok := stripPrefix(req.URL.Path)
	if !ok {
		r.notFound(w, req)
		return
	}
	parts := split(rest)
	for i, s := range parts { // 주소 인코딩 풀기(한글 groupKey 등)
		if u, err := unescape(s); err == nil {
			parts[i] = u
		}
	}
	rt, params := r.match(parts)
	if rt == nil {
		r.notFound(w, req)
		return
	}
	serve(r.App, rt, prefix, params, w, req)
}

func (r *Router) notFound(w http.ResponseWriter, req *http.Request) {
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 Not Found"))
}
