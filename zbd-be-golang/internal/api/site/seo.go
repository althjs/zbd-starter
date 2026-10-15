package site

// GET /seo?p=<원 요청 경로> — 사용자 사이트의 문서(검색엔진·공유 미리보기용 내용을 채운 index.html). 원본: pages/api/seo/index.ts
//
// nginx 가 사이트 경로를 ?p= 로 넘긴다. 회사는 Host 로 정한다(talk.zbd.kr = 즈비드의 다른 출처, {주소}.zbd.kr = 미니홈피).
// 페이지마다 제목·설명·이미지·canonical·robots·JSON-LD·<noscript> 본문을 정하고, 화면이 첫 요청을 줄이게 사이트 바탕 데이터
// (window.___zbd_data_model)를 함께 싣는다. 문서 틀은 배포된 index.html(seo_html.go).

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"zbd.kr/be/internal/sitepub"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// 검색 색인에서 뺄 게시판(글·목록 모두 noindex, follow) — 이용약관·개인정보 처리방침
var noindexBoards = []string{"TERMS"}

const talkHost = "talk.zbd.kr"

// 환경 접두사 — 회사는 운영 주소 기준으로 등록돼 있어 떼고 찾는다
var envHostPrefixes = []string{"staging-", "alpha-", "dev-"}

func stripEnvHostPrefix(h string) string {
	for _, p := range envHostPrefixes {
		if strings.HasPrefix(h, p) {
			return h[len(p):]
		}
	}
	return h
}

func isEnvHost(h string) bool {
	for _, p := range envHostPrefixes {
		if strings.HasPrefix(h, p) {
			return true
		}
	}
	return false
}

// ─── 스캐너 탐침 — 조회도 로그도 없이 404 ───

var ignoredPrefixes = []string{"/wp-admin", "/wp-includes", "/wp-content", "/wordpress", "/wp", "/xmlrpc.php", "/phpmyadmin",
	"/phpMyAdmin", "/administrator", "/cgi-bin", "/vendor/", "/.env", "/.git", "/.aws", "/.vscode", "/actuator", "/solr",
	"/hudson", "/jenkins", "/telescope", "/_ignition", "/autodiscover", "/owa/", "/boaform", "/HNAP1"}

var ignoredExts = []string{".php", ".asp", ".aspx", ".jsp", ".cgi", ".env"}

// isIgnoredSeoPath — 우리 서비스에 없는 확장자·디렉터리(경계를 본다 — /wp 가 /wpath 를 잡지 않게)
func isIgnoredSeoPath(p string) bool {
	p, _, _ = strings.Cut(p, "?")
	p = strings.ToLower(p)
	if p == "" {
		return false
	}
	for _, e := range ignoredExts {
		if strings.HasSuffix(p, e) {
			return true
		}
	}
	for _, raw := range ignoredPrefixes {
		prefix := strings.ToLower(raw)
		if strings.HasSuffix(prefix, "/") {
			if strings.HasPrefix(p, prefix) {
				return true
			}
			continue
		}
		if !strings.HasPrefix(p, prefix) {
			continue
		}
		if len(p) == len(prefix) || p[len(prefix)] == '/' || p[len(prefix)] == '.' {
			return true
		}
	}
	return false
}

// seoPathParam — ?p 를 문자열로(원 경로에 p 가 또 있으면 배열 — 첫 원소가 nginx 가 붙인 경로)
func seoPathParam(c *web.Ctx) string {
	switch v := c.QV("p").(type) {
	case nil:
		return ""
	case []any:
		if len(v) == 0 || v[0] == nil {
			return ""
		}
		return jsx.Str(v[0])
	default:
		return jsx.Str(v)
	}
}

// parseQuery — 경로의 쿼리를 {k: v} 로(지금 서비스의 utils.parseQuery 와 같은 규칙 — 첫 조각은 ? 뒤만)
func parseQuery(p string) map[string]string {
	if p == "" {
		return nil
	}
	m := map[string]string{}
	for i, s := range strings.Split(p, "&") {
		if s == "" {
			continue
		}
		if i == 0 {
			parts := strings.Split(s, "?")
			s = ""
			if len(parts) > 1 {
				s = parts[1]
			}
		}
		kv := strings.Split(s, "=")
		v := ""
		if len(kv) > 1 {
			if d, err := url.PathUnescape(kv[1]); err == nil && utf8.ValidString(d) {
				v = d
			} else {
				v = kv[1]
			}
		}
		m[kv[0]] = v
	}
	return m
}

