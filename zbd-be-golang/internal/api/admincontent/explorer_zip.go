package admincontent

// 웹하드 폴더 내려받기 — 폴더를 zip 으로 묶어 바로 흘려보낸다(서버에 임시 파일을 만들지 않는다).
// 읽기 권한이 없는 하위 폴더·파일은 빼고 묶는다(숨김이든 아니든 — explorer_acl.go). 링크(symlink)는 따라가지 않는다.

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"zbd.kr/be/internal/web"
)

func adminFileExplorerZip(c *web.Ctx) {
	if c.Method != "GET" {
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
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		c.Ecb("폴더를 찾을 수 없습니다.")
		return
	}
	canRead, ok := explorerReadFilter(c, cid, rootFolder, root, dir) // 폴더 권한(explorer_acl.go — 스타터 기본은 늘 읽기)
	if !ok {
		return
	}
	name := filepath.Base(dir)
	if dir == root {
		name = orStr(rootFolder)
	}
	c.W.Header().Set("Content-Type", "application/zip")
	c.W.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+encodeURIComponent(name+".zip"))
	c.W.Header().Set("X-Accel-Buffering", "no")
	c.W.WriteHeader(200)
	zw := zip.NewWriter(c.W)
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 못 읽는 것은 건너뛴다
		}
		if p == dir {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 || d.Name() == ".DS_Store" {
			return nil
		}
		if !canRead(p) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		zname := name + "/" + filepath.ToSlash(rel)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if d.IsDir() {
			_, err := zw.CreateHeader(&zip.FileHeader{Name: zname + "/", Modified: info.ModTime()})
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		h, _ := zip.FileInfoHeader(info)
		h.Name, h.Method = zname, zipMethod(p)
		w, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		log.Printf("[file-explorer/zip] %s: %v", name, err) // 받는 쪽이 끊으면 여기 — 머리말은 이미 나가 오류 응답은 못 준다
	}
	_ = zw.Close()
}

// zipMethod — 이미 압축된 형식은 그대로 담아(Store) CPU 를 아낀다
func zipMethod(p string) uint16 {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(p), ".")) {
	case "zip", "gz", "tgz", "7z", "rar", "jpg", "jpeg", "png", "gif", "webp", "mp4", "mov", "mp3", "m4a", "pdf", "docx", "xlsx", "pptx", "hwpx", "zps", "heic", "avif":
		return zip.Store
	}
	return zip.Deflate
}
