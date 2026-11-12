package web

// 관리자 API 영구 토큰(2026-11-12 대표) — 토큰 칸 13(Etc)에 `ptk:{번호}` 가 든 관리자 토큰은 표 admin_api_tokens 에서
// 쓰는 중(status 0)·기간 안일 때만 통과한다(지우면 바로 막힘). 판정은 순수 함수 ApiTokenUsable(정책 테스트).

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/lib/auth"
)

// ApiTokenPrefix — 영구 토큰 표시(토큰 Etc 칸 앞머리)
const ApiTokenPrefix = "ptk:"

// ApiTokenID — 토큰의 영구 토큰 번호(아니면 0)
func ApiTokenID(t *auth.Token) int64 {
	if t == nil || !strings.HasPrefix(t.Etc, ApiTokenPrefix) {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimPrefix(t.Etc, ApiTokenPrefix), 10, 64)
	return n
}

// ApiTokenUsable — 영구 토큰을 쓸 수 있는가: 쓰는 중(status 0) + 주인 번호·사이트가 같음 + 만료 안 됨(expiresAt 0 = 없음). 순수 함수
func ApiTokenUsable(status int, rowAdmin, tokAdmin int64, rowHosting, tokHosting string, expiresAt, now int64) bool {
	return status == 0 && rowAdmin == tokAdmin && rowHosting == tokHosting && (expiresAt == 0 || expiresAt > now)
}

type apiTokCache struct {
	ok  bool
	exp time.Time
}

var (
	apiTokMu    sync.Mutex
	apiTokMemo  = map[int64]apiTokCache{}
	apiTokTouch = map[int64]time.Time{}
)

// ApiTokenForget — 폐기·삭제 때 기억을 바로 지운다
func ApiTokenForget(id int64) {
	apiTokMu.Lock()
	delete(apiTokMemo, id)
	apiTokMu.Unlock()
}

// ApiTokenActive — 이 토큰이 영구 토큰이면 표를 확인(20초 기억, 못 읽으면 막음). 영구 토큰이 아니면 true
func ApiTokenActive(ctx context.Context, a *app.App, t *auth.Token) bool {
	id := ApiTokenID(t)
	if id == 0 {
		return true
	}
	apiTokMu.Lock()
	m, hit := apiTokMemo[id]
	apiTokMu.Unlock()
	if hit && time.Now().Before(m.exp) {
		return m.ok
	}
	var status int
	var admin, exp int64
	var hosting string
	err := a.DB.QueryRowContext(ctx, "SELECT status, adminId, hostingId, expiresAt FROM admin_api_tokens WHERE tokenId = ?", id).Scan(&status, &admin, &hosting, &exp)
	ok := err == nil && ApiTokenUsable(status, admin, t.ID(), hosting, t.HostingID, exp, time.Now().UnixMilli())
	apiTokMu.Lock()
	apiTokMemo[id] = apiTokCache{ok: ok, exp: time.Now().Add(20 * time.Second)}
	last := apiTokTouch[id]
	touch := ok && time.Since(last) > 10*time.Minute
	if touch {
		apiTokTouch[id] = time.Now()
	}
	apiTokMu.Unlock()
	if touch {
		go func() {
			_, _ = a.DB.Exec("UPDATE admin_api_tokens SET lastUsedAt = ? WHERE tokenId = ?", time.Now().UnixMilli(), id)
		}()
	}
	return ok
}
