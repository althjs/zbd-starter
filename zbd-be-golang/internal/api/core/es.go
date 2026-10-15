package core

// 검색엔진(ES) 글 색인 — 글을 고치거나 지울 때 색인도 맞춘다. 실패해도 막지 않는다(지금 서비스와 같다).
//   주소: 설정 ES_HOST(기본 http://localhost:9202) · 계정 elastic / ES_PASSWORD

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

const esPostIndex = "posts"

func esDo(c *web.Ctx, method, path string, body any) (int, []byte, error) {
	host := strings.TrimRight(c.App.Cfg.GetOr("ES_HOST", "http://localhost:9202"), "/")
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(jsx.Marshal(body))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, host+path, rd)
	if err != nil {
		return 0, nil, err
	}
	req.SetBasicAuth("elastic", c.App.Cfg.Get("ES_PASSWORD"))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	return res.StatusCode, b, nil
}

// esPutPost — 색인 고치기(없으면 새로 — 그때 createdAt 은 지금)
func esPutPost(c *web.Ctx, post *jsx.OrderedMap) {
	p := jsx.Assign(nil, post)
	for _, k := range []string{"json", "jsonUser"} {
		if s, ok := p.V(k).(string); ok {
			if s == "" {
				s = "{}"
			}
			p.Set(k, jsx.ParseOr(s, jsx.NewOrderedMap()))
		}
	}
	id := jsx.Str(p.V("postId"))
	st, _, err := esDo(c, "POST", "/"+esPostIndex+"/_update/"+id, jsx.Obj("doc", p))
	if err == nil && st < 300 {
		return
	}
	p.Set("createdAt", nowMs())
	if st, b, err := esDo(c, "PUT", "/"+esPostIndex+"/_create/"+id, p); err != nil || st >= 300 {
		log.Printf("[es] 글 색인 실패 %s %d %v %s", id, st, err, clip(string(b), 200))
	}
}

// esDeletePost — 색인 지우기(없으면 그만)
func esDeletePost(c *web.Ctx, postID string) {
	if _, _, err := esDo(c, "DELETE", "/"+esPostIndex+"/_doc/"+postID, nil); err != nil {
		log.Printf("[es] 글 색인 삭제 실패 %s %v", postID, err)
	}
}
