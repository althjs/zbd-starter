package core

import (
	"context"
	"os"
	"path/filepath"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// 탈퇴 회원 정리 — pages/api/admin/user/remove-user-data
//   HIDE_USER_PROFILE: 프로필 사진 지우고 글·댓글의 이름을 「탈퇴회원」 으로 / DELETE_ALL: 글 지우고 댓글 숨김

// deleteUserProfileImage — 회원 프로필 사진 파일
func deleteUserProfileImage(c *web.Ctx, hostingID, userID string) {
	dir := c.App.Cfg.Get("CDN_UPLOAD_DIR")
	if dir == "" {
		return
	}
	_ = os.Remove(filepath.Join(dir, hostingID, "u_"+userID+"_my", "profile.webp"))
}

// postGroupUsesES — 그 게시판이 검색엔진 색인을 쓰는가(사용자 게시판 → 관리자 게시판 순)
func postGroupUsesES(c *web.Ctx, hostingID string, groupKey any) bool {
	gs, _ := c.App.Settings.Config(hostingID).V("POST_GROUPS").(*jsx.OrderedMap)
	for _, list := range []string{"userPostGroups", "adminPostGroups"} {
		for _, x := range gs.Arr(list) {
			if g, _ := x.(*jsx.OrderedMap); g != nil && looseEq(g.V("groupKey"), groupKey) {
				return jsx.Truthy(g.Path("feature", "useES"))
			}
		}
	}
	return false
}

func userUCC(ctx context.Context, q db.Querier, hostingID, userID string) (posts, comments []*jsx.OrderedMap, err error) {
	posts, err = query(ctx, q, `SELECT * from posts WHERE hostingId = ? AND userId = ?`, hostingID, userID)
	if err != nil {
		return
	}
	return // 댓글은 고급 게시판 플러그인 — 스타터엔 없다(comments 는 늘 비어 있다)
}

// adminRemoveUserData — DELETE ?actionType=HIDE_USER_PROFILE|DELETE_ALL&userId=
func adminRemoveUserData(c *web.Ctx) {
	ctx := c.Context()
	action := c.Q("actionType")
	uid := c.Q("userId")
	if action == "" || uid == "" {
		c.Ecb("잘못된 요청")
		return
	}
	if c.Method != "DELETE" {
		c.Ecb("잘못된 요청")
		return
	}
	if action != "HIDE_USER_PROFILE" && action != "DELETE_ALL" {
		c.Ecb("잘못된 요청")
		return
	}
	cid := c.Token.HostingID
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	deleteUserProfileImage(c, cid, uid)
	posts, comments, err := userUCC(ctx, conn, cid, uid)
	if err != nil {
		c.Ecb(err)
		return
	}
	postCount, commentCount := 0.0, 0.0
	if action == "HIDE_USER_PROFILE" {
		for _, p := range posts {
			ju, err := parseJSON(p.V("jsonUser"), "{}")
			if err != nil {
				c.Ecb(err)
				return
			}
			jm, _ := ju.(*jsx.OrderedMap)
			p.Set("jsonUser", ju)
			if jm == nil || !jsx.Truthy(jm.V("name")) {
				continue
			}
			jm.Set("name", "탈퇴회원")
			o, err := db.Exec(ctx, conn, `UPDATE posts SET jsonUser = ? WHERE postId = ? AND hostingId = ?`, jsx.Stringify(jm), p.V("postId"), p.V("hostingId"))
			if err != nil {
				c.Ecb(err)
				return
			}
			if postGroupUsesES(c, cid, p.V("groupKey")) {
				esPutPost(c, p)
			}
			postCount += jsx.NumOr(o.V("affectedRows"), 0)
		}
		for _, cm := range comments {
			o, err := db.Exec(ctx, conn, `UPDATE post_comments SET userNickname = ? WHERE commentId = ? AND hostingId = ?`, "탈퇴회원", cm.V("commentId"), cm.V("hostingId"))
			if err != nil {
				c.Ecb(err)
				return
			}
			commentCount += jsx.NumOr(o.V("affectedRows"), 0)
		}
	} else {
		// 지운 글 보관 표는 없어졌다(지금 서비스도 넣기를 시도하고 실패를 무시한다) — 바로 지운다
		for _, p := range posts {
			o, err := db.Exec(ctx, conn, `DELETE FROM posts WHERE postId = ? AND hostingId = ?`, p.V("postId"), p.V("hostingId"))
			if err != nil {
				c.Ecb(err)
				return
			}
			if postGroupUsesES(c, cid, p.V("groupKey")) {
				esDeletePost(c, jsx.Str(p.V("postId")))
			}
			postCount += jsx.NumOr(o.V("affectedRows"), 0)
		}
		for _, cm := range comments {
			o, err := db.Exec(ctx, conn, `UPDATE post_comments SET status = 99, updatedAt = NOW() 
  WHERE commentId = ?`, cm.V("commentId"))
			if err != nil {
				c.Ecb(err)
				return
			}
			commentCount += jsx.NumOr(o.V("affectedRows"), 0)
		}
	}
	c.Scb(jsx.Obj("postCount", postCount, "commentCount", commentCount))
}
