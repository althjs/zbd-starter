#!/usr/bin/env bash
# 서버 설치·갱신 — deploy.sh 가 꾸러미와 함께 올려 sudo 로 실행한다(여러 번 돌려도 같다). Ubuntu 22.04·24.04.
#   bash install.sh --pkg <꾸러미 폴더> --name zbd-starter [--domain example.com --admin admin.example.com[,…]] [--admin-apps "chat mail"]
#   bash /srv/<이름>/install.sh --rollback --name <이름>   앞 판으로 되돌리기(zbd deploy --rollback) — DB 는 그대로
#   꾸러미에 dev-db.sql.gz(배포한 컴퓨터의 개발 DB)가 있으면: 서버 DB 가 비었을 때(처음 설치) 그것을 넣는다 · --replace-db 면 서버 DB 를 백업하고 바꾼다.
#     DB 이름·계정·비밀번호는 서버 .env 그대로, 주소 설정(domains·홈 주소·쿠키 범위)만 서버 값(처음이면 --domain)으로 — 2026-10-20
#   DB 서버(2026-10-20): .env 의 DB_HOST·DB_PORT 를 따른다 — 127.0.0.1·localhost 면 이 서버 MariaDB(DB·계정을 만든다), 아니면 **다른 DB 서버**
#     (DB·계정은 그 서버에서 미리 만든다 — 이 스크립트는 접속만 확인, MariaDB 서버는 설치하지 않고 접속 도구만). 꾸러미에 db-conn(deploy.sh --db-host)이 있으면
#     .env 의 DB_* 를 그 값으로 바꾼다(앞 .env 백업) — 새 DB 가 비었으면 앞 DB 를 그대로 옮기고, 주소 설정이 없거나 개발 값이면 앞 DB 의 주소를 넣는다
#   새 판의 서버 프로그램이 뜨지 않으면(시작하자마자 죽음·응답 없음) 묻지 않고 앞 판으로 되돌린다(2026-10-11)
#   주소의 원본 = DB 사이트 설정 domains(관리자 「환경 설정 › 도메인」 — 2026-10-06 「하나로만」). --domain·--admin 은 **처음 설치 때만** 쓴다
#   (DB 에 주소가 이미 있으면 무시 — 화면에서 바꾼 주소를 옛 배포 설정이 덮지 않게). nginx·인증서는 DB 주소대로 맞춘다.
#   --admin-apps = 플러그인이 필요로 하는 관리자 단독 앱 주소의 앞이름(chat → chat.<대표 사용자 주소>) — 없으면 DB 관리자 주소에 더한다
#
# /srv/<이름>/
#   .env                 설정(처음 한 번 — 키는 무작위. 잃으면 회원 개인정보를 못 읽는다 → 따로 보관)
#   releases/<시각>/      배포한 판(bin·web·db) — 마지막 3개만 남긴다 · current → 지금 판
#   data/upload|secure|file-explorer   올린 파일(배포와 무관하게 남는다)
#   socket/.env          소켓 서버 설정(웹소켓 플러그인이 있을 때)
#   account.txt          첫 관리자 비밀번호(600) — 처음 한 번
# 서비스: <이름>(서버 프로그램 127.0.0.1:5001~) · <이름>-socket(소켓 서버 127.0.0.1:4006~) · nginx 사이트 <이름> — 포트는 비어 있는 것을 골라 .env 에 남긴다
set -euo pipefail
PKG="" NAME="zbd-starter" DOMAIN="" ADMIN="" ADMIN_APPS="" BE_PORT_ARG="" SOCK_PORT_ARG="" SSL_EMAIL="" SYNC_FROM="" ROLLBACK=0 REPLACE_DB=0
while [ $# -gt 0 ]; do
	case "$1" in
		--pkg) PKG="$2"; shift 2 ;; --name) NAME="$2"; shift 2 ;;
		--domain) DOMAIN="$2"; shift 2 ;; --admin) ADMIN="$2"; shift 2 ;; --admin-apps) ADMIN_APPS="$2"; shift 2 ;;
		--port) BE_PORT_ARG="$2"; shift 2 ;; --socket-port) SOCK_PORT_ARG="$2"; shift 2 ;;
		--ssl) SSL_EMAIL="$2"; shift 2 ;;
		--sync-from) SYNC_FROM="$2"; shift 2 ;; # 개발 컴퓨터에서 올린 파일(deploy.sh 가 보낸 곳 — upload·secure·file-explorer)
		--rollback) ROLLBACK=1; shift ;;
		--replace-db) REPLACE_DB=1; shift ;; # 서버 DB 를 꾸러미의 개발 DB 로 바꾼다(먼저 백업)
		*) echo "✗ 모르는 옵션: $1"; exit 1 ;;
	esac
done
[ "$(id -u)" = 0 ] || { echo "✗ sudo 로 실행하세요"; exit 1; }
[[ "$NAME" =~ ^[a-z][a-z0-9-]{1,30}$ ]] || { echo "✗ --name 은 영문 소문자·숫자·- 만"; exit 1; }
APP="/srv/$NAME"

# ─── 판 바꾸기 · 확인 · 되돌리기 ──────────────────────────
# releases 이름 = 배포 시각(이름 순 = 시간 순). 실패해 되돌린 판은 .failed, 손으로 되돌린 판은 .rolledback 을 붙여 다시 고르지 않는다
say() { printf '\033[1;36m▶ %s\033[0m\n' "$*"; }
rel_rev() { cat "$1/REVISION" 2>/dev/null || echo "-"; }
rel_prev() { # $1 = 지금 판 → 그보다 앞의 정상 판 중 가장 최근(없으면 빈 값)
	local cur r prev=""
	cur="$(basename "$1")"
	for r in $(ls -1 "$APP/releases" 2>/dev/null | grep -E '^[0-9]{8}-[0-9]{6}$' | sort || true); do
		[[ "$r" < "$cur" ]] || break
		prev="$r"
	done
	[ -z "$prev" ] || echo "$APP/releases/$prev"
}
rel_switch() { ln -sfn "$1" "$APP/current.new" && mv -T "$APP/current.new" "$APP/current"; }
site_restart() {
	systemctl restart "$NAME"
	if [ -f "/etc/systemd/system/$NAME-socket.service" ] && [ -x "$APP/current/bin/zbd-socket" ]; then systemctl restart "$NAME-socket"; fi
}
# site_up — 서버 프로그램이 응답하고 3초 뒤에도 살아 있는지(시작하자마자 죽는 판을 잡는다) · 소켓 서버가 있으면 healthz 도. 최대 30초
site_up() {
	local port sport i code
	port="$(grep -E '^PORT=' "$APP/.env" | tail -1 | cut -d= -f2-)"
	for i in $(seq 1 30); do
		sleep 1
		code="$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$port/service/design/ZBD/base" || true)"
		case "$code" in 2* | 3* | 4*) break ;; esac
		[ "$i" = 30 ] && return 1
	done
	sleep 3
	systemctl is-active --quiet "$NAME" || return 1
	code="$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$port/service/design/ZBD/base" || true)"
	case "$code" in 2* | 3* | 4*) ;; *) return 1 ;; esac
	if [ -f "/etc/systemd/system/$NAME-socket.service" ] && [ -x "$APP/current/bin/zbd-socket" ]; then
		sport="$(grep -E '^SOCKET_URL=' "$APP/.env" | tail -1 | sed -nE 's#.*:([0-9]+)$#\1#p')"
		[ -n "$sport" ] || return 0
		for i in $(seq 1 15); do
			[ "$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$sport/healthz" || true)" = 200 ] && return 0
			sleep 1
		done
		return 1
	fi
	return 0
}
rel_list() {
	local r mark cur
	cur="$(basename "$(readlink -f "$APP/current" 2>/dev/null || echo -)")"
	echo "  서버에 남은 판(최근 것 위):"
	for r in $(ls -1 "$APP/releases" 2>/dev/null | sort -r || true); do
		mark=""; [ "$r" = "$cur" ] && mark=" ← 지금"
		printf '    %s  판 %s%s\n' "$r" "$(rel_rev "$APP/releases/$r")" "$mark"
	done
}
if [ "$ROLLBACK" = 1 ]; then
	[ -L "$APP/current" ] || { echo "✗ $APP 에 배포한 판이 없습니다"; exit 1; }
	CUR="$(readlink -f "$APP/current")"
	PREV="$(rel_prev "$CUR")"
	if [ -z "$PREV" ] || [ ! -d "$PREV" ]; then
		echo "✗ 되돌릴 앞 판이 없습니다(서버에는 최근 판 3개만 남깁니다)"
		rel_list
		exit 1
	fi
	say "앞 판으로 되돌리기 — $(basename "$CUR") (판 $(rel_rev "$CUR")) → $(basename "$PREV") (판 $(rel_rev "$PREV"))"
	rel_switch "$PREV"
	case "$(basename "$CUR")" in *.*) ;; *) mv "$CUR" "$CUR.rolledback" ;; esac
	site_restart
	if site_up; then
		echo "✓ 되돌렸습니다 — 사이트는 판 $(rel_rev "$PREV") 으로 돕니다"
		echo "  DB 는 그대로입니다(새 판이 더한 표·칸은 남음 — 앞 판은 그것을 쓰지 않습니다)"
		echo "  다시 새 판으로: 고친 뒤 zbd deploy"
		exit 0
	fi
	echo "✗ 앞 판도 응답하지 않습니다 — journalctl -u $NAME -n 50"
	rel_list
	exit 1
