-- 0001 되돌리기 — 표를 모두 지운다(데이터도 사라진다)
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `admin_menu_apis`;
DROP TABLE IF EXISTS `admin_menus`;
DROP TABLE IF EXISTS `admin_role_menus`;
DROP TABLE IF EXISTS `admin_roles`;
DROP TABLE IF EXISTS `admin_token_logs`;
DROP TABLE IF EXISTS `admins`;
DROP TABLE IF EXISTS `passkeys`;
DROP TABLE IF EXISTS `work_items`;
DROP TABLE IF EXISTS `login_devices`;
DROP TABLE IF EXISTS `members`;
DROP TABLE IF EXISTS `post_views`;
DROP TABLE IF EXISTS `posts`;
DROP TABLE IF EXISTS `settings`;
SET FOREIGN_KEY_CHECKS = 1;
