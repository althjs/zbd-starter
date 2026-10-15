package core

import (
	"context"
	"database/sql"
	"net"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/adminauth"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
	"zbd.kr/be/internal/web"
)

// 관리자 메뉴 — pages/api/admin/menus/*

// menuSpecials — 특수 하위메뉴(사이드 메뉴가 하위 항목을 그 자리에서 채운다). 화면이 아는 이름만.
var menuSpecials = []string{"mail", "adminPosts", "userPosts"} // mail = 웹메일 플러그인(화면 목록에 늘 있다 — 빠져 있어 저장 때 버려졌다, 2026-10-09)

func specialOf(it *jsx.OrderedMap) any {
	s, ok := it.V("special").(string)
	if !ok {
		return nil
	}
	for _, x := range menuSpecials {
		if x == s {
			return s
		}
	}
	return nil
}

// allAdminMenus — 회사 메뉴 트리(API 개수 포함)
func allAdminMenus(ctx context.Context, q db.Querier, hostingID string) ([]any, error) {
	where := ""
	var args []any
	if hostingID != "" {
		where = "WHERE a.hostingId = ?"
		args = append(args, hostingID)
	}
	rows, err := query(ctx, q, `select a.menuId, parentId, sortNo, name, description, icon, path, special, IFNULL(b.cnt, 0) as api_count, a.hostingId from admin_menus a
LEFT JOIN (
	select menuId, hostingId, count(*) as cnt from admin_menu_apis GROUP BY menuId, hostingId
) as b
ON a.menuId = b.menuId AND a.hostingId = b.hostingId
`+where+`
;`, args...)
	if err != nil {
		return nil, err
	}
	menus, _, _ := adminauth.MakeMenuTree(rows)
	return rowsAny(menus), nil
}

// hostingOf — 토큰의 호스팅, 없으면 주소의 hostingId
func hostingOf(c *web.Ctx) string {
	if c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return c.Q("hostingId")
}

func updateMenuItem(ctx context.Context, q db.Execer, it *jsx.OrderedMap, hostingID string) (*jsx.OrderedMap, error) {
	if jsx.Truthy(it.V("isDeleted")) {
		return db.Exec(ctx, q, `DELETE FROM admin_menus WHERE menuId = ? AND hostingId = ?`, it.V("menuId"), hostingID)
	}
	path := strings.TrimSpace(jsx.StrOrEmpty(it.V("path")))
	return db.Exec(ctx, q, `INSERT INTO admin_menus (menuId, parentId, sortNo, name, icon, path, description, special, hostingId) VALUES (?,?,?,?,?,?,?,?,?) 
  ON DUPLICATE KEY UPDATE sortNo = ?, name = ?, icon = ?, path = ?, 
  description = ?, special = ?,
  updatedAt = current_timestamp()`,
		it.V("menuId"), it.V("parentId"), it.V("sortNo"), it.V("name"), orEmptyStr(it.V("icon")), path, orEmptyStr(it.V("description")), specialOf(it), hostingID,
		it.V("sortNo"), orEmptyStr(it.V("name")), orEmptyStr(it.V("icon")), path, orEmptyStr(it.V("description")), specialOf(it))
}

