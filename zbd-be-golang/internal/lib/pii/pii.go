// Package pii — 회원 개인정보(이름·이메일·휴대폰) 암호화와 찾기 값.
//
//   - 값: `v1:` + base64url(nonce 12바이트 | 암호문 | 인증 태그 16바이트) — AES-256-GCM, 매번 다른 nonce.
//     `v1` 은 키 판 번호 — 키를 바꿀 때 v2 를 더하고 옛 값은 그대로 읽는다.
//   - 찾기: HMAC-SHA256(찾기 키, "{종류}:{정규화한 값}") 16진수 64자 — `email_hash = ?` 처럼 같은지만 본다.
//     정규화: email = 앞뒤 공백 제거 + 소문자 · mobile = 숫자만 · name = 앞뒤 공백 제거.
//
// 키: 설정 PII_ENC_KEY_V1 · PII_INDEX_KEY (각각 base64 32바이트). 없으면 쓰는 순간 오류(조용히 평문을 넣지 않게).
package pii

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"zbd.kr/be/internal/jsx"
)

const version = "v1"

type Keys struct {
	enc, index []byte
}

// New — base64 키 두 개로. 길이가 32바이트가 아니면 오류.
func New(encB64, indexB64 string) (*Keys, error) {
	dec := func(name, s string) ([]byte, error) {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
		if err != nil || len(b) != 32 {
			return nil, fmt.Errorf("[pii] %s 가 없거나 32바이트가 아닙니다", name)
		}
		return b, nil
	}
	e, err := dec("PII_ENC_KEY_V1", encB64)
	if err != nil {
		return nil, err
	}
	i, err := dec("PII_INDEX_KEY", indexB64)
	if err != nil {
		return nil, err
	}
	return &Keys{enc: e, index: i}, nil
}

func empty(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	return false
}

// Enc — 암호화. 빈 값은 "".
func (k *Keys) Enc(v any) (string, error) {
	if empty(v) {
		return "", nil
	}
	if k == nil {
		return "", errors.New("[pii] 키가 없다")
	}
	block, err := aes.NewCipher(k.enc)
	if err != nil {
		return "", err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := g.Seal(nil, nonce, []byte(jsx.Str(v)), nil) // 암호문 | 태그
	out := append(nonce, sealed...)
	return version + ":" + base64.RawURLEncoding.EncodeToString(out), nil
}

// MustEnc — 실패하면 panic(키가 없으면 저장하면 안 된다)
func (k *Keys) MustEnc(v any) string {
	s, err := k.Enc(v)
	if err != nil {
		panic(err)
	}
	return s
}

// Dec — 복호화. 빈 값·모르는 모양·위조는 "".
func (k *Keys) Dec(v any) string {
	if empty(v) || k == nil {
		return ""
	}
	s := jsx.Str(v)
	if !strings.HasPrefix(s, version+":") {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s[len(version)+1:], "="))
	if err != nil || len(raw) < 12+16 {
		return ""
	}
	block, err := aes.NewCipher(k.enc)
	if err != nil {
		return ""
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	p, err := g.Open(nil, raw[:12], raw[12:], nil)
	if err != nil {
		return ""
	}
	return string(p)
}

type Kind string

const (
	Email  Kind = "email"
	Mobile Kind = "mobile"
	Name   Kind = "name"
)

func normalize(kind Kind, v any) string {
	s := ""
	if v != nil {
		s = strings.TrimSpace(jsx.Str(v))
	}
	switch kind {
	case Email:
		return strings.ToLower(s)
	case Mobile:
		var b strings.Builder
		for _, r := range s {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return strings.TrimFunc(s, unicode.IsSpace)
}

// Hash — 찾기 값. 빈 값은 "".
func (k *Keys) Hash(kind Kind, v any) string {
	n := normalize(kind, v)
	if n == "" || k == nil {
		return ""
	}
	m := hmac.New(sha256.New, k.index)
	m.Write([]byte(string(kind) + ":" + n))
	return hex.EncodeToString(m.Sum(nil))
}

// DecMember — 회원 행의 name·email·mobile 을 풀어 준다(있는 칸만, 제자리).
func (k *Keys) DecMember(m *jsx.OrderedMap) *jsx.OrderedMap {
	if m == nil {
		return m
	}
	for _, c := range []string{"name", "email", "mobile"} {
		if m.Has(c) {
			m.Set(c, k.Dec(m.V(c)))
		}
	}
	return m
}

// DecMembers — 여러 행
func (k *Keys) DecMembers(rows []*jsx.OrderedMap) []*jsx.OrderedMap {
	for _, r := range rows {
		k.DecMember(r)
	}
	return rows
}
