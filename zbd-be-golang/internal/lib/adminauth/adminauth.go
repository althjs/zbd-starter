// Package adminauth — 관리자 권한 캐시: 메뉴 트리 · 메뉴별 API · 관리자별 메뉴 권한(R/RW).
//
// 1분마다 다시 읽고, 관리자를 더하거나 권한을 바꾸면 소켓 신호(onUpdateCodes)로 바로 다시 읽는다.
// 결과 모양(키 순서 포함)은 지금 서비스의 캐시와 같다 — 관리자 메뉴 응답이 이 값을 그대로 쓴다.
//
//	menus          메뉴 트리(1단 + items), sortNo 오름차순
//	hostingApisMap {hostingId: {api: [menuId…]}} — 그 API 를 쓰려면 그 메뉴들 중 하나의 권한이 있어야 한다
//	apisMap        호스팅 없는("") 메뉴의 API
//	memberAuthMap  {adminId: {adminId, groupId, hostingId, name, nickname, email, mobile, image, authMenuMap, isDeveloper?, isSuperAdmin?}}
package adminauth

import (
	"context"
	"database/sql"
	"log"
	"sort"
	"strings"
	"sync"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

type Snapshot struct {
	Menus          []*jsx.OrderedMap
	ApisMap        *jsx.OrderedMap // api → []any(menuId)
	HostingApisMap *jsx.OrderedMap // hostingId → (api → []any)
	MemberAuthMap  *jsx.OrderedMap // adminId(문자열) → 항목
}

type Cache struct {
	mu   sync.RWMutex
	snap *Snapshot
}

func New() *Cache {
	return &Cache{snap: &Snapshot{ApisMap: jsx.NewOrderedMap(), HostingApisMap: jsx.NewOrderedMap(), MemberAuthMap: jsx.NewOrderedMap()}}
}

// Get — 지금 캐시(읽기 전용으로 쓴다 — 고치려면 jsx.Clone)
func (c *Cache) Get() *Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snap
}

// Reload — DB 에서 다시 읽는다. 실패하면 기존 캐시를 그대로 둔다.
func (c *Cache) Reload(ctx context.Context, d *sql.DB) error {
	s, err := build(ctx, d)
	if err != nil {
		log.Printf("@@ adminauth 캐시 갱신 실패: %v", err)
		return err
	}
	c.mu.Lock()
	c.snap = s
	c.mu.Unlock()
	return nil
}

// Member — 관리자 한 명의 권한 항목(없으면 nil)
func (c *Cache) Member(adminID string) *jsx.OrderedMap {
	return c.Get().MemberAuthMap.Map(adminID)
}