fi
[ -d "$PKG/bin" ] || { echo "✗ --pkg 가 필요합니다"; exit 1; }
# 포트 — 설정(.env)에 이미 있으면 그 값(다시 배포해도 그대로), 처음이면 이 서버의 다른 사이트·쓰는 중인 포트와 겹치지 않는 값
#   (한 서버에 사이트 여럿 — 2026-10-05). 직접 정하려면 처음 설치 때 --port · --socket-port
port_taken() {
	ss -ltnH "sport = :$1" 2>/dev/null | grep -q . && return 0
	for f in /srv/*/.env /srv/*/socket/.env; do
		[ "$f" = "$APP/.env" ] && continue
		grep -qsE "^(PORT|SOCKET_PORT)=$1\$|127\.0\.0\.1:$1(\$|[^0-9])" "$f" && return 0
	done
	return 1
}
pick_port() { local p=$1; while port_taken "$p"; do p=$((p + 1)); done; echo "$p"; }
old_env() { [ -f "$APP/.env" ] && grep -E "^$1=" "$APP/.env" | tail -1 | cut -d= -f2- || true; }
BE_PORT="${BE_PORT_ARG:-$(old_env PORT)}"
[ -n "$BE_PORT" ] || BE_PORT="$(pick_port 5001)"
SOCK_PORT="${SOCK_PORT_ARG:-$(old_env SOCKET_URL | sed -nE 's#.*:([0-9]+)$#\1#p')}"
[ -n "$SOCK_PORT" ] || SOCK_PORT="$(pick_port 4006)"
# 다시 배포 — 적어 둔 포트를 이 사이트가 아닌 프로그램이 듣고 있으면(설치 뒤 다른 프로그램이 차지) 빈 포트를 새로 고른다.
#   이 사이트 서비스($NAME · $NAME-socket)가 듣는 것은 정상(다시 뜰 자리). 바뀐 포트는 아래에서 .env · 소켓 설정 · nginx 에 같이 들어간다
port_foreign() {
	local pid
	for pid in $(ss -ltnpH "sport = :$1" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2 | sort -u); do
		grep -qsE "/($NAME|$NAME-socket)\.service" "/proc/$pid/cgroup" || return 0
	done
	return 1
}
if [ -z "$BE_PORT_ARG" ] && port_foreign "$BE_PORT"; then
	p="$(pick_port $((BE_PORT + 1)))"
	echo "  ⚠ 서버 포트 $BE_PORT 를 다른 프로그램이 쓰고 있어 $p 로 옮깁니다"
	BE_PORT="$p"
fi
if [ -z "$SOCK_PORT_ARG" ] && port_foreign "$SOCK_PORT"; then
	p="$(pick_port $((SOCK_PORT + 1)))"
	echo "  ⚠ 소켓 포트 $SOCK_PORT 를 다른 프로그램이 쓰고 있어 $p 로 옮깁니다"
	SOCK_PORT="$p"
fi
rnd() { local s; s="$(head -c 96 /dev/urandom | base64 | tr -dc 'A-Za-z0-9')"; printf '%s' "${s:0:$1}"; }
b64() { head -c 32 /dev/urandom | base64 | tr -d '\n'; }
HAS_SOCK=0; [ -x "$PKG/bin/zbd-socket" ] && HAS_SOCK=1

# DB 서버 — 꾸러미의 db-conn(이번에 바꿀 값) → 지금 .env → 이 서버. 다른 서버면 MariaDB 서버를 설치·켜지 않는다
connget() { [ -f "$PKG/db-conn" ] && grep -E "^$1=" "$PKG/db-conn" | tail -1 | cut -d= -f2- || true; }
DB_HOST_NEXT="$(connget DB_HOST)"; [ -n "$DB_HOST_NEXT" ] || DB_HOST_NEXT="$(old_env DB_HOST)"
case "${DB_HOST_NEXT:-127.0.0.1}" in 127.0.0.1 | localhost) DB_REMOTE=0 ;; *) DB_REMOTE=1 ;; esac

# ─── 1. 프로그램 ─────────────────────────────────────────
if ! command -v nginx >/dev/null || ! command -v mariadb >/dev/null; then
	export DEBIAN_FRONTEND=noninteractive
	apt-get update -qq
	if [ "$DB_REMOTE" = 1 ]; then
		say "nginx · MariaDB 접속 도구 설치(apt — DB 는 다른 서버 $DB_HOST_NEXT)"
		apt-get install -y -qq nginx mariadb-client curl >/dev/null
	else
		say "nginx · MariaDB 설치(apt)"
		apt-get install -y -qq nginx mariadb-server curl >/dev/null
	fi
fi
[ "$DB_REMOTE" = 1 ] || systemctl enable --now mariadb >/dev/null 2>&1 || true
systemctl enable --now nginx >/dev/null 2>&1 || true
id -u "$NAME" >/dev/null 2>&1 || useradd --system --home-dir "$APP" --shell /usr/sbin/nologin "$NAME"

# ─── 2. 폴더 · 설정(처음 한 번) ──────────────────────────
mkdir -p "$APP/releases" "$APP/data/upload" "$APP/data/secure" "$APP/data/file-explorer" "$APP/socket"
if [ ! -f "$APP/.env" ]; then
	say "설정 만들기($APP/.env — 키는 무작위)"
	umask 077
	cat >"$APP/.env" <<ENV
# install.sh 가 만든 운영 설정 — 키를 잃으면 회원 개인정보를 못 읽는다. 따로 보관할 것.
PORT=$BE_PORT
DB_HOST=$( [ -f "$PKG/db-conn" ] && connget DB_HOST || echo 127.0.0.1 )
DB_PORT=$( [ -f "$PKG/db-conn" ] && connget DB_PORT || echo 3306 )
DB_NAME=$( [ -f "$PKG/db-conn" ] && connget DB_NAME || echo "${NAME//-/_}" )
DB_USER=$( [ -f "$PKG/db-conn" ] && connget DB_USER || echo "${NAME//-/_}" )
DB_PASSWORD=$( [ -f "$PKG/db-conn" ] && connget DB_PASSWORD || rnd 28 )
DB_CHARSET=utf8mb4
DB_CONN_LIMIT=20
AES_256_ENCRYPTION_KEY=$(rnd 32)
AES_256_ENCRYPTION_IV=$(rnd 16)
PII_ENC_KEY_V1=$(b64)
PII_INDEX_KEY=$(b64)
AUTH_COOKIE_KEY=zbd_admin_session
AUTH_COOKIE_KEY_USER=zbd_user_session
CDN_UPLOAD_DIR=$APP/data/upload/
FE_INDEX_HTML=$APP/current/web/index.html
RUN_JOBS=true
ENV
	umask 022
	DB_NEW_ENV=1
fi
envget() { grep -E "^$1=" "$APP/.env" | tail -1 | cut -d= -f2-; }
envset() { grep -qE "^$1=" "$APP/.env" || echo "$1=$2" >>"$APP/.env"; }
# 관리자 주소는 이제 DB(도메인 설정) 한 곳 — 옛 설정 줄은 지운다(서버 프로그램은 DB 에 주소가 없을 때만 이 값을 읽었다)
sed -i '/^ADMIN_HOSTS=/d' "$APP/.env"
sed -i "s|^PORT=.*|PORT=$BE_PORT|" "$APP/.env"
if [ "$HAS_SOCK" = 1 ]; then
	envset SOCKET_URL "http://127.0.0.1:$SOCK_PORT"
	sed -i "s|^SOCKET_URL=.*|SOCKET_URL=http://127.0.0.1:$SOCK_PORT|" "$APP/.env"
	envset SOCKET_SECRET "$(rnd 32)"
	envset READY_SOCKET_SECRET "$(rnd 32)" # 대기열 관리 채널 서명(대기열 관리 플러그인 — 없는 기능이면 쓰이지 않는다)
