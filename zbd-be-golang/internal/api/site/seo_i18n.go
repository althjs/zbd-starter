package site

// 다국어 SEO(2026-10-08 대표 — 「언어별 로고·SEO 커스텀 · 서비스 전체 다국어 SEO」).
//
// 언어별 주소 = 같은 주소 + ?lang={코드}(기본 언어는 붙이지 않는다). 사이트 언어 = SITE.feature.i18n(둘 이상일 때만), 기본 = feature.defaultLanguage.
// 문서에는 <html lang> · canonical(그 언어 주소) · hreflang 대체 주소(모든 언어 + x-default) · og:locale 을 넣고, 사이트맵은 언어마다 한 줄씩 대체 주소와 함께.
//
// 언어별 값 = 설정 I18N_SEO(관리자 「다국어 › 언어별 로고·SEO」, 플러그인 i18n):
//
//	{ "en": { "title", "description", "keyword", "image": {name}, "pages": { "/store": {"title","desc","keyword"} } }, … }
//
// 고르는 순서: 그 주소의 언어별 값(pages) → 디자인·사이트 사전 번역(doI18n) → 아직 한국어 사이트 기본 문구 그대로면 그 언어의 사이트 문구.
// 언어별 로고는 화면이 SITE.design.i18nLogo 로 그린다(사용자 기본 데이터에 실려야 해서 SITE 에 둔다).
//
// 회원이 쓴 글(미니홈피·공유 링크·게시글·스토어 앱)은 언어 주소를 만들지 않는다 — 본문이 한 언어라 같은 글이 언어 수만큼 겹친다.

import (
	"strings"

	"zbd.kr/be/internal/jsx"
)

