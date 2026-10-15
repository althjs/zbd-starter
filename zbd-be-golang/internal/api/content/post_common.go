package content

// 글(posts) 공용 조회 — 게시판·댓글·사용자 글 라우트가 같이 쓴다.
// 사본: pages/api/admin/post/constants.ts · admin/post/[postId].ts 의 일부 — 전환 뒤 합칠 것

import (
	"context"
	"math"
	"regexp"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
)

// getPostAdmin — 글 하나(json·jsonUser·hostingId 포함). hostingID 가 있으면 그 회사 것만.
func getPostAdmin(ctx context.Context, q db.Querier, postID any, hostingID string) (*jsx.OrderedMap, error) {
	query := `select postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt,  UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
          FROM posts where postId = ?`
	args := []any{postID}
	if hostingID != "" {
		query += ` AND hostingId = ?`
		args = append(args, hostingID)
	}
	return db.QueryOne(ctx, q, query, args...)
}

// getPostByPostIDs — 글 여러 개(회사 안에서)
func getPostByPostIDs(ctx context.Context, q db.Querier, hostingID string, postIDs []any) ([]*jsx.OrderedMap, error) {
	if len(postIDs) == 0 {
		return []*jsx.OrderedMap{}, nil
	}
	return db.Query(ctx, q, `SELECT postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
          FROM posts WHERE hostingId = ? AND postId IN (?)`, hostingID, postIDs)
}

// countUpPost — 글 조회수 +1(시간 단위 행)
func countUpPost(ctx context.Context, q db.Execer, postID any, hostingID string) (*jsx.OrderedMap, error) {
	ymdh := util.YMD(time.Now(), true, false)[:13] + ":00"
	return db.Exec(ctx, q, `INSERT INTO post_views (postId, pv, groupKey, hostingId)
VALUES (?, 1, ?, ?)
ON DUPLICATE KEY UPDATE pv = pv+1;`, postID, ymdh, hostingID)
}

// postSearch — getAllGroupPost 의 조건
type postSearch struct {
	groupKey    string
	groupKeys   []any
	groupSubKey string
	userID      string // 숫자 문자열 — 있으면 그 회원의 글만
	status      *float64
	pageNo      float64
	pageSize    float64
	orderBy     string
	orderByDesc string
}

// getAllGroupPost — 게시판 글(쪽 나누기를 주면 {posts, totalCount})
func getAllGroupPost(ctx context.Context, q db.Querier, hostingID string, s postSearch) ([]*jsx.OrderedMap, any, error) {
	from := ` FROM posts WHERE 1 `
	args := []any{}
	if hostingID != "" {
		if s.groupKey == "TIP" {
			from = ` FROM posts WHERE 1  AND (hostingId = ? OR postId in (?))  `
			args = append(args, hostingID)
			tips, err := db.Query(ctx, q, `select postId from posts where hostingId = "" AND groupKey = "TIP"`)
			if err != nil {
				return nil, nil, err
			}
			ids := []any{}
			for _, t := range tips {
				ids = append(ids, t.V("postId"))
			}
			args = append(args, ids)
		} else {
			from = ` FROM posts WHERE 1  AND (hostingId = ? )  `
			args = append(args, hostingID)
		}
	}
	if s.status != nil {
		from += ` AND status = ` + jsx.NumStr(*s.status)
	}
	if s.groupKey != "" {
		from += ` AND groupKey = ?`
		args = append(args, s.groupKey)
	} else if len(s.groupKeys) > 0 {
		from += ` AND groupKey in (?)`
		args = append(args, s.groupKeys)
	}
	if s.groupSubKey != "" {
		from += ` AND groupSubKey = ?`
		args = append(args, s.groupSubKey)
	}
	if s.userID != "" {
		from += ` AND userId = ?`
		args = append(args, s.userID)
	}
	paged := s.pageNo != 0 && s.pageSize != 0 && !math.IsNaN(s.pageNo) && !math.IsNaN(s.pageSize)
	var total any
	limit, orderBy := "", ""
	if paged {
		r, err := db.QueryOne(ctx, q, `SELECT count(*) as cnt `+from, args...)
		if err != nil {
			return nil, nil, err
		}
		total = r.V("cnt")
		limit = "LIMIT " + jsx.NumStr((s.pageNo-1)*s.pageSize) + "," + jsx.NumStr(s.pageSize)
		// 지금 서비스는 주소 값을 그대로 SQL 에 붙였다(끼워 넣기 가능) — 이름 모양만 받는다
		if !sqlIdentRe.MatchString(s.orderBy) || !sqlDirRe.MatchString(s.orderByDesc) {
			return nil, nil, errString("잘못된 정렬 조건")
		}
		orderBy = "ORDER BY " + s.orderBy + " " + s.orderByDesc
	}
	rows, err := db.Query(ctx, q, `select postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt, UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
          `+from+` `+orderBy+`
          `+limit, args...)
	if err != nil {
		return nil, nil, err
	}
	return rows, total, nil
}