fi
# 시스템 모니터 플러그인(zbd add system-monitor)이 읽는 서비스 — 없는 기능이면 쓰이지 않는다
SVCS="$NAME"; [ "$HAS_SOCK" = 1 ] && SVCS="$NAME,$NAME-socket"
envset SYSTEM_MONITOR_SERVICES "$SVCS"
sed -i "s|^SYSTEM_MONITOR_SERVICES=.*|SYSTEM_MONITOR_SERVICES=$SVCS|" "$APP/.env"
# E2E 테스트 플러그인(zbd add e2e-test — 메뉴 seed 가 있으면) — 결과 올리기 비밀값. 배포한 컴퓨터의 개발 설정 값(deploy.sh 가 실어 보냄)과 같게, 없으면 무작위
if [ -f "$PKG/db/seed/0120_e2e-test.sql" ]; then
	E2E_S="$(cat "$PKG/e2e-upload-secret" 2>/dev/null || true)"
	if [[ "$E2E_S" =~ ^[A-Za-z0-9_-]{16,128}$ ]]; then
		envset E2E_UPLOAD_SECRET "$E2E_S"
		sed -i "s|^E2E_UPLOAD_SECRET=.*|E2E_UPLOAD_SECRET=$E2E_S|" "$APP/.env"
	else
		envset E2E_UPLOAD_SECRET "$(rnd 32)"
	fi
fi
rm -f "$PKG/e2e-upload-secret"
# 외부 공격 감지 플러그인(zbd add attack-block — 메뉴 seed 가 있으면) — 이 사이트 nginx 로그를 서버 프로그램이 읽고(adm 그룹),
#   관리자 화면의 차단 목록은 이 사이트 전용 적용 스크립트 하나만 sudo 로(인자 없음 — 목록을 다시 검사해 conf 생성 → nginx -t → reload, 실패하면 되돌림)
HAS_ATTACK=0; [ -f "$PKG/db/seed/0120_attack-block.sql" ] && HAS_ATTACK=1
DENY_SCRIPT="/usr/local/sbin/zbd-nginx-deny-$NAME"
if [ "$HAS_ATTACK" = 1 ]; then
	usermod -aG adm "$NAME" # /var/log/nginx 읽기(우분투 root:adm 640)
	for kv in "NGINX_ATTACK_STORE=db" "NGINX_ATTACK_LOG_DIR=/var/log/nginx" "NGINX_ATTACK_LOG_GLOB=$NAME.access.log" \
		"NGINX_DENY_DIR=$APP/data/nginx-deny" "NGINX_DENY_CONF=/etc/nginx/conf.d/zz-deny-$NAME.conf" "NGINX_DENY_APPLY_SCRIPT=$DENY_SCRIPT"; do
		envset "${kv%%=*}" "${kv#*=}"
		sed -i "s|^${kv%%=*}=.*|$kv|" "$APP/.env"
	done
	mkdir -p "$APP/data/nginx-deny"
	cat >"$DENY_SCRIPT.tmp" <<'DENY'
#!/bin/bash
# __NAME__ 공격 IP 차단 목록 적용 — install.sh 가 만든다(root 755, sudoers 에 이 파일만). 인자 없음.
#   관리자 화면(외부 공격 감지)이 __APP__/data/nginx-deny/deny.txt 를 고치고 이 스크립트를 부른다.
#   ① 목록 검사(IP·CIDR 형식만, 루프백·내부망·사무실 IP(.env OFFICE_IPS) 거부) ② conf 생성 ③ nginx -t → reload ④ 실패하면 이전 conf 로 되돌림.
#   차단은 이 서버의 모든 사이트에 걸린다(nginx http 단계 conf.d).
set -uo pipefail
SRC=__APP__/data/nginx-deny/deny.txt
CONF=/etc/nginx/conf.d/zz-deny-__NAME__.conf
PREV=/etc/nginx/zz-deny-__NAME__.conf.prev # conf.d 밖(include 되지 않게)
OFFICE_IPS="__OFFICE__"
exec 9>/run/zbd-nginx-deny-__NAME__.lock
flock -w 30 9 || { echo "다른 적용이 진행 중" >&2; exit 4; }
[ -f "$SRC" ] || { echo "목록 파일 없음: $SRC" >&2; exit 2; }
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
{
	echo "# 공격 IP 차단 — 관리자 화면(외부 공격 감지)이 관리한다. 직접 고치면 다음 적용 때 덮어쓴다."
	echo "# 적용: $(date '+%Y-%m-%d %H:%M:%S')"
} >"$TMP"
n=0
while IFS= read -r line || [ -n "$line" ]; do
	ip="${line%%[[:space:]#]*}"
	[ -z "$ip" ] && continue
	if [[ "$ip" =~ ^([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})(/([89]|[12][0-9]|3[0-2]))?$ ]]; then
		for o in "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}" "${BASH_REMATCH[4]}"; do
			[ "$o" -le 255 ] || { echo "잘못된 IP: $ip" >&2; exit 2; }
		done
		a=${BASH_REMATCH[1]}; b=${BASH_REMATCH[2]}
		if [ "$a" = 127 ] || [ "$a" = 10 ] || [ "$a" = 0 ] || { [ "$a" = 192 ] && [ "$b" = 168 ]; } || { [ "$a" = 172 ] && [ "$b" -ge 16 ] && [ "$b" -le 31 ]; }; then
			echo "내부망·루프백은 차단할 수 없음: $ip" >&2; exit 2
		fi
		for off in $OFFICE_IPS; do [ "${ip%%/*}" = "$off" ] && { echo "사무실 IP 는 차단할 수 없음: $ip" >&2; exit 2; }; done
	elif [[ "$ip" =~ ^[0-9a-fA-F:]+(/[0-9]{1,3})?$ ]] && [[ "$ip" == *:* ]]; then
		case "${ip,,}" in ::1*|fe80*|fc*|fd*) echo "내부 IPv6 는 차단할 수 없음: $ip" >&2; exit 2;; esac
	else
		echo "잘못된 형식: $ip" >&2; exit 2
	fi
	echo "deny $ip;" >>"$TMP"
	n=$((n + 1))
done <"$SRC"
[ -f "$CONF" ] && cp -p "$CONF" "$PREV"
install -m 644 -o root -g root "$TMP" "$CONF"
if OUT=$(nginx -t 2>&1); then
	systemctl reload nginx && echo "OK 차단 ${n}건 적용"
else
	echo "nginx -t 실패 — 이전 설정으로 되돌림" >&2
	echo "$OUT" >&2
	if [ -f "$PREV" ]; then install -m 644 -o root -g root "$PREV" "$CONF"; else rm -f "$CONF"; fi
	exit 3
fi
DENY
	sed -i -e "s|__APP__|$APP|g" -e "s|__NAME__|$NAME|g" -e "s|__OFFICE__|$(envget OFFICE_IPS | tr ',' ' ')|" "$DENY_SCRIPT.tmp"
	install -m 755 -o root -g root "$DENY_SCRIPT.tmp" "$DENY_SCRIPT"; rm -f "$DENY_SCRIPT.tmp"
fi
# 배포한 컴퓨터의 설정(deploy.sh 가 dev-env 로 보냄 — 서버가 정하는 값·localhost 값은 빼고 옴): 같은 이름은 그 값으로, 없으면 더한다. 서버에만 있는 줄은 둔다
if [ -f "$PKG/dev-env" ]; then
	CH="$(awk -F= 'NR==FNR{k=$1; sub(/^[^=]*=/,""); v[k]=$0; next} {k=$1; sub(/^[^=]*=/,""); if(k in v){seen[k]=1; if(v[k]!=$0) print k}} END{for(k in v) if(!(k in seen)) print k}' "$PKG/dev-env" "$APP/.env" | grep -E '^[A-Z][A-Z0-9_]*$' | sort | tr '\n' ' ' || true)"
	if [ -n "$CH" ]; then
		install -m 600 -o root -g root "$APP/.env" "$APP/.env.bak_$(date +%Y%m%d-%H%M%S)_devenv"
		awk -F= 'NR==FNR{if($0 ~ /^[A-Z][A-Z0-9_]*=/){k=$1; sub(/^[^=]*=/,""); v[k]=$0; o[++n]=k}; next}
			{k=$1} ($0 ~ /^[A-Z][A-Z0-9_]*=/ && k in v){print k "=" v[k]; done[k]=1; next} {print}
			END{for(i=1;i<=n;i++) if(!(o[i] in done)){print o[i] "=" v[o[i]]; done[o[i]]=1}}' "$PKG/dev-env" "$APP/.env" >"$APP/.env.new"
		mv "$APP/.env.new" "$APP/.env"
		echo "  · 설정 맞춤(개발 컴퓨터 값): $CH"
	fi
	rm -f "$PKG/dev-env"
fi
chown root:"$NAME" "$APP/.env"; chmod 640 "$APP/.env"

# ─── 3. 새 판 ───────────────────────────────────────────
REL="$APP/releases/$(date +%Y%m%d-%H%M%S)"
say "새 판 $(cat "$PKG/REVISION" 2>/dev/null || echo -) → $REL"
mkdir -p "$REL"
cp -a "$PKG/bin" "$PKG/web" "$PKG/db" "$REL/"
[ -f "$PKG/REVISION" ] && cp "$PKG/REVISION" "$REL/"
chmod 755 "$REL/bin/"*
PREV_REL="$(readlink -f "$APP/current" 2>/dev/null || true)"
rel_switch "$REL"
# 되돌리기용 — 이 스크립트를 서버에 남긴다(zbd deploy --rollback 이 부른다)
install -m 700 -o root -g root "$PKG/install.sh" "$APP/install.sh"

