package apps

// 회원 파일 — 내 폴더 목록·지우기 · 내려받기 · 올리기(사진·동영상) · 보안 파일 올리기/받기/옮기기
//
// 올린 파일은 설정 CDN_UPLOAD_DIR 아래 {회사}/{폴더}/. 보안 파일은 그 옆 secure/ (nginx 로 열리지 않는다 — 받기는 늘 API).
// 경로는 지금 서비스처럼 문자열을 이어 만든다(설정값 끝의 / 유무까지 같게).

import (
	"context"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// nodeJoin — Node 의 path.join(하나): 정리하되 끝의 / 는 남긴다
func nodeJoin(p string) string {
	if p == "" {
		return "."
	}
	trail := strings.HasSuffix(p, "/")
	c := path.Clean(p)
	if trail && c != "/" {
		c += "/"
	}
	return c
}

// uploadDir — 지금 서비스의 UPLOAD_DIR. 늘 「/」 로 끝난다 — 부르는 쪽이 `uploadDir + 회사 + "/…"` 로 붙인다(설정에 끝 / 가 없으면 uploadZBD 가 됐다)
func uploadDir(a *app.App) string { return withSlash(nodeJoin(a.Cfg.Get("CDN_UPLOAD_DIR"))) }

func withSlash(p string) string {
	if strings.HasSuffix(p, "/") {
		return p
	}
	return p + "/"
}

// secureDir — UPLOAD_DIR/../secure/
func secureDir(a *app.App) string { return nodeJoin(uploadDir(a) + "/../secure/") }

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func mkdirAll(p string) { _ = os.MkdirAll(p, 0o755) }

// nodeErr — Node 의 fs 오류 문구(첫 줄) 모양: "Error: ENOENT: no such file or directory, unlink '…'"
func nodeErr(op, p string, err error) string {
	code, msg := "EIO", "i/o error"
	switch {
	case errors.Is(err, os.ErrNotExist):
		code, msg = "ENOENT", "no such file or directory"
	case errors.Is(err, os.ErrPermission):
		code, msg = "EACCES", "permission denied"
	case strings.Contains(err.Error(), "is a directory"):
		code, msg = "EISDIR", "illegal operation on a directory"
	case strings.Contains(err.Error(), "operation not permitted"):
		code, msg = "EPERM", "operation not permitted"
	}
	return "Error: " + code + ": " + msg + ", " + op + " '" + p + "'"
}

// insideDir — p 가 base 안인가(경로 정리 후)
func insideDir(base, p string) bool {
	b, _ := filepath.Abs(base)
	f, _ := filepath.Abs(p)
	return f == b || strings.HasPrefix(f, strings.TrimRight(b, "/")+"/")
}

// GET·DELETE /user/file/[folder]?name=&userId=&hostingId= — 회원 폴더(u_{회원}_{folder}) 파일 목록·지우기
func userFolder(c *web.Ctx) {
	cid := c.Token.HostingID
	if cid == "" {
		cid = jsx.StrOrEmpty(c.QV("hostingId"))
	}
	uid := jsx.StrOrEmpty(c.QV("userId"))
	if uid == "" {
		uid = tokenKey(c)
		if !c.Token.Valid() {
			uid = "undefined"
		}
	}
	folder := "u_" + uid + "_" + tmplU(c.QV("folder"), c.Query.Has("folder"))
	dir := uploadDir(c.App)
	if cid != "" {
		dir += cid + "/"
	}
	dir += folder + "/"
	if !exists(dir) {
		c.Ecb("존재하지 않는 경로입니다.")
		return
	}
	switch c.Method {
	case "GET":
		ents, err := os.ReadDir(dir)
		if err != nil {
			c.JSON(web.Ecb(nodeErr("scandir", dir, err)))
			return
		}
		files := []any{}
		for _, e := range ents {
			if !strings.Contains(e.Name(), "400.jpg") {
				files = append(files, e.Name())
			}
		}
		c.JSON(web.Scb(jsx.Obj("files", files), nil))
	case "DELETE":
		name := jsx.StrOrEmpty(c.QV("name"))
		if name == "" {
			c.Ecb("잘못된 요청")
			return
		}
		fp := nodeJoin(dir + name)
		// 보안: 폴더 밖 파일은 지우지 않는다(지금 서비스엔 이 검사가 없었다 — ../ 로 서버 파일을 지울 수 있었다)
		if !insideDir(dir, fp) {
			c.Ecb("잘못된 요청")
			return
		}
		if err := os.Remove(fp); err != nil {
			c.JSON(web.Ecb(nodeErr("unlink", fp, err)))
			return
		}
		if strings.HasSuffix(name, ".mp4") {
			n := name[:strings.LastIndex(name, ".")]
			for _, x := range []string{n + ".webp", n + "_share.jpg"} {
				p := nodeJoin(dir + x)
				if err := os.Remove(p); err != nil {
					c.JSON(web.Ecb(nodeErr("unlink", p, err)))
					return
				}
			}
		}
		c.Scb("DELETED")
	default:
		c.Ecb("잘못된 요청")
	}
}

// GET /user/file/download?name= — download/ 폴더의 파일 내려받기(폴더 밖은 막는다)
func userDownload(c *web.Ctx) {
	dir := nodeJoin(uploadDir(c.App) + "../download/")
	notFound := `<script>alert("NO FILE FOUND");window.close()</script>`
	if !exists(dir) {
		c.Send(notFound)
		return
	}
	name := jsx.StrOrEmpty(c.QV("name"))
	if name == "" {
		c.Status(204).Send(nil)
		return
	}
	fp := nodeJoin(dir + name)
	if !insideDir(dir, fp) {
		log.Printf("@@ user download path traversal blocked: %s", fp)
		c.Status(204).Send(nil)
		return
	}
	st, err := os.Stat(fp)
	if err != nil {
		c.Send(notFound)
		return
	}
	base := path.Base(name)
	ct := "application/octet-stream"
	if strings.ToLower(path.Ext(base)) == ".json" {
		ct = "application/json; charset=utf-8"
	}
	enc := encodeURIComponent(base)
	f, err := os.Open(fp)
	if err != nil {
		c.JSON(web.Ecb(nodeErr("open", fp, err)))
		return
	}
	defer f.Close()
	h := c.W.Header()
	h.Set("Content-Type", ct)
	h.Set("Content-Length", jsx.NumStr(float64(st.Size())))
	h.Set("Content-Disposition", `attachment; filename="`+enc+`"; filename*=UTF-8''`+enc)
	c.W.WriteHeader(200)
	_, _ = io.Copy(c.W, f)
}

// encodeURIComponent — JS 와 같은 규칙
func encodeURIComponent(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	for _, k := range []string{"!", "'", "(", ")", "*"} {
		e = strings.ReplaceAll(e, url.QueryEscape(k), k)
	}
	return e
}

// savedFile — 올린 파일 하나(지금 서비스 업로드 도구의 파일 정보와 같은 칸)
type savedFile struct {
	field, filepath, newFilename, originalFilename string
	mimetype                                       any
	size                                           int64
}

var errEmptyFile = errors.New("options.allowEmptyFiles is false, file size should be greater than 0")

// saveMultipart — 본문의 파일 칸을 dir 에 저장한다. nameFn(원래 이름) → 저장 이름. 폴더 밖이면 invalid-name
func saveMultipart(c *web.Ctx, dir string, nameFn func(string) string, maxTotal int64) ([]savedFile, error) {
	ct := c.Header("Content-Type")
	if !strings.Contains(strings.ToLower(ct), "multipart/form-data") {
		return nil, nil
	}
	mr, err := c.R.MultipartReader()
	if err != nil {
		return nil, err
	}
	var out []savedFile
	var total int64
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		if !hasFilenameParam(part) {
			_, _ = io.Copy(io.Discard, part)
			continue
		}
		orig := rawFileName(part)
		name := nameFn(orig)
		fp := path.Join(dir, name)
		if !strings.HasPrefix(fp, strings.TrimRight(dir, "/")) {
			fp = path.Join(dir, "invalid-name")
		}
		f, err := os.Create(fp)
		if err != nil {
			return out, err
		}
		n, err := io.Copy(f, io.LimitReader(part, maxTotal-total+1))
		f.Close()
		total += n
		if err != nil {
			return out, err
		}
		if total > maxTotal {
			return out, errors.New("options.maxTotalFileSize (" + jsx.NumStr(float64(maxTotal)) + " bytes) exceeded, received " + jsx.NumStr(float64(total)) + " bytes of file data")
		}
		if n == 0 {
			_ = os.Remove(fp)
			return out, errEmptyFile
		}
		var mt any
		if v := part.Header.Get("Content-Type"); v != "" {
			mt = v
		}
		out = append(out, savedFile{field: part.FormName(), filepath: fp, newFilename: path.Base(fp), originalFilename: orig, mimetype: mt, size: n})
	}
	return out, nil
}

