package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"testing"
)

// 가짜 인증기 — 브라우저가 실제로 만드는 모양 그대로(authenticatorData·clientDataJSON·서명)
func pkFakeAuthData(rpID string, flags byte, count uint32, credID []byte) []byte {
	h := sha256.Sum256([]byte(rpID))
	b := append([]byte{}, h[:]...)
	b = append(b, flags)
	b = binary.BigEndian.AppendUint32(b, count)
	if credID != nil {
		b = append(b, make([]byte, 16)...) // aaguid
		b = binary.BigEndian.AppendUint16(b, uint16(len(credID)))
		b = append(b, credID...)
		b = append(b, 0xa5) // COSE 키(읽지 않는다)
	}
	return b
}

func pkFakeClientData(typ, ch, origin string) []byte {
	b, _ := json.Marshal(map[string]any{"type": typ, "challenge": ch, "origin": origin, "crossOrigin": false})
	return b
}

// 정책: 패스키 로그인은 출처·종류·RP ID·사람/본인 확인(UP·UV)·서명·사용 횟수를 모두 본다
func TestPasskeyVerifyPolicy(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spki, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	sign := func(ad, cd []byte) []byte {
		h := sha256.Sum256(cd)
		d := sha256.Sum256(append(append([]byte{}, ad...), h[:]...))
		s, _ := ecdsa.SignASN1(rand.Reader, key, d[:])
		return s
	}
	const origin = "https://backyard.zbd.kr"
	cd := pkFakeClientData("webauthn.get", "abc", origin)
	if ch, err := pkCheckClientData(cd, "webauthn.get", origin); err != nil || ch != "abc" {
		t.Fatalf("정상 clientData 거절: %v", err)
	}
	if _, err := pkCheckClientData(cd, "webauthn.get", "https://evil.example"); err == nil {
		t.Fatal("다른 출처를 받았다")
	}
	if _, err := pkCheckClientData(cd, "webauthn.create", origin); err == nil {
		t.Fatal("다른 종류를 받았다")
	}

	ad := pkFakeAuthData("zbd.kr", pkFlagUP|pkFlagUV, 5, nil)
	if a, err := pkParseAuthData(ad, "zbd.kr", false); err != nil || a.signCount != 5 {
		t.Fatalf("정상 authData 거절: %v", err)
	}
	if _, err := pkParseAuthData(ad, "evil.example", false); err == nil {
		t.Fatal("다른 RP ID 를 받았다")
	}
	if _, err := pkParseAuthData(pkFakeAuthData("zbd.kr", pkFlagUP, 5, nil), "zbd.kr", false); err == nil {
		t.Fatal("본인 확인(UV) 없이 받았다")
	}

	if err := pkVerifySig(spki, ad, cd, sign(ad, cd)); err != nil {
		t.Fatalf("정상 서명 거절: %v", err)
	}
	if err := pkVerifySig(spki, ad, pkFakeClientData("webauthn.get", "xyz", origin), sign(ad, cd)); err == nil {
		t.Fatal("다른 clientData 의 서명을 받았다")
	}
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ospki, _ := x509.MarshalPKIXPublicKey(&other.PublicKey)
	if err := pkVerifySig(ospki, ad, cd, sign(ad, cd)); err == nil {
		t.Fatal("다른 열쇠의 서명을 받았다")
	}
	if pkCheckPublicKey(spki) != nil {
		t.Fatal("P-256 공개키 거절")
	}

	// 등록: 붙은 열쇠 id 를 읽는다, AT 없으면 거절
	reg := pkFakeAuthData("zbd.kr", pkFlagUP|pkFlagUV|pkFlagAT, 0, []byte("cred-1"))
	if a, err := pkParseAuthData(reg, "zbd.kr", true); err != nil || string(a.credID) != "cred-1" {
		t.Fatalf("등록 authData: %v", err)
	}
	if _, err := pkParseAuthData(ad, "zbd.kr", true); err == nil {
		t.Fatal("등록 정보 없이 받았다")
	}

	// 사용 횟수
	for _, c := range []struct {
		stored, got uint32
		ok          bool
	}{{0, 0, true}, {5, 6, true}, {5, 5, false}, {5, 3, false}, {5, 0, false}, {0, 1, true}} {
		if pkCountOK(c.stored, c.got) != c.ok {
			t.Fatalf("횟수 %d→%d 기대 %v", c.stored, c.got, c.ok)
		}
	}
}

// 정책: 도전값은 한 번만·같은 용도·같은 사람(관리자/회원)·같은 RP 에서만
func TestPasskeyChallengePolicy(t *testing.T) {
	ch := pkNewChallenge("reg", "340", "zbd.kr")
	if pkTakeChallenge(ch, "login", "340", "zbd.kr") {
		t.Fatal("다른 용도로 받았다")
	}
	ch = pkNewChallenge("reg", "340", "zbd.kr")
	if pkTakeChallenge(ch, "reg", "352", "zbd.kr") {
		t.Fatal("다른 관리자로 받았다")
	}
	ch = pkNewChallenge("login", "A", "zbd.kr")
	if pkTakeChallenge(ch, "login", "U", "zbd.kr") {
		t.Fatal("관리자 로그인 도전값을 회원 로그인에 받았다")
	}
	ch = pkNewChallenge("login", "A", "zbd.kr")
	if !pkTakeChallenge(ch, "login", "A", "zbd.kr") {
		t.Fatal("정상 도전값 거절")
	}
	if pkTakeChallenge(ch, "login", "A", "zbd.kr") {
		t.Fatal("같은 도전값을 두 번 받았다")
	}
}
