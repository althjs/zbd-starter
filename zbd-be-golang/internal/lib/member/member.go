// Package member — 여러 기능이 같이 쓰는 회원 조회와 공개용 모양.
package member

import (
	"context"
	"strings"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
)

func hashPassword(p string) (string, error) { return util.BcryptHash(p) }

// StoreHostingID — 회사를 정할 단서가 없을 때(비로그인 공개 화면)
const StoreHostingID = "ZBD"

// ByIDs — 회원 여러 명(이름·이메일·전화는 풀어서). hostingId 가 ""면 회사 조건 없이. withJSON 이면 json 칸까지.
func ByIDs(ctx context.Context, a *app.App, q db.Querier, hostingID string, userIDs []any, withJSON bool) ([]*jsx.OrderedMap, error) {
	query := `select userId, groupId, name
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
  ` + map[bool]string{true: ", json", false: ""}[withJSON] + `
  FROM members 
  WHERE userId IN (?)`
	args := []any{userIDs}
	if hostingID != "" {
		query += ` AND hostingId = ?`
		args = append(args, hostingID)
	}
	rows, err := db.Query(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	return a.PII.DecMembers(rows), nil
}

// PublicCreator — 남에게 보여 줄 회원 모양 {userId, nickname, description, image}. 본명은 첫 글자 + * 로.
func PublicCreator(m *jsx.OrderedMap) *jsx.OrderedMap {
	if m == nil {
		return nil
	}
	name := strings.TrimSpace(jsx.StrOrEmpty(m.V("name")))
	masked := name
	if r := []rune(name); len(r) > 1 {
		masked = string(r[0]) + strings.Repeat("*", len(r)-1)
	}
	nick := jsx.StrOrEmpty(m.V("nickname"))
	if nick == "" {
		nick = masked
	}
	if nick == "" {
		nick = "사용자"
	}
	o := jsx.NewOrderedMap()
	o.Set("userId", m.V("userId"))
	o.Set("nickname", nick)
	o.Set("description", orEmpty(m.V("description")))
	o.Set("image", orEmpty(m.V("image")))
	return o
}

func orEmpty(v any) any {
	if !jsx.Truthy(v) {
		return ""
	}
	return v
}

const memberCols = `userId, groupId, name
  , email
  , mobile
  , nickname
  , description
  , image
  , UNIX_TIMESTAMP(createdAt)*1000 as createdAt
  , UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt
  , status, json, hostingId, socialId`

// Get — 회원 한 명(이름·이메일·전화는 풀어서, json 포함). 없으면 nil. (지금 서비스의 getUser)
func Get(ctx context.Context, a *app.App, q db.Querier, userID any) (*jsx.OrderedMap, error) {
	r, err := db.QueryOne(ctx, q, `select `+memberCols+`
  FROM members
  WHERE userId = ?`, userID)
	if err != nil || r == nil {
		return nil, err
	}
	return a.PII.DecMember(r), nil
}

// Update — 회원 정보 통째로(이름·이메일·전화는 암호화 + 찾기 값, 비밀번호가 있으면 bcrypt). (지금 서비스의 updateUser)
//
//	m 은 {userId, groupId, name, email, mobile, password, json, nickname, image, description, status}
func Update(ctx context.Context, a *app.App, q db.Execer, m *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	var pw any
	if p := jsx.StrOrEmpty(m.V("password")); p != "" {
		h, err := hashPassword(p)
		if err != nil {
			return nil, err
		}
		pw = h
	}
	var group any
	if jsx.Truthy(m.V("groupId")) {
		group = m.V("groupId")
	}
	enc := func(k string) (string, error) { return a.PII.Enc(m.V(k)) }
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
	status := m.V("status")
	if !jsx.Truthy(status) {
		status = float64(0)
	}
	return db.Exec(ctx, q, `UPDATE members m SET 
  groupId = COALESCE(NULLIF(?, 0), m.groupId),
  name = ?, name_hash = ?,
  email = ?, email_hash = ?,
  password = COALESCE(NULLIF(?, ''), m.password),
  mobile = ?, mobile_hash = ?,
  json = ?,
  nickname = ?,
  image = ?,
  description = ?,
  status = ?,
  updatedAt = current_timestamp() WHERE userId = ?`,
		group, name, a.PII.Hash("name", m.V("name")), email, a.PII.Hash("email", m.V("email")), pw,
		mobile, a.PII.Hash("mobile", m.V("mobile")), m.V("json"),
		jsx.StrOrEmpty(m.V("nickname")), jsx.StrOrEmpty(m.V("image")), jsx.StrOrEmpty(m.V("description")),
		status, m.V("userId"))
}
