package core

import (
	"context"
	"log"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/web"
)

// 로그인(이메일·비밀번호) — pages/api/login

// setLoginDevice — 이 브라우저로 로그인했다(이메일 칸은 찾기 값)
func setLoginDevice(c *web.Ctx, q db.Execer, userID any, memberType string, email any, browserUID, hostingID, ip string) error {
	if memberType == "" {
		memberType = "X"
	}
	_, err := db.Exec(context.Background(), q, `INSERT INTO login_devices (userId, memberType, email, browserUid, hostingId, ip, lastLoginedAt) VALUES (?, ?, ?, ?, ?, ?, current_timestamp())
  ON DUPLICATE KEY UPDATE 
  lastLoginedAt = current_timestamp(),
  ip = ?
  `, userID, memberType, c.App.PII.Hash("email", email), browserUID, hostingID, ip, ip)
	return err
}

// userWithPassword — 이메일 가입 회원 확인(소셜 가입은 빼고). 없으면 nil.
func userWithPassword(c *web.Ctx, q db.Querier, id, password any, hostingID string) (*jsx.OrderedMap, error) {
	oo, err := queryOne(c.Context(), q, `select userId, groupId, name, nickname, email, password from members WHERE email_hash = ? AND socialId = '' AND hostingId = ? AND status >= 0 AND status <> 99`,
		c.App.PII.Hash("email", id), hostingID)
	if err != nil || oo == nil {
		return nil, err
	}
	c.App.PII.DecMember(oo)
	ok, err := bcryptCompareAny(password, oo.V("password"))
	if err != nil || !ok {
		return nil, err
	}
	name := oo.V("nickname")
	if !jsx.Truthy(name) {
		name = oo.V("name")
	}
	return jsx.Obj("key", oo.V("userId"), "name", name, "nickname", orEmptyStr(oo.V("nickname")), "email", oo.V("email"),
		"memberType", "U", "hostingId", "", "groupId", oo.V("groupId")), nil
}

// checkRequiredUserInfo — 회사가 가입 때 받기로 한 정보(이름·닉네임·생년월일·주소·휴대폰 인증)가 빠졌는가
func checkRequiredUserInfo(c *web.Ctx, u *jsx.OrderedMap) bool {
	site := c.App.Settings.Site(jsx.StrOrEmpty(u.V("hostingId")))
	f := site.Map("feature")
	if !jsx.Truthy(f.V("userMemberJoinForceAdditionalInfo")) {
		return false
	}
	J, _ := u.V("Json").(*jsx.OrderedMap)
	j, _ := u.V("json").(*jsx.OrderedMap) // 문자열이면 비어 있다(지금 서비스와 같다)
	switch {
	case !jsx.Truthy(u.V("name")):
		return true
	case jsx.Truthy(f.V("userMemberNicknameRequired")) && !jsx.Truthy(u.V("nickname")):
		return true
	case jsx.Truthy(f.V("useMemberDemographic")) && !jsx.Truthy(J.V("birthDate")) && !jsx.Truthy(j.V("birthDate")):
		return true
	case jsx.Truthy(f.V("isRequiredMemberAddress")) && !jsx.Truthy(J.Path("address", "address1")) && !jsx.Truthy(j.Path("address", "address1")):
		return true
	case jsx.Truthy(f.V("useMemberJoinMobileVerify")) && !jsx.Truthy(j.V("me")) && !jsx.Truthy(J.V("me")) && !jsx.Truthy(J.V("isMobileVerified")):
		return true
	}
	return false
}

// hostingTokenIDs — 호스팅 설정에서 토큰에 넣을 사이트 관리자·하위 관리자 번호(즈비드엔 설정이 없어 늘 0)
func hostingTokenIDs(c *web.Ctx, cid string) (siteAdmin, sub string) {
	co := c.App.Settings.Hosting(cid)
	sa := co.V("siteAdminId")
	if !jsx.Truthy(sa) {
		sa = co.Path("server", "associationAdminId")
	}
	sb := co.V("subAdminId")
	z := func(v any) string {
		if !jsx.Truthy(v) {
			return "0"
		}
		return jsx.Str(v)
	}
	return z(sa), z(sb)
}

// login — POST {id, password, memberType: A|U, hostingId, code?(2단계 인증)}
func login(c *web.Ctx) {
	ctx := c.Context()
	bd := c.BodyMap()
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	browserUID := c.BrowserUID()
	if browserUID == "" {
		browserUID = "UNKNOWN"
	}
	if !jsx.Truthy(bd.V("id")) || !jsx.Truthy(bd.V("password")) {
		c.Ecb("잘못된 요청 (POST)")
		return
	}
	cid := jsx.StrOrEmpty(bd.V("hostingId"))
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	var o *jsx.OrderedMap
	isAdmin := jsx.Str(bd.V("memberType")) == "A"
	if isAdmin {
		o, err = adminWithPassword(c, conn, bd.V("id"), bd.V("password"), cid)
	} else {
		o, err = userWithPassword(c, conn, bd.V("id"), bd.V("password"), cid)
	}
	if err != nil {
		c.Ecb(err)
		return
	}
	if o == nil {
		c.Ecb("계정 정보를 확인하세요")
		return
	}
	partnerID := "0"
	if js := jsx.StrOrEmpty(o.V("json")); js != "" {
		pj, err := jsx.ParseString(js)
		if err != nil {
			c.Ecb(err)
			return
		}
		if m, ok := pj.(*jsx.OrderedMap); ok && jsx.Truthy(m.V("partnerId")) {
			partnerID = jsx.Str(m.V("partnerId"))
		}
	}
	ip := ipOf(reqHeaders(c))
	if err := setLoginDevice(c, conn, o.V("key"), jsx.Str(o.V("memberType")), o.V("email"), browserUID, cid, ip); err != nil {
		c.Ecb(err)
		return
	}
	extra := false
	if jsx.Str(o.V("memberType")) == "U" {
		u, err := member.Get(ctx, c.App, conn, o.V("key"))
		if err != nil {
			c.Ecb(err)
			return
		}
		if u == nil {
			c.Ecb("TypeError: Cannot read properties of undefined (reading 'json')")
			return
		}
		if s, ok := u.V("json").(string); ok {
			jv, err := jsx.ParseString(s)
			if err != nil {
				c.Ecb(err)
				return
			}
			u.Set("Json", jv)
		}
		extra = checkRequiredUserInfo(c, u)
	}
	sa, sub := hostingTokenIDs(c, cid)
	tok := makeToken(c, jsx.Str(o.V("key")), jsx.StrOrEmpty(o.V("email")), jsx.Str(o.V("memberType")), browserUID, cid,
		jsx.StrOrEmpty(o.V("name")), jsx.StrOrEmpty(o.V("groupId")), partnerID, sa, jsx.StrOrEmpty(o.V("nickname")), 0, sub)
	r := jsx.NewOrderedMap()
	if !c.IssueSession(tok) { // 세션 쿠키로 줬으면 본문에 토큰을 싣지 않는다
		r.Set("token", tok)
	}
	if extra {
		r.Set("isAdditionalInfoRequired", true)
	}
	log.Printf("@ 로그인 %s %s %s", jsx.Str(o.V("memberType")), cid, jsx.Str(o.V("key")))
	c.Scb(r)
}
