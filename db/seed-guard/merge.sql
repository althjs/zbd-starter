-- seed 파일 하나 뒤(in.sql 과 짝) — 받은 메뉴를 진짜 표로 옮긴다. 메뉴 번호가 이미 다른 메뉴(주소가 다름)에 쓰였으면
-- 조용히 건너뛰지 않고 1000번 이후 빈 번호로 옮겨 넣고, 그 파일의 하위 메뉴·권한·API 연결도 같이 따라간다(2026-10-19 대표 —
-- 고객이 만든 메뉴가 60~90번을 먼저 써서 그룹웨어·페이지 디자이너 등 플러그인 메뉴 11개가 소리 없이 빠졌다).
-- 옮긴 번호는 seed_menu_ids 에 남겨 다음 배포 때 같은 자리로(두 번 생기지 않게).
CREATE TEMPORARY TABLE seed_map (
	hostingId VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
	seedId INT NOT NULL,
	menuId INT NOT NULL,
	PRIMARY KEY (hostingId, seedId)
);
-- 같은 메뉴인가 = 주소가 같다(주소 없는 묶음 메뉴는 이름이 같다)
-- 1) 전에 옮겨 둔 번호 — 그 자리가 비었거나 같은 메뉴일 때만
INSERT IGNORE INTO seed_map (hostingId, seedId, menuId)
SELECT s.hostingId, s.menuId, k.menuId FROM seed_in_menus s
JOIN seed_menu_ids k ON k.file = @seed_file AND k.hostingId = s.hostingId AND k.seedId = s.menuId
LEFT JOIN admin_menus m ON m.menuId = k.menuId AND m.hostingId = k.hostingId
WHERE m.menuId IS NULL
   OR (COALESCE(s.path, '') <> '' AND m.path = s.path)
   OR (COALESCE(s.path, '') = '' AND COALESCE(m.path, '') = '' AND m.name = s.name);
-- 2) 같은 번호에 같은 메뉴가 이미 있다 — 그대로
INSERT IGNORE INTO seed_map (hostingId, seedId, menuId)
SELECT s.hostingId, s.menuId, s.menuId FROM seed_in_menus s
JOIN admin_menus m ON m.menuId = s.menuId AND m.hostingId = s.hostingId
WHERE (COALESCE(s.path, '') <> '' AND m.path = s.path)
   OR (COALESCE(s.path, '') = '' AND COALESCE(m.path, '') = '' AND m.name = s.name);
-- 3) 다른 번호에 같은 메뉴가 있다(다른 파일이 이미 옮겨 넣은 묶음 메뉴 등) — 그 번호로
INSERT IGNORE INTO seed_map (hostingId, seedId, menuId)
SELECT q.hostingId, q.seedId, q.menuId FROM (
	SELECT s.hostingId, s.menuId AS seedId,
		(SELECT MIN(m.menuId) FROM admin_menus m WHERE m.hostingId = s.hostingId
			AND ((COALESCE(s.path, '') <> '' AND m.path = s.path) OR (COALESCE(s.path, '') = '' AND COALESCE(m.path, '') = '' AND m.name = s.name))) AS menuId
	FROM seed_in_menus s
) q WHERE q.menuId IS NOT NULL;
-- 4) 번호가 비어 있다 — 그대로
INSERT IGNORE INTO seed_map (hostingId, seedId, menuId)
SELECT s.hostingId, s.menuId, s.menuId FROM seed_in_menus s
WHERE NOT EXISTS (SELECT 1 FROM admin_menus m WHERE m.menuId = s.menuId AND m.hostingId = s.hostingId);
-- 5) 번호를 다른 메뉴가 쓴다 — 1000번 이후 빈 번호로 옮긴다
SET @seed_next := GREATEST(999,
	(SELECT COALESCE(MAX(menuId), 0) FROM admin_menus),
	(SELECT COALESCE(MAX(menuId), 0) FROM seed_in_menus),
	(SELECT COALESCE(MAX(menuId), 0) FROM seed_menu_ids));
INSERT IGNORE INTO seed_map (hostingId, seedId, menuId)
SELECT q.hostingId, q.seedId, (@seed_next := @seed_next + 1) FROM (
	SELECT s.hostingId, s.menuId AS seedId FROM seed_in_menus s
	WHERE NOT EXISTS (SELECT 1 FROM seed_map p WHERE p.hostingId = s.hostingId AND p.seedId = s.menuId)
	ORDER BY s.menuId LIMIT 100000
) q;
REPLACE INTO seed_menu_ids (file, hostingId, seedId, menuId)
SELECT @seed_file, hostingId, seedId, menuId FROM seed_map WHERE seedId <> menuId;
SELECT CONCAT('· 메뉴 번호가 겹쳐 옮겨 넣음 (', @seed_file, '): ', GROUP_CONCAT(CONCAT(seedId, '→', menuId) ORDER BY seedId SEPARATOR ', ')) AS seed_guard
FROM seed_map WHERE seedId <> menuId HAVING COUNT(*) > 0;
-- 메뉴 — 칸 목록은 표에서 읽는다(칸이 늘어도 그대로). 상위 메뉴가 다른 파일에서 옮겨졌으면 그 번호로
SET @seed_cols := (SELECT GROUP_CONCAT(CONCAT('`', column_name, '`') ORDER BY ordinal_position)
	FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'admin_menus');
SET @seed_vals := (SELECT GROUP_CONCAT(CASE column_name
		WHEN 'menuId' THEN 'COALESCE(a.menuId, s.menuId)'
		WHEN 'parentId' THEN 'COALESCE(b.menuId, (SELECT k.menuId FROM seed_menu_ids k WHERE k.hostingId = s.hostingId AND k.seedId = s.parentId ORDER BY k.createdAt LIMIT 1), s.parentId)'
		ELSE CONCAT('s.`', column_name, '`') END ORDER BY ordinal_position)
	FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'admin_menus');
SET @seed_q := CONCAT('INSERT IGNORE INTO admin_menus (', @seed_cols, ') SELECT ', @seed_vals,
	' FROM seed_in_menus s LEFT JOIN seed_map a ON a.hostingId = s.hostingId AND a.seedId = s.menuId',
	' LEFT JOIN seed_map b ON b.hostingId = s.hostingId AND b.seedId = s.parentId');
PREPARE seed_stmt FROM @seed_q;
EXECUTE seed_stmt;
DEALLOCATE PREPARE seed_stmt;
-- 권한 — 권한 표엔 호스팅 칸이 없다(이 파일의 번호표로)
INSERT IGNORE INTO admin_role_menus (groupId, menuId, auth, status)
SELECT r.groupId, COALESCE((SELECT MIN(p.menuId) FROM seed_map p WHERE p.seedId = r.menuId), r.menuId), r.auth, r.status FROM seed_in_roles r;
-- 메뉴 API 연결
INSERT INTO admin_menu_apis (menuId, api, description, hostingId)
SELECT COALESCE(p.menuId, x.menuId), x.api, x.description, x.hostingId FROM seed_in_apis x
LEFT JOIN seed_map p ON p.hostingId = x.hostingId AND p.seedId = x.menuId
WHERE NOT EXISTS (SELECT 1 FROM admin_menu_apis z WHERE z.menuId = COALESCE(p.menuId, x.menuId) AND z.api = x.api);
DROP TEMPORARY TABLE IF EXISTS seed_in_menus, seed_in_roles, seed_in_apis, seed_map;
