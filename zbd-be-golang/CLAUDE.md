# CLAUDE.md — 즈비드 BE (Go)

즈비드 서비스의 새 백엔드. 실행·설정·폴더는 README.md.

## 원칙

- **완전히 새 구현이다.** 예전 회사·플랫폼의 이름·도메인·코드를 들이지 않는다(코드·주석·문서·커밋 모두).
  지금 운영 중인 서비스는 「지금 서비스(Node BE)」 라고만 부른다.
- **한 번에 전환한다**(2026-10-03 결정). 그 전까지 옮긴 API 는 **지금 서비스와 응답이 글자 단위로 같아야** 한다 — 화면을 고치지 않고 바꿔 끼운다.
  같은 요청을 두 서버에 보내 비교한다(맥미니 비교 환경 — 아래).
- 즈비드 한 회사만 다룬다. 다른 회사·사이트빌더 분기는 옮기지 않는다(데이터도 지웠다).
- 비밀값(DB 비밀번호·토큰 키·API 키)은 `.env.local`·서버 환경변수에만. 코드·로그·커밋·응답에 넣지 않는다.
- 공유 메모리는 잠근다(`sync.Mutex`·`sync.Map`·`atomic`). 테스트는 `go test -race ./...`.

## 구조

| 위치 | 내용 |
|---|---|
| `cmd/server` | 시작 · 기능별 라우트 등록(`routes.go`) |
| `internal/app` | 프로세스 하나가 공유하는 것 — 설정·DB 풀·설정 캐시·관리자 권한 캐시·개인정보 키·소켓, 1분 하트비트 일(`EveryMinute`) |
| `internal/web` | 경로·입구. `Router.Handle("/admin/design/[designId]", web.Auth, h)` — 지금 서비스와 같은 경로 모양, `/service`·`/api/v1`·`/api` 세 주소로 받는다 |
| `internal/api/<기능>` | 옮긴 API — content·apps·core·admincontent·collab·site·mcp·pay |
| `internal/jsx` | JS 와 같은 값·JSON(키 순서 보존, 숫자 키는 앞·오름차순, 숫자 표기) |
| `internal/lib/db` | MariaDB — 값 모양을 지금 서비스 드라이버와 같게, `IN (?)` 배열·`SET ?` 객체 펼치기, OkPacket, 트랜잭션 |
| `internal/lib/…` | 여러 기능이 같이 쓰는 것: auth(토큰) · pii(개인정보 암호화) · settings · adminauth · member · push · mailsend · sock · util |

## 입구 (지금 서비스의 감싸기와 같다)

- `web.Plain` — 감싸지 않은 라우트(CORS·요청 번호 없음)
- `web.Cors` — withCors: CORS + 요청 번호 + 응답 `meta.requestId`
- `web.Auth` — withAuth: Cors + 회원 토큰은 헤더 `zbduid`(쿠키 `zbd_uid` — `c.BrowserUID()`) 와 브라우저 id 가 같아야 + `/user/…` 가 아닌 주소는 관리자만, 관리자는 메뉴에 묶인 API 면 그 메뉴 권한(GET 은 R, 그 밖은 RW)
- 토큰은 13칸 미만이거나 memberType 이 A·U 가 아니면 해석하지 않는다(위조·옛 토큰).

## 핸들러 쓰는 법

```go
func h(c *web.Ctx) {
	id := c.Q("postId")            // req.query.postId (경로 변수 포함)
	bd := c.BodyMap()              // req.body
	rows, err := db.Query(c.Context(), c.App.DB, `SELECT … WHERE id IN (?)`, ids)
	if err != nil { c.Ecb(err); return }
	c.Scb(rows)                    // res.send(scb(rows))
}
```
- 응답 객체는 `jsx.Obj(...)`/`*jsx.OrderedMap` — 키 순서가 응답에 그대로 나간다. JSON 칸은 `jsx.ParseOr`/`jsx.ParseMap`.
- 한 요청에서 DB 를 여러 번 쓰면 `c.App.DB.Conn(ctx)` 로 연결 하나를 잡고 `defer conn.Close()`.
- 회원 이름·이메일·전화는 암호문이다 — 읽으면 `c.App.PII.DecMember`, 쓰면 `PII.Enc` + `PII.Hash`(찾기 값 칸 `*_hash`).

