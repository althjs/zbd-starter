package web

// 사이트 공개 여부(2026-10-08 대표 — 즈비드 호스팅 고객만): SITE 설정 siteClosed = true 면 사용자 주소는 「준비 중」 화면만.
// 호스팅 사이트(.env HOSTING_DOMAIN — 즈비드 호스팅 서버가 넣는다)에서만 먹는다. 관리자 주소는 늘 열려 있다.

import (
	"strings"

	"zbd.kr/be/internal/jsx"
)

// IsHostedSite — 즈비드 호스팅으로 올린 사이트인가(.env HOSTING_DOMAIN)
func IsHostedSite(env func(string) string) bool {
	return strings.TrimSpace(env("HOSTING_DOMAIN")) != ""
}

// SiteOpen — 사이트를 밖에 보이는가(값 없음 = 공개)
func SiteOpen(site *jsx.OrderedMap) bool { return site == nil || !jsx.Truthy(site.V("siteClosed")) }

// ShowClosedPage — 이 요청에 「준비 중」 화면을 낼지
func ShowClosedPage(hosted, open, adminHost bool) bool { return hosted && !open && !adminHost }
