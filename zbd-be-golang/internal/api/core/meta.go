package core

import (
	"context"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// 회사 설정(settings 표) — pages/api/admin/meta/*  (사이트빌더 분기는 옮기지 않았다)

func metaAll(ctx context.Context, q db.Querier, hostingID string) ([]any, error) {
	sqlText := `SELECT metaKey, name, memo, json, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId from settings`
	var args []any
	if hostingID != "" {
		sqlText += " WHERE hostingId = ?"
		args = append(args, hostingID)
	}
	rows, err := query(ctx, q, sqlText, args...)
	if err != nil {
		return nil, err
	}
	out := []any{}
	for _, o := range rows {
		n := jsx.Assign(nil, o)
		j, err := jsParse(o.V("json"))
		if err != nil {
			return nil, err
		}
		n.Set("json", j)
		out = append(out, n)
	}
	return out, nil
}

// jsParse — JSON.parse(v): null → null, 없는 값은 JS 와 같은 오류
func jsParse(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	s := jsx.Str(v)
	r, err := jsx.ParseString(s)
	if err != nil {
		return nil, errFmt("Unexpected token in JSON")
	}
	return r, nil
}

func metaGet(ctx context.Context, q db.Querier, metaKey, hostingID any) (*jsx.OrderedMap, error) {
	o, err := queryOne(ctx, q, `SELECT name, memo, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, json, hostingId from settings WHERE metaKey = ? AND hostingId = ?`, metaKey, hostingID)
	if err != nil {
		return nil, err
	}
	r := jsx.NewOrderedMap()
	var js any = "{}"
	if o != nil {
		for _, k := range []string{"name", "memo", "updatedAt"} {
			r.Set(k, o.V(k))
		}
		if jsx.Truthy(o.V("json")) {
			js = o.V("json")
		}
	}
	j, err := jsParse(js)
	if err != nil {
		return nil, err
	}
	r.Set("json", j)
	if o != nil {
		r.Set("hostingId", o.V("hostingId"))
	}
	r.Set("metaKey", metaKey)
	return r, nil
}

// metaJSONGet — 설정 하나의 json(없으면 지금 서비스처럼 오류)
func metaJSONGet(ctx context.Context, q db.Querier, metaKey, hostingID string) (any, error) {
	o, err := queryOne(ctx, q, `SELECT json from settings WHERE metaKey = ? AND hostingId = ?`, metaKey, hostingID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errFmt(`SyntaxError: "undefined" is not valid JSON`)
	}
	return jsParse(o.V("json"))
}

// metaJSONSet — 설정 json 저장(없으면 추가) + 설정 캐시 갱신 알림
func metaJSONSet(c *web.Ctx, q db.Execer, metaKey string, data any, hostingID string) error {
	var js any
	if data != nil {
		js = jsx.Stringify(data)
	}
	o, err := db.Exec(c.Context(), q, `INSERT INTO settings (metaKey, json, hostingId) VALUES (?, ?, ?)
  ON DUPLICATE KEY UPDATE json = ?, updatedAt = current_timestamp()`, metaKey, js, hostingID, js)
	if err != nil {
		return err
	}
	if db.Affected(o) < 1 {
		return errFmt("메타 테이블 처리 실패!!")
	}
	settingsChanged(c, hostingID)
	return nil
}

// settingsChanged — 설정이 바뀌었다: 이 프로세스 캐시를 바로 다시 읽고 다른 서버에도 알린다(doUpdateCodes)
func settingsChanged(c *web.Ctx, hostingID string) {
	c.App.ReloadCaches(c.Context())
	c.App.Emit("doUpdateCodes", jsx.Obj("hostingId", hostingID))
}

func metaHosting(c *web.Ctx) string {
	if c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return c.Q("hostingId")
}

// adminMeta — GET(?metaKey 없으면 전체) · PUT {metaKey, name, memo, json}
func adminMeta(c *web.Ctx) {
	ctx := c.Context()
	hostingID := metaHosting(c)
	bd := c.BodyMap()
	metaKey := c.Q("metaKey")
	if metaKey == "" {
		metaKey = jsx.StrOrEmpty(bd.V("metaKey"))
	}
	switch c.Method {
	case "GET":
		var r any
		var err error
		if metaKey != "" {
			r, err = metaGet(ctx, c.App.DB, metaKey, hostingID)
		} else {
			r, err = metaAll(ctx, c.App.DB, hostingID)
		}
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(r)
	case "PUT":
		if c.Body() == nil || !jsx.Truthy(bd.V("metaKey")) || !jsx.Truthy(bd.V("memo")) {
			c.Ecb("잘못된 요청")
			return
		}
		var js any
		if jsx.Truthy(bd.V("json")) {
			js = bd.V("json")
		}
		jsStr := jsx.Stringify(js)
		o, err := db.Exec(ctx, c.App.DB, `INSERT INTO settings (metaKey, hostingId, name, memo, json)
    VALUES (?,?,?,?,?)
    ON DUPLICATE KEY UPDATE name = ?, memo = ?, json = ?, updatedAt = current_timestamp()`,
			bd.V("metaKey"), hostingID, bd.V("name"), bd.V("memo"), jsStr, bd.V("name"), bd.V("memo"), jsStr)
		if err != nil {
			c.Ecb(err)
			return
		}
		if db.Affected(o) < 1 {
			c.Ecb("메타 테이블 메모 업데이트 실패!!")
			return
		}
		c.App.ReloadCaches(ctx)
		r, err := metaGet(ctx, c.App.DB, bd.V("metaKey"), hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(r)
	default:
		c.Ecb("잘못된 요청")
	}
}

// adminMetaKey — 설정 하나의 json 읽기·쓰기(본문 전체가 json)
func adminMetaKey(c *web.Ctx) {
	ctx := c.Context()
	hostingID := metaHosting(c)
	metaKey := c.Q("metaKey")
	switch c.Method {
	case "GET":
		r, err := metaJSONGet(ctx, c.App.DB, metaKey, hostingID)
		if err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(r)
	case "PUT":
		bd := c.Body()
		if err := metaJSONSet(c, c.App.DB, metaKey, bd, hostingID); err != nil {
			c.Ecb(err)
			return
		}
		c.Scb(bd)
	default:
		c.Ecb("잘못된 요청")
	}
}
