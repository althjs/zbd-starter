package site

// 사본: server-libs/utils.ts (decryptNumber) · pages/api/share/constants.ts (getShareItem)
//       · pages/api/admin/post/[postId].ts (getPost) · server-libs/freecell-card.ts — 전환 뒤 합칠 것
// (Go 쪽 원본은 internal/api/content/share.go · admincontent/post.go · apps/freecell.go — 고치면 같이)

import (
	"context"
	"regexp"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

// getPost — 글 하나
func getPost(ctx context.Context, q db.Querier, postID, hostingID string) (*jsx.OrderedMap, error) {
	return db.QueryOne(ctx, q, `select postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser,
          status, UNIX_TIMESTAMP(createdAt)*1000 as createdAt,  UNIX_TIMESTAMP(updatedAt)*1000 as updatedAt, hostingId
          FROM posts where postId = ? AND hostingId = ?`, postID, hostingID)
}

// ─── 프리셀 도전장·판 번호별 페이지 ───────────────────

var fcPostIDRule = regexp.MustCompile(`^[a-z0-9]{6,12}$`)

type fcRecord struct {
	v, gameNumber float64
	isWon         bool
	dealType      string
	sec           any
	completed     any
}

type fcCard struct {
	nickname   string
	gameNumber float64
	sec        float64
}

const wonMsWhere = `hostingId = ? AND groupKey = 'FREECELL' AND status = 0
  AND JSON_EXTRACT(jsonUser, '$.v') = 2
  AND JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.dealType')) = 'ms'
  AND (JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.result')) = 'won' OR JSON_UNQUOTE(JSON_EXTRACT(jsonUser, '$.isWon')) = 'true')`

type fcTop struct {
	nickname string
	sec      float64
}

type fcGame struct {
	gameNumber float64
	solvers    int
	bestSec    float64
	top        []fcTop
}

type fcSolved struct{ gameNumber, at float64 }