func hasFilenameParam(p *multipart.Part) bool {
	_, ps, err := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
	_, ok := ps["filename"]
	return err == nil && ok
}

// rawFileName — 보낸 파일 이름 그대로(Go 의 FileName 은 경로를 떼어 내 지금 서비스와 달라진다)
func rawFileName(p *multipart.Part) string {
	_, ps, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
	return ps["filename"]
}

var ctlRule = regexp.MustCompile("[\x00-\x1f]")

// safeUploadName — 원래 파일명에서 경로 부분을 떼어 낸다(../ 로 폴더 밖에 저장되지 않게)
func safeUploadName(name string) string {
	parts := regexp.MustCompile(`[\\/]`).Split(name, -1)
	base := parts[len(parts)-1]
	clean := jsTrim(ctlRule.ReplaceAllString(base, ""))
	if clean == "" || clean == "." || clean == ".." {
		return "file"
	}
	return clean
}

func runCmd(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).Run()
}

// syncCDN — CDN 원본 저장소로 복사(설정 USE_CDN_UPLOAD=true · CDN_RCLONE_REMOTE 가 있을 때만)
func syncCDN(a *app.App, realDir string) {
	if a.Cfg.Get("USE_CDN_UPLOAD") != "true" {
		return
	}
	base := a.Cfg.Get("CDN_UPLOAD_DIR")
	remoteBase := a.Cfg.Get("CDN_RCLONE_REMOTE") // 예: 「원격:버킷/폴더/upload/」 — 이름이 바뀌므로 설정으로 받는다
	if !strings.Contains(realDir, "cdn/upload") || base == "" || remoteBase == "" {
		return
	}
	b, _ := filepath.Abs(base)
	d, _ := filepath.Abs(realDir)
	if d != b && !strings.HasPrefix(d, b+"/") {
		log.Printf("@ syncCDN 업로드 폴더 밖 경로 — 건너뜀: %s", realDir)
		return
	}
	rel, _ := filepath.Rel(b, d)
	if rel == "." {
		rel = ""
	}
	remote := strings.TrimRight(remoteBase, "/") + "/"
	if rel != "" {
		remote += rel + "/"
	}
	if err := runCmd("rclone", "copy", d+"/", remote, "--include", "*.{js,jpg,gif,png,webp,svg,jpeg,ico,pdf,json}"); err != nil {
		log.Printf("@@ syncCDN err: %v", err)
	}
}