// getHostingID — Host → 회사(회사 목록의 homepage 와 같은 것)
func getHostingID(a *app.App, host string) string {
	if host == "" {
		return ""
	}
	// SITE 설정 domains(사용자·관리자 주소) 먼저 — 없으면 옛 방식(호스팅 목록의 homepage)
	if id := web.HostingIDFor(a, host); id != "" {
		return id
	}
	homepage := "https://" + stripEnvHostPrefix(host) + "/"
	if homepage == "https://"+talkHost+"/" {
		homepage = "https://zbd.kr/"
	}
	for _, x := range a.Settings.Hostings() {
		if m := asMap(x); m != nil && looseEq(m.V("homepage"), homepage) {
			return orStr(m.V("hostingId"))
		}
	}
	return ""
}

// firstHostingID — 설정 HOSTINGS 의 첫 호스팅
func firstHostingID(a *app.App) string {
	for _, x := range a.Settings.Hostings() {
		if m := asMap(x); m != nil {
			if id := orStr(m.V("hostingId")); id != "" {
				return id
			}
		}
	}
	return ""
}

var reMiniSub = regexp.MustCompile(`^([a-z][a-z0-9-]{2,19})\.zbd\.kr$`)

// ─── 페이지 판정 ───

type seoPage struct {
	typ, path, groupKey, postID, handle, shareHashKey string
	userID                                            float64
	freecellGame                                      float64 // 0 이면 아님
}

var (
	reFreecellGame = regexp.MustCompile(`^/app/freecell/(\d{1,7})/?$`)
	reStoreW       = regexp.MustCompile(`^/store/w/([^/]+)$`)
	reStoreU       = regexp.MustCompile(`^/store/u/(\d+)$`)
	reMini         = regexp.MustCompile(`^/(?:mini|life/u)/([^/?#]+)`)
	reBoardPost    = regexp.MustCompile(`^/board/([^/]+)/([^/]+)$`)
	reLeadInt      = regexp.MustCompile(`^[ \t\n\r]*([+-]?\d+)`)
)

// parseIntStr — parseInt(s, 10) 을 문자열로(NaN 이면 "NaN")
func parseIntStr(s string) string {
	m := reLeadInt.FindStringSubmatch(s)
	if m == nil {
		return "NaN"
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return "NaN"
	}
	return jsx.NumStr(n)
}

