#!/usr/bin/env bash
# 운영 배포 — 내 컴퓨터에서 실행한다(macOS·Linux·Windows 의 Git Bash). 서버는 Ubuntu(22.04·24.04) — ssh 로 sudo 가 되는 계정이면 된다.
#
#   ./deploy/deploy.sh <ssh 대상> --domain example.com [--admin admin.example.com] [--name zbd-starter] [--port ssh포트] [--ssl 이메일]   처음 한 번
#   ./deploy/deploy.sh                      두 번째부터는 deploy/deploy.conf 에 남긴 값으로(주소는 서버 DB 의 도메인 설정을 쓴다)
#   ./deploy/deploy.sh --rollback           서버의 앞 판으로 되돌리기(빌드·올리기 없이 몇 초 — DB 는 그대로). 새 판이 뜨지 않으면 배포가 알아서 되돌린다
#
# 기본으로 맞추는 것(2026-10-08): 암호화 키 = 이 컴퓨터 .env.local 과 같게(--server-keys 로 끔) · 개발에서 올린 파일 → 서버(새 것만, 지우지 않음 — --no-files 로 끔)
# 하는 일: 서버 CPU 확인 → 서버 프로그램(Go)·소켓 서버(웹소켓 플러그인이 있으면)·화면(Svelte)을 리눅스용으로 빌드 →
#          꾸러미 하나로 올림 → 서버에서 deploy/server/install.sh(nginx·MariaDB·설정·DB·첫 관리자·systemd·nginx 연결·확인)
# 주소의 원본 = 서버 DB 의 도메인 설정(관리자 「환경 설정 › 도메인」 — 사용자 주소·관리자 주소). --domain·--admin(기본 admin.<도메인>, 쉼표로 여럿)은
#   **처음 설치 때만** 그 값을 만든다 — 그 뒤 주소를 바꾸려면 관리자 화면에서 바꾸고 이 스크립트를 다시 돌린다(nginx·인증서를 DB 주소대로). 모든 주소의 DNS 를 이 서버로.
# 플러그인이 관리자 단독 앱 주소를 더할 수 있다(PWA 채팅 앱 = chat.<대표 주소>) — 아래 「관리자 주소」 줄 뒤에서 ADMIN_APPS 에 앞이름을. HTTPS 는 --ssl 이메일(Let's Encrypt — 인증서·443·http→https·자동 갱신).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONF="$ROOT/deploy/deploy.conf"
HOST="" PORT="" DOMAIN="" ADMIN="" NAME="" OUT="" ARCH="" SSL="" ADMIN_APPS="" KEYS="" FILES="" ROLLBACK=0
[ -f "$CONF" ] && . "$CONF"
while [ $# -gt 0 ]; do
	case "$1" in
		--domain) DOMAIN="$2"; shift 2 ;;
		--admin) ADMIN="$2"; shift 2 ;;
		--name) NAME="$2"; shift 2 ;;
		--port) PORT="$2"; shift 2 ;;   # ssh 포트(22 가 아닐 때). 한 번 주면 deploy.conf 에 남는다
		--ssl) SSL="$2"; shift 2 ;;     # HTTPS — Let's Encrypt 만료 안내 이메일(두 주소 인증서·443·자동 갱신). 한 번 주면 deploy.conf 에 남는다
		--server-keys) KEYS=server; shift ;; # 서버 키를 따로 둔다(개발 컴퓨터 키로 맞추지 않음). 한 번 주면 deploy.conf 에 남는다 — 되돌리기 --dev-keys
		--dev-keys) KEYS=""; shift ;;
		--no-files) FILES=0; shift ;;        # 개발에서 올린 파일을 보내지 않는다. 한 번 주면 남는다 — 되돌리기 --files
		--files) FILES=""; shift ;;
		--out) OUT="$2"; shift 2 ;;     # 꾸러미만 만들어 이 파일로(서버에 접속하지 않는다 — 손으로 올릴 때·시험)
		--arch) ARCH="$2"; shift 2 ;;   # --out 과 함께: amd64(기본) · arm64
		--rollback) ROLLBACK=1; shift ;; # 앞 판으로 되돌리기
		-*) echo "✗ 모르는 옵션: $1"; exit 1 ;;
		*) [ -z "${ARG_HOST:-}" ] || { echo "✗ 서버(ssh 대상)는 하나만 넣습니다: $ARG_HOST 와 $1 — 이메일이면 앞에 --ssl 을 붙이세요(--ssl $1)"; exit 1; }
			ARG_HOST="$1"; HOST="$1"; shift ;;
	esac
