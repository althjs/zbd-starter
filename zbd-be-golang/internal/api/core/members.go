package core

import (
	"context"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// 관리자 회원 — pages/api/admin/member/*

const adminCols = `adminId, groupId, name, email, mobile, 
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt, 
      UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, 
      status, json, hostingId`

// adminGet — 관리자 한 명(없으면 nil). withOtp 면 otp(풀어서), withPassword 면 password 까지.
func adminGet(c *web.Ctx, q db.Querier, adminID any, withOtp, withPassword bool) (*jsx.OrderedMap, error) {
	cols := adminCols
	if withOtp {
		cols += " , otp"
	}
	if withPassword {
		cols += "\n      , password"
	}
	r, err := queryOne(c.Context(), q, `select `+cols+`
    from admins
    WHERE adminId = ?`, adminID)
	if err != nil || r == nil {
		return r, err
	}
	if withOtp {
		r.Set("otp", c.App.PII.Dec(r.V("otp")))
	}
	return r, nil
}

// adminMembersAll — 회사 관리자 목록(+ 그룹 이름). 지금 서비스의 60초 캐시는 두지 않는다(프로세스 하나 · 늘 최신).
func adminMembersAll(ctx context.Context, q db.Querier, hostingID string) ([]*jsx.OrderedMap, error) {
	sqlText := `select m.adminId, m.groupId, m.name, m.email, m.mobile, 
        UNIX_TIMESTAMP(m.createdAt)*1000 as createdAt, 
        UNIX_TIMESTAMP(m.updatedAt)*1000 as updatedAt, 
        m.status, m.hostingId, m.json, 
        g.name AS groupName 
      from admins as m
        join admin_roles as g on m.groupId = g.groupId WHERE m.status >= 0`
	var args []any
	if hostingID != "" {
		sqlText += ` AND m.hostingId = ?`
		args = append(args, hostingID)
	}
	return query(ctx, q, sqlText, args...)
}

// canManageAdmins — 관리자 회원 메뉴 권한(R 이상)이 있는 사람 — 메뉴 ↔ API 묶음 「W:/service/admin/member」 의 메뉴.
// 아니면 관리자 목록·한 사람 조회는 이름·사진·조직만(adminDirectoryRow) — 업무 보드 담당자·채팅 이름에는 그것으로 충분하다.
func canManageAdmins(c *web.Ctx) bool {
	return c.App.Admin.CanRead(c.Token.HostingID, c.Token.Key, "/service/admin/member")
}

// adminDirectoryRow — 관리자 한 줄에서 동료에게 보여도 되는 칸만(이메일·전화·권한 표시 등은 뺀다)
func adminDirectoryRow(m *jsx.OrderedMap) *jsx.OrderedMap {
	o := jsx.NewOrderedMap()
	for _, k := range []string{"adminId", "groupId", "name", "status", "hostingId", "groupName"} {
		if m.Has(k) {
			o.Set(k, m.V(k))
		}
	}
	var src *jsx.OrderedMap
	switch j := m.V("json").(type) {
	case string:
		src = jsx.ParseMap(j)
	case *jsx.OrderedMap:
		src = j
	}
	pub := jsx.NewOrderedMap()
	if src != nil {
		for _, k := range []string{"image", "nickname", "org", "isAI"} {
			if src.Has(k) {
				pub.Set(k, src.V(k))
			}
		}
	}
	o.Set("json", jsx.Stringify(pub))
	o.Set("email", "")
	o.Set("mobile", "")
	return o
}

// canSetAiFields — 「AI 계정」 칸(isAI·aiCharacter)을 바꿀 수 있는 관리자: 개발자 또는 중계 계정(MESSENGER_NOTIFY_ADMIN_IDS, 기본 352)
func canSetAiFields(c *web.Ctx, caller string) bool {
	id := jsx.NumOr(caller, 0)
	if id == 0 {
		return false
	}
	if m := c.App.Admin.Member(jsx.NumStr(id)); m != nil && jsx.Truthy(m.V("isDeveloper")) {
		return true
	}
	for _, s := range strings.Split(c.App.Cfg.GetOr("MESSENGER_NOTIFY_ADMIN_IDS", "352"), ",") {
		if jsx.Number(strings.TrimSpace(s)) == id && s != "" {
			return true
		}
	}
	return false
}

// keepAiFlag — 저장할 json 의 AI 칸을 이전 값으로 되돌린다(문자열이면 문자열로, 객체면 객체로)
func keepAiFlag(next, prev any) any { return keepFlags(next, prev, "isAI", "aiCharacter") }

// adminFlagKeys — 저장하는 관리자가 바꿀 수 없는 json 칸(정책 테스트 members_flags_policy_test.go, 2026-10-05)
//
//	isDeveloper   = 개발자만 · claudeManager(업무 채팅 「관리자별 허용 권한」 관리) = 개발자·최고 관리자
//	(전에는 화면에서만 가렸고 서버는 json 을 그대로 받아 누구나 isDeveloper 를 켤 수 있었다)
func adminFlagKeys(isDev, isSuper bool) []string {
	var keys []string
	if !isDev {
		keys = append(keys, "isDeveloper")
		if !isSuper {
			keys = append(keys, "claudeManager")
		}
	}
	return keys
}

// keepFlags — 저장할 json 의 keys 칸을 이전 값으로 되돌린다(없던 칸이면 뺀다)
func keepFlags(next, prev any, keys ...string) any {
	asObj := func(v any) *jsx.OrderedMap {
		if !jsx.Truthy(v) {
			return jsx.NewOrderedMap()
		}
		if s, ok := v.(string); ok {
			m, _ := jsx.ParseOr(s, nil).(*jsx.OrderedMap)
			if m == nil {
				return jsx.NewOrderedMap()
			}
			return m
		}
		if m, ok := v.(*jsx.OrderedMap); ok {
			return m
		}
		return jsx.NewOrderedMap()
	}
	n := jsx.Assign(nil, asObj(next))
	p := asObj(prev)
	for _, k := range keys {
		if p.Has(k) {
			n.Set(k, p.V(k))
		} else {
			n.Delete(k)
		}
	}
	if _, ok := next.(string); ok {
		return jsx.Stringify(n)
	}
	return n
}

// callerFlags — 저장하는 관리자가 개발자·최고 관리자인가(관리자 캐시). 최고 관리자 = isSuperAdmin 또는 그룹 1(화면 isSuperAdmin() 과 같게)
func callerFlags(c *web.Ctx) (isDev, isSuper bool) {
	if m := c.App.Admin.Member(c.Token.Key); m != nil {
		return jsx.Truthy(m.V("isDeveloper")), jsx.Truthy(m.V("isSuperAdmin")) || jsx.Number(m.V("groupId")) == 1
	}
	return false, false
}

// adminMember — GET 목록 · POST 추가 · DELETE(?adminId)
func adminMember(c *web.Ctx) {
	ctx := c.Context()
	switch c.Method {
	case "GET":
		ms, err := adminMembersAll(ctx, c.App.DB, c.Token.HostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		if !canManageAdmins(c) {
			for i, m := range ms {
				ms[i] = adminDirectoryRow(m)
			}
		}
		c.Scb(rowsAny(ms))
	case "POST":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청")
			return
		}
		if !canSetAiFields(c, c.Token.Key) {
			bd.Set("json", keepAiFlag(bd.V("json"), nil))
		}
		if keys := adminFlagKeys(callerFlags(c)); len(keys) > 0 {
			bd.Set("json", keepFlags(bd.V("json"), nil, keys...))
		}
		pw, err := bcryptAny(bd.V("password"))
		if err != nil {
			c.Ecb(err)
			return
		}
		group := bd.V("groupId")
		if !jsx.Truthy(group) {
			group = float64(0)
		}
		o, err := db.Exec(ctx, c.App.DB, `INSERT INTO admins (groupId, name, email, password, mobile, json, hostingId)
  VALUES (?,?,?,?,?,?,?)`, group, bd.V("name"), bd.V("email"), pw, bd.V("mobile"), bd.V("json"), orEmptyStr(bd.V("hostingId")))
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) > 0 {
			a, err := adminGet(c, c.App.DB, db.InsertID(o), false, false)
			if err != nil {
				c.Ecb(err)
				return
			}
			c.Scb(nilIfNull(a))
		} else {
			c.Ecb(o)
		}
	case "DELETE":
		id := c.QV("adminId")
		if !jsx.Truthy(id) {
			c.Ecb("잘못된 요청")
			return
		}
		sqlText := `DELETE FROM admins WHERE adminId = ?`
		args := []any{id}
		if c.Token.HostingID != "" {
			sqlText += ` AND hostingId = ?`
			args = append(args, c.Token.HostingID)
		}
		o, err := db.Exec(ctx, c.App.DB, sqlText, args...)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) > 0 {
			c.Scb("DELETED")
		} else {
			c.Ecb(o)
		}
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminUpdate — 관리자 정보 고치기(updateAdmin). 비밀번호는 있을 때만, otp 는 있으면 암호화해 따로.
func adminUpdate(c *web.Ctx, q db.Execer, bd *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	ctx := c.Context()
	var group, pw, cid any
	if jsx.Truthy(bd.V("groupId")) {
		group = bd.V("groupId")
	}
	if jsx.Truthy(bd.V("password")) {
		h, err := util.BcryptHash(jsx.Str(bd.V("password")))
		if err != nil {
			return nil, err
		}
		pw = h
	}
	if jsx.Truthy(bd.V("hostingId")) {
		cid = bd.V("hostingId")
	}
	o, err := db.Exec(ctx, q, `UPDATE admins m SET 
  groupId = COALESCE(NULLIF(?, 0), m.groupId),
  name = ?,
  email = ?, 
  password = COALESCE(NULLIF(?, ''), m.password),
  mobile = ?, json = ?, 
  hostingId = COALESCE(NULLIF(?, ''), m.hostingId),
  updatedAt = current_timestamp() WHERE adminId = ?`, group, bd.V("name"), bd.V("email"), pw, bd.V("mobile"), bd.V("json"), cid, bd.V("adminId"))
	if err != nil {
		return nil, err
	}
	if jsx.Truthy(bd.V("otp")) {
		enc, err := c.App.PII.Enc(bd.V("otp"))
		if err != nil {
			return nil, err
		}
		o, err = db.Exec(ctx, q, `UPDATE admins SET
    otp = ? WHERE adminId = ?`, enc, bd.V("adminId"))
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

// adminMemberOne — GET · PUT /admin/member/[adminId]. 고칠 대상은 주소의 번호로만.
func adminMemberOne(c *web.Ctx) {
	adminID := c.QV("adminId")
	if !jsx.Truthy(adminID) {
		c.Ecb("잘못된 요청")
		return
	}
	switch c.Method {
	case "GET":
		a, err := adminGet(c, c.App.DB, adminID, false, false)
		if err != nil {
			c.Ecb(err)
			return
		}
		if a != nil && jsx.Str(adminID) != c.Token.Key && !canManageAdmins(c) {
			a = adminDirectoryRow(a)
		}
		c.Scb(nilIfNull(a))
	case "PUT":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		bd.Set("adminId", jsx.Number(adminID))
		flagKeys := adminFlagKeys(callerFlags(c))
		if !canSetAiFields(c, c.Token.Key) || len(flagKeys) > 0 {
			prev, err := adminGet(c, c.App.DB, adminID, false, false)
			if err != nil {
				c.Ecb(err)
				return
			}
			var pj any
			if prev != nil {
				pj = prev.V("json")
			}
			if !canSetAiFields(c, c.Token.Key) {
				bd.Set("json", keepAiFlag(bd.V("json"), pj))
			}
			if len(flagKeys) > 0 {
				bd.Set("json", keepFlags(bd.V("json"), pj, flagKeys...))
			}
		}
		o, err := adminUpdate(c, c.App.DB, bd)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("업데이트 오류")
			return
		}
		a, err := adminGet(c, c.App.DB, adminID, false, false)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(nilIfNull(a))
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminMemberByGroup — 그룹의 관리자들
func adminMemberByGroup(c *web.Ctx) {
	gid := c.QV("groupId")
	if gid == nil {
		gid = ""
	}
	ms, err := adminMembersAll(c.Context(), c.App.DB, c.Token.HostingID)
	if err != nil {
		c.Ecb(err)
		return
	}
	out := []any{}
	for _, m := range ms {
		if looseEq(m.V("groupId"), gid) {
			out = append(out, m)
		}
	}
	c.Scb(out)
}
