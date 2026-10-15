package site

import (
	"strings"
	"testing"
)

var testParts = docParts{
	lang:     "ko",
	head:     "<title>새 제목</title><meta property=\"og:title\" content=\"새\" />",
	data:     "<script>window.___zbd_data_model = {}</script>",
	bodyTop:  "<div id=\"top\"></div>",
	noscript: "<noscript>본문</noscript>",
}

// 지금 React 빌드 — 자리표 없이 고정 위치에, 원래 있던 제목·OG·icon 은 지운다
func TestRenderDocFixedPositions(t *testing.T) {
	index := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/logo.svg" />
    <title>옛 제목</title>
    <meta property="og:title" content="옛" />
    <meta
      name="description"
      content="옛 설명"
    />
    <link rel="manifest" href="/manifest.json" />
  </head>
  <body class="x">
    <div id="root"></div>
  </body>
</html>`
	out := renderDoc(index, testParts)
	for _, bad := range []string{"옛 제목", `content="옛"`, "옛 설명", "/logo.svg", `lang="en"`} {
		if strings.Contains(out, bad) {
			t.Fatalf("남으면 안 되는 것이 남았다: %s\n%s", bad, out)
		}
	}
	for _, want := range []string{`<html lang="ko">`, `<meta charset="UTF-8" />`, `/manifest.json`} {
		if !strings.Contains(out, want) {
			t.Fatalf("있어야 할 것이 없다: %s\n%s", want, out)
		}
	}
	head := out[:strings.Index(out, "</head>")]
	if !strings.Contains(head, testParts.head) || !strings.Contains(head, testParts.data) {
		t.Fatalf("head 묶음·데이터는 </head> 앞에\n%s", out)
	}
	if !strings.Contains(out, `<body class="x">`+"\n"+testParts.bodyTop) {
		t.Fatalf("본문 위 HTML 은 <body> 바로 뒤\n%s", out)
	}
	if !strings.Contains(out, testParts.noscript+"</body>") {
		t.Fatalf("noscript 는 </body> 앞\n%s", out)
	}
}

// 새 화면 빌드 — 자리표가 있으면 그 자리에
func TestRenderDocMarks(t *testing.T) {
	index := `<html><head><meta charset="utf-8"><!--zbd:head--><title>앱</title><!--zbd:data--></head>` +
		`<body><div id="app"></div><!--zbd:body--><script src="/assets/a.js"></script></body></html>`
	out := renderDoc(index, testParts)
	want := `<html lang="ko"><head><meta charset="utf-8">` + testParts.head + testParts.data + `</head>` +
		`<body><div id="app"></div>` + testParts.bodyTop + testParts.noscript + `<script src="/assets/a.js"></script></body></html>`
	if out != want {
		t.Fatalf("자리표 끼우기가 다르다\n got %s\nwant %s", out, want)
	}
}

func TestToHTMLLang(t *testing.T) {
	for in, want := range map[any]string{nil: "ko", "": "ko", "MICE": "ko", "en": "en", "zh-CN": "zh-CN", "ko ": "ko"} {
		if got := toHTMLLang(in); got != want {
			t.Fatalf("toHTMLLang(%v) = %s, want %s", in, got, want)
		}
	}
}

func TestIsIgnoredSeoPath(t *testing.T) {
	for p, want := range map[string]bool{"/wp-admin/x": true, "/WP": true, "/wpath": false, "/a/b.PHP": true, "/vendor/x": true, "/vendors": false, "/": false} {
		if isIgnoredSeoPath(p) != want {
			t.Fatalf("isIgnoredSeoPath(%s) != %v", p, want)
		}
	}
}

func TestPageNameOf(t *testing.T) {
	cases := map[string]string{
		"솔리테어 — 즈비드 스토어": "솔리테어",
		"공지 - 즈비드":       "공지",
		"즈비드":            "즈비드",
		"제목만":            "제목만",
	}
	for in, want := range cases {
		if got := pageNameOf(in, "즈비드"); got != want {
			t.Fatalf("pageNameOf(%s) = %s, want %s", in, got, want)
		}
	}
}