var (
	sqlIdentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)
	sqlDirRe   = regexp.MustCompile(`^(?i:asc|desc)$`)
)

// getUserIDToNicknameMap — userId → {nickname, image}
func getUserIDToNicknameMap(ctx context.Context, q db.Querier, userIDs []any) (map[string]*jsx.OrderedMap, error) {
	out := map[string]*jsx.OrderedMap{}
	if len(userIDs) == 0 {
		return out, nil
	}
	seen := map[string]bool{}
	uniq := []any{}
	for _, u := range userIDs {
		k := jsx.Str(u)
		if !seen[k] {
			seen[k] = true
			uniq = append(uniq, jsx.Param(u))
		}
	}
	rows, err := db.Query(ctx, q, `SELECT userId, nickname, image FROM members WHERE userId IN (?)`, uniq)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[jsx.Str(r.V("userId"))] = jsx.Obj("nickname", r.V("nickname"), "image", r.V("image"))
	}
	return out, nil
}

// getPostSummaryMap — 글별 반응(좋아요·별점·신고)·조회수·댓글 수. 관리자 + 글 하나면 신고 내역까지(반응·댓글은 post_reactions.go)
func getPostSummaryMap(ctx context.Context, q db.Querier, hostingID string, postIDs []any, isAdmin bool) (*jsx.OrderedMap, error) {
	in := ""
	args := []any{hostingID}
	if len(postIDs) > 0 {
		in = ` AND postId IN (?)`
		args = append(args, postIDs)
	}
	m, cMap, reportMap, err := postReactionSummary(ctx, q, hostingID, postIDs, in, args, isAdmin)
	if err != nil {
		return nil, err
	}
	pvRows, err := db.Query(ctx, q, `SELECT postId, sum(pv) as pv FROM post_views WHERE hostingId = ? `+in+` group by postId;`, args...)
	if err != nil {
		return nil, err
	}
	pvMap := jsx.NewOrderedMap()
	for _, r := range pvRows {
		pvMap.Set(jsx.Str(r.V("postId")), orVal(decNum(r.V("pv")), 0.0))
	}
	out := jsx.Obj("reactionMap", m, "pvMap", pvMap, "commentMap", cMap)
	if isAdmin {
		out.Set("reportMap", reportMap)
	}
	return out, nil
}

// postIDsOf — 행들의 postId
func postIDsOf(rows []*jsx.OrderedMap) []any {
	out := []any{}
	for _, r := range rows {
		out = append(out, r.V("postId"))
	}
	return out
}

// parseJSONField — JSON.parse(s || "{}") (실패하면 오류)
func parseJSONField(v any) (any, error) {
	return jsx.ParseString(orStr(v, "{}"))
}

// trimJS — JS 의 String.prototype.trim (공백·줄바꿈·BOM 등)
func trimJS(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r', 0xA0, 0xFEFF, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000:
			return true
		}
		return r >= 0x2000 && r <= 0x200A
	})
}

// decNum — DECIMAL 값(SUM 등)을 숫자로. 지금 서비스 드라이버(mysql, supportBigNumbers 없음)는 DECIMAL 을 숫자로 준다.
func decNum(v any) any {
	if s, ok := v.(string); ok {
		return jsx.Number(s)
	}
	return v
}
