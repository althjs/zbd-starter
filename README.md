# zbd-starter

자기 서비스를 **Claude 와 함께** 만들고 운영하기 위한 스타터. 이 폴더를 받아 Claude Code 워크스페이스로 열면 바로 시작할 수 있다.

| 폴더 | 내용 |
|---|---|
| `zbd-be-golang/` | 백엔드 — Go, HTTP API 한 프로세스(인증·권한·설정·글·파일·SEO 문서) |
| `zbd-fe-svelte5/` | 화면 — SvelteKit(Svelte 5). 사용자 사이트와 관리자(`/admin`)가 한 앱 |
| `db/` | MariaDB 마이그레이션(번호 붙은 SQL + 되돌리기) · 기본 데이터 |
| `run_dev.sh` | 개발 모드 — 도구 확인·설치 → DB 준비 → BE·FE 함께 실행 |
| `run_dev.bat` | Windows 판(더블클릭) — 실제 일은 `run_dev.ps1`, Go·MariaDB 는 관리자 권한 없이 `%LOCALAPPDATA%\zbd-tools` 에 |
| `reset.sh` | 처음 상태로 — 개발 DB·`.dev/` 지우기(`--all` 은 `.env.local` 까지) |
| `CLAUDE.md` | 규칙집 — Claude 가 이 저장소에서 일하는 방식 |

## 시작

> 새로 시작하는 분은 **[docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)** 부터(받기 → 띄우기 → 관리자 → 첫 화면).

```bash
git clone <이 저장소> my-service && cd my-service
./run_dev.sh
```

처음 실행하면 알아서:

1. **도구 확인** — Go · Node.js(22 이상) · MariaDB · air(Go 자동 재시작). 없으면 설치한다(macOS 는 Homebrew, Linux 는 apt).
   Homebrew 자체는 관리자 비밀번호가 필요해 대신 설치하지 않는다 — https://brew.sh
2. **MariaDB 켜기** — macOS 는 `brew services start mariadb`(로그인할 때 자동으로 켜지는 서비스로 등록된다).
3. **설정 만들기** — `zbd-be-golang/.env.local` 에 DB 접속 정보와 암호화 키를 무작위로 만든다.
4. **DB 준비** — `db/setup.sh`: DB `zbd_starter`·계정 `zbd_dev`(이 DB 에만 권한) → 마이그레이션 → 기본 데이터.
5. **첫 관리자·시험 회원** — 관리자 `admin@example.com` · 회원 `user@example.com`, 비밀번호는 무작위로 `.dev/account.txt` 에.
6. **실행** — BE http://127.0.0.1:5001 · FE http://localhost:5000 (관리자 http://localhost:5000/admin). 로그는 `[be]`·`[fe]` 머리를 달고 한 화면에.

**Ctrl+C** 로 둘 다 끈다. 포트는 `zbd-be-golang/.env.local` 의 `PORT`(서버)·`FE_PORT`(화면)·`SOCK_PORT`(소켓 데몬)로 바꾼다 — 한 번만은 `BE_PORT=… FE_PORT=… ./run_dev.sh`(docs/GETTING_STARTED.md 「포트 바꾸기」).

**Windows** 는 `run_dev.bat` 을 더블클릭 — 같은 일을 한다(서버는 따로 뜨는 `[be]` 창, DB 는 `.dev\mariadb` 에 두고 뒤에서 계속 돈다).

> ⚠ `zbd-be-golang/.env.local` 의 `PII_ENC_KEY_V1`·`PII_INDEX_KEY` 를 잃으면 회원 이름·이메일·전화를 다시 읽을 수 없다. git 에는 올라가지 않는다 — 따로 보관할 것.

## DB

```bash
./db/setup.sh            # 빠진 마이그레이션·기본 데이터만 적용(여러 번 돌려도 같다)
./db/setup.sh --status   # 적용된 마이그레이션
```

- 새 표·칸은 `db/migrations/NNNN_이름.up.sql` + 되돌리기 `NNNN_이름.down.sql` 로 더한다. 「처음 쓸 때 CREATE TABLE」 은 쓰지 않는다.
- 시간은 UTC 로 저장하고 화면에서만 지역 시간으로 바꾼다.

## 라이선스

MIT — `LICENSE`. 「즈비드」 이름·로고는 상표라 이 라이선스에 들어가지 않는다(갈라져 나가면 이름을 바꿀 것).
