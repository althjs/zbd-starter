package site

// GET /seo/parse-open-graph?url= — 남의 페이지에서 제목·설명·대표 그림(base64)을 뽑는다(로그인 없이). 원본: pages/api/seo/parse-open-graph.ts
//
// 🔥 보안(2026-10-02 사고 — 지금 서비스): 셸을 거친 curl 로 명령이 실행됐고, 127.0.0.1 로 서버 내부 페이지를 받아 갔다.
//   여기서는 셸을 아예 쓰지 않고, http/https 공인 주소만 · 넘겨주기는 단계마다 다시 검사 · 실제 접속도 다이얼러에서 한 번 더 막는다.
//   실패 사유(내부 오류)는 응답에 싣지 않는다.

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/korean"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

const ogUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

// 받는 크기 상한 — 지금 서비스(curl 10MB)와 같게
const ogMaxBody = 10 << 20

func isPrivateIP(ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		a, b := v4[0], v4[1]
		return a == 0 || a == 10 || a == 127 || (a == 100 && b >= 64 && b <= 127) || (a == 169 && b == 254) ||
			(a == 172 && b >= 16 && b <= 31) || (a == 192 && b == 168) || a >= 224
	}
	x := strings.ToLower(ip.String())
	return x == "::" || x == "::1" || strings.HasPrefix(x, "fc") || strings.HasPrefix(x, "fd") || strings.HasPrefix(x, "fe80")
}

// assertPublicURL — 공인 http(s) 주소면 URL, 아니면 사유
func assertPublicURL(ctx context.Context, raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil, errors.New("잘못된 주소")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("http/https 주소만 허용")
	}
	host := u.Hostname()
	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else if addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host); err == nil {
		for _, a := range addrs {
			ips = append(ips, a.IP)
		}
	}
	if len(ips) == 0 {
		return nil, errors.New("주소를 찾을 수 없음")
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return nil, errors.New("내부 주소는 허용하지 않음")
		}
	}
	return u, nil
}

// ogClient — 넘겨주기를 따라가지 않고(직접 검사하며 따라간다), 접속 순간에도 내부 주소면 끊는다(DNS 바꿔치기 대비)
var ogClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
			Control: func(_, address string, _ syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				if ip := net.ParseIP(host); ip == nil || isPrivateIP(ip) {
					return errors.New("내부 주소는 허용하지 않음")
				}
				return nil
			},
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

func ogGet(ctx context.Context, u string, accept bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ogUA)
	if accept {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "ko-KR,ko;q=0.8,en-US;q=0.5,en;q=0.3")
	}
	return ogClient.Do(req)
}

// safeFetch — 넘겨주기를 직접 따라가며(최대 5번) 단계마다 공인 주소인지 다시 본다
func safeFetch(ctx context.Context, raw string, accept bool) (*http.Response, error) {
	cur := raw
	for i := 0; i <= 5; i++ {
		u, err := assertPublicURL(ctx, cur)
		if err != nil {
			return nil, err
		}
		res, err := ogGet(ctx, u.String(), accept)
		if err != nil {
			return nil, err
		}
		if loc := res.Header.Get("Location"); res.StatusCode >= 300 && res.StatusCode < 400 && loc != "" {
			res.Body.Close()
			next, err := u.Parse(loc)
			if err != nil {
				return nil, err
			}
			cur = next.String()
			continue
		}
		return res, nil
	}
	return nil, errors.New("리다이렉트가 너무 많음")
}

var reEucKr = regexp.MustCompile(`(?i)charset\s*=\s*(euc-kr|euckr|ksc5601|ks_c_5601-1987)`)

// decodeBody — EUC-KR(KSC5601 포함)면 풀고, 아니면 UTF-8 그대로
func decodeBody(b []byte, contentType string) string {
	ct := strings.ToLower(contentType)
	euc := strings.Contains(ct, "euc-kr") || strings.Contains(ct, "euckr") || strings.Contains(ct, "ksc5601")
	if !euc {
		head := b
		if len(head) > 2048 {
			head = head[:2048]
		}
		euc = reEucKr.Match(head)
	}
	if euc {
		if d, err := korean.EUCKR.NewDecoder().Bytes(b); err == nil {
			return string(d)
		}
	}
	return strings.ToValidUTF8(string(b), "�")
}

