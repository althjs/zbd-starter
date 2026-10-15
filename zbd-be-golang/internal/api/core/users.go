package core

import (
	"log"
	"regexp"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/web"
)

// 회원(사용자) — pages/api/admin/user/constants 의 회원 함수들

const userCols = `userId, groupId, name
  , email
  , mobile
  , nickname
  , description
  , image
  , UNIX_TIMESTAMP(createdAt)*1000 as createdAt
  , UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt
  , status
  , hostingId
  , socialId
  `

// usersWithSearch — 회원 검색(이름·전화·이메일은 찾기 값으로)
func usersWithSearch(c *web.Ctx, q db.Querier, hostingID string, search *jsx.OrderedMap, withJSON bool) ([]*jsx.OrderedMap, error) {
	sqlText := `select ` + userCols + map[bool]string{true: ", json", false: ""}[withJSON] + `
  FROM members 
  WHERE TRUE`
	var args []any
	if hostingID != "" {
		sqlText += ` AND hostingId = ?`
		args = append(args, hostingID)
	}
	if jsx.Truthy(search.V("groupId")) {
		sqlText += ` AND groupId = ?`
		args = append(args, search.V("groupId"))
	}
	if kw := search.V("keyword"); jsx.Truthy(kw) {
		switch jsx.StrOrEmpty(search.V("type")) {
		case "name":
			sqlText += ` AND name_hash = ?`
			args = append(args, c.App.PII.Hash("name", kw))
		case "mobile":
			sqlText += ` AND mobile_hash = ?`
			args = append(args, c.App.PII.Hash("mobile", kw))
		case "email":
			sqlText += ` AND email_hash = ?`
			args = append(args, c.App.PII.Hash("email", kw))
		case "groupId":
			sqlText += ` AND groupId = ?`
			args = append(args, kw)
		case "status":
			sqlText += ` AND status = ?`
			args = append(args, kw)
		}
	}
	rows, err := query(c.Context(), q, sqlText, args...)
	if err != nil {
		return nil, err
	}
	return c.App.PII.DecMembers(rows), nil
}

