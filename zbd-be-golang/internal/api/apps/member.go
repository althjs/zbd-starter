package apps

// 회원 — 가입 · 탈퇴 · 휴대폰 인증 · 내 정보 · 개인 액세스 토큰(MCP 등)

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

func resolveJoinGroupID(bd *jsx.OrderedMap) float64 {
	t := strings.ToUpper(jsx.StrOrEmpty(bd.V("joinType")))
	if t == "COMPANY" {
		return 2
	}
	if t == "INDIVIDUAL" {
		return 1
	}
	if g := jsx.Number(bd.V("groupId")); g == 1 || g == 2 {
		return g
	}
	return 0
}

// addUser — 회원 넣기(이름·이메일·전화 암호화 + 찾기 값, 비밀번호 bcrypt) + 가입 알림(알림톡) · 환영 메일
func addUser(ctx context.Context, c *web.Ctx, q dbConn, m *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	if _, ok := m.V("password").(string); !ok {
		return nil, errors.New("data and salt arguments required")
	}
	pw, err := util.BcryptHash(m.S("password"))
	if err != nil {
		return nil, err
	}
	p := c.App.PII
	enc := func(k string) (string, error) { return p.Enc(m.V(k)) }
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
	g := m.V("groupId")
	if !jsx.Truthy(g) {
		g = float64(0)
	}
	ok, err := db.Exec(ctx, q, `INSERT INTO members (groupId, name, email, password, mobile, json, nickname, image, description, socialId, hostingId,
    name_hash, email_hash, mobile_hash)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jsx.Param(g), name, email, pw, mobile, m.V("json"), m.V("nickname"), orEmptyAny(m.V("image")), orEmptyAny(m.V("description")),
		orEmptyAny(m.V("socialId")), orEmptyAny(m.V("hostingId")), p.Hash("name", m.V("name")), p.Hash("email", m.V("email")), p.Hash("mobile", m.V("mobile")))
	if err != nil {
		return nil, err
	}
	m.Set("userId", ok.V("insertId"))
	return ok, nil
}

var emailLike = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// POST /user/member/join — 회원 가입(로그인한 소셜 회원이면 부가정보만 고친다). 가입 등급은 서버가 정한다
func memberJoin(c *web.Ctx) {
	bd := c.BodyMap() // 본문이 없으면 빈 객체 — 지금 서비스처럼 아래 검사에서 「잘못된 요청」 으로
	if looseEq(bd.V("memberType"), "U") && !jsx.Truthy(bd.V("hostingId")) {
		c.Ecb("잘못된 요청")
		return
	}
	ctx := c.Context()
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	defer conn.Close()
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
		c.Ecb("잘못된 요청")
		return
	}
	cid := jsx.StrOrEmpty(bd.V("hostingId"))
	site := c.App.Settings.Site(cid)
	if site == nil {
		c.Ecb("사이트 메타 정보가 없습니다.")
		return
	}
	if tokenKey(c) != "" {
		u, err := member.Get(ctx, c.App, conn, c.Token.Key)
		if err != nil {
			c.Ecb(errMsg(err))
			return
		}
		if u == nil {
			c.Ecb("Cannot set properties of null (setting 'json')")
			return
		}
		uj := parseJSONObj(u.V("json"))
		bj := jsx.ParseMap(jsx.StrOrEmpty(bd.V("json")))
		if !jsx.Truthy(u.V("name")) && jsx.Truthy(bd.V("name")) {
			u.Set("name", bd.V("name"))
		}
		pick := func(k string) any {
			if v := bd.V(k); jsx.Truthy(v) {
				return v
			}
			if v := u.V(k); jsx.Truthy(v) {
				return v
			}
			return ""
		}
		u.Set("mobile", pick("mobile"))
		u.Set("nickname", pick("nickname"))
		u.Set("description", pick("description"))
		u.Set("json", jsx.Stringify(jsx.Assign(jsx.Assign(nil, uj), bj)))
		o, err := member.Update(ctx, c.App, conn, u)
		if err != nil {
			c.Ecb(errMsg(err))
			return
		}
		log.Printf("소셜회원 가입 회원 부가정보 업데이트 결과 %v %v", u.V("userId"), o.V("affectedRows"))
		c.Scb(u)
		return
	}
	if jsx.Truthy(site.Path("feature", "useCustomAccountIdExceptEmail")) {
		j := jsx.ParseMap(jsx.StrOrEmpty(bd.V("json")))
		if j == nil {
			c.Ecb("Unexpected end of JSON input")
			return
		}
		j.Set("email", bd.V("email"))
		bd.Set("email", bd.V("accountId"))
		bd.Set("json", jsx.Stringify(j))
	}
	bd.Set("groupId", resolveJoinGroupID(bd))
	o, err := addUser(ctx, c, conn, bd)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	if db.Affected(o) == 0 {
		c.Ecb(o)
		return
	}
	mm, err := member.Get(ctx, c.App, conn, o.V("insertId"))
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	c.Scb(mapOrNil(mm))
}

// DELETE /user/member/leave — 탈퇴(status 99)
func memberLeave(c *web.Ctx) {
	if tokenKey(c) == "" || c.Method != "DELETE" {
		c.Ecb("잘못된 요청")
		return
	}
	o, err := db.Exec(c.Context(), c.App.DB, `UPDATE members SET status = 99, updatedAt = current_timestamp()
  WHERE hostingId = ? AND userId = ?`, c.Token.HostingID, c.Token.Key)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	if db.Affected(o) == 0 {
		c.Ecb("처리 도중 오류가 발생하였습니다.")
		return
	}
	c.Scb("탈퇴가 완료되었습니다.")
}

// GET·PUT /user/my/account — 내 정보 보기·고치기(고치는 대상은 로그인한 본인, 등급·상태는 그대로)
func myAccount(c *web.Ctx) {
	uid := c.Token.Key
	cid := c.Token.HostingID
	if cid == "" {
		cid = jsx.StrOrEmpty(c.QV("hostingId"))
	}
	ctx := c.Context()
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(errMsg(err))
		return
	}
	defer conn.Close()
	var m *jsx.OrderedMap
	if uid != "" {
		if m, err = member.Get(ctx, c.App, conn, uid); err != nil {
			c.Ecb(errMsg(err))
			return
		}
	}
	switch c.Method {
	case "GET":
		c.Scb(mapOrNil(m))
	case "PUT":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("email")) || !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		// 비밀번호 바꾸기 — 로그인한 본인 행의 기존 비밀번호가 맞아야 한다(예전엔 passwordOld 를 빼고 보내면 확인 없이 바뀌었고,
		// 확인 대상도 본문 email 이었다 — 2026-10-06 E2E K8). 판정은 myPasswordChangeOK(정책 테스트)
		if newPw, oldPw := jsx.StrOrEmpty(bd.V("password")), jsx.StrOrEmpty(bd.V("passwordOld")); newPw != "" || oldPw != "" {
			stored := ""
			if uid != "" {
				r, err := db.QueryOne(ctx, conn, `select password from members WHERE userId = ? AND socialId = '' AND status >= 0 AND status <> 99`, uid)
				if err != nil {
					c.Ecb(errMsg(err))
					return
				}
				if r != nil {
					stored = r.S("password")
				}
			}
			if !myPasswordChangeOK(newPw, oldPw, stored, util.BcryptCompare) {
				ecbCode(c, "기존 패스워드가 일치하지 않습니다.", "0001")
				return
			}
		}
		var uidV any
		if uid != "" {
			uidV = uid
		}
		bd.Set("userId", uidV)
		if m != nil && m.V("hostingId") != nil {
			bd.Set("hostingId", m.V("hostingId"))
		} else {
			bd.Set("hostingId", cid)
		}
		bd.Set("groupId", nil)
		if m != nil && m.V("status") != nil {
			bd.Set("status", m.V("status"))
		} else {
			bd.Set("status", float64(0))
		}
		setDefaultProfileImage(c, bd)
		o, err := member.Update(ctx, c.App, conn, bd)
		if err != nil {
			c.Ecb(errMsg(err))
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("업데이트 오류")
			return
		}
		m2, err := member.Get(ctx, c.App, conn, uid)
		if err != nil {
			c.Ecb(errMsg(err))
			return
		}
		c.Scb(mapOrNil(m2))
	default:
		c.Ecb("잘못된 요청")
	}
}

// setDefaultProfileImage — 프로필 사진을 profile.webp 로 고정해 둔다(올린 사진을 복사하거나, 소셜 프로필 사진을 받아 둔다)
func setDefaultProfileImage(c *web.Ctx, bd *jsx.OrderedMap) {
	if bd.V("image") == "profile.webp" {
		return
	}
	dir := uploadDir(c.App) + tmpl(bd.V("hostingId")) + "/u_" + tmpl(bd.V("userId")) + "_my/"
	if !exists(dir) {
		mkdirAll(dir)
	}
	img := tmplU(bd.V("image"), bd.Has("image"))
	fixed := dir + "profile.webp"
	if jsx.Truthy(bd.V("image")) && exists(dir+img) {
		if b, err := os.ReadFile(dir + img); err == nil {
			_ = os.WriteFile(fixed, b, 0o644)
		}
		syncCDN(c.App, dir)
		return
	}
	var j *jsx.OrderedMap
	switch v := bd.V("json").(type) {
	case string:
		j = jsx.ParseMap(v)
	case *jsx.OrderedMap:
		j = v
	}
	url := j.Path("me", "picture")
	if !jsx.Truthy(url) {
		url = j.Path("me", "response", "profile_image")
	}
	if !jsx.Truthy(url) {
		url = j.Path("me", "properties", "profile_image")
	}
	if !jsx.Truthy(url) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", jsx.Str(url), nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("프로필 이미지 다운로드 실패: %v", err)
		return
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 20<<20))
	if err != nil || res.StatusCode >= 300 {
		return
	}
	if os.WriteFile(fixed, b, 0o644) == nil {
		bd.Set("image", "profile.webp")
		syncCDN(c.App, dir)
	}
}

// ─── 개인 액세스 토큰 ───

var apiTokenScopes = []string{"widget:read", "widget:write", "mini:read", "mini:write"}

// myPasswordChangeOK — 내 정보에서 비밀번호를 바꿀 수 있는지. 새 비밀번호든 기존 비밀번호든 보냈으면
// 본인 행에 비밀번호가 있고(소셜 가입은 없다) 기존 비밀번호가 그것과 맞아야 한다.
func myPasswordChangeOK(newPw, oldPw, stored string, compare func(plain, hash string) bool) bool {
	if newPw == "" && oldPw == "" {
		return true
	}
	return oldPw != "" && stored != "" && compare(oldPw, stored)
}
