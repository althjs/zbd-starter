#!/usr/bin/env bash
# DB 준비 — 데이터베이스·접속 계정 만들기 → 마이그레이션(번호 순, 한 번씩) → 기본 데이터.
# 여러 번 돌려도 같다. 접속 정보는 zbd-be-golang/.env.local(없으면 run_dev.sh 가 만든다).
#
#   ./db/setup.sh            준비(빠진 것만)
#   ./db/setup.sh --status   적용된 마이그레이션 보기
#
# 계정·DB 만들기는 MariaDB 관리자 권한이 필요하다 — macOS(brew) 는 지금 사용자, Linux 는 sudo 로 붙는다.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/zbd-be-golang/.env.local"
[ -f "$ENV_FILE" ] || { echo "✗ $ENV_FILE 이 없습니다 — ./run_dev.sh 를 먼저 실행하세요"; exit 1; }

envv() { grep -E "^$1=" "$ENV_FILE" | tail -1 | cut -d= -f2- | sed -e 's/^"//' -e 's/"$//'; }
DB_HOST="$(envv DB_HOST)"; DB_PORT="$(envv DB_PORT)"; DB_NAME="$(envv DB_NAME)"; DB_USER="$(envv DB_USER)"; DB_PASSWORD="$(envv DB_PASSWORD)"
DB_PORT="${DB_PORT:-3306}"

# 이 컴퓨터 안의 접속이라 TLS 를 쓰지 않는다(--skip-ssl — MariaDB 11.4+ 클라이언트의 인증서 경고를 끈다)
# 관리자로 붙기(계정·DB 만들 때만)
admin_sql() {
	if [ "$(uname)" = "Darwin" ]; then mariadb --skip-ssl -e "$1"; else sudo mariadb --skip-ssl -e "$1"; fi
}
# 앱 계정으로 붙기 — 비밀번호는 환경변수로만(명령줄에 쓰면 ps 에 보인다)
app_sql() { MYSQL_PWD="$DB_PASSWORD" mariadb --skip-ssl --default-character-set=utf8mb4 -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" "$DB_NAME" "$@"; }

if [ "${1:-}" = "--status" ]; then
	app_sql -e "SELECT version, appliedAt FROM schema_migrations ORDER BY version"
	exit 0
fi

# 1) DB·계정 — 이 DB 하나에만 권한(다른 DB·권한 부여 없음)
if ! MYSQL_PWD="$DB_PASSWORD" mariadb --skip-ssl -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -e "USE \`$DB_NAME\`" 2>/dev/null; then
	echo "· DB $DB_NAME · 계정 $DB_USER 만들기"
	admin_sql "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';
CREATE USER IF NOT EXISTS '$DB_USER'@'127.0.0.1' IDENTIFIED BY '$DB_PASSWORD';
ALTER USER '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';
ALTER USER '$DB_USER'@'127.0.0.1' IDENTIFIED BY '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'localhost';
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'127.0.0.1';"
fi

# 2) 마이그레이션 — db/migrations/NNNN_이름.up.sql 을 번호 순으로, 적용 기록은 schema_migrations
app_sql -e "CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(100) NOT NULL PRIMARY KEY, appliedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)"
# 플러그인 빼기(zbd remove)가 남긴 DB 지우기 — 마이그레이션보다 먼저, 파일마다 한 번(schema_migrations 'remove:이름').
#   파일 머리 「-- migrations: …」 가 하나도 적용 안 된 DB(그 플러그인을 붙인 적 없음)면 돌리지 않고 기록만 한다
for f in "$ROOT"/db/removals/*.sql; do
	[ -f "$f" ] || continue
	v="remove:$(basename "$f" .sql)"
	[ "$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")" = 0 ] || continue
	had=0; vs="$(sed -n 's/^-- migrations: //p' "$f" | head -1)"
	[ -z "$vs" ] && had=1
	for m in $vs; do [ "$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$m'")" = 0 ] || had=1; done
	if [ "$had" = 1 ]; then echo "· 플러그인 빼기 $(basename "$f" .sql)"; app_sql <"$f"; fi
	app_sql -e "INSERT INTO schema_migrations (version) VALUES ('$v')"
done
for f in "$ROOT"/db/migrations/*.up.sql; do
	v="$(basename "$f" .up.sql)"
	done_v="$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")"
	[ "$done_v" = "0" ] || continue
	echo "· 마이그레이션 $v"
	app_sql < "$f"
	app_sql -e "INSERT INTO schema_migrations (version) VALUES ('$v')"
done

# 3) 기본 데이터 — INSERT IGNORE 라 이미 있는 행(순서·이름을 바꾼 메뉴)은 그대로.
#    관리자가 지운 메뉴·권한은 다시 넣지 않게 seed-guard 로 감싸 한 세션에서 돌린다(db/seed-guard/build.sh — 새로·다시 붙인 플러그인의 메뉴는 되살린다)
sh "$ROOT"/db/seed-guard/build.sh "$ROOT"/db/seed | app_sql
echo "✓ DB 준비됨 ($DB_NAME)"
