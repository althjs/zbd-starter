// Package config — 서버 설정. **환경변수**로만 받는다(값은 git 에 두지 않는다).
//
// 읽는 순서: 실행 폴더의 `.env` → `.env.local`(덮어씀) → 프로세스 환경변수(가장 우선).
// `.env*` 는 git 에 올리지 않는다 — 항목 이름은 `.env.example` 에 있다.
//
//	PORT                     기본 5001
//	DB_HOST · DB_PORT(3306) · DB_NAME · DB_USER · DB_PASSWORD · DB_CHARSET(utf8mb4) · DB_CONN_LIMIT(20)
//	AES_256_ENCRYPTION_KEY · AES_256_ENCRYPTION_IV   로그인 토큰 복호화(지금 서비스가 발급한 토큰과 같은 키)
//	AUTH_COOKIE_KEY          쿠키로 온 토큰의 이름(선택)
package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type DBConfig struct {
	Host      string
	Port      int
	Database  string
	User      string
	Password  string
	Charset   string
	ConnLimit int
}

type Config struct {
	env        map[string]string
	Port       string
	DB         DBConfig
	AESKey     string // 32자로 맞춘 키
	AESIV      string // 16자
	CookieName string
}

// padKey — 모자라면 "0" 으로 채우고 넘치면 자른다(토큰을 발급하는 쪽과 같은 규칙).
func padKey(s string, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i < len(s) {
			b.WriteByte(s[i])
		} else {
			b.WriteByte('0')
		}
	}
	return b.String()
}

// readDotEnv — `KEY=value`, `KEY = value`(공백), 따옴표 값을 읽는다. 파일이 없으면 빈 값.
func readDotEnv(path string, into map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		into[strings.TrimSpace(k)] = dotEnvValue(v)
	}
}

// dotEnvValue — 따옴표 값은 닫는 따옴표까지만(뒤의 ` # 설명` 은 버린다 — 2026-10-03 ES 비밀번호에 주석이 붙어 ES 인증이 실패했다).
// 따옴표 없는 값은 공백을 품을 수 있어(옛 AES 키) 그대로 두되, 앞에 공백이 있는 ` #` 뒤는 설명으로 본다.
func dotEnvValue(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') {
		if end := strings.IndexByte(v[1:], v[0]); end >= 0 {
			return v[1 : 1+end]
		}
	}
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return strings.Trim(v, `"'`)
}

func Load() (*Config, error) {
	env := map[string]string{}
	readDotEnv(".env", env)
	readDotEnv(".env.local", env)
	get := func(k, d string) string {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
		if v := env[k]; v != "" {
			return v
		}
		return d
	}
	num := func(k string, d int) int {
		if n, err := strconv.Atoi(get(k, "")); err == nil && n > 0 {
			return n
		}
		return d
	}

	c := &Config{
		env:  env,
		Port: get("PORT", "5001"),
		DB: DBConfig{
			Host:      get("DB_HOST", ""),
			Port:      num("DB_PORT", 3306),
			Database:  get("DB_NAME", ""),
			User:      get("DB_USER", ""),
			Password:  get("DB_PASSWORD", ""),
			Charset:   get("DB_CHARSET", "utf8mb4"),
			ConnLimit: num("DB_CONN_LIMIT", 20),
		},
		AESKey:     padKey(get("AES_256_ENCRYPTION_KEY", ""), 32),
		AESIV:      padKey(get("AES_256_ENCRYPTION_IV", ""), 16),
		CookieName: get("AUTH_COOKIE_KEY", ""),
	}
	if c.DB.Host == "" || c.DB.Database == "" {
		return nil, errors.New("DB 설정이 없습니다 — .env.local 또는 환경변수 DB_HOST·DB_NAME 등을 확인하세요")
	}
	return c, nil
}

// Get — 아무 설정값(환경변수 → .env.local → .env). 없으면 "".
//
//	비밀값(API 키·비밀번호)은 여기서 읽어 쓰고 **로그·응답에 찍지 않는다**.
func (c *Config) Get(k string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	if c == nil {
		return ""
	}
	return c.env[k]
}

// GetOr — 없으면 d
func (c *Config) GetOr(k, d string) string {
	if v := c.Get(k); v != "" {
		return v
	}
	return d
}

// Bool — "true"·"1" 이면 true
func (c *Config) Bool(k string) bool {
	v := strings.ToLower(c.Get(k))
	return v == "true" || v == "1"
}
