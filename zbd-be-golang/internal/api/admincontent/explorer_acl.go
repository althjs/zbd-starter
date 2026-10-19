package admincontent

// 업무 웹하드 빈 판 — 스타터 기본 웹하드는 폴더 권한·개인 웹하드가 없다(모든 관리자가 같은 폴더를 읽고 쓴다).
// 「업무 웹하드」(zbd add work-webhard) 플러그인이 이 파일을 진짜 판(폴더·파일 권한 + 개인 웹하드 MY)으로 갈아 끼운다.
// 실시간 공동 편집(coedit.go)이 부르는 자리만 둔다.

import (
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// explorerCoeditRoom — 함께 편집 방 열쇠(루트 이름). 권한 검사 없음
func explorerCoeditRoom(c *web.Ctx, hostingID string, rootFolder any, root, p string) (string, bool) {
	return jsx.Str(rootFolder), true
}
