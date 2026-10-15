package core

import (
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// 로그아웃 · 세션 옮기기 — 비밀번호 찾기(이메일·문자 인증)는 메일 설정이 필요해 스타터에 없다

// loginLogout — GET: 이 브라우저의 로그인 기기 기록을 지운다
func loginLogout(c *web.Ctx) {
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	uid := c.BrowserUID()
	if uid == "" {
		uid = "UNKNOWN"
	}
	mt := c.Token.MemberType
	if mt == "" {
		mt = "X"
	}
	var email any
	if c.Token.Valid() {
		email = c.Token.Email
	}
	c.ClearSession()
	o, err := db.Exec(c.Context(), c.App.DB, `DELETE FROM login_devices WHERE memberType = ? AND email = ? AND browserUid = ?`, mt, c.App.PII.Hash("email", email), uid)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.Scb(o)
}

// loginSession — POST: 화면에 남아 있던 옛 토큰(localStorage, 헤더로 보냄)을 세션 쿠키로 옮긴다(2026-10 옮기는 기간 — 지울 때 이 표시를 찾는다).
// 회원 토큰은 이 브라우저 것이어야 하고, 토큰의 쪽(관리자/회원)이 주소와 맞아야 한다. 갱신 시각을 지금으로 바꿔 심는다.
func loginSession(c *web.Ctx) {
	t := c.Token
	if c.Method != "POST" || !t.Valid() || c.FromSessionCookie() {
		c.Ecb("잘못된 요청")
		return
	}
	if t.MemberType == "U" && t.BrowserUID != c.BrowserUID() {
		c.Ecb("브라우저 정보가 맞지 않습니다.")
		return
	}
	if !c.IssueSession(t.Refreshed(time.Now().UnixMilli(), c.App.Cfg.AESKey, c.App.Cfg.AESIV)) {
		c.Ecb("이 주소에서는 쓸 수 없습니다.")
		return
	}
	c.Scb(jsx.NewOrderedMap())
}

// rowCount — 인증 코드 표 UPDATE 의 「맞은 행 / 바뀐 행」 (드라이버가 바뀐 행 수를 주지 않아 미리 센다)
type rowCount struct{ matched, changed float64 }
