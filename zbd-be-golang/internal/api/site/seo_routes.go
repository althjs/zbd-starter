package site

// 없는 주소 = 404(2026-10-07 E2E J9 — 그 전엔 어떤 주소든 200 이라 검색엔진에 「가짜 페이지」 로 보였다).
// 화면 빌드가 index.html 옆에 남기는 routes.json(scripts/routes-list.mjs — 화면 라우트 폴더 그대로)에 맞는 주소가 없으면 404 + noindex.
// 문서는 그대로 내보내 화면이 「없는 페이지」 를 그린다. routes.json 이 없으면(개발 서버·옛 빌드) 판단하지 않는다(200 그대로).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type routeFile struct {
	mtime  time.Time
	routes [][]string
}

var (
	routesMu    sync.Mutex
	routesCache = map[string]*routeFile{}
)

// loadRoutes — index.html 과 같은 폴더의 routes.json(고친 시각이 같으면 캐시). 없거나 깨졌으면 nil = 판단 안 함
func loadRoutes(indexHTML string) [][]string {
	p := filepath.Join(filepath.Dir(indexHTML), "routes.json")
	st, err := os.Stat(p)
	if err != nil {
		return nil
	}
	routesMu.Lock()
	defer routesMu.Unlock()
	if c := routesCache[p]; c != nil && c.mtime.Equal(st.ModTime()) {
		return c.routes
	}
	b, err := os.ReadFile(p)
	var list []string
	if err != nil || json.Unmarshal(b, &list) != nil || len(list) == 0 {
		return nil
	}
	rs := make([][]string, 0, len(list))
	for _, r := range list {
		rs = append(rs, routeSegs(r))
	}
	routesCache[p] = &routeFile{mtime: st.ModTime(), routes: rs}
	return rs
}

func routeSegs(p string) []string {
	var out []string
	for _, s := range strings.Split(p, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// routeKnown — 주소가 목록의 한 라우트에 맞는가. 순수 함수(정책 테스트)
//
//	[x] 한 칸 · [[x]] 있어도 없어도 · [...x] 나머지 전부(없어도 됨) · 그 밖은 글자 그대로
func routeKnown(routes [][]string, path string) bool {
	segs := routeSegs(path)
	for _, r := range routes {
		if routeMatch(r, segs) {
			return true
		}
	}
	return false
}

func routeMatch(r, segs []string) bool {
	if len(r) == 0 {
		return len(segs) == 0
	}
	h := r[0]
	switch {
	case strings.HasPrefix(h, "[..."):
		return true
	case strings.HasPrefix(h, "[[") && strings.HasSuffix(h, "]]"):
		if routeMatch(r[1:], segs) {
			return true
		}
		return len(segs) > 0 && routeMatch(r[1:], segs[1:])
	case strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]"):
		return len(segs) > 0 && routeMatch(r[1:], segs[1:])
	default:
		return len(segs) > 0 && segs[0] == h && routeMatch(r[1:], segs[1:])
	}
}
