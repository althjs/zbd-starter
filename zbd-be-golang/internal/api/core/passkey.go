package core

// 패스키(지문·얼굴) 로그인 — 관리자(2026-10-04 대표) · 회원(같은 날, /user/my/passkey · 사용자 주소의 /login/passkey)
//
//	등록  POST /admin/my/passkey {a:"options"} → 도전값·내 열쇠 목록 → 브라우저 navigator.credentials.create
//	      POST /admin/my/passkey {a:"register", id, clientDataJSON, authenticatorData, publicKey, publicKeyAlgorithm, transports, name}
//	목록  GET  /admin/my/passkey · 지우기 DELETE /admin/my/passkey?id=
//	로그인 POST /login/passkey {a:"options"} → 도전값 → navigator.credentials.get(열쇠 고르기는 기기가) →
//	      POST /login/passkey {a:"verify", id, clientDataJSON, authenticatorData, signature} → 세션(비밀번호 로그인과 같은 토큰)
//
// RP ID = 대표 도메인(zbd.kr) — backyard·chat·mail 이 한 패스키를 같이 쓴다(web.PasskeyRP). 본인 확인(UV) 필수라 2단계 인증을 대신한다.
// 표 passkeys(처음 쓸 때 생성, 되돌리기 DROP TABLE passkeys). 공개키만 저장 — 지문·개인키는 기기 밖으로 나오지 않는다.
// 도전값은 프로세스 메모리(3분, 한 번만) — Go BE 는 프로세스 하나.

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/member"
	"zbd.kr/be/internal/web"
)

const pkChallengeTTL = 3 * time.Minute

type pkChallenge struct {
	purpose string // reg | login
	adminID string // reg 만
	rpID    string
	until   time.Time
}

var pkChallenges = struct {
	sync.Mutex
	m map[string]pkChallenge
}{m: map[string]pkChallenge{}}

func pkNewChallenge(purpose, adminID, rpID string) string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	ch := pkB64Enc(b)
	now := time.Now()
	pkChallenges.Lock()
	for k, v := range pkChallenges.m {
		if now.After(v.until) {
			delete(pkChallenges.m, k)
		}
	}
	if len(pkChallenges.m) < 10000 {
		pkChallenges.m[ch] = pkChallenge{purpose, adminID, rpID, now.Add(pkChallengeTTL)}
	}
	pkChallenges.Unlock()
	return ch
}

// pkTakeChallenge — 한 번만 쓴다(꺼내면 지운다)
func pkTakeChallenge(ch, purpose, adminID, rpID string) bool {
	pkChallenges.Lock()
	defer pkChallenges.Unlock()
	v, ok := pkChallenges.m[ch]
	delete(pkChallenges.m, ch)
	return ok && time.Now().Before(v.until) && v.purpose == purpose && v.adminID == adminID && v.rpID == rpID
}

var (
	passkeyReadyMu sync.Mutex
	passkeyReady   bool
)

