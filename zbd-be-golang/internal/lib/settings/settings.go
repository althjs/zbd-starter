// Package settings — 설정(settings 표) 캐시. 회사별 {metaKey: json}.
//
// 1분마다·소켓 신호(onUpdateCodes) 때 다시 읽는다. 값 모양·키 순서는 지금 서비스의 메타 캐시와 같다.
// 회사 공통("")의 COMMON 은 각 회사 설정에 얹어 준다(지금 서비스가 getConfig 때 하는 일).
package settings

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

type Cache struct {
	mu   sync.RWMutex
	data *jsx.OrderedMap // hostingId → (metaKey → json)
}

// HostingsKey — 호스팅 목록 설정의 키(공통 "" 호스팅)
const HostingsKey = "HOSTINGS"

// 옛 이름 companyId — 2026-10 옮기는 기간: 목록 설정의 옛 키
const oldHostingsKey = "COMPANIES"

func New() *Cache { return &Cache{data: jsx.NewOrderedMap()} }

// Reload — DB 에서 다시 읽는다. 실패하면 기존 캐시를 둔다.
func (c *Cache) Reload(ctx context.Context, d *sql.DB) error {
	rows, err := db.Query(ctx, d, `SELECT metaKey, name, memo, json, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId from settings`)
	if err != nil {
		log.Printf("@@ settings 캐시 갱신 실패: %v", err)
		return err
	}
	m := jsx.NewOrderedMap()
	for _, r := range rows {
		cid := jsx.StrOrEmpty(r.V("hostingId"))
		cm := m.Map(cid)
		if cm == nil {
			cm = jsx.NewOrderedMap()
			m.Set(cid, cm)
		}
		cm.Set(jsx.Str(r.V("metaKey")), jsx.ParseOr(jsx.StrOrEmpty(r.V("json")), nil))
	}
	migrateOldNames(m)
	fixup(m, HostingsKey)
	c.mu.Lock()
	c.data = m
	c.mu.Unlock()
	return nil
}

// migrateOldNames — 옛 이름 companyId — 2026-10 옮기는 기간.
// 공통("") 의 HOSTINGS 가 없고 옛 키 COMPANIES 가 있으면 그것을 HOSTINGS 로 읽고,
// 목록 항목에 hostingId 가 없고 companyId 가 있으면 hostingId 로 바꿔 읽는다(캐시만 — DB 는 그대로).
func migrateOldNames(m *jsx.OrderedMap) {
	common := m.Map("")
	if common == nil {
		return
	}
	if !common.Has(HostingsKey) && common.Has(oldHostingsKey) {
		common.Set(HostingsKey, common.V(oldHostingsKey))
		common.Delete(oldHostingsKey)
	}
	for _, x := range common.Map(HostingsKey).Arr("items") {
		if it, _ := x.(*jsx.OrderedMap); it != nil && !it.Has("hostingId") && it.Has("companyId") {
			it.RenameKey("companyId", "hostingId")
		}
	}
}

// fixup — 지금 서비스의 getConfig 가 캐시 객체에 얹는 값을 미리 얹는다
//
//	· COMMON = 회사 공통("") 의 COMMON
//	· SITE.smsSenderNumber / SITE.noreplyMail 이 비었으면 회사 목록의 server.nhn.sendNo / server.noreplyMail
func fixup(m *jsx.OrderedMap, hostingsKey string) {
	common := m.Map("").V("COMMON")
	if common == nil {
		common = jsx.NewOrderedMap()
	}
	hostings := m.Map("").Map(hostingsKey).Arr("items")
	find := func(cid string) *jsx.OrderedMap {
		for _, x := range hostings {
			if c, _ := x.(*jsx.OrderedMap); c != nil && jsx.StrOrEmpty(c.V("hostingId")) == cid {
				return c
			}
		}
		return nil
	}
	for _, cid := range m.Keys() {
		cm := m.Map(cid)
		cm.Set("COMMON", common)
		site := cm.Map("SITE")
		if site == nil {
			continue
		}
		co := find(cid)
		if !jsx.Truthy(site.V("smsSenderNumber")) {
			v := co.Path("server", "nhn", "sendNo")
			if !jsx.Truthy(v) {
				v = ""
			}
			site.Set("smsSenderNumber", v)
		}
		if !jsx.Truthy(site.V("noreplyMail")) {
			v := co.Path("server", "noreplyMail")
			if !jsx.Truthy(v) {
				v = ""
			}
			site.Set("noreplyMail", v)
		}
	}
}

// All — 캐시 전체(hostingId → 설정). 읽기 전용으로 쓴다.
func (c *Cache) All() *jsx.OrderedMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

// Config — 회사 설정 {SITE, COMMON_CODES, POST_GROUPS, …, COMMON}. 없는 회사면 {COMMON} 만.
func (c *Cache) Config(hostingID string) *jsx.OrderedMap {
	all := c.All()
	if m := all.Map(hostingID); m != nil {
		return m
	}
	n := jsx.NewOrderedMap()
	common := all.Map("").V("COMMON")
	if common == nil {
		common = jsx.NewOrderedMap()
	}
	n.Set("COMMON", common)
	return n
}

// Meta — 회사의 설정 하나(없으면 nil)
func (c *Cache) Meta(hostingID, metaKey string) any {
	return c.All().Map(hostingID).V(metaKey)
}

// Site — 회사 SITE 설정(없으면 nil)
func (c *Cache) Site(hostingID string) *jsx.OrderedMap {
	m, _ := c.Meta(hostingID, "SITE").(*jsx.OrderedMap)
	return m
}

// PostGroup — POST_GROUPS 의 어느 묶음에서든 groupKey 가 같은 게시판(복사본). 없으면 nil.
func (c *Cache) PostGroup(hostingID, groupKey string) *jsx.OrderedMap {
	gs, _ := c.Meta(hostingID, "POST_GROUPS").(*jsx.OrderedMap)
	for _, k := range gs.Keys() {
		var found *jsx.OrderedMap
		for _, x := range gs.Arr(k) {
			if g, _ := x.(*jsx.OrderedMap); g != nil && jsx.Str(g.V("groupKey")) == groupKey {
				found = jsx.Assign(nil, g) // { ...g } — 같은 묶음에서 여럿이면 마지막
			}
		}
		if found != nil {
			return found
		}
	}
	return nil
}

// Companies — 회사 목록(공통 설정 CompaniesKey 의 items)
func (c *Cache) Hostings() []any {
	return c.All().Map("").Map(HostingsKey).Arr("items")
}

// Hosting — 호스팅 하나(없으면 nil)
func (c *Cache) Hosting(hostingID string) *jsx.OrderedMap {
	for _, x := range c.Hostings() {
		if m, _ := x.(*jsx.OrderedMap); m != nil && jsx.StrOrEmpty(m.V("hostingId")) == hostingID {
			return m
		}
	}
	return nil
}
