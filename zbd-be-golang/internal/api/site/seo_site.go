package site

// 사이트 공통 구조화 데이터(JSON-LD) — 모든 페이지에 WebSite·Organization·메뉴·이동 경로(원본: pages/api/seo/seo-site-content.ts)
//
// 메뉴의 원본은 GNB 의 공개 메뉴(로그인 전 화면 = 크롤러가 보는 메뉴). 못 읽으면 SEO 항목의 「메뉴 이름(navName)」.
// ⚠ 검색엔진에만 보이는 메뉴를 만들면 클로킹이다 — 화면과 같은 목록만.

import (
	"net/url"
	"regexp"
	"sort"
	"strings"

	"zbd.kr/be/internal/jsx"
)

type navItem struct{ name, path, url string }

// toSeoPath — 전체 주소 → 경로(끝 / 없이, 빈 값이면 /)
func toSeoPath(name string) string {
	if u, err := url.Parse(name); err == nil && u.Scheme != "" && u.Host != "" {
		p := trimRightSlash(u.EscapedPath())
		if p == "" {
			return "/"
		}
		return p
	}
	p := trimRightSlash(name)
	if p == "" {
		return "/"
	}
	return p
}

// getSiteNavItems — SEO 항목 중 메뉴 이름이 있는 것(관리 화면 순서)
func getSiteNavItems(items []any, homepage string) []navItem {
	home := trimRightSlash(homepage)
	var out []navItem
	for _, x := range items {
		o := asMap(x)
		if o == nil || !notFalse(o.V("use")) || jsTrim(tmpl(orAny(o.V("navName")))) == "" {
			continue
		}
		p := toSeoPath(tmpl(o.V("name")))
		u := home + p
		if p == "/" {
			u = home + "/"
		}
		out = append(out, navItem{jsTrim(tmpl(o.V("navName"))), p, u})
	}
	return out
}

// orAny — (v || "")
func orAny(v any) any {
	if !jsx.Truthy(v) {
		return ""
	}
	return v
}

// pageNameOf — 「솔리테어 — 즈비드 스토어」 → 「솔리테어」 : 제목 끝의 사이트 이름 꼬리를 뗀다
func pageNameOf(title, siteName string) string {
	t := jsTrim(title)
	if siteName == "" {
		return t
	}
	re, err := regexp.Compile(`[` + jsSpace + `]*(?:-|—|::|\|)[` + jsSpace + `]*[^-—|]*` + regexp.QuoteMeta(siteName) + `[^-—|]*$`)
	if err != nil {
		return t
	}
	// JS 의 replace 는 맨 앞 일치 하나 — Go 는 가장 왼쪽 일치를 찾는다(같다)
	cut := t
	if loc := re.FindStringIndex(t); loc != nil {
		cut = t[:loc[0]] + t[loc[1]:]
	}
	cut = jsTrim(cut)
	if cut == "" {
		return t
	}
	return cut
}

type siteJSONLdInput struct {
	homepage, siteName, logo, path, pageTitle string
	nav                                       []navItem
}

