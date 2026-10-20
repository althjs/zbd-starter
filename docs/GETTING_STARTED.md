# 처음 시작하기 — 새 고객용 안내

zbd-starter 를 받아 **내 컴퓨터에서 내 사이트를 띄우고, 관리자로 들어가 첫 화면을 만드는 데까지** 안내한다.
(2026-10-04 시험판 기준 — 아래 「아직 없는 것」 도 꼭 읽을 것)

---

## 0. 준비물

| | 내용 |
|---|---|
| 컴퓨터 | macOS(Apple 실리콘·인텔) · Linux(Ubuntu 계열) · **Windows 10(1809 이상)·11** |
| macOS | **Homebrew** 를 먼저 설치 — https://brew.sh (관리자 비밀번호가 필요해 스크립트가 대신 깔지 않는다) |
| git · GitHub | 저장소를 받을 수 있는 계정. 시험 기간엔 비공개 저장소라 **초대(Collaborator)를 받은 계정**이어야 한다 |
| 나머지 | Go · Node.js 22+ · MariaDB · air 는 `run_dev.sh`(Windows 는 `run_dev.bat`)가 없으면 알아서 설치한다 — Windows 는 Go·MariaDB 를 관리자 권한 없이 `%LOCALAPPDATA%\zbd-tools` 에 둔다 |
| 권장 | Claude Code — 이 폴더를 열면 `CLAUDE.md` 규칙대로 Claude 가 같이 일한다 |

포트 **5000(화면) · 5001(서버) · 3306(DB)** 를 쓴다.
⚠ macOS 는 「AirPlay 수신 모드」 가 5000 을 쓰고 있을 수 있다 → 시스템 설정 › 일반 › AirDrop 및 Handoff 에서 끄거나 포트를 바꾼다(아래 「포트 바꾸기」).

---

## 1. 받기

```bash
git clone git@github.com:althjs/zbd-starter.git my-service
cd my-service
```

`my-service` 는 원하는 이름으로. 이 폴더가 앞으로 **내 서비스의 저장소**가 된다 — 나중에 내 GitHub 저장소로 옮기려면
`git remote set-url origin git@github.com:<나>/<저장소>.git` 후 push.

---

## 2. 띄우기 — 명령 하나

```bash
./run_dev.sh
```

