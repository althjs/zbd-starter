# CLAUDE.md — 즈비드 화면 (Svelte 5)

관리자(backyard)·사용자(zbd.kr) 두 React 앱을 **하나의 Svelte 5 앱**으로 합친다. 실행·구조·스타일 규칙은 README.md.

## 꼭 지킬 것 (2026-10-03 결정)

### 1. 관리자 화면은 관리자 주소에서만 연다 — 보안

빌드 결과가 하나라 막지 않으면 `zbd.kr/admin` 으로도 관리자 화면이 열린다. 로그인 토큰은 **출처(주소)별로** 저장되므로,
관리자가 zbd.kr 에서 로그인하면 관리자 토큰이 **회원이 만든 내용(앱·미니홈피)과 같은 출처**에 남는다 — 그 출처의 스크립트 사고 하나가 관리자 권한으로 번진다.

| 주소 | 여는 화면 |
|---|---|
| `backyard.zbd.kr` · `chat.zbd.kr` · `mail.zbd.kr` | `(admin)` 만 (chat·mail 은 각자 단독 앱 모드) |
| `zbd.kr` · `talk.zbd.kr` · `{주소}.zbd.kr`(미니홈피) | `(user)` 만 |

- **앱 안에서**: 루트 레이아웃(또는 hooks)이 지금 주소를 보고, 맞지 않는 묶음 경로면 그리지 않고 옳은 주소로 보낸다
  (사용자 주소에서 `/admin/…` → `https://backyard.zbd.kr/admin/…`, 관리자 주소에서 사용자 경로 → `/admin`).
- **nginx 에서도**: 사용자 주소 블록은 `/admin` 을 관리자 주소로 넘기고, 관리자 주소 블록은 사용자 경로를 SPA 로 넘기지 않는다. 앱 검사만 믿지 않는다(두 겹).
- 로그인은 서버의 HttpOnly 세션 쿠키(2026-10-03) — 화면은 토큰을 저장·읽지 않는다(`api.ts` isLoggedIn·endSession, 개발 서버만 localStorage). 관리자 쿠키는 관리자 주소에서만, 회원 쿠키는 사용자 주소에서만 서버가 받는다. 한 화면에서 두 쪽을 섞지 않는다(`setRealm`).
- 개발(localhost)은 한 출처라 위 검사를 끈다 — 설정값으로만(코드에 개발용 예외를 흩뿌리지 않는다).

### 2. 검색엔진용 HTML 은 서버가 「끼워 넣는다」 — 이 앱은 `index.html` 하나만 낸다

`ssr = false` · adapter-static 그대로 간다. 사용자 페이지의 제목·설명·OG·구조화 데이터·`<noscript>` 본문·첫 화면 데이터는
**백엔드(Go 로 옮기는 중)가 배포된 `index.html` 을 읽어 정해진 자리에 끼워 넣는다**. 예전처럼 서버가 빌드 청크 목록을 읽어 HTML 을 통째로 만들지 않는다.

- `src/app.html` 에 서버가 채울 **자리표를 두고 이름을 바꾸지 않는다**(바꾸면 서버도 같이 바꾼다):
  - `<head>` 안: 제목·설명·OG·canonical·robots·JSON-LD 를 넣을 자리 (예: `<!--zbd:head-->`)
  - `<body>` 안: `<noscript>` 본문 자리 (예: `<!--zbd:body-->`)
  - 첫 화면 데이터: `<!--zbd:data-->` → 서버가 **미리 받기 링크만** 넣는다(`<link rel=preload as=fetch id=zbd-base href=/service/design/ZBD/base?v=지문>` + 디자인 페이지면 `id=zbd-page`). 데이터 자체는 HTML 에 싣지 않는다(2026-10-03 대표 — 소스보기에 설정이 통째로 보였다). 앱은 `site.svelte.ts takePreload` 로 같은 주소를 받아 미리 받은 응답을 쓴다. base 는 Go `internal/sitepub` 허용 목록 — **화면이 새 칸을 쓰면 거기에 더한다**(빠지면 조용히 빈 값).
  - 자리표를 정하면 이 표에 적고, 백엔드 쪽에도 알린다.
- `index.html` 과 정적 파일은 **`/assets/` 아래로만** 낸다 — nginx 는 `/assets/` 에만 파일 경로가 걸려 있고 그 밖은 전부 SPA(index.html) 로 흐른다.
- 관리자 화면도 같은 `index.html` 을 쓴다(업무 카드 링크 미리보기가 이미 「배포된 index.html 에 제목·요약만 끼우는」 방식이다).

### 3. 응답 형식 — 새 화면은 `/api/v1` 의 새 형식을 쓴다 (대표 결정 10-03)

백엔드를 Go 로 바꾼 뒤 `/api/v1/…` 는 새 형식으로 낸다(옛 `/service/…` 는 React 화면이 사라질 때까지 옛 `code:"0000"` 형식 유지).

```jsonc
// 성공 — HTTP 200
{ "data": …, "meta": { "requestId": "…", "total": 120 } }
// 실패 — HTTP 4xx/5xx
{ "error": { "code": "POST_NOT_FOUND", "message": "글을 찾을 수 없습니다" }, "meta": { "requestId": "…" } }
```
- 성공·실패는 **HTTP 상태(`res.ok`)** 로 판단한다 — `lib/shared/api.ts` 의 `request()` 한 곳에서. 화면 코드에 `code == "0000"` 을 새로 쓰지 않는다.
- 실패 분기는 `error.code`(뜻이 읽히는 문자열) 로 한다. **message 문구로 분기하지 않는다**(문구는 바뀐다).
- 그 전까지(Go 전환 전) 는 `/service/…` 옛 형식을 `request()` 안에서 새 모양으로 바꿔 화면에 넘긴다 — 화면은 처음부터 새 모양만 본다.
- 이 앱이 해야 할 `request()` 보강: FormData 올리기 · blob 받기 · 401 이면 다시 로그인 · AbortController(화면을 떠나면 끊기).

