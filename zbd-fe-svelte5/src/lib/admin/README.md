# lib/admin — 관리자 화면(backyard) 바탕

관리자 화면의 틀·메뉴·권한·공용 부품. 화면을 새로 만들 때 **여기 것을 쓰고, 같은 부품을 새로 만들지 않는다.**

## 라우트 규칙

| 주소 | 파일 |
| --- | --- |
| `/admin/xxx` | `src/routes/(admin)/admin/xxx/+page.svelte` |
| `/admin/xxx/:id` | `src/routes/(admin)/admin/xxx/[id]/+page.svelte` (`page.params.id`) |
| `/login/admin-login` | `src/routes/(admin)/login/admin-login/+page.svelte` |

- `+page.svelte` 는 얇게 두고 본문은 `lib/admin/<묶음>/…` 에(예: `lib/admin/dashboard/Dashboard.svelte`).
- 화면 틀: `<PageContainer>` 로 감싼다 — 제목·설명은 지금 메뉴(권한 메뉴 API)에서 자동. `title`/`subtitle` 을 주면 그것.
- 묶음 레이아웃 `(admin)/+layout.svelte`: 관리자 주소 검사(HOST_REALM — 서버가 「도메인」 설정으로 알려 줌, 개발 한 주소는 끔) · `setRealm('admin', 401→로그인)` · `<html class="zbd-admin">` + 테마 · `<AdminDialogs/>` · 휴대폰 뒤로가기.
  `(admin)/admin/+layout.svelte`: `loadAdmin()`(메뉴·프로필) → 틀(AdminShell) · 관리자 소켓 시작 · 사이트 설정 파비콘. 단독 앱(chat/mail.zbd.kr)·`?popup=1` 은 메뉴 없는 틀.
- 권한 없는 주소로 들어오면 첫 메뉴로 보낸다(`authCheckAndRedirect`) — 예외: /admin/my, /admin/work-invite/, 개발 서버의 /admin/dev-ui.
- 개발 서버 `/admin/dev-ui` = 공용 부품 견본(모두 한 화면에).

## 꼭 지킬 것 (실측으로 정한 것)

1. **관리자 CSS 는 `:where(html.zbd-admin) .Mui…`**(styles/admin.css) — 구체성이 클래스 하나라 화면의 `sx`(zx- 클래스, 나중에 붙음)가 이긴다. `html.zbd-admin .X` 로 쓰면 sx 를 이겨 버린다(링크 색·입력칸 여백이 안 바뀌었다).
2. **MUI 이름표 `<button>` 을 직접 그리는 부품에는 `<style>` 을 두지 않는다.** Svelte 가 svelte-xxx 클래스를 붙이면 공용 app.css 의 `button[class*='svelte-']` 초기화(0,1,1)가
   sx·MUI 여백·바탕을 이긴다. 그런 부품의 모양은 `styles/parts.css`(adm-* 이름). 부품 컴포넌트(`<Button>`·`<IconButton>`)를 쓰면 상관없다.
3. **아이콘은 `ui/Icon`(머티리얼 심볼 Outlined)** · `ui/SvgIcon`(@mui/icons-material 경로). 공용 shared/mui/Icon 은 Material Icons(채움) — 모양이 다르다.
4. **알림·확인 창·토스트는 `ui/modal.svelte.ts`**(alert/confirm/prompt/modal/toast). 공용 shared/ui/dialog 는 사용자 화면용.
5. 색은 변수: 대표색 `--theme-color-main|dark|light|lightest|alpha8…` · `--adm-text|text2|bg|paper|card|divider|hover|action` · `--adm-n50…n900`(neutral) ·
   `--adm-success-main|error-main|warning-main|info-main|secondary-main` · 공용 회색 `var(--gt-200, #f2f2f2)` · 코드 색 `var(--zc-fff, #fff)`(다크에서 바뀜). 화면에 hex 를 새로 적지 말 것.