**Windows** 는 탐색기에서 폴더를 열고 **`run_dev.bat` 을 더블클릭**한다(git 은 [Git for Windows](https://git-scm.com/download/win) 로 받는다).
서버는 따로 열리는 창 `[be]` 에서, 화면은 그 창에서 돈다 — 화면 창에서 Ctrl+C 로 끄면 서버 창도 같이 꺼진다. DB(MariaDB)는 뒤에서 계속 돈다.
처음 실행 때 방화벽 창이 뜨면 「취소」 해도 이 컴퓨터 안에서는 된다.

같은 네트워크의 다른 기기(휴대폰·다른 PC)에서도 `http://<이 컴퓨터 IP>:5000` 으로 바로 열린다(화면에 「Network:」 주소가 나온다 — 관리자는 `/admin`). 이 컴퓨터에서만 열려면 `zbd-be-golang/.env.local` 에 `FE_HOST=localhost`. 서버·DB 는 그대로 이 컴퓨터 안에서만 받는다. Windows 는 방화벽 창에서 「개인 네트워크」 를 허용한다.

처음 한 번은 몇 분 걸린다. 하는 일:

1. 도구 확인·설치(Go · Node · MariaDB · air)
2. MariaDB 켜기(macOS 는 로그인 때 자동으로 켜지는 서비스로 등록)
3. 설정 파일 `zbd-be-golang/.env.local` 생성 — DB 비밀번호·암호화 키를 **무작위로**
4. DB `zbd_starter` · 계정 `zbd_dev` 만들기 → 표 13개 → 기본 데이터
5. 첫 관리자 `admin@example.com` · 시험 회원 `user@example.com` — 비밀번호는 무작위로 `.dev/account.txt` 에
6. 서버(5001)·화면(5000) 실행 — 로그가 `[be]`·`[fe]` 로 한 화면에

끝나면 `▶ BE http://127.0.0.1:5001 · FE http://localhost:5000` 이 보인다. **Ctrl+C** 로 둘 다 끈다. 다음부터는 같은 명령으로 몇 초 만에 뜬다.

> ⚠ `.env.local` 의 `PII_ENC_KEY_V1`·`PII_INDEX_KEY` 를 **잃으면 회원 이름·이메일·전화를 다시 읽을 수 없다.** git 에 올라가지 않으니 비밀번호 관리자 등에 따로 보관한다.

---

## 3. 관리자로 들어가기

> 개발 모드에서는 홈(http://localhost:5000/) 맨 아래 「관리자 암호 보기 · 사용자 암호 보기」 로도 계정을 볼 수 있다.
> 처음 설치하면 예시 데이터가 들어 있다 — 공지 2개(첨부파일 포함)·자유게시판 글 2개(시험 회원)·약관 2개·업무 카드 7장·웹하드 시작 가이드. 확인한 뒤 지워도 된다.

1. `cat .dev/account.txt` 로 이메일·비밀번호 확인
2. http://localhost:5000/admin → 로그인
3. 오른쪽 위 내 정보에서 **비밀번호를 바꾼다**
4. (선택) 내 정보 › **지문·얼굴 로그인** — 이 기기를 등록하면 다음부터 비밀번호 없이 들어간다
   (개발 주소의 패스키는 `localhost` 전용이다. 실제 도메인에선 다시 등록)

왼쪽 메뉴(기본):

| 메뉴 | 하는 일 |
|---|---|
| 대시보드 | 요약(처음엔 비어 있다) |
| 업무관리 | 할 일 보드(TASK) — Claude 와 일을 주고받는 곳 |
| 웹하드 | 문서·파일 |
| 게시판 | 공지사항·약관 등 관리자 게시판 글 |
| 회원 › 사용자 회원 / 관리자 회원 / 권한 그룹 / 로그인 기기 | 회원·관리자·권한·로그인한 브라우저 |
| 설정 › 환경 설정 / 메뉴 설정 / 공통 코드 | 사이트 이름·로고·로그인 정책 등 |

---

## 4. 내 사이트 화면 만들기

스타터에는 「페이지 디자이너」가 없다 — 사용자 화면은 **코드(Svelte)** 로 만든다. Claude Code 에 「홈에 소개 섹션 넣어 줘」 처럼 시키면 된다.

| 무엇 | 파일 |
|---|---|
| 홈(/) | `zbd-fe-svelte5/src/routes/(user)/(site)/+page.svelte` |
| 머리 메뉴 | `zbd-fe-svelte5/src/lib/user/layout/SiteHeader.svelte` (위쪽 `MENU`) |
| 바닥글 | `zbd-fe-svelte5/src/lib/user/layout/SiteFooter.svelte` (사업자 정보는 환경 설정 값) |
| 새 페이지 | `src/routes/(user)/(site)/{주소}/+page.svelte` |

- 사이트 이름·회사 정보·약관 정책은 관리자 **설정 › 환경 설정**.
- 공지사항은 관리자 **게시판 › 공지사항** 에서 쓰고 http://localhost:5000/board/notice 에 나온다.
- 약관은 **이용약관·개인정보 처리방침 표준 양식**이 기본으로 들어 있다 — 관리자 **게시판 › 약관** 에서 `[회사명]`·`[시행일]`·`[연락처]` 를 실제 값으로 고친다(회원 가입 화면에 나온다).
- 회원 가입은 **이메일·비밀번호**(http://localhost:5000/user/join), 로그인 뒤 내 정보에서 지문·얼굴(패스키)을 등록할 수 있다. 비밀번호 찾기·2단계 인증은 메일 설정이 필요해 스타터에 없다(플러그인).

---

## 5. 자주 쓰는 것

```bash
./run_dev.sh                         # 띄우기(Ctrl+C 로 끄기)
./db/setup.sh --status               # 적용된 DB 마이그레이션
BE_PORT=5101 FE_PORT=5100 ./run_dev.sh   # 이번 한 번만 다른 포트로
```

### 포트 바꾸기

늘 다른 포트로 띄우려면 `zbd-be-golang/.env.local` 에 적는다(처음 실행 때 만들어진다 — git 에 안 올라간다).

```bash
# zbd-be-golang/.env.local
PORT=5101        # 서버(BE) — 기본 5001 (BE_PORT=5101 로 적어도 된다)
FE_PORT=5100     # 화면(FE) — 기본 5000. 관리자 주소도 http://localhost:5100/admin 이 된다
SOCK_PORT=4016   # 소켓 데몬 — 기본 4006, 웹소켓 플러그인을 붙였을 때만 쓴다
```

- 고친 뒤 `./run_dev.sh` 를 껐다 켠다.
- 순서: 명령 앞 값(`FE_PORT=… ./run_dev.sh`) → `.env.local` → 기본값. 명령 앞 값은 그 실행에만 쓰이고 파일은 안 바뀐다.
- 서버·화면 포트가 같거나 숫자가 아니면 바로 멈추고 알린다. 포트가 이미 쓰이고 있으면 무엇이 쥐고 있는지 알려 준다.
- 소켓 데몬 포트를 다른 프로그램이 쓰고 있으면 빈 포트(+10씩)를 골라 `.env.local SOCK_PORT` 에 남긴다.
- DB 포트는 같은 파일의 `DB_PORT`(기본 3306).
- Claude 통합 중계(`zbd install claude-integration`)를 붙였다면, 중계 설정의 관리자 주소도 바뀐 화면 포트(`http://localhost:5100`)로 맞춘다.

- 코드를 저장하면 서버(air)·화면(vite)이 알아서 다시 읽는다. `.env.local` 을 고쳤으면 `run_dev.sh` 를 껐다 켠다.
- 검사: 서버 `gofmt -l . && go vet ./... && go build ./...` · 화면 `npx svelte-check && npx prettier --check .`
- 새 표·칸은 `db/migrations/NNNN_이름.up.sql` + `.down.sql`(되돌리기)로 더한다.

**처음부터 다시(내 컴퓨터의 데이터가 전부 지워진다)**

```bash
./reset.sh        # DB·.dev/(계정·업로드·웹하드) 지우기 — 묻고 진행, 서버가 떠 있으면 멈춘다
./run_dev.sh      # 새 DB·예시 데이터·새 계정
```

`./reset.sh --all` 은 설정 파일(`.env.local`, 암호화 키)과 DB 계정까지 지운다. `-y` 는 묻지 않는다.

---

## 6. 유료 플러그인

페이지 디자이너·소셜 로그인(간편 가입)·웹 푸시·웹소켓(실시간)·문자 인증·메신저·웹메일·결제·외부 공격 감지·SSL 자동 갱신 등은 **스타터에 들어 있지 않다** — 구매하면 `zbd` 명령으로 내 저장소에 끼워 넣는다.

```bash
curl -fsSL https://zbd.kr/upload/ZBD/cli/install.sh | sh   # zbd 명령 설치
zbd login            # 즈비드 마이페이지에서 발급한 토큰
zbd install ssl      # 예: SSL 인증서 자동 갱신
```

플러그인 목록·가격은 https://zbd.kr/app/zbd-starter

**라이선스 확인** — 플러그인을 붙이면 `zbd` 가 `zbd-be-golang/internal/api/core/zz_plugins_gen.go` 에 붙인 플러그인 키·판과 구매자 표시를 적는다(고치지 않는다 — 다음 `zbd add`·`zbd update` 가 다시 쓴다).
운영 서버는 켠 뒤 1분, 그 뒤 몇 시간마다 https://zbd.kr 에 **구매자 표시 · 사이트 주소 · 플러그인 키와 판**만 보내 라이선스를 확인한다(회원·글 같은 사이트 데이터는 보내지 않는다).
구입 기록이 없거나 정지·기간이 끝난 플러그인이 있으면 **관리자 화면 맨 위에 안내가 뜬다** — 사이트는 멈추지 않는다. 내 컴퓨터(run_dev)에서는 확인하지 않는다.
한 구매로 여러 사이트를 돌릴 수 있는 수에는 한도가 있다(개발·시험 주소는 세지 않는다).

---

## 7. 운영 서버에 올리기 — `zbd deploy`

우분투 서버(22.04·24.04, ssh 로 sudo 가 되는 계정) 하나에 사이트·관리자·DB·(웹소켓 플러그인이 붙어 있으면)소켓 서버까지 한 번에 올린다.
도메인 두 개의 DNS 를 서버 IP 로 먼저 돌려 둔다 — 사이트(`example.com`)와 관리자(`admin.example.com`).
`zbd` 명령이 없으면 먼저 설치한다(아래 「플러그인 붙이기」 의 설치 한 줄).

```bash
zbd deploy ubuntu@서버주소 --domain example.com --ssl me@example.com   # 처음 한 번 — 관리자 주소는 admin.example.com, HTTPS 포함
zbd deploy ubuntu@서버주소 --domain example.com --admin office.example.com
zbd deploy                                                   # 두 번째부터 — 주소 옵션 없이
```

- **주소(도메인)는 관리자 「환경 설정 › 도메인」 한 곳**에서 정한다 — 사용자 주소·관리자 주소(각각 여럿, 첫 번째가 대표). 홈 주소·로그인 쿠키 범위·사이트맵 주소는 여기서 자동으로 계산된다.
  `--domain`·`--admin` 은 **처음 설치 때만** 그 값을 만든다. 그 뒤 주소를 바꾸려면 화면에서 바꾸고 → 새 주소의 DNS 를 서버로 → `zbd deploy` 한 번(nginx·인증서를 DB 주소대로 맞춘다). 옵션을 다시 줘도 DB 주소가 이긴다.
  내 컴퓨터(개발)는 사용자·관리자 모두 `localhost:5000` 으로 등록돼 있다(한 주소에서 둘 다 연다).

- **커밋된 것만 올린다** — 커밋 안 된 변경이 있으면 멈추고 목록을 보여 준다(무엇이 나갔는지 git 기록으로 알 수 있게).
- 안에서는 저장소의 `deploy/deploy.sh` 를 그대로 돌린다(스크립트도 `zbd update` 로 같이 최신이 된다).
- 내 컴퓨터에서 리눅스용으로 빌드(서버 CPU 에 맞춰 amd64/arm64) → 꾸러미 하나를 올려 서버에서 `install.sh` 가 nginx·MariaDB 설치, 설정(`/srv/{이름}/.env`, 처음 한 번 무작위 키), DB 준비, 첫 관리자, systemd 서비스, nginx 연결, 확인까지 한다.
- 첫 관리자 계정은 서버 `/srv/{이름}/account.txt` — 로그인해 비밀번호를 바꾸고 파일을 지운다.
- 다시 배포하면 새 판(`releases/`)으로 바꿔 끼운다(앞 판 3개 보관). `.env`·올린 파일(`data/`)·DB 는 그대로.
- 한 서버에 사이트 여럿: `--name` 을 다르게 주면 된다. 포트는 처음 설치 때 비어 있는 것을 골라(서버 프로그램 5001~·소켓 서버 4006~) `.env` 에 남기고, 다시 배포해도 그대로다.
- **ssh 포트**: 서버 ssh 가 22 가 아니면 `--port 2222` 처럼. 한 번 주면 `deploy.conf` 에 남는다(`~/.ssh/config` 의 Host 별칭을 서버 자리에 써도 된다).
- **HTTPS**: `--ssl 내이메일@example.com` 을 붙이면 모든 주소(사용자·관리자)의 Let's Encrypt 인증서를 한 장으로 받고, nginx 를 443 으로 바꾸고, http 는 https 로 넘기고, 자동 갱신(갱신 뒤 nginx 다시 읽기)까지 켠다. (주의: `--ssl` 을 빼고 이메일만 쓰면 서버 주소로 읽힌다 — 이제는 오류로 알려 준다) 한 번 주면 `deploy.conf` 에 남는다. 두 주소의 DNS 가 서버를 가리키고 80 포트가 열려 있어야 한다 — 발급이 안 되면 사이트는 http 로 계속 돌고 원인을 보여 준다.
### 내 컴퓨터와 서버 — 처음엔 개발한 그대로, 그 뒤엔 서버 DB 가 원본

- **암호화 키는 내 컴퓨터와 같게 맞춘다(기본)** — 회원 이름·이메일·전화를 잠그는 키(`PII_ENC_KEY_V1`·`PII_INDEX_KEY`)와 로그인 토큰 키(`AES_256_*`)를
  `zbd deploy` 가 내 컴퓨터 개발 설정(`zbd-be-golang/.env.local`)에서 가져가 서버 `.env` 에 넣는다.
  서버 키가 달랐으면 서버 DB 의 회원 정보를 새 키로 다시 잠근 뒤 바꾼다(옛 설정은 `/srv/{이름}/.env.bak_…_keys` 로 남김). 이때 모두 한 번 다시 로그인한다.
  그래서 DB 를 옮기거나 같이 써도 회원 정보를 읽을 수 있다. 서버 키를 따로 두려면 `zbd deploy --server-keys`(되돌리기 `--dev-keys`).
  ⚠ 키를 잃으면 회원 정보를 못 읽는다 — `.env.local` 과 서버 `.env` 를 따로 백업해 둔다. 다른 컴퓨터에서 배포하면 그 컴퓨터의 키로 바뀐다(먼저 `.env.local` 의 네 키를 같게 옮겨 두기).
- **내 컴퓨터에서 올린 파일도 같이 간다(기본)** — 로고·이미지(`upload`)·웹하드(`file-explorer`)는 새 것·바뀐 것만 서버로 보낸다. 서버 쪽 파일이 더 새것이면 그대로 두고, 지우지는 않는다.
  `secure`(키 파일 등)는 서버에 없는 것만. 보내지 않으려면 `--no-files`(되돌리기 `--files`).
- **설정(.env)도 내 컴퓨터 값으로(기본)** — `.env.local` 에 직접 넣은 값(소셜 로그인·결제·문자·AI 키 등)을 서버 `.env` 에 같은 값으로 넣는다(바뀐 이름을 배포 끝에 보여 주고, 앞 설정은 `.env.bak_…_devenv`).
  서버가 정하는 값(포트·DB·폴더·쿠키 이름·소켓·nginx·메일 서버·개발 전용 `MAIL_DEV_LOG`)과 `localhost` 를 가리키는 값(예: `PAY_BASE_URL=http://localhost:5000`)은 보내지 않는다 — 뒤의 것은 경고가 뜨니 서버 `.env` 에 운영 값을 넣는다.
  서버 설정을 따로 두려면 `--server-env`(되돌리기 `--dev-env`).
- **처음 배포는 개발 DB 를 그대로 옮긴다** — 서버에 이 사이트가 없을 때 `zbd deploy` 가 내 컴퓨터 개발 DB 를 떠서 서버 DB 에 넣는다: 회원·글·메뉴·디자인·설정이 개발한 그대로, **관리자 계정·비밀번호도 run_dev 에서 쓰던 그대로**(`account.txt` 는 안 생김).
  DB 이름·계정·비밀번호는 서버 것(개발용 계정을 운영에 쓰지 않는다), 주소 설정(도메인·홈 주소·쿠키 범위)만 `--domain`·`--admin` 으로 바뀐다. 빈 스타터로 시작하려면 `--no-db`.
- **그 뒤 배포는 DB 를 건드리지 않는다** — 운영에서 쌓인 데이터가 원본이다. 서버 DB 를 다시 개발 DB 로 덮어야 하면 `zbd deploy --with-db`(그 한 번만 — 서버 DB 를 `/srv/{이름}/backups/db/…-before-dev-db.sql.gz` 로 먼저 백업하고, 서버 주소 설정은 그대로 둔다. 되살리기: `gunzip -c 그파일 | sudo mariadb {DB 이름}`).
  내 컴퓨터와 서버가 **같은 장비**(개발 컴퓨터에 바로 배포)면 둘 다 같은 DB 를 쓸 수도 있다 — 개발에서 고친 설정·디자인이 운영에 바로 보인다(시험 데이터도).
- **서버 첫 관리자** = 서버의 `/srv/{이름}/account.txt`(`sudo cat` — 첫 배포 끝에 경로를 알려 준다).

- 서버 없이 꾸러미만: `zbd deploy --out pkg.tgz --domain example.com`.
- 문제가 생기면 서버에서 `journalctl -u {이름} -n 50`(소켓 서버는 `{이름}-socket`) · `sudo nginx -t`.

## 8. 스타터 업데이트 — `zbd update`

스타터 원본에 고친 것·새 기능이 나오면 **내 저장소에서 한 줄**로 받는다. 받은 뒤 확인하고 `zbd deploy`.

```bash
git status                 # 깨끗해야 한다(커밋 먼저)
zbd update                 # 원본 받기 → 합치기 → 붙인 플러그인 다시 붙이기 → 빌드 확인 → 커밋
./run_dev.sh               # 내 컴퓨터에서 확인
zbd deploy                 # 운영 서버에
```

- 원본은 git 원격 **`zbd-starter`** — 없으면 처음에 더한다. `git clone` 으로 받은 그대로 `origin` 이 원본이면 그대로 둬도 되고, 내 GitHub 저장소로 옮겼다면 원본은 `zbd-starter` 로 따로 남는다.
- 붙인 플러그인(`.zbd-plugins.json`)은 **최신 꾸러미로 다시 붙인다**(플러그인 업데이트도 같이 — 1회 구매는 업데이트 기간 안에서). 플러그인이 고친 파일이 원본과 부딪히면 원본을 받고 플러그인을 다시 붙여 알아서 푼다.
- **내가 직접 고친 파일**이 원본과 부딪히면 합치기를 되돌리고 그 파일 목록을 보여 준다 → `git merge zbd-starter/main` 으로 직접 합쳐 커밋한 뒤 다시 `zbd update`.
- ⚠ **플러그인 파일은 업데이트 때 덮어쓴다.** 플러그인이 붙인 파일(`.zbd-plugins.json` 의 `files` 목록)을 내가 고쳐 두었으면, 다시 붙일 때 그 파일을 `.zbd/backup/<시각>/` 에 남기고 **플러그인의 새 판으로 덮는다**(덮은 파일 목록이 화면에 나온다). 내 수정이 필요하면 업데이트 뒤 백업에서 다시 옮겨 담는다 — 커밋해 두었다면 `git diff HEAD~1 -- <파일>` 로도 볼 수 있다.
  - 바꾸고 싶은 것은 가능하면 플러그인 파일이 아닌 곳에 — 관리자 설정(문구·메뉴·디자인)이나 내가 만든 파일.
- 커밋은 두 개까지 생긴다: 「zbd update: 스타터 원본 합침」 · 「zbd update: 플러그인 다시 붙임」. 되돌리려면 `git reset --hard` 로 그 앞으로.

## 9. 아직 없는 것 (시험판)

- 호스팅 id 가 `ZBD` 로 고정돼 있고 즈비드 문구가 일부 남아 있다.

막히면 `[be]` 로그의 오류 줄과 함께 문의한다.

---

## 플러그인 붙이기 — `zbd add`

산 플러그인(예: **웹소켓 연동** — 업무관리·웹하드·고객 상담 등이 새로고침 없이 바로 맞춰진다 `zbd add websocket` · **Claude 통합** — 관리자 화면 「업무 채팅」 에서 내 장비의 Claude 와 일하기, 실시간 소켓 서버 포함)의 코드는 **스타터 폴더에서 명령 한 줄**로 붙인다.

```bash
# zbd 명령이 없으면 먼저 — macOS·Linux
curl -fsSL https://zbd.kr/upload/ZBD/cli/install.sh | sh
zbd login                      # zbd.kr 마이페이지 › 플러그인에서 발급한 토큰
zbd add claude-integration     # 이 폴더에 코드 붙이기(라이선스 확인 → 받기 → 붙이기)
```
Windows 는 https://zbd.kr/upload/ZBD/cli/claude-install.bat 을 받아 더블클릭하면 zbd 설치·로그인까지 이어진다(그다음 이 폴더에서 `zbd add claude-integration`).

붙인 뒤:
1. `run_dev` 를 다시 띄운다 — 소켓 서버(`plugins/zbd-socket`)도 같이 빌드·실행되고, 관리자 메뉴에 「Claude 협업 › 업무 채팅」 이 생긴다.
2. 바뀐 파일을 `git status` 로 보고 커밋한다(플러그인 코드도 내 저장소의 일부가 된다). 단, 플러그인이 붙인 파일은 **다음 `zbd update` 때 플러그인의 새 판으로 덮인다**(8절) — 직접 고치지 않는 것이 좋다.
3. 업무 채팅의 🔌 「중계 연결 만들기」 → Claude 를 돌릴 장비에서 `zbd install claude-integration`(Windows 는 `claude-install.bat`) — 묻는 칸에 관리자 주소(개발 PC 면 `http://localhost:5000`)·번호·비밀값·일할 폴더.

운영 서버에는 `zbd deploy`(7절)가 소켓 서버(systemd `{이름}-socket`)와 nginx `/ws` 연결까지 같이 올린다.