func build(ctx context.Context, d *sql.DB) (*Snapshot, error) {
	rows, err := db.Query(ctx, d, `select a.menuId, parentId, sortNo, name, description, icon, path, special,
IFNULL(b.apis, "[]") as apis, a.hostingId from admin_menus a
LEFT JOIN (
	select menuId, hostingId, JSON_ARRAYAGG(api) AS apis from admin_menu_apis group by menuId, hostingId
) as b
ON a.menuId = b.menuId AND a.hostingId = b.hostingId`)
	if err != nil {
		return nil, err
	}
	menus, apisMap, hostingApisMap := MakeMenuTree(rows)

	ga, err := db.Query(ctx, d, `select groupId, menuId, auth from admin_role_menus WHERE auth <> ''`)
	if err != nil {
		return nil, err
	}
	groupAuth := map[string]*jsx.OrderedMap{}
	for _, r := range ga {
		g := jsx.Str(r.V("groupId"))
		if groupAuth[g] == nil {
			groupAuth[g] = jsx.NewOrderedMap()
		}
		groupAuth[g].Set(jsx.Str(r.V("menuId")), r.V("auth"))
	}

	ad, err := db.Query(ctx, d, `SELECT adminId, groupId, name, email, mobile, json, hostingId FROM admins WHERE status = 0`)
	if err != nil {
		return nil, err
	}
	mam := jsx.NewOrderedMap()
	for _, r := range ad {
		id := jsx.Str(r.V("adminId"))
		if mam.Has(id) {
			continue
		}
		gm := jsx.Clone(groupAuth[jsx.Str(r.V("groupId"))]).(*jsx.OrderedMap)
		if gm == nil {
			gm = jsx.NewOrderedMap()
		}
		J := jsx.ParseMap(jsx.StrOrEmpty(r.V("json")))
		authMenuMap := jsx.NewOrderedMap()
		for _, x := range J.Arr("authMenus") {
			am, _ := x.(*jsx.OrderedMap)
			if am == nil {
				continue
			}
			mid := jsx.Str(am.V("menuId"))
			if !jsx.Truthy(gm.V(mid)) {
				authMenuMap.Set(mid, am.V("auth"))
			} else if jsx.Str(gm.V(mid)) == "RW" {
				authMenuMap.Set(mid, gm.V(mid))
			} else {
				authMenuMap.Set(mid, am.V("auth"))
			}
		}
		for _, mid := range gm.Keys() {
			if !jsx.Truthy(authMenuMap.V(mid)) {
				authMenuMap.Set(mid, gm.V(mid))
			} else if jsx.Str(gm.V(mid)) == "RW" {
				authMenuMap.Set(mid, gm.V(mid))
			}
		}
		e := jsx.NewOrderedMap()
		e.Set("adminId", r.V("adminId"))
		e.Set("groupId", r.V("groupId"))
		e.Set("hostingId", r.V("hostingId"))
		e.Set("name", r.V("name"))
		nick := J.V("nickname")
		if !jsx.Truthy(nick) {
			nick = ""
		}
		e.Set("nickname", nick)
		e.Set("email", r.V("email"))
		e.Set("mobile", r.V("mobile"))
		img := J.V("image")
		if !jsx.Truthy(img) {
			img = jsx.NewOrderedMap()
		}
		e.Set("image", img)
		e.Set("authMenuMap", authMenuMap)
		if jsx.Truthy(J.V("isDeveloper")) {
			e.Set("isDeveloper", true)
		}
		if jsx.Truthy(J.V("isSuperAdmin")) {
			e.Set("isSuperAdmin", true)
		}
		mam.Set(id, e)
	}
	return &Snapshot{Menus: menus, ApisMap: apisMap, HostingApisMap: hostingApisMap, MemberAuthMap: mam}, nil
}

// MakeMenuTree — 메뉴 행 → 트리(1단 + items) · API 맵. 행은 고쳐진다(apis 를 배열로, api_count·items 추가).
func MakeMenuTree(rows []*jsx.OrderedMap) (menus []*jsx.OrderedMap, apisMap, hostingApisMap *jsx.OrderedMap) {
	apisMap = jsx.NewOrderedMap()
	hostingApisMap = jsx.NewOrderedMap()
	m := jsx.NewOrderedMap() // 키: menuId+hostingId (숫자만인 키는 JS 처럼 앞·오름차순)
	for _, o := range rows {
		var apis []any
		switch x := o.V("apis").(type) {
		case string:
			if p, ok := jsx.ParseOr(x, []any{}).([]any); ok {
				apis = p
			}
		case []any:
			apis = x
		}
		if apis == nil {
			apis = []any{}
		}
		o.Set("apis", apis)
		cid := jsx.StrOrEmpty(o.V("hostingId"))
		for _, a := range apis {
			api := jsx.Str(a)
			if cid != "" {
				cm := hostingApisMap.Map(cid)
				if cm == nil {
					cm = jsx.NewOrderedMap()
					hostingApisMap.Set(cid, cm)
				}
				arr, _ := cm.V(api).([]any)
				cm.Set(api, append(arr, o.V("menuId")))
			} else {
				arr, _ := apisMap.V(api).([]any)
				apisMap.Set(api, append(arr, o.V("menuId")))
			}
		}
		if !jsx.Truthy(o.V("api_count")) {
			o.Set("api_count", float64(len(apis)))
		}
		if jsx.Truthy(o.V("parentId")) {
			key := jsx.Str(o.V("parentId")) + cid
			p := m.Map(key)
			if p == nil {
				p = jsx.NewOrderedMap()
				p.Set("items", []any{})
				m.Set(key, p)
			} else if !p.Has("items") {
				p.Set("items", []any{})
			}
			p.Set("items", append(p.Arr("items"), o))
		} else {
			key := jsx.Str(o.V("menuId")) + cid
			if pre := m.Map(key); pre != nil && len(pre.Arr("items")) > 0 {
				o.Set("items", append(append([]any{}, o.Arr("items")...), pre.Arr("items")...))
			}
			m.Set(key, o)
		}
	}
	for _, k := range m.Keys() { // Object.values — JS 키 순서
		o := m.Map(k)
		if items := o.Arr("items"); o.Has("items") {
			o.Set("items", orderBySortNo(items))
		}
		menus = append(menus, o)
	}
	sortMaps(menus)
	return menus, apisMap, hostingApisMap
}