func ensurePasskeyTable(ctx context.Context, q db.Execer) error {
	passkeyReadyMu.Lock()
	defer passkeyReadyMu.Unlock()
	if passkeyReady {
		return nil
	}
	_, err := db.Exec(ctx, q, `CREATE TABLE IF NOT EXISTS passkeys (
       id INT UNSIGNED NOT NULL AUTO_INCREMENT,
       hostingId VARCHAR(50) NOT NULL DEFAULT '',
       memberType CHAR(1) NOT NULL COMMENT 'A 관리자 / U 회원',
       ownerId INT UNSIGNED NOT NULL,
       credId VARCHAR(255) NOT NULL COMMENT 'base64url',
       publicKey TEXT NOT NULL COMMENT 'SPKI DER base64url',
       alg INT NOT NULL DEFAULT 0,
       signCount INT UNSIGNED NOT NULL DEFAULT 0,
       rpId VARCHAR(100) NOT NULL,
       name VARCHAR(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
       transports VARCHAR(100) NOT NULL DEFAULT '',
       createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
       lastUsedAt DATETIME NULL,
       PRIMARY KEY (id),
       UNIQUE KEY uq_passkeys_cred (credId),
       KEY ix_passkeys_owner (memberType, ownerId)
     ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err == nil {
		passkeyReady = true
	}
	return err
}

func pkUserHandle(mt, hostingID, id string) string {
	return pkB64Enc([]byte(mt + hostingID + ":" + id))
}

// adminMyPasskey · userMyPasskey — 내 패스키 목록·등록·지우기(관리자는 관리자 주소에서, 회원은 사용자 주소에서만)
func adminMyPasskey(c *web.Ctx) { myPasskey(c, "A") }
func userMyPasskey(c *web.Ctx)  { myPasskey(c, "U") }

func myPasskey(c *web.Ctx, mt string) {
	ctx := c.Context()
	if !c.Token.Valid() || c.Token.MemberType != mt {
		c.Ecb("로그인이 필요합니다")
		return
	}
	adminID := c.Token.Key // 관리자면 adminId, 회원이면 userId
	if err := ensurePasskeyTable(ctx, c.App.DB); err != nil {
		c.Ecb(err)
		return
	}
	list := func() ([]*jsx.OrderedMap, error) {
		return query(ctx, c.App.DB, `SELECT id, name, rpId, createdAt, lastUsedAt FROM passkeys WHERE memberType = ? AND ownerId = ? ORDER BY id`, mt, adminID)
	}
	switch c.Method {
	case "GET":
		rows, err := list()
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rows)
	case "DELETE":
		if _, err := db.Exec(ctx, c.App.DB, `DELETE FROM passkeys WHERE id = ? AND memberType = ? AND ownerId = ?`, c.Q("id"), mt, adminID); err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(true)
	case "POST":
		rpID, origin, isAdmin, ok := c.PasskeyRP()
		if !ok || isAdmin != (mt == "A") {
			c.Ecb("이 주소에서는 패스키를 등록할 수 없습니다")
			return
		}
		bd := c.BodyMap()
		switch jsx.Str(bd.V("a")) {
		case "options":
			rows, err := query(ctx, c.App.DB, `SELECT credId, transports FROM passkeys WHERE memberType = ? AND ownerId = ? AND rpId = ?`, mt, adminID, rpID)
			if err != nil {
				c.Ecb(err)
				return
			}
			ex := []any{}
			for _, r := range rows {
				tr := []any{}
				for _, t := range strings.Split(jsx.StrOrEmpty(r.V("transports")), ",") {
					if t != "" {
						tr = append(tr, t)
					}
				}
				ex = append(ex, jsx.Obj("id", r.V("credId"), "transports", tr))
			}
			name := c.Token.Name
			if mt == "U" && c.Token.Nickname != "" {
				name = c.Token.Nickname
			}
			if name == "" {
				name = c.Token.Email
			}
			userName := c.Token.Email
			if userName == "" {
				userName = name
			}
			if mt == "A" { // 한 기기에 회원 열쇠와 같이 있으면 고르는 창에서 구분되게
				name += " (관리자)"
				userName += " (관리자)"
			}
			c.Scb(jsx.Obj("challenge", pkNewChallenge("reg", mt+":"+adminID, rpID), "rpId", rpID, "userId", pkUserHandle(mt, c.Token.HostingID, adminID),
				"userName", userName, "displayName", name, "exclude", ex, "timeout", pkChallengeTTL.Milliseconds()))
		case "register":
			if err := pkRegister(c, bd, mt, adminID, rpID, origin); err != nil {
				c.Ecb(err.Error())
				return
			}
			rows, err := list()
			if err != nil {
				c.Ecb(err)
				return
			}
			c.Scb(rows)
		default:
			c.Ecb("잘못된 요청")
		}
	default:
		c.Ecb("잘못된 요청")
	}
}

func pkRegister(c *web.Ctx, bd *jsx.OrderedMap, mt, adminID, rpID, origin string) error {
	cd, err1 := pkB64(jsx.StrOrEmpty(bd.V("clientDataJSON")))
	ad, err2 := pkB64(jsx.StrOrEmpty(bd.V("authenticatorData")))
	pub, err3 := pkB64(jsx.StrOrEmpty(bd.V("publicKey")))
	cred, err4 := pkB64(jsx.StrOrEmpty(bd.V("id")))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || len(pub) == 0 || len(cred) == 0 || len(cred) > 180 {
		return errBad("등록 정보가 올바르지 않습니다 — 이 브라우저가 패스키 등록 정보를 주지 않습니다(브라우저를 최신으로)")
	}
	ch, err := pkCheckClientData(cd, "webauthn.create", origin)
	if err != nil {
		return err
	}
	if !pkTakeChallenge(ch, "reg", mt+":"+adminID, rpID) {
		return errBad("시간이 지났습니다 — 다시 시도해 주세요")
	}
	a, err := pkParseAuthData(ad, rpID, true)
	if err != nil {
		return err
	}
	if string(a.credID) != string(cred) {
		return errBad("열쇠 id 가 맞지 않습니다")
	}
	if err := pkCheckPublicKey(pub); err != nil {
		return err
	}
	name := []rune(strings.TrimSpace(jsx.StrOrEmpty(bd.V("name"))))
	if len(name) > 60 {
		name = name[:60]
	}
	tr := []string{}
	if arr, ok := bd.V("transports").([]any); ok {
		for _, t := range arr {
			if s := jsx.StrOrEmpty(t); s != "" && len(s) <= 20 && !strings.Contains(s, ",") {
				tr = append(tr, s)
			}
		}
	}
	_, err = db.Exec(c.Context(), c.App.DB, `INSERT INTO passkeys (hostingId, memberType, ownerId, credId, publicKey, alg, signCount, rpId, name, transports) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Token.HostingID, mt, adminID, pkB64Enc(cred), pkB64Enc(pub), int(jsx.Number(bd.V("publicKeyAlgorithm"))), a.signCount, rpID, string(name), strings.Join(tr, ","))
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return errBad("이미 등록된 열쇠입니다")
		}
		return err
	}
	log.Printf("@ 패스키 등록 %s %s", mt, adminID)
	return nil
}

