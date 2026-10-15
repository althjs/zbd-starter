package apps

// 작은 조회·설정 라우트 — 친구 채팅 로그인 넘겨주기 · 내 댓글 · 내 설정 · 내 기기

import (
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/web"
)

// GET·PUT /user/my/settings?settingKey= — 내 설정 한 칸(회원 json.settings[settingKey])
func mySettings(c *web.Ctx) {
	key := c.Q("settingKey")
	if !c.Token.Valid() || c.Token.MemberType != "U" || key == "" {
		c.Ecb("잘못된 요청")
		return
	}
	ctx := c.Context()
	m, err := member.Get(ctx, c.App, c.App.DB, c.Token.Key)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	if m == nil {
		c.Ecb("TypeError: Cannot read properties of undefined (reading 'json')")
		return
	}
	js := jsx.ParseMap(m.S("json"))
	if js == nil {
		js = jsx.NewOrderedMap()
	}
	st := js.Map("settings")
	if st == nil {
		st = jsx.NewOrderedMap()
	}
	switch c.Method {
	case "GET":
		v := st.V(key)
		if !jsx.Truthy(v) {
			v = nil
		}
		c.Scb(v)
	case "PUT":
		bd := c.Body()
		if !jsx.Truthy(bd) {
			c.Ecb("잘못된 요청 bd")
			return
		}
		// 객체 설정(라이프·라벨·데스크탑)은 객체만 받는다 — 화면이 「"OK"」 응답을 설정으로 쓰다 {"0":"O","1":"K",…} 로 통째로 덮은 일이 있다(2026-10-03)
		if key == "LIFE_LOG" || key == "LABEL" || key == "LIFE_DESKTOP" {
			o, ok := bd.(*jsx.OrderedMap)
			if !ok || (o.Has("0") && o.Has("1") && o.V("0") == "O" && o.V("1") == "K") {
				c.Ecb("잘못된 설정 모양입니다. 화면을 새로고침한 뒤 다시 저장해 주세요.")
				return
			}
		}
		st.Set(key, bd)
		js.Set("settings", st)
		m.Set("json", jsx.Stringify(js))
		o, err := member.Update(ctx, c.App, c.App.DB, m)
		if err != nil {
			c.Ecb(errMsg(err))
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("업데이트 오류")
			return
		}
		c.Scb("OK")
	default:
		c.Ecb("잘못된 요청")
	}
}

// GET /user/my/device — 내 로그인 기기
func myDevices(c *web.Ctx) {
	cid := c.Token.HostingID
	if cid == "" {
		cid = c.Q("hostingId")
	}
	if c.Method != "GET" {
		c.Ecb("잘못된 요청")
		return
	}
	email, mt := c.Token.Email, c.Token.MemberType
	if email == "" || mt == "" {
		c.Ecb("잘못된 요청")
		return
	}
	rows, err := db.Query(c.Context(), c.App.DB, `select browserUid, 
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt, 
      UNIX_TIMESTAMP(lastLoginedAt)*1000 as lastLoginedAt
    FROM login_devices WHERE userId = ? AND memberType = ? AND email = ? AND hostingId = ?`,
		c.Token.Key, mt, c.App.PII.Hash("email", email), cid)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	c.Scb(rowsAny(rows))
}

// DELETE /user/my/device/[browserUid] — 내 로그인 기기 하나 지우기
func myDeviceDelete(c *web.Ctx) {
	if c.Method != "DELETE" {
		c.Ecb("잘못된 요청")
		return
	}
	email, mt, uid := c.Token.Email, c.Token.MemberType, c.Q("browserUid")
	if email == "" || mt == "" || uid == "" {
		c.Ecb("잘못된 요청")
		return
	}
	o, err := db.Exec(c.Context(), c.App.DB, `DELETE FROM login_devices 
  WHERE memberType = ? AND email = ? AND browserUid = ?`, mt, c.App.PII.Hash("email", email), uid)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	c.Scb(o)
}
