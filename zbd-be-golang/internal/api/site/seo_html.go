package site

// 검색엔진용 HTML — **배포된 index.html 을 읽어 필요한 부분만 끼워 넣는다**(대표 결정 2026-10-03).
//
// 지금 서비스는 빌드 청크 목록(manifest)을 읽어 문서를 통째로 새로 만들었다 → 빌드 산출물과 어긋나면 흰 화면.
// 여기서는 화면 빌드가 낸 index.html 을 그대로 쓰고 제목·설명·OG·canonical·robots·JSON-LD·첫 화면 데이터·<noscript> 만 바꾼다.
//
// 넣는 자리:
//   - 자리표가 있으면 그 자리(새 화면 빌드 — <!--zbd:head--> · <!--zbd:body--> · <!--zbd:data-->)
//   - 없으면 고정 위치(지금 React 빌드): head 묶음·데이터는 </head> 앞, 사이트 본문 위 HTML 은 <body> 바로 뒤, <noscript> 는 </body> 앞
// 어느 쪽이든 index.html 에 원래 있던 제목·설명·OG·canonical·icon 은 지운다(같은 태그가 둘이면 검색엔진·미리보기가 앞의 것을 쓴다).
//
// 파일은 설정 FE_INDEX_HTML(기본 ./data/fe/index.html), 관리자는 FE_ADMIN_INDEX_HTML. 고친 시각이 바뀌면 다시 읽는다.

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"zbd.kr/be/internal/jsx"
)

const (
	markHead = "<!--zbd:head-->"
	markBody = "<!--zbd:body-->"
	markData = "<!--zbd:data-->"
)

type indexFile struct {
	mtime time.Time
	size  int64
	html  string
}

var (
	indexMu    sync.Mutex
	indexCache = map[string]*indexFile{}
)