### 4. 목록은 서버에서 나누고 찾는다 (대표 결정 10-03)

- 화면이 전체 목록을 받아 `filter`·`slice` 로 거르거나 쪽을 나누지 않는다. 쪽·검색·정렬·필터는 요청 값(`page`·`size`·`q`·`sort`·필터)으로 서버에 넘기고, 받은 쪽만 그린다(`meta.total` 로 쪽 수).
- 목록은 Go `/api/v1/…` 목록 API 를 쓴다(대상·진행은 `~/zbd-port/LIST_AUDIT.md`). React 판 API 가 전체를 주던 화면은 옮길 때 v1 목록이 생길 때까지 「v1 목록 대기」 로 표시해 둔다.

### 5. 디자인(페이지·GNB·위젯) — 레이아웃 하나 · 얕은 DOM (대표 결정 10-03)

- **GNB 는 늘 레이아웃 모드**다. 「레이아웃 사용 안 함」(옛 기본 모드 `items`) 의 그리기·편집 UI 는 옮기지 않는다(운영 GNB 13 은 이미 레이아웃 모드, 옛 items 는 데이터에만 남음).
- 디자인 화면의 DOM 은 **AI 가 읽기 쉬운 얕은 구조로 다시 짜도 된다**(지금은 MUI 감싸개가 겹겹이라 깊다). 단 운영자가 써 둔 CSS(전역·GNB·페이지·위젯)가 지금 클래스(.MuiStack-root 등)에 기대므로 **두 단계로** 간다:
  1. 옮기는 동안: 지금 DOM·클래스 그대로(React 판과 픽셀 비교) — 운영자 CSS 가 그대로 먹게.
  2. Svelte 판으로 바꿔 끼울 때: 새 DOM 약속(아래) + 운영자 CSS 일괄 변환(옛 선택자 → 새 선택자, DB 백업 뒤, 변환 전후 픽셀 비교) + 디자인 MCP 의 선택자 안내(get_widget_selectors·domHint) 갱신을 **한 번에**.
- 새 DOM 약속(초안 — 확정 전 대표 확인): 위젯 하나 = `<section class="zw zw-{종류} {위젯 className}" data-widget-id data-widget-type>` 아래 내용 2~3단, GNB = `<header class="zg"> <div class="zg-row" data-row="{이름}"> <div class="zg-col">`, 의미 있는 클래스 이름(zw-title·zw-list·zw-item…)만, MUI 클래스 없음. 약속 문서는 `docs/design-dom.md`(만들 것) 한 곳.

## 함께 알아 둘 것

- **코드 모양은 prettier**(`.prettierrc`) — VS Code 는 저장할 때 자동, 손으로는 `npm run format`. 배포 스크립트가 `prettier --check` 를 통과해야 올린다. `npm install 패키지` 뒤 편집기 타입 오류가 나면 `npm run prepare`(svelte-kit sync — `node_modules/$app` 을 다시 만든다).

- **API 는 고정이다.** 백엔드를 Go 로 한 번에 옮기는 중이고 응답은 지금과 글자 단위로 같게 맞춘다. 화면은 지금처럼 `/service/…` 를 부른다.
  응답 모양을 바꿔야 하면 혼자 바꾸지 말고 백엔드 작업의 「합친 뒤 할 일」(맥미니 `~/zbd-port/AFTER_MERGE.md`) 에 적어 같이 바꾼다.
- **새 서버 데이터 타입(응답·저장 모양)은 `lib/shared/types/{주제}.ts` 에**(2026-10-03 — 글·회원·관리자·디자인·앱·라이프·미니홈피·메일·메신저·캔버스·미디어…). 관리자·사용자가 여기서 import 하고, 같은 모양을 화면마다 다시 적지 않는다. 화면 전용 타입(폼 상태·표 칸·props)은 각 묶음에.
  관리자·사용자 둘 다 쓰는 유틸·화면 조각도 사본을 만들지 말고 `lib/shared` 에 한 벌(예: `diff.ts`·`ui/DiffPreview.svelte` — 10-03 두 벌을 합침). 관리자 웹 푸시 상태·켜기는 `lib/admin/collab/push/admin-push.ts` 한 곳.
- **앱 격리 실행(widget-sandbox — iframe 에 `allow-same-origin` 절대 금지)과 소켓 클라이언트는 `lib/shared` 에 한 벌만.** 관리자 미리보기와 사용자 실행이 같은 코드를 쓴다.
- **브라우저 저장소 키 = `zbd:{영역}:{이름}`**(번호가 붙으면 `:{번호}`) — 전부 `src/lib/shared/storage-keys.ts` 의 `SK` 한 곳(2026-10-03). 화면 코드에 문자열 키를 쓰지 않는다.
  옛 키는 같은 파일의 `migrateStorage()` 가 옮긴다(표시 「옛 localStorage 키」 — 몇 주 뒤 지운다). ⚠ BE 소셜 로그인 콜백은 아직 옛 `__user_token` 에 쓰므로 토큰만은 옛 값이 새 값을 덮는다.
- 서비스 워커는 하나로 합친다(푸시·알림 클릭). 설치 앱 정보(manifest) 는 서버가 접속 주소별로 준다.
- React 두 레포는 Go 전환이 끝난 뒤 내린다(지금 Node BE 타입체크가 React 관리자 레포의 types 폴더에 기댄다).
- 바꿔 끼우기: 화면 묶음별로 옮겨 시험 주소에 나란히 띄워 픽셀 비교 → 다 되면 주소를 한 번에 돌린다.
