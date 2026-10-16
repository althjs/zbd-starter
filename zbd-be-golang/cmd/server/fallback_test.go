package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiteAssetFallback(t *testing.T) {
	up := t.TempDir()
	os.MkdirAll(filepath.Join(up, "ZBD", "site"), 0o755)
	os.WriteFile(filepath.Join(up, "ZBD", "site", "192x192_favicon.png"), []byte("x"), 0o644)
	cases := map[string]bool{
		"/upload/ZBD/site/192x192_favicon.png":   false, // 업로드 폴더에 있음
		"/upload/ZBD/site/512x512_favicon.svg":   true,  // 저장소 원본에서
		"/upload/ZBD/site/없는파일.png":              false,
		"/upload/ZBD/site/..":                    false,
		"/upload/ZBD/other/512x512_favicon.svg":  false,
		"/upload/ZBD/site/a/512x512_favicon.svg": false,
	}
	for p, want := range cases {
		if got := siteAssetFallback(up, p) != ""; got != want {
			t.Errorf("%s: %v, want %v", p, got, want)
		}
	}
}
