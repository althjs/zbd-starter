package core

import (
	"context"
	"log"
	"strconv"
	"time"

	"zbd.kr/be/internal/lib/auth"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// 관리자 첫 진입 — 권한 메뉴·프로필·회사 설정·알림. pages/api/admin/common/menus

// siteDefault — SITE 설정이 없는 관리자에게 주는 기본값
func siteDefault() *jsx.OrderedMap {
	return jsx.Obj(
		"hostingId", "",
		"title", "협회 가맹점",
		"serviceType", "TOUR",
		"disableTourAgency", true,
		"feature", jsx.Obj("useMediaImage", true, "pageWidgets", []any{}),
		"disableTourLand", true,
	)
}

// spread — {...v} (객체면 얕은 복사, 배열·문자열은 번호 키, 그 밖은 빈 객체)
func spread(v any) *jsx.OrderedMap {
	switch x := v.(type) {
	case *jsx.OrderedMap:
		return jsx.Assign(nil, x)
	case []any:
		m := jsx.NewOrderedMap()
		for i, e := range x {
			m.Set(strconv.Itoa(i), e)
		}
		return m
	case string:
		m := jsx.NewOrderedMap()
		for i, r := range []rune(x) {
			m.Set(strconv.Itoa(i), string(r))
		}
		return m
	}
	return jsx.NewOrderedMap()
}

// parentHostingOf — 토큰의 상위 호스팅: 자기 SITE.parentHostingId 가 있으면 그것, 아니면 자기 호스팅
func parentHostingOf(c *web.Ctx) string {
	own := c.Token.HostingID
	if site := c.App.Settings.Site(own); site != nil {
		if p := trimStr(jsx.StrOrEmpty(site.V("parentHostingId"))); p != "" && p != own {
			return p
		}
	}
	return own
}

// authorizedMenus — 지금 서비스의 getAuthorizedMenus: {profile, admin:{menus}}. 권한 항목이 없으면 nil.
func authorizedMenus(c *web.Ctx) *jsx.OrderedMap {
	t := c.Token
	if !t.Valid() {
		return jsx.Obj("profile", nil, "admin", nil)
	}
	m := c.App.Admin.Member(t.Key)
	if m == nil {
		return nil
	}
	hostingID := t.HostingID
	menus := c.App.Admin.AuthorizedMenus(hostingID, t.Key)
	var assoc any
	if t.SiteAdmin != 0 {
		assoc = float64(t.SiteAdmin)
	} else if v := c.App.Settings.Hosting(hostingID).Path("server", "associationAdminId"); jsx.Truthy(v) {
		assoc = v
	}
	or := func(v, d any) any {
		if !jsx.Truthy(v) {
			return d
		}
		return v
	}
	p := jsx.Obj(
		"name", or(m.V("name"), ""),
		"nickname", or(m.V("nickname"), ""),
		"email", or(m.V("email"), ""),
		"mobile", or(m.V("mobile"), ""),
		"groupId", m.V("groupId"),
		"adminId", m.V("adminId"),
		"image", or(m.V("image"), jsx.NewOrderedMap()),
		"hostingId", hostingID,
	)
	if jsx.Number(m.V("groupId")) == 1 {
		p.Set("isMaster", true)
	}
	if jsx.Truthy(m.V("isDeveloper")) {
		p.Set("isDeveloper", true)
	}
	if jsx.Truthy(m.V("isSuperAdmin")) {
		p.Set("isSuperAdmin", true)
	}
	if t.PartnerID != 0 {
		p.Set("partnerId", float64(t.PartnerID))
	}
	if assoc != nil {
		p.Set("associationAdminId", assoc)
	}
	if isOfficeReq(c) {
		p.Set("isOffice", true)
	}
	return jsx.Obj("profile", p, "admin", jsx.Obj("menus", menus))
}

// checkLoginDevice — 그 브라우저로 로그인한 기록이 있는가(로그아웃·기기 삭제 뒤 옛 토큰을 막는다)
func checkLoginDevice(c *web.Ctx, q db.Querier, userID any, memberType, email, browserUID, hostingID string) (bool, error) {
	o, err := query(c.Context(), q, `SELECT count(*) as cnt FROM login_devices WHERE userId = ? AND memberType = ? AND email = ? AND browserUid = ? AND hostingId = ?`,
		userID, memberType, c.App.PII.Hash("email", email), browserUID, hostingID)
	if err != nil {
		return false, err
	}
	return len(o) > 0 && jsx.Number(o[0].V("cnt")) > 0, nil
}

// makeToken — 토큰 13칸을 잠근 문자열. ts 가 있으면(다시 발급) 첫 로그인 시각을 두고 갱신 시각을 지금으로.
func makeToken(c *web.Ctx, key, email, memberType, browserUID, hostingID, name, groupID, partnerID, siteAdminID, nickname string, ts int64, subAdminID string) string {
	now := time.Now().UnixMilli()
	first, rts := strconv.FormatInt(now, 10), "0"
	if ts != 0 {
		first, rts = strconv.FormatInt(ts, 10), strconv.FormatInt(now, 10)
	}
	if groupID == "" {
		groupID = "-"
	}
	z := func(s string) string {
		if s == "" {
			return "0"
		}
		return s
	}
	parts := []string{key, email, first, memberType, browserUID, hostingID, name, groupID, z(partnerID), z(siteAdminID), nickname, rts, z(subAdminID)}
	s := ""
	for i, x := range parts {
		if i > 0 {
			s += "|"
		}
		s += x
	}
	return util.AES{Key: c.App.Cfg.AESKey, IV: c.App.Cfg.AESIV}.Encrypt(s)
}

func i64s(n int64) string { return strconv.FormatInt(n, 10) }

// adminNotification — 관리자 알림: 내 1:1 문의(처리된 것)·공지 10개·관리자 푸시 기록, 30일 안, 지운 것 빼고
func adminNotification(ctx context.Context, c *web.Ctx, q db.Querier) (items []any, checked, quick any, userJSON *jsx.OrderedMap, err error) {
	t := c.Token
	rows, err := query(ctx, q, `(
  (SELECT postId, groupKey, groupSubKey, title, json, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, status FROM posts WHERE adminId = ? AND groupKey = "INQUIRY")  
  UNION
  (SELECT postId, groupKey, groupSubKey, title, json, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, status FROM posts WHERE hostingId = "" AND groupKey = "NOTICE" AND status = 0
  ORDER BY postId DESC LIMIT 10)
)
ORDER BY updatedAt desc, postId desc`, t.Key)
	if err != nil {
		return
	}
	var os []*jsx.OrderedMap
	for _, o := range rows {
		if !jsx.Truthy(o.V("json")) {
			continue
		}
		j, e := jsx.ParseString(jsx.Str(o.V("json")))
		if e != nil {
			err = e
			return
		}
		jm, _ := j.(*jsx.OrderedMap)
		o.Set("Json", j)
		o.Delete("json")
		if jm != nil {
			jm.Delete("memo")
		}
		if jsx.Str(o.V("groupKey")) == "INQUIRY" && (jm == nil || !jsx.Truthy(jm.V("processStatus"))) {
			continue
		}
		os = append(os, o)
	}
	a, err := queryOne(ctx, q, `SELECT json FROM admins WHERE adminId = ?`, t.Key)
	if err != nil {
		return
	}
	var aj any
	if a != nil {
		aj = a.V("json")
	}
	j, err := parseJSON(aj, "{}")
	if err != nil {
		return
	}
	userJSON, _ = j.(*jsx.OrderedMap)
	if userJSON == nil {
		userJSON = jsx.NewOrderedMap()
	}
	deleted := userJSON.Arr("deletedNotiIds")
	has := func(s string) bool {
		for _, d := range deleted {
			if ds, ok := d.(string); ok && ds == s {
				return true
			}
		}
		return false
	}
	ymd := util.YMD(time.Now().AddDate(0, 0, -30), false, false)
	items = []any{}
	for _, oo := range os {
		if has(jsx.Str(oo.V("postId"))) {
			continue
		}
		if nid, ok := oo.Get("notiId"); ok && has(jsx.Str(nid)) {
			continue
		} else if !ok && has("undefined") {
			continue
		}
		if ymd < util.YMD(time.UnixMilli(int64(jsx.Number(oo.V("createdAt")))), false, false) {
			if !jsx.Truthy(oo.V("updatedAt")) {
				oo.Set("updatedAt", oo.V("createdAt"))
			}
			items = append(items, oo)
		}
	}
	checked = userJSON.V("checkedNotiIds")
	if !jsx.Truthy(checked) {
		checked = []any{}
	}
	quick = userJSON.V("quickMenuIds")
	if !jsx.Truthy(quick) {
		quick = []any{}
	}
	return
}

// adminCommonMenus — 관리자 첫 진입 정보. 토큰-브라우저·로그인 기기·세션 시간을 확인한다.
func adminCommonMenus(c *web.Ctx) {
	ctx := c.Context()
	t := c.Token
	o := authorizedMenus(c)
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()

	now := time.Now().UnixMilli()
	if t.BrowserUID == auth.AnyBrowser {
		if t.TS+24*60*60*1000 < now {
			c.Ecb("timeout")
			return
		}
	} else if t.BrowserUID != c.BrowserUID() {
		c.Ecb("유효하지 않은 토큰입니다.")
		return
	} else {
		ok, err := checkLoginDevice(c, conn, t.Key, t.MemberType, t.Email, t.BrowserUID, t.HostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		if !ok {
			c.Ecb("유효하지 않은 토큰입니다.")
			return
		}
	}
	hosting := c.App.Settings.Hosting(t.HostingID)
	if t.BrowserUID != auth.AnyBrowser {
		if h := jsx.Number(hosting.Path("server", "sessionTimeoutHourAdmin")); h > 0 {
			uts := t.RTS
			if uts == 0 {
				uts = t.TS
			}
			if float64(uts)+1000*60*60*h < float64(now) {
				log.Printf("@@ 관리자 세션만료 %s adminId: %s", t.HostingID, t.Key)
				c.Ecb("세션이 만료되었습니다. 다시 로그인해 주세요.")
				return
			}
		}
	}
	if o == nil {
		c.Ecb("Cannot set properties of null (setting 'config')")
		return
	}
	profile := o.Map("profile")
	admin := o.Map("admin")

	config := c.App.Settings.Config(t.HostingID)
	cfg := jsx.NewOrderedMap()
	for _, k := range config.Keys() {
		oo := spread(config.V(k))
		if k == "SITE" {
			if v := hosting.Path("server", "associationAdminId"); jsx.Truthy(v) {
				oo.Set("associationAdminId", v)
			}
			if jsx.Truthy(hosting.Path("server", "PG_KSNET", "NONECERT_MID")) && jsx.Truthy(hosting.Path("server", "PG_KSNET", "NONECERT_pgapi")) {
				oo.Set("useKsNetNoneCertPay", true)
			}
		}
		if k == "COMMON_CODES" {
			if myParent := parentHostingOf(c); myParent != "" && myParent != t.HostingID {
				pc, _ := c.App.Settings.Config(myParent).V("COMMON_CODES").(*jsx.OrderedMap)
				var inherit []*jsx.OrderedMap
				for _, x := range pc.Arr("items") {
					if it, _ := x.(*jsx.OrderedMap); it != nil && jsx.StrOrEmpty(it.Path("setting", "parentHostingId")) == myParent {
						inherit = append(inherit, it)
					}
				}
				if len(inherit) > 0 {
					key := func(x *jsx.OrderedMap) string {
						return jsx.StrOrEmpty(x.V("group")) + "|" + jsx.StrOrEmpty(x.V("key")) + "|" + jsx.StrOrEmpty(x.V("name"))
					}
					have := map[string]bool{}
					for _, x := range oo.Arr("items") {
						if it, _ := x.(*jsx.OrderedMap); it != nil {
							have[key(it)] = true
						}
					}
					items := append([]any{}, oo.Arr("items")...)
					added := false
					for _, p := range inherit {
						if have[key(p)] {
							continue
						}
						n := jsx.Assign(nil, p)
						n.Set("id", jsx.NumOr(p.V("id"), 0)*-1-1)
						items = append(items, n)
						added = true
					}
					if added {
						oo.Set("items", items)
					}
				}
			}
		}
		oo.Delete("server")
		if oo.Has("items") && jsx.Truthy(oo.V("items")) {
			arr := []any{}
			for _, x := range oo.Arr("items") {
				ooo := spread(x)
				ooo.Delete("server")
				arr = append(arr, ooo)
			}
			oo.Set("items", arr)
		}
		cfg.Set(k, oo)
	}
	o.Set("config", cfg)

	if !jsx.Truthy(cfg.V("SITE")) {
		a, err := adminGet(c, conn, t.Key, false, false)
		if err != nil {
			c.Ecb(err)
			return
		}
		var x *jsx.OrderedMap
		if a != nil {
			if s, ok := a.V("json").(string); ok {
				x, _ = jsx.ParseOr(s, nil).(*jsx.OrderedMap)
			}
		}
		var site any
		if x != nil && jsx.Truthy(x.V("SITE")) {
			site = x.V("SITE")
		} else if x != nil {
			for _, mi := range x.Arr("META_ITEMS") {
				if mm, _ := mi.(*jsx.OrderedMap); mm != nil && jsx.Str(mm.V("metaKey")) == "SITE" {
					site = mm.V("json")
					break
				}
			}
		}
		if !jsx.Truthy(site) {
			site = siteDefault()
		}
		cfg.Set("SITE", site)
	}

	o.Set("hostingId", t.HostingID)
	o.Set("notification", jsx.Obj("items", []any{}, "checkedNotiIds", []any{}))

	common, _ := c.App.Settings.Config("").V("COMMON").(*jsx.OrderedMap)
	if cm := common.Arr("menus"); len(cm) > 0 {
		menus, _ := admin.V("menus").([]any)
		for _, x := range jsx.Clone(cm).([]any) {
			if mm, _ := x.(*jsx.OrderedMap); mm != nil && jsx.Truthy(mm.V("items")) && t.GroupID == "1" {
				for _, it := range mm.Arr("items") {
					if im, _ := it.(*jsx.OrderedMap); im != nil {
						im.Set("auth", "RW")
					}
				}
			}
			menus = append(menus, x)
		}
		admin.Set("menus", menus)

		if c.Q("check") != "1" {
			email := t.Email
			mt := t.MemberType
			if mt == "" {
				mt = "U"
			}
			if tok := makeToken(c, t.Key, email, mt, t.BrowserUID, t.HostingID, t.Name, t.GroupID, i64s(t.PartnerID), i64s(t.SiteAdmin), t.Nickname, t.TS, i64s(t.SubAdminID)); !c.IssueSession(tok) {
				profile.Set("token", tok)
			}
		}
		items, checked, quick, uj, err := adminNotification(ctx, c, conn)
		if err != nil {
			log.Printf("@ NOTI ERR: %v", err)
		} else {
			o.Set("notification", jsx.Obj("items", items, "checkedNotiIds", checked))
			if jsx.Truthy(uj.V("SITE")) {
				cfg.Set("SITE", uj.V("SITE"))
			}
			if s, ok := uj.V("settings").(*jsx.OrderedMap); ok && s != nil {
				profile.Set("settings", s)
			} else if arr, ok := uj.V("settings").([]any); ok {
				profile.Set("settings", arr)
			}
			profile.Set("quickMenuIds", quick)
		}
	}
	if t.SubAdminID != 0 {
		profile.Set("subAdminId", float64(t.SubAdminID))
	}
	if pcid := parentHostingOf(c); pcid != "" && pcid != jsx.StrOrEmpty(profile.V("hostingId")) {
		profile.Set("parentHostingId", pcid)
	}
	c.Scb(o)
}