# ─── 4. DB ──────────────────────────────────────────────
# DB 서버 바꾸기(deploy.sh --db-host) — 앞 접속값을 떠 두고 .env 의 DB_* 를 바꾼다. 아래에서 새 DB 에 못 붙으면 .env 를 되돌리고 멈춘다
DB_SWITCHED=0
if [ -f "$PKG/db-conn" ] && [ "${DB_NEW_ENV:-0}" != 1 ]; then
	OH="$(envget DB_HOST)" OP="$(envget DB_PORT)" ON="$(envget DB_NAME)" OU="$(envget DB_USER)" OPW="$(envget DB_PASSWORD)"
	if [ "$(connget DB_HOST):$(connget DB_PORT)/$(connget DB_NAME)/$(connget DB_USER)/$(connget DB_PASSWORD)" != "$OH:${OP:-3306}/$ON/$OU/$OPW" ]; then
		ENV_BAK="$APP/.env.bak_$(date +%Y%m%d-%H%M%S)_db"
		install -m 600 -o root -g root "$APP/.env" "$ENV_BAK"
		for k in DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD; do
			v="$(connget "$k")"
			if grep -qE "^$k=" "$APP/.env"; then sed -i "s|^$k=.*|$k=$v|" "$APP/.env"; else echo "$k=$v" >>"$APP/.env"; fi
		done
		DB_SWITCHED=1
		say "DB 서버 바꾸기: ${OH:-127.0.0.1}/$ON → $(connget DB_HOST):$(connget DB_PORT)/$(connget DB_NAME)(앞 설정 백업 $ENV_BAK)"
	fi
fi
rm -f "$PKG/db-conn"
DBN="$(envget DB_NAME)" DBU="$(envget DB_USER)" DBP="$(envget DB_PASSWORD)" DBH="$(envget DB_HOST)" DBPORT="$(envget DB_PORT)"
DBH="${DBH:-127.0.0.1}" DBPORT="${DBPORT:-3306}"
case "$DBH" in 127.0.0.1 | localhost) DB_REMOTE=0 ;; *) DB_REMOTE=1 ;; esac
app_sql() { MYSQL_PWD="$DBP" mariadb --default-character-set=utf8mb4 -h"$DBH" -P"$DBPORT" -u"$DBU" "$DBN" "$@"; }
# 앞 DB(바꾸기 전) — 로컬이면 root, 다른 서버면 앞 계정으로
old_sql() {
	case "${OH:-127.0.0.1}" in
		127.0.0.1 | localhost) mariadb --default-character-set=utf8mb4 "$ON" "$@" ;;
		*) MYSQL_PWD="$OPW" mariadb --default-character-set=utf8mb4 -h"$OH" -P"${OP:-3306}" -u"$OU" "$ON" "$@" ;;
	esac
}
old_dump() {
	case "${OH:-127.0.0.1}" in
		127.0.0.1 | localhost) mariadb-dump --single-transaction --default-character-set=utf8mb4 "$ON" ;;
		*) MYSQL_PWD="$OPW" mariadb-dump --single-transaction --no-tablespaces --default-character-set=utf8mb4 -h"$OH" -P"${OP:-3306}" -u"$OU" "$ON" ;;
	esac
}
# 서버 DB 백업(이 DB 의 지금 내용) — 다른 서버면 앱 계정으로
db_dump() {
	if [ "$DB_REMOTE" = 1 ]; then MYSQL_PWD="$DBP" mariadb-dump --single-transaction --no-tablespaces --default-character-set=utf8mb4 -h"$DBH" -P"$DBPORT" -u"$DBU" "$DBN"
	else mariadb-dump --single-transaction --default-character-set=utf8mb4 "$DBN"; fi
}
say "DB 준비($DBN$( [ "$DB_REMOTE" = 1 ] && echo " @ $DBH:$DBPORT"))"
if [ "$DB_REMOTE" = 1 ]; then
	if ! app_sql -e "SELECT 1" >/dev/null 2>"$APP/.db-err"; then
		echo "✗ DB 서버($DBH:$DBPORT)의 $DBN 에 계정 $DBU 로 접속하지 못했습니다: $(tail -1 "$APP/.db-err")"
		echo "  다른 DB 서버는 그 서버에서 DB·계정을 먼저 만듭니다(이 서버 IP 에서 들어올 수 있게, 그 DB 에 모든 권한):"
		echo "    CREATE DATABASE $DBN CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; CREATE USER '$DBU'@'<이 서버 IP>' IDENTIFIED BY '…'; GRANT ALL ON $DBN.* TO '$DBU'@'<이 서버 IP>';"
		rm -f "$APP/.db-err"
		if [ "$DB_SWITCHED" = 1 ]; then install -m 600 -o root -g root "$ENV_BAK" "$APP/.env"; echo "  · .env 는 앞 값으로 되돌렸습니다(사이트는 앞 DB 그대로)"; fi
		rel_switch "$PREV_REL" 2>/dev/null || true
		exit 1
	fi
	rm -f "$APP/.db-err"
