package admincontent

// 관리자 파일 — /admin/file/[folder] · download · secure-download · unlink-file · upload · upload-mp4
// 원본: pages/api/admin/file/*

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

const noFileHTML = `<script>alert("NO FILE FOUND");window.close()</script>`

// fileRealDir — 업로드 폴더 안 호스팅/preFolder/folder/subFolder (템플릿 폴더는 호스팅 구분 없음)
func fileRealDir(c *web.Ctx) string {
	cid := hostingID(c)
	folder := orStr(c.QV("folder"))
	pre := orStr(c.QV("preFolder"))
	sub := orStr(c.QV("subFolder"))
	if pre == "widget_templates" || pre == "page_templates" {
		cid = ""
	}
	d := uploadDir(c)
	if cid != "" {
		d += cid + "/"
	}
	if pre != "" {
		d += pre + "/"
	}
	if folder != "" {
		d += folder + "/"
	}
	if sub != "" {
		d += sub + "/"
	}
	return nodeJoin(d)
}

// 폴더 파일 목록 · 파일 하나 지우기
func adminFileFolder(c *web.Ctx) {
	dir := fileRealDir(c)
	if !fileExists(dir) {
		c.Ecb("존재하지 않는 경로입니다.")
		return
	}
	switch c.Method {
	case "GET":
		ents, err := os.ReadDir(dir)
		if err != nil {
			c.JSON(web.Ecb(jsErrText(err)))
			return
		}
		files := []any{}
		for _, e := range ents {
			n := e.Name()
			if strings.HasPrefix(n, ".nfs") || strings.Contains(n, "400.jpg") {
				continue
			}
			files = append(files, n)
		}
		c.JSON(web.Scb(jsx.Obj("files", files), nil))
	case "DELETE":
		name := orStr(c.QV("name"))
		if name == "" {
			c.Ecb("잘못된 요청")
			return
		}
		p := nodeJoin(dir + name)
		// 업로드 폴더 밖은 지우지 않는다(원본엔 이 검사가 없었다 — name 에 ../ 를 넣으면 밖의 파일이 지워졌다)
		base, _ := filepath.Abs(uploadDir(c))
		abs, _ := filepath.Abs(p)
		if !strings.HasPrefix(abs, base) {
			c.JSON(web.Ecb("Error: ENOENT: no such file or directory, unlink '" + p + "'"))
			return
		}
		if err := os.Remove(p); err != nil {
			c.JSON(web.Ecb("Error: ENOENT: no such file or directory, unlink '" + p + "'"))
			return
		}
		c.Scb("DELETED")
	default:
		c.Ecb("잘못된 요청")
	}
}

// 내려받기 폴더(업로드 폴더 옆 download/)의 파일
func adminFileDownload(c *web.Ctx) {
	up := uploadDir(c)
	real := nodeJoin(up + "../download/")
	root := nodeJoin(up, "../")
	if !fileExists(real) {
		c.Send(noFileHTML)
		return
	}
	name := orStr(c.QV("name"))
	if name == "" {
		c.Ecb("잘못된 요청")
		return
	}
	fp := nodeJoin(real + name)
	rr, _ := filepath.Abs(root)
	rf, _ := filepath.Abs(fp)
	if rf != rr && !strings.HasPrefix(rf, rr+string(filepath.Separator)) {
		c.Status(403).Send(`<script>alert("INVALID PATH");window.close()</script>`)
		return
	}
	st, err := os.Stat(fp)
	if err != nil {
		c.Send(noFileHTML)
		return
	}
	base := filepath.Base(name)
	ct := "application/octet-stream"
	if strings.ToLower(filepath.Ext(base)) == ".json" {
		ct = "application/json; charset=utf-8"
	}
	enc := encodeURIComponent(base)
	streamFile(c, fp, st.Size(), map[string]string{
		"Content-Type":        ct,
		"Content-Disposition": `attachment; filename="` + enc + `"; filename*=UTF-8''` + enc,
	})
}

// streamFile — 파일을 그대로 흘려 보낸다
func streamFile(c *web.Ctx, path string, size int64, headers map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		c.JSON(web.Ecb(jsErrText(err)))
		return
	}
	defer f.Close()
	for k, v := range headers {
		c.W.Header().Set(k, v)
	}
	c.W.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	c.W.WriteHeader(200)
	_, _ = io.Copy(c.W, f)
}

func mimeByExt(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	case "pdf":
		return "application/pdf"
	case "mp4":
		return "video/mp4"
	case "zip":
		return "application/zip"
	}
	return "application/octet-stream"
}

// 보안 폴더(secure/{호스팅}/{imageId}/) 파일 받기·지우기 — 로그인 필요
func adminFileSecureDownload(c *web.Ctx) {
	var bd *jsx.OrderedMap
	if b := body(c); jsx.Truthy(b) {
		bd = asMap(b)
	} else {
		bd = c.Query
	}
	cid := c.Token.HostingID
	imageID := orStr(bd.V("imageId"))
	if !c.Token.Valid() || bd == nil || imageID == "" || !jsx.Truthy(bd.V("fileName")) {
		c.Ecb("잘못된 요청")
		return
	}
	secure := nodeJoin(uploadDir(c) + "/../secure/")
	real := nodeJoin(secure + cid + "/" + jsx.Str(bd.V("imageId")) + "/")
	fp := nodeJoin(real + jsx.Str(bd.V("fileName")))
	if !strings.HasPrefix(fp, real) {
		c.Ecb("잘못된 요청")
		return
	}
	switch c.Method {
	case "DELETE":
		if err := os.Remove(fp); err != nil {
			c.Ecb("Error: ENOENT: no such file or directory, unlink '" + fp + "'")
			return
		}
		c.Scb("DELETED")
	default:
		st, err := os.Stat(fp)
		if err != nil {
			c.Ecb("파일을 찾을 수 없습니다.")
			return
		}
		parts := strings.Split(jsx.Str(bd.V("fileName")), ".")
		ext := strings.ToLower(parts[len(parts)-1])
		streamFile(c, fp, st.Size(), map[string]string{"Content-Type": mimeByExt(ext)})
	}
}

// 내려받기 폴더의 내 회사 파일 지우기
func adminFileUnlink(c *web.Ctx) {
	real := nodeJoin(uploadDir(c) + "../download/")
	if !fileExists(real) {
		c.Send(noFileHTML)
		return
	}
	name := orStr(c.QV("name"))
	if name == "" {
		c.Ecb("잘못된 요청")
		return
	}
	fp := nodeJoin(real + name)
	if !c.Token.Valid() || !strings.Contains(fp, c.Token.HostingID) {
		c.Ecb("권한이 없습니다.")
		return
	}
	if !fileExists(fp) {
		c.Send(noFileHTML)
		return
	}
	if err := os.Remove(fp); err != nil {
		c.JSON(web.Ecb(jsErrText(err)))
		return
	}
	c.Scb("DELETED")
}

// 올리기 — 저장 이름은 원래 파일명(경로 뗀 것). 응답은 파일 정보 목록(서버 경로 없음).
func adminFileUpload(c *web.Ctx) {
	dir := fileRealDir(c)
	_ = os.MkdirAll(dir, 0o755)
	_, files, err := parseMultipart(c, dir, 200<<20, 0, safeUploadName)
	if err != nil {
		c.JSON(web.Ecb(jsErrText(err)))
		return
	}
	out := []any{}
	for _, f := range files["files[]"] {
		out = append(out, f.toJSON(false))
	}
	c.JSON(web.Scb(out, nil))
}