// loginPasskey — 패스키로 로그인(관리자 주소면 관리자, 사용자 주소면 회원)
func loginPasskey(c *web.Ctx) {
	ctx := c.Context()
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	rpID, origin, isAdmin, ok := c.PasskeyRP()
	if !ok {
		c.Ecb("이 주소에서는 패스키로 로그인할 수 없습니다")
		return
	}
	mt := "U"
	if isAdmin {
		mt = "A"
	}
	bd := c.BodyMap()
	switch jsx.Str(bd.V("a")) {
	case "options":
		c.Scb(jsx.Obj("challenge", pkNewChallenge("login", mt, rpID), "rpId", rpID, "timeout", pkChallengeTTL.Milliseconds()))
		return
	case "verify":
	default:
		c.Ecb("잘못된 요청")
		return
	}
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	if err := ensurePasskeyTable(ctx, conn); err != nil {
		c.Ecb(err)
		return
	}
	fail := func(why string) {
		log.Printf("@ 패스키 로그인 실패: %s", why)
		c.Ecb("패스키로 로그인하지 못했습니다 — " + why)
	}
	cd, err1 := pkB64(jsx.StrOrEmpty(bd.V("clientDataJSON")))
	ad, err2 := pkB64(jsx.StrOrEmpty(bd.V("authenticatorData")))
	sig, err3 := pkB64(jsx.StrOrEmpty(bd.V("signature")))
	cred, err4 := pkB64(jsx.StrOrEmpty(bd.V("id")))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || len(cred) == 0 {
		fail("요청이 올바르지 않습니다")
		return
	}
	ch, err := pkCheckClientData(cd, "webauthn.get", origin)
	if err != nil {
		fail(err.Error())
		return
	}
	if !pkTakeChallenge(ch, "login", mt, rpID) {
		fail("시간이 지났습니다. 다시 시도해 주세요")
		return
	}
	pk, err := queryOne(ctx, conn, `SELECT id, ownerId, publicKey, signCount FROM passkeys WHERE credId = ? AND memberType = ? AND rpId = ?`, pkB64Enc(cred), mt, rpID)
	if err != nil {
		c.Ecb(err)
		return
	}
	if pk == nil {
		// 한 기기에 관리자·회원 열쇠가 같이 있으면 고르는 창에 둘 다 뜬다 — 다른 쪽 열쇠를 골랐으면 그렇게 알려 준다
		if o, _ := queryOne(ctx, conn, `SELECT memberType FROM passkeys WHERE credId = ? AND rpId = ?`, pkB64Enc(cred), rpID); o != nil {
			if mt == "A" {
				fail("회원용 열쇠를 고르셨어요. 고르는 창에서 「(관리자)」 가 붙은 열쇠를 골라 주세요")
			} else {
				fail("관리자용 열쇠를 고르셨어요. 고르는 창에서 다른 열쇠(회원 계정)를 골라 주세요")
			}
			return
		}
		if mt == "A" {
			fail("등록되지 않은 열쇠입니다(마이페이지에서 지웠거나 다른 기기의 열쇠)")
		} else {
			fail("등록되지 않은 열쇠입니다(내 정보에서 지웠거나 관리자용 열쇠)")
		}
		return
	}
	a, err := pkParseAuthData(ad, rpID, false)
	if err != nil {
		fail(err.Error())
		return
	}
	pub, _ := pkB64(jsx.StrOrEmpty(pk.V("publicKey")))
	if err := pkVerifySig(pub, ad, cd, sig); err != nil {
		fail(err.Error())
		return
	}
	if !pkCountOK(uint32(jsx.Int(pk.V("signCount"))), a.signCount) {
		fail("열쇠 사용 횟수가 맞지 않습니다(복제 의심)")
		return
	}
	if mt == "U" {
		pkLoginUser(c, conn, pk, a.signCount)
		return
	}
	o, err := queryOne(ctx, conn, `SELECT adminId, groupId, name, email, json, hostingId FROM admins WHERE adminId = ? AND status >= 0`, pk.V("ownerId"))
	if err != nil {
		c.Ecb(err)
		return
	}
	if o == nil {
		fail("쓸 수 없는 계정입니다")
		return
	}
	if _, err := db.Exec(ctx, conn, `UPDATE passkeys SET signCount = ?, lastUsedAt = current_timestamp() WHERE id = ?`, a.signCount, pk.V("id")); err != nil {
		c.Ecb(err)
		return
	}
	cid := jsx.StrOrEmpty(o.V("hostingId"))
	partnerID := "0"
	if pj, ok := jsx.ParseOr(jsx.StrOrEmpty(o.V("json")), nil).(*jsx.OrderedMap); ok && jsx.Truthy(pj.V("partnerId")) {
		partnerID = jsx.Str(pj.V("partnerId"))
	}
	browserUID := c.BrowserUID()
	if browserUID == "" {
		browserUID = "UNKNOWN"
	}
	if err := setLoginDevice(c, conn, o.V("adminId"), "A", o.V("email"), browserUID, cid, ipOf(reqHeaders(c))); err != nil {
		c.Ecb(err)
		return
	}
	sa, sub := hostingTokenIDs(c, cid)
	tok := makeToken(c, jsx.Str(o.V("adminId")), jsx.StrOrEmpty(o.V("email")), "A", browserUID, cid,
		jsx.StrOrEmpty(o.V("name")), jsx.StrOrEmpty(o.V("groupId")), partnerID, sa, "", 0, sub)
	r := jsx.NewOrderedMap()
	if !c.IssueSession(tok) {
		r.Set("token", tok)
	}
	log.Printf("@ 로그인(패스키) A %s %s", cid, jsx.Str(o.V("adminId")))
	c.Scb(r)
}