// POST /user/file/upload?folder=&userId=(관리자만)&partnerType=&meta= — 회원 파일 올리기(동영상은 포스터 만들고 인코딩 대기열로)
func userUpload(c *web.Ctx) {
	if tokenKey(c) == "" {
		ecbCode(c, "로그인이 필요합니다.", "0401")
		return
	}
	cid := c.Token.HostingID
	if cid == "" {
		cid = jsx.StrOrEmpty(c.QV("hostingId"))
	}
	uid := tokenKey(c)
	if c.Token.MemberType != "U" {
		uid = jsx.StrOrEmpty(c.QV("userId"))
		if uid == "" {
			uid = tokenKey(c)
		}
	}
	fq := tmplU(c.QV("folder"), c.Query.Has("folder"))
	imageID := "u_" + uid + "_" + fq
	folder := imageID
	if pt := jsx.StrOrEmpty(c.QV("partnerType")); pt != "" {
		folder = pt + "_" + tokenKey(c) + "_" + fq
	}
	up := uploadDir(c.App)
	dir := up
	if cid != "" {
		dir += cid + "/"
	}
	dir += folder + "/"
	if cid != "" && !exists(up+cid) {
		mkdirAll(up + cid)
	}
	if !exists(dir) {
		mkdirAll(dir)
	}
	lastName := ""
	files, err := saveMultipart(c, dir, func(o string) string { lastName = safeUploadName(o); return lastName }, 1000*1024*1024)
	if err != nil {
		c.JSON(web.Ecb(err.Error()))
		return
	}
	site := c.App.Settings.Site(c.Token.HostingID)
	if jsx.Truthy(site.Path("feature", "useMobileImageUpload")) {
		for _, f := range files {
			if f.field != "files[]" || strings.HasSuffix(f.newFilename, "gif") {
				continue
			}
			if err := runCmd("convert", dir+f.newFilename, "-resize", "400x", "-quality", "75", "-strip", dir+f.newFilename+".400.jpg"); err != nil {
				log.Printf("모바일 이미지 변환 실패 %v", err)
				break
			}
		}
	}
	if jsx.Truthy(c.QV("meta")) {
		c.Send(web.Scb([]any{jsx.Obj("newFilename", lastName, "geo", exifGPS(dir+"/"+lastName))}, nil))
		return
	}
	out := []any{}
	for _, f := range files {
		if f.field != "files[]" {
			continue
		}
		// 동영상 인코딩(포스터·대기열)은 동영상 플러그인 — 스타터는 파일만 저장한다
		out = append(out, jsx.Obj("size", float64(f.size), "newFilename", f.newFilename, "mimetype", f.mimetype, "originalPathname", f.originalFilename))
	}
	syncCDN(c.App, dir)
	c.JSON(web.Scb(out, nil))
}

