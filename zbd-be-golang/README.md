# 즈비드 BE (Go)

즈비드 서비스의 백엔드. **HTTP API 를 한 프로세스**에서 처리한다(설정 캐시·DB 풀이 한 벌). 운영 = zbd.kr 서버 `/srv/zbd-be-go`, pm2 `zbd-be`(포트 4011).
규칙·구조는 CLAUDE.md.

## 실행

```bash
cp .env.example .env.local   # 값을 채운다(.env* 는 git 에 올리지 않는다)
make dev     # 저장하면 다시 빌드·재시작 (air 필요: go install github.com/air-verse/air@latest)
make run     # 한 번 실행
make build   # bin/server
```

기본 포트 **5001**. 상태 확인 `GET /api/v1/healthz`. 코드 모양은 gofmt(VS Code 는 저장할 때 자동).

## 배포 · 로그

```bash
./deploy_zbd.sh             # gofmt·vet 검사 → linux/arm64 빌드 → 올리기 → 바꿔 끼우고 재시작 → healthz 확인
./deploy_zbd.sh --rollback  # 바로 앞 실행 파일로
```

- 작업 폴더의 코드를 그대로 빌드한다 — **커밋한 뒤** 돌린다.
- 서버 설정 `/srv/zbd-be-go/.env`(600) 는 손으로 둔다. `KEY="값" # 설명` 처럼 값 뒤에 주석을 달지 말 것.
- 로그: `ssh zbd.kr pm2 logs zbd-be`(실시간) · `--lines 500 --nostream`(최근만). 파일은 `~/.pm2/logs/zbd-be-{out,error}.log`. 처리 중 오류는 `@@` 로 시작한다.

## 설정 — 환경변수

`.env` → `.env.local`(덮어씀) → 프로세스 환경변수(가장 우선). 항목은 `.env.example`.

| 항목 | 뜻 |
|---|---|
| `PORT` | 기본 5001 |
| `DB_HOST` · `DB_PORT` · `DB_NAME` · `DB_USER` · `DB_PASSWORD` · `DB_CHARSET` · `DB_CONN_LIMIT` | MariaDB(`ZBD`). 풀은 프로세스에 하나 |
| `AES_256_ENCRYPTION_KEY` · `AES_256_ENCRYPTION_IV` | 로그인 토큰·티켓 암호화 |
| `PII_ENC_KEY_V1` · `PII_INDEX_KEY` | 회원 개인정보(이름·이메일·전화) 암호화·찾기 값 — 잃으면 못 읽는다 |
| ~~`ADMIN_HOSTS`~~ | (옛 설정) 관리자 주소는 이제 DB 사이트 설정 `domains`(관리자 「환경 설정 › 도메인」) — 이 값은 domains 가 없을 때만 읽는다 |
| `RUN_JOBS` | `true` 면 1분 주기 작업(알림 예약·정리 등)을 이 프로세스가 돈다 |
| `ES_HOST` · `ES_PASSWORD` · `ES_TAG_INDEX` | Elasticsearch(글 검색·웹로그) |
| `FE_INDEX_HTML` | 화면 index.html(SEO 를 끼워 넣어 돌려준다) |
| 그 밖 | 메일·푸시(VAPID)·소켓·PG·오라클·AI 키 — `.env.example` |

## API 주소

- 같은 핸들러를 `/service/…` · `/api/v1/…` · `/api/…` 로 받는다. `/service` 는 `{code:"0000", data}` 형식, `/api/v1` 은 HTTP 상태 + `{data, meta}` / `{error}` 형식.
- 라우트는 `internal/api/<기능>/` 의 `Register` 에서 등록한다(`/admin/design/[designId]` 같은 경로 모양).
- 로그인 = 대표 도메인 HttpOnly 세션 쿠키(`internal/web/session.go`). 스크립트는 `Authorization: Bearer` 헤더도 쓸 수 있다.

## DB 표 (`ZBD`, 2026-10-03 기준 행 수)

표는 복수형 의미 이름, 인덱스는 `ix_/uq_/ft_{표}_{뜻}`. 호스팅 칸은 `hostingId`(`ZBD`, 공통 `""`). 회원 개인정보 칸은 암호문 + `*_hash`(찾기 값).
「처음 쓸 때 생성」 표는 코드가 `CREATE TABLE IF NOT EXISTS` 로 만든다.

### 회원 · 로그인

| 표 | 행 | 설명 |
|---|---:|---|
| `members` | 41 | 사용자 회원. 이름·이메일·전화는 암호문(`name_hash`·`email_hash`·`mobile_hash` 로 찾는다), 설정·미니홈피는 `json` |
| `login_devices` | 136 | 로그인한 브라우저(회원·관리자) — 토큰이 묶인 브라우저 id 확인 |
| `verification_codes` | 52 | 이메일·휴대폰 인증 코드(비밀번호 찾기 등) |
| `push_subscriptions` | 4 | 웹 푸시 구독(기기·출처별) |
| `notifications` | 84 | 앱 안 알림함 |
| `visit_events` | 357 | 방문 흐름 이벤트(방문·프리셀 첫 수·도전장 등, 90일 보관) |

