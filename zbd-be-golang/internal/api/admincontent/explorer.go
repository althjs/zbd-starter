package admincontent

// 관리자 웹하드(파일 탐색기) — /admin/file-explorer/[rootFolder] · move · presence · upload
// 원본: pages/api/admin/file-explorer/*, server-libs/file-explorer.ts
//
// 경로 = {업로드 폴더}/../file-explorer/{회사}/{rootFolder}. rootFolder 는 영문 대문자·숫자·_ 만, 모든 경로는 루트 안인지 확인한다.

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

var (
	reRootFolder = regexp.MustCompile(`^[A-Z0-9_]{1,40}$`)
	reHostingID  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)
)

// explorerRoot — 웹하드 루트(잘못된 이름이면 "")
func explorerRoot(c *web.Ctx, hostingID string, rootFolder any) string {
	rf := orStr(rootFolder)
	if !reRootFolder.MatchString(rf) {
		return ""
	}
	if hostingID != "" && !reHostingID.MatchString(hostingID) {
		return ""
	}
	base, _ := filepath.Abs(uploadDir(c))
	return filepath.Join(base, "../file-explorer/", hostingID, rf)
}

// explorerPath — 루트 안의 경로(루트 자신 포함). 밖이면 ""
func explorerPath(root string, rel any) string {
	r := ""
	if jsx.Truthy(rel) {
		r = jsx.Str(rel)
	}
	p := filepath.Join(root, "./"+strings.ReplaceAll(r, "\\", "/"))
	if p != root && !strings.HasPrefix(p, root+string(filepath.Separator)) {
		return ""
	}
	return p
}

// 관리자 이름(신호의 「누가」) — 10분 캐시
var (
	adminNameMu    sync.Mutex
	adminNameCache = map[string]struct {
		name string
		at   time.Time
	}{}
)

func adminName(ctx context.Context, c *web.Ctx, hostingID string, adminID any) string {
	id := jsNum(adminID)
	if id == 0 {
		return ""
	}
	k := hostingID + ":" + jsx.NumStr(id)
	adminNameMu.Lock()
	hit, ok := adminNameCache[k]
	adminNameMu.Unlock()
	if ok && time.Since(hit.at) < 10*time.Minute {
		return hit.name
	}
	r, err := db.QueryOne(ctx, c.App.DB, `SELECT name FROM admins WHERE adminId = ? AND hostingId = ?`, jsx.Param(id), hostingID)
	if err != nil || r == nil {
		return ""
	}
	name := orStr(r.V("name"))
	if name != "" {
		adminNameMu.Lock()
		adminNameCache[k] = struct {
			name string
			at   time.Time
		}{name, time.Now()}
		adminNameMu.Unlock()
	}
	return name
}

// notifyExplorer — 웹하드 변경 신호(열어 둔 화면이 다시 읽는다)
func notifyExplorer(ctx context.Context, c *web.Ctx, hostingID string, rootFolder any, root string, abs []string, action string, by any, extra *jsx.OrderedMap) {
	paths := []any{}
	for _, p := range abs {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}
		if rel == "." {
			rel = ""
		}
		paths = append(paths, filepath.ToSlash(rel))
	}
	var first any
	if len(paths) > 0 {
		first = paths[0]
	}
	ev := jsx.Obj("hostingId", hostingID, "zbduid", "", "adminId", float64(0), "groupKey", "WEB_HDD", "key", rootFolder,
		"path", first, "paths", paths, "action", action, "by", jsNum(by), "byName", adminName(ctx, c, hostingID, by))
	jsx.Assign(ev, extra)
	c.App.Emit("doSocketDataUpdate", ev)
}

// textHash — 문서 지문(FNV-1a 32bit, UTF-16 단위) — 화면과 같은 계산
func textHash(s string) string {
	u := utf16.Encode([]rune(s))
	h := uint32(0x811c9dc5)
	for _, ch := range u {
		h ^= uint32(ch)
		h *= 0x01000193
	}
	return fmt.Sprintf("%x:%d", h, len(u))
}

var textExts = []string{"md", "txt", "json", "js", "jsx", "ts", "tsx", "css", "html", "xml", "yml", "yaml", "csv"}