// seoSiteLangs — 사이트 언어(옛 값 kor 는 뺀다). 둘 미만이면 nil(다국어 안 씀)
func seoSiteLangs(feature *jsx.OrderedMap) []string {
	var out []string
	seen := map[string]bool{}
	for _, x := range arrOf(feature.V("i18n")) {
		l, _ := x.(string)
		l = strings.TrimSpace(l)
		if l == "" || l == "kor" || seen[l] || !reHTMLLangCode.MatchString(l) {
			continue
		}
		seen[l] = true
		out = append(out, l)
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

// seoDefaultLang — 기본 언어(사이트 언어에 없으면 첫 언어, 그것도 없으면 ko)
func seoDefaultLang(feature *jsx.OrderedMap, langs []string) string {
	d := strings.TrimSpace(orStr(feature.V("defaultLanguage")))
	for _, l := range langs {
		if l == d {
			return d
		}
	}
	for _, l := range langs {
		if l == "ko" {
			return "ko"
		}
	}
	if len(langs) > 0 {
		return langs[0]
	}
	return "ko"
}

// seoReqLang — 요청 언어(?lang=). 사이트 언어가 아니거나 기본 언어면 "" (기본 주소)
func seoReqLang(param string, langs []string, def string) string {
	param = strings.TrimSpace(param)
	if param == "" || param == def {
		return ""
	}
	for _, l := range langs {
		if l == param {
			return l
		}
	}
	return ""
}

// i18nSeoPath — 언어 주소를 만드는 주소인가(회원 글·공유·로그인 화면은 아니다). 순수 함수
func i18nSeoPath(p string) bool {
	p = trimRightSlash(p)
	if p == "" {
		return true
	}
	for _, pre := range []string{"/mini/", "/life/", "/s/", "/store/w/", "/store/u/", "/talk", "/login", "/user/"} {
		if strings.HasPrefix(p+"/", pre) || strings.HasPrefix(p, pre) {
			return false
		}
	}
	// 게시글 /board/{게시판}/{번호}
	if strings.HasPrefix(p, "/board/") && len(strings.Split(p, "/")) >= 4 {
		return false
	}
	return true
}

// langURL — 그 언어의 주소(기본 언어는 그대로, 나머지는 ?lang= · 이미 ? 가 있으면 &). 순수 함수
func langURL(u, lang, def string) string {
	if u == "" || lang == "" || lang == def {
		return u
	}
	if i := strings.Index(u, "://"); i >= 0 && !strings.Contains(u[i+3:], "/") {
		u += "/" // 홈 https://zbd.kr → https://zbd.kr/?lang=en
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + "lang=" + lang
}

// hreflangOf — 사이트 언어 코드 → hreflang(중국어는 옛 코드 cn)
func hreflangOf(l string) string {
	if l == "cn" {
		return "zh"
	}
	return l
}

var ogLocales = map[string]string{
	"ko": "ko_KR", "en": "en_US", "ja": "ja_JP", "cn": "zh_CN", "zh": "zh_CN", "fr": "fr_FR", "vi": "vi_VN",
	"th": "th_TH", "id": "id_ID", "es": "es_ES", "de": "de_DE", "ru": "ru_RU",
}

// ogLocaleOf — og:locale 값(모르는 코드면 "")
func ogLocaleOf(l string) string { return ogLocales[l] }

type seoAlt struct{ hreflang, href string }

// seoAlternates — 모든 언어 주소 + x-default(기본 언어 주소). 순수 함수
func seoAlternates(base string, langs []string, def string) []seoAlt {
	if base == "" || len(langs) < 2 {
		return nil
	}
	var out []seoAlt
	for _, l := range langs {
		out = append(out, seoAlt{hreflangOf(l), langURL(base, l, def)})
	}
	return append(out, seoAlt{"x-default", base})
}

// seoKoBase — 언어별 값을 고르기 전의 한국어 사이트 기본 문구(이것과 같으면 「페이지가 따로 정하지 않았다」)
type seoKoBase struct {
	title, desc, tags any
	image             string
}

// applyI18nSeo — 그 언어의 운영자 값으로 제목·설명·키워드·OG 이미지를 바꾼다(빈 칸은 그대로)
func applyI18nSeo(s *seoData, conf *jsx.OrderedMap, lang, path string, ko seoKoBase, imgPrefix string) {
	L := conf.Map(lang)
	if L == nil || lang == "" {
		return
	}
	str := func(v any) string { return strings.TrimSpace(orStr(v)) }
	same := func(a, b any) bool { return strings.TrimSpace(orStr(a)) == strings.TrimSpace(orStr(b)) }
	if pg := L.Map("pages").Map(seoPathKey(path)); pg != nil {
		if v := str(pg.V("title")); v != "" {
			s.title = v
		}
		if v := str(pg.V("desc")); v != "" {
			s.desc = v
		}
		if v := str(pg.V("keyword")); v != "" {
			s.tags = v
		}
		if n := str(pg.Path("image", "name")); n != "" {
			s.image = imgPrefix + "/seo/" + n
		}
	}
	if v := str(L.V("title")); v != "" && same(s.title, ko.title) {
		s.title = v
	}
	if v := str(L.V("description")); v != "" && (same(s.desc, ko.desc) || same(s.desc, ko.title)) {
		s.desc = v
	}
	if v := str(L.V("keyword")); v != "" && (same(s.tags, ko.tags) || same(s.tags, ko.title)) {
		s.tags = v
	}
	if n := str(L.Path("image", "name")); n != "" && s.image == ko.image {
		s.image = imgPrefix + "/site/" + n
	}
}

// seoPathKey — I18N_SEO pages 의 열쇠(끝 / 없이, 홈은 /)
func seoPathKey(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	p = trimRightSlash(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// i18nHead — 문서 머리말에 더할 줄(hreflang · og:locale)
func i18nHead(s *seoData) string {
	var b strings.Builder
	for _, a := range s.alternates {
		b.WriteString(`    <link rel="alternate" hreflang="` + a.hreflang + `" href="` + escAttr(a.href) + "\" />\n")
	}
	if loc := ogLocaleOf(s.lang); loc != "" && len(s.alternates) > 0 {
		b.WriteString(`    <meta property="og:locale" content="` + loc + "\" />\n")
		for _, a := range s.alternates {
			if a.hreflang == "x-default" || a.hreflang == hreflangOf(s.lang) {
				continue
			}
			l := a.hreflang
			if l == "zh" {
				l = "cn"
			}
			if o := ogLocaleOf(l); o != "" {
				b.WriteString(`    <meta property="og:locale:alternate" content="` + o + "\" />\n")
			}
		}
	}
	return b.String()
}
