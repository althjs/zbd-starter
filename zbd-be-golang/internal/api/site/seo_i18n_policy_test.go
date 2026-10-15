package site

import (
	"strings"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 다국어 SEO(2026-10-08 대표) — 언어 주소·hreflang·언어별 운영자 값·회원 글 제외
func TestSeoI18nPolicy(t *testing.T) {
	f := jsx.ParseMap(`{"i18n":["ko","en","kor","ja","cn"],"defaultLanguage":"ko"}`)
	langs := seoSiteLangs(f)
	if strings.Join(langs, ",") != "ko,en,ja,cn" {
		t.Fatalf("사이트 언어(kor 뺌): %v", langs)
	}
	if seoSiteLangs(jsx.ParseMap(`{"i18n":["ko"]}`)) != nil {
		t.Error("언어 하나면 다국어 아님")
	}
	def := seoDefaultLang(f, langs)
	if def != "ko" {
		t.Error("기본 언어 ko")
	}
	for in, want := range map[string]string{"en": "en", "ko": "", "": "", "xx": "", "cn": "cn"} {
		if got := seoReqLang(in, langs, def); got != want {
			t.Errorf("요청 언어 %q → %q, want %q", in, got, want)
		}
	}
	if langURL("https://zbd.kr/app/store", "en", "ko") != "https://zbd.kr/app/store?lang=en" || langURL("https://zbd.kr/x?a=1", "ja", "ko") != "https://zbd.kr/x?a=1&lang=ja" || langURL("https://zbd.kr/x", "ko", "ko") != "https://zbd.kr/x" || langURL("https://zbd.kr", "en", "ko") != "https://zbd.kr/?lang=en" {
		t.Error("언어 주소")
	}
	alts := seoAlternates("https://zbd.kr/app/store", langs, def)
	if len(alts) != 5 || alts[0].href != "https://zbd.kr/app/store" || alts[3].hreflang != "zh" || alts[4].hreflang != "x-default" || alts[4].href != "https://zbd.kr/app/store" {
		t.Errorf("대체 주소: %+v", alts)
	}
	if seoAlternates("https://zbd.kr/", []string{"ko"}, "ko") != nil {
		t.Error("언어 하나면 대체 주소 없음")
	}
	for p, want := range map[string]bool{"/": true, "": true, "/app/store": true, "/board/notice": true, "/board/notice/abc123": false,
		"/mini/claude": false, "/s/xyz": false, "/store/w/abc": false, "/life/today": false, "/login": false, "/talk": false, "/app/freecell/12": true} {
		if i18nSeoPath(p) != want {
			t.Errorf("언어 주소 대상 %q → %v", p, !want)
		}
	}
}

func TestApplyI18nSeoPolicy(t *testing.T) {
	conf := jsx.ParseMap(`{"en":{"title":"ZBD — private life log","description":"EN desc","keyword":"life log","image":{"name":"en.png"},
		"pages":{"/app/store":{"title":"Store — ZBD","desc":"Apps"}}}}`)
	ko := seoKoBase{title: "즈비드", desc: "설명", tags: "태그", image: "P/site/ko.png"}

	// 페이지가 따로 정하지 않은 주소 → 그 언어의 사이트 문구·이미지
	s := &seoData{title: "즈비드", desc: "설명", tags: "태그", image: "P/site/ko.png"}
	applyI18nSeo(s, conf, "en", "/app/life", ko, "P")
	if s.title != "ZBD — private life log" || s.desc != "EN desc" || s.tags != "life log" || s.image != "P/site/en.png" {
		t.Errorf("사이트 문구: %+v", s)
	}
	// 그 주소의 언어별 값이 이긴다(빈 칸은 그대로)
	s = &seoData{title: "스토어", desc: "앱", tags: "태그", image: "P/seo/store.png"}
	applyI18nSeo(s, conf, "en", "/app/store/", ko, "P")
	if s.title != "Store — ZBD" || s.desc != "Apps" || s.image != "P/seo/store.png" {
		t.Errorf("주소별 값: %+v", s)
	}
	// 페이지가 따로 정한 한국어 제목(사전 번역도 없음)은 사이트 문구로 덮지 않는다
	s = &seoData{title: "프리셀 12번 판", desc: "설명", tags: "태그"}
	applyI18nSeo(s, conf, "en", "/app/freecell/12", ko, "P")
	if s.title != "프리셀 12번 판" || s.desc != "EN desc" {
		t.Errorf("페이지 제목 유지: %+v", s)
	}
	// 값이 없는 언어는 아무것도 안 바꾼다
	s = &seoData{title: "즈비드"}
	applyI18nSeo(s, conf, "ja", "/", ko, "P")
	if s.title != "즈비드" {
		t.Error("값 없는 언어")
	}
}

func TestSitemapI18nPolicy(t *testing.T) {
	o := siteMapItem{"https://zbd.kr/app/store", "2026-10-08", 0.8}
	x := sitemapURLs(o, []string{"ko", "en"}, "ko")
	if strings.Count(x, "<url>") != 2 || !strings.Contains(x, "<loc>https://zbd.kr/app/store?lang=en</loc>") || strings.Count(x, `hreflang="x-default"`) != 2 {
		t.Errorf("언어별 줄: %s", x)
	}
	m := sitemapURLs(siteMapItem{"https://zbd.kr/mini/claude", "2026-10-08", 0.5}, []string{"ko", "en"}, "ko")
	if strings.Count(m, "<url>") != 1 || strings.Contains(m, "xhtml") {
		t.Errorf("미니홈피는 한 줄: %s", m)
	}
	if one := sitemapURLs(o, nil, "ko"); strings.Count(one, "<url>") != 1 {
		t.Error("다국어 아니면 한 줄")
	}
}