// pkLoginUser — 회원 패스키 로그인 마무리(소셜 로그인과 같은 토큰). 탈퇴·정지 회원은 거절
func pkLoginUser(c *web.Ctx, conn *sql.Conn, pk *jsx.OrderedMap, count uint32) {
	ctx := c.Context()
	u, err := member.Get(ctx, c.App, conn, pk.V("ownerId"))
	if err != nil {
		c.Ecb(err)
		return
	}
	if u == nil || jsx.Number(u.V("status")) < 0 || jsx.Number(u.V("status")) == 99 {
		log.Printf("@ 패스키 로그인 실패: 쓸 수 없는 회원 %s", jsx.Str(pk.V("ownerId")))
		c.Ecb("패스키로 로그인하지 못했습니다 — 쓸 수 없는 계정입니다")
		return
	}
	if _, err := db.Exec(ctx, conn, `UPDATE passkeys SET signCount = ?, lastUsedAt = current_timestamp() WHERE id = ?`, count, pk.V("id")); err != nil {
		c.Ecb(err)
		return
	}
	cid := jsx.StrOrEmpty(u.V("hostingId"))
	browserUID := c.BrowserUID()
	if browserUID == "" {
		browserUID = "UNKNOWN"
	}
	if err := setLoginDevice(c, conn, u.V("userId"), "U", u.V("email"), browserUID, cid, ipOf(reqHeaders(c))); err != nil {
		c.Ecb(err)
		return
	}
	if s, ok := u.V("json").(string); ok {
		u.Set("Json", jsx.ParseOr(s, nil))
	}
	extra := checkRequiredUserInfo(c, u)
	name := jsx.StrOrEmpty(u.V("nickname")) // 닉네임, 없으면 이름
	if name == "" {
		name = jsx.StrOrEmpty(u.V("name"))
	}
	assoc := "0"
	if v := c.App.Settings.Hosting(cid).Path("server", "associationAdminId"); jsx.Truthy(v) {
		assoc = jsx.Str(v)
	}
	tok := makeToken(c, jsx.Str(u.V("userId")), jsx.StrOrEmpty(u.V("email")), "U", browserUID, cid, name, "-", "0", assoc, jsx.StrOrEmpty(u.V("nickname")), 0, "0")
	r := jsx.NewOrderedMap()
	if !c.IssueSession(tok) {
		r.Set("token", tok)
	}
	if extra {
		r.Set("isAdditionalInfoRequired", true)
	}
	log.Printf("@ 로그인(패스키) U %s %s", cid, jsx.Str(u.V("userId")))
	c.Scb(r)
}
