package admincontent

// 사이트 바탕 데이터 — 사용자 화면이 처음 한 번 받는 사이트 설정·게시판 목록·공통 코드(`/service/design/{호스팅}/base`).
// 주소의 design 은 옛 이름(화면이 이 주소를 부른다) — 스타터에는 페이지 디자인 기능이 없다.

import (
	"context"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/sitepub"
	"zbd.kr/be/internal/web"
)

// 공통("") COMMON 설정을 DB 에서 바로 읽는다(지금 서비스의 강제 로딩)
func commonMetaFromDB(ctx context.Context, q db.Querier) (*jsx.OrderedMap, error) {
	r, err := db.QueryOne(ctx, q, `SELECT json FROM settings WHERE hostingId = "" AND metaKey = "COMMON"`)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errMsg("TypeError: Cannot read properties of undefined (reading 'json')")
	}
	v, err := parseJSONStr(r.V("json"))
	if err != nil {
		return nil, err
	}
	return asMap(v), nil
}

// getBaseData — 사이트 공통 바탕(사이트 설정·게시판·공통 코드 …). 스타터에는 페이지 디자인(GNB) 이 없다
func getBaseData(ctx context.Context, c *web.Ctx, q db.Querier, hostingID string) (*jsx.OrderedMap, error) {
	var err error
	meta, _ := jsx.Clone(c.App.Settings.Config(hostingID)).(*jsx.OrderedMap)
	postGroups := asMap(jsx.Clone(meta.V("POST_GROUPS")))
	siteMeta := asMap(jsx.Clone(meta.V("SITE")))
	if co := c.App.Settings.Hosting(hostingID); co != nil {
		if v := co.Path("server", "associationAdminId"); jsx.Truthy(v) {
			if siteMeta == nil {
				return nil, errMsg("TypeError: Cannot set properties of undefined (setting 'associationAdminId')")
			}
			siteMeta.Set("associationAdminId", v)
		}
	}
	var label any
	if v := meta.V("PRODUCT_LABLES"); jsx.Truthy(v) {
		label = jsx.Clone(v)
	}
	var promotionTemplates any
	if v := meta.Path("PROMOTION_TEMPLATES", "items"); jsx.Truthy(v) {
		promotionTemplates = jsx.Clone(v)
	}
	if !meta.Has("POST_GROUPS") {
		postGroups = jsx.Obj("adminPostGroups", []any{}, "userPostGrouips", []any{})
	}
	if postGroups != nil {
		apg := []any{}
		for _, x := range postGroups.Arr("adminPostGroups") {
			if jsx.Truthy(asMap(x).V("isBoardStore")) {
				apg = append(apg, x)
			}
		}
		postGroups.Set("adminPostGroups", apg)
	}

	common, err := commonMetaFromDB(ctx, q)
	if err != nil {
		return nil, err
	}
	commonCodes := c.App.Settings.Config(hostingID).Path("COMMON_CODES", "items")
	if !jsx.Truthy(commonCodes) {
		commonCodes = []any{}
	}

	if siteMeta != nil {
		if langs, ok := siteMeta.Path("feature", "i18n").([]any); ok && len(langs) > 0 {
			design := siteMeta.Map("design")
			if design == nil {
				return nil, errMsg("TypeError: Cannot read properties of undefined (reading 'i18n')")
			}
			if !design.Has("i18n") {
				design.Set("i18n", jsx.Obj("dictionary", []any{}, "langs", langs))
			}
			i18n := design.Map("i18n")
			dict, _ := i18n.V("dictionary").([]any)
			if extra, ok := jsx.Clone(common.Path("i18n", "dictionary")).([]any); ok {
				dict = append(append([]any{}, dict...), extra...)
			}
			for _, x := range dict {
				if m := asMap(x); m != nil {
					m.Delete("desc")
					m.Delete("group")
				}
			}
			if dict == nil {
				dict = []any{}
			}
			i18n.Set("dictionary", dict)
		}
		siteMeta.Delete("tourReserveNotiSetting")
	} else {
		return nil, errMsg("TypeError: Cannot convert undefined or null to object")
	}

	data := jsx.NewOrderedMap()
	data.Set("siteMeta", siteMeta)
	if postGroups != nil {
		data.Set("postGroupsMeta", postGroups)
	}
	data.Set("commonCodes", commonCodes)
	data.Set("promotionTemplates", promotionTemplates)
	data.Set("categories", []any{})
	if common.Has("holydays") {
		data.Set("holydays", common.V("holydays"))
	}
	data.Set("label", label)
	if common.Has("CSS") {
		data.Set("CSS", common.V("CSS"))
	}
	return data, nil
}

func designBase(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		base, err := getBaseData(ctx, c, q, c.Q("hostingId"))
		if err != nil {
			return err
		}
		sm := base.Map("siteMeta")
		if sm == nil {
			return errMsg("TypeError: Cannot read properties of undefined (reading 'design')")
		}
		sm.Map("design").Delete("globalHTML")
		// 공개 주소 — 사용자 화면이 쓰는 칸만(sitepub). 화면 HTML 이 ?v=지문 으로 부르면 길게 캐시(내용이 바뀌면 지문이 바뀐다)
		pub := sitepub.PublicBase(base)
		if c.Q("v") != "" {
			c.SetHeader("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.SetHeader("Cache-Control", "no-cache")
		}
		c.Scb(pub)
		return nil
	})
}
