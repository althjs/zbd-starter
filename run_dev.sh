#!/usr/bin/env bash
# 개발 모드 — 필요한 도구 확인(없으면 설치) → 설정·DB 준비 → BE(Go)·FE(Svelte) 를 함께 띄운다.
#   ./run_dev.sh        로그는 [be] · [fe] 머리를 붙여 한 화면에. Ctrl+C 로 둘 다 끈다.
#
# 처음 한 번 만드는 것(모두 git 에 올리지 않는다):
#   zbd-be-golang/.env.local   DB 접속·암호화 키(무작위)       ← 잃으면 회원 개인정보를 못 읽는다
#   .dev/account.txt           첫 관리자·시험 회원 이메일·비밀번호(600)
#   .dev/index.html · .dev/upload/   BE 가 쓰는 문서 틀 · 업로드 폴더
#
# 포트: BE 5001 · FE 5000 · 소켓 데몬 4006(웹소켓 플러그인을 붙였을 때만)
#   바꾸기 = zbd-be-golang/.env.local 에 PORT=(BE) · FE_PORT= · SOCK_PORT= 를 적는다(늘 그 포트로)
#   한 번만 = BE_PORT=… FE_PORT=… SOCK_PORT=… ./run_dev.sh (명령 앞 값이 .env.local 보다 먼저)
# 외부 접속 주소(zbd tunnel): zbd 로그인 + 구매가 유효하면 자동으로 같이 뜬다 — 끄기 TUNNEL=0
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
BE="$ROOT/zbd-be-golang"
FE="$ROOT/zbd-fe-svelte5"
DEV="$ROOT/.dev"
NODE_MIN=22
OS="$(uname)"

say() { printf '\033[1;36m▶ %s\033[0m\n' "$*"; }
die() { printf '\033[1;31m✗ %s\033[0m\n' "$*"; exit 1; }

# ─── 포트 — 명령 앞 값 → .env.local → 기본값. .env.local 의 BE 포트는 BE 가 읽는 PORT(BE_PORT 라고 적어도 된다) ───
envlocal() { [ -f "$BE/.env.local" ] && grep -E "^$1=" "$BE/.env.local" | tail -1 | cut -d= -f2- | tr -d '"'"'"' \r' || true; }
BE_PORT="${BE_PORT:-$(envlocal BE_PORT)}"
BE_PORT="${BE_PORT:-$(envlocal PORT)}"
BE_PORT="${BE_PORT:-5001}"
FE_PORT="${FE_PORT:-$(envlocal FE_PORT)}"
FE_PORT="${FE_PORT:-5000}"
SOCK_PORT_ARG="${SOCK_PORT:-}" # 직접 준 값이 있으면 그것, 없으면 .env.local SOCK_PORT(지난번 고른 포트) → 4006
SOCK_PORT="${SOCK_PORT:-4006}"
for _p in "$BE_PORT" "$FE_PORT" "$SOCK_PORT"; do
	case "$_p" in ''|*[!0-9]*) die "포트 값이 숫자가 아닙니다: '$_p' — zbd-be-golang/.env.local 의 PORT·FE_PORT·SOCK_PORT 를 확인하세요" ;; esac
done
[ "$BE_PORT" != "$FE_PORT" ] || die "BE 포트와 FE 포트가 같습니다($BE_PORT) — 다르게 정해 주세요"

# ─── 1. 도구 ───────────────────────────────────────────────
need_brew() {
	command -v brew >/dev/null || die "Homebrew 가 없습니다. https://brew.sh 의 설치 명령을 터미널에서 실행한 뒤 다시 실행하세요(관리자 비밀번호를 물어서 대신 설치할 수 없습니다)."
}
install_pkg() { # $1 = 명령 이름, $2 = brew 패키지, $3 = apt 패키지
	command -v "$1" >/dev/null && return 0
	say "$1 설치"
	if [ "$OS" = "Darwin" ]; then need_brew; brew install "$2"
	elif command -v apt-get >/dev/null; then
		# 이 패키지와 상관없는 깨진 패키지(예: 커널 모듈)가 있으면 apt 가 실패로 끝난다 → 결과는 명령이 생겼는지로 본다
		sudo apt-get update -qq || true
		sudo apt-get install -y "$3" || true
		command -v "$1" >/dev/null || die "$1 설치에 실패했습니다 — 위 apt 메시지를 확인하세요"
	else die "$1 을(를) 설치해 주세요"; fi
}

