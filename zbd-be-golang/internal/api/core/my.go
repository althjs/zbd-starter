package core

import (
	"sync"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// 내 계정(관리자) — pages/api/admin/my/*

// ecbCode — ecb(msg, code)
func ecbCode(c *web.Ctx, msg any, code string) {
	m := web.Ecb(msg)
	m.Set("code", code)
	c.Send(m)
}

// bcryptCompareAny — bcrypt.compare(v, hash): 값이 없으면 그 라이브러리와 같은 오류
func bcryptCompareAny(pw, hash any) (bool, error) {
	p, ok1 := pw.(string)
	h, ok2 := hash.(string)
	if pw == nil || hash == nil {
		return false, errFmt("data and hash arguments required")
	}
	if !ok1 || !ok2 {
		return false, errFmt("data must be a string or Buffer and hash must be strings")
	}
	return util.BcryptCompare(p, h), nil
}

// adminWithPassword — 이메일·비밀번호가 맞는 관리자(login 의 getAdminWithPassword). 없으면 nil.
func adminWithPassword(c *web.Ctx, q db.Querier, id, password any, hostingID string) (*jsx.OrderedMap, error) {
	oo, err := queryOne(c.Context(), q, `SELECT adminId, groupId, name, email, password, json, otp 
  FROM admins 
  WHERE email = ? AND hostingId = ? AND status >= 0`, id, hostingID)
	if err != nil || oo == nil {
		return nil, err
	}
	ok, err := bcryptCompareAny(password, oo.V("password"))
	if err != nil || !ok {
		return nil, err
	}
	return jsx.Obj(
		"key", oo.V("adminId"),
		"groupId", oo.V("groupId"),
		"name", oo.V("name"),
		"email", oo.V("email"),
		"memberType", "A",
		"json", oo.V("json"),
		"otp", c.App.PII.Dec(oo.V("otp")),
	), nil
}

func myAccountView(m *jsx.OrderedMap) *jsx.OrderedMap {
	return jsx.Obj("adminId", m.V("adminId"), "name", m.V("name"), "email", m.V("email"), "mobile", m.V("mobile"), "json", m.V("json"))
}

// adminMyAccount — 내 정보 GET · PUT(비밀번호를 바꾸면 기존 비밀번호 확인)
func adminMyAccount(c *web.Ctx) {
	adminID := c.Token.Key
	hostingID := metaHosting(c)
	switch c.Method {
	case "GET":
		m, err := adminGet(c, c.App.DB, adminID, false, false)
		if err != nil {
			c.Ecb(err)
			return
		}
		if m == nil {
			c.Ecb("Cannot read properties of null (reading 'adminId')")
			return
		}
		c.Scb(myAccountView(m))
	case "PUT":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		if jsx.Truthy(bd.V("passwordOld")) || jsx.Truthy(bd.V("password")) {
			m, err := adminWithPassword(c, c.App.DB, bd.V("email"), bd.V("passwordOld"), hostingID)
			if err != nil {
				c.Ecb(err)
				return
			}
			// 확인한 계정이 로그인한 본인이어야 한다 — 본문 email 을 다른 관리자 것으로 바꾸면 그 사람 비밀번호로 통과했다(2026-10-06)
			if m == nil || !sameAdmin(m.V("key"), adminID) {
				ecbCode(c, "기존 패스워드가 일치하지 않습니다.", "0001")
				return
			}
		}
		// 고칠 대상은 토큰의 본인만 — 지금 서비스는 본문의 adminId 를 그대로 썼다(남의 계정을 고칠 수 있던 구멍)
		bd.Set("adminId", jsx.Number(adminID))
		o, err := adminUpdate(c, c.App.DB, bd)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("업데이트 오류")
			return
		}
		m, err := adminGet(c, c.App.DB, adminID, false, false)
		if err != nil || m == nil {
			c.Ecb(err)
			return
		}
		c.Scb(myAccountView(m))
	default:
		c.Ecb("잘못된 요청")
	}
}

// 비밀번호 재확인 실패 횟수(5분에 10번까지) — 세션을 뺏긴 상태의 무차별 대입 막기
var (
	verifyFailMu sync.Mutex
	verifyFail   = map[int64]struct {
		count int
		first time.Time
	}{}
)

func verifyFailState(id int64) (int, time.Time, bool) {
	s, ok := verifyFail[id]
	if !ok {
		return 0, time.Time{}, false
	}
	if time.Since(s.first) > 5*time.Minute {
		delete(verifyFail, id)
		return 0, time.Time{}, false
	}
	return s.count, s.first, true
}

// adminMyVerifyPassword — POST {password}: 본인 비밀번호 재확인(일치 여부만)
func adminMyVerifyPassword(c *web.Ctx) {
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	adminID := c.Token.ID()
	if adminID == 0 || c.Token.MemberType != "A" {
		ecbCode(c, "로그인 정보가 없습니다.", "0401")
		return
	}
	pv := c.BodyMap().V("password")
	password := ""
	if pv != nil {
		password = jsx.Str(pv)
	}
	if password == "" {
		ecbCode(c, "비밀번호를 입력해주세요.", "0001")
		return
	}
	verifyFailMu.Lock()
	n, _, _ := verifyFailState(adminID)
	verifyFailMu.Unlock()
	if n >= 10 {
		ecbCode(c, "비밀번호 확인 시도가 너무 많습니다. 잠시 후 다시 시도해주세요.", "0429")
		return
	}
	m, err := queryOne(c.Context(), c.App.DB, `SELECT adminId, password FROM admins WHERE adminId = ? AND status >= 0`, adminID)
	if err != nil {
		c.Ecb(err)
		return
	}
	if m == nil || !jsx.Truthy(m.V("password")) || !util.BcryptCompare(password, jsx.Str(m.V("password"))) {
		verifyFailMu.Lock()
		cnt, first, ok := verifyFailState(adminID)
		if !ok {
			first = time.Now()
		}
		verifyFail[adminID] = struct {
			count int
			first time.Time
		}{cnt + 1, first}
		verifyFailMu.Unlock()
		ecbCode(c, "비밀번호가 일치하지 않습니다.", "0001")
		return
	}
	verifyFailMu.Lock()
	delete(verifyFail, adminID)
	verifyFailMu.Unlock()
	c.Scb(jsx.Obj("verified", true))
}

// sameAdmin — 비밀번호를 확인한 관리자 번호가 토큰의 관리자인지
func sameAdmin(key any, adminID string) bool {
	return adminID != "" && jsx.StrOrEmpty(key) == adminID
}