func parsePage(a *app.App, p, hostingID string) *seoPage {
	page := &seoPage{}
	switch {
	case strings.HasPrefix(p, "/s/"):
		rest := strings.SplitN(p, "/s/", 3)[1]
		page.typ = "LIFE"
		page.shareHashKey = strings.Split(rest, "/")[0]
		return page
	case strings.HasPrefix(p, "/category/"):
		page.typ, page.groupKey = "CATEGORY", "CATEGORY"
		page.path = "category/" + parseIntStr(strings.Split(strings.Replace(p, "/category/", "", 1), "/")[0])
	case strings.HasPrefix(p, "/promotion/"):
		page.typ, page.groupKey = "COMMERCE_CATEGORY", "COMMERCE_CATEGORY"
		page.path = "promotion/" + parseIntStr(strings.Split(strings.Replace(p, "/promotion/", "", 1), "/")[0])
	case reFreecellGame.MatchString(p):
		page.typ, page.path, page.groupKey = "DESIGN_PAGE", "freecell", "PAGE"
		n, _ := strconv.ParseFloat(reFreecellGame.FindStringSubmatch(p)[1], 64)
		page.freecellGame = n // 0 번은 원본도 「판 번호 페이지」 로 보지 않는다(거짓값)
	case strings.HasPrefix(p, "/app/"):
		page.typ, page.groupKey = "DESIGN_PAGE", "PAGE"
		page.path = strings.Split(strings.Split(p, "/app/")[1], "?")[0]
	case reStoreW.MatchString(p):
		page.typ = "STORE_WIDGET"
		page.postID = reStoreW.FindStringSubmatch(p)[1]
	case reStoreU.MatchString(p):
		page.typ = "STORE_CREATOR"
		page.userID, _ = strconv.ParseFloat(reStoreU.FindStringSubmatch(p)[1], 64)
	case reMini.MatchString(p):
		page.typ = "MINI_HOME"
		raw := reMini.FindStringSubmatch(p)[1]
		page.handle = raw
		if d, err := url.PathUnescape(raw); err == nil && utf8.ValidString(d) {
			page.handle = d // 잘못된 % 는 원문 그대로
		}
	case p == "/store" || strings.HasPrefix(p, "/store?"):
		page.typ = "STORE"
	case reBoardPost.MatchString(p):
		page.typ, page.groupKey = "POST_ITEM", "POST"
		page.postID = reBoardPost.FindStringSubmatch(p)[2]
	case strings.HasPrefix(p, "/board/"):
		gk := strings.ToUpper(strings.Split(strings.SplitN(p, "/board/", 3)[1], "?")[0])
		pgs := a.Settings.Config(hostingID).Map("POST_GROUPS")
		var g *jsx.OrderedMap
		for _, k := range []string{"adminPostGroups", "userPostGroups"} {
			for _, x := range pgs.Arr(k) {
				if m := asMap(x); m != nil && looseEq(m.V("groupKey"), gk) {
					g = m
					break
				}
			}
			if g != nil {
				break
			}
		}
		if jsx.Truthy(g.V("useDesignWidget")) {
			page.typ, page.groupKey = "DESIGN_PAGE", "POST"
			page.path = "board/" + strings.ToLower(jsx.Str(g.V("groupKey")))
		}
	case len(strings.Split(p, "/")) == 3:
		page.typ, page.groupKey = "DESIGN_PAGE", "PAGE"
		page.path = strings.Split(strings.Split(p, "/")[2], "?")[0]
	}
	return page
}

// ─── 문서 값 ───

type seoData struct {
	lang       string   // 문서 언어(다국어 — 기본 언어면 "", seo_i18n.go)
	alternates []seoAlt // hreflang 대체 주소

	missing                        bool // 없는 주소 — 404 + noindex(seo_routes.go). 플러그인이 손질하는 줄(cid…)과 떨어뜨려 둔다(zbd update 부딪힘)
	title, desc, tags              any  // 운영자 설정이 비어 있으면 undefined(nil) — 문서에는 기본값이 나간다
	image, html, canonical, robots string
	jsonLd, siteJSONLd, siteName   string
	favicon, appBarBgColor, ga4ID  string
	model                          *jsx.OrderedMap
	cid, baseVer, pageKey          string // 화면이 따로 받는 데이터의 미리 받기 주소(데이터는 HTML 에 싣지 않는다)
}

// tplKey — `${o.k}` (칸이 없으면 "undefined", null 이면 "null")
func tplKey(o *jsx.OrderedMap, k string) string {
	if !o.Has(k) {
		return "undefined"
	}
	return jsx.Str(o.V(k))
}

// 사이트 GNB(홈 디자인이 쓰는 GNB) — 메뉴 구조화 데이터의 원본. 5분 캐시
type gnbEntry struct {
	at  time.Time
	gnb *jsx.OrderedMap
}

var (
	siteGnbMu    sync.Mutex
	siteGnbCache = map[string]gnbEntry{}
)

// doI18n — 언어(?lang=)가 있으면 사이트·페이지 사전으로 제목 등 글자를 바꾼다(문자열 칸, 통째로 같은 글자만)
func doI18n(s *seoData, page *jsx.OrderedMap, siteI18n *jsx.OrderedMap, lang string) {
	if lang == "" || siteI18n == nil {
		return
	}
	dict := append(append([]any{}, arrOf(page.Path("i18n", "dictionary"))...), arrOf(siteI18n.V("dictionary"))...)
	m := map[string]string{}
	for _, x := range dict {
		o := asMap(x)
		ko := jsTrim(orStr(o.V("ko")))
		if ko == "" {
			continue
		}
		m[ko] = orStr(o.V(lang))
	}
	conv := func(v string) (string, bool) {
		if t, ok := m[jsTrim(v)]; ok && t != "" {
			return t, true
		}
		return v, false
	}
	for _, f := range []*any{&s.title, &s.desc, &s.tags} {
		if str, ok := (*f).(string); ok {
			if t, ok := conv(str); ok {
				*f = t
			}
		}
	}
	for _, f := range []*string{&s.image, &s.html, &s.canonical, &s.robots, &s.jsonLd, &s.favicon, &s.appBarBgColor, &s.ga4ID} {
		if t, ok := conv(*f); ok {
			*f = t
		}
	}
}

