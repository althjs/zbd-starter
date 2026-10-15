package admincontent

import (
	"path/filepath"
	"testing"
)

// 정책: 웹하드 경로 — 루트 밖으로 못 나간다(../ · 역슬래시 · 섞인 경로), 루트 이름은 영문 대문자·숫자·_ 만
func TestExplorerPathPolicy(t *testing.T) {
	root := filepath.FromSlash("/srv/cdn/file-explorer/ZBD/WEB_HDD")
	for in, inside := range map[string]bool{
		"": true, "문서/계획.md": true, "a/./b": true, "/etc/passwd": true, // 앞 / 는 루트 안으로 붙는다
		"../x": false, "a/../../x": false, "..\\..\\x": false, "a/b/../../../ZBD2": false,
	} {
		p := explorerPath(root, in)
		if got := p != ""; got != inside {
			t.Errorf("%q → %q (안=%v, 기대 %v)", in, p, got, inside)
		}
	}
	for name, ok := range map[string]bool{"WEB_HDD": true, "A1": true, "web_hdd": false, "../X": false, "": false, "A B": false} {
		if reRootFolder.MatchString(name) != ok {
			t.Errorf("루트 이름 %q: 기대 %v", name, ok)
		}
	}
}