// lodash orderBy(arr, ["sortNo"], ["asc"]) — 안정 정렬, 값이 없으면 뒤로
func less(a, b *jsx.OrderedMap) bool {
	av, bv := a.V("sortNo"), b.V("sortNo")
	if av == nil || bv == nil {
		return av != nil && bv == nil
	}
	return jsx.Number(av) < jsx.Number(bv)
}

func sortMaps(arr []*jsx.OrderedMap) {
	sort.SliceStable(arr, func(i, j int) bool { return less(arr[i], arr[j]) })
}

func orderBySortNo(items []any) []any {
	ms := make([]*jsx.OrderedMap, 0, len(items))
	for _, it := range items {
		if mm, ok := it.(*jsx.OrderedMap); ok {
			ms = append(ms, mm)
		}
	}
	sortMaps(ms)
	out := make([]any, len(ms))
	for i, x := range ms {
		out[i] = x
	}
	return out
}

// Allowed — 관리자 API 권한 검사.
//
//	routePath 는 경로 정의(예 "/service/admin/design/[designId]"). 메뉴에 묶인 API 중 그 앞부분과 맞는 것이 있으면
//	**가장 긴(가장 구체적인) 것 하나**의 메뉴들 중 하나에 권한이 있어야 한다 — GET 은 R 이상, 그 밖의 메서드는 RW.
//	(예전엔 맞는 것을 차례로 보며 마지막 것이 결과를 정해, /admin/member 와 /admin/member/login-devices 가 둘 다
//	묶이면 순서에 따라 결과가 바뀌었다.)
//	「W:」 로 시작하는 API 는 쓰기(GET 아닌 요청)에만 — 업무 보드 담당자 이름처럼 여러 화면이 같이 읽는 목록용.
//	아무 메뉴에도 묶이지 않은 API 는 통과.
func (c *Cache) Allowed(hostingID, adminID, method, routePath string) bool {
	return c.allowed(hostingID, adminID, method, routePath, false)
}

// CanRead — 쓰기 전용(W:) 묶음까지 읽기 권한으로 본다: 그 묶음의 메뉴에 R 이상이 있는지.
// 묶인 목록을 누구나 읽게 열어 둔 API 에서 「그 메뉴 사람」 에게만 칸을 더 줄 때(관리자 목록의 이메일 등).
func (c *Cache) CanRead(hostingID, adminID, routePath string) bool {
	return c.allowed(hostingID, adminID, "GET", routePath, true)
}

func (c *Cache) allowed(hostingID, adminID, method, routePath string, strictRead bool) bool {
	s := c.Get()
	var apis *jsx.OrderedMap
	if hostingID != "" {
		apis = s.HostingApisMap.Map(hostingID)
	} else {
		apis = s.ApisMap
	}
	member := s.MemberAuthMap.Map(adminID)
	binds := map[string][]string{}
	for _, api := range apis.Keys() {
		for _, mid := range apis.Arr(api) {
			binds[api] = append(binds[api], jsx.Str(mid))
		}
	}
	return allowedBy(binds, func(menuID string) string {
		if member == nil {
			return ""
		}
		return jsx.StrOrEmpty(member.Map("authMenuMap").V(menuID))
	}, method, routePath, strictRead)
}

// AllowedBy — Allowed 의 판정(순수 함수, 정책 테스트용). binds = {API: [menuId…]}, auth(menuId) = "" | "R" | "RW".
func AllowedBy(binds map[string][]string, auth func(menuID string) string, method, routePath string) bool {
	return allowedBy(binds, auth, method, routePath, false)
}

func allowedBy(binds map[string][]string, auth func(menuID string) string, method, routePath string, strictRead bool) bool {
	best, bestLen := "", -1
	for api := range binds {
		path, writeOnly := strings.CutPrefix(api, "W:")
		if writeOnly && method == "GET" && !strictRead {
			continue
		}
		if path == "" || !strings.HasPrefix(routePath, path) {
			continue
		}
		// 같은 길이면 전체(읽기 포함) 묶음이 이긴다 — 더 엄격한 쪽
		if len(path) > bestLen || (len(path) == bestLen && !writeOnly) {
			best, bestLen = api, len(path)
		}
	}
	if bestLen < 0 {
		return true
	}
	for _, mid := range binds[best] {
		a := auth(mid)
		if a == "" {
			continue
		}
		if method == "GET" || a == "RW" {
			return true
		}
	}
	return false
}

