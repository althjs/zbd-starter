package core

import (
	"context"
	"database/sql"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// 관리자 권한 그룹 — pages/api/admin/group/*

func adminGroupsAll(ctx context.Context, q db.Querier, hostingID string) ([]*jsx.OrderedMap, error) {
	sqlText := `select groupId, name, description, 
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt, 
      UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, 
      status, hostingId
    from admin_roles`
	var args []any
	if hostingID != "" {
		sqlText += ` WHERE hostingId = ?`
		args = append(args, hostingID)
	}
	return query(ctx, q, sqlText, args...)
}

func adminGroupGet(ctx context.Context, q db.Querier, groupID any) (*jsx.OrderedMap, error) {
	return queryOne(ctx, q, `select groupId, name, description, 
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt, 
      UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, 
      status, json, hostingId
    from admin_roles
    WHERE groupId = ?`, groupID)
}

// orEmptyStr — v || ""
func orEmptyStr(v any) any {
	if !jsx.Truthy(v) {
		return ""
	}
	return v
}

func nilIfNull(m *jsx.OrderedMap) any {
	if m == nil {
		return nil
	}
	return m
}

// adminGroup — GET 목록 · POST 추가
func adminGroup(c *web.Ctx) {
	ctx := c.Context()
	switch c.Method {
	case "GET":
		gs, err := adminGroupsAll(ctx, c.App.DB, c.Token.HostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rowsAny(gs))
	case "POST":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청")
			return
		}
		o, err := db.Exec(ctx, c.App.DB, `INSERT INTO admin_roles (name, description, json, hostingId)
  VALUES (?,?,?,?)`, bd.V("name"), bd.V("description"), bd.V("json"), orEmptyStr(bd.V("hostingId")))
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) > 0 {
			g, err := adminGroupGet(ctx, c.App.DB, db.InsertID(o))
			if err != nil {
				c.Ecb(err)
				return
			}
			c.Scb(nilIfNull(g))
		} else {
			c.Ecb(o)
		}
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminGroupOne — GET · PUT · DELETE /admin/group/[groupId]
func adminGroupOne(c *web.Ctx) {
	ctx := c.Context()
	groupID := c.QV("groupId")
	if !jsx.Truthy(groupID) {
		c.Ecb("잘못된 요청")
		return
	}
	switch c.Method {
	case "GET":
		g, err := adminGroupGet(ctx, c.App.DB, groupID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(nilIfNull(g))
	case "PUT":
		bd := c.BodyMap()
		if !jsx.Truthy(bd.V("name")) {
			c.Ecb("잘못된 요청 (PUT)")
			return
		}
		o, err := db.Exec(ctx, c.App.DB, `UPDATE admin_roles SET name = ?, description = ?, json = ?, 
  hostingId = ?,
  updatedAt = current_timestamp() WHERE groupId = ?`, bd.V("name"), bd.V("description"), bd.V("json"), orEmptyStr(bd.V("hostingId")), bd.V("groupId"))
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("업데이트 오류")
			return
		}
		g, err := adminGroupGet(ctx, c.App.DB, groupID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(nilIfNull(g))
	case "DELETE":
		o, err := db.Exec(ctx, c.App.DB, `DELETE FROM admin_roles WHERE groupId = ? AND hostingId = ?`, groupID, c.Token.HostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) == 0 {
			c.Ecb("삭제 오류")
		} else {
			c.Scb("삭제 완료")
		}
	default:
		c.Ecb("잘못된 요청")
	}
}

func groupAuthGet(ctx context.Context, q db.Querier, groupID any) ([]*jsx.OrderedMap, error) {
	return query(ctx, q, `SELECT menuId, auth FROM admin_role_menus WHERE groupId = ? AND status = 0`, groupID)
}

// adminGroupAuth — 그룹의 메뉴 권한 GET · PUT(배열 [{menuId, auth}] 을 한 트랜잭션으로)
func adminGroupAuth(c *web.Ctx) {
	ctx := c.Context()
	groupID := c.QV("groupId")
	if !jsx.Truthy(groupID) {
		c.Ecb("잘못된 요청")
		return
	}
	switch c.Method {
	case "GET":
		rows, err := groupAuthGet(ctx, c.App.DB, groupID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rowsAny(rows))
	case "PUT":
		items, _ := c.Body().([]any)
		err := db.Tx(ctx, c.App.DB, func(tx *sql.Tx) error {
			for _, x := range items {
				it, _ := x.(*jsx.OrderedMap)
				if it == nil {
					it = jsx.NewOrderedMap()
				}
				if _, err := db.Exec(ctx, tx, `INSERT INTO admin_role_menus (groupId, menuId, auth) VALUES (?,?,?) 
  ON DUPLICATE KEY UPDATE auth = ?`, groupID, it.V("menuId"), it.V("auth"), it.V("auth")); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			c.Ecb(err)
			return
		}
		rows, err := groupAuthGet(ctx, c.App.DB, groupID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(rowsAny(rows))
	default:
		c.Ecb("잘못된 요청")
	}
}