// ─── 보안 파일 ───

// parentHostingID — 토큰의 상위 회사(SITE.parentHostingId 가 있으면 그것, 아니면 자기 회사)
func parentHostingID(c *web.Ctx) string {
	own := c.Token.HostingID
	if p := jsTrim(jsx.StrOrEmpty(c.App.Settings.Site(own).V("parentHostingId"))); p != "" && p != own {
		return p
	}
	return own
}

var metaOwnerRule = regexp.MustCompile(`^meta_(S\d+)$`)

// metaOwnerHostingID — 사이트 메타 첨부의 주인 회사(즈비드는 자기 회사 서류일 때만)
func metaOwnerHostingID(c *web.Ctx, imageID any) string {
	m := metaOwnerRule.FindStringSubmatch(jsx.StrOrEmpty(imageID))
	if m == nil || c.Token.MemberType != "A" {
		return ""
	}
	if c.Token.HostingID == m[1] {
		return m[1]
	}
	return ""
}

// POST /user/file/secure-upload?hostingId= (헤더 imageid) — 보안 파일 올리기
func secureUpload(c *web.Ctx) {
	imageID := c.Header("imageid")
	cid := jsx.StrOrEmpty(c.QV("hostingId"))
	if cid == "" && c.Token.MemberType == "A" {
		cid = parentHostingID(c)
	}
	if !jsx.Truthy(c.QV("hostingId")) {
		if o := metaOwnerHostingID(c, imageID); o != "" {
			cid = o
		}
	}
	sd := secureDir(c.App)
	if cid != "" && !exists(sd+cid+"/") {
		mkdirAll(sd + cid + "/")
	}
	dir := sd
	if cid != "" {
		dir += cid + "/"
	}
	if imageID != "" {
		dir += imageID + "/"
	}
	if !exists(dir) {
		mkdirAll(dir)
	}
	var real any
	files, err := saveMultipart(c, dir, func(o string) string { return o }, 1000*1024*1024)
	if err != nil {
		c.Ecb(err.Error())
		return
	}
	for _, f := range files {
		real = f.originalFilename
	}
	var iid any
	if imageID != "" {
		iid = imageID
	}
	c.Scb(objU("secureImageId", undefIfNil(iid), "fileName", undefIfNil(real), "hostingId", cid))
}

func undefIfNil(v any) any {
	if v == nil {
		return undef
	}
	return v
}

// 업무관리 카드 첨부 폴더 work_{번호}
var workFileRule = regexp.MustCompile(`(?:^|/)work_(\d+)$`)
var postRule = regexp.MustCompile(`^([^/]+)/(?:u_\d+_)?post_([A-Za-z0-9]+)$`)