// adminMenus — GET 메뉴 트리 · PUT 메뉴 전체 저장(순서 = 배열 순서) · DELETE(?menuId)
func adminMenus(c *web.Ctx) {
	ctx := c.Context()
	hostingID := hostingOf(c)
	switch c.Method {
	case "GET":
		menus, err := allAdminMenus(ctx, c.App.DB, hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(menus)
	case "PUT":
		items, _ := c.Body().([]any)
		if len(items) == 0 {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		updated, inserted, deleted, errs := 0.0, 0.0, 0.0, 0.0
		err := db.Tx(ctx, c.App.DB, func(tx *sql.Tx) error {
			for i, x := range items {
				it, _ := x.(*jsx.OrderedMap)
				if it == nil {
					it = jsx.NewOrderedMap()
				}
				it.Set("sortNo", float64(i))
				r, err := updateMenuItem(ctx, tx, it, hostingID)
				if err != nil {
					return err
				}
				switch db.Affected(r) {
				case 2:
					updated++
				case 1:
					inserted++
				}
				if it.Has("items") && jsx.Truthy(it.V("items")) {
					for j, y := range it.Arr("items") {
						sub, _ := y.(*jsx.OrderedMap)
						if sub == nil {
							sub = jsx.NewOrderedMap()
						}
						sub.Set("sortNo", float64(j))
						r, err := updateMenuItem(ctx, tx, sub, hostingID)
						if err != nil {
							return err
						}
						if jsx.Truthy(sub.V("isDeleted")) {
							if db.Affected(r) == 1 {
								deleted++
							} else {
								errs++
							}
						} else if db.Affected(r) == 2 {
							updated++
						} else if db.Affected(r) == 1 {
							inserted++
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			c.Ecb(err)
			return
		}
		menus, err := allAdminMenus(ctx, c.App.DB, hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(jsx.Obj(
			"summary", jsx.Obj("updated", updated, "inserted", inserted, "deleted", deleted, "errors", errs),
			"admin", jsx.Obj("menus", menus),
		))
	case "DELETE":
		menuID := c.QV("menuId")
		if !jsx.Truthy(menuID) {
			c.Ecb("잘못된 요청")
			return
		}
		if _, err := db.Exec(ctx, c.App.DB, `DELETE FROM admin_menus WHERE hostingId = ? AND menuId = ?`, hostingID, menuID); err != nil {
			c.Ecb(err)
			return
		}
		menus, err := allAdminMenus(ctx, c.App.DB, hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(jsx.Obj("admin", jsx.Obj("menus", menus)))
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminMenuApis — 메뉴에 묶인 API: GET 목록 · PUT 하나 저장(apiId 가 있으면 고치기)
func adminMenuApis(c *web.Ctx) {
	ctx := c.Context()
	menuID := c.QV("menuId")
	hostingID := hostingOf(c)
	switch c.Method {
	case "GET":
		sqlText := `SELECT apiId, api, description FROM admin_menu_apis WHERE menuId = ?`
		args := []any{menuID}
		if hostingID != "" {
			sqlText += ` AND hostingId = ?`
			args = append(args, hostingID)
		}
		rows, err := query(ctx, c.App.DB, sqlText, args...)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rowsAny(rows))
	case "PUT":
		bd, ok := c.Body().(*jsx.OrderedMap)
		if !ok || bd == nil {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		bd.Set("hostingId", hostingID)
		var o *jsx.OrderedMap
		var err error
		if jsx.Truthy(bd.V("apiId")) {
			o, err = db.Exec(ctx, c.App.DB, `UPDATE admin_menu_apis SET api = ?, description = ? WHERE apiId = ?`, bd.V("api"), bd.V("description"), bd.V("apiId"))
		} else {
			o, err = db.Exec(ctx, c.App.DB, `INSERT INTO admin_menu_apis (menuId, api, description, hostingId) VALUES (?,?,?,?)`, bd.V("menuId"), bd.V("api"), bd.V("description"), hostingID)
		}
		if err != nil {
			c.Ecb(err)
			return
		}
		if !jsx.Truthy(bd.V("apiId")) {
			bd.Set("apiId", o.V("insertId"))
		}
		c.Scb(bd)
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminMenuApiDelete — DELETE /admin/menus/apis/[menuId]/[apiId]
func adminMenuApiDelete(c *web.Ctx) {
	if c.Method != "DELETE" {
		c.Ecb("잘못된 요청")
		return
	}
	o, err := db.Exec(c.Context(), c.App.DB, `DELETE FROM admin_menu_apis WHERE apiId = ? AND menuId = ?`, c.QV("apiId"), c.QV("menuId"))
	if err != nil {
		c.Ecb(err)
		return
	}
	c.Scb(o)
}

// isOfficeReq — 사내 요청인가(표시 용도). 프록시 헤더가 없으면 접속 주소를 본다(지금 서비스 웹서버가 그 주소를 x-forwarded-for 로 채운다).
func isOfficeReq(c *web.Ctx) bool {
	h := c.R.Header.Clone()
	if h.Get("x-forwarded-for") == "" {
		if host, _, err := net.SplitHostPort(c.R.RemoteAddr); err == nil {
			h.Set("x-forwarded-for", host)
		}
	}
	return util.IsOffice(h, util.OfficeIPs(c.App.Cfg.Get("OFFICE_IPS")))
}