## 전환 때 할 일(잊지 말 것)

- 서버 `.env`: DB_HOST·DB_NAME(ZBD)·DB_USER(zbd_app)… · AES_256_ENCRYPTION_* · AUTH_COOKIE_KEY · PII_* · VAPID_* · RUN_JOBS=true · 그 밖 기능별 키
- 서버 `.env` 2차: FE_INDEX_HTML·FE_ADMIN_INDEX_HTML·CDN_URL_PREFIX·MCP_UPLOAD_TOKEN·PAY_*·PG_INICIS_*·PG_KSNET_*
- 운영 DDL(대표 승인 2026-10-03): `ALTER TABLE payments ADD COLUMN keepId INT UNSIGNED NULL AFTER pgId, ADD UNIQUE KEY uk_payments_keep (keepId)` — 되돌리기 `DROP INDEX uk_payments_keep, DROP COLUMN keepId`
- nginx: seo·robots·/admin/work OG 를 Go 포트로(Host 헤더 유지)

## 전환 뒤 — 응답 형식 v1 (대표 결정 2026-10-03)

- `/api/v1/…` 는 새 형식: 성공 HTTP 200 `{data, meta:{requestId, …}}` · 실패 HTTP 4xx/5xx `{error:{code, message}, meta:{requestId}}`. 오류 코드는 뜻이 읽히는 문자열(상수 한 곳).
- `/service/…` 는 React 화면이 사라질 때까지 옛 형식(`code:"0000"`/`"9999"`) 그대로 — 전환 전 비교 검증도 이 형식으로.
- 바꿀 곳: `internal/web/ctx.go` 의 `Scb`·`Ecb`(주소 앞부분으로 판 구분) + 기능 패키지의 비공개 사본(scb*/ecb*).

## 목록 API = 서버에서 나누고 찾는다 (대표 결정 2026-10-03)

- **전체 목록을 내려 화면이 거르거나 나누는 API 를 만들지 않는다.** 쪽 나누기·검색·정렬·필터는 모두 서버(SQL)에서.
- 새 목록 API 는 `/api/v1/…` 새 형식으로: 요청 `?page=1&size=20&q=검색어&sort=createdAt,-title&{필터}=…`(size 최대 100, 기본 20) →
  응답 `{data: [...], meta: {requestId, total, page, size}}`. 검색은 색인이 있는 칸만(필요하면 인덱스·전문 검색 추가 — DDL 은 대표 확인).
- `/service/…` 옛 목록은 React 화면이 사라질 때까지 지금 모양 그대로(비교 검증 대상). Svelte 판은 v1 목록만 쓴다.
- 대상 목록은 감사표 `~/zbd-port/LIST_AUDIT.md`.
- 코드: `internal/api/v1/` — `v1x`(입구·쪽·검색 도우미) · `admin` · `me`(본인 것만) · `pub`(공개). 입구는 `web.Cors` 로 등록하고 `v1x.Admin(원래경로, h)`·`v1x.User`·`v1x.Public` 으로 감싼다(/api/v1 이 아니면 404).
  관리자 권한은 **원래 API 경로의 메뉴 권한**으로 본다(새 경로는 admin_menu_apis 연결이 없다). 정렬은 `v1x.List(…, tie)` — 끝에 기본키를 붙여 쪽 사이 순서가 흔들리지 않게.
  회원 검색은 `v1x.MemberSearch`(이메일·전화·실명 정확 일치, 닉네임 앞부분). 공개 화면은 닉네임만(이메일로 가입 여부를 알아낼 수 없게).
- 인덱스 제안 `~/zbd-port/LIST_DDL.sql`(실행 전 대표 확인). MariaDB 는 `WITH` 안의 `JSON_TABLE` 이 「Table doesn't exist in engine」 오류 — 파생 표로 쓴다.

