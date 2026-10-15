#!/usr/bin/env bash
# 처음 설치 상태로 되돌리기 — 내 컴퓨터의 개발 DB 와 .dev/(계정·업로드·웹하드)를 지운다. 다음 ./run_dev.sh 가 새로 만든다.
#   ./reset.sh          DB 지우기 + .dev/ 지우기 (설정 .env.local 은 그대로 — 같은 DB 이름·계정·키로 다시 만든다)
#   ./reset.sh --all    + .env.local·DB 계정까지 지움 (암호화 키도 새로 만든다)
#   ./reset.sh -y       묻지 않고 진행
#
# ⚠ 되돌릴 수 없다 — 이 DB 의 회원·글·업무 카드·올린 파일이 전부 사라진다. 운영 서버에서 돌리지 말 것.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$ROOT/zbd-be-golang/.env.local"
DEV="$ROOT/.dev"
ALL=0
YES=0
for a in "$@"; do
	case "$a" in
	--all) ALL=1 ;;
	-y | --yes) YES=1 ;;
	-h | --help) sed -n '2,7p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
	*) echo "모르는 옵션: $a (./reset.sh --help)"; exit 1 ;;
	esac
done

say() { printf '\033[1;36m▶ %s\033[0m\n' "$*"; }
die() { printf '\033[1;31m✗ %s\033[0m\n' "$*"; exit 1; }

# .env.local 에서 값 하나 읽기(따옴표 벗김)
envget() { [ -f "$ENV_FILE" ] && sed -n "s/^$1=//p" "$ENV_FILE" | tail -1 | sed 's/^"\(.*\)"$/\1/' || true; }
DB_NAME="$(envget DB_NAME)"
DB_USER="$(envget DB_USER)"
BE_PORT="$(envget PORT)"
DB_NAME="${DB_NAME:-zbd_starter}"
BE_PORT="${BE_PORT:-5001}"

# 운영 DB 이름이면 거절 — 실수로 운영 .env 를 가리키는 경우
case "$DB_NAME" in ZBD | zbd | mysql | information_schema | performance_schema | sys) die "DB 이름이 '$DB_NAME' — 개발 DB 가 아니라서 멈춥니다" ;; esac

# 서버가 떠 있으면 먼저 끄게 한다(열린 연결·쓰는 중인 파일)
if command -v lsof >/dev/null && lsof -nP -iTCP:"$BE_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
	die "포트 $BE_PORT 에 서버가 떠 있습니다 — run_dev.sh 를 Ctrl+C 로 끈 뒤 다시 실행하세요"
fi

echo "지울 것:"
echo "  · DB $DB_NAME (표·데이터 전부)"
echo "  · $DEV/ (계정 파일·업로드·첨부·웹하드)"
[ "$ALL" = 1 ] && echo "  · $ENV_FILE (암호화 키 포함) · DB 계정 ${DB_USER:-zbd_dev}"
if [ "$YES" != 1 ]; then
	printf '정말 지울까요? 되돌릴 수 없습니다 [y/N] '
	read -r ans
	[ "$ans" = "y" ] || [ "$ans" = "Y" ] || { echo "그만둡니다."; exit 0; }
fi

# MariaDB 관리자 권한 — setup.sh 와 같게(macOS brew = 지금 사용자, Linux = sudo)
admin_sql() { if [ "$(uname)" = "Darwin" ]; then mariadb --skip-ssl -e "$1"; else sudo mariadb --skip-ssl -e "$1"; fi; }
mariadb-admin ping >/dev/null 2>&1 || (sudo -n true 2>/dev/null && sudo mariadb-admin ping >/dev/null 2>&1) || die "MariaDB 가 꺼져 있습니다 — ./run_dev.sh 를 한 번 실행하거나 brew services start mariadb"

say "DB $DB_NAME 지우기"
admin_sql "DROP DATABASE IF EXISTS \`$DB_NAME\`;"
if [ "$ALL" = 1 ] && [ -n "$DB_USER" ]; then
	say "DB 계정 $DB_USER 지우기"
	admin_sql "DROP USER IF EXISTS '$DB_USER'@'localhost'; DROP USER IF EXISTS '$DB_USER'@'127.0.0.1';"
fi

say ".dev/ 지우기"
rm -rf "$DEV"
if [ "$ALL" = 1 ]; then
	say ".env.local 지우기"
	rm -f "$ENV_FILE"
fi

printf '\033[1;32m✓ 처음 상태입니다. ./run_dev.sh 로 다시 띄우면 DB·예시 데이터·새 계정(.dev/account.txt)이 만들어집니다.\033[0m\n'