// seoHandler — /seo
func seoHandler(c *web.Ctx) {
	rawP := seoPathParam(c)
	p, _, _ := strings.Cut(rawP, "?")
	params := parseQuery(rawP)
	if isIgnoredSeoPath(p) {
		c.Status(404).End()
		return
	}
	if params == nil { // ?p 가 없거나 빈 값 — nginx 는 늘 넘긴다. 지금 서비스도 여기서 오류로 끝난다
		sendText(c, "SERVER ERROR")
		return
	}
	ctx := c.Context()
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		sendText(c, "SERVER ERROR")
		return
	}
	defer conn.Close()
	if err := renderSeo(ctx, c, conn, rawP, p, params); err != nil {
		log.Printf("seo/index %s: %v", p, err)
		if !c.Wrote() {
			sendText(c, "SERVER ERROR")
		}
	}
}

// sendText — 문자열 응답(지금 서비스는 text/html 로 낸다)
func sendText(c *web.Ctx, s string) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Send(s)
}

// cdnPrefix — 올린 파일 주소 앞부분(운영 .env CDN_URL_PREFIX, 예 https://도메인/upload/). 없으면 이 서버의 /upload/(개발 모드)
func cdnPrefix(c *web.Ctx) string {
	if v := c.App.Cfg.Get("CDN_URL_PREFIX"); v != "" {
		return v
	}
	return "/upload/"
}

