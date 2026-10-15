-- seed 파일 하나 뒤 — 이번에 새로 생긴 행 중 「넣은 적 있는데 관리자가 지운 것」 은 다시 지운다(처음 보는 파일이면 그대로 둔다)
DELETE m FROM admin_menus m
JOIN seed_marks s ON s.kind = 'menu' AND s.k = CONCAT(m.hostingId, ':', m.menuId)
LEFT JOIN seed_had_menus h ON h.menuId = m.menuId AND h.hostingId = m.hostingId
WHERE h.menuId IS NULL AND NOT @seed_new;
DELETE r FROM admin_role_menus r
JOIN seed_marks s ON s.kind = 'role' AND s.k = CONCAT(r.groupId, ':', r.menuId)
LEFT JOIN seed_had_roles h ON h.groupId = r.groupId AND h.menuId = r.menuId
WHERE h.groupId IS NULL AND NOT @seed_new;
-- 지운 메뉴에 딸려 새로 붙은 권한도 뺀다(메뉴가 없으면 권한도 의미 없다)
DELETE r FROM admin_role_menus r
LEFT JOIN seed_had_roles h ON h.groupId = r.groupId AND h.menuId = r.menuId
WHERE h.groupId IS NULL AND NOT @seed_new AND NOT EXISTS (SELECT 1 FROM admin_menus m WHERE m.menuId = r.menuId);
-- 이번에 처음 넣은 것을 기록
INSERT IGNORE INTO seed_marks (kind, k) SELECT 'menu', CONCAT(hostingId, ':', menuId) FROM admin_menus;
INSERT IGNORE INTO seed_marks (kind, k) SELECT 'role', CONCAT(groupId, ':', menuId) FROM admin_role_menus;
DROP TEMPORARY TABLE IF EXISTS seed_had_menus, seed_had_roles;
INSERT IGNORE INTO seed_files (name) VALUES (@seed_file);
