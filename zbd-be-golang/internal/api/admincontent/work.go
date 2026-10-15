package admincontent

// 업무관리 — /admin/work, /admin/work/[workId], /admin/work/[workId]/[shapeId], /admin/work/quick-update
// 원본: pages/api/admin/work/*
//
// 접근: 본인 회사 항목 · 공유 보드(공통 COMMON_CODES 의 그 groupKey setting.sharedHostingIds) · 초대 메일(inviteEmails) · 부모 회사 스코프.

import (
	"context"
	"log"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/auth"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// 사본: pages/api/admin/member/constants.ts getAllAdminMembers — 전환 뒤 합칠 것(캐시 없이 바로 읽는다)
func adminMembers(ctx context.Context, q db.Querier, hostingID string) ([]*jsx.OrderedMap, error) {
	query := `select m.adminId, m.groupId, m.name, m.email, m.mobile,
        UNIX_TIMESTAMP(m.createdAt)*1000 as createdAt,
        UNIX_TIMESTAMP(m.updatedAt)*1000 as updatedAt,
        m.status, m.hostingId, m.json,
        g.name AS groupName
      from admins as m
        join admin_roles as g on m.groupId = g.groupId WHERE m.status >= 0`
	var args []any
	if hostingID != "" {
		query += ` AND m.hostingId = ?`
		args = append(args, hostingID)
	}
	return db.Query(ctx, q, query, args...)
}

// 사본: pages/api/admin/member/constants.ts getAdmin(캐시 안 씀)
func getAdminRow(ctx context.Context, q db.Querier, adminID any) (*jsx.OrderedMap, error) {
	return db.QueryOne(ctx, q, `select adminId, groupId, name, email, mobile,
      UNIX_TIMESTAMP(createdAt)*1000 as createdAt,
      UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt,
      status, json, hostingId

    from admins
    WHERE adminId = ?`, jsx.Param(adminID))
}

// 사본: pages/api/admin/member/constants.ts updateAdmin(비밀번호·otp 를 바꾸지 않는 호출만)
func updateAdmin(ctx context.Context, q db.Execer, m *jsx.OrderedMap) (*jsx.OrderedMap, error) {
	var group any
	if jsx.Truthy(m.V("groupId")) {
		group = m.V("groupId")
	}
	var cid any
	if jsx.Truthy(m.V("hostingId")) {
		cid = m.V("hostingId")
	}
	return db.Exec(ctx, q, `UPDATE admins m SET
  groupId = COALESCE(NULLIF(?, 0), m.groupId),
  name = ?,
  email = ?,
  password = COALESCE(NULLIF(?, ''), m.password),
  mobile = ?, json = ?,
  hostingId = COALESCE(NULLIF(?, ''), m.hostingId),
  updatedAt = current_timestamp() WHERE adminId = ?`,
		jsx.Param(group), jsx.Param(m.V("name")), jsx.Param(m.V("email")), nil, jsx.Param(m.V("mobile")), jsx.Param(m.V("json")), cid, jsx.Param(m.V("adminId")))
}

// browserUID — token._browserUid || token.browserUid
func browserUID(t *auth.Token) string {
	if t.HeaderUID != "" {
		return t.HeaderUID
	}
	return t.BrowserUID
}

// commonCodeItem — 회사 COMMON_CODES.items 에서 group 이 같은 첫 항목
func commonCodeItem(c *web.Ctx, hostingID string, groupKey any, strict bool) *jsx.OrderedMap {
	items, _ := c.App.Settings.Config(hostingID).Path("COMMON_CODES", "items").([]any)
	for _, x := range items {
		m := asMap(x)
		if m == nil {
			continue
		}
		if (strict && strictEq(m.V("group"), groupKey)) || (!strict && looseEq(m.V("group"), groupKey)) {
			return m
		}
	}
	return nil
}

func sharedHostingIDs(c *web.Ctx, groupKey string) []string {
	if groupKey == "" {
		return nil
	}
	code := commonCodeItem(c, "", groupKey, true)
	ids, _ := code.Path("setting", "sharedHostingIds").([]any)
	var out []string
	for _, x := range ids {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func parentHostingIDForGroupKey(c *web.Ctx, groupKey string, extra ...string) string {
	if groupKey == "" {
		return ""
	}
	scopes := []string{""}
	for _, s := range extra {
		if !contains(scopes, s) {
			scopes = append(scopes, s)
		}
	}
	for _, sc := range scopes {
		code := commonCodeItem(c, sc, groupKey, true)
		if p, ok := code.Path("setting", "parentHostingId").(string); ok && p != "" {
			return p
		}
	}
	return ""
}

func isHostingShared(c *web.Ctx, t *auth.Token, groupKey string) bool {
	if t.HostingID == "" || groupKey == "" {
		return false
	}
	return contains(sharedHostingIDs(c, groupKey), t.HostingID)
}

func workAccessibleHostingIDs(c *web.Ctx, t *auth.Token, groupKey string) []string {
	own := t.HostingID
	if groupKey == "" {
		return []string{own}
	}
	if own == "" {
		return []string{""}
	}
	if isHostingShared(c, t, groupKey) {
		return []string{"", own}
	}
	return []string{own}
}

func canAccessWorkItem(c *web.Ctx, t *auth.Token, item *jsx.OrderedMap) bool {
	if t == nil || item == nil {
		return false
	}
	if strictEq(item.V("hostingId"), t.HostingID) {
		return true
	}
	if strictEq(item.V("hostingId"), "") && isHostingShared(c, t, orStr(item.V("groupKey"))) {
		return true
	}
	return false
}

func resolveWorkWriteHostingID(c *web.Ctx, t *auth.Token, groupKey string, item *jsx.OrderedMap) string {
	if s, ok := item.V("hostingId").(string); ok && s != "" {
		return s
	}
	if s, ok := item.V("hostingId").(string); ok && s == "" {
		if t.HostingID == "" || isHostingShared(c, t, groupKey) {
			return ""
		}
	}
	if t.HostingID != "" && isHostingShared(c, t, groupKey) {
		return ""
	}
	return t.HostingID
}

// resolveParentHostingID — 토큰의 상위 회사(SITE.parentHostingId). 즈비드에는 판매점·협회(사이트빌더)가 없다.
func resolveParentHostingID(c *web.Ctx, t *auth.Token) string {
	own := t.HostingID
	p := strings.TrimSpace(orStr(c.App.Settings.Config(own).Path("SITE", "parentHostingId")))
	if p != "" && p != own {
		return p
	}
	return own
}

// isInvited — (Json.inviteEmails || "").includes(token.email || "")
func isInvited(item *jsx.OrderedMap, email string) bool {
	v := item.Path("Json", "inviteEmails")
	switch x := v.(type) {
	case []any:
		for _, e := range x {
			if strictEq(e, email) {
				return true
			}
		}
		return false
	case string:
		return strings.Contains(x, email)
	}
	if jsx.Truthy(v) {
		return false
	}
	return strings.Contains("", email)
}

func isParentScoped(c *web.Ctx, t *auth.Token, item *jsx.OrderedMap) bool {
	gp := parentHostingIDForGroupKey(c, orStr(item.V("groupKey")), orStr(item.V("hostingId")), t.HostingID)
	return gp != "" && strictEq(item.V("hostingId"), gp) && gp == resolveParentHostingID(c, t)
}

const workCols = `
      workId,
      adminId,
      createdAt,
      Json,
      parentWorkId,
      hostingId,
      groupKey,
      updatedAt
    FROM work_items`

type workSearch struct {
	hostingIDs  []string
	groupKey    string
	workIDs     any
	inquiryType string
}

func getWorkItems(ctx context.Context, q db.Querier, s workSearch) ([]*jsx.OrderedMap, error) {
	where := "WHERE 1=1"
	var args []any
	ids := s.hostingIDs
	if ids == nil {
		ids = []string{""}
	}
	cond := "hostingId = ?"
	if len(ids) != 1 {
		cond = "hostingId IN (" + db.Placeholders(len(ids)) + ")"
	}
	var cargs []any
	for _, x := range ids {
		cargs = append(cargs, x)
	}
	where += " AND " + cond
	args = append(args, cargs...)
	if s.groupKey != "" {
		where += " AND groupKey = ?"
		args = append(args, s.groupKey)
	}
	if jsx.Truthy(s.workIDs) {
		where += " AND workId IN (?)"
		args = append(args, s.workIDs)
	}
	if s.inquiryType != "" {
		where += " AND JSON_UNQUOTE(JSON_EXTRACT(Json, '$.inquiryType')) = ?"
		args = append(args, s.inquiryType)
	}
	rows, err := db.Query(ctx, q, `
    SELECT `+workCols+`
    `+where+`
    ORDER BY workId DESC
  `, args...)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if jsx.Truthy(r.V("Json")) {
			v, err := parseJSONStr(r.V("Json"))
			if err != nil {
				return nil, err
			}
			r.Set("Json", v)
		} else {
			r.Set("Json", jsx.NewOrderedMap())
		}
	}
	return rows, nil
}

func getWorkItem(ctx context.Context, q db.Querier, workID any) (*jsx.OrderedMap, error) {
	r, err := db.QueryOne(ctx, q, `
    SELECT `+workCols+`
    WHERE workId = ?
  `, workID)
	if err != nil || r == nil {
		return nil, err
	}
	if jsx.Truthy(r.V("Json")) {
		v, err := parseJSONStr(r.V("Json"))
		if err != nil {
			return nil, err
		}
		r.Set("Json", v)
	}
	return r, nil
}

// notifyAdmin — 관리자 알림 자리(알림함·웹 푸시는 플러그인). 스타터는 보내지 않아 false
func notifyAdmin(ctx context.Context, c *web.Ctx, q dbq, hostingID string, noti *jsx.OrderedMap, adminID any) bool {
	return false
}

func pushData(title, body, url string) *jsx.OrderedMap {
	return jsx.Obj("title", title, "body", body, "url", url, "icon", "")
}

func titleOr(v any) string {
	if jsx.Truthy(v) {
		return jsx.Str(v)
	}
	return "제목없음"
}

// tplU — `${v}` (없으면 "undefined")
func tplU(m *jsx.OrderedMap, k string) string {
	if !m.Has(k) {
		return "undefined"
	}
	return tplStr(m.V(k))
}

// includesNum — arr.includes(n) (엄격 비교 — 숫자만 맞는다)
func includesNum(arr []any, n float64) bool {
	for _, x := range arr {
		if f, ok := x.(float64); ok && f == n {
			return true
		}
	}
	return false
}

func nowMs() float64 { return float64(time.Now().UnixMilli()) }

// workDoNoti — 담당자 할당·멘션·좋아요 알림(보낸 표시를 item.Json 에 남긴다)
func workDoNoti(ctx context.Context, c *web.Ctx, q dbq, item *jsx.OrderedMap) error {
	cid := orStr(item.V("hostingId"))
	ms, err := adminMembers(ctx, q, cid)
	if err != nil {
		return err
	}
	js := item.Map("Json")
	if js == nil {
		return nil // Json 이 없으면 알릴 것도 없다(지금 서비스도 ?. 로 지나간다)
	}
	groupKey := tplU(item, "groupKey")
	base := "/admin/work/" + groupKey + "/" + tplU(item, "workId")

	// 담당자 알림
	adminIDs, _ := js.V("adminIds").([]any)
	var targets []any
	for _, a := range adminIDs {
		if jsx.Truthy(js.Map("adminNotiMap").V(jsx.Str(a))) {
			continue
		}
		targets = append(targets, a)
	}
	for _, a := range targets {
		noti := pushData("[할당] "+groupKey+" "+tplU(js, "title"), titleOr(js.V("title")), base)
		if !js.Has("adminNotiMap") {
			js.Set("adminNotiMap", jsx.NewOrderedMap())
		}
		if notifyAdmin(ctx, c, q, cid, noti, a) {
			if m := js.Map("adminNotiMap"); m != nil {
				m.Set(jsx.Str(a), nowMs())
			}
		}
	}
	if m := js.Map("adminNotiMap"); m != nil {
		for _, k := range m.Keys() {
			if !includesNum(adminIDs, jsx.Number(k)) {
				m.Delete(k)
			}
		}
	}

	now := nowMs()
	mention := func(members any, shapeID any, hasShape bool, subIdx any, hint string) {
		list, _ := members.([]any)
		if len(list) == 0 {
			return
		}
		var params []string
		if hasShape && shapeID != nil && jsx.Str(shapeID) != "" {
			params = append(params, "shapeId="+jsx.Str(shapeID))
		}
		if subIdx != nil {
			params = append(params, "subImageIndex="+jsx.Str(subIdx))
		}
		qs := ""
		if len(params) > 0 {
			qs = "?" + strings.Join(params, "&")
		}
		suffix := ""
		if hint != "" {
			suffix = " — " + hint
		}
		for _, x := range list {
			mb := asMap(x)
			if jsx.Truthy(mb.V("notiAt")) {
				continue
			}
			noti := pushData("[멘션] "+groupKey+" "+titleOr(js.V("title")), titleOr(js.V("title"))+" 에서 멘션되었습니다."+suffix, base+qs)
			if notifyAdmin(ctx, c, q, cid, noti, mb.V("adminId")) {
				mb.Set("notiAt", now)
			}
		}
	}
	mention(js.V("mentionedMembers"), nil, false, nil, "")
	images, _ := js.V("images").([]any)
	for i, img := range images {
		mention(asMap(img).V("mentionedMembers"), nil, false, float64(i), "")
	}
	shapes, _ := js.V("shapes").([]any)
	for _, s := range shapes {
		sm := asMap(s)
		hint := ""
		if jsx.Truthy(sm.V("title")) {
			hint = jsx.Str(sm.V("title"))
		}
		mention(sm.V("mentionedMembers"), sm.V("id"), true, nil, hint)
	}

	like := func(t *jsx.OrderedMap, subIdx any) {
		if t == nil {
			return
		}
		likeIDs, _ := t.V("likeAdminIds").([]any)
		if m := t.Map("likeNotiMap"); m != nil {
			for _, k := range m.Keys() {
				if !includesNum(likeIDs, jsx.Number(k)) {
					m.Delete(jsx.NumStr(jsx.Number(k)))
				}
			}
		}
		if len(likeIDs) == 0 {
			return
		}
		if !t.Has("likeNotiMap") {
			t.Set("likeNotiMap", jsx.NewOrderedMap())
		}
		lm := t.Map("likeNotiMap")
		for _, a := range likeIDs {
			if jsx.Truthy(lm.V(jsx.Str(a))) {
				continue
			}
			var admin *jsx.OrderedMap
			for _, m := range ms {
				if looseEq(m.V("adminId"), a) {
					admin = m
					break
				}
			}
			if admin == nil {
				continue
			}
			url := base
			if subIdx != nil {
				url += "?subImageIndex=" + jsx.Str(subIdx)
			}
			noti := pushData("[반응] "+groupKey+" "+titleOr(js.V("title")), tplStr(admin.V("name"))+"님이 좋아요를 눌렀습니다.", url)
			owner := jsx.Number(item.V("adminId"))
			ok := true
			if owner > 0 && !strictEq(a, item.V("adminId")) {
				ok = notifyAdmin(ctx, c, q, cid, noti, item.V("adminId"))
			}
			if ok && lm != nil {
				lm.Set(jsx.Str(a), now)
			}
		}
	}
	like(js, nil)
	for i, img := range images {
		like(asMap(img), float64(i))
	}
	return nil
}

func emitWork(c *web.Ctx, data *jsx.OrderedMap) { c.App.Emit("doWorkUpdate", data) }

// updateWorkItem — 알림 처리 → 저장 → 다시 읽기
func updateWorkItem(ctx context.Context, c *web.Ctx, q dbq, workID any, item *jsx.OrderedMap, ignoreSocket bool) (*jsx.OrderedMap, error) {
	setUpdated := ", updatedAt = NOW()"
	if ignoreSocket || jsx.Truthy(item.V("ignoreUpdatedAt")) {
		setUpdated = ""
	}
	if err := workDoNoti(ctx, c, q, item); err != nil {
		return nil, err
	}
	js := item.V("Json")
	if !jsx.Truthy(js) {
		js = jsx.NewOrderedMap()
	}
	parent := item.V("parentWorkId")
	if !jsx.Truthy(parent) {
		parent = float64(0)
	}
	if _, err := db.Exec(ctx, q, `
    UPDATE work_items
    SET
      Json = ?,
      parentWorkId = ?,
      groupKey = ?
      `+setUpdated+`
    WHERE workId = ?
  `, jsx.Stringify(js), jsx.Param(parent), orStr(item.V("groupKey")), jsx.Param(workID)); err != nil {
		return nil, err
	}
	if ignoreSocket {
		log.Printf("[doWorkUpdate] skip(ignoreSocket) workId=%v", workID)
	} else {
		ev := jsx.NewOrderedMap()
		if item.Has("hostingId") {
			ev.Set("hostingId", item.V("hostingId"))
		}
		ev.Set("workId", workID)
		if item.Has("groupKey") {
			ev.Set("groupKey", item.V("groupKey"))
		}
		ev.Set("zbduid", browserUID(c.Token))
		ev.Set("oxuid", browserUID(c.Token)) // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
		emitWork(c, ev)
	}
	return getWorkItem(ctx, q, workID)
}

func deleteWorkItem(ctx context.Context, c *web.Ctx, q dbq, workID any) (*jsx.OrderedMap, error) {
	ok, err := db.Exec(ctx, q, `DELETE FROM work_items WHERE workId = ?`, workID)
	if err != nil {
		return nil, err
	}
	emitWork(c, jsx.Obj("hostingId", c.Token.HostingID, "workId", workID, "zbduid", browserUID(c.Token), "oxuid", browserUID(c.Token) /* 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다 */))
	return ok, nil
}

// errNaN — 숫자가 아닌 번호(NaN)가 SQL 에 들어갔을 때 지금 서비스가 내는 오류
var errNaN = errMsg("Error: ER_BAD_FIELD_ERROR: Unknown column 'NaN' in 'WHERE'")

// workIDParam — parseInt(req.query.workId) (NaN 이면 ok=false)
// 옛 이름 canvas — 2026-10 옮기는 기간(업무관리 canvas → work, 2026-10-04): 주소 /admin/canvas…(admincontent.go) 와
// 칸 canvasId·parentCanvasId·canvasIds 도 받는다. 지울 때 grep "옛 이름 canvas".
var legacyWorkKeys = [][2]string{{"canvasId", "workId"}, {"parentCanvasId", "parentWorkId"}, {"canvasIds", "workIds"}, {"canvasIds[]", "workIds[]"}}

func renameLegacyWorkKeys(m *jsx.OrderedMap) {
	if m == nil {
		return
	}
	for _, k := range legacyWorkKeys {
		if m.Has(k[0]) && !m.Has(k[1]) {
			m.Set(k[1], m.V(k[0]))
		}
	}
}

func workIDParam(c *web.Ctx) (float64, bool) {
	n, ok := parseIntJS(c.QV("workId"))
	return float64(n), ok
}

func workIndex(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		t := c.Token
		cid := hostingID(c)
		switch c.Method {
		case "GET":
			groupKey := c.Q("groupKey")
			s := workSearch{groupKey: groupKey, hostingIDs: []string{cid}}
			if groupKey != "" {
				s.hostingIDs = workAccessibleHostingIDs(c, t, groupKey)
			}
			if v := c.QV("workIds"); jsx.Truthy(v) {
				s.workIDs = v
			}
			if v := c.QV("workIds[]"); jsx.Truthy(v) {
				s.workIDs = v
			}
			if v := c.QV("inquiryType"); jsx.Truthy(v) {
				s.inquiryType = jsx.Str(v)
			}
			items, err := getWorkItems(ctx, q, s)
			if err != nil {
				return err
			}
			var meta *jsx.OrderedMap
			if groupKey != "" && t.HostingID != "" && isHostingShared(c, t, groupKey) {
				if code := commonCodeItem(c, "", groupKey, true); code != nil && jsx.Truthy(code.V("items")) {
					cc := jsx.Obj("items", code.V("items"))
					if code.Has("name") {
						cc.Set("name", code.V("name"))
					}
					meta = jsx.Obj("workCodes", cc)
				}
			}
			scbMeta(c, jsx.Obj("items", rowsAny(items)), meta)
		case "POST":
			item := asMap(bodyMap(c).V("item"))
			if item == nil {
				return errMsg("TypeError: Cannot read properties of undefined (reading 'groupKey')")
			}
			writeCID := resolveWorkWriteHostingID(c, t, orStr(item.V("groupKey")), item)
			parent := item.V("parentWorkId")
			if !jsx.Truthy(parent) {
				parent = float64(0)
			}
			js := item.V("Json")
			if !jsx.Truthy(js) {
				js = jsx.NewOrderedMap()
			}
			adminID := any(t.Key)
			if !t.Valid() {
				adminID = float64(0)
			}
			ok, err := db.Exec(ctx, q, `
    INSERT INTO work_items (
      adminId,
      Json,
      parentWorkId,
      hostingId,
      groupKey,
      createdAt,
      updatedAt
    ) VALUES (
      ?,
      ?,
      ?,
      ?,
      ?,
      NOW(),
      NOW()
    )
  `, adminID, jsx.Stringify(js), jsx.Param(parent), writeCID, orStr(item.V("groupKey")))
			if err != nil {
				return err
			}
			id := float64(db.InsertID(ok))
			newItem, err := getWorkItem(ctx, q, id)
			if err != nil {
				return err
			}
			ev := jsx.NewOrderedMap()
			if item.Has("hostingId") {
				ev.Set("hostingId", item.V("hostingId"))
			}
			ev.Set("workId", id)
			if item.Has("groupKey") {
				ev.Set("groupKey", item.V("groupKey"))
			}
			ev.Set("zbduid", browserUID(t))
			ev.Set("oxuid", browserUID(t)) // 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
			emitWork(c, ev)
			if newItem != nil {
				if err := workDoNoti(ctx, c, q, newItem); err != nil {
					log.Printf("@ Socket doWorkUpdate emit failed %v", err)
				}
			}
			scbOne(c, newItem)
		case "PUT":
			item := asMap(bodyMap(c).V("item"))
			if item == nil {
				return errMsg("TypeError: Cannot read properties of undefined (reading 'workId')")
			}
			if !jsx.Truthy(item.V("workId")) {
				c.Ecb("workId가 필요합니다.")
				return nil
			}
			existing, err := getWorkItem(ctx, q, jsx.Param(item.V("workId")))
			if err != nil {
				return err
			}
			if existing == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			if !canAccessWorkItem(c, t, existing) {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			item.Set("hostingId", existing.V("hostingId"))
			item.Set("groupKey", existing.V("groupKey"))
			r, err := updateWorkItem(ctx, c, q, jsx.Param(item.V("workId")), item, false)
			if err != nil {
				return err
			}
			scbOne(c, r)
		case "DELETE":
			id, ok := workIDParam(c)
			if !ok || id == 0 {
				c.Ecb("workId가 필요합니다.")
				return nil
			}
			existing, err := getWorkItem(ctx, q, id)
			if err != nil {
				return err
			}
			if existing == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			if !canAccessWorkItem(c, t, existing) {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			if _, err := deleteWorkItem(ctx, c, q, id); err != nil {
				return err
			}
			c.Scb("DELETED")
		default:
			c.Ecb("잘못된 요청")
		}
		return nil
	})
}

func workOne(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		t := c.Token
		id, idOK := workIDParam(c)
		switch c.Method {
		case "GET":
			var idArg any = id
			if !idOK {
				return errNaN // parseInt 실패 = NaN 이 SQL 에 그대로 들어가 오류
			}
			item, err := getWorkItem(ctx, q, idArg)
			if err != nil {
				return err
			}
			if item == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			if !looseEq(t.HostingID, item.V("hostingId")) {
				if !canAccessWorkItem(c, t, item) && !isInvited(item, t.Email) && !isParentScoped(c, t, item) {
					c.Ecb("초대받지 않은 업무관리입니다.")
					return nil
				}
				code := commonCodeItem(c, orStr(item.V("hostingId")), item.V("groupKey"), false)
				items := code.V("items")
				if !jsx.Truthy(items) {
					items = []any{}
				}
				cc := jsx.Obj("items", items)
				if code.Has("name") {
					cc.Set("name", code.V("name"))
				}
				ms, err := adminMembers(ctx, q, orStr(item.V("hostingId")))
				if err != nil {
					return err
				}
				list := []any{}
				for _, m := range ms {
					if looseEq(m.V("hostingId"), item.V("hostingId")) {
						list = append(list, m)
					}
				}
				c.Scb(item, jsx.Obj("workCodes", cc, "adminMembers", list))
				return nil
			}
			c.Scb(item)
		case "PUT":
			item := asMap(bodyMap(c).V("item"))
			if item == nil {
				return errMsg("TypeError: Cannot read properties of undefined (reading 'workId')")
			}
			if !jsx.Truthy(item.V("workId")) {
				c.Ecb("workId가 필요합니다.")
				return nil
			}
			existing, err := getWorkItem(ctx, q, jsx.Param(item.V("workId")))
			if err != nil {
				return err
			}
			if existing == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			if !canAccessWorkItem(c, t, existing) && !isInvited(existing, t.Email) && !isParentScoped(c, t, existing) {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			item.Set("hostingId", existing.V("hostingId"))
			item.Set("groupKey", existing.V("groupKey"))
			r, err := updateWorkItem(ctx, c, q, jsx.Param(item.V("workId")), item, false)
			if err != nil {
				return err
			}
			scbOne(c, r)
		case "DELETE":
			if !idOK || id == 0 {
				c.Ecb("workId가 필요합니다.")
				return nil
			}
			existing, err := getWorkItem(ctx, q, id)
			if err != nil {
				return err
			}
			if existing == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			if !canAccessWorkItem(c, t, existing) {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			if _, err := deleteWorkItem(ctx, c, q, id); err != nil {
				return err
			}
			c.Scb("DELETED")
		default:
			c.Ecb("잘못된 요청")
		}
		return nil
	})
}

// workShapeNoti — 쉐이프 멘션·좋아요·답글 알림(보낸 표시를 쉐이프에 남긴다)
func workShapeNoti(ctx context.Context, c *web.Ctx, q dbq, o *jsx.OrderedMap) error {
	cid := orStr(o.V("hostingId"))
	ms, err := adminMembers(ctx, q, cid)
	if err != nil {
		return err
	}
	js := o.Map("Json")
	groupKey := tplU(o, "groupKey")
	title := tplU(js, "title")
	findAdmin := func(id any) *jsx.OrderedMap {
		for _, m := range ms {
			if looseEq(m.V("adminId"), id) {
				return m
			}
		}
		return nil
	}
	shapeURL := func(shape *jsx.OrderedMap, subIdx any) string {
		var params []string
		if id := shape.V("id"); id != nil && jsx.Str(id) != "" {
			params = append(params, "shapeId="+jsx.Str(id))
		}
		if subIdx != nil {
			params = append(params, "subImageIndex="+jsx.Str(subIdx))
		}
		qs := ""
		if len(params) > 0 {
			qs = "?" + strings.Join(params, "&")
		}
		return "/admin/work/" + groupKey + "/" + tplU(o, "workId") + qs
	}
	checkShapes := func(shapes []any, subIdx any) {
		for _, s := range shapes {
			shape := asMap(s)
			list, _ := shape.V("mentionedMembers").([]any)
			for _, x := range list {
				m := asMap(x)
				if jsx.Truthy(m.V("notiAt")) {
					continue
				}
				noti := pushData("[멘션] "+groupKey+" "+title, titleOr(shape.V("title")), shapeURL(shape, subIdx))
				if notifyAdmin(ctx, c, q, cid, noti, m.V("adminId")) {
					m.Set("notiAt", nowMs())
				}
			}
		}
	}
	checkLikes := func(shapes []any, subIdx any) {
		for _, s := range shapes {
			shape := asMap(s)
			likeIDs, _ := shape.V("likeAdminIds").([]any)
			if !jsx.Truthy(shape.V("likeNotiMap")) {
				shape.Set("likeNotiMap", jsx.NewOrderedMap())
			}
			lm := shape.Map("likeNotiMap")
			for _, k := range lm.Keys() {
				n := jsx.Number(k)
				if !includesNum(likeIDs, n) {
					lm.Delete(jsx.NumStr(n))
				}
			}
			for _, a := range likeIDs {
				if jsx.Truthy(lm.V(jsx.Str(a))) {
					continue
				}
				admin := findAdmin(a)
				if admin == nil {
					continue
				}
				noti := pushData("[반응] "+groupKey+" "+title, tplStr(admin.V("name"))+"님이 좋아요를 눌렀습니다.", shapeURL(shape, subIdx))
				owner := shape.V("adminId")
				if !jsx.Truthy(owner) {
					owner = float64(0)
				}
				ok := true
				if jsx.Number(owner) > 0 && !strictEq(a, owner) {
					ok = notifyAdmin(ctx, c, q, cid, noti, owner)
				}
				if ok && lm != nil {
					lm.Set(jsx.Str(a), nowMs())
				}
			}
		}
	}
	checkParent := func(shapes []any, subIdx any) {
		for _, s := range shapes {
			shape := asMap(s)
			if !(jsx.Truthy(shape.V("parentShapeId")) && jsx.Truthy(shape.V("activeEditorName")) && jsx.Truthy(shape.V("html")) && !jsx.Truthy(shape.V("parentShapeNotiTs"))) {
				continue
			}
			var all []any
			all = append(all, js.Arr("shapes")...)
			for _, img := range js.Arr("images") {
				all = append(all, asMap(img).Arr("shapes")...)
			}
			var parent *jsx.OrderedMap
			for _, x := range all {
				if jsx.Str(asMap(x).V("id")) == jsx.Str(shape.V("parentShapeId")) {
					parent = asMap(x)
					break
				}
			}
			if parent != nil && jsx.Truthy(parent.V("adminId")) && jsx.Truthy(shape.V("adminId")) && !strictEq(parent.V("adminId"), shape.V("adminId")) {
				admin := findAdmin(shape.V("adminId"))
				if admin == nil {
					shape.Set("parentShapeNotiTs", nowMs())
					continue
				}
				noti := pushData("[답글] "+groupKey+" "+title, tplStr(admin.V("name"))+" | "+titleOr(parent.V("title"))+" 의 답글", shapeURL(shape, subIdx))
				if notifyAdmin(ctx, c, q, cid, noti, parent.V("adminId")) {
					shape.Set("parentShapeNotiTs", nowMs())
				}
			} else {
				shape.Set("parentShapeNotiTs", nowMs())
			}
		}
	}
	if shapes, ok := js.V("shapes").([]any); ok {
		checkShapes(shapes, nil)
		checkLikes(shapes, nil)
		checkParent(shapes, nil)
	}
	for i, img := range js.Arr("images") {
		shapes := asMap(img).Arr("shapes")
		if len(shapes) == 0 {
			continue
		}
		checkShapes(shapes, float64(i))
		checkLikes(shapes, float64(i))
		checkParent(shapes, float64(i))
	}
	return nil
}

func workShape(c *web.Ctx) {
	withConn(c, func(ctx context.Context, q dbqConn) error {
		t := c.Token
		id, idOK := workIDParam(c)
		shapeID := c.QV("shapeId")
		shape := bodyMap(c)
		if !jsx.Truthy(shape.V("id")) {
			c.Ecb("잘못된 요청")
			return nil
		}
		switch c.Method {
		case "PUT":
			var idArg any = id
			if !idOK {
				return errNaN
			}
			item, err := getWorkItem(ctx, q, idArg)
			if err != nil {
				return err
			}
			if item == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			js := item.Map("Json")
			if js == nil {
				return errMsg("TypeError: Cannot read properties of null (reading 'shapes')")
			}
			if !js.Has("shapes") {
				js.Set("shapes", []any{})
			}
			isOwner := looseEq(item.V("hostingId"), t.HostingID)
			isShared := !isOwner && canAccessWorkItem(c, t, item)
			invited := !isOwner && !isShared && isInvited(item, t.Email)
			if !isOwner && !isShared && !invited {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			if invited {
				if !jsx.Truthy(js.V("invitedAdmins")) {
					js.Set("invitedAdmins", []any{})
				}
				list, _ := js.V("invitedAdmins").([]any)
				found := false
				for _, x := range list {
					if looseEq(asMap(x).V("adminId"), t.Key) {
						found = true
						break
					}
				}
				if !found {
					all, err := adminMembers(ctx, q, "")
					if err != nil {
						return err
					}
					var admin *jsx.OrderedMap
					for _, m := range all {
						if looseEq(m.V("adminId"), jsx.Number(t.Key)) {
							admin = m
							break
						}
					}
					if admin == nil {
						return errMsg("TypeError: Cannot read properties of undefined (reading 'json')")
					}
					var x *jsx.OrderedMap
					if s, ok := admin.V("json").(string); ok {
						v, err := jsx.ParseString(s)
						if err != nil {
							return err
						}
						x = asMap(v)
					}
					profile := ""
					if n := x.Path("image", "name"); jsx.Truthy(n) {
						profile = imageURLPrefix(c, t.HostingID) + "/admin_profile_" + jsx.Str(admin.V("adminId")) + "/" + jsx.Str(n)
					}
					js.Set("invitedAdmins", append(list, jsx.Obj("adminId", jsx.Number(t.Key), "email", t.Email, "name", t.Name, "profileImage", profile)))
				}
			}
			var subIdx any
			if shape.Has("subImageIndex") {
				subIdx = shape.V("subImageIndex")
			}
			if !js.Has("images") {
				js.Set("images", []any{})
			}
			var target *jsx.OrderedMap
			if subIdx == nil {
				target = js
			} else {
				imgs := js.Arr("images")
				i := int(jsx.Number(subIdx))
				if f := jsx.Number(subIdx); f == float64(i) && i >= 0 && i < len(imgs) {
					target = asMap(imgs[i])
				}
			}
			if target == nil {
				return errMsg("TypeError: Cannot read properties of undefined (reading 'shapes')")
			}
			if !target.Has("shapes") {
				target.Set("shapes", []any{})
			}
			isNew := true
			next := []any{}
			for _, s := range target.Arr("shapes") {
				if strictEq(asMap(s).V("id"), shapeID) {
					isNew = false
					if jsx.Truthy(shape.V("isDeleted")) {
						continue
					}
					next = append(next, shape)
				} else {
					next = append(next, s)
				}
			}
			if isNew {
				next = append(next, shape)
			}
			target.Set("shapes", next)

			if err := workShapeNoti(ctx, c, q, item); err != nil {
				return err
			}
			r, err := updateWorkItem(ctx, c, q, item.V("workId"), item, true)
			if err != nil {
				return err
			}
			scbOne(c, r)
			emitWork(c, jsx.Obj("hostingId", item.V("hostingId"), "workId", idArg, "zbduid", browserUID(t), "oxuid", browserUID(t) /* 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다 */, "shape", shape))
		case "DELETE":
			if !idOK || id == 0 {
				c.Ecb("workId가 필요합니다.")
				return nil
			}
			existing, err := getWorkItem(ctx, q, id)
			if err != nil {
				return err
			}
			if existing == nil {
				c.Ecb("업무관리를 찾을 수 없습니다.")
				return nil
			}
			isOwner := looseEq(existing.V("hostingId"), t.HostingID)
			isShared := !isOwner && canAccessWorkItem(c, t, existing)
			invited := !isOwner && !isShared && isInvited(existing, t.Email)
			if !isOwner && !isShared && !invited {
				c.Ecb("이 업무관리에 대한 권한이 없습니다.")
				return nil
			}
			if _, err := deleteWorkItem(ctx, c, q, id); err != nil {
				return err
			}
			c.Scb("DELETED")
		default:
			c.Ecb("잘못된 요청")
		}
		return nil
	})
}

// 여러 업무관리를 한 번에(정렬 번호 등) — 항목마다 권한을 본다
func workQuickUpdate(c *web.Ctx) {
	bd := bodyMap(c)
	if !jsx.Truthy(bd.V("actionType")) || !jsx.Truthy(bd.V("workIds")) {
		c.Ecb("잘못된 요청")
		return
	}
	t := c.Token
	ids := bd.V("workIds")
	actionType := jsx.Str(bd.V("actionType"))
	idList, _ := ids.([]any)
	q, err := conn(c)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer q.Close()
	ctx := c.Context()
	os, err := getWorkItems(ctx, q, workSearch{workIDs: ids, hostingIDs: []string{"", t.HostingID}})
	if err != nil {
		c.Ecb(jsErrText(err))
		return
	}
	indexOf := func(v any) int {
		for i, x := range idList {
			if strictEq(x, v) {
				return i
			}
		}
		return -1
	}
	sortStable(os, func(a, b *jsx.OrderedMap) bool { return indexOf(a.V("workId")) < indexOf(b.V("workId")) })
	if _, err := q.ExecContext(ctx, "START TRANSACTION"); err != nil {
		c.Ecb(jsErrText(err))
		return
	}
	for i, o := range os {
		if !canAccessWorkItem(c, t, o) {
			continue
		}
		if strings.Contains(actionType, "sortNo") {
			if js := o.Map("Json"); js != nil {
				js.Set(actionType, float64(i))
			}
		}
		x, err := updateWorkItem(ctx, c, q, o.V("workId"), o, true)
		if err != nil {
			_, _ = q.ExecContext(ctx, "ROLLBACK")
			c.Ecb(jsErrText(err))
			return
		}
		if x == nil || !jsx.Truthy(x.V("workId")) {
			_, _ = q.ExecContext(ctx, "ROLLBACK")
			c.Ecb("업무관리 " + jsx.Str(o.V("workId")) + " 업데이트 실패")
			return
		}
	}
	if _, err := q.ExecContext(ctx, "COMMIT"); err != nil {
		c.Ecb(jsErrText(err))
		return
	}
	for _, bc := range uniqStr([]string{"", t.HostingID}) {
		emitWork(c, jsx.Obj("hostingId", bc, "workIds", ids, "zbduid", browserUID(t), "oxuid", browserUID(t) /* 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다 */))
	}
	c.Scb(true)
}

func uniqStr(a []string) []string {
	var out []string
	for _, s := range a {
		if !contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

// sortStable — Array.prototype.sort(안정 정렬)
func sortStable(a []*jsx.OrderedMap, less func(x, y *jsx.OrderedMap) bool) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && less(a[j], a[j-1]); j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// imageURLPrefix — getImageUrlPrefix(hostingId) = CDN_URL_PREFIX + "/" + 회사 + "/"
func imageURLPrefix(c *web.Ctx, hostingID string) string {
	p := c.App.Cfg.Get("CDN_URL_PREFIX")
	if hostingID == "" {
		return p
	}
	return p + "/" + hostingID + "/"
}
