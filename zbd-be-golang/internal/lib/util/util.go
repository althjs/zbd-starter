// Package util — 여러 기능이 같이 쓰는 작은 도우미(날짜 문자열·자리 채우기·암호·사무실 IP 등).
//
// 지금 서비스의 공용 도우미와 결과가 같다. 새 도우미가 둘 이상의 기능에서 필요하면 여기 더한다.
package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// KST — 서버가 쓰는 시간대(지금 서비스는 TZ=Asia/Seoul 로 돈다)
var KST = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Seoul"); err == nil {
		return loc
	}
	return time.FixedZone("KST", 9*3600)
}()

// YMD — "2026-10-03" (getYMD). hm=true 면 " 13:05", sec=true 면 ":09" 까지. 한국 시각 기준.
func YMD(t time.Time, hm, sec bool) string {
	t = t.In(KST)
	s := t.Format("2006-01-02")
	if hm {
		s += " " + t.Format("15:04")
		if sec {
			s += ":" + t.Format("05")
		}
	}
	return s
}

// Today — 오늘 "YYYY-MM-DD"(한국)
func Today() string { return YMD(time.Now(), false, false) }

// AES — 지금 서비스의 encryptAES256/decryptAES256 (AES-256-CBC, base64). 키 32자·IV 16자.
type AES struct{ Key, IV string }

func (a AES) Encrypt(s string) string {
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return ""
	}
	b := []byte(s)
	pad := aes.BlockSize - len(b)%aes.BlockSize
	for i := 0; i < pad; i++ {
		b = append(b, byte(pad))
	}
	out := make([]byte, len(b))
	cipher.NewCBCEncrypter(block, []byte(a.IV)).CryptBlocks(out, b)
	return base64.StdEncoding.EncodeToString(out)
}

func (a AES) Decrypt(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(strings.TrimSpace(s), "="))
		if err != nil {
			return "", err
		}
	}
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return "", err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return "", errBad
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, []byte(a.IV)).CryptBlocks(out, data)
	pad := int(out[len(out)-1])
	if pad == 0 || pad > aes.BlockSize {
		return "", errBad
	}
	return string(out[:len(out)-pad]), nil
}

type strErr string

const errBad = strErr("bad decrypt")

func (e strErr) Error() string { return string(e) }

// BcryptHash · BcryptCompare — 비밀번호(bcryptjs 와 같은 $2a/$2b 형식, 비용 10)
func BcryptHash(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	return string(b), err
}

func BcryptCompare(pw, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pw)) == nil
}

// ClientIPs — 요청 헤더에서 본 IP 후보(x-forwarded-for, x-real-ip, x-client-ip, cf-connecting-ip). "::ffff:" 는 뗀다.
func ClientIPs(h http.Header) []string {
	var out []string
	push := func(v string) {
		for _, x := range strings.Split(v, ",") {
			ip := strings.TrimSpace(x)
			ip = strings.TrimPrefix(strings.TrimPrefix(ip, "::ffff:"), "::FFFF:")
			if ip != "" {
				out = append(out, ip)
			}
		}
	}
	push(h.Get("x-forwarded-for"))
	push(h.Get("x-real-ip"))
	push(h.Get("x-client-ip"))
	push(h.Get("cf-connecting-ip"))
	return out
}

// OfficeIPsFromSettings — 관리자 「환경 설정 › 관리자 페이지 › 사무실 IP」(SITE.officeIps). app 이 시작할 때 채운다(없으면 .env 만).
var OfficeIPsFromSettings func() []string

// OfficeIPs — 사무실 IP = 설정 OFFICE_IPS(.env, 쉼표) + 관리자 화면 설정. 코드에 박아 둔 기본값은 없다(2026-10-16).
func OfficeIPs(conf string) []string {
	return MergeIPs(SplitIPs(conf), settingsOfficeIPs())
}

func settingsOfficeIPs() []string {
	if OfficeIPsFromSettings == nil {
		return nil
	}
	return OfficeIPsFromSettings()
}

// SplitIPs — 쉼표·공백·줄바꿈으로 나뉜 IP 목록
func SplitIPs(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r' })
}

// MergeIPs — 겹치지 않게 합친다(앞 순서 유지)
func MergeIPs(lists ...[]string) []string {
	var out []string
	seen := map[string]bool{}
	for _, l := range lists {
		for _, s := range l {
			if s = strings.TrimSpace(s); s != "" && !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	return out
}

// IsOffice — 사무실·서버 자신(127.0.0.1) 에서 온 요청인가
func IsOffice(h http.Header, office []string) bool {
	for _, ip := range ClientIPs(h) {
		if ip == "127.0.0.1" || strings.Contains(ip, "ffff:127.0.0.1") {
			return true
		}
		for _, o := range office {
			if ip == o {
				return true
			}
		}
	}
	return false
}
