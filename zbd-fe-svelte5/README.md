# zbd-fe-svelte5 — 즈비드 화면 (Svelte 5)

즈비드의 **사용자 화면(zbd.kr)과 관리자 화면(backyard.zbd.kr)을 한 앱**으로 그린다. 운영 = zbd.kr 서버 `/srv/zbd-web`.
꼭 지킬 규칙은 CLAUDE.md, 관리자 화면 부품·규칙은 `src/lib/admin/README.md`.

## 실행

```bash
npm install
npm run dev            # http://localhost:5000  — API 는 운영(zbd.kr) 으로 프록시
ZBD_API=http://127.0.0.1:5001 npm run dev   # 맥미니 Go BE 를 볼 때
npm run check          # 타입 검사(svelte-check) — 0 오류 유지
npm run build          # build/ (정적 SPA, index.html 폴백)
npm run format         # 코드 모양 정리(prettier) — VS Code 는 저장할 때 자동
```

- SvelteKit 3 · Svelte 5(runes) · Vite 8 · adapter-static. `ssr = false`(전부 브라우저에서 그린다).
- 검색엔진용 HTML(제목·설명·OG·구조화 데이터·`<noscript>` 본문)은 Go BE 가 배포된 `index.html` 의 자리표에 끼워 넣는다(CLAUDE.md 2).
- `npm install 패키지` 뒤 편집기에 타입 오류가 쏟아지면 `npm run prepare`(svelte-kit sync — `node_modules/$app` 을 다시 만든다).

## 배포

```bash
./deploy_zbd.sh             # 커밋된 HEAD 를 ../.wt-sv5-deploy 에 꺼내 prettier 검사 → 빌드 → 올리기 → index.html 바꿔 끼우기 → 홈 확인
./deploy_zbd.sh --built     # 이 폴더의 build/ 를 그대로
./deploy_zbd.sh --rollback  # 앞 index.html 로
```

- 커밋 안 한 수정은 올라가지 않는다. 정적 파일은 `static/assets/` 아래로만 낸다(nginx 는 `/assets/` 에만 파일 경로를 건다).
- 사이트 주소(사용자·관리자)·쿠키 범위는 빌드에 넣지 않는다 — 서버가 문서에 알려 준다(`<meta name="zbd:host-realm|admin-hosts|cookie-host">`, 원본 = 관리자 「환경 설정 › 도메인」).

## 구조 — 관리자 / 사용자 / 공용

```
src/
  app.html · app.css           바탕 — 서버가 채울 자리표(<!--zbd:head--> 등), 설치 앱 시작 화면, 기본 파비콘
  routes/
    +layout.svelte             공용: 전역 CSS, 알림창, 서비스 워커, 저장소 키 옮겨 담기
    (user)/                    사용자 화면(zbd.kr · talk.zbd.kr · {주소}.zbd.kr)
    (admin)/                   관리자 화면(backyard · chat · mail.zbd.kr) — 관리자 주소에서만 열린다
  lib/
    shared/                    둘이 같이 쓰는 것 — 여기 것은 user·admin 을 import 하지 않는다
      types/                   서버 데이터 모양(post·member·admin·design·widget·life·mini·mail·messenger·work·media·message·file·freecell·claude·socket)
      api.ts · config.ts · cookie.ts · site.svelte.ts · storage-keys.ts · diff.ts · push.ts · pull-refresh.ts
      ui/ · mui/ · media/ · image/ · socket/ · widget/ …
    user/                      사용자 전용 — session(로그인 상태·doLogin), 라이프·미니홈피·스토어·앱 빌더·프리셀·타자연습·친구 채팅
    admin/                     관리자 전용 — 틀·메뉴·권한·설정·디자인·콘텐츠·협업(메신저·웹메일·업무 채팅·업무관리)
```

- import 는 `#lib/...` 이고 **확장자를 붙인다**(`#lib/shared/api.ts`) — package.json `imports` 규칙.
- 서버 데이터 모양은 `lib/shared/types/{주제}.ts`, 화면 전용 타입(폼 상태·props)은 각 묶음에. 두 화면이 같이 쓰는 유틸·부품은 `lib/shared` 에 한 벌.
- 로그인은 서버가 심는 HttpOnly 세션 쿠키 — 화면은 토큰을 저장·읽지 않는다(`api.ts` 의 `isLoggedIn`·`endSession`). 화면 묶음이 `setRealm('user' | 'admin')` 으로 쪽을 정한다.
- 브라우저 저장소 키는 `zbd:{영역}:{이름}` — 전부 `lib/shared/storage-keys.ts` 의 `SK`.

## 스타일 규칙

1. **숫자를 화면마다 새로 적지 않는다** — 색·그림자·모서리·글꼴은 `lib/shared/styles/tokens.css` 의 변수.
2. 상자는 `<Card variant radius pad>`, 버튼은 `<Button variant tone pill icon>`.
3. 팝업·전체 화면 보기는 **`use:portal`** 로 body 에 그린다. 화면 안에 두면 미니홈피 글꼴(`webfonts.css` 의 `*` 규칙)을 물려받는다.
4. 글꼴 두 가지: 바탕 `--zbd-font`(Pretendard) · 부품 `--zbd-font-ui`(Roboto 계열). 미니홈피는 `.minihome :is(p,h5,h6,a,button)` 로 맞춘다.
5. 글자 간격: `p` 만 `--zbd-ls-body`(0.00938em), body·div 는 normal. 가짜 굵게 금지(`font-synthesis: none`).
6. Svelte 의 scoped style 을 기본으로, 자식 부품을 겨냥할 때만 `:global(...)` 을 좁게 쓴다.
7. 운영자가 써 둔 디자인 CSS(전역·GNB·페이지·위젯)가 지금 클래스(`.MuiStack-root` 등)에 기대므로, 디자인 화면의 DOM·클래스를 바꿀 때는 CLAUDE.md 5 의 두 단계로.

## 개발 로그인

- **관리자**: `http://localhost:5000/admin` → 관리자 로그인 화면에서 이메일·비밀번호로 그대로 로그인한다. 개발 서버는 요청에 `X-Token-Delivery: body` 를 붙여 토큰을 본문으로 받아 localStorage 에 둔다(운영은 HttpOnly 쿠키라 화면이 토큰을 못 본다).
- **회원**: 소셜 로그인은 운영 주소로만 돌아와 localhost 에서는 안 된다 → dev.zbd.kr(밖에서 보는 개발 서버)에서 zbd.kr 로그인을 넘겨받는다(`talk-handoff`). `/dev/login` 의 「복사 한 줄」 은 토큰이 쿠키로 바뀐 뒤(2026-10-03) 운영 화면에서 토큰을 읽을 수 없어 쓰지 못한다.
- 개발 서버의 `/admin`·`/login` 은 운영 사용자 문서 머리말을 받아 오지 않는다(운영 zbd.kr 은 `/admin` 을 홈으로 넘겨서, 받으러 가면 개발에서도 홈으로 튕겼다 — `hooks.server.ts`).
- ⚠ 기본 프록시는 운영 API — 개발 화면에서 하는 쓰기는 **운영 데이터**다.
