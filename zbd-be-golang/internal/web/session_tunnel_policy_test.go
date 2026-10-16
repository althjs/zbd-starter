package web

import (
	"net/http/httptest"
	"testing"
)

// 외부 접속 주소 — 터널 헤더를 믿는 조건(session_tunnel.go)
func TestTunnelPairPolicy(t *testing.T) {
	cases := []struct{ base, pub, wantBase, wantPub string }{
		{"starter-76.zbd.co.kr", "starter-76.zbd.co.kr", "starter-76.zbd.co.kr", "starter-76.zbd.co.kr"},
		{"starter-76.zbd.co.kr", "chat.starter-76.zbd.co.kr", "starter-76.zbd.co.kr", "chat.starter-76.zbd.co.kr"},
		{"starter-76.zbd.co.kr", "admin.starter-76.zbd.co.kr:443", "starter-76.zbd.co.kr", "admin.starter-76.zbd.co.kr"},
		{"starter-76.zbd.co.kr", "evil.starter-76.zbd.co.kr", "", ""},   // 정해진 앱 주소만
		{"starter-76.zbd.co.kr", "chat.starter-11.zbd.co.kr", "", ""},   // 다른 고객
		{"starter-76.zbd.co.kr", "x.chat.starter-76.zbd.co.kr", "", ""}, // 한 단계만
		{"zbd.kr", "zbd.kr", "", ""}, // 사이트 주소 모양
		{"starter-76.zbd.co.kr.evil.com", "starter-76.zbd.co.kr.evil.com", "", ""},
		{"", "starter-76.zbd.co.kr", "", ""},
	}
	for _, x := range cases {
		b, p := tunnelPair(x.base, x.pub)
		if b != x.wantBase || p != x.wantPub {
			t.Errorf("tunnelPair(%q,%q) = %q,%q want %q,%q", x.base, x.pub, b, p, x.wantBase, x.wantPub)
		}
	}
}

// 이 PC 안(개발 화면 프록시)에서 온 요청만 — 배포한 사이트(Host = 실제 주소)·밖에서 온 요청은 헤더를 무시
func TestTunnelLoopbackOnly(t *testing.T) {
	mk := func(host, remote string) bool {
		r := httptest.NewRequest("GET", "/service/x", nil)
		r.Host, r.RemoteAddr = host, remote
		r.Header.Set("X-Zbd-Tunnel-Base", "starter-76.zbd.co.kr")
		r.Header.Set("X-Zbd-Public-Host", "chat.starter-76.zbd.co.kr")
		_, pub := tunnelHosts(r)
		return pub != ""
	}
	if !mk("127.0.0.1:5001", "127.0.0.1:50000") || !mk("localhost:5001", "[::1]:50000") {
		t.Error("루프백 요청은 터널로 봐야 한다")
	}
	if mk("example.com", "127.0.0.1:50000") || mk("127.0.0.1:5001", "203.0.113.7:50000") {
		t.Error("실제 주소·밖에서 온 요청은 터널 헤더를 무시해야 한다")
	}
}
