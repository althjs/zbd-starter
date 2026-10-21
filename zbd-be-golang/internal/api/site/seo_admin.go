package site

import (
	"strings"

	"zbd.kr/be/internal/jsx"
)

// 관리자 주소 링크 미리보기 — SITE.adminSeo {title, description, image:{name}}(「환경 설정 › 관리자 페이지 › 관리자 주소 링크 미리보기」).
// 관리자 주소(admin.…)로 들어온 문서만 이 값을 쓰고, 비운 칸은 사이트 값 그대로. 관리자 화면은 검색에 안 나오게 noindex.
func applyAdminSeo(s *seoData, m *jsx.OrderedMap, imgBase, origin, path string) {
	if m != nil {
		if t := strings.TrimSpace(orStr(m.V("title"))); t != "" {
			s.title = t
			s.siteName = t
		}
		if d := strings.TrimSpace(orStr(m.V("description"))); d != "" {
			s.desc = d
		}
		if n := strings.TrimSpace(orStr(m.Map("image").V("name"))); n != "" {
			s.image = imgBase + "/site/" + n
		}
	}
	if origin != "" {
		s.canonical = origin + path
	}
	s.robots = "noindex, nofollow"
}

// adminOrigin — 관리자 문서의 og:url 앞부분(개발 주소는 http)
func adminOrigin(host string, local bool) string {
	if host == "" {
		return ""
	}
	if local {
		return "http://" + host
	}
	return "https://" + host
}
