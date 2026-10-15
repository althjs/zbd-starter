-- seed 파일 하나 앞 — 넣기 전 모습(build.sh 가 앞에 SET @seed_file := '이름' 을 붙인다)
SET @seed_new := NOT EXISTS (SELECT 1 FROM seed_files WHERE name = @seed_file);
DROP TEMPORARY TABLE IF EXISTS seed_had_menus, seed_had_roles;
CREATE TEMPORARY TABLE seed_had_menus AS SELECT menuId, hostingId FROM admin_menus;
CREATE TEMPORARY TABLE seed_had_roles AS SELECT groupId, menuId FROM admin_role_menus;
