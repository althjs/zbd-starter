-- seed 파일 하나 앞(메뉴 번호를 맞춰 넣는 파일만) — 파일의 메뉴·권한·API 를 먼저 임시 표에 받는다(build.sh 가 표 이름을 바꿔 넣는다).
DROP TEMPORARY TABLE IF EXISTS seed_in_menus, seed_in_roles, seed_in_apis, seed_map;
CREATE TEMPORARY TABLE seed_in_menus LIKE admin_menus;
CREATE TEMPORARY TABLE seed_in_roles LIKE admin_role_menus;
CREATE TEMPORARY TABLE seed_in_apis LIKE admin_menu_apis;