6. sx 의 반응형은 미디어 쿼리 키로: `{ padding: 16, '@media (min-width:600px)': { padding: 24 } }` — 관리자 구간 sm 600 · md 900 · lg 1280 · xl 1440. 코드에서 폭은 `media.mdDown`/`media.lgUp`(media.svelte.ts).
7. MUI Stack 의 `spacing` 은 gap 이 아니라 **둘째 자식부터 margin**(TopNav `.adm-sp2`) — 배치가 다를 수 있으니 계산값을 재고 맞춘다.
8. API 는 `api`(shared/api.ts) — 응답은 `{ok,data,meta}` / `{ok:false,error}`. `code=="0000"` 금지.
9. 서버 데이터 타입은 `lib/shared/types`, 관리자 화면 전용 타입은 `lib/admin/types.ts` + 묶음별 `types.ts`.

## 상태·도우미

| 파일                    | 쓰는 법                                                                                                                                                                                                                                      |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `profile.svelte.ts`     | `admin.profile` · `admin.menus` · `admin.notification` · `checkAuth(path)` · `updateProfile` · `logout()` · `getFirstAuthMenuPath`                                                                                       |
| `meta.svelte.ts`        | `getMeta('SITE','feature','useESPost')` · `feature(name)` · `getHostingId()` · `isSuperAdmin(profile)` · `adminImageUrl(folder, name)`(= /upload/{호스팅}/폴더/파일)                                                                         |
| `nav.svelte.ts`         | `nav.current`(지금 메뉴) · `activeNav()`(권한 auth) · `setSubDepthTitle` · 메뉴 펼침 상태                                                                                                                                                    |
| `socket.svelte.ts`      | `onSocketDataUpdate('WEB_HDD', (d, isSelf) => …)` → 정리 함수(`$effect` 에서 return) · `onAdminSocket('onWorkUpdate', fn)` · `adminSocket()?.emit(…)` · `socketState.connected` — 기본은 내 브라우저가 일으킨 신호를 건너뜀(`{self:true}`) |
| `dark.svelte.ts`        | `dark.resolved`('light'·'dark') · `setDarkMode('auto')`                                                                                                                                                                                      |
| `backToClose.svelte.ts` | 창(서랍·전체 화면 Dialog)은 자동. 창이 아닌 덮는 화면만 `useBackToClose(() => open && media.mdDown, () => (open = false))`                                                                                                                   |
| `media.svelte.ts`       | `media.smDown · mdDown · lgUp · xlUp`                                                                                                                                                                                                        |
| `format.ts`             | `shortDate()` — 그 밖(date·number·timeAgo) 은 shared/format.ts                                                                                                                                                                               |
| `config.ts`             | ADMIN_HOSTS · LS 키(이름은 storage-keys.ts) · SINGLE_APP · 메뉴 폭                                                                                                                                                                               |
| `store/mail.svelte.ts`  | `mailUnread`(안 읽은 수) · `mailFolders`(웹메일 폴더 — 왼쪽 메뉴 「웹메일」 하위)                                                                                                                                                                 |

## 공용 부품 `ui/`

