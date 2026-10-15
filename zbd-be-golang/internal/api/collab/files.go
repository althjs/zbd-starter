package collab

// 파일 공용 — 올린 파일 폴더(CDN_UPLOAD_DIR), 여러 부분 본문(multipart) 받기, 사진 미리보기(webp), 내려보내기.

import (
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"regexp"

	_ "golang.org/x/image/webp"
)

var errTooBig = errors.New("maxFileSize exceeded")

// formFile — 받은 파일 하나(임시 파일)
type formFile struct {
	path, name, mime string
	size             int64
}

var reNonASCII = regexp.MustCompile(`[^\x20-\x7e]`)
