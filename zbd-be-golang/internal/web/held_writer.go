package web

import (
	"bytes"
	"net/http"
)

// heldWriter — 설정을 바꾸는 라우트(Options.Changes)의 응답을 캐시 다시 읽기가 끝날 때까지 붙잡아 둔다.
// 바로 보내면 큰 응답(메뉴 목록 등)은 핸들러가 끝나기 전에 나가, 받은 쪽이 곧바로 부른 다음 요청이 옛 캐시를 봤다(E2E K9).
type heldWriter struct {
	w      http.ResponseWriter
	status int
	buf    bytes.Buffer
	used   bool
}

func (h *heldWriter) Header() http.Header { return h.w.Header() }

func (h *heldWriter) WriteHeader(s int) {
	if !h.used {
		h.status, h.used = s, true
	}
}

func (h *heldWriter) Write(b []byte) (int, error) {
	if !h.used {
		h.status, h.used = http.StatusOK, true
	}
	return h.buf.Write(b)
}

// release — 붙잡아 둔 응답을 실제로 보낸다(아무것도 안 썼으면 그대로 둔다)
func (h *heldWriter) release() {
	if !h.used {
		return
	}
	h.w.WriteHeader(h.status)
	_, _ = h.w.Write(h.buf.Bytes())
}
