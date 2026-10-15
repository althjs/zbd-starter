package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"zbd.kr/be/internal/lib/settings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// healthCheck — 서버 상태(회사 수·글 수·권한 캐시 수). pages/api/health-check
func healthCheck(c *web.Ctx) {
	start := time.Now()
	o, err := queryOne(c.Context(), c.App.DB, `select count(*) as cnt FROM posts`)
	if err != nil {
		c.Ecb(err)
		return
	}
	var hostingCount any
	if items := c.App.Settings.All().Map("").Map(settings.HostingsKey); items != nil {
		hostingCount = float64(len(items.Arr("items")))
	}
	m := jsx.NewOrderedMap()
	if hostingCount != nil {
		m.Set("hostingCount", hostingCount)
	}
	m.Set("postCount", o.V("cnt"))
	m.Set("duration", durationSec(start))
	m.Set("authCacheCount", float64(c.App.Admin.Get().MemberAuthMap.Len()))
	c.Scb(m)
}

// 관리자 화면 빌드 버전 — 관리자 화면의 vite manifest 의 시작 파일과 수정 시각(배포마다 바뀐다)
var (
	adminVerMu   sync.Mutex
	adminVerAt   time.Time
	adminVerMod  time.Time
	adminVerHash string
	adminVerStr  string
)

// adminManifestPath — 설정 ADMIN_MANIFEST_FILE, 없으면 실행 폴더 옆 관리자 화면 빌드
func adminManifestPath(c *web.Ctx) string {
	if p := c.App.Cfg.Get("ADMIN_MANIFEST_FILE"); p != "" {
		return p
	}
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(wd), "fe-admin", "dist", ".vite", "manifest.json")
}

func currentAdminVersion(c *web.Ctx) *jsx.OrderedMap {
	adminVerMu.Lock()
	defer adminVerMu.Unlock()
	p := adminManifestPath(c)
	if st, err := os.Stat(p); err == nil && (st.ModTime() != adminVerMod || adminVerAt.IsZero()) {
		adminVerMod = st.ModTime()
		adminVerAt = time.Now()
		if b, err := os.ReadFile(p); err == nil {
			var m map[string]struct {
				File string `json:"file"`
			}
			if json.Unmarshal(b, &m) == nil {
				adminVerHash = m["index.html"].File
			}
		}
		adminVerStr = st.ModTime().In(util.KST).Format("20060102.1504")
	}
	return jsx.Obj("version", adminVerStr, "hash", adminVerHash)
}

// adminVersion — GET /admin/version: 관리자 화면 빌드 버전(새 버전 안내용)
func adminVersion(c *web.Ctx) {
	if c.Method != "GET" {
		c.Status(405).Send(jsx.Obj("error", "Method "+c.Method+" not allowed"))
		return
	}
	c.SetHeader("Cache-Control", "no-store")
	c.Scb(currentAdminVersion(c))
}

// adminCommonMeta — 공통 설정(COMMON)
func adminCommonMeta(c *web.Ctx) {
	if !c.Token.Valid() {
		c.Ecb("잘못된 요청")
		return
	}
	c.Scb(c.App.Settings.Config("").V("COMMON"))
}