// buildSiteJSONLd — WebSite · Organization · (메뉴) · (이동 경로)
func buildSiteJSONLd(i siteJSONLdInput) string {
	home := trimRightSlash(i.homepage)
	if home == "" || i.siteName == "" {
		return ""
	}
	homeURL := home + "/"
	org := jsx.Obj("@type", "Organization", "@id", homeURL+"#organization", "name", i.siteName, "url", homeURL)
	if i.logo != "" {
		org.Set("logo", jsx.Obj("@type", "ImageObject", "url", i.logo))
	}
	graph := []any{
		jsx.Obj("@type", "WebSite", "@id", homeURL+"#website", "name", i.siteName, "url", homeURL, "inLanguage", "ko-KR",
			"publisher", jsx.Obj("@id", homeURL+"#organization")),
		org,
	}
	if len(i.nav) > 0 {
		els := []any{}
		for idx, n := range i.nav {
			els = append(els, jsx.Obj("@type", "SiteNavigationElement", "position", float64(idx+1), "name", n.name, "url", n.url))
		}
		graph = append(graph, jsx.Obj("@type", "ItemList", "@id", homeURL+"#navigation", "name", i.siteName+" 메뉴", "itemListElement", els))
	}

	p, _, _ := strings.Cut(i.path, "?")
	if p == "" {
		p = "/"
	}
	p = trimRightSlash(p)
	if p == "" {
		p = "/"
	}
	if p != "/" {
		under := func(x string) bool { return x != "/" && (p == x || strings.HasPrefix(p, x+"/")) }
		var cands []navItem
		for _, n := range i.nav {
			if under(n.path) || under(reAppPrefix.ReplaceAllString(n.path, "$1")) {
				cands = append(cands, n)
			}
		}
		sort.SliceStable(cands, func(a, b int) bool { return jsLen(cands[a].path) > jsLen(cands[b].path) })
		type crumb struct{ name, item string }
		crumbs := []crumb{{i.siteName, homeURL}}
		var section *navItem
		if len(cands) > 0 {
			section = &cands[0]
			crumbs = append(crumbs, crumb{section.name, section.url})
		}
		if section == nil || section.path != p {
			if name := pageNameOf(i.pageTitle, i.siteName); name != "" {
				crumbs = append(crumbs, crumb{name, home + p})
			}
		}
		if len(crumbs) >= 2 {
			els := []any{}
			for idx, c := range crumbs {
				els = append(els, jsx.Obj("@type", "ListItem", "position", float64(idx+1), "name", c.name, "item", c.item))
			}
			graph = append(graph, jsx.Obj("@type", "BreadcrumbList", "itemListElement", els))
		}
	}
	return ltEscape(jsx.Stringify(jsx.Obj("@context", "https://schema.org", "@graph", graph)))
}

// "/app/x" → "/x" (JS 의 /^\/app(?=\/)/ — 앞보기가 없어 / 를 잡아 되돌린다)
var reAppPrefix = regexp.MustCompile(`^/app(/)`)

// buildArticleJSONLd — 게시판 글(공지 등 공개 글)
func buildArticleJSONLd(urlS, title, desc, image string, createdAt, updatedAt any, siteName, homepage string) string {
	homeURL := trimRightSlash(homepage) + "/"
	iso := func(v any) (string, bool) {
		var ms float64
		switch x := v.(type) {
		case string:
			if reDigits.MatchString(x) {
				ms = jsx.Number(x)
			} else {
				ms = jsDateMs(x)
			}
		default:
			ms = jsDateMs(v)
		}
		return isoMs(ms)
	}
	o := jsx.NewOrderedMap()
	o.Set("@context", "https://schema.org")
	o.Set("@type", "Article")
	o.Set("headline", jsSub(title, 110))
	if desc != "" {
		o.Set("description", desc)
	}
	if image != "" {
		o.Set("image", []any{image})
	}
	if s, ok := iso(createdAt); ok {
		o.Set("datePublished", s)
	}
	mod := updatedAt
	if !jsx.Truthy(mod) {
		mod = createdAt
	}
	if s, ok := iso(mod); ok {
		o.Set("dateModified", s)
	}
	o.Set("mainEntityOfPage", urlS)
	o.Set("author", jsx.Obj("@type", "Organization", "name", siteName, "url", homeURL))
	o.Set("publisher", jsx.Obj("@id", homeURL+"#organization"))
	return ltEscape(jsx.Stringify(o))
}

var (
	reDigits   = regexp.MustCompile(`^\d+$`)
	reFirstImg = regexp.MustCompile(`(?i)<img[^>]+src="([^"]+)"`)
	reHTTP     = regexp.MustCompile(`^https?://`)
)

// firstImageOf — 본문 HTML 의 첫 이미지(대표 이미지를 따로 안 둔 글)
func firstImageOf(html, homepage string) string {
	m := reFirstImg.FindStringSubmatch(html)
	if m == nil {
		return ""
	}
	src := m[1]
	if reHTTP.MatchString(src) {
		return src
	}
	if !strings.HasPrefix(src, "/") {
		src = "/" + src
	}
	return trimRightSlash(homepage) + src
}
