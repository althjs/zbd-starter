package admincontent

// 여러 부분 본문(파일 올리기) 읽기 — 지금 서비스가 쓰던 업로드 라이브러리와 같은 결과 모양으로.

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

type upFile struct {
	newName  string
	origName string
	size     int64
	mimetype any // 없으면 null
	mtime    time.Time
	path     string
}

// toJSON — 업로드 라이브러리의 파일 객체 JSON(filepath 를 지운 모양 — 화면에 경로를 내보내지 않는다)
func (f upFile) toJSON(withPath bool) *jsx.OrderedMap {
	m := jsx.NewOrderedMap()
	m.Set("size", float64(f.size))
	if withPath {
		m.Set("filepath", f.path)
	}
	m.Set("newFilename", f.newName)
	m.Set("mimetype", f.mimetype)
	m.Set("mtime", f.mtime.UTC().Format("2006-01-02T15:04:05.000Z"))
	m.Set("originalFilename", f.origName)
	return m
}

var reCtrl = regexp.MustCompile(`[\x00-\x1f]`)

// safeUploadName — 원래 파일명에서 경로 부분·제어문자를 뗀다(비면 "file")
func safeUploadName(name string) string {
	parts := regexp.MustCompile(`[\\/]`).Split(name, -1)
	base := parts[len(parts)-1]
	clean := strings.TrimSpace(reCtrl.ReplaceAllString(base, ""))
	if clean == "" || clean == "." || clean == ".." {
		return "file"
	}
	return clean
}

var errUploadTooLarge = errors.New("options.maxFileSize exceeded")

// parseMultipart — 파일은 dir 에 nameFn(원래 이름) 으로 저장, 필드는 값 목록으로.
func parseMultipart(c *web.Ctx, dir string, maxFile, maxTotal int64, nameFn func(string) string) (map[string][]string, map[string][]upFile, error) {
	if maxTotal == 0 {
		maxTotal = maxFile
	}
	_, params, err := mime.ParseMediaType(c.Header("Content-Type"))
	if err != nil || params["boundary"] == "" {
		return nil, nil, errors.New("bad content-type header, no multipart boundary")
	}
	rd, err := c.RawBodyReader()
	if err != nil {
		return nil, nil, err
	}
	mr := multipart.NewReader(rd, params["boundary"])
	fields := map[string][]string{}
	files := map[string][]upFile{}
	var total int64
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		name := p.FormName()
		if p.FileName() == "" && !strings.Contains(p.Header.Get("Content-Disposition"), "filename=") {
			b, err := io.ReadAll(io.LimitReader(p, 20<<20))
			if err != nil {
				return nil, nil, err
			}
			fields[name] = append(fields[name], string(b))
			continue
		}
		orig := p.FileName()
		if _, cd, err := mime.ParseMediaType(p.Header.Get("Content-Disposition")); err == nil {
			orig = cd["filename"] // 경로까지 원래 그대로(라이브러리가 자르지 않는다)
		}
		nn := nameFn(orig)
		path := filepath.Join(dir, nn)
		out, err := os.Create(path)
		if err != nil {
			return nil, nil, err
		}
		n, err := io.Copy(out, io.LimitReader(p, maxFile+1))
		out.Close()
		if err != nil {
			return nil, nil, err
		}
		total += n
		if n > maxFile || total > maxTotal {
			_ = os.Remove(path)
			return nil, nil, errUploadTooLarge
		}
		var mt any
		if v := p.Header.Get("Content-Type"); v != "" {
			mt = v
		}
		files[name] = append(files[name], upFile{newName: nn, origName: orig, size: n, mimetype: mt, mtime: time.Now(), path: path})
	}
	return fields, files, nil
}