# Go — go.mod 가 요구하는 판. 리눅스 배포판 패키지(golang-go)는 오래됐고 자동 판 받기(GOTOOLCHAIN)도 꺼 두어
# 빌드가 안 된다 → 없거나 낮으면 공식 판을 /usr/local/go 에 받는다(맥은 brew).
GO_NEED="$(sed -n 's/^go \([0-9.]*\).*/\1/p' "$BE/go.mod")"
[ -x /usr/local/go/bin/go ] && export PATH="/usr/local/go/bin:$PATH"
[ -x /usr/local/node/bin/node ] && export PATH="/usr/local/node/bin:$PATH"
go_ok() { command -v go >/dev/null && [ "$(printf '%s\n%s\n' "$GO_NEED" "$(go env GOVERSION | sed 's/^go//')" | sort -V | head -1)" = "$GO_NEED" ]; }
install_go() {
	go_ok && return 0
	if [ "$OS" = "Darwin" ]; then install_pkg go go golang-go; brew upgrade go >/dev/null 2>&1 || true; go_ok || die "Go $GO_NEED 이상이 필요합니다(brew upgrade go)"; return 0; fi
	local arch; case "$(uname -m)" in x86_64) arch=amd64 ;; aarch64 | arm64) arch=arm64 ;; *) die "Go $GO_NEED 이상을 설치해 주세요(https://go.dev/dl)" ;; esac
	say "Go $GO_NEED 설치(/usr/local/go — 공식 판)"
	local tgz="/tmp/go$GO_NEED.linux-$arch.tar.gz"
	curl -fsSL -o "$tgz" "https://go.dev/dl/go$GO_NEED.linux-$arch.tar.gz" || die "Go 내려받기 실패(https://go.dev/dl)"
	sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "$tgz" && rm -f "$tgz"
	export PATH="/usr/local/go/bin:$PATH"
	go_ok || die "Go $GO_NEED 설치 뒤에도 판이 맞지 않습니다"
	grep -qs '/usr/local/go/bin' "$HOME/.profile" || echo 'export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"' >>"$HOME/.profile"
	echo "  (새 터미널부터 go 가 잡힙니다 — ~/.profile 에 경로를 더했습니다)"
}

say "도구 확인"
install_go
install_pkg mariadb mariadb mariadb-server
if ! command -v node >/dev/null || [ "$(node -p 'process.versions.node.split(".")[0]')" -lt "$NODE_MIN" ]; then
	say "Node.js $NODE_MIN 이상 설치"
	if [ "$OS" = "Darwin" ]; then need_brew; brew install node
	else
		# 리눅스 — 공식 판(/usr/local/node). 배포판 nodejs 패키지는 오래됐다
		case "$(uname -m)" in x86_64) narch=x64 ;; aarch64 | arm64) narch=arm64 ;; *) die "Node.js $NODE_MIN 이상을 설치해 주세요(https://nodejs.org)" ;; esac
		nbase="https://nodejs.org/dist/latest-v$NODE_MIN.x"
		nfile="$(curl -fsSL "$nbase/SHASUMS256.txt" | grep -o "node-v[0-9.]*-linux-$narch.tar.xz" | head -1)"
		[ -n "$nfile" ] || die "Node.js 내려받기 실패(https://nodejs.org)"
		curl -fsSL -o "/tmp/$nfile" "$nbase/$nfile" || die "Node.js 내려받기 실패"
		sudo rm -rf /usr/local/node && sudo mkdir -p /usr/local/node && sudo tar -C /usr/local/node --strip-components=1 -xJf "/tmp/$nfile" && rm -f "/tmp/$nfile"
		export PATH="/usr/local/node/bin:$PATH"
		grep -qs '/usr/local/node/bin' "$HOME/.profile" || echo 'export PATH="/usr/local/node/bin:$PATH"' >>"$HOME/.profile"
	fi