else
mariadb -e "CREATE DATABASE IF NOT EXISTS \`$DBN\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$DBU'@'localhost' IDENTIFIED BY '$DBP'; CREATE USER IF NOT EXISTS '$DBU'@'127.0.0.1' IDENTIFIED BY '$DBP';
ALTER USER '$DBU'@'localhost' IDENTIFIED BY '$DBP'; ALTER USER '$DBU'@'127.0.0.1' IDENTIFIED BY '$DBP';
GRANT ALL PRIVILEGES ON \`$DBN\`.* TO '$DBU'@'localhost'; GRANT ALL PRIVILEGES ON \`$DBN\`.* TO '$DBU'@'127.0.0.1';"
fi
# DB 서버를 바꿨는데 새 DB 가 비었으면 앞 DB 를 그대로 옮긴다(데이터가 있으면 그대로 쓴다). 주소 설정은 아래 「사이트 주소」 앞에서 맞춘다
OLD_DOMS=""
if [ "$DB_SWITCHED" = 1 ]; then
	OLD_DOMS="$(old_sql -N -e "SELECT JSON_EXTRACT(json, '\$.domains') FROM settings WHERE metaKey='SITE' AND hostingId='ZBD'" 2>/dev/null || true)"
	[[ "$OLD_DOMS" =~ ^[][{}\"a-zA-Z0-9.:,\ _-]+$ ]] || OLD_DOMS=""
	NT0="$(app_sql -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()")"
	if [ "$NT0" = 0 ]; then
		say "새 DB 가 비어 있어 앞 DB 를 옮깁니다"
		old_dump | sed -e '1{/enable the sandbox mode/d;}' -e 's/ DEFINER=[^ ]* / /g' | app_sql ||
			{ echo "✗ 앞 DB 를 옮기지 못했습니다 — .env 를 되돌립니다"; install -m 600 -o root -g root "$ENV_BAK" "$APP/.env"; rel_switch "$PREV_REL" 2>/dev/null || true; exit 1; }
		echo "  · 옮김: 표 $(app_sql -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()")개 · 관리자 $(app_sql -N -e "SELECT COUNT(*) FROM admins" 2>/dev/null || echo 0)명 · 회원 $(app_sql -N -e "SELECT COUNT(*) FROM members" 2>/dev/null || echo 0)명(앞 DB 는 그대로 남음)"
	else
		echo "  · 새 DB 에 이미 데이터가 있어 그대로 씁니다(표 ${NT0}개 — 앞 DB 는 그대로 남음)"
	fi
fi
# 개발 DB 옮기기 — 회원 정보는 개발 키로 잠겨 있어 개발 키(dev-keys)가 같이 와야 넣는다. 넣은 뒤 주소 설정은 지우고(아래 「사이트 주소」 가 다시 만든다)
#   서버 DB 를 바꿀 때는 서버 주소 설정을 먼저 떠 두었다가 되살린다
DEVDB_LOADED=0
if [ -f "$PKG/dev-db.sql.gz" ]; then
	NT="$(app_sql -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name<>'schema_migrations'")"
	if [ ! -f "$PKG/dev-keys" ]; then
		echo "  ⚠ 개발 컴퓨터 키가 같이 오지 않아 개발 DB 는 넣지 않습니다(회원 정보를 못 읽습니다)"
	elif [ "$NT" != 0 ] && [ "$REPLACE_DB" != 1 ]; then
		echo "  · 서버 DB 에 이미 데이터가 있어 개발 DB 는 넣지 않습니다(서버 DB 를 개발 DB 로 바꾸려면: zbd deploy --with-db)"
	else
		SAVED_DOMS=""
		if [ "$NT" != 0 ]; then
			SAVED_DOMS="$(app_sql -N -e "SELECT JSON_EXTRACT(json, '\$.domains') FROM settings WHERE metaKey='SITE' AND hostingId='ZBD'" 2>/dev/null || true)"
			[[ "$SAVED_DOMS" =~ ^[][{}\"a-zA-Z0-9.:,\ _-]+$ ]] || SAVED_DOMS=""
			mkdir -p "$APP/backups/db"; chmod 700 "$APP/backups" "$APP/backups/db"
			B="$APP/backups/db/$(date +%Y%m%d-%H%M%S)-before-dev-db.sql.gz"
			say "서버 DB 백업 → $B"
			db_dump | gzip >"$B" || { echo "✗ 서버 DB 를 백업하지 못해 바꾸지 않았습니다"; exit 1; }
			chmod 600 "$B"
			if [ "$DB_REMOTE" = 1 ]; then # 다른 DB 서버 — 앱 계정은 DB 를 지우고 만들 권한이 없을 수 있어 표·뷰를 지운다
				DROPS="$(app_sql -N -e "SELECT CONCAT('DROP ', IF(table_type='VIEW','VIEW','TABLE'), ' IF EXISTS \`', table_name, '\`;') FROM information_schema.tables WHERE table_schema=DATABASE()")"
				printf 'SET FOREIGN_KEY_CHECKS=0;\n%s\n' "$DROPS" | app_sql
			else
				mariadb -e "DROP DATABASE \`$DBN\`; CREATE DATABASE \`$DBN\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
			fi
		fi
		say "개발 DB 넣기"
		gunzip -c "$PKG/dev-db.sql.gz" | app_sql || { echo "✗ 개발 DB 를 넣지 못했습니다${B:+ — 되살리기: gunzip -c $B | (DB 에 넣기)}"; exit 1; }
		if [ -n "$SAVED_DOMS" ] && [ "$SAVED_DOMS" != NULL ]; then
			app_sql -e "UPDATE settings SET json = JSON_SET(json, '\$.domains', JSON_EXTRACT('$SAVED_DOMS', '\$')) WHERE metaKey='SITE' AND hostingId='ZBD'"
		else
			app_sql -e "UPDATE settings SET json = JSON_REMOVE(json, '\$.domains') WHERE metaKey='SITE' AND hostingId='ZBD'"
		fi
		echo "  · 넣음: 표 $(app_sql -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()")개 · 관리자 $(app_sql -N -e "SELECT COUNT(*) FROM admins" 2>/dev/null || echo 0)명 · 회원 $(app_sql -N -e "SELECT COUNT(*) FROM members" 2>/dev/null || echo 0)명${B:+ · 앞 서버 DB 백업 $B}"
		DEVDB_LOADED=1
	fi
fi
rm -f "$PKG/dev-db.sql.gz"
app_sql -e "CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(100) NOT NULL PRIMARY KEY, appliedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)"
# 플러그인 빼기(zbd remove)가 남긴 DB 지우기 — 마이그레이션보다 먼저, 파일마다 한 번(schema_migrations 'remove:이름').
#   파일 머리 「-- migrations: …」 가 하나도 적용 안 된 DB(그 플러그인을 붙인 적 없음)면 돌리지 않고 기록만 한다
for f in "$REL"/db/removals/*.sql; do
	[ -f "$f" ] || continue
	v="remove:$(basename "$f" .sql)"
	[ "$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")" = 0 ] || continue
	had=0; vs="$(sed -n 's/^-- migrations: //p' "$f" | head -1)"
	[ -z "$vs" ] && had=1
	for m in $vs; do [ "$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$m'")" = 0 ] || had=1; done
	if [ "$had" = 1 ]; then echo "  · 플러그인 빼기 $(basename "$f" .sql)"; app_sql <"$f"; fi
	app_sql -e "INSERT INTO schema_migrations (version) VALUES ('$v')"
done
for f in "$REL"/db/migrations/*.up.sql; do
	v="$(basename "$f" .up.sql)"
	[ "$(app_sql -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")" = 0 ] || continue
	echo "  · 마이그레이션 $v"
	app_sql <"$f"
	app_sql -e "INSERT INTO schema_migrations (version) VALUES ('$v')"
done
# 기본 데이터 — 지운 메뉴·권한은 다시 넣지 않게 seed-guard 로 감싸 한 세션에서(db/seed-guard/build.sh — 새로·다시 붙인 플러그인의 메뉴는 되살린다)
sh "$REL"/db/seed-guard/build.sh "$REL"/db/seed | app_sql
# DB 서버를 바꿨고 새 DB 의 주소 설정이 없거나 개발 값(localhost)이면 앞 DB 의 주소를 넣고 홈 주소·쿠키 범위도 다시 계산한다
FORCE_ADDR=0
if [ "$DB_SWITCHED" = 1 ] && [ -n "$OLD_DOMS" ] && [ "$OLD_DOMS" != NULL ]; then
	CUR_U1="$(app_sql -N -e "SELECT JSON_UNQUOTE(JSON_EXTRACT(json, '\$.domains.user[0]')) FROM settings WHERE metaKey='SITE' AND hostingId='ZBD'")"
	case "$CUR_U1" in "" | NULL | localhost* | *.localhost* | 127.0.0.1*)
		app_sql -e "UPDATE settings SET json = JSON_SET(json, '\$.domains', JSON_EXTRACT('$OLD_DOMS', '\$')) WHERE metaKey='SITE' AND hostingId='ZBD'"
		echo "  · 새 DB 의 주소 설정을 앞 DB 값으로 맞춤"
		FORCE_ADDR=1 ;;
	esac
fi
# 사이트 주소 — 원본은 DB 사이트 설정 domains {user[], admin[]}(관리자 「환경 설정 › 도메인」). 처음이면 --domain·--admin 으로 만든다.
#   홈 주소·쿠키 범위·사이트맵 주소·호스팅 목록 homepage 는 domains 로 계산해 같이 저장한다(서버 web/domains.go 와 같은 규칙)
site_json() { app_sql -N -e "SELECT JSON_UNQUOTE(JSON_EXTRACT(json, '$1')) FROM settings WHERE metaKey='SITE' AND hostingId='ZBD'"; }
jlist() { tr -d '[]" ' | tr ',' ' '; } # ["a","b"] → a b
USER_HOSTS="$(site_json '$.domains.user' | jlist)"
ADMINS="$(site_json '$.domains.admin' | jlist)"
[ "$USER_HOSTS" = NULL ] && USER_HOSTS=""
[ "$ADMINS" = NULL ] && ADMINS=""
# 기본 데이터의 개발 주소(localhost:5000) 는 「아직 없음」 — 서버 첫 설치는 --domain 으로
case "${USER_HOSTS%% *}" in localhost* | *.localhost* | 127.0.0.1*) USER_HOSTS="" ADMINS="" ;; esac
if [ -n "$USER_HOSTS" ] && [ -n "$ADMINS" ]; then
	echo "  주소(DB 도메인 설정): 사용자 $USER_HOSTS · 관리자 $ADMINS"
	{ [ -n "$DOMAIN" ] && [ "$DOMAIN" != "${USER_HOSTS%% *}" ]; } && echo "  · 배포 설정의 --domain $DOMAIN 은 쓰지 않습니다 — 주소는 관리자 「환경 설정 › 도메인」 에서 바꿉니다"
	DOMAIN="${USER_HOSTS%% *}"
	NEW_INIT=0
else
	[ -n "$DOMAIN" ] || { echo "✗ 처음 설치에는 --domain 이 필요합니다(그다음부터는 관리자 「환경 설정 › 도메인」 의 주소를 씁니다)"; exit 1; }
	USER_HOSTS="$DOMAIN"
	ADMINS="$(printf '%s' "${ADMIN:-admin.$DOMAIN}" | tr ',' ' ')"
	NEW_INIT=1
fi
# 플러그인이 필요로 하는 단독 앱 주소 — 없으면 더한다. 앞에 붙일 주소 = 대표 관리자 주소가 사용자 주소 아래면 사용자 주소(chat.example.com),
# 밖이면 대표 관리자 주소(chat.admin.example.com — 관리자 로그인 쿠키를 같이 쓴다, 2026-10-20 오토카지: chat.user.… 는 관리자 admin.… 와 로그인이 안 이어졌다).
# 밖인데 예전 자리(chat.<사용자 주소>)가 있으면 새 자리로 옮긴다(DNS 에 새 이름이 있어야 인증서가 나온다)
ADD=0
A1="${ADMINS%% *}"
APP_BASE="$DOMAIN"
case "$A1" in "$DOMAIN" | *".$DOMAIN") ;; *) APP_BASE="$A1" ;; esac
for a in $ADMIN_APPS; do
	h="$a.$APP_BASE"
	if [ "$APP_BASE" != "$DOMAIN" ]; then
		case " $ADMINS " in *" $a.$DOMAIN "*) ADMINS="$(printf ' %s ' "$ADMINS" | sed "s/ $a\.$DOMAIN / /; s/^ *//; s/ *$//")"; ADD=1; echo "  · 관리자 주소 옮김: $a.$DOMAIN → $h(관리자 로그인을 같이 쓰게)" ;; esac
	fi
	case " $ADMINS " in *" $h "*) ;; *) ADMINS="$ADMINS $h"; ADD=1; echo "  · 관리자 주소 더함: $h(플러그인)" ;; esac
done
ADMIN1="${ADMINS%% *}" # 대표 관리자 주소
jarr() { local o="" h; for h in "$@"; do o="$o${o:+,}'$h'"; done; printf 'JSON_ARRAY(%s)' "$o"; }
DOMS="JSON_OBJECT('user', $(jarr $USER_HOSTS), 'admin', $(jarr $ADMINS))"
if [ "$NEW_INIT" = 1 ] || [ "$DEVDB_LOADED" = 1 ] || [ "$FORCE_ADDR" = 1 ]; then # 개발 DB 를 넣었으면 홈 주소·쿠키 범위·사이트맵도 개발 값(localhost)이라 다시
	URL="https://$DOMAIN/"
	app_sql -e "UPDATE settings SET json = JSON_SET(json, '\$.items[0].homepage', '$URL') WHERE metaKey='HOSTINGS' AND hostingId='';
UPDATE settings SET json = JSON_SET(json, '\$.domains', $DOMS, '\$.cookieHost', '.$DOMAIN', '\$.homepage', '$URL', '\$.design.homepage', '$URL',
  '\$.design.robots', 'User-agent: *\nAllow: /\nDisallow: /admin\n\nSitemap: ${URL}sitemap.xml') WHERE metaKey='SITE' AND hostingId='ZBD';"
elif [ "$ADD" = 1 ]; then
	app_sql -e "UPDATE settings SET json = JSON_SET(json, '\$.domains', $DOMS) WHERE metaKey='SITE' AND hostingId='ZBD';"
fi
# 암호화 키 = 배포한 컴퓨터의 개발 키와 같게(deploy.sh 가 dev-keys 로 보냄 — --server-keys 면 안 옴).
#   다르면: .env 백업 → 서버 DB 의 회원 이름·이메일·전화·관리자 2단계 인증을 새 키로 다시 잠금(bin/rekey, 이미 새 키인 값은 그대로) → .env 키 바꿈.
#   로그인 토큰 키도 바뀌어 모두 한 번 다시 로그인한다. 다시 잠그기가 실패하면 키를 바꾸지 않고 멈춘다
KEYS_CHANGED=0
if [ -f "$PKG/dev-keys" ]; then
	nk() { grep -E "^$1=" "$PKG/dev-keys" | tail -1 | cut -d= -f2-; }
	NK_AES="$(nk AES_256_ENCRYPTION_KEY)" NK_IV="$(nk AES_256_ENCRYPTION_IV)" NK_PE="$(nk PII_ENC_KEY_V1)" NK_PI="$(nk PII_INDEX_KEY)"
	rm -f "$PKG/dev-keys"
	if ! [[ "$NK_AES" =~ ^[A-Za-z0-9]{32}$ && "$NK_IV" =~ ^[A-Za-z0-9]{16}$ && "$NK_PE" =~ ^[A-Za-z0-9+/]{43}=$ && "$NK_PI" =~ ^[A-Za-z0-9+/]{43}=$ ]]; then
		echo "  ⚠ 개발 컴퓨터 키 모양이 달라 서버 키를 그대로 둡니다"
	elif [ "$NK_AES$NK_IV$NK_PE$NK_PI" != "$(envget AES_256_ENCRYPTION_KEY)$(envget AES_256_ENCRYPTION_IV)$(envget PII_ENC_KEY_V1)$(envget PII_INDEX_KEY)" ]; then
		say "암호화 키를 개발 컴퓨터와 맞춥니다"
		BAK="$APP/.env.bak_$(date +%Y%m%d-%H%M%S)_keys"
		install -m 600 -o root -g root "$APP/.env" "$BAK"
		if [ "$DEVDB_LOADED" = 1 ]; then
			echo "  · 개발 DB 를 넣어 회원 정보가 이미 개발 키로 잠겨 있습니다 — 다시 잠그지 않고 키만 바꿉니다"
		else
			OUT="$(cd "$APP" && NEW_PII_ENC_KEY_V1="$NK_PE" NEW_PII_INDEX_KEY="$NK_PI" "$REL/bin/rekey" 2>&1)" || { echo "$OUT"; echo "✗ 회원 정보를 새 키로 다시 잠그지 못해 키를 바꾸지 않았습니다(서버 키 그대로) — 위 오류를 확인하세요"; exit 1; }
			echo "  · $OUT" | sed '2,$s/^/  /'
		fi
		for kv in "AES_256_ENCRYPTION_KEY=$NK_AES" "AES_256_ENCRYPTION_IV=$NK_IV" "PII_ENC_KEY_V1=$NK_PE" "PII_INDEX_KEY=$NK_PI"; do
			k="${kv%%=*}"; v="${kv#*=}"
			if grep -qE "^$k=" "$APP/.env"; then sed -i "s|^$k=.*|$k=$v|" "$APP/.env"; else echo "$kv" >>"$APP/.env"; fi
		done
		echo "  · 앞 설정 백업: $BAK (옛 키) — 로그인은 한 번 다시 해야 합니다"
		KEYS_CHANGED=1
	fi
fi
# 개발 컴퓨터에서 올린 파일 → 사이트 폴더(서버 쪽이 더 새 파일이면 그대로, 지우지 않음, secure 는 서버에 없는 것만)
if [ -n "$SYNC_FROM" ] && [ -d "$SYNC_FROM" ]; then
	n=0
	for sd in upload file-explorer secure; do
		[ -d "$SYNC_FROM/$sd" ] || continue
		mkdir -p "$APP/data/$sd"
		if [ "$sd" = secure ]; then c="$(cp -a -n -v "$SYNC_FROM/$sd/." "$APP/data/$sd/" | wc -l)"; else c="$(cp -a -u -v "$SYNC_FROM/$sd/." "$APP/data/$sd/" | wc -l)"; fi
		n=$((n + c))
	done
	echo "  · 개발에서 올린 파일 반영: ${n}개(새 것·바뀐 것만)"
fi
mkdir -p "$APP/data/upload/ZBD/site"
for f in "$REL"/db/assets/site/*; do [ -f "$APP/data/upload/ZBD/site/$(basename "$f")" ] || cp "$f" "$APP/data/upload/ZBD/site/"; done

# 첫 관리자 — 관리자가 하나도 없을 때만
if [ "$(app_sql -N -e "SELECT COUNT(*) FROM admins")" = 0 ]; then
	PW="$(rnd 16)"
	(cd "$APP" && ADMIN_PASSWORD="$PW" "$REL/bin/create-admin" -email "admin@$DOMAIN" -name "관리자" -image "$REL/db/assets/profile/400x400_profile_admin.webp" >/dev/null)
	umask 077
	printf '# 첫 관리자 — 로그인 뒤 비밀번호를 바꾸고 이 파일을 지우세요\n[관리자] https://%s/admin\nemail=admin@%s\npassword=%s\n' "$ADMIN1" "$DOMAIN" "$PW" >"$APP/account.txt"
	umask 022
	FIRST_ADMIN=1
fi
chown -R "$NAME":"$NAME" "$APP/data"

# ─── 5. 서비스(systemd) ─────────────────────────────────
say "서비스 등록"
cat >"/etc/systemd/system/$NAME.service" <<UNIT
[Unit]
Description=$NAME 서버 프로그램
After=network-online.target mariadb.service

[Service]
User=$NAME
WorkingDirectory=$APP
ExecStart=$APP/current/bin/zbd-be
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
UNIT
if [ "$HAS_SOCK" = 1 ]; then
	umask 077
	cat >"$APP/socket/.env" <<SENV
PORT=$SOCK_PORT
ADMIN_SOCKET_SECRET=$(envget SOCKET_SECRET)
AES_256_ENCRYPTION_KEY=$(envget AES_256_ENCRYPTION_KEY)
AES_256_ENCRYPTION_IV=$(envget AES_256_ENCRYPTION_IV)
CLAUDE_BE_URL=http://127.0.0.1:$BE_PORT
ORIGINS=$(printf '%s' "$USER_HOSTS $ADMINS" | tr ' ' ','),*.$DOMAIN
READY_SOCKET_SECRET=$(envget READY_SOCKET_SECRET)
READY_QUEUE_HOSTING_IDS=ZBD
SENV
	umask 022
	chown root:"$NAME" "$APP/socket/.env"; chmod 640 "$APP/socket/.env"
	cat >"/etc/systemd/system/$NAME-socket.service" <<UNIT
[Unit]
Description=$NAME 소켓 서버(웹소켓)
After=network-online.target

[Service]
User=$NAME
WorkingDirectory=$APP/socket
ExecStart=$APP/current/bin/zbd-socket
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
UNIT
fi
# 이 사이트 서비스만 다시 시작할 수 있게(시스템 모니터 「재시작」) — 다른 명령·다른 서비스는 안 된다
SC="$(command -v systemctl)"
{
	echo "# zbd 스타터 $NAME — 자기 서비스 재시작만(install.sh)"
	echo "$NAME ALL=(root) NOPASSWD: $SC restart $NAME$([ "$HAS_SOCK" = 1 ] && echo ", $SC restart $NAME-socket")"
	if [ "$HAS_ATTACK" = 1 ]; then echo "$NAME ALL=(root) NOPASSWD: $DENY_SCRIPT"; fi
} >"/etc/sudoers.d/zbd-$NAME"
chmod 440 "/etc/sudoers.d/zbd-$NAME"
visudo -cf "/etc/sudoers.d/zbd-$NAME" >/dev/null || { rm -f "/etc/sudoers.d/zbd-$NAME"; echo "  ⚠ sudoers 확인 실패 — 재시작 허락을 뺐습니다"; }
systemctl daemon-reload
systemctl enable "$NAME" >/dev/null 2>&1
systemctl restart "$NAME"
if [ "$HAS_SOCK" = 1 ]; then systemctl enable "$NAME-socket" >/dev/null 2>&1; systemctl restart "$NAME-socket"; fi
# 새 판이 뜨는지 — 시작하자마자 죽거나 응답이 없으면 묻지 않고 앞 판으로(사이트가 멈춘 채 남지 않게). DB 마이그레이션은 되돌리지 않는다
if ! site_up; then
	echo "✗ 새 판(판 $(rel_rev "$REL"))의 서버 프로그램이 뜨지 않습니다:"
	journalctl -u "$NAME" -n 15 --no-pager -o cat 2>/dev/null | sed 's/^/    /' || true
	if [ -n "$PREV_REL" ] && [ -d "$PREV_REL" ] && [ "$PREV_REL" != "$REL" ]; then
		rel_switch "$PREV_REL"
		mv "$REL" "$REL.failed"
		site_restart
		if site_up; then
			echo "↩ 앞 판(판 $(rel_rev "$PREV_REL"))으로 자동으로 되돌렸습니다 — 사이트는 앞 판으로 돌고 있습니다"
			echo "  DB 마이그레이션은 되돌리지 않았습니다 · 원인을 고친 뒤 다시 zbd deploy"
		else
			echo "✗ 앞 판도 응답하지 않습니다 — journalctl -u $NAME -n 50"
		fi
	fi
	exit 1
fi
# 새 판이 떴다 — 정상 판은 최근 3개, 실패·되돌린 판은 최근 1개만 남긴다
{ ls -1dt "$APP"/releases/*/ 2>/dev/null | grep -vE '\.(failed|rolledback)/$' | tail -n +4 | xargs -r rm -rf; } || true
{ ls -1dt "$APP"/releases/*.failed/ "$APP"/releases/*.rolledback/ 2>/dev/null | tail -n +2 | xargs -r rm -rf; } || true

# ─── 6. 웹서버(nginx) · HTTPS ───────────────────────────
say "nginx 연결(사용자 $USER_HOSTS · 관리자 $ADMINS)"
USER_RE="$(printf '%s' "$USER_HOSTS" | sed 's/\./\\./g; s/ /|/g')" # nginx 정규식(사용자 주소들)
ACME=/var/www/letsencrypt # 인증서 확인 파일 자리 — 판(releases)과 무관하게 늘 같은 곳(갱신이 배포 뒤에도 된다)
mkdir -p "$ACME"
CERT="/etc/letsencrypt/live/$NAME" # 인증서 이름 = 사이트 이름(두 주소를 한 장에)
WS=""
[ "$HAS_SOCK" = 1 ] && WS="
  location = /ws {
    proxy_pass http://127.0.0.1:$SOCK_PORT;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection \"upgrade\";
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_read_timeout 3600s;
  }"
# 올리기 한도 — 서버 프로그램이 주소마다 따로 막는다(일반 파일 200MB · 동영상 플러그인 원본 1GB). 큰 파일은 디스크에 모으지 않고 바로 넘긴다
APP_LOC="  access_log /var/log/nginx/$NAME.access.log; # 이 사이트 접속 기록(외부 공격 감지가 읽는다 · 회전은 nginx 기본 logrotate)
  client_max_body_size 1100m;
  root $APP/current/web;
  # 사용자 주소의 /admin → 대표 관리자 주소로(관리자 화면은 관리자 주소에서만 열린다 — 화면은 사용자 주소에서 홈으로 돌려보낸다)
  if (\$host ~* ^($USER_RE)\$) { rewrite ^/admin(/.*)?\$ \$scheme://$ADMIN1/admin\$1 redirect; }

  location /assets/ { alias $APP/current/web/assets/; expires 7d; }
  location /upload/ { alias $APP/data/upload/; }
  location /service/ {
    proxy_http_version 1.1;
    proxy_request_buffering off;
    proxy_read_timeout 600s;
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_pass http://127.0.0.1:$BE_PORT/service/;
  }
  location /api/ {
    proxy_http_version 1.1;
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_pass http://127.0.0.1:$BE_PORT/api/;
  }$WS
  # 그 밖의 주소 = 서버 프로그램이 페이지 문서(제목·설명·검색엔진용 본문)를 끼워 넣은 index.html
  # 화면 폴더에 실제 있는 파일(images/·서비스 워커·라이브러리)은 그대로, 나머지는 페이지 문서로
  location / { try_files \$uri @page; }
  location @page {
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Proto \$scheme;
    rewrite ^ /service/seo?p=\$request_uri? break;
    proxy_pass http://127.0.0.1:$BE_PORT;
  }"
ACME_LOC="  location ^~ /.well-known/acme-challenge/ { root $ACME; default_type text/plain; }"
# 옛 nginx(1.25 미만 — 우분투 24.04 는 1.24)는 「http2 on」 을 모른다 → 「listen 443 ssl http2」
if nginx -v 2>&1 | grep -qE 'nginx/1\.([0-9]|1[0-9]|2[0-4])\.'; then LISTEN443="listen 443 ssl http2;"; H2ON=""; else LISTEN443="listen 443 ssl;"; H2ON="  http2 on;"; fi
# 기본 사이트 파일이 패키지 원본 그대로인가(dpkg 가 기억하는 지문과 비교)
stock_default() {
	local want have
	want="$(dpkg-query -W -f='${Conffiles}\n' nginx-common nginx 2>/dev/null | awk '$1=="/etc/nginx/sites-available/default"{print $2; exit}')"
	have="$(md5sum /etc/nginx/sites-available/default 2>/dev/null | cut -d' ' -f1)"
	[ -n "$want" ] && [ "$want" = "$have" ]
}
write_nginx() {
	if [ -f "$CERT/fullchain.pem" ]; then
		cat >"/etc/nginx/sites-available/$NAME.conf" <<NGX
# install.sh 가 만든다 — 손으로 고친 것은 다음 배포 때 덮인다. 인증서 $CERT(certbot 자동 갱신)
server {
  listen 80;
  server_name $USER_HOSTS $ADMINS;
$ACME_LOC
  location / { return 301 https://\$host\$request_uri; }
}
server {
  $LISTEN443
$H2ON
  server_name $USER_HOSTS $ADMINS;
  ssl_certificate $CERT/fullchain.pem;
  ssl_certificate_key $CERT/privkey.pem;
$APP_LOC
}
NGX
	else
		cat >"/etc/nginx/sites-available/$NAME.conf" <<NGX
# install.sh 가 만든다 — 손으로 고친 것은 다음 배포 때 덮인다. HTTPS 는 deploy.sh --ssl 이메일
server {
  listen 80;
  server_name $USER_HOSTS $ADMINS;
$ACME_LOC
$APP_LOC
}
NGX
	fi
	ln -sfn "/etc/nginx/sites-available/$NAME.conf" "/etc/nginx/sites-enabled/$NAME.conf"
	# 우분투 기본 사이트(「Welcome to nginx」) — 원본 그대로이고 다른 사이트가 없는 새 서버일 때만 끈다.
	#   이미 다른 사이트가 있는 서버(운영 중)는 기본 사이트가 「모르는 주소」 를 받는 자리일 수 있어 건드리지 않는다(우리 주소는 이름으로 잡힌다)
	local others
	others="$(find /etc/nginx/sites-enabled -mindepth 1 -maxdepth 1 ! -name default ! -name "$NAME.conf" | wc -l)"
	if [ -L /etc/nginx/sites-enabled/default ] && [ "$others" = 0 ] && stock_default; then rm -f /etc/nginx/sites-enabled/default; fi
	if ! nginx -t -q; then
		rm -f "/etc/nginx/sites-enabled/$NAME.conf" # 깨진 설정을 남기면 이 서버 nginx 의 다음 reload 가 모두 막힌다
		echo "  ✗ nginx 설정 검사 실패 — 이 사이트 설정은 빼 두었습니다(sudo nginx -t 로 원인 확인)"; exit 1
	fi
	systemctl reload nginx
	sleep 2 # reload 는 비동기 — 바로 재면 옛 설정이 답한다
}
write_nginx

# HTTPS — --ssl 이메일 이면 모든 주소 인증서를 한 장으로 받고(Let's Encrypt, 80 포트로 확인) nginx 를 443 으로 다시 쓴다.
#   모든 주소의 DNS 가 이 서버를 가리켜야 한다. 갱신은 certbot 타이머, 갱신 뒤 nginx 다시 읽기는 deploy 훅.
#   인증서가 있어도 새 주소(예: 플러그인이 더한 chat.<도메인>)가 빠져 있으면 그 주소까지 넓혀 다시 받는다(--expand)
cert_covers() {
	local san h
	san="$(openssl x509 -in "$CERT/fullchain.pem" -noout -ext subjectAltName 2>/dev/null)" || return 1
	for h in $USER_HOSTS $ADMINS; do printf '%s' "$san" | grep -qE "DNS:$h(,|\$| )" || return 1; done
}
# certbot 실패 메시지를 읽어 주소마다 원인 하나를 알려 준다(Domain:/Detail: 묶음). 못 알아보면 흔한 원인 목록
cert_hint() {
	local log="$1" myip found=0 dom det ip
	myip="$(curl -s -m 5 https://checkip.amazonaws.com 2>/dev/null | tr -d '[:space:]' || true)"
	echo "  ✗ 인증서를 받지 못했습니다 — 원인:"
	if grep -qiE 'invalidEmail|invalid contact|unable to register|contact email' "$log"; then
		echo "    · 이메일 주소($SSL_EMAIL)를 Let's Encrypt 가 받지 않습니다 — 실제로 받을 수 있는 주소로: --ssl 내메일@주소"; found=1
	fi
	if grep -qiE 'rateLimited|too many (certificates|failed)' "$log"; then
		echo "    · Let's Encrypt 발급 한도에 걸렸습니다 — 한 시간쯤 뒤 다시 배포하세요(같은 실패를 여러 번 되풀이하면 걸립니다)"; found=1
	fi
	while IFS=$'\t' read -r dom det; do
		[ -n "$dom" ] || continue; found=1
		ip="$(printf '%s' "$det" | grep -oE '^[0-9a-fA-F.:]+:' | sed 's/:$//' || true)"
		case "$det" in
			*NXDOMAIN*|*"DNS problem"*|*"no valid A records"*)
				echo "    · $dom — DNS 기록이 없습니다. 도메인 관리 화면에서 A 기록을 ${myip:-이 서버 IP} 로 만드세요" ;;
			*"Timeout during connect"*|*"Connection refused"*|*"Connection reset"*)
				echo "    · $dom — 밖에서 이 서버 80 포트에 닿지 않습니다(${ip:-?}). 방화벽·클라우드 보안 규칙·공유기 포트포워딩에서 80 을 여세요" ;;
			*)
				if [ -n "$ip" ] && [ -n "$myip" ] && [ "$ip" != "$myip" ]; then
					case "$ip" in
						104.1[6-9].*|104.2[0-9].*|104.3[01].*|172.6[4-9].*|172.7[01].*|188.114.*|162.158.*|141.101.*|108.162.*|190.93.*|197.234.*|198.41.*|173.245.*|103.2[12].*|103.31.*)
							echo "    · $dom — Cloudflare 프록시(주황 구름)를 거칩니다($ip). DNS 화면에서 회색 구름(DNS only)으로 바꾸세요" ;;
						*) echo "    · $dom — DNS 가 다른 서버($ip)를 가리킵니다. A 기록을 이 서버 $myip 로 바꾸세요(바꾼 뒤 몇 분 기다림)" ;;
					esac
				else
					echo "    · $dom — $(printf '%s' "$det" | cut -c1-160)"
				fi ;;
		esac
	done < <(awk '/^ *Domain: /{d=$2} /^ *Detail: /{sub(/^ *Detail: */,""); if(d!=""){print d "\t" $0; d=""}}' "$log")
	[ "$found" = 1 ] || echo "    · 위 certbot 메시지를 보세요(흔한 원인: 이메일 주소가 틀림 · 주소의 DNS 가 이 서버(${myip:-IP})를 가리키지 않음 · 80 포트가 막힘)"
	return 0
}
if [ -n "$SSL_EMAIL" ] && { [ ! -f "$CERT/fullchain.pem" ] || ! cert_covers; }; then
	say "HTTPS 인증서 받기(Let's Encrypt — $USER_HOSTS $ADMINS)"
	command -v certbot >/dev/null || { export DEBIAN_FRONTEND=noninteractive; apt-get update -qq; apt-get install -y -qq certbot >/dev/null 2>&1; }
	DARGS=(); for h in $USER_HOSTS $ADMINS; do DARGS+=(-d "$h"); done
	CBLOG="$(mktemp)"
	if certbot certonly --non-interactive --agree-tos -m "$SSL_EMAIL" --webroot -w "$ACME" --cert-name "$NAME" "${DARGS[@]}" --keep-until-expiring --expand 2>&1 | tee "$CBLOG"; [ "${PIPESTATUS[0]}" = 0 ]; then
		mkdir -p /etc/letsencrypt/renewal-hooks/deploy
		printf '#!/bin/sh\n# 인증서가 갱신되면 nginx 가 새 인증서를 읽게(install.sh)\nsystemctl is-active --quiet nginx && systemctl reload nginx\nexit 0\n' >/etc/letsencrypt/renewal-hooks/deploy/zbd-reload.sh
		chmod 755 /etc/letsencrypt/renewal-hooks/deploy/zbd-reload.sh
		systemctl enable --now certbot.timer >/dev/null 2>&1 || true
		write_nginx
	else
		cert_hint "$CBLOG"
		echo "  → 고친 뒤 다시 배포하세요(사이트는 그동안 http 로 돕니다)."
	fi
	rm -f -- "$CBLOG"
fi
SCHEME=http; [ -f "$CERT/fullchain.pem" ] && SCHEME=https

# ─── 7. 확인 ────────────────────────────────────────────
say "확인"
ok=1
for i in $(seq 1 20); do curl -s -o /dev/null "http://127.0.0.1:$BE_PORT/" && break; sleep 1; done
# 200 만 보지 않고 내용도 본다 — 화면 문서엔 자리표 흔적(zbd), base 엔 "data". HTTPS 면 이 서버의 443 으로(인증서 이름 확인은 건너뜀 — 아래 줄에서 따로)
CHECKS=("$DOMAIN/|_app/"); for h in $ADMINS; do CHECKS+=("$h/admin|_app/"); done; CHECKS+=("$DOMAIN/service/design/ZBD/base|\"data\"")
for u in "${CHECKS[@]}"; do
	want="${u#*|}"; u="${u%%|*}"; h="${u%%/*}"; p="/${u#*/}"
	if [ "$SCHEME" = https ]; then
		body="$(curl -sk -w '\n%{http_code}' --resolve "$h:443:127.0.0.1" "https://$h$p")"
	else
		body="$(curl -s -w '\n%{http_code}' -H "Host: $h" "http://127.0.0.1$p")"
	fi
	code="${body##*$'\n'}"
	mark=""; printf '%s' "$body" | grep -q "$want" || { mark=" (내용이 즈비드 화면이 아님)"; code="x$code"; }
	printf '  %s %s%s\n' "$code" "$SCHEME://$u" "$mark"
	[ "$code" = 200 ] || ok=0
done
if [ "$SCHEME" = https ]; then
	code="$(curl -s -o /dev/null -w '%{http_code}' -H "Host: $DOMAIN" "http://127.0.0.1/")"
	printf '  %s http → https 넘기기\n' "$code"; [ "$code" = 301 ] || ok=0
fi
if [ "$HAS_SOCK" = 1 ]; then
	code="$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$SOCK_PORT/healthz")"
	printf '  %s 소켓 서버 healthz\n' "$code"; [ "$code" = 200 ] || ok=0
fi
echo
if [ "$ok" = 1 ]; then echo "✓ 배포 끝 — 사이트 $SCHEME://$DOMAIN · 관리자 $SCHEME://$ADMIN1/admin"; else echo "✗ 확인 실패 — journalctl -u $NAME -n 50 · nginx -t"; fi
[ "${FIRST_ADMIN:-}" = 1 ] && echo "  첫 관리자 계정: sudo cat $APP/account.txt (로그인 뒤 비밀번호를 바꾸고 지우세요)"
[ "${FIRST_ADMIN:-}" = 1 ] && echo "  ※ 개발 DB 를 옮기지 않고 빈 스타터 DB 로 시작했습니다 — 개발 DB 로 바꾸려면 zbd deploy --with-db(docs/GETTING_STARTED.md 7절)"
[ "$DEVDB_LOADED" = 1 ] && echo "  개발 DB 를 그대로 옮겼습니다 — 관리자 로그인은 내 컴퓨터(run_dev)에서 쓰던 계정·비밀번호 그대로입니다"
[ "$SCHEME" = http ] && echo "  HTTPS: 모든 주소의 DNS 를 이 서버로 걸고 ./deploy/deploy.sh --ssl 이메일 로 다시 배포(인증서·443·자동 갱신)"
[ "$ok" = 1 ]
