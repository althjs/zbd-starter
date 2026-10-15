// Package auth — 로그인 토큰 해석.
//
// 토큰은 지금 서비스가 발급하는 것과 같은 모양이다: AES-256-CBC(키 32자·IV 16자, PKCS#7) 로 잠근 문자열을 `|` 로 가른 칸.
//
//	0 key(회원·관리자 번호) · 1 email · 2 발급 시각 · 3 memberType(A 관리자 / U 회원) · 4 브라우저 id · 5 hostingId
//	6 name · 7 groupId · 8~9 (예약) · 10 nickname · 11 갱신 시각 · 12 (예약)
//
// 칸이 13개보다 적거나 memberType 이 A·U 가 아니면 위조·옛 토큰으로 보고 **해석하지 않는다**.
// 받는 곳: `Authorization: Bearer …` 헤더. 세션 쿠키(zbd_user_session·zbd_admin_session)는 web/session.go 가 읽는다.
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const minTokenFields = 13

// AnyBrowser — 브라우저에 묶이지 않은 토큰의 브라우저 id(관리자가 발급해 붙여 넣는 토큰)
const AnyBrowser = "ANY_BROWSER"

type Token struct {
	Key        string
	Email      string
	TS         int64
	MemberType string // A | U
	BrowserUID string
	HostingID  string
	Name       string
	GroupID    string
	PartnerID  int64
	SiteAdmin  int64 // 칸 9
	Nickname   string
	RTS        int64
	SubAdminID int64 // 칸 12
	Etc        string
	// HeaderUID — 브라우저에 묶이지 않은 토큰(브라우저 id 가 AnyBrowser) 은 요청 헤더 zbduid 를 쓴다
	HeaderUID string
}

// Valid — 해석에 성공했는가
func (t *Token) Valid() bool { return t != nil && t.Key != "" && t.Key != "0" }

// IsAdmin · IsUser — 관리자 / 회원
func (t *Token) IsAdmin() bool { return t.Valid() && t.MemberType == "A" }
func (t *Token) IsUser() bool  { return t.Valid() && t.MemberType == "U" }

// ID — 번호(관리자면 adminId, 회원이면 userId). 숫자가 아니면 0
func (t *Token) ID() int64 {
	n, _ := strconv.ParseInt(t.Key, 10, 64)
	return n
}

// parseLeadingInt — 앞쪽 숫자만 읽는다(없으면 0).
func parseLeadingInt(s string) int64 {
	s = strings.TrimSpace(s)
	end := 0
	for end < len(s) && ((s[end] >= '0' && s[end] <= '9') || (end == 0 && (s[end] == '-' || s[end] == '+'))) {
		end++
	}
	n, _ := strconv.ParseInt(s[:end], 10, 64)
	return n
}

// decodeBase64Lenient — 패딩 없음·URL-safe 문자도 받는다.
func decodeBase64Lenient(s string) ([]byte, error) {
	s = strings.NewReplacer("-", "+", "_", "/", " ", "+").Replace(strings.TrimSpace(s))
	return base64.RawStdEncoding.DecodeString(strings.TrimRight(s, "="))
}

// Decrypt — AES-256-CBC 복호화(위젯 티켓 등 같은 키로 잠근 다른 값에도 쓴다).
func Decrypt(text, key, iv string) (string, error) {
	data, err := decodeBase64Lenient(text)
	if err != nil {
		return "", err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return "", errors.New("bad ciphertext length")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, []byte(iv)).CryptBlocks(out, data)
	pad := int(out[len(out)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(out) {
		return "", errors.New("bad padding")
	}
	for _, c := range out[len(out)-pad:] {
		if int(c) != pad {
			return "", errors.New("bad padding")
		}
	}
	return string(out[:len(out)-pad]), nil
}

// ParseString — 토큰 문자열 하나를 해석한다. 실패하면 빈 토큰(Valid()=false).
func ParseString(raw, aesKey, aesIV string) *Token {
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "Bearer "))
	if raw == "" {
		return &Token{}
	}
	plain, err := Decrypt(raw, aesKey, aesIV)
	if err != nil || plain == "" {
		return &Token{}
	}
	f := strings.Split(plain, "|")
	if len(f) < minTokenFields || (f[3] != "A" && f[3] != "U") {
		return &Token{}
	}
	t := &Token{
		Key:        f[0],
		Email:      f[1],
		TS:         parseLeadingInt(f[2]),
		MemberType: f[3],
		BrowserUID: f[4],
		HostingID:  f[5],
		Name:       f[6],
		GroupID:    f[7],
		PartnerID:  parseLeadingInt(f[8]),
		SiteAdmin:  parseLeadingInt(f[9]),
		Nickname:   f[10],
		RTS:        parseLeadingInt(f[11]),
		SubAdminID: parseLeadingInt(f[12]),
	}
	if len(f) > 13 {
		t.Etc = f[13]
	}
	if t.GroupID == "" {
		t.GroupID = "-"
	}
	return t
}

// Parse — 요청에서 토큰을 꺼내 해석한다.
func Parse(r *http.Request, aesKey, aesIV, cookieName string) *Token {
	raw := r.Header.Get("Authorization")
	if strings.TrimSpace(raw) == "" && cookieName != "" {
		if c, err := r.Cookie(cookieName); err == nil {
			raw = c.Value
		}
	}
	t := ParseString(raw, aesKey, aesIV)
	if t.BrowserUID == AnyBrowser {
		t.HeaderUID = r.Header.Get("zbduid")
		if t.HeaderUID == "" {
			// 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
			t.HeaderUID = r.Header.Get("oxuid")
		}
	}
	return t
}

// Encrypt — Decrypt 의 반대(AES-256-CBC, PKCS#7, base64 표준).
func Encrypt(text, key, iv string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}
	b := []byte(text)
	pad := aes.BlockSize - len(b)%aes.BlockSize
	for i := 0; i < pad; i++ {
		b = append(b, byte(pad))
	}
	out := make([]byte, len(b))
	cipher.NewCBCEncrypter(block, []byte(iv)).CryptBlocks(out, b)
	return base64.StdEncoding.EncodeToString(out)
}

// Refreshed — 같은 토큰을 갱신 시각만 rts 로 바꿔 다시 잠근다(세션 쿠키 연장용).
func (t *Token) Refreshed(rts int64, key, iv string) string {
	i := func(n int64) string { return strconv.FormatInt(n, 10) }
	g := t.GroupID
	if g == "" {
		g = "-"
	}
	parts := []string{t.Key, t.Email, i(t.TS), t.MemberType, t.BrowserUID, t.HostingID, t.Name, g,
		i(t.PartnerID), i(t.SiteAdmin), t.Nickname, i(rts), i(t.SubAdminID)}
	if t.Etc != "" {
		parts = append(parts, t.Etc)
	}
	return Encrypt(strings.Join(parts, "|"), key, iv)
}

// LastActive — 마지막으로 발급·갱신된 시각(ms)
func (t *Token) LastActive() int64 {
	if t.RTS > t.TS {
		return t.RTS
	}
	return t.TS
}