// GET·DELETE /user/file/secure-download?imageId=&fileName= — 보안 파일 받기·지우기
func secureDownload(c *web.Ctx) {
	var bd *jsx.OrderedMap
	if m, ok := c.Body().(*jsx.OrderedMap); ok && m != nil {
		bd = m
	} else {
		bd = c.Query
	}
	imgID := jsx.StrOrEmpty(bd.V("imageId"))
	if tokenKey(c) == "" || imgID == "" || !jsx.Truthy(bd.V("fileName")) {
		c.Ecb("잘못된 요청")
		return
	}
	ctx := c.Context()
	cid := ""
	if c.Token.MemberType == "A" {
		cid = parentHostingID(c)
	}
	if m := workFileRule.FindStringSubmatch(imgID); m != nil && c.Token.MemberType == "A" {
		r, err := db.QueryOne(ctx, c.App.DB, `SELECT workId, hostingId, groupKey FROM work_items WHERE workId = ?`, m[1])
		if err == nil && r != nil && jsx.StrOrEmpty(r.V("hostingId")) == c.Token.HostingID {
			cid = jsx.StrOrEmpty(r.V("hostingId"))
			imgID = "work_" + m[1]
		}
	}
	if m := postRule.FindStringSubmatch(imgID); m != nil && cid != "" {
		r, err := db.QueryOne(ctx, c.App.DB, `SELECT hostingId, json FROM posts WHERE postId = ?`, m[2])
		if err == nil && r != nil {
			pj := parseJSONObj(r.V("json"))
			if jsx.StrOrEmpty(r.V("hostingId")) == cid || jsx.StrOrEmpty(pj.V("hostingId")) == cid {
				cid = ""
			}
		}
	}
	if o := metaOwnerHostingID(c, imgID); o != "" {
		cid = o
	}
	dir := secureDir(c.App)
	if cid != "" {
		dir += cid + "/"
	}
	dir += imgID + "/"
	fp := nodeJoin(dir + jsx.Str(bd.V("fileName")))
	if !strings.HasPrefix(fp, dir) {
		c.Ecb("잘못된 요청")
		return
	}
	if c.Method == "DELETE" {
		if err := os.Remove(fp); err != nil {
			c.Ecb(nodeErr("unlink", fp, err))
			return
		}
		c.Scb("DELETED")
		return
	}
	b, err := os.ReadFile(fp)
	if err != nil {
		c.Ecb(nodeErr("open", fp, err))
		return
	}
	c.SetHeader("Content-Type", "image/jpg")
	c.Status(200).Send(b)
}

var imageIDRule = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

// POST /user/file/secure-move {fromImageId, toImageId} — 내 회사 보안 폴더 안에서 옮기기(이미 있는 파일은 건너뜀)
func secureMove(c *web.Ctx) {
	if c.Method != "POST" {
		c.Status(405).Send(jsx.Obj("error", "Method "+c.Method+" not allowed"))
		return
	}
	bd := c.BodyMap()
	from, okF := bd.V("fromImageId").(string)
	to, okT := bd.V("toImageId").(string)
	if !okF || !okT || !imageIDRule.MatchString(from) || !imageIDRule.MatchString(to) {
		c.Ecb("imageId 형식이 올바르지 않습니다")
		return
	}
	base := secureDir(c.App)
	if cid := c.Token.HostingID; cid != "" {
		base += cid + "/"
	}
	fromDir, toDir := path.Join(base, from), path.Join(base, to)
	moved, skipped := 0.0, 0.0
	if exists(fromDir) {
		if !exists(toDir) {
			mkdirAll(toDir)
		}
		ents, err := os.ReadDir(fromDir)
		if err != nil {
			c.Ecb(nodeErr("scandir", fromDir, err))
			return
		}
		for _, e := range ents {
			src, dst := path.Join(fromDir, e.Name()), path.Join(toDir, e.Name())
			if exists(dst) {
				skipped++
				continue
			}
			if err := os.Rename(src, dst); err != nil {
				c.Ecb(nodeErr("rename", src, err))
				return
			}
			moved++
		}
		if rest, err := os.ReadDir(fromDir); err == nil && len(rest) == 0 {
			_ = os.Remove(fromDir)
		}
	}
	c.Scb(jsx.Obj("moved", moved, "skipped", skipped))
}
