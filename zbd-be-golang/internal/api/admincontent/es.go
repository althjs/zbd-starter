package admincontent

// 검색 색인(ES) — 게시판 「검색 색인」 기능(feature.useES)이 켜진 글만.
// 주소는 설정 ES_HOST 가 있을 때만 쓴다(비어 있으면 건너뛴다 — 시험 환경에서 운영 색인을 건드리지 않게).

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

func esRequest(c *web.Ctx, method, path string, body any) (int, error) {
	host := c.App.Cfg.Get("ES_HOST")
	if host == "" {
		log.Printf("@ ES_HOST 미설정 — 색인 건너뜀 %s %s", method, path)
		return 0, nil
	}
	var rd *bytes.Buffer
	if body != nil {
		rd = bytes.NewBufferString(jsx.Stringify(body))
	} else {
		rd = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, host+path, rd)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("elastic", c.App.Cfg.Get("ES_PASSWORD"))
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// esPutPost — 있으면 고치고 없으면 만든다(json·jsonUser 는 객체로)
func esPutPost(c *web.Ctx, post *jsx.OrderedMap) {
	doc := jsx.Assign(nil, post)
	for _, k := range []string{"json", "jsonUser"} {
		if s, ok := doc.V(k).(string); ok {
			doc.Set(k, jsx.ParseOr(orStrDef(s, "{}"), jsx.NewOrderedMap()))
		}
	}
	id := jsx.Str(doc.V("postId"))
	code, err := esRequest(c, "POST", "/posts/_update/"+id, jsx.Obj("doc", doc))
	if err == nil && code >= 200 && code < 300 {
		return
	}
	if code == 0 && err == nil {
		return
	}
	doc.Set("createdAt", nowMs())
	if _, err := esRequest(c, "PUT", "/posts/_create/"+id, doc); err != nil {
		log.Printf("ERROR ES 색인 %v", err)
	}
}

func esDelete(c *web.Ctx, index, id string) {
	if _, err := esRequest(c, "DELETE", "/"+index+"/_doc/"+id, nil); err != nil {
		log.Printf("@ 게시글 ES 삭제 실패: %v", err)
	}
}
