package content

// 글 반응(좋아요·별점·신고)·댓글 수·댓글 검색 자리 — 플러그인 「고급 게시판」(zbd add advanced-board)이 이 파일을 운영판으로 갈아 끼운다.
// 스타터는 반응·댓글 표가 없어 늘 빈 값. 함수 모양은 운영판과 같게 둔다.

import (
	"context"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

// doCommentSearchPostIDs — 댓글 본문 검색으로 글 id. 스타터는 늘 빈 목록
func doCommentSearchPostIDs(ctx context.Context, q db.Querier, hostingID, keyword string) ([]any, error) {
	return []any{}, nil
}

// postReactionSummary — getPostSummaryMap 의 반응·댓글 수·신고 부분. 스타터는 빈 표
func postReactionSummary(ctx context.Context, q db.Querier, hostingID string, postIDs []any, in string, args []any, isAdmin bool) (m, cMap, reportMap *jsx.OrderedMap, err error) {
	return jsx.NewOrderedMap(), jsx.NewOrderedMap(), jsx.Obj("isPostReported", false, "reportedCommentIdMap", jsx.NewOrderedMap()), nil
}

// userCommentCount — 회원이 쓴 댓글 수. 스타터는 0
func userCommentCount(ctx context.Context, q db.Querier, hostingID string, userID any) (any, error) {
	return 0.0, nil
}