// addUser — 회원 추가(이름·이메일·전화 암호화 + 찾기 값, 비밀번호 bcrypt). 가입 알림톡·환영 메일까지. member.userId 를 채운다.
func addUser(c *web.Ctx, q db.Execer, m *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	ctx := c.Context()
	pw, err := bcryptAny(m.V("password"))
	if err != nil {
		return nil, err
	}
	enc := func(k string) (string, error) { return c.App.PII.Enc(m.V(k)) }
	name, err := enc("name")
	if err != nil {
		return nil, err
	}
	email, err := enc("email")
	if err != nil {
		return nil, err
	}
	mobile, err := enc("mobile")
	if err != nil {
		return nil, err
	}
	group := m.V("groupId")
	if !jsx.Truthy(group) {
		group = float64(0)
	}
	o, err := db.Exec(ctx, q, `INSERT INTO members (groupId, name, email, password, mobile, json, nickname, image, description, socialId, hostingId,
    name_hash, email_hash, mobile_hash)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		group, name, email, pw, mobile, m.V("json"), m.V("nickname"), orEmptyStr(m.V("image")), orEmptyStr(m.V("description")),
		orEmptyStr(m.V("socialId")), orEmptyStr(m.V("hostingId")),
		c.App.PII.Hash("name", m.V("name")), c.App.PII.Hash("email", m.V("email")), c.App.PII.Hash("mobile", m.V("mobile")))
	if err != nil {
		return nil, err
	}
	m.Set("userId", o.V("insertId"))
	sendJoinSystemNoti(c, jsx.StrOrEmpty(m.V("hostingId")), m)
	return o, nil
}

// sendJoinSystemNoti — 가입 알림톡(시스템 알람 1번). 회사 설정에 카카오 발신 키가 있을 때만.
//
//	⚠ 알림톡 발송(외부 카카오 비즈메시지)은 아직 옮기지 않았다 — 즈비드는 발신 키·템플릿이 없어 지금 서비스도 보내지 않는다.
func sendJoinSystemNoti(c *web.Ctx, hostingID string, m *jsx.OrderedMap) {
	key := jsx.StrOrEmpty(c.App.Settings.Hosting(hostingID).Path("server", "nhn", "senderKey"))
	if key == "" {
		return
	}
	log.Printf("⚠ 가입 알림톡: %s 에 발신 키가 있지만 이 서버는 알림톡을 아직 보내지 않는다(U%s)", hostingID, jsx.Str(m.V("userId")))
}

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func userHosting(c *web.Ctx) string {
	if c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return c.Q("hostingId")
}

// adminUser — 회원 목록(GET ?userIds · 검색) · 추가(POST) · 삭제/되살리기(DELETE ?userId[&status=DO_NORMAL])
func adminUser(c *web.Ctx) {
	ctx := c.Context()
	cid := userHosting(c)
	switch c.Method {
	case "GET":
		withJSON := jsx.Truthy(c.QV("isIncludeJson"))
		var rows []*jsx.OrderedMap
		var err error
		if s := c.Q("userIds"); s != "" {
			var ids []any
			for _, x := range strings.Split(s, ",") {
				ids = append(ids, x)
			}
			rows, err = member.ByIDs(ctx, c.App, c.App.DB, cid, ids, withJSON)
		} else {
			search := jsx.NewOrderedMap()
			for _, k := range []string{"type", "keyword", "groupId", "status"} {
				var v any
				if jsx.Truthy(c.QV(k)) {
					v = c.QV(k)
				}
				search.Set(k, v)
			}
			rows, err = usersWithSearch(c, c.App.DB, cid, search, withJSON)
		}
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rowsAny(rows))
	case "POST":
		bd := c.BodyMap()
		bd.Set("hostingId", cid)
		j, err := jsx.ParseString(jsx.Str(undefStr(bd, "json")))
		if err != nil {
			c.Ecb("SyntaxError: \"undefined\" is not valid JSON")
			return
		}
		jm, ok := j.(*jsx.OrderedMap)
		if !ok || jm == nil {
			c.Ecb("TypeError: Cannot set properties of null (setting 'registrant')")
			return
		}
		jm.Set("registrant", c.Token.Key)
		bd.Set("json", jsx.Stringify(jm))
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청")
			return
		}
		o, err := addUser(c, c.App.DB, bd)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) > 0 {
			u, err := member.Get(ctx, c.App, c.App.DB, o.V("insertId"))
			if err != nil {
				c.Ecb(err)
				return
			}
			c.Scb(nilIfNull(u))
		} else {
			c.Ecb(o)
		}
	case "DELETE":
		uid := c.Q("userId")
		if uid == "" {
			c.Ecb("잘못된 요청")
			return
		}
		var o *jsx.OrderedMap
		var err error
		if c.Q("status") == "DO_NORMAL" {
			o, err = db.Exec(ctx, c.App.DB, `UPDATE members SET status = 0, updatedAt = current_timestamp()
  WHERE hostingId = ? AND userId = ?`, cid, uid)
		} else {
			sqlText := `UPDATE members SET status = -1, updatedAt = current_timestamp()
  WHERE hostingId = ? AND userId = ?`
			u, err2 := member.Get(ctx, c.App, c.App.DB, uid)
			if err2 != nil {
				c.Ecb(err2)
				return
			}
			if u != nil && jsx.Number(u.V("status")) == -1 {
				// 탈퇴완료된 회원을 다시 삭제하면 DB 에서 지운다
				sqlText = `DELETE FROM members WHERE hostingId = ? AND userId = ?`
			}
			o, err = db.Exec(ctx, c.App.DB, sqlText, cid, uid)
		}
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("처리할회원이 없습니다.")
			return
		}
		c.Scb("정상적으로 처리되었습니다.")
	default:
		c.Ecb("잘못된 요청")
	}
}

// undefStr — 본문 칸이 없으면 "undefined"(JSON.parse(undefined) 와 같은 실패를 내려고)
func undefStr(m *jsx.OrderedMap, k string) any {
	if !m.Has(k) || m.V(k) == nil {
		if m.Has(k) {
			return "null"
		}
		return "undefined"
	}
	return m.V(k)
}

// adminUserOne — 회원 한 명 GET · PUT(회원 정보 통째로 — 본문의 userId)
func adminUserOne(c *web.Ctx) {
	ctx := c.Context()
	uid := c.QV("userId")
	if !jsx.Truthy(uid) {
		c.Ecb("잘못된 요청")
		return
	}
	cid := userHosting(c)
	switch c.Method {
	case "GET":
		u, err := member.Get(ctx, c.App, c.App.DB, uid)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(nilIfNull(u))
	case "PUT":
		bd := c.BodyMap()
		bd.Set("hostingId", cid)
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청")
			return
		}
		if jsx.Truthy(bd.V("password")) {
			if _, ok := bd.V("password").(string); !ok {
				c.Ecb("data must be a string or Buffer and salt must either be a salt string or a number of rounds")
				return
			}
		}
		o, err := member.Update(ctx, c.App, c.App.DB, bd)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) > 0 {
			bd.Set("password", "")
			bd.Set("updatedAt", nowMs())
			c.Scb(bd)
		} else {
			c.Ecb(o)
		}
	default:
		c.Ecb("잘못된 요청")
	}
}
