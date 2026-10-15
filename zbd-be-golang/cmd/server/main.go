// 즈비드 BE 서버 — HTTP API (그리고 앞으로 실시간) 를 **한 프로세스**에서.
//
// Go 는 프로세스 하나가 CPU 코어를 모두 쓰고 메모리를 공유한다 — 캐시·DB 풀이 한 벌이다
// (프로세스마다 따로 들던 구조가 아니다). 그래서 여러 개를 띄우지 않는다.
//
// 개발: `make dev` · 설정은 .env.local(항목은 .env.example).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// kstLog — 로그 줄 앞 시각을 KST 로(서버 시계는 UTC). 시간대(time.Local) 는 바꾸지 않는다 — 날짜 계산이 어긋날 수 있다(CLAUDE.md)
type kstLog struct{}

func (kstLog) Write(p []byte) (int, error) {
	return os.Stderr.Write(append([]byte(time.Now().In(util.KST).Format("2006-01-02 15:04:05 ")), p...))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(kstLog{})
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer a.DB.Close()

	r := web.NewRouter(a)
	registerRoutes(r)
	go a.Run(ctx)

	var h http.Handler = r
	// 올린 파일(/upload/…) — 운영은 nginx 가 CDN_UPLOAD_DIR 를 바로 내보낸다. 이 서버도 내보내 개발 모드(nginx 없음)에서 그림이 보이게
	if dir := a.Cfg.Get("CDN_UPLOAD_DIR"); dir != "" {
		files := http.StripPrefix("/upload/", http.FileServer(http.Dir(dir)))
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/upload/") && (req.Method == http.MethodGet || req.Method == http.MethodHead) {
				if strings.HasSuffix(req.URL.Path, "/") { // 폴더 목록은 보이지 않게
					http.NotFound(w, req)
					return
				}
				files.ServeHTTP(w, req)
				return
			}
			r.ServeHTTP(w, req)
		})
	}
	srv := &http.Server{Addr: ":" + a.Cfg.Port, Handler: logRequests(h), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("즈비드 BE :%s (DB %s:%d/%s · 라우트 %d개)", a.Cfg.Port, a.Cfg.DB.Host, a.Cfg.DB.Port, a.Cfg.DB.Database, len(r.Routes()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	// 받은 요청은 끝까지 처리하고 내려간다(배포 때 끊기지 않게)
	shutdownCtx, c2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer c2()
	_ = srv.Shutdown(shutdownCtx)
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s (%dms)", r.Method, r.URL.Path, time.Since(t0).Milliseconds())
	})
}