// dirTree — 폴더 나무(숨김 파일 제외)
func dirTree(dir, root string) []any {
	out := []any{}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		rel, _ := filepath.Rel(root, p)
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		mt := st.ModTime().UTC().Format("2006-01-02T15:04:05.000Z")
		if st.IsDir() {
			out = append(out, jsx.Obj("name", e.Name(), "path", filepath.ToSlash(rel), "type", "directory", "size", float64(st.Size()), "updatedAt", mt, "items", dirTree(p, root)))
		} else {
			out = append(out, jsx.Obj("name", e.Name(), "path", filepath.ToSlash(rel), "type", "file", "size", float64(st.Size()), "updatedAt", mt, "extension", filepath.Ext(e.Name())))
		}
	}
	return out
}

func adminFileExplorer(c *web.Ctx) {
	cid := hostingID(c)
	ctx := c.Context()
	filePath := ""
	if v := c.QV("filePath"); jsx.Truthy(v) {
		filePath = jsx.Str(v)
	}
	rootFolder := c.QV("rootFolder")
	root := explorerRoot(c, cid, rootFolder)
	target := ""
	if root != "" {
		target = explorerPath(root, filePath)
	}
	if root == "" || target == "" {
		c.Ecb("잘못된 경로입니다.")
		return
	}
	if c.Method != "GET" && target == root {
		c.Ecb("최상위 폴더는 바꿀 수 없습니다.")
		return
	}
	_ = os.MkdirAll(root, 0o755)
	bd := bodyMap(c)
	switch c.Method {
	case "GET":
		if filePath != "" {
			st, err := os.Stat(target)
			if err != nil {
				c.Ecb("파일을 찾을 수 없습니다.")
				return
			}
			if !st.Mode().IsRegular() {
				c.Ecb("해당 경로는 파일이 아닙니다.")
				return
			}
			parts := strings.Split(filePath, ".")
			ext := strings.ToLower(parts[len(parts)-1])
			if contains(textExts, ext) {
				b, err := os.ReadFile(target)
				if err != nil {
					c.Ecb(jsErrText(err))
					return
				}
				c.Scb(string(b))
				return
			}
			streamFile(c, target, st.Size(), map[string]string{
				"Content-Type":        mimeByExt(ext),
				"Content-Disposition": "inline; filename*=UTF-8''" + encodeURIComponent(filepath.Base(target)),
			})
			return
		}
		st, err := os.Stat(root)
		if err != nil {
			c.Ecb("디렉토리를 찾을 수 없습니다.")
			return
		}
		if !st.IsDir() {
			c.Ecb("해당 경로는 디렉토리가 아닙니다.")
			return
		}
		c.Scb(dirTree(root, root))
	case "POST":
		isFolder := strictEq(bd.V("isFolder"), true) || nullishStr(bd.V("isFolder")) == "true"
		content := orStr(bd.V("content"))
		if isFolder {
			if err := os.MkdirAll(target, 0o755); err != nil {
				log.Printf("PUT ERROR: %v", err)
				c.Ecb("생성/저장 중 오류가 발생했습니다.")
				return
			}
			c.Scb("폴더가 생성되었습니다.")
		} else {
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err == nil {
				err = os.WriteFile(target, []byte(content), 0o644)
				if err != nil {
					log.Printf("PUT ERROR: %v", err)
					c.Ecb("생성/저장 중 오류가 발생했습니다.")
					return
				}
			} else {
				c.Ecb("생성/저장 중 오류가 발생했습니다.")
				return
			}
			c.Scb("파일이 저장되었습니다.")
		}
		notifyExplorer(ctx, c, cid, rootFolder, root, []string{target}, "create", c.Token.Key, nil)
	case "PUT":
		content := orStr(bd.V("text"))
		if _, err := os.Stat(target); err != nil {
			c.Ecb("수정 중 오류가 발생했습니다.")
			return
		}
		if base := bd.V("baseHash"); jsx.Truthy(base) {
			cur, err := os.ReadFile(target)
			if err != nil {
				c.Ecb("수정 중 오류가 발생했습니다.")
				return
			}
			if textHash(string(cur)) != jsx.Str(base) {
				c.Send(jsx.Obj("code", "4090", "status", "CONFLICT", "message", "다른 곳에서 먼저 수정했습니다.", "data", string(cur)))
				return
			}
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			c.Ecb("수정 중 오류가 발생했습니다.")
			return
		}
		c.Scb("파일이 수정되었습니다.")
		notifyExplorer(ctx, c, cid, rootFolder, root, []string{target}, "update", c.Token.Key, nil)
	case "DELETE":
		if _, err := os.Stat(target); err != nil {
			c.Ecb("대상을 찾을 수 없거나 삭제 중 오류가 발생했습니다.")
			return
		}
		if err := os.RemoveAll(target); err != nil {
			c.Ecb("대상을 찾을 수 없거나 삭제 중 오류가 발생했습니다.")
			return
		}
		c.Scb("삭제되었습니다.")
		notifyExplorer(ctx, c, cid, rootFolder, root, []string{target}, "delete", c.Token.Key, nil)
	default:
		c.Ecb("잘못된 요청")
	}
}

