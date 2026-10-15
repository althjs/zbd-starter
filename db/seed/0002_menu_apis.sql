-- 메뉴 ↔ API 묶음 — 묶인 API 는 그 메뉴 권한이 있어야 부를 수 있다(GET = R 이상, 그 밖 = RW). 묶이지 않은 API 는 누구나 통과.
-- 이 묶음이 없던 때는 웹하드 메뉴만 받은 관리자도 최고 관리자 계정을 만들 수 있었다(2026-10-06 E2E K5).
-- 「W:」 = 쓰기(GET 아닌 요청)만 — 관리자 목록은 업무 보드 담당자·게시판 글쓴이 이름으로 여러 화면이 읽는다.
-- 앞부분이 맞는 것 중 가장 긴 묶음 하나가 정한다(adminauth.AllowedBy). 배포 때마다 돌아도 빠진 줄만 넣는다.
INSERT INTO admin_menu_apis (menuId, api, description, hostingId)
SELECT x.menuId, x.api, x.description, 'ZBD' FROM (
	SELECT 31 AS menuId, '/service/admin/user' AS api, '사용자 회원 목록·상세·수정' AS description
	UNION ALL SELECT 32, 'W:/service/admin/member', '관리자 만들기·고치기·지우기(목록 읽기는 모두)'
	UNION ALL SELECT 33, '/service/admin/group', '권한 그룹·그룹 메뉴 권한'
	UNION ALL SELECT 34, '/service/admin/member/login-devices', '로그인 기기'
	UNION ALL SELECT 41, '/service/admin/meta', '사이트 설정(비밀값이 들 수 있다)'
	UNION ALL SELECT 43, '/service/admin/meta', '공통 코드'
	UNION ALL SELECT 41, '/service/admin/domain', '도메인'
	UNION ALL SELECT 42, 'W:/service/admin/menus', '메뉴·메뉴 API 고치기(읽기는 권한 화면도 쓴다)'
	UNION ALL SELECT 10, '/service/admin/work', '업무관리'
	UNION ALL SELECT 11, '/service/admin/file-explorer', '웹하드'
) x
JOIN admin_menus m ON m.menuId = x.menuId AND m.hostingId = 'ZBD'
WHERE NOT EXISTS (SELECT 1 FROM admin_menu_apis a WHERE a.menuId = x.menuId AND a.api = x.api AND a.hostingId = 'ZBD');
