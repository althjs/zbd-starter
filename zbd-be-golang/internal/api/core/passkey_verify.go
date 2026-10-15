package core

// 패스키(WebAuthn) 검증 — 외부 라이브러리 없이 표준 라이브러리로(2026-10-04 대표 「PWA 지문 로그인」).
//
// 등록: 브라우저가 response.getPublicKey()(SPKI DER) 를 같이 보내 CBOR(COSE 키) 를 풀 필요가 없다. attestation 은 "none" —
// 로그인한 관리자가 자기 기기를 등록하는 것이라 기기 제조사 증명은 보지 않는다.
// 로그인: 서명 = 개인키로 authenticatorData || SHA256(clientDataJSON). 공개키 형식은 x509.ParsePKIXPublicKey 가 가린다(ES256·RS256·EdDSA).
// 테스트 passkey_policy_test.go

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
)

const (
	pkFlagUP = 0x01 // 사람이 있다(눌렀다)
	pkFlagUV = 0x04 // 지문·얼굴·기기 잠금으로 본인 확인
	pkFlagAT = 0x40 // 등록 정보(새 열쇠) 가 붙어 있다
)

// pkB64 — base64url(패딩 없음·있음 모두)
func pkB64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(strings.TrimSpace(s), "="))
}

func pkB64Enc(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

type pkClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

// pkCheckClientData — 종류(webauthn.create|get)·도전값·출처가 맞는가. 맞으면 도전값(base64url) 을 돌려준다
func pkCheckClientData(raw []byte, wantType, origin string) (string, error) {
	var cd pkClientData
	if err := json.Unmarshal(raw, &cd); err != nil {
		return "", errors.New("clientData 를 읽을 수 없습니다")
	}
	if cd.Type != wantType {
		return "", errors.New("요청 종류가 맞지 않습니다")
	}
	if cd.Origin != origin {
		return "", errors.New("출처가 맞지 않습니다")
	}
	if cd.Challenge == "" {
		return "", errors.New("도전값이 없습니다")
	}
	return strings.TrimRight(cd.Challenge, "="), nil
}

type pkAuthData struct {
	flags     byte
	signCount uint32
	credID    []byte // 등록 때만
}

// pkParseAuthData — RP ID 해시가 맞고 사람·본인 확인(UP·UV) 이 됐는지. 등록이면 붙은 열쇠 id 까지
func pkParseAuthData(b []byte, rpID string, register bool) (*pkAuthData, error) {
	if len(b) < 37 {
		return nil, errors.New("authenticatorData 가 짧습니다")
	}
	h := sha256.Sum256([]byte(rpID))
	if subtle.ConstantTimeCompare(b[:32], h[:]) != 1 {
		return nil, errors.New("RP ID 가 맞지 않습니다")
	}
	a := &pkAuthData{flags: b[32], signCount: binary.BigEndian.Uint32(b[33:37])}
	if a.flags&pkFlagUP == 0 || a.flags&pkFlagUV == 0 {
		return nil, errors.New("본인 확인(지문·얼굴·화면 잠금) 이 되지 않았습니다")
	}
	if register {
		if a.flags&pkFlagAT == 0 || len(b) < 37+16+2 {
			return nil, errors.New("등록 정보가 없습니다")
		}
		n := int(binary.BigEndian.Uint16(b[53:55]))
		if n == 0 || len(b) < 55+n {
			return nil, errors.New("열쇠 id 가 없습니다")
		}
		a.credID = b[55 : 55+n]
	}
	return a, nil
}

// pkVerifySig — spki(공개키 DER) 로 authenticatorData || SHA256(clientDataJSON) 의 서명 확인
func pkVerifySig(spki, authData, clientData, sig []byte) error {
	pub, err := x509.ParsePKIXPublicKey(spki)
	if err != nil {
		return errors.New("공개키를 읽을 수 없습니다")
	}
	ch := sha256.Sum256(clientData)
	msg := append(append([]byte{}, authData...), ch[:]...)
	d := sha256.Sum256(msg)
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		if ecdsa.VerifyASN1(k, d[:], sig) {
			return nil
		}
	case *rsa.PublicKey:
		if rsa.VerifyPKCS1v15(k, crypto.SHA256, d[:], sig) == nil {
			return nil
		}
	case ed25519.PublicKey:
		if ed25519.Verify(k, msg, sig) {
			return nil
		}
	default:
		return errors.New("지원하지 않는 열쇠 형식")
	}
	return errors.New("서명이 맞지 않습니다")
}

// pkCountOK — 서명 횟수: 둘 다 0 이 아니면 늘어야 한다(같거나 줄면 열쇠가 복제된 것으로 본다). 애플·구글 동기화 패스키는 늘 0
func pkCountOK(stored, got uint32) bool {
	return got == 0 && stored == 0 || got > stored
}

func errBad(s string) error { return errors.New(s) }

// pkCheckPublicKey — 등록할 공개키가 읽히고 지원하는 형식인가(P-256 ECDSA · RSA 2048 이상 · Ed25519)
func pkCheckPublicKey(spki []byte) error {
	pub, err := x509.ParsePKIXPublicKey(spki)
	if err != nil {
		return errors.New("공개키를 읽을 수 없습니다")
	}
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		if k.Curve.Params().Name == "P-256" {
			return nil
		}
	case *rsa.PublicKey:
		if k.N.BitLen() >= 2048 {
			return nil
		}
	case ed25519.PublicKey:
		return nil
	}
	return errors.New("지원하지 않는 열쇠 형식")
}
