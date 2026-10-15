-- 0001 기본 스키마 — 표 13개(스타터 포함분만 — 2026-10-04 즈비드 운영 구조에서 데이터 없이 떠 옴. 플러그인·즈비드 전용·페이지 디자인 표는 뺐다).
-- 시간은 UTC 로 저장한다(서버·DB 모두). 개인정보 칸(members.name·email·mobile)은 BE 가 암호화해 넣는다.
-- 되돌리기: 0001_schema.down.sql
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

/*M!999999\- enable the sandbox mode */ 

CREATE TABLE IF NOT EXISTS `admin_menu_apis` (
  `apiId` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'API ID',
  `menuId` int(10) NOT NULL DEFAULT 0 COMMENT '메뉴 ID',
  `api` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'ex) /service/admin/menus',
  `description` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'API 설명',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`apiId`,`menuId`,`api`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `admin_menus` (
  `menuId` int(10) NOT NULL DEFAULT 0 COMMENT '메뉴 ID',
  `parentId` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '2depth 이상인 경우 부모 menuId',
  `sortNo` int(10) NOT NULL DEFAULT 0 COMMENT '정렬순서 - 내림차순',
  `name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '메뉴 명',
  `icon` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '메뉴 아이콘',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '수정일',
  `path` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '링크 패스 ex) /home',
  `special` varchar(30) DEFAULT NULL,
  `description` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '메뉴 설명',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`menuId`,`hostingId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `admin_role_menus` (
  `groupId` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '관리자 그룹 ID',
  `menuId` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '메뉴 ID',
  `auth` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'R' COMMENT '권한 R: 읽기전용, RW: 읽기/쓰기',
  `status` int(10) NOT NULL DEFAULT 0 COMMENT '상태 0: 정상, 99: 비활성',
  PRIMARY KEY (`groupId`,`menuId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `admin_roles` (
  `groupId` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '관리자 그룹 ID',
  `name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '관리자 그룹 명',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '수정일',
  `status` int(10) NOT NULL DEFAULT 0 COMMENT '상태 0: 정상, 99: 비활성',
  `json` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '부가정보 JSON 저장',
  `description` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '그룹 설명',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`groupId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `admin_token_logs` (
  `logId` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `kind` varchar(20) NOT NULL COMMENT 'DEV_LOGIN | TOKEN_PASTE',
  `result` varchar(10) NOT NULL COMMENT 'OK | DENIED',
  `reason` varchar(200) NOT NULL DEFAULT '',
  `issuerAdminId` int(10) unsigned NOT NULL DEFAULT 0,
  `issuerName` varchar(100) NOT NULL DEFAULT '',
  `issuerHostingId` varchar(50) NOT NULL DEFAULT '',
  `targetAdminId` int(10) unsigned NOT NULL DEFAULT 0,
  `targetName` varchar(100) NOT NULL DEFAULT '',
  `targetHostingId` varchar(50) NOT NULL DEFAULT '',
  `ip` varchar(50) NOT NULL DEFAULT '',
  `ipCandidates` varchar(300) NOT NULL DEFAULT '',
  `isOffice` tinyint(4) NOT NULL DEFAULT 0,
  `host` varchar(200) NOT NULL DEFAULT '',
  `ua` varchar(500) NOT NULL DEFAULT '',
  `createdAt` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`logId`),
  KEY `ix_admin_token_logs_created` (`createdAt`),
  KEY `ix_admin_token_logs_target` (`targetAdminId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='DXAUTH 토큰 발급 기록(성공·거절)';
CREATE TABLE IF NOT EXISTS `admins` (
  `adminId` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '관리자 ID',
  `groupId` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '관리자 그룹 ID',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '관리자 명',
  `email` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '관리자 이메일',
  `password` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '패스워드 SHA512 hash',
  `mobile` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '휴대폰 번호',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '수정일',
  `status` int(10) NOT NULL DEFAULT 0 COMMENT '상태 0: 정상, 99: 비활성',
  `json` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '부가정보 JSON 저장',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `otp` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '2FA 인증 정보',
  PRIMARY KEY (`adminId`,`email`),
  UNIQUE KEY `uq_admins_email_hosting` (`email`,`hostingId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `passkeys` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `hostingId` varchar(50) NOT NULL DEFAULT '',
  `memberType` char(1) NOT NULL COMMENT 'A 관리자 / U 회원',
  `ownerId` int(10) unsigned NOT NULL,
  `credId` varchar(255) NOT NULL COMMENT 'base64url',
  `publicKey` text NOT NULL COMMENT 'SPKI DER base64url',
  `alg` int(11) NOT NULL DEFAULT 0,
  `signCount` int(10) unsigned NOT NULL DEFAULT 0,
  `rpId` varchar(100) NOT NULL,
  `name` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `transports` varchar(100) NOT NULL DEFAULT '',
  `createdAt` datetime NOT NULL DEFAULT current_timestamp(),
  `lastUsedAt` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_passkeys_cred` (`credId`),
  KEY `ix_passkeys_owner` (`memberType`,`ownerId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='패스키(지문·얼굴) 로그인 열쇠 — 관리자(A)·회원(U)';
CREATE TABLE IF NOT EXISTS `work_items` (
  `workId` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '업무 ID',
  `adminId` int(11) NOT NULL DEFAULT 0 COMMENT '작성자 adminId',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '생성 timestamp',
  `Json` text NOT NULL DEFAULT '{}' COMMENT '콘텐츠 내용',
  `parentWorkId` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '부모 업무 ID',
  `hostingId` varchar(50) NOT NULL DEFAULT '' COMMENT 'hostingId',
  `groupKey` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '분류 키',
  `updatedAt` timestamp NULL DEFAULT NULL ON UPDATE current_timestamp() COMMENT '업데이트 시간',
  PRIMARY KEY (`workId`),
  KEY `ix_work_items_hosting_group` (`hostingId`,`groupKey`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='업무관리(work) 카드';
CREATE TABLE IF NOT EXISTS `login_devices` (
  `memberType` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'A: 어드민, U: 일반유저',
  `email` varchar(200) NOT NULL DEFAULT '' COMMENT '이메일 찾기 값(HMAC) — 화면 이메일은 회원 표에서',
  `browserUid` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '브라우저 고유ID: oxuid 헤더값, ox_uid 쿠키값',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `lastLoginedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '마지막 로그인 일시 (refresh token 으로 갱신한 경우도 업데이트)',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `ip` varchar(50) NOT NULL DEFAULT '' COMMENT '로그인 IP(v4/v6 지원)',
  `userId` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '소셜추가되면서 email uniq 가 아님..',
  PRIMARY KEY (`memberType`,`email`,`browserUid`,`hostingId`,`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `members` (
  `userId` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '사용자 ID',
  `groupId` int(11) DEFAULT 0 COMMENT '사용자 그룹 ID',
  `name` varchar(400) NOT NULL DEFAULT '' COMMENT '이름 — BE 암호문(pii-crypto v1)',
  `name_hash` char(64) NOT NULL DEFAULT '' COMMENT '이름 찾기 값(HMAC)',
  `email` varchar(400) NOT NULL DEFAULT '' COMMENT '이메일 — BE 암호문(pii-crypto v1)',
  `email_hash` char(64) NOT NULL DEFAULT '' COMMENT '이메일 찾기 값(HMAC, 소문자)',
  `password` varchar(200) NOT NULL COMMENT '패스워드 SHA512 hash',
  `mobile` varchar(200) NOT NULL DEFAULT '' COMMENT '휴대폰 — BE 암호문(pii-crypto v1)',
  `mobile_hash` char(64) NOT NULL DEFAULT '' COMMENT '휴대폰 찾기 값(HMAC, 숫자만)',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '수정일',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '상태 0: 정상, 99: 비활성',
  `json` text DEFAULT '' COMMENT '부가정보 JSON 저장',
  `nickname` varchar(100) DEFAULT '' COMMENT '사용자 닉네임',
  `image` varchar(30) NOT NULL DEFAULT '' COMMENT '프로필 이미지 파일명 imageId 는 u_[userId]_my',
  `description` varchar(200) NOT NULL DEFAULT '' COMMENT '프로필 소개글',
  `socialId` varchar(100) NOT NULL DEFAULT '' COMMENT '소셜 로그인 고유 ID (카카오/네이버/구글 등)',
  `hostingId` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`userId`,`email`),
  UNIQUE KEY `uq_members_email_social` (`email_hash`,`socialId`,`hostingId`),
  KEY `ix_members_group` (`groupId`),
  KEY `ix_members_hosting` (`hostingId`),
  KEY `ix_members_name_hash` (`name_hash`),
  KEY `ix_members_email_hash` (`email_hash`),
  KEY `ix_members_mobile_hash` (`mobile_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
CREATE TABLE IF NOT EXISTS `post_views` (
  `postId` varchar(15) NOT NULL DEFAULT '' COMMENT '포스트 ID',
  `pv` int(10) unsigned NOT NULL DEFAULT 0,
  `groupKey` varchar(20) NOT NULL DEFAULT '' COMMENT '시간 interval key ex) 2025-06-19 14:00',
  `hostingId` varchar(50) NOT NULL DEFAULT '',
  UNIQUE KEY `uq_post_views_post_group` (`postId`,`groupKey`),
  KEY `ix_post_views_hosting_post` (`hostingId`,`postId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
CREATE TABLE IF NOT EXISTS `posts` (
  `postId` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '포스트 ID',
  `adminId` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '작성자 - 관리자 ID',
  `userId` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '작성자 - 사용자 ID',
  `groupKey` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '분류 키',
  `groupSubKey` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '분류 서브 키',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '포스트 타이틀',
  `json` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '포스트 본문 등 JSON 저장',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '수정일',
  `status` int(10) NOT NULL DEFAULT 0 COMMENT '상태 0: 정상, 99: 비활성',
  `jsonUser` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '사용자용 json',
  `hostingId` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`postId`),
  KEY `ix_posts_lookup` (`postId`,`hostingId`,`groupKey`,`userId`,`groupSubKey`,`status`,`createdAt`,`updatedAt`) USING BTREE,
  KEY `ix_posts_group_status` (`groupKey`,`status`),
  FULLTEXT KEY `ix_posts_subgroup` (`groupSubKey`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `settings` (
  `metaKey` varchar(30) NOT NULL DEFAULT '' COMMENT '메타정보 키',
  `json` mediumtext NOT NULL DEFAULT '' COMMENT '메타정보 json',
  `memo` varchar(700) NOT NULL DEFAULT '' COMMENT '메타정보 설명',
  `updatedAt` timestamp NULL DEFAULT NULL COMMENT '수정일',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '메타 키 이름 (관리자에 노출됨)',
  `hostingId` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`metaKey`,`hostingId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET FOREIGN_KEY_CHECKS = 1;