// 옮기기·이름 바꾸기(대상이 폴더면 그 안으로)
func adminFileExplorerMove(c *web.Ctx) {
	cid := hostingID(c)
	bd := bodyMap(c)
	rootFolder := c.QV("rootFolder")
	root := explorerRoot(c, cid, rootFolder)
	from, to := "", ""
	if root != "" {
		from = explorerPath(root, bd.V("fromPath"))
		to = explorerPath(root, bd.V("toPath"))
	}
	if root == "" || from == "" || to == "" || from == root {
		c.Ecb("잘못된 경로입니다.")
		return
	}
	if to == from || strings.HasPrefix(to, from+string(filepath.Separator)) {
		c.Ecb("자기 안으로는 옮길 수 없습니다.")
		return
	}
	if _, err := os.Stat(from); err != nil {
		c.Ecb("원본을 찾을 수 없습니다.")
		return
	}
	if st, err := os.Stat(to); err == nil && st.IsDir() {
		to = filepath.Join(to, filepath.Base(from))
	}
	if err := os.Rename(from, to); err != nil {
		c.Ecb("이동에 실패했습니다.")
		return
	}
	notifyExplorer(c.Context(), c, cid, rootFolder, root, []string{from, to}, "move", c.Token.Key, nil)
	rf, _ := filepath.Rel(root, from)
	rt, _ := filepath.Rel(root, to)
	c.Scb(jsx.Obj("fromPath", rf, "toPath", rt))
}

// 「보는 중 / 편집 중」 신호(저장하지 않는다)
func adminFileExplorerPresence(c *web.Ctx) {
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	cid := c.Token.HostingID
	bd := bodyMap(c)
	rootFolder := c.QV("rootFolder")
	root := explorerRoot(c, cid, rootFolder)
	p := ""
	if root != "" {
		p = explorerPath(root, bd.V("filePath"))
	}
	if root == "" || p == "" || p == root {
		c.Ecb("잘못된 경로입니다.")
		return
	}
	notifyExplorer(c.Context(), c, cid, rootFolder, root, []string{p}, "presence", c.Token.Key,
		jsx.Obj("editing", jsx.Truthy(bd.V("editing")), "on", !strictEq(bd.V("on"), false)))
	c.Scb("ok")
}

// 올리기 — files[] 를 웹하드 안 폴더(?folderPath=)에
func adminFileExplorerUpload(c *web.Ctx) {
	if c.Method != "POST" {
		c.Ecb("잘못된 요청")
		return
	}
	cid := hostingID(c)
	rootFolder := c.QV("rootFolder")
	root := explorerRoot(c, cid, rootFolder)
	dir := ""
	if root != "" {
		dir = explorerPath(root, c.QV("folderPath"))
	}
	if root == "" || dir == "" {
		c.Ecb("잘못된 경로입니다.")
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	_, files, err := parseMultipart(c, dir, 1<<30, 2<<30, safeUploadName)
	if err != nil {
		log.Printf("[file-explorer/upload] %v", err)
		c.Ecb("올리기에 실패했습니다.")
		return
	}
	list := files["files[]"]
	out := []any{}
	var paths []string
	for _, f := range list {
		out = append(out, jsx.Obj("name", filepath.Base(f.path), "size", float64(f.size)))
		paths = append(paths, f.path)
	}
	c.Scb(out)
	if len(paths) == 0 {
		paths = []string{dir}
	}
	notifyExplorer(c.Context(), c, cid, rootFolder, root, paths, "upload", c.Token.Key, nil)
}