### 관리자 · 권한

| 표 | 행 | 설명 |
|---|---:|---|
| `admins` | 6 | 관리자 계정(OTP 는 암호문, 조직·AI 계정 정보는 `json`) |
| `admin_roles` | 4 | 관리자 권한 그룹 |
| `admin_menus` | 67 | 관리자 왼쪽 메뉴(트리, 특수 메뉴 `special`) |
| `admin_role_menus` | 114 | 그룹별 메뉴 권한(R·RW) |
| `admin_menu_apis` | 33 | 메뉴에 묶인 관리자 API — 그 메뉴 권한이 있어야 부를 수 있다 |
| `admin_token_logs` | 4 | 브라우저에 묶이지 않은 관리자 토큰(ANY_BROWSER) 발급 기록 |
| `org_units` | 0 | 조직도(회사·부서·팀) — 처음 쓸 때 생성 |

### 사이트 · 디자인 · 설정

| 표 | 행 | 설명 |
|---|---:|---|
| `settings` | 10 | 호스팅별 설정(`SITE`·`HOSTINGS`·`COMMON`·`POST_GROUPS`·공통코드 등) — JSON 한 덩어리씩 |
| `designs` | 18 | 디자인 페이지·GNB·푸터(위젯 JSON) |
| `design_page_templates` | 0 | 디자인 페이지 템플릿(관리자 템플릿 관리) |
| `design_widget_templates` | 0 | 디자인 위젯 템플릿 |

### 글 · 댓글 · 미디어

| 표 | 행 | 설명 |
|---|---:|---|
| `posts` | 924 | 모든 글 — 게시판·라이프 기록·앱(위젯)·앱 데이터·프리셀 기록·도전·리뷰 등(`groupKey` 로 구분) |
| `post_comments` | 50 | 게시판 댓글(Mroonga 엔진 — 한글 전문 검색, 다른 DB 로 RENAME 불가) |
| `post_reactions` | 54 | 글 좋아요 등 반응 |
| `post_views` | 296 | 글 조회 기록 |
| `share_links` | 16 | 라이프 기록 공유 링크 |
| `videos` | 42 | 관리자 동영상(삭제 = 숨김 status -1) |
| `youtube_videos` | 307 | 유튜브 영상(수집) |
| `youtube_comments` | 24,072 | 유튜브 댓글(수집) |
| `typing_rounds` | 33,164 | 타자연습 한 판 한 줄 |

### 미니홈피 · 앱

| 표 | 행 | 설명 |
|---|---:|---|
| `friendships` | 1 | 미니홈피 친구(승인제) |
| `minihome_comments` | 12 | 미니홈피 메모·앱 카드 댓글(주인·친구만) |
| `app_reminders` | 2 | 앱 알림 예약(`zbd.notify`, KST 다음 시각) |

### 메신저 · 업무

| 표 | 행 | 설명 |
|---|---:|---|
| `chat_rooms` | 8 | 메신저 방(관리자 사내 메신저 + 회원 친구 채팅 — 호스팅 칸으로 공간을 가른다) |
| `chat_members` | 16 | 방 구성원·읽은 위치·알림 끔 |
| `chat_messages` | 303 | 메시지(글타래 `threadId`, 반응·첨부·링크 미리보기는 `json`) |
| `chat_presence` | 0 | 「보는 중」 표시(45초 유효 — 그 사람에게 알림을 안 보낸다) |
| `support_messages` | 9 | 고객 상담 1:1 |
| `work_items` | 84 | 업무관리 보드 카드(코멘트·첨부는 `Json`) |

### 메일 · 메시지 발송

| 표 | 행 | 설명 |
|---|---:|---|
| `mail_accounts` | 3 | 관리자 ↔ 웹메일 계정 연결 · 개인 메일 규칙 · 항상 보기 발신자 |
| `mail_large_files` | 2 | 대용량 첨부(30일 내려받기 링크) |
| `message_logs` | 4 | 메일·SMS·알림톡 발송 기록(자체 메일 서버·NHN) |
| `kakao_templates` | 0 | 카카오 알림톡 템플릿(아직 안 씀 — 남겨 둠) |

### 결제 · 플러그인

| 표 | 행 | 설명 |
|---|---:|---|
| `payments` | 0 | PG 결제(아직 없음 — 남겨 둠) |
| `payment_requests` | 0 | PG 결제 요청(결제창 열기 전) |
| `plugin_tokens` | 3 | `zbd` 명령(플러그인 스토어) 토큰 |
| `plugin_licenses` | 2 | 플러그인 구독(체험·구독·정지·만료) |
| `plugin_installs` | 0 | 장비별 설치·검증 기록 |
