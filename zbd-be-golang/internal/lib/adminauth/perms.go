package adminauth

// 세부 권한 — 메뉴 권한(R/RW)만으로 가를 수 없는 「설정을 바꿀 수 있는 사람」 같은 권한.
//
// 권한 그룹(admin_roles.json.perms = ["attendance.ip", …])에 붙는다 — 표를 따로 두지 않고 그룹 화면이 저장하는 json 그대로.
// 개발자·최고 관리자는 늘 허용. 권한 이름은 기능마다 init() 에서 RegisterPerm 으로 등록한다(플러그인 코드가 자기 권한을 들고 다님).

import (
	"sort"
	"strings"
	"sync"

	"zbd.kr/be/internal/jsx"
)

type PermDef struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Group string `json:"group"` // 화면 묶음 이름(예: 「근무 관리」)
	Desc  string `json:"desc,omitempty"`
}

var (
	permMu   sync.RWMutex
	permDefs = map[string]PermDef{}
)

// RegisterPerm — 세부 권한 등록(같은 키는 덮어씀)
func RegisterPerm(d PermDef) {
	permMu.Lock()
	permDefs[d.Key] = d
	permMu.Unlock()
}

// PermDefs — 등록된 세부 권한(묶음·키 순)
func PermDefs() []PermDef {
	permMu.RLock()
	out := make([]PermDef, 0, len(permDefs))
	for _, d := range permDefs {
		out = append(out, d)
	}
	permMu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// RolePermsOf — 그룹 json 의 perms 를 정리(문자열만 · 공백 제거 · 중복 제거)
func RolePermsOf(J *jsx.OrderedMap) map[string]bool {
	out := map[string]bool{}
	if J == nil {
		return out
	}
	for _, x := range J.Arr("perms") {
		if s, ok := x.(string); ok {
			if s = strings.TrimSpace(s); s != "" {
				out[s] = true
			}
		}
	}
	return out
}

// HasPermBy — 판정(순수 함수, 정책 테스트용): 개발자·최고 관리자는 늘, 아니면 그룹 권한에 있을 때만. 계정이 없으면 거절.
func HasPermBy(member *jsx.OrderedMap, rolePerms map[string]map[string]bool, key string) bool {
	if member == nil || key == "" {
		return false
	}
	if jsx.Truthy(member.V("isDeveloper")) || jsx.Truthy(member.V("isSuperAdmin")) {
		return true
	}
	return rolePerms[jsx.Str(member.V("groupId"))][key]
}

// HasPerm — 이 관리자가 세부 권한을 가졌나
func (c *Cache) HasPerm(adminID, key string) bool {
	s := c.Get()
	return HasPermBy(s.MemberAuthMap.Map(adminID), s.RolePerms, key)
}