func readLimited(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, ogMaxBody))
}

// ogInfo — 문서에서 og 값(cheerio 의 $('meta[...]').attr('content') 와 같게 — 첫 요소의 content)
type ogDoc struct {
	metas []map[string]string
	title strings.Builder
}

func parseOgDoc(s string) *ogDoc {
	d := &ogDoc{}
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return d
	}
	var walk func(n *html.Node, inTitle bool)
	walk = func(n *html.Node, inTitle bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "meta":
				m := map[string]string{}
				for _, at := range n.Attr {
					if _, ok := m[at.Key]; !ok {
						m[at.Key] = at.Val
					}
				}
				d.metas = append(d.metas, m)
			case "title":
				inTitle = true
			}
		}
		if n.Type == html.TextNode && inTitle {
			d.title.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inTitle)
		}
	}
	walk(root, false)
	return d
}

// attr — 첫 meta[key="val"] 의 content (없으면 ok=false)
func (d *ogDoc) attr(key, val string) (string, bool) {
	for _, m := range d.metas {
		if v, ok := m[key]; ok && v == val {
			c, ok := m["content"]
			return c, ok
		}
	}
	return "", false
}

// first — 앞에서부터 비지 않은 값
func (d *ogDoc) first(cands ...[2]string) string {
	for _, c := range cands {
		if v, ok := d.attr(c[0], c[1]); ok && v != "" {
			return v
		}
	}
	return ""
}

// resolveRef — 그림 주소를 절대 주소로
func resolveRef(base, ref string) string {
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	if strings.HasPrefix(ref, "//") {
		if strings.HasPrefix(base, "https://") {
			return "https:" + ref
		}
		return "http:" + ref
	}
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	r, err := b.Parse(ref)
	if err != nil {
		return ""
	}
	return r.String()
}

// imageToBase64 — 그림을 data: 주소로(실패면 "")
func imageToBase64(ctx context.Context, u string) string {
	res, err := safeFetch(ctx, u, false)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ""
	}
	b, err := readLimited(res.Body)
	if err != nil {
		return ""
	}
	ct := res.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(b)
}

func parseOpenGraph(c *web.Ctx) {
	raw := jsTrim(jsx.StrOrEmpty(c.QV("url")))
	if raw == "" {
		c.Ecb("잘못된 요청")
		return
	}
	ctx := c.Context()
	u, err := assertPublicURL(ctx, raw)
	if err != nil {
		c.Ecb(err.Error())
		return
	}

	page := ""
	if res, err := safeFetch(ctx, u.String(), true); err == nil {
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			if b, err := readLimited(res.Body); err == nil {
				page = decodeBody(b, res.Header.Get("Content-Type"))
			}
		}
		res.Body.Close()
	}
	if page == "" {
		// 한 번 더 — 넘겨주기를 따라가지 않고 상태와 무관하게 본문을 읽는다(지금 서비스의 curl 재시도와 같은 동작, 셸 없이)
		res, err := ogGet(ctx, u.String(), false)
		if err != nil {
			c.Ecb("페이지를 가져오지 못했습니다.")
			return
		}
		b, err := readLimited(res.Body)
		res.Body.Close()
		if err != nil {
			c.Ecb("페이지를 가져오지 못했습니다.")
			return
		}
		page = decodeBody(b, "")
	}

	d := parseOgDoc(page)
	title := d.first([2]string{"property", "og:title"}, [2]string{"name", "og:title"})
	if title == "" {
		title = d.title.String()
	}
	desc := d.first([2]string{"property", "og:description"}, [2]string{"name", "og:description"}, [2]string{"name", "description"})
	img := d.first([2]string{"property", "og:image"}, [2]string{"name", "og:image"})

	b64 := ""
	if img != "" {
		if abs := resolveRef(raw, img); abs != "" {
			b64 = imageToBase64(ctx, abs)
		}
	}
	c.Scb(jsx.Obj("url", raw, "title", jsTrim(title), "description", jsTrim(desc), "base64Image", b64))
}