done
# 흔한 실수 막기 — 메일 주소를 서버로 넣음(--ssl 빠뜨림) · --ssl 값이 이메일이 아님
case "${HOST#*@}" in
	gmail.com|naver.com|daum.net|hanmail.net|kakao.com|nate.com|outlook.com|hotmail.com|live.com|icloud.com|me.com|yahoo.com|proton.me|protonmail.com)
		echo "✗ 서버 자리에 메일 주소($HOST)가 들어갔습니다 — HTTPS 이메일이면 앞에 --ssl 을 붙이세요: --ssl $HOST"; exit 1 ;;
esac
[ -z "$PORT" ] || [[ "$PORT" =~ ^[0-9]{1,5}$ ]] || { echo "✗ --port 다음엔 숫자(ssh 포트)를 넣습니다: $PORT"; exit 1; }
SSH_P=(); SCP_P=(); [ -z "$PORT" ] || { SSH_P=(-p "$PORT"); SCP_P=(-P "$PORT"); }
[ -z "$SSL" ] || [[ "$SSL" =~ ^[^@[:space:]-][^@[:space:]]*@[^@[:space:]]+\.[^@[:space:]]+$ ]] || { echo "✗ --ssl 다음엔 이메일을 넣습니다(인증서 만료 안내 받을 주소): $SSL"; exit 1; }
NAME="${NAME:-zbd-starter}"
[ -n "$HOST" ] || [ -n "$OUT" ] || { echo "✗ 서버(ssh 대상)를 넣어 주세요: ./deploy/deploy.sh ubuntu@서버주소 --domain example.com"; exit 1; }
# 처음 설치엔 --domain 이 필요하다(서버에 주소가 없으면 install.sh 가 알려 준다). 그 뒤엔 서버 DB 주소를 쓴다
[ -z "$DOMAIN" ] || ADMIN="${ADMIN:-admin.$DOMAIN}"
# 관리자 주소 — 플러그인이 이 줄 뒤에 주소를 더한다(ADMIN_APPS="$ADMIN_APPS chat" → chat.<대표 사용자 주소> · mail · office · photo)
[[ "$NAME" =~ ^[a-z][a-z0-9-]{1,30}$ ]] || { echo "✗ --name 은 영문 소문자·숫자·- 만"; exit 1; }

say() { printf '\033[1;36m▶ %s\033[0m\n' "$*"; }
if [ "$ROLLBACK" = 1 ]; then
	[ -n "$HOST" ] || { echo "✗ 배포한 적이 없습니다(deploy/deploy.conf 없음)"; exit 1; }
	say "앞 판으로 되돌리기 ($HOST · $NAME)"
	ssh ${SSH_P[@]+"${SSH_P[@]}"} -t "$HOST" "if [ ! -f /srv/$NAME/install.sh ]; then echo '✗ 이 서버에는 아직 되돌리기 도구가 없습니다 — 새 판으로 한 번 배포한 뒤부터 됩니다(앞 판은 /srv/$NAME/releases 에 있습니다)'; exit 1; fi; sudo bash /srv/$NAME/install.sh --rollback --name $NAME"
	exit $?