// readIndexHTML — index.html 내용(고친 시각·크기가 같으면 캐시)
func readIndexHTML(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	indexMu.Lock()
	c := indexCache[path]
	indexMu.Unlock()
	if c != nil && c.mtime.Equal(st.ModTime()) && c.size == st.Size() {
		return c.html, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	indexMu.Lock()
	indexCache[path] = &indexFile{mtime: st.ModTime(), size: st.Size(), html: string(b)}
	indexMu.Unlock()
	return string(b), nil
}

var (
	reTitleTag  = regexp.MustCompile(`(?is)<title[^>]*>.*?</title>\s*`)
	reMetaTag   = regexp.MustCompile(`(?is)<meta\b[^>]*>\s*`)
	reLinkTag   = regexp.MustCompile(`(?is)<link\b[^>]*>\s*`)
	reMetaKey   = regexp.MustCompile(`(?is)\b(?:name|property)\s*=\s*["']([^"']*)["']`)
	reLinkRel   = regexp.MustCompile(`(?is)\brel\s*=\s*["']([^"']*)["']`)
	reHTMLLang  = regexp.MustCompile(`(?is)<html\b[^>]*>`)
	reLangAttr  = regexp.MustCompile(`(?is)\slang\s*=\s*("[^"]*"|'[^']*'|[^\s>]*)`)
	reBodyOpen  = regexp.MustCompile(`(?is)<body\b[^>]*>`)
	reManifest  = regexp.MustCompile(`(<link rel="manifest" href=")/manifest\.json(")`)
	replacedTag = map[string]bool{
		"title": true, "description": true, "keywords": true, "robots": true, "theme-color": true,
		"apple-mobile-web-app-status-bar-style": true, "naver-site-verification": true,
	}
)

// stripSeoTags — index.html 에 원래 있던 문서 정보(제목·설명·OG·twitter·canonical·icon)를 지운다
func stripSeoTags(html string) string {
	html = reTitleTag.ReplaceAllString(html, "")
	html = reMetaTag.ReplaceAllStringFunc(html, func(t string) string {
		m := reMetaKey.FindStringSubmatch(t)
		if m == nil {
			return t
		}
		k := strings.ToLower(m[1])
		if replacedTag[k] || strings.HasPrefix(k, "og:") || strings.HasPrefix(k, "twitter:") {
			return ""
		}
		return t
	})
	html = reLinkTag.ReplaceAllStringFunc(html, func(t string) string {
		m := reLinkRel.FindStringSubmatch(t)
		if m == nil {
			return t
		}
		switch strings.ToLower(m[1]) {
		case "canonical", "icon", "shortcut icon":
			return ""
		}
		return t
	})
	return html
}

// setHTMLLang — <html lang="…">
func setHTMLLang(html, lang string) string {
	loc := reHTMLLang.FindStringIndex(html)
	if loc == nil {
		return html
	}
	tag := html[loc[0]:loc[1]]
	var nt string
	if reLangAttr.MatchString(tag) {
		nt = reLangAttr.ReplaceAllString(tag, ` lang="`+lang+`"`)
	} else {
		nt = strings.Replace(tag, "<html", `<html lang="`+lang+`"`, 1)
	}
	return html[:loc[0]] + nt + html[loc[1]:]
}

// insertBefore — 마지막 needle 앞에(없으면 끝에)
func insertBefore(html, needle, s string) string {
	i := strings.LastIndex(strings.ToLower(html), needle)
	if i < 0 {
		return html + s
	}
	return html[:i] + s + html[i:]
}

// renderDoc — 문서 조각들을 index.html 에 끼운다
type docParts struct {
	lang     string
	head     string // 제목·meta·link·JSON-LD·사이트 전역 HTML·템플릿 CSS·GA
	data     string // <script> 첫 화면 데이터 </script>
	bodyTop  string // 사이트 본문 위 HTML
	noscript string // <noscript>…</noscript>
}

func renderDoc(index string, p docParts) string {
	html := stripSeoTags(index)
	html = setHTMLLang(html, p.lang)

	if strings.Contains(html, markHead) {
		html = strings.Replace(html, markHead, p.head, 1)
	} else {
		html = insertBefore(html, "</head>", p.head)
	}
	if strings.Contains(html, markData) {
		html = strings.Replace(html, markData, p.data, 1)
	} else {
		html = insertBefore(html, "</head>", p.data)
	}
	if strings.Contains(html, markBody) {
		html = strings.Replace(html, markBody, p.bodyTop+p.noscript, 1)
	} else {
		if p.bodyTop != "" {
			if loc := reBodyOpen.FindStringIndex(html); loc != nil {
				html = html[:loc[1]] + "\n" + p.bodyTop + html[loc[1]:]
			}
		}
		html = insertBefore(html, "</body>", p.noscript)
	}
	return html
}

// defaultSeo — 페이지·사이트에 SEO 값이 하나도 없을 때(새로 설치한 사이트). 관리자 「환경 설정」·페이지 SEO 에서 채운다
var defaultSeo = struct{ title, desc, tags, image string }{
	title: "내 사이트",
	desc:  "",
	tags:  "",
	image: "",
}

var reHTMLLangCode = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// toHTMLLang — <html lang> 값(언어코드 모양이 아니면 ko — 사전 변형 키가 같은 칸에 들어온다)
func toHTMLLang(v any) string {
	s := ""
	if v != nil {
		s = strings.TrimSpace(jsx.Str(v))
	}
	if s == "" || !reHTMLLangCode.MatchString(s) {
		return "ko"
	}
	return s
}

// docLang — <html lang>: 요청 언어(다국어)가 있으면 그것, 없으면 사이트 기본 언어
func docLang(s *seoData, siteMeta *jsx.OrderedMap) string {
	if s.lang != "" {
		return toHTMLLang(hreflangOf(s.lang))
	}
	return toHTMLLang(hreflangOf(orStr(siteMeta.Path("feature", "defaultLanguage"))))
}

// buildDoc — seo 값으로 문서 조각을 만든다(지금 서비스 getHTML 과 같은 내용)
func buildDoc(s *seoData, globalHTML, templateCSS, preloadTags, globalHTMLBodyTop string) docParts {
	nl := func(v string) string { return strings.ReplaceAll(v, "\n", "") }
	title := escAttr(nl(orDefault(s.title, defaultSeo.title)))
	desc := escAttr(nl(orDefault(s.desc, defaultSeo.desc)))
	tags := nl(orDefault(s.tags, defaultSeo.tags))
	image := nl(orDefault(s.image, defaultSeo.image))
	siteMeta := s.model.Map("base").Map("siteMeta")
	canonical := escAttr(s.canonical)
	ogURL := canonical
	if ogURL == "" {
		ogURL = orStr(siteMeta.V("homepage"))
	}

	var h strings.Builder
	w := func(x ...string) {
		for _, v := range x {
			h.WriteString(v)
		}
	}
	w("\n    <meta name=\"title\" content=\"", title, "\" />\n")
	w("    <meta name=\"description\" content=\"", desc, "\" />\n")
	w("    <meta name=\"keywords\" content=\"", tags, "\" />\n")
	w("    <meta property=\"og:type\" content=\"website\" />\n")
	w("    <meta property=\"og:url\" content=\"", ogURL, "\" />\n")
	if s.siteName != "" {
		w("    <meta property=\"og:site_name\" content=\"", s.siteName, "\" />\n")
	}
	w("    <meta property=\"og:title\" content=\"", title, "\" />\n")
	w("    <meta property=\"og:image\" content=\"", escAttr(absURL(image, ogURL)), "\" />\n")
	w("    <meta property=\"og:description\" content=\"", desc, "\" />\n")
	if k := siteMeta.V("webMasterToolNaverKey"); jsx.Truthy(k) {
		w("    <meta name=\"naver-site-verification\" content=\"", jsx.Str(k), "\" />\n")
	}
	w("    <title>", title, "</title>\n")
	if s.appBarBgColor != "" {
		w("    <meta name=\"theme-color\" content=\"", s.appBarBgColor, "\" />\n")
		w("    <meta name=\"apple-mobile-web-app-status-bar-style\" content=\"", s.appBarBgColor, "\" />\n")
	}
	if canonical != "" {
		w("    <link rel=\"canonical\" href=\"", canonical, "\" />\n")
	}
	w(i18nHead(s))
	if s.robots != "" {
		w("    <meta name=\"robots\" content=\"", escAttr(s.robots), "\" />\n")
	}
	if s.favicon != "" {
		w("    <link rel=\"icon\" href=\"", s.favicon, "\" type=\"image/png\">\n")
	}
	if s.siteJSONLd != "" {
		w("    <script type=\"application/ld+json\">", s.siteJSONLd, "</script>\n")
	}
	if s.jsonLd != "" {
		w("    <script type=\"application/ld+json\">", s.jsonLd, "</script>\n")
	}
	if preloadTags != "" {
		w("    ", preloadTags, "\n")
	}
	if globalHTML != "" {
		w(globalHTML, "\n")
	}
	w("    <style>", templateCSS, "</style>\n")
	w("    <script>\n    window.dataLayer = window.dataLayer || []\n    function gtag() {\n      dataLayer.push(arguments)\n    }\n    </script>\n")
	if s.ga4ID != "" {
		w(`    <script async src="https://www.googletagmanager.com/gtag/js?id=`, s.ga4ID, "\"></script>\n")
		w("    <script>\n    gtag(\"js\", new Date())\n    gtag(\"config\", \"", s.ga4ID, "\")\n    </script>\n")
	}

	// 첫 화면 데이터는 HTML 에 싣지 않는다(2026-10-03 대표 — 소스보기에 사이트 설정이 통째로 보였다).
	// 화면이 따로 받는 주소를 미리 받기로 걸어 앱 스크립트와 동시에 받기 시작한다 — base 는 지문 주소라 길게 캐시된다.
	var d strings.Builder
	// 로그인 쿠키 도메인(호스팅 SITE 설정 cookieHost) — 화면이 브라우저 id·돌아갈 주소 쿠키를 대표 도메인에 심는다(세션 쿠키와 같은 범위)
	if ch := strings.TrimSpace(jsx.StrOrEmpty(siteMeta.V("cookieHost"))); ch != "" {
		d.WriteString(`<meta name="zbd:cookie-host" content="` + escAttr(ch) + `">` + "\n")
	}
	if s.cid != "" {
		d.WriteString(`<link rel="preload" id="zbd-base" as="fetch" crossorigin="anonymous" href="/service/design/` + s.cid + `/base?v=` + s.baseVer + `">` + "\n")
		if s.pageKey != "" {
			d.WriteString(`<link rel="preload" id="zbd-page" as="fetch" crossorigin="anonymous" href="/service/design/` + s.cid + `/` + escAttr(s.pageKey) + `">` + "\n")
		}
	}

	ns := s.html
	if ns == "" {
		ns = "<p>" + desc + "</p>"
	}
	return docParts{
		lang:     docLang(s, siteMeta),
		head:     h.String(),
		data:     d.String(),
		bodyTop:  globalHTMLBodyTop,
		noscript: "<noscript>\n      <div class=\"seo-content\">" + ns + "</div>\n    </noscript>\n",
	}
}

// orDefault — (v || d) 문자열
// absURL — 「/upload/…」 같은 상대 주소를 사이트 주소(base)로 완성한다. 카톡·페이스북 미리보기는 og:image 가 절대 주소여야 읽는다.
func absURL(v, base string) string {
	if !strings.HasPrefix(v, "/") || strings.HasPrefix(v, "//") {
		return v
	}
	i := strings.Index(base, "://")
	if i < 0 {
		return v
	}
	if j := strings.Index(base[i+3:], "/"); j >= 0 {
		base = base[:i+3+j]
	}
	return base + v
}

func orDefault(v any, d string) string {
	if !jsx.Truthy(v) {
		return d
	}
	return jsx.Str(v)
}