fi
AIR="$(command -v air || echo "$(go env GOPATH)/bin/air")"
if [ ! -x "$AIR" ]; then say "air 설치(Go 저장 시 다시 시작)"; go install github.com/air-verse/air@latest; fi
echo "  go $(go env GOVERSION) · node $(node -v) · $(mariadb --version | sed 's/.*Distrib //;s/,.*//;s/^mariadb from //')"

# ─── 2. MariaDB 켜기 ──────────────────────────────────────
db_up() { mariadb-admin ping >/dev/null 2>&1 || (sudo -n true 2>/dev/null && sudo mariadb-admin ping >/dev/null 2>&1); }
if ! db_up; then
	say "MariaDB 시작"
	if [ "$OS" = "Darwin" ]; then brew services start mariadb >/dev/null; else sudo systemctl start mariadb; fi
	for _ in $(seq 1 30); do db_up && break; sleep 1; done
	db_up || die "MariaDB 가 뜨지 않습니다 — brew services list / journalctl -u mariadb 를 확인하세요"
fi

# ─── 3. 설정 (처음 한 번) ─────────────────────────────────
mkdir -p "$DEV/upload"
chmod 700 "$DEV"
# 무작위 문자열 — 넉넉히 받아 잘라 쓴다(`tr </dev/urandom | head` 는 pipefail 에 걸려 스크립트가 조용히 끝난다)
rnd() { local s; s="$(head -c 96 /dev/urandom | base64 | LC_ALL=C tr -dc 'A-Za-z0-9')"; printf '%s' "${s:0:$1}"; }
b64() { head -c 32 /dev/urandom | base64 | tr -d '\n'; }
if [ ! -f "$BE/.env.local" ]; then
	say "BE 설정 만들기(.env.local — 키는 무작위)"
	umask 077
	cat >"$BE/.env.local" <<EOF
# run_dev.sh 가 만든 개발 설정 — git 에 올리지 않는다. 키를 잃으면 이 DB 의 회원 개인정보를 못 읽는다.
# 포트 — 고치면 run_dev.sh 를 껐다 켠다(PORT = 서버, FE_PORT = 화면, SOCK_PORT = 소켓 데몬)
PORT=$BE_PORT
FE_PORT=$FE_PORT
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=zbd_starter
DB_USER=zbd_dev
DB_PASSWORD=$(rnd 24)
DB_CHARSET=utf8mb4
DB_CONN_LIMIT=10
AES_256_ENCRYPTION_KEY=$(rnd 32)
AES_256_ENCRYPTION_IV=$(rnd 16)
PII_ENC_KEY_V1=$(b64)
PII_INDEX_KEY=$(b64)
AUTH_COOKIE_KEY=zbd_admin_session
AUTH_COOKIE_KEY_USER=zbd_user_session
CDN_UPLOAD_DIR=$DEV/upload/
FE_INDEX_HTML=$DEV/index.html
RUN_JOBS=false
EOF
	umask 022
fi
if [ ! -f "$DEV/index.html" ]; then
	# BE 가 페이지 문서를 만들 때 쓰는 틀 — 개발 서버는 여기서 머리말·사이트 데이터만 가져다 쓴다
	cat >"$DEV/index.html" <<'EOF'
<!doctype html>
<html lang="ko">
<head><meta charset="utf-8" /></head>
<body><div id="root"></div></body>
</html>
EOF
fi