fi
if [ -n "$OUT" ]; then M="${ARCH:-amd64}"; else
	say "서버 확인 ($HOST)"
	M="$(ssh ${SSH_P[@]+"${SSH_P[@]}"} -o ConnectTimeout=15 "$HOST" uname -m)" || { echo "✗ 서버($HOST)에 ssh 로 접속하지 못했습니다 — 주소·계정·ssh 키·포트(22 가 아니면 --port 번호)를 확인하세요(설정은 저장하지 않았습니다)"; exit 1; }
	# 접속이 된 값만 다음 배포용으로 남긴다(틀린 값이 deploy.conf 에 남지 않게)
	printf 'HOST=%q\nPORT=%q\nDOMAIN=%q\nADMIN=%q\nNAME=%q\nSSL=%q\nKEYS=%q\nFILES=%q\n' "$HOST" "$PORT" "$DOMAIN" "$ADMIN" "$NAME" "$SSL" "$KEYS" "$FILES" >"$CONF"
fi
case "$M" in x86_64|amd64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) echo "✗ 지원하지 않는 CPU: $M"; exit 1 ;; esac
echo "  linux/$ARCH · 이름 $NAME${DOMAIN:+ · 처음 설치 주소 $DOMAIN · $ADMIN}"

PKG="$(mktemp -d)"; trap 'rm -rf "$PKG"' EXIT
mkdir -p "$PKG/bin" "$PKG/web" "$PKG/db"
say "서버 프로그램 빌드(Go)"
(cd "$ROOT/zbd-be-golang" && for c in server create-admin rekey; do
	out="zbd-be"; [ "$c" = server ] || out="$c"
	GOOS=linux GOARCH=$ARCH CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$PKG/bin/$out" "./cmd/$c"
done)
if [ -f "$ROOT/plugins/zbd-socket/main.go" ]; then
	say "소켓 서버 빌드(웹소켓 플러그인)"
	(cd "$ROOT/plugins/zbd-socket" && GOOS=linux GOARCH=$ARCH CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$PKG/bin/zbd-socket" .)
fi
say "화면 빌드(Svelte)"
(cd "$ROOT/zbd-fe-svelte5" && { [ -d node_modules ] || npm ci --no-audit --no-fund; } &&
	npm run build >/dev/null) # 주소는 빌드에 넣지 않는다 — 서버가 문서에 알려 준다(도메인 설정)
cp -R "$ROOT/zbd-fe-svelte5/build/." "$PKG/web/"
cp -R "$ROOT/db/migrations" "$ROOT/db/seed" "$ROOT/db/seed-guard" "$ROOT/db/assets" "$PKG/db/"
[ -d "$ROOT/db/removals" ] && cp -R "$ROOT/db/removals" "$PKG/db/" # zbd remove 가 남긴 DB 지우기
cp "$ROOT/deploy/server/install.sh" "$PKG/"
# E2E 테스트 플러그인(zbd add e2e-test) — 결과 올리기 비밀값을 이 컴퓨터 개발 설정과 같게(저장소의 e2e 키트가 운영 사이트에도 같은 값으로 올린다)
envlocal() { grep -E "^$1=" "$ROOT/zbd-be-golang/.env.local" 2>/dev/null | tail -1 | cut -d= -f2- | tr -d '"' || true; }
E2E_S="$(envlocal E2E_UPLOAD_SECRET)"
[ -z "$E2E_S" ] || (umask 077 && printf '%s' "$E2E_S" >"$PKG/e2e-upload-secret")
# 암호화 키 = 이 컴퓨터 개발 설정(.env.local)과 같게(기본 — 2026-10-08). 개발에서 만든 회원·관리자 2단계 인증을 서버도 읽는다.
#   서버 키가 다르면 install.sh 가 서버 DB 의 회원 정보를 새 키로 다시 잠근 뒤(bin/rekey) .env 를 바꾼다 — 로그인은 한 번 풀린다.
#   --server-keys 면 보내지 않는다(서버 키 따로). 꾸러미만 만들 때(--out)도 키는 넣지 않는다(파일로 돌아다니지 않게)
if [ "$KEYS" != server ] && [ -z "$OUT" ] && [ -f "$ROOT/zbd-be-golang/.env.local" ]; then
	(umask 077 && for k in AES_256_ENCRYPTION_KEY AES_256_ENCRYPTION_IV PII_ENC_KEY_V1 PII_INDEX_KEY; do
		v="$(envlocal "$k")"; [ -z "$v" ] || printf '%s=%s\n' "$k" "$v"
	done >"$PKG/dev-keys")
fi
# 관리자 단독 앱 주소(PWA 채팅·메일 플러그인) — 즈비드 호스팅(zbd hosting deploy)은 이 파일로 받는다(install.sh 는 --admin-apps)
[ -z "${ADMIN_APPS# }" ] || printf '%s\n' "${ADMIN_APPS# }" >"$PKG/admin-apps"
(cd "$ROOT" && git rev-parse --short HEAD 2>/dev/null || echo "-") >"$PKG/REVISION"
COPYFILE_DISABLE=1 tar -czf "$PKG.tgz" $( [ "$(uname)" = Darwin ] && echo --no-mac-metadata --no-xattrs ) -C "$PKG" .
echo "  꾸러미 $(du -h "$PKG.tgz" | cut -f1)"

if [ -n "$OUT" ]; then
	mv "$PKG.tgz" "$OUT"
	echo "✓ 꾸러미: $OUT — 서버에서: mkdir d && tar -xzf 꾸러미 -C d && sudo bash d/install.sh --pkg d --name $NAME ${DOMAIN:+--domain $DOMAIN --admin $ADMIN}${ADMIN_APPS:+ --admin-apps \"${ADMIN_APPS# }\"}${SSL:+ --ssl $SSL}"
	exit 0
fi
say "올리기 · 설치"
scp ${SCP_P[@]+"${SCP_P[@]}"} -q "$PKG.tgz" "$HOST:/tmp/$NAME-deploy.tgz"
rm -f "$PKG.tgz"
# 개발에서 올린 파일(로고·이미지·첨부·웹하드) → 서버. 서버의 받는 자리(~/.zbd-sync/<이름>)에 먼저 보내고 install.sh 가 사이트 폴더로 옮긴다:
#   서버 쪽이 더 새 파일이면 그대로 · 지우지 않는다 · secure(키 파일 등)는 서버에 없는 것만. --no-files 면 보내지 않는다
SYNC_ARG=""
if [ "$FILES" != 0 ]; then
	UPD="$(envlocal CDN_UPLOAD_DIR)"; DEVD="${UPD%/}"; DEVD="${DEVD%/upload}"; [ -n "$UPD" ] || DEVD="$ROOT/.dev"
	SUBS=(); for s in upload secure file-explorer; do [ -d "$DEVD/$s" ] && SUBS+=("$s"); done
	if [ ${#SUBS[@]} -gt 0 ]; then
		say "개발에서 올린 파일 보내기(${SUBS[*]})"
		R=".zbd-sync/$NAME"
		ssh ${SSH_P[@]+"${SSH_P[@]}"} "$HOST" "mkdir -p $R && chmod 700 .zbd-sync"
		if command -v rsync >/dev/null && ssh ${SSH_P[@]+"${SSH_P[@]}"} "$HOST" "command -v rsync" >/dev/null 2>&1; then
			for s in "${SUBS[@]}"; do rsync -a -e "ssh ${PORT:+-p $PORT}" "$DEVD/$s/" "$HOST:$R/$s/"; done
		else
			COPYFILE_DISABLE=1 tar -cf - -C "$DEVD" "${SUBS[@]}" | ssh ${SSH_P[@]+"${SSH_P[@]}"} "$HOST" "tar -xf - -C $R"
		fi
		SYNC_ARG="--sync-from \$HOME/$R"
	fi
fi
ssh ${SSH_P[@]+"${SSH_P[@]}"} -t "$HOST" "set -e; D=\$(mktemp -d); tar -xzf /tmp/$NAME-deploy.tgz -C \$D; rm -f /tmp/$NAME-deploy.tgz; sudo bash \$D/install.sh --pkg \$D --name $NAME $SYNC_ARG ${DOMAIN:+--domain $DOMAIN --admin $ADMIN}${ADMIN_APPS:+ --admin-apps \"${ADMIN_APPS# }\"}${SSL:+ --ssl $SSL}; rm -rf \$D"
