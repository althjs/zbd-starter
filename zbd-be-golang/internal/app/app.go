// Package app — 프로세스 하나가 공유하는 것: 설정 · DB 풀 · 캐시(설정·관리자 권한) · 개인정보 키 · 실시간 통로.
//
// Go 는 프로세스 하나가 모든 요청을 받는다 — 캐시도 한 벌이다(프로세스마다 따로 들던 구조가 아니다).
// 캐시는 1분마다 다시 읽고, 관리자 화면에서 설정·권한을 바꾸면 그 요청에서 바로 다시 읽는다(web.Options{Changes}).
//
// 실시간 통로(Realtime) 는 스타터 기본엔 없다 — Emit 은 아무것도 안 하고 화면은 새로고침해야 반영된다.
// 웹소켓 연동 플러그인이 이 패키지에 파일 하나를 더해 realtimeFactory 를 채운다.
package app

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/adminauth"
	"zbd.kr/be/internal/lib/config"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/pii"
	"zbd.kr/be/internal/lib/settings"
)

type App struct {
	Cfg      *config.Config
	DB       *sql.DB
	Settings *settings.Cache
	Admin    *adminauth.Cache
	PII      *pii.Keys // 키가 없으면 nil — 암호화 시 오류
	Realtime Realtime  // 웹소켓 연동 플러그인이 있을 때만 — 없으면 nil
	Started  time.Time

	jobsMu sync.Mutex
	jobs   []job
}

type job struct {
	name string
	fn   func(ctx context.Context)
}

// Realtime — 실시간 통로. Emit = 소켓 서버로 신호(doSocketDataUpdate 등), Status = 접속 현황("ready" 대기열 · "users" 사용자 소켓, 없으면 nil)
type Realtime interface {
	Emit(ev string, data any)
	Status(kind string) *jsx.OrderedMap
}

// realtimeFactory — 플러그인이 init() 에서 채운다(스타터 기본은 nil)
var realtimeFactory func(a *App) Realtime

// New — 설정을 읽고 DB·캐시·(있으면) 실시간 통로를 준비한다.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	d, err := db.Open(cfg.DB)
	if err != nil {
		return nil, err
	}
	a := &App{Cfg: cfg, DB: d, Settings: settings.New(), Admin: adminauth.New(), Started: time.Now()}
	if k, err := pii.New(cfg.Get("PII_ENC_KEY_V1"), cfg.Get("PII_INDEX_KEY")); err == nil {
		a.PII = k
	} else {
		log.Printf("⚠ 개인정보 키 없음 — 회원 이름·이메일을 읽고 쓸 수 없다: %v", err)
	}
	a.ReloadCaches(ctx)

	if realtimeFactory != nil {
		a.Realtime = realtimeFactory(a)
	}
	return a, nil
}

// ReloadCaches — 설정·관리자 권한 캐시를 다시 읽는다.
func (a *App) ReloadCaches(ctx context.Context) {
	_ = a.Settings.Reload(ctx, a.DB)
	_ = a.Admin.Reload(ctx, a.DB)
}

// Emit — 실시간 통로로 신호. 통로가 없으면(스타터 기본) 아무것도 안 한다.
func (a *App) Emit(ev string, data any) {
	if a.Realtime != nil {
		a.Realtime.Emit(ev, data)
	}
}

// RealtimeStatus — 접속 현황(통로가 없으면 nil)
func (a *App) RealtimeStatus(kind string) *jsx.OrderedMap {
	if a.Realtime == nil {
		return nil
	}
	return a.Realtime.Status(kind)
}

// EveryMinute — 1분마다 할 일(알림 예약 발송·오래된 기록 정리 등). 설정 RUN_JOBS=true 일 때만 돈다(개발·시험 서버에선 안 돈다).
func (a *App) EveryMinute(name string, fn func(ctx context.Context)) {
	a.jobsMu.Lock()
	a.jobs = append(a.jobs, job{name, fn})
	a.jobsMu.Unlock()
}

// Run — 1분 하트비트(캐시 갱신 + 등록된 일). ctx 가 끝나면 멈춘다.
func (a *App) Run(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	runJobs := a.Cfg.Bool("RUN_JOBS")
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c, cancel := context.WithTimeout(ctx, 50*time.Second)
			a.ReloadCaches(c)
			if runJobs {
				a.jobsMu.Lock()
				js := append([]job(nil), a.jobs...)
				a.jobsMu.Unlock()
				for _, j := range js {
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("@@ %s 실패: %v", j.name, r)
							}
						}()
						j.fn(c)
					}()
				}
			}
			cancel()
		}
	}
}
