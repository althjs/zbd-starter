package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/lib/adminauth"
	"zbd.kr/be/internal/lib/auth"
	"zbd.kr/be/internal/lib/config"
	"zbd.kr/be/internal/lib/settings"
)

// 정책: 설정·권한·디자인을 바꾸는 라우트(Options.Changes)는 GET 이 아닌 요청이 성공했을 때만 「바뀜」 을 알린다(캐시 즉시 갱신 + 소켓, 2026-10-04)
func TestChangeNotifiesPolicy(t *testing.T) {
	ok := &Ctx{status: 200, Token: &auth.Token{Key: "1"}}
	cases := []struct {
		name   string
		o      Options
		method string
		c      *Ctx
		want   bool
	}{
		{"표시 + PUT 성공", Options{Changes: true}, "PUT", ok, true},
		{"표시 없으면 안 알림", Options{}, "PUT", ok, false},
		{"GET 은 안 알림", Options{Changes: true}, "GET", ok, false},
		{"실패 응답(Ecb)", Options{Changes: true}, "POST", &Ctx{status: 200, failed: true, Token: &auth.Token{Key: "1"}}, false},
		{"상태 400 이상", Options{Changes: true}, "DELETE", &Ctx{status: 403, Token: &auth.Token{Key: "1"}}, false},
		{"로그인 안 함", Options{Changes: true}, "PUT", &Ctx{status: 200, Token: &auth.Token{}}, false},
	}
	for _, c := range cases {
		if got := changeNotifies(c.o, c.method, c.c); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

// 라우터를 실제로 돌려 본다 — 성공한 PUT 은 알리고, Ecb 로 실패한 PUT · GET 은 알리지 않는다
func TestChangesRouteEndToEnd(t *testing.T) {
	key, iv := strings.Repeat("k", 32), strings.Repeat("v", 16)
	a := &app.App{Cfg: &config.Config{AESKey: key, AESIV: iv}, Settings: settings.New(), Admin: adminauth.New()}
	var got []string
	old := onChanged
	onChanged = func(_ *app.App, _ context.Context, h string) { got = append(got, h) }
	defer func() { onChanged = old }()

	r := NewRouter(a)
	r.Handle("/t/ok", Cors, func(c *Ctx) { c.Scb("ok") }, Options{Changes: true})
	r.Handle("/t/fail", Cors, func(c *Ctx) { c.Ecb("안 됨") }, Options{Changes: true})
	r.Handle("/t/plain", Cors, func(c *Ctx) { c.Scb("ok") })
	now := time.Now().UnixMilli()
	tok := (&auth.Token{Key: "352", MemberType: "A", BrowserUID: auth.AnyBrowser, HostingID: "ZBD", TS: now}).Refreshed(now, key, iv)

	do := func(method, path string) {
		req := httptest.NewRequest(method, "/service"+path, strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s: HTTP %d %s", method, path, w.Code, w.Body.String())
		}
	}
	do("PUT", "/t/ok")
	do("GET", "/t/ok")
	do("PUT", "/t/fail")
	do("PUT", "/t/plain")
	if len(got) != 1 || got[0] != "ZBD" {
		t.Fatalf("알림 = %v, 기대 [ZBD] 한 번(성공한 PUT 만)", got)
	}
}
