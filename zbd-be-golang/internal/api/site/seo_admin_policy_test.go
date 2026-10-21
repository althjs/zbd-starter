package site

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

func TestApplyAdminSeo(t *testing.T) {
	base := func() *seoData {
		return &seoData{title: "내 사이트", desc: "사이트 설명", image: "/upload/ZBD/site/og.png", siteName: "내 회사"}
	}

	s := base()
	applyAdminSeo(s, jsx.Obj("title", "오예 그룹웨어", "description", "오예 공식 그룹웨어", "image", jsx.Obj("name", "admin_og.png")), "/upload/ZBD", "https://admin.example.com", "/")
	if s.title != "오예 그룹웨어" || s.siteName != "오예 그룹웨어" || s.desc != "오예 공식 그룹웨어" || s.image != "/upload/ZBD/site/admin_og.png" {
		t.Fatalf("값이 안 들어감: %+v", s)
	}
	if s.canonical != "https://admin.example.com/" || s.robots != "noindex, nofollow" {
		t.Fatalf("주소·robots: %q %q", s.canonical, s.robots)
	}

	// 비운 칸은 사이트 값 그대로
	s = base()
	applyAdminSeo(s, jsx.Obj("title", "  ", "image", jsx.Obj("name", "")), "/upload/ZBD", "", "/admin")
	if s.title != "내 사이트" || s.desc != "사이트 설명" || s.image != "/upload/ZBD/site/og.png" || s.siteName != "내 회사" || s.canonical != "" {
		t.Fatalf("비운 칸이 바뀜: %+v", s)
	}
	s = base()
	applyAdminSeo(s, nil, "/upload/ZBD", "http://localhost:5000", "/admin")
	if s.title != "내 사이트" || s.canonical != "http://localhost:5000/admin" {
		t.Fatalf("설정 없음: %+v", s)
	}
	if adminOrigin("localhost:5000", true) != "http://localhost:5000" || adminOrigin("a.b", false) != "https://a.b" || adminOrigin("", false) != "" {
		t.Fatal("adminOrigin")
	}
}