func renderSeo(ctx context.Context, c *web.Ctx, conn *sql.Conn, rawP, p string, params map[string]string) error {
	a := c.App
	cdn := cdnPrefix(c)
	host := c.R.Host
	hostingID := params["hostingId"]
	if hostingID == "" {
		hostingID = getHostingID(a, host)
	}
	if hostingID == "" {
		hostingID = c.Q("hostingId")
	}
	// 관리자 주소(admin.…)는 호스팅 homepage 와 다르다 — 스타터는 사이트가 하나라 첫 호스팅으로
	if hostingID == "" && c.IsAdminHost(host) {
		hostingID = firstHostingID(a)
	}
	full := rawP
	if !strings.HasPrefix(full, "/") {
		full = "/" + full
	}

	if hostingID == "" {
		sendText(c, "hostingId NOT FOUND")
		return nil
	}

	// 사이트 비공개(즈비드 호스팅 — site_open.go): 사용자 주소는 준비 중 화면·robots 전부 막음·사이트맵 없음
	if web.ShowClosedPage(web.IsHostedSite(a.Cfg.Get), web.SiteOpen(a.Settings.Site(hostingID)), c.IsAdminHost(host)) {
		sendClosedPage(c, a.Settings.Site(hostingID), p)
		return nil
	}

	// /sitemap.xml — 표준 위치에서도(옛 /service/sitemap/… 는 robots.txt 에 막혀 검색엔진이 못 읽는다)
	if p == "/sitemap.xml" {
		xml, err := buildSitemapXML(ctx, a, conn, hostingID)
		if err != nil {
			return err
		}
		c.SetHeader("Content-Type", "application/xml")
		c.Send(xml)
		return nil
	}

	// /robots.txt — 스타터 nginx 는 이 주소를 따로 넘기지 않아 화면(HTML)이 나갔다(2026-10-06 E2E J9). 여기서 바로 준다
	if p == "/robots.txt" {
		robotsHandler(c)
		return nil
	}

	page := parsePage(a, p, hostingID)

	base, err := seoBaseData(c, hostingID)
	if err != nil {
		return err
	}
	siteMeta := base.Map("siteMeta")

	// 정규 도메인으로(개발·스테이징 접두사 주소·관리자 주소·개발 주소는 그대로).
	// 개발 주소(localhost·127.0.0.1·사설 IP)를 빼지 않으면 개발 화면이 문서를 127.0.0.1:5001 로 받을 때 홈 주소(localhost:5000)로 넘겨
	// 화면 → 서버 → 화면이 끝없이 302 를 돌았다(2026-10-06 새로 깐 리눅스 — 홈 주소가 DB 에 있으면 난다)
	if !isEnvHost(host) && !c.IsAdminHost(host) && !web.IsLocalHost(host) {
		if hp := orStr(siteMeta.V("homepage")); hp != "" && !strings.Contains(hp, host) {
			c.Redirect(307, hp+strings.TrimPrefix(rawP, "/"))
			return nil
		}
	}

	design := siteMeta.Map("design")
	favicon := design.Map("favicon").V("name")
	image := design.Map("image").V("name")
	globalHTML := orStr(design.V("globalHTML"))
	design.Delete("globalHTML")
	globalHTMLBodyTop := orStr(design.V("globalHTMLBodyTop"))
	design.Delete("globalHTMLBodyTop")

	title := design.V("title")
	if !jsx.Truthy(title) {
		title = siteMeta.V("title")
	}
	s := &seoData{model: jsx.Obj("base", base), title: title}
	s.desc = design.V("description")
	if !jsx.Truthy(s.desc) {
		s.desc = title
	}
	s.tags = design.V("keyword")
	if !jsx.Truthy(s.tags) {
		s.tags = title
	}
	if jsx.Truthy(favicon) {
		s.favicon = cdn + hostingID + "/site/" + jsx.Str(favicon)
	}
	if jsx.Truthy(image) {
		s.image = cdn + hostingID + "/site/" + jsx.Str(image)
	}
	s.appBarBgColor = orStr(design.V("appBarBgColor"))
	s.ga4ID = orStr(siteMeta.V("ga4Id"))
	koBase := seoKoBase{title: s.title, desc: s.desc, tags: s.tags, image: s.image} // 다국어 — 페이지가 따로 정하지 않은 한국어 사이트 문구(seo_i18n.go)

	// 공통 코드 SEO 항목 — 경로가 정확히 같은 것 먼저, 없으면 옛 부분일치(홈은 맨 앞 항목 — 의도된 것)
	var seoCodeItems []any
	for _, x := range arrOf(a.Settings.Config(hostingID).Path("COMMON_CODES", "items")) {
		if looseEq(asMap(x).V("component"), "SEO") {
			seoCodeItems = arrOf(asMap(x).V("items"))
			break
		}
	}
	wantPath := trimRightSlash(p)
	if p == "" || wantPath == "" {
		wantPath = "/"
	}
	var customSeo *jsx.OrderedMap
	for _, x := range seoCodeItems {
		if o := asMap(x); jsx.Truthy(o.V("use")) && toSeoPath(orStr(o.V("name"))) == wantPath {
			customSeo = o
			break
		}
	}
	if customSeo == nil {
		for _, x := range seoCodeItems {
			if o := asMap(x); jsx.Truthy(o.V("use")) && strings.Contains(orStr(o.V("name")), p) {
				customSeo = o
				break
			}
		}
	}
	if customSeo != nil {
		s.title = customSeo.V("title")
		s.desc = customSeo.V("desc")
		if img := customSeo.Map("image"); jsx.Truthy(img.V("name")) {
			s.image = cdn + tplKey(img, "hostingId") + "/seo/" + jsx.Str(img.V("name"))
		}
	}

	homepageRaw := orStr(siteMeta.V("homepage"))
	homepage := trimRightSlash(homepageRaw)
	if homepage != "" {
		if p != "" && p != "/" {
			s.canonical = homepage + p
		} else {
			s.canonical = homepage
		}
	}

	switch page.typ {
	case "POST_ITEM":
		post, err := getPost(ctx, conn, page.postID, hostingID)
		if err != nil {
			return err
		}
		if post == nil {
			break
		}
		gk := strings.ToUpper(orStr(post.V("groupKey")))
		for _, b := range noindexBoards {
			if b == gk {
				s.robots = "noindex, follow" // 약관은 검색 결과(사이트링크)에 나오지 않게 — 링크는 따라가도 된다
			}
		}
		pg := a.Settings.PostGroup(hostingID, jsx.Str(post.V("groupKey")))
		siteName := orStr(siteMeta.V("companyName"))
		if siteName != "" && !strings.Contains(orStr(post.V("title")), siteName) {
			s.title = jsx.Str(post.V("title")) + " - " + siteName
		} else {
			s.title = post.V("title")
		}
		col := orStr(pg.V("dataColumn"))
		if col == "" {
			col = "json"
		}
		jv, err := parseJSONStr(orDefault(post.V(col), "{}"))
		if err != nil {
			return err
		}
		js := asMap(jv)
		img := asMap(js.V("image"))
		if img != nil && jsx.Truthy(img.V("name")) && jsx.Truthy(img.V("imageId")) {
			cid := orStr(img.V("hostingId"))
			if cid == "" {
				cid = hostingID
			}
			s.image = cdn + cid + "/" + jsx.Str(img.V("imageId")) + "/" + encodeURIComponent(jsx.Str(img.V("name")))
		}
		if body, ok := js.V("body").(string); ok && body != "" {
			if sd := jsTrim(orStr(js.V("seoDescription"))); sd != "" {
				s.desc = js.V("seoDescription")
			} else {
				text := jsTrim(strings.ReplaceAll(collapse(reTags.ReplaceAllString(body, "")), "&nbsp;", " "))
				if text != "" {
					if jsLen(text) > 100 {
						text = jsSub(text, 100) + "..."
					}
					s.desc = text
				}
			}
		}
		if jsx.Number(post.V("status")) == 0 {
			bodyImage := firstImageOf(orStr(js.V("body")), homepageRaw)
			if !jsx.Truthy(img.V("name")) && bodyImage != "" {
				s.image = bodyImage
			}
			s.jsonLd = buildArticleJSONLd(s.canonical, orStr(post.V("title")), orStr(s.desc), s.image,
				post.V("createdAt"), post.V("updatedAt"), siteName, homepageRaw)
		}
	}

	// 다국어 — 요청 언어(?lang=)·언어별 운영자 값·hreflang(seo_i18n.go). 사이트 언어가 둘 이상일 때만(플러그인 i18n)
	feature := siteMeta.Map("feature")
	siteLangs := seoSiteLangs(feature)
	defLang := seoDefaultLang(feature, siteLangs)
	reqLang := ""
	if i18nSeoPath(p) {
		reqLang = seoReqLang(params["lang"], siteLangs, defLang)
	}
	doI18n(s, nil, design.Map("i18n"), reqLang)
	if reqLang != "" {
		applyI18nSeo(s, asMap(a.Settings.Meta(hostingID, "I18N_SEO")), reqLang, wantPath, koBase, cdn+hostingID)
		s.lang = reqLang
	}
	if i18nSeoPath(p) && !strings.Contains(s.robots, "noindex") && !s.missing {
		s.alternates = seoAlternates(s.canonical, siteLangs, defLang)
		if reqLang != "" {
			s.canonical = langURL(s.canonical, reqLang, defLang)
		}
	}

	// 사이트 공통 구조화 데이터 — 메뉴는 공통 코드 SEO 항목의 메뉴 이름(navName)
	s.siteName = orStr(siteMeta.V("companyName"))
	nav := getSiteNavItems(seoCodeItems, homepageRaw)
	pth := p
	if pth == "" {
		pth = "/"
	}
	s.siteJSONLd = buildSiteJSONLd(siteJSONLdInput{homepage: homepageRaw, siteName: s.siteName, logo: s.favicon, path: pth,
		pageTitle: tmpl(orAny(s.title)), nav: nav})

	idxPath := indexPath(c, "FE_INDEX_HTML", "./data/fe/index.html")
	// 없는 주소 — 화면 라우트에 없거나(routes.json) 등록 안 된 디자인 페이지(/app/…)면 404 + noindex. 문서는 그대로(화면이 「없는 페이지」)
	if rs := loadRoutes(idxPath); rs != nil && !routeKnown(rs, p) {
		s.missing = true
	}
	if s.missing {
		s.robots = "noindex, follow"
	}
	index, err := readIndexHTML(idxPath)
	if err != nil {
		return err
	}
	s.cid = hostingID
	s.baseVer = sitepub.Version(sitepub.PublicBase(s.model.Map("base")))
	parts := buildDoc(s, globalHTML, "", "", globalHTMLBodyTop)
	// 관리자 주소 목록(SITE 설정 domains — 화면 주소 검사·대표 관리자 주소·단독 앱 주소) — 관리자 주소로 온 문서에만(사용자 화면 소스에 관리자 주소를 싣지 않는다)
	//   zbd:host-realm = 이 주소가 관리자(admin)·사용자(user) 어느 쪽인가 — 화면이 맞지 않는 묶음을 그리지 않는다. 개발 한 주소(localhost)는 싣지 않는다(검사 끔)
	if ah := c.AdminHostsFor(host); ah != nil {
		parts.data += `<meta name="zbd:host-realm" content="admin">` + "\n" + `<meta name="zbd:admin-hosts" content="` + escAttr(strings.Join(ah, ",")) + `">` + "\n"
	} else if !web.IsLocalHost(host) {
		parts.data += `<meta name="zbd:host-realm" content="user">` + "\n"
		// 디자인 미리보기 창만 — 관리자 편집 화면이 보내는 내용을 받으려면 관리자 주소를 알아야 한다(다른 사용자 화면에는 싣지 않는다)
		if strings.HasPrefix(p, "/widget-preview-iframe") {
			if ah := c.AdminHostsOfUserHost(host); ah != nil {
				parts.data += `<meta name="zbd:admin-hosts" content="` + escAttr(strings.Join(ah, ",")) + `">` + "\n"
			}
		}
	}
	html := renderDoc(index, parts)
	html = deferFontCSS(html)
	c.SetHeader("Content-Type", "text/html; charset=UTF-8")
	c.SetHeader("Cache-Control", "no-cache") // 화면 배포가 곧바로 보이게 — 옛 index.html(옛 청크 목록) 을 쓰지 않게
	if s.missing {
		c.Status(404)
	}
	c.Send(html)
	return nil
}

