package core

// 즈비드 플러그인 라이선스 확인 — 관리자 화면 맨 위 배너
//
//	`zbd add`·`zbd update` 로 붙인 플러그인이 있으면 zbd 명령이 zz_plugins_gen.go 를 만들어
//	pluginLicMark(꾸러미를 받은 회원의 구매자 표시)·pluginLicInstalled(플러그인 키 → 판)를 적는다.
//	서버(RUN_JOBS=true)는 켠 지 1분 뒤, 그 뒤로는 즈비드가 알려 준 간격(기본 6시간)마다
//	https://zbd.kr/service/plugin/site-check 에 {표시, 사이트 주소, 플러그인 목록}을 보내고 답을 기억한다.
//	문제가 있으면 GET /admin/plugin-license 가 돌려주고 관리자 화면이 배너로 보인다. 사이트는 멈추지 않는다.
//	즈비드에 닿지 않으면 앞 결과를 그대로 둔다(처음이면 배너 없음). 보내는 것: 표시·사이트 주소·플러그인 키와 판(회원·글 데이터는 없음).
//	주소를 바꾸려면 .env ZBD_LICENSE_URL. 판정 규칙(순수 함수)은 plugin_license_policy_test.go.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// zbd 명령이 zz_plugins_gen.go 에서 채운다(없으면 확인하지 않는다)
var (
	pluginLicMark      string
	pluginLicInstalled = map[string]string{}
)

type pluginLicProblem struct {
	Code    string `json:"code"`
	Plugin  string `json:"plugin,omitempty"`
	Message string `json:"message"`
}

var pluginLic = struct {
	sync.Mutex
	next     time.Time
	at       int64
	problems []pluginLicProblem
	checkMin int
}{}

const pluginLicDefaultURL = "https://zbd.kr/service/plugin/site-check"

// pluginLicNext — 다음에 물을 때. 성공이면 checkMin 분 뒤(없거나 이상하면 6시간, 최소 30분·최대 1일), 실패면 30분 뒤. 순수 함수
func pluginLicNext(now time.Time, ok bool, checkMin int) time.Time {
	if !ok {
		return now.Add(30 * time.Minute)
	}
	if checkMin <= 0 {
		checkMin = 360
	}
	if checkMin < 30 {
		checkMin = 30
	}
	if checkMin > 1440 {
		checkMin = 1440
	}
	return now.Add(time.Duration(checkMin) * time.Minute)
}

// pluginLicParse — 즈비드 답({code:"0000", data:{ok, problems, checkMin}}) 해석. ok=false 면 답을 못 믿는다(앞 결과 유지). 순수 함수
func pluginLicParse(b []byte) (problems []pluginLicProblem, checkMin int, ok bool) {
	var r struct {
		Code string `json:"code"`
		Data struct {
			Problems []pluginLicProblem `json:"problems"`
			CheckMin int                `json:"checkMin"`
		} `json:"data"`
	}
	if json.Unmarshal(b, &r) != nil || r.Code != "0000" {
		return nil, 0, false
	}
	out := []pluginLicProblem{}
	for _, p := range r.Data.Problems {
		if p.Code == "" || len(out) >= 20 {
			continue
		}
		if len(p.Message) > 300 {
			p.Message = p.Message[:300]
		}
		out = append(out, p)
	}
	return out, r.Data.CheckMin, true
}

// pluginLicHost — 보고할 사이트 주소(사이트 설정의 사용자 주소 첫 번째)
func pluginLicHost(a *app.App) string {
	for _, x := range a.Settings.Hostings() {
		m, _ := x.(*jsx.OrderedMap)
		if m == nil {
			continue
		}
		id := jsx.StrOrEmpty(m.V("hostingId"))
		if id == "" {
			continue
		}
		if d := web.SiteDomains(a.Settings.Site(id), a.Cfg.Get("ADMIN_HOSTS")); len(d.User) > 0 {
			return d.User[0]
		}
	}
	return ""
}

var pluginLicHTTP = &http.Client{Timeout: 20 * time.Second}

// pluginLicTick — 1분 작업. 때가 되면 한 번 묻는다
func pluginLicTick(ctx context.Context, a *app.App) {
	if len(pluginLicInstalled) == 0 {
		return
	}
	now := time.Now()
	pluginLic.Lock()
	due := !now.Before(pluginLic.next)
	pluginLic.Unlock()
	if !due {
		return
	}
	url := a.Cfg.Get("ZBD_LICENSE_URL")
	if url == "" {
		url = pluginLicDefaultURL
	}
	body, _ := json.Marshal(map[string]any{"mark": pluginLicMark, "host": pluginLicHost(a), "plugins": pluginLicInstalled})
	ok := false
	var problems []pluginLicProblem
	checkMin := 0
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		if res, err := pluginLicHTTP.Do(req); err == nil {
			b, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
			res.Body.Close()
			problems, checkMin, ok = pluginLicParse(b)
		}
	}
	pluginLic.Lock()
	defer pluginLic.Unlock()
	pluginLic.next = pluginLicNext(now, ok, checkMin)
	if !ok {
		log.Printf("[plugin-license] 즈비드에 닿지 않음 — 앞 결과 유지, 30분 뒤 다시")
		return
	}
	pluginLic.at, pluginLic.problems, pluginLic.checkMin = now.UnixMilli(), problems, checkMin
	if len(problems) > 0 {
		log.Printf("[plugin-license] 확인 결과 문제 %d개 — 관리자 화면에 배너", len(problems))
	}
}

// adminPluginLicense — GET /admin/plugin-license (관리자 화면 배너)
func adminPluginLicense(c *web.Ctx) {
	pluginLic.Lock()
	ps := append([]pluginLicProblem{}, pluginLic.problems...)
	at := pluginLic.at
	pluginLic.Unlock()
	items := make([]any, 0, len(ps))
	for _, p := range ps {
		items = append(items, jsx.Obj("code", p.Code, "plugin", p.Plugin, "message", p.Message))
	}
	c.Scb(jsx.Obj("checkedAt", at, "problems", items, "plugins", len(pluginLicInstalled)))
}

func registerPluginLicense(r *web.Router) {
	r.Handle("/admin/plugin-license", web.Auth, adminPluginLicense)
	a := r.App
	pluginLic.Lock()
	pluginLic.next = time.Now() // 켠 뒤 첫 1분 작업에서
	pluginLic.Unlock()
	r.App.EveryMinute("플러그인 라이선스 확인", func(ctx context.Context) { pluginLicTick(ctx, a) })
}
