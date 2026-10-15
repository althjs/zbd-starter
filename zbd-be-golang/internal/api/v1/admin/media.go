package admin

// 미디어·기기 목록
//
//	GET /api/v1/admin/files/[folder]?preFolder&subFolder&q&page&size&sort=-mtime           (2-13, 업로드 폴더의 파일)
//	GET /api/v1/admin/login-devices?tab=U|A&from&to&q&page&size&sort=-lastLoginedAt        (3-2)

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"zbd.kr/be/internal/api/v1/v1x"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

var ytVideoSort = map[string]string{"publishedAt": "publishedAt", "viewCount": "viewCount", "likeCount": "likeCount", "commentCount": "commentCount", "videoId": "videoId", "title": "title"}

var ytCommentSort = map[string]string{"timeAt": "timeAt", "likes": "likes", "author": "author", "commentId": "commentId"}

var videoSort = map[string]string{"videoId": "videoId", "createdAt": "createdAt", "updatedAt": "updatedAt"}

var reFolderPart = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

var fileSort = map[string]string{"mtime": "", "name": "", "size": ""}

// files — 업로드 폴더(회사/preFolder/folder/subFolder)의 파일. 폴더 이름은 영문·숫자·_·- 만(.. 으로 밖을 못 본다).
func files(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, fileSort, "-mtime", "")
	if !ok {
		return
	}
	parts := []string{c.Token.HostingID}
	for _, k := range []string{"preFolder", "folder", "subFolder"} {
		v := c.Q(k)
		if v == "" {
			if k == "folder" {
				c.Fail(400, web.ErrBadRequest, "folder 가 필요합니다")
				return
			}
			continue
		}
		if !reFolderPart.MatchString(v) {
			c.Fail(400, web.ErrBadRequest, k+" 는 영문·숫자·_·- 만")
			return
		}
		parts = append(parts, v)
	}
	base := c.App.Cfg.Get("CDN_UPLOAD_DIR")
	if base == "" {
		c.Fail(503, web.ErrFailed, "업로드 폴더가 설정되지 않았습니다")
		return
	}
	dir := filepath.Join(append([]string{base}, parts...)...)
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.Fail(404, web.ErrNotFound, "없는 폴더입니다")
			return
		}
		c.Ecb(err)
		return
	}
	type fi struct {
		name  string
		size  int64
		mtime int64
	}
	q := strings.ToLower(lq.Q)
	var list []fi
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, ".") || strings.Contains(n, "400.jpg") {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(n), q) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		list = append(list, fi{n, info.Size(), info.ModTime().UnixMilli()})
	}
	// 파일 시스템이라 SQL 대신 여기서 정렬·자른다(폴더 하나 단위)
	key, desc := "mtime", true
	if s := c.Q("sort"); s != "" {
		f := strings.Split(s, ",")[0]
		desc = strings.HasPrefix(f, "-")
		key = strings.TrimPrefix(f, "-")
	}
	cmp := func(a, b fi) int {
		var d int64
		switch key {
		case "name":
			return strings.Compare(a.name, b.name)
		case "size":
			d = a.size - b.size
		default:
			d = a.mtime - b.mtime
		}
		switch {
		case d < 0:
			return -1
		case d > 0:
			return 1
		}
		return strings.Compare(a.name, b.name) // 같으면 이름으로(쪽 사이 순서 고정)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if desc {
			return cmp(list[i], list[j]) > 0
		}
		return cmp(list[i], list[j]) < 0
	})
	total := int64(len(list))
	items := []any{}
	for i := lq.Offset(); i < len(list) && i < lq.Offset()+lq.Size; i++ {
		f := list[i]
		items = append(items, jsx.Obj("name", f.name, "size", f.size, "mtime", f.mtime))
	}
	c.SendList(items, total, lq)
}

var deviceSort = map[string]string{"lastLoginedAt": "d.lastLoginedAt", "createdAt": "d.createdAt", "userId": "d.userId"}

// loginDevices — 로그인 기기. tab=U(회원, 기본) | A(관리자). q: 회원 = 개인정보 규칙(v1x.MemberSearch), 관리자 = 번호.
func loginDevices(c *web.Ctx) {
	if !v1x.OnlyGET(c) {
		return
	}
	lq, ok := v1x.List(c, deviceSort, "-lastLoginedAt", "d.browserUid DESC, d.userId")
	if !ok {
		return
	}
	tab := c.Q("tab")
	if tab == "" {
		tab = "U"
	}
	if tab != "U" && tab != "A" {
		c.Fail(400, web.ErrBadRequest, "tab 은 U|A")
		return
	}
	cid := c.Token.HostingID
	w := &v1x.Where{}
	w.Add("d.hostingId = ?", cid)
	w.Add("d.memberType = ?", tab)
	if !v1x.DateRange(c, w, "d.lastLoginedAt") {
		return
	}
	if lq.Q != "" {
		if tab == "U" {
			mw := &v1x.Where{}
			v1x.MemberSearch(mw, c.App.PII, "u.", lq.Q)
			w.Add(strings.TrimPrefix(mw.SQL(), " WHERE "), mw.Args...)
		} else {
			w.Add("d.userId = ?", lq.Q)
		}
	}
	from := " FROM login_devices d LEFT JOIN members u ON d.memberType = 'U' AND d.userId = u.userId AND d.hostingId = u.hostingId"
	ctx := c.Context()
	rows, total, err := v1x.Page(ctx, c.App.DB, `SELECT d.userId, d.memberType, d.browserUid, d.ip,
  UNIX_TIMESTAMP(d.createdAt)*1000 AS createdAt, UNIX_TIMESTAMP(d.lastLoginedAt)*1000 AS lastLoginedAt,
  u.email, u.nickname, u.name, u.groupId`, from, w, lq)
	if err != nil {
		c.Ecb(err)
		return
	}
	for _, r := range rows {
		if tab == "A" {
			m := c.App.Admin.Member(jsx.Str(r.V("userId")))
			r.Set("email", v1x.OrEmpty(m.V("email")))
			r.Set("name", v1x.OrEmpty(m.V("name")))
			r.Set("nickname", "")
			r.Set("groupId", nil)
		} else {
			c.App.PII.DecMember(r)
		}
	}
	u, err := db.QueryOne(ctx, c.App.DB, "SELECT COUNT(DISTINCT d.userId) AS n"+from+w.SQL(), w.Args...)
	if err != nil {
		c.Ecb(err)
		return
	}
	c.SendListSummary(v1x.Any(rows), total, lq, jsx.Obj("uniqUserIdCount", jsx.Number(u.V("n"))))
}
