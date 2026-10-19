-- 기본 데이터(db/seed) 앞 — 「한 번 넣은 메뉴·권한」 은 관리자가 지우면 다시 넣지 않는다.
-- 조립은 build.sh: pre → (파일마다 snap → 그 seed 파일 → post) → 끝에 빠진 파일 잊기. 한 세션에서 돈다.
-- seed 는 INSERT IGNORE 라 있는 행(순서·이름을 바꾼 메뉴)은 그대로지만, 지운 행은 배포마다 되살아났다.
-- 그래서 넣기 전 모습을 떠 두고(post.sql 이 비교), 넣은 적 있는 행은 seed_marks 에 남긴다. 한 세션에서 pre → seed → post 순서로 돈다.
SET @seed_files_first := (SELECT COUNT(*) = 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'seed_files');
SET @seed_first := (SELECT COUNT(*) = 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'seed_marks');
CREATE TABLE IF NOT EXISTS seed_marks (
	kind VARBINARY(20) NOT NULL COMMENT 'menu | role',
	k VARBINARY(100) NOT NULL COMMENT 'menu = 호스팅:메뉴 번호 · role = 그룹:메뉴 번호(바이트 비교 — 서버마다 기본 정렬 규칙이 달라도)',
	createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (kind, k)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '기본 데이터가 넣은 적 있는 메뉴·권한(지우면 다시 안 넣음)';
-- 이 장치가 처음 도는 사이트 = 지금 있는 메뉴·권한은 전부 예전 seed 가 넣은 것으로 본다(지금부터 지우면 안 돌아온다)
INSERT IGNORE INTO seed_marks (kind, k) SELECT 'menu', CONCAT(hostingId, ':', menuId) FROM admin_menus WHERE @seed_first;
INSERT IGNORE INTO seed_marks (kind, k) SELECT 'role', CONCAT(groupId, ':', menuId) FROM admin_role_menus WHERE @seed_first;
-- 어떤 seed 파일을 넣어 봤는지 — 처음 보는 파일(플러그인을 새로·다시 붙임)은 지운 적 있는 메뉴라도 되살린다(post.sql 의 @seed_new).
-- 플러그인을 빼서 파일이 없어지면 이름을 잊는다(build.sh 끝) → 다시 붙이면 메뉴가 돌아온다(2026-10-06 QA — 페이지 디자이너를 뺐다 다시 붙였는데 메뉴가 안 돌아왔다)
CREATE TABLE IF NOT EXISTS seed_files (
	name VARBINARY(150) NOT NULL PRIMARY KEY COMMENT 'db/seed 파일 이름(.sql 뺌)',
	createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '넣어 본 기본 데이터 파일(처음 보는 파일은 지운 메뉴 기억을 무시)';
-- 메뉴 번호가 겹쳐 옮겨 넣은 플러그인 메뉴(merge.sql) — 파일·원래 번호 → 이 사이트의 번호(2026-10-19)
CREATE TABLE IF NOT EXISTS seed_menu_ids (
	file VARBINARY(150) NOT NULL COMMENT 'db/seed 파일 이름(.sql 뺌)',
	hostingId VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
	seedId INT NOT NULL COMMENT '파일에 적힌 메뉴 번호',
	menuId INT NOT NULL COMMENT '이 사이트에 넣은 번호',
	createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (file, hostingId, seedId)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '번호가 겹쳐 옮겨 넣은 기본 데이터 메뉴';