# ─── 4. DB ───────────────────────────────────────────────
say "DB 준비"
"$ROOT/db/setup.sh"
# 첫 관리자 · 시험 회원(기본 프로필 사진 db/assets/profile) — 처음 한 번(.dev/account.txt 가 없을 때). 비밀번호는 무작위, 이 파일에만(600)
if [ ! -f "$DEV/account.txt" ]; then
	ADMIN_EMAIL="admin@example.com"
	ADMIN_PASSWORD="$(rnd 16)"
	(cd "$BE" && ADMIN_PASSWORD="$ADMIN_PASSWORD" go run ./cmd/create-admin -email "$ADMIN_EMAIL" -name "관리자" -image "$ROOT/db/assets/profile/400x400_profile_admin.webp")
	# 시험 회원 — 사용자 화면(로그인·마이페이지·글쓰기) 확인용, 자유게시판 예시 글도 이 회원 이름으로. 운영 전에 「사용자 회원」 에서 지운다
	MEMBER_EMAIL="user@example.com"
	MEMBER_PASSWORD="$(rnd 16)"
	(cd "$BE" && MEMBER_PASSWORD="$MEMBER_PASSWORD" go run ./cmd/create-member -email "$MEMBER_EMAIL" -name "시험회원" -nickname "테스터" -image "$ROOT/db/assets/profile/400x400_profile_user.webp")
	umask 077
	cat >"$DEV/account.txt" <<ACCOUNT
# 개발용 계정 — git 에 올라가지 않는다
[관리자] http://localhost:$FE_PORT/admin
email=$ADMIN_EMAIL
password=$ADMIN_PASSWORD

[사용자] http://localhost:$FE_PORT/login
email=$MEMBER_EMAIL
password=$MEMBER_PASSWORD
ACCOUNT
	umask 022
	unset ADMIN_PASSWORD MEMBER_PASSWORD
fi