// indexPath — 설정의 index.html 경로
func indexPath(c *web.Ctx, key, def string) string {
	if v := c.App.Cfg.Get(key); v != "" {
		return v
	}
	return def
}

// reFontCSS — 운영자 머리말의 바깥 글꼴 CSS(<link … fonts.googleapis.com … rel="stylesheet">)
var reFontCSS = regexp.MustCompile(`<link([^>]*?)href="(https://fonts\.googleapis\.com/[^"]+)"([^>]*?)rel="stylesheet"([^>]*)>`)

// deferFontCSS — 바깥 글꼴 CSS 가 그리기를 막지 않게(받는 대로 적용). 휴대폰에서 매번 새로 접속하느라 첫 그림이 0.5초 늦어
// 설치 앱 시작 때 창이 투명하게 비었다(2026-10-03 대표). 글꼴은 원래 display=swap 이라 늦게 바뀌어도 된다. 스크립트가 꺼진 곳은 noscript 로 그대로.
func deferFontCSS(html string) string {
	return reFontCSS.ReplaceAllString(html, `<link${1}href="${2}"${3}rel="stylesheet" media="print" onload="this.media='all'"${4}><noscript><link href="${2}" rel="stylesheet"></noscript>`)
}

// seoBaseData — 문서에 실을 사이트 바탕(사이트 설정·게시판 목록·공통 코드). 화면 base 와 같은 지문(sitepub.Version)을 쓴다
func seoBaseData(c *web.Ctx, hostingID string) (*jsx.OrderedMap, error) {
	meta, _ := jsx.Clone(c.App.Settings.Config(hostingID)).(*jsx.OrderedMap)
	siteMeta := asMap(meta.V("SITE"))
	if siteMeta == nil {
		return nil, errText("TypeError: Cannot convert undefined or null to object")
	}
	if siteMeta.Map("design") == nil {
		siteMeta.Set("design", jsx.NewOrderedMap())
	}
	base := jsx.NewOrderedMap()
	base.Set("siteMeta", siteMeta)
	if pg := asMap(meta.V("POST_GROUPS")); pg != nil {
		base.Set("postGroupsMeta", pg)
	}
	cc := meta.Path("COMMON_CODES", "items")
	if !jsx.Truthy(cc) {
		cc = []any{}
	}
	base.Set("commonCodes", cc)
	return base, nil
}