// AuthorizedMenus — 그 관리자가 볼 수 있는 메뉴(지금 서비스의 _getAuthorizedMenus).
func (c *Cache) AuthorizedMenus(hostingID, adminID string) []any {
	s := c.Get()
	ua := s.MemberAuthMap.Map(adminID).Map("authMenuMap")
	out := []any{}
	for _, o := range s.Menus {
		has := jsx.Truthy(ua.V(jsx.Str(o.V("menuId"))))
		if !has {
			for _, it := range o.Arr("items") {
				if oo, _ := it.(*jsx.OrderedMap); oo != nil && jsx.Truthy(ua.V(jsx.Str(oo.V("menuId")))) {
					has = true
					break
				}
			}
		}
		if !has {
			continue
		}
		items := []any{}
		for _, it := range o.Arr("items") {
			oo, _ := it.(*jsx.OrderedMap)
			if oo == nil || jsx.StrOrEmpty(oo.V("hostingId")) != hostingID || !jsx.Truthy(ua.V(jsx.Str(oo.V("menuId")))) {
				continue
			}
			x := jsx.NewOrderedMap()
			x.Set("menuId", oo.V("menuId"))
			x.Set("name", oo.V("name"))
			x.Set("description", oo.V("description"))
			x.Set("icon", oo.V("icon"))
			x.Set("path", oo.V("path"))
			if jsx.Truthy(oo.V("special")) {
				x.Set("special", oo.V("special"))
			}
			x.Set("auth", ua.V(jsx.Str(oo.V("menuId"))))
			x.Set("hostingId", oo.V("hostingId"))
			items = append(items, x)
		}
		if jsx.StrOrEmpty(o.V("hostingId")) == hostingID {
			x := jsx.NewOrderedMap()
			x.Set("menuId", o.V("menuId"))
			x.Set("name", o.V("name"))
			x.Set("description", o.V("description"))
			x.Set("icon", o.V("icon"))
			x.Set("path", o.V("path"))
			if jsx.Truthy(o.V("special")) {
				x.Set("special", o.V("special"))
			}
			x.Set("items", items)
			// 특수 메뉴(게시판 목록 등)는 화면이 하위 메뉴를 채우고 그 권한을 대메뉴에서 물려받는다 — 실제 권한을 싣는다
			auth := any("R")
			if jsx.Truthy(o.V("special")) && jsx.Truthy(ua.V(jsx.Str(o.V("menuId")))) {
				auth = ua.V(jsx.Str(o.V("menuId")))
			}
			x.Set("auth", auth)
			x.Set("hostingId", o.V("hostingId"))
			out = append(out, x)
		}
	}
	return out
}

// MenuAuth — 그 관리자가 메뉴 주소(path)에 가진 권한 "" | "R" | "RW"(2026-10-09 대표 — 서버 관리 화면을 메뉴 권한으로 열기).
// 같은 주소 메뉴가 여럿이면 높은 쪽. 호스팅이 다른 메뉴는 안 본다.
func (c *Cache) MenuAuth(hostingID, adminID, path string) string {
	s := c.Get()
	ua := s.MemberAuthMap.Map(adminID).Map("authMenuMap")
	return MenuAuthBy(s.Menus, func(menuID string) string { return jsx.StrOrEmpty(ua.V(menuID)) }, hostingID, path)
}

// MenuAuthBy — MenuAuth 의 판정(순수 함수, 정책 테스트용). menus = 메뉴 트리(1단 + items)
func MenuAuthBy(menus []*jsx.OrderedMap, auth func(menuID string) string, hostingID, path string) string {
	best := ""
	see := func(o *jsx.OrderedMap) {
		if o == nil || path == "" || jsx.StrOrEmpty(o.V("path")) != path || jsx.StrOrEmpty(o.V("hostingId")) != hostingID {
			return
		}
		switch auth(jsx.Str(o.V("menuId"))) {
		case "RW":
			best = "RW"
		case "R":
			if best == "" {
				best = "R"
			}
		}
	}
	for _, o := range menus {
		see(o)
		for _, it := range o.Arr("items") {
			oo, _ := it.(*jsx.OrderedMap)
			see(oo)
		}
	}
	return best
}