| 부품 | 쓰는 법 |
| --- | --- |
| `ui/modal.svelte.ts` + `ui/AdminDialogs.svelte` | alert/confirm/prompt/modal(child 스니펫) · `toast()` — 문자열 = 성공(초록), 객체 type 없으면 오류(빨강) |
| `layout/PageContainer.svelte` | 화면 틀 — 이동 경로·북마크·제목·도움말(TIP) |
| 그냥 `<a href>` | 화면 이동(SvelteKit 이 가로챈다) |
| `ui/SortDialog.svelte` · `ui/SortList.svelte` | 순서 바꾸기 — 끌기 + ▲▼⤒ |
| `ui/TablePagination.svelte` + `ui/pageSize.ts` | 쪽 나누기 |
| `ui/DataTable.svelte` | 표 — 정렬·선택·줄 누르기·빈 목록·번호(#) |
| 공용 `shared/mui/Select.svelte` | 고르기(outlined 작은 상자 — 관리자 모양은 admin.css 가 덮음) |
| `ui/TextField.svelte` | 입력칸 — 라벨 위 고정·hiddenLabel·여러 줄·helperText·error·start/end |
| `ui/AuthButton.svelte` | 지금 메뉴 권한이 RW 일 때만 보이는 버튼 |
| `layout/Loading.svelte` | 불러오는 중 |
| `ui/TagInput.svelte` | 태그 입력(값은 "a,b,c") |
| `ui/StateChip.svelte` | 상태 칩(isAutoColor) |
| `ui/DateRangeModal.svelte` · `ui/DateRange.svelte` | 기간 고르기 창 · 날짜 칸 + 빠른 버튼 — 여러 달 달력·빠른 칩·키보드·직접 입력 |
| `ui/AdminDrawer.svelte` | 오른쪽 편집 서랍(폭 규칙은 admin.css — 900 미만 꽉, 이상 최소 min(560px,100vw)) |
| `ui/Dialog.svelte` | 창 — maxWidth·fullWidth·mobileFullScreen·actions 스니펫 |
| `ui/Tabs.svelte` | 탭(count 표시) |
| `ui/Tooltip.svelte` · `ui/Popover.svelte` | 말풍선 · 팝오버 |
| `ui/Button.svelte` · `ui/IconButton.svelte` · `ui/Card.svelte` | 버튼(color·fullWidth) · 아이콘 버튼(href) · 카드 |
| `ui/SearchBar.svelte` | 검색 줄 — filters 스니펫 + 검색어 + 버튼 |
| 공용 `shared/image/ImagePicker.svelte` | 이미지 고르기·올리기 |
| 공용 `shared/editor/TinyMceEditor.svelte` | 글 편집기 |
| `design/ui/JsonEditorDrawer.svelte` · `settings/ui/JsonEditDialog.svelte` | JSON 직접 고치기 |
| `collab/ui/MarkdownView.svelte` | 마크다운 보기 |
| `window/` | 화면 위 창 — `openRouteWindow(path, title)` · `openWindow({id,title,component,props,width,height})`(아래) |

## 확인

`npm run check`(0 오류) · `npm run build` · `npm run format:check`.

## 화면 위 창 (`window/`)

- 여러 창 · 제목줄 끌기 · 8방향 크기 조절(최소 600×500) · 누르면 맨 앞 · 최소화(→ 아래 오른쪽 줄 `#minimized-dock`, 오른쪽 가운데 창 목록) · 최대화(제목줄 두 번 누르기) · Ctrl+Shift+화살표 반쪽 붙이기.
- **경로 창** `openRouteWindow('/admin/…', '메뉴 이름')` — 같은 출처 iframe(`경로?popup=1`). 창 id = `win_`+base64(경로) → 같은 경로는 같은 창(복원 + 맨 앞).
  iframe 안은 `data-zbd-win` 으로 알아본다(`window/frame.ts IN_WIN_FRAME`) — 링크로 옮겨 다녀도 메뉴 없는 틀, 제목은 그 주소의 메뉴 이름. 창 안의 `100vh` 는 창 높이다.
- **부품 창** `openWindow({ id, title, component, props })` — Svelte 부품을 직접(props 에 `windowId` 가 더해진다). 제목줄 버튼을 더하려면 `setWindowActions(windowId, [{icon,title,onclick}])`.
  iframe 안에서 연 부품 창은 그 iframe 안에 뜬다(다른 문서의 부품은 맨 위 문서에서 못 그린다). 경로 창은 맨 위 문서로 넘긴다.
- 자리·크기 = localStorage `zbd:admin:window:{창 id}`, 최소화 줄 표시 = `zbd:admin:showMinimizedDock`(키는 #lib/shared/storage-keys.ts). 「모든 창 닫기」 는 기억한 자리도 지운다.
- z-index: 창 1250 + 순번(최대 1299 — 메뉴 바 1100·서랍 1200 위, 팝오버·Dialog 1300 아래), 최소화 줄 2000, 창 목록 2100(말풍선 2200).
- 렌더러 `<WindowHost/>` 는 AdminShell·PopupLayout 에 한 번씩. 휴대폰(<900)은 창이 화면 전체, 뒤로가기 = 닫기.
- 최고 관리자는 관리자 링크를 **Shift+누르기** = 그 화면을 창으로. 연결: 왼쪽 메뉴 창 아이콘 · 머리줄 AI 업무채팅(`window.__openClaudeChat`) · 이미지 편집 도구 · Claude 파일 미리보기 · 보드 검색 Shift · 글 보기 회원(👤).