# 사이트 기본 그림(로고·파비콘·OG) — 업로드 폴더 ZBD/site 로(없는 파일만). 사이트 설정(SITE)의 기본값이 이 이름을 가리킨다
mkdir -p "$DEV/upload/ZBD/site"
for f in "$ROOT"/db/assets/site/*; do
	[ -f "$DEV/upload/ZBD/site/$(basename "$f")" ] || cp "$f" "$DEV/upload/ZBD/site/"
done

# 공지 예시 첨부파일 — 보안 폴더(웹으로 직접 열리지 않는다, 받기는 API). 공지 notice01 의 「시작 가이드.md」
NOTICE_FILES="$DEV/secure/ZBD/post_notice01"
if [ ! -f "$NOTICE_FILES/시작 가이드.md" ]; then
	mkdir -p "$NOTICE_FILES" && cp "$ROOT/docs/GETTING_STARTED.md" "$NOTICE_FILES/시작 가이드.md"
fi

# 웹하드 기본 문서 — 처음 한 번(폴더가 없을 때만) 시작 가이드(docs/GETTING_STARTED.md)를 넣는다. 웹하드 = {업로드 폴더}/../file-explorer/{호스팅}/WEB_HDD
WEBHARD="$DEV/file-explorer/ZBD/WEB_HDD"
if [ ! -d "$WEBHARD" ]; then
	mkdir -p "$WEBHARD" && cp "$ROOT/docs/GETTING_STARTED.md" "$WEBHARD/시작 가이드.md"
	say "웹하드에 시작 가이드 넣음"
fi

# ─── 5. FE 패키지 ────────────────────────────────────────
if [ ! -d "$FE/node_modules" ] || [ "$FE/package-lock.json" -nt "$FE/node_modules/.package-lock.json" ]; then
	say "FE 패키지 설치(npm ci)"
	(cd "$FE" && npm ci --no-audit --no-fund)
fi

# ─── 5½. 소켓 데몬(웹소켓 플러그인을 붙였을 때만 — plugins/zbd-socket) ─────
SOCK="$ROOT/plugins/zbd-socket"
if [ -f "$SOCK/main.go" ]; then
	envget() { grep -E "^$1=" "$BE/.env.local" | tail -1 | cut -d= -f2-; }
	if [ -z "$SOCK_PORT_ARG" ]; then
		SAVED_SOCK_PORT="$(envget SOCK_PORT || true)"
		[ -n "$SAVED_SOCK_PORT" ] && SOCK_PORT="$SAVED_SOCK_PORT"
	fi
	# BE → 소켓 데몬 주소. 이 컴퓨터 주소(127.0.0.1)면 SOCK_PORT 에 맞춘다(포트를 바꿔 다시 띄워도 따라오게)
	SOCK_URL="http://127.0.0.1:$SOCK_PORT"
	if [ -z "$(envget SOCKET_URL)" ]; then echo "SOCKET_URL=$SOCK_URL" >>"$BE/.env.local"
	elif envget SOCKET_URL | grep -qE '^http://127\.0\.0\.1:[0-9]+$' && [ "$(envget SOCKET_URL)" != "$SOCK_URL" ]; then
		sed -i.bak "s#^SOCKET_URL=.*#SOCKET_URL=$SOCK_URL#" "$BE/.env.local" && rm -f "$BE/.env.local.bak"
	fi
	[ -n "$(envget SOCKET_SECRET)" ] || echo "SOCKET_SECRET=$(rnd 32)" >>"$BE/.env.local"
	SD="$DEV/socket"; mkdir -p "$SD"
	say "소켓 데몬 빌드(plugins/zbd-socket)"
	(cd "$SOCK" && go build -buildvcs=false -o "$SD/zbd-socket" .)
	umask 077
	cat >"$SD/.env" <<SOCKENV
ADMIN_SOCKET_SECRET=$(envget SOCKET_SECRET)
AES_256_ENCRYPTION_KEY=$(envget AES_256_ENCRYPTION_KEY)
AES_256_ENCRYPTION_IV=$(envget AES_256_ENCRYPTION_IV)
CLAUDE_BE_URL=http://127.0.0.1:$BE_PORT
ORIGINS=*:$FE_PORT
READY_SOCKET_SECRET=$(envget READY_SOCKET_SECRET)
READY_QUEUE_HOSTING_IDS=ZBD
PORT=$SOCK_PORT
SOCKENV
	umask 022
	SOCK_RUN=1
fi

# ─── 5¾. Claude 통합(붙였을 때만 — plugins/claude-bridge): Claude Code 설치·로그인·중계 등록을 여기서 한 번에 ─────
if [ -f "$ROOT/plugins/claude-bridge/dev-hook.sh" ] && [ "${SOCK_RUN:-}" = 1 ]; then
	. "$ROOT/plugins/claude-bridge/dev-hook.sh"
fi

# killMine 포트 — 그 포트를 쥔 프로세스 중 이 스타터 폴더(작업 폴더 기준)에서 뜬 것만 끈다(TERM → 5초 → KILL). 남의 프로그램은 건드리지 않는다
killMine() {
	command -v lsof >/dev/null || return 0
	local pid cwd
	for pid in $(lsof -nP -t -iTCP:"$1" -sTCP:LISTEN 2>/dev/null); do
		cwd="$(lsof -a -p "$pid" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p' | head -1 || true)"
		case "$cwd" in "$ROOT" | "$ROOT"/*)
			[ -n "${2:-}" ] && echo "  포트 $1: $2(pid $pid) — 끕니다"
			kill -TERM "$pid" 2>/dev/null || true
			for _ in 1 2 3 4 5 6 7 8 9 10; do kill -0 "$pid" 2>/dev/null || break; sleep 0.5; done
			kill -KILL "$pid" 2>/dev/null || true
			;;
		esac
	done
}

# ─── 5⅞. git — 스타터 원본으로 push 막기 ─────────────────────
# 스타터 원본(althjs/zbd-starter)에서 받은 저장소면 그쪽 push 를 막는다 — 내 사이트 코드·유료 플러그인이 원본에 올라가지 않게.
# 받기(zbd update)는 그대로. 내 저장소에 올리려면 다른 이름의 원격을 더한다(git remote add mine …). 즈비드 개발 저장소는 git config zbd.starterDev true
if command -v git >/dev/null && git -C "$ROOT" remote get-url origin 2>/dev/null | grep -qE 'althjs/zbd-starter(\.git)?$' \
	&& [ "$(git -C "$ROOT" config zbd.starterDev || true)" != true ] && [ -z "$(git -C "$ROOT" config remote.origin.pushurl || true)" ]; then
	git -C "$ROOT" remote set-url --push origin no-push && echo "  스타터 원본(origin)으로 push 를 막았습니다 — 받기(zbd update)는 그대로"
fi

# ─── 6. 실행 ─────────────────────────────────────────────
# 작업마다 프로세스 그룹을 따로(set -m) — Ctrl+C 는 이 스크립트만 받고, 정리할 때 그룹째(air·go 서버·vite·로그) 끈다
set -m
cleanup() {
	trap - INT TERM EXIT
	echo
	say "끄는 중…"
	local groups; groups="$(jobs -p)"
	# air 는 Go 서버를 자기 그룹 밖에 띄운다 — air 가 서버를 끄고 나갈 시간을 준다(INT → 최대 8초 → 남은 것 강제)
	for g in $groups; do kill -INT -- "-$g" 2>/dev/null || true; done
	for _ in $(seq 1 16); do
		alive=0
		for g in $groups; do kill -0 -- "-$g" 2>/dev/null && alive=1; done
		[ "$alive" = 0 ] && break
		sleep 0.5
	done
	for g in $groups; do kill -KILL -- "-$g" 2>/dev/null || true; done
	# 그래도 남은 것(포트를 쥔 이 폴더의 BE·vite) — 다음 실행이 「포트 사용 중」 으로 막히지 않게.
	# air 는 Go 서버를 그룹 밖에 띄우고, npx 가 띄운 vite 도 Ctrl+C 뒤에 남은 적이 있다(2026-10-06 리눅스)
	killMine "$BE_PORT"
	killMine "$FE_PORT"
	[ "${SOCK_RUN:-}" = 1 ] && killMine "$SOCK_PORT"
	echo "  꺼졌습니다."
}
trap cleanup INT TERM EXIT

port_busy() { (exec 3<>"/dev/tcp/127.0.0.1/$1") 2>/dev/null; } # 그 포트에 무언가 떠 있나

prefix() { # 줄마다 머리 붙이기(버퍼 없이)
	awk -v p="$1" '{ print p " " $0; fflush() }'
}

# 포트가 이미 쓰이면 바로 알린다 — macOS 는 「AirPlay 수신 모드」 가 5000 을 쥐고 있을 수 있다(시스템 설정 › 일반 › AirDrop 및 Handoff 에서 끄거나 FE_PORT 로 바꾼다)
if command -v lsof >/dev/null; then
	for p in "$BE_PORT" "$FE_PORT"; do
		# 지난번 run_dev 가 남긴 것(이 스타터 폴더에서 뜬 vite·Go 서버)이면 끄고 이어 간다
		killMine "$p" "지난번 실행이 남긴 프로그램"
		if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
			echo "✗ 포트 $p 를 이미 다른 프로그램이 쓰고 있습니다: $(lsof -nP -iTCP:"$p" -sTCP:LISTEN | awk 'NR==2{print $1" (pid "$2")"}')"
			[ "$p" = 5000 ] && [ "$(uname)" = "Darwin" ] && echo "  macOS 라면 「AirPlay 수신 모드」 일 수 있어요 — 시스템 설정 › 일반 › AirDrop 및 Handoff 에서 끄거나 FE_PORT=5100 ./run_dev.sh"
			exit 1
		fi
	done
	# 소켓 데몬 포트 — 다른 zbd-socket(다른 스타터 등)이면 그대로 쓰고, 다른 프로그램이면 멈춘다(남의 프로그램에 붙지 않게)
	if [ "${SOCK_RUN:-}" = 1 ]; then
		killMine "$SOCK_PORT" "지난번 실행이 남긴 소켓 데몬"
		# 다른 계정의 프로그램은 lsof 에 안 보일 수 있다(리눅스) — 그래서 실제로 붙어 보는 것으로도 확인한다
		SOCK_CMD="$(lsof -nP -iTCP:"$SOCK_PORT" -sTCP:LISTEN -Fc 2>/dev/null | sed -n 's/^c//p' | head -1 || true)"
		if [ "$SOCK_CMD" != zbd-socket ] && port_busy "$SOCK_PORT"; then
			# 비어 있는 포트를 스스로 골라(+10 씩, 열 번까지) 기억한다 — 다음부터는 그 포트로
			OLD_SOCK_PORT="$SOCK_PORT"
			for _ in 1 2 3 4 5 6 7 8 9 10; do
				SOCK_PORT=$((SOCK_PORT + 10))
				port_busy "$SOCK_PORT" || break
			done
			if port_busy "$SOCK_PORT"; then
				echo "✗ 소켓 데몬에 쓸 빈 포트를 찾지 못했습니다($OLD_SOCK_PORT~$SOCK_PORT) — SOCK_PORT=번호 ./run_dev.sh 로 정해 주세요"
				exit 1
			fi
			echo "  소켓 데몬 포트 $OLD_SOCK_PORT 는 다른 프로그램(${SOCK_CMD:-다른 계정의 프로그램})이 쓰고 있어 $SOCK_PORT 를 씁니다"
		fi
		# 고른 포트를 설정에 맞춘다(서버 → 데몬 주소 · 데몬 .env) 그리고 기억
		if [ "$(envget SOCK_PORT || true)" != "$SOCK_PORT" ]; then
			if grep -q '^SOCK_PORT=' "$BE/.env.local"; then sed -i.bak "s#^SOCK_PORT=.*#SOCK_PORT=$SOCK_PORT#" "$BE/.env.local" && rm -f "$BE/.env.local.bak"
			else echo "SOCK_PORT=$SOCK_PORT" >>"$BE/.env.local"; fi
		fi
		if envget SOCKET_URL | grep -qE '^http://127\.0\.0\.1:[0-9]+$'; then
			sed -i.bak "s#^SOCKET_URL=.*#SOCKET_URL=http://127.0.0.1:$SOCK_PORT#" "$BE/.env.local" && rm -f "$BE/.env.local.bak"
		fi
		sed -i.bak "s#^PORT=.*#PORT=$SOCK_PORT#" "$SD/.env" && rm -f "$SD/.env.bak"
	fi
fi

# 화면은 기본으로 같은 네트워크의 다른 기기(휴대폰·다른 PC)에서도 열린다(vite 가 「Network:」 주소를 보여 준다). 이 컴퓨터에서만 = .env.local 에 FE_HOST=localhost. 서버(BE)·DB 는 그대로 이 컴퓨터 안에서만
if [ -z "${FE_HOST:-}" ]; then FE_HOST="$(grep -E '^FE_HOST=' "$BE/.env.local" 2>/dev/null | tail -1 | cut -d= -f2- || true)"; fi
[ -n "${FE_HOST:-}" ] && export FE_HOST
say "BE http://127.0.0.1:$BE_PORT · FE http://localhost:$FE_PORT (관리자 http://localhost:$FE_PORT/admin — 계정은 .dev/account.txt)"
(cd "$BE" && PORT="$BE_PORT" exec "$AIR" 2>&1) | prefix $'\033[33m[be]\033[0m' &
(cd "$FE" && ZBD_API="http://127.0.0.1:$BE_PORT" ZBD_WS="ws://127.0.0.1:$SOCK_PORT" FE_PORT="$FE_PORT" exec npx vite dev 2>&1) | prefix $'\033[35m[fe]\033[0m' &
# 외부 접속 주소(터널) — zbd 로그인 + 「외부 접속 주소」 구매가 유효하면 https://starter-{회원 번호}.zbd.co.kr 로 같이 연다. 끄기 TUNNEL=0
if [ "${TUNNEL:-}" != 0 ]; then
	ZBD_BIN="$(command -v zbd || true)"
	[ -z "$ZBD_BIN" ] && [ -x "$HOME/.local/bin/zbd" ] && ZBD_BIN="$HOME/.local/bin/zbd"
	# 구매 확인 — zbd 설정(토큰)으로 zbd.kr 에 한 번 묻는다. 안 됐으면 조용히 건너뛴다
	if [ -n "$ZBD_BIN" ] && TUNNEL_HOST="$(node -e '
		const fs = require("fs"), os = require("os"), path = require("path");
		const c = [process.env.ZBD_CONFIG, path.join(process.env.XDG_CONFIG_HOME || path.join(os.homedir(), ".config"), "zbd/config.json"), process.env.APPDATA && path.join(process.env.APPDATA, "zbd/config.json")].filter(Boolean).find((f) => fs.existsSync(f));
		if (!c) process.exit(1);
		const cfg = JSON.parse(fs.readFileSync(c, "utf8"));
		if (!cfg.token) process.exit(1);
		fetch(process.env.ZBD_SERVER || cfg.server || "https://zbd.kr/service/plugin/cli", {
			method: "POST", signal: AbortSignal.timeout(8000),
			headers: { "content-type": "application/json", authorization: "Bearer " + cfg.token },
			body: JSON.stringify({ a: "tunnel", plugin: "tunnel", machineId: cfg.machineId })
		}).then((r) => r.json()).then((r) => { if (r.code !== "0000") process.exit(1); console.log(r.data.host); }).catch(() => process.exit(1));
	' 2>/dev/null)"; then
		say "외부 접속 주소 https://$TUNNEL_HOST (끄기: TUNNEL=0 ./run_dev.sh)"
		("$ZBD_BIN" tunnel --port "$FE_PORT" 2>&1) | prefix $'\033[34m[tn]\033[0m' &
	fi
fi
if [ "${SOCK_RUN:-}" = 1 ]; then
	if port_busy "$SOCK_PORT"; then echo "  포트 $SOCK_PORT 에 이미 소켓 데몬이 떠 있습니다 — 그대로 씁니다"
	else (cd "$SD" && PORT="$SOCK_PORT" exec ./zbd-socket 2>&1) | prefix $'\033[36m[ws]\033[0m' & fi
fi
if [ "${CC_RUN:-}" = 1 ]; then
	# 업무 채팅 중계 — 화면 주소로 붙는다(소켓·API 모두 vite 를 거친다). 설정·대화 목록은 .dev/claude-bridge
	# 중계 = zbd 안의 Go 중계(zbd bridge run — Claude 통합 구독 확인), 옛 플러그인 판은 JS 중계(node bridge.mjs)
	CC_CMD=(node bridge.mjs)
	[ -n "${CC_ZBD:-}" ] && CC_CMD=("$CC_ZBD" bridge run)
	(cd "$ROOT/plugins/claude-bridge" && CC_CONF_DIR="$DEV/claude-bridge" CLAUDE_BRIDGE_SECRET="$CC_SECRET" CLAUDE_BRIDGE_OWNER="$CC_OWNER" \
		CLAUDE_BE_URL="http://localhost:$FE_PORT" SOCKET_URL="http://localhost:$FE_PORT" CC_WORKSPACE="$ROOT" CC_CLAUDE_BIN="$CLAUDE" \
		exec "${CC_CMD[@]}" 2>&1) | prefix $'\033[32m[cc]\033[0m' &
fi
# 플러그인을 붙이거나(zbd add) 스타터를 업데이트하면(zbd update) .zbd-plugins.json 이 바뀐다 — 그때 스스로 껐다 다시 띄운다
# (새 메뉴 seed·마이그레이션·소켓 데몬·Claude 중계까지 run_dev 가 다시 준비한다)
plugin_stamp() { cksum "$ROOT/.zbd-plugins.json" 2>/dev/null || echo none; }
STAMP="$(plugin_stamp)"
while [ -n "$(jobs -pr)" ]; do
	sleep 2
	if [ "$(plugin_stamp)" != "$STAMP" ]; then
		sleep 3 # 붙이기 도구가 파일을 다 쓸 때까지
		say "플러그인이 바뀌어 다시 띄웁니다"
		cleanup
		exec "$ROOT/run_dev.sh" "$@"
	fi
done
