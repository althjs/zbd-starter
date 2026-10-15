package site

// 사이트맵 — GET /sitemap/[hostingId] (옛 주소, robots.txt 의 /service/ 에 막혀 있다) · /seo?p=/sitemap.xml (표준 위치)
// 원본: pages/api/sitemap/[hostingId].ts
//
// 싣는 것: 공통 코드 SEO 항목(사이트맵 제외 아닌 것) · 디자인 페이지 · 공개 앱 상세·제작자 · 누군가 푼 프리셀 판
//         · 「사이트맵 노출」 게시판의 공개 글 · 검색엔진 연동을 끄지 않은 공개 미니홈피. 약관 게시판은 뺀다(noindex).

import (
	"context"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

type siteMapItem struct {
	loc, lastmod string
	priority     float64
}

var reTermsBoard = regexp.MustCompile(`/board/terms(/|$)`)

// isoOr — new Date(ms).toISOString() (못 읽으면 "Invalid Date" 대신 빈 값 — 원본은 RangeError 로 그 묶음을 건너뛴다)
func isoOr(ms float64) string {
	s, _ := isoMs(ms)
	return s
}

// buildSitemapXML — 사이트맵 XML
func buildSitemapXML(ctx context.Context, a *app.App, q db.Querier, hostingID string) (string, error) {
	now := float64(time.Now().UnixMilli())
	conf := a.Settings.Config(hostingID)
	co := a.Settings.Hosting(hostingID)
	if co == nil {
		return "", errText("TypeError: Cannot read properties of undefined (reading 'homepage')")
	}
	var seoCode *jsx.OrderedMap
	for _, x := range arrOf(conf.Path("COMMON_CODES", "items")) {
		if looseEq(asMap(x).V("component"), "SEO") {
			seoCode = asMap(x)
			break
		}
	}
	codeMod := now
	if jsx.Truthy(seoCode.V("updatedAt")) {
		codeMod = jsDateMs(seoCode.V("updatedAt"))
	}
	addCodeLastMod := isoOr(codeMod)

	var items []siteMapItem
	for _, x := range arrOf(seoCode.V("items")) {
		o := asMap(x)
		if !jsx.Truthy(o.V("use")) || jsx.Truthy(o.V("ignoreSitemap")) {
			continue
		}
		pr := parseFloatJS(o.V("priority"))
		if pr == 0 {
			pr = 0.5
		}
		items = append(items, siteMapItem{tmpl(o.V("name")), addCodeLastMod, pr})
	}
	homepage := tmpl(co.V("homepage"))

	// 게시판 — 「사이트맵 노출」 을 켠 게시판의 공개 글
	{
		pgs := conf.Map("POST_GROUPS")
		var keys []string
		seen := map[string]bool{}
		for _, k := range pgs.Keys() {
			vals := []any{pgs.V(k)}
			if arr, ok := pgs.V(k).([]any); ok {
				vals = arr
			}
			for _, x := range vals {
				g := asMap(x)
				if !jsx.Truthy(g.V("groupKey")) || !jsx.Truthy(g.Path("feature", "useSitemap")) {
					continue
				}
				gk := jsx.Str(g.V("groupKey"))
				if !seen[gk] {
					seen[gk] = true
					keys = append(keys, gk)
				}
			}
		}
		if len(keys) > 0 {
			for _, k := range keys {
				items = append(items, siteMapItem{homepage + "board/" + strings.ToLower(k), addCodeLastMod, 0.6})
			}
			var ks []any
			for _, k := range keys {
				ks = append(ks, k)
			}
			rows, err := db.Query(ctx, q, `SELECT postId, groupKey, createdAt, updatedAt FROM posts
           WHERE hostingId = ? AND groupKey IN (?) AND status = 0
           ORDER BY updatedAt DESC LIMIT 2000`, hostingID, ks)
			if err != nil {
				log.Printf("@@ sitemap board err: %v", err)
			} else {
				for _, r := range rows {
					mod := r.V("updatedAt")
					if !jsx.Truthy(mod) {
						mod = r.V("createdAt")
					}
					t := jsDateMs(mod)
					if !jsx.Truthy(mod) {
						t = now
					}
					items = append(items, siteMapItem{homepage + "board/" + strings.ToLower(jsx.Str(r.V("groupKey"))) + "/" + encodeURIComponent(jsx.Str(r.V("postId"))), isoOr(t), 0.6})
				}
			}
		}
	}

	// 약관 게시판 빼기 → 우선순위·수정 시각 내림차순 → 같은 주소는 하나만
	var kept []siteMapItem
	for _, o := range items {
		if !reTermsBoard.MatchString(o.loc) {
			kept = append(kept, o)
		}
	}
	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].priority != kept[j].priority {
			return kept[i].priority > kept[j].priority
		}
		return kept[i].lastmod > kept[j].lastmod
	})
	// 다국어 — 사이트 언어가 둘 이상이면 언어마다 한 줄 + 서로의 대체 주소(seo_i18n.go)
	feature := asMap(conf.Path("SITE", "feature"))
	siteLangs := seoSiteLangs(feature)
	defLang := seoDefaultLang(feature, siteLangs)
	seen := map[string]bool{}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset
  xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
  xmlns:xhtml="http://www.w3.org/1999/xhtml"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
  `)
	for _, o := range kept {
		k := trimRightSlash(o.loc)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		b.WriteString(sitemapURLs(o, siteLangs, defLang))
	}
	b.WriteString(`</urlset>`)
	return b.String(), nil
}

var reFloatPrefix = regexp.MustCompile(`^[ \t\n\r]*([+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?)`)

// parseFloatJS — parseFloat(v) (NaN 이면 0 — 호출부가 || 기본값)
func parseFloatJS(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	m := reFloatPrefix.FindStringSubmatch(tmpl(v))
	if m == nil {
		return 0
	}
	return jsx.Number(m[1])
}

// GET /sitemap/[hostingId] — 옛 주소(?type=json 은 빈 목록)
func sitemapHandler(c *web.Ctx) {
	ctx := c.Context()
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Send(err.Error())
		return
	}
	defer conn.Close()
	cid := c.Q("hostingId")
	if c.Q("type") == "json" {
		if _, err := buildSitemapXML(ctx, c.App, conn, cid); err != nil {
			log.Printf("@ sitemap err: %v", err)
			c.Send(err.Error())
			return
		}
		c.Scb([]any{})
		return
	}
	xml, err := buildSitemapXML(ctx, c.App, conn, cid)
	if err != nil {
		log.Printf("@ sitemap err: %v", err)
		c.Send(err.Error()) // 원본은 오류 스택을 그대로 낸다 — 여기선 첫 줄만
		return
	}
	c.SetHeader("Content-Type", "application/xml")
	c.Send(xml)
}

// sitemapURLs — 주소 하나의 <url> 줄들. 다국어면 언어마다(기본 언어는 원래 주소) hreflang 대체 주소를 같이. 회원 글 주소는 원래 한 줄
func sitemapURLs(o siteMapItem, langs []string, def string) string {
	one := func(loc, alts string) string {
		return "<url>\n  <loc>" + xmlEsc(loc) + "</loc>\n" + alts + "  <lastmod>" + o.lastmod + "</lastmod>\n  <priority>" + jsx.NumStr(o.priority) + "</priority>\n</url>\n"
	}
	path := o.loc
	if i := strings.Index(path, "://"); i >= 0 {
		if j := strings.IndexByte(path[i+3:], '/'); j >= 0 {
			path = path[i+3+j:]
		} else {
			path = "/"
		}
	}
	alts := seoAlternates(o.loc, langs, def)
	if len(alts) == 0 || !i18nSeoPath(path) {
		return one(o.loc, "")
	}
	var al strings.Builder
	for _, a := range alts {
		al.WriteString(`  <xhtml:link rel="alternate" hreflang="` + a.hreflang + `" href="` + xmlEsc(a.href) + "\" />\n")
	}
	var b strings.Builder
	for _, l := range langs {
		b.WriteString(one(langURL(o.loc, l, def), al.String()))
	}
	return b.String()
}

// xmlEsc — XML 글자(& < > ")
func xmlEsc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
