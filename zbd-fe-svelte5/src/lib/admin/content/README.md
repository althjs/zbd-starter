# lib/admin/content — 관리자 콘텐츠 묶음

게시물(post)·게시판(board)·고객센터(cs)·동영상(media/video)·유튜브(youtube)·메시지(message)·내 정보(my).
라우트는 `routes/(admin)/admin/<경로>/+page.svelte`(얇게), 본문은 여기.

| 폴더 | 무엇 |
| --- | --- |
| `post/` | 게시판 설정 읽기(`postGroup.ts`)·글 상태(`postStore.svelte.ts`)·편집 서랍(`PostManageDrawer` — `inline` 이면 페이지 안)·본문 보기·다중 답변 |
| `board/` | 게시판 목록(`BoardGroup` — 고객센터도 이걸)·검색 칸·표·글 보기·글 편집 페이지·사용자 글/댓글·신고 |
| `cs/` | 고객센터 목록·글(답변 창·답변 메일) |
| `media/` | 동영상 목록·편집 서랍·올리기 |
| `youtube/` | 영상·댓글(서버 쪽 나눔, 영상 고르기는 서버 검색) |
| `message/` | 이메일·SMS 발송, 대상 고르기, 템플릿, 알림톡 템플릿(`kakao/`) |
| `my/` | 마이페이지 — 계정·푸시(`PushDevices` 는 메신저 알림 설정에서도 쓴다)·인증 기기·2단계 인증 |
| `ui/` | 이 묶음의 범용 부품: `ScrollTabs`(넘치면 ‹ ›, 고른 탭 자동 스크롤) · `DateField` · `OutlinedInput`(앞·뒤 자리) · `SwitchField` · `JsonEditDialog` · `EsMarkTitle` · `table.ts`(표 칸 바로잡기) |

- ⚠ 서버 목록 API 대기(CLAUDE.md 4): `BoardGroup`(비ES 게시판)·`PostManageHome`·`ReportedPosts`·일반 회원 고르기는 지금 API 가 전체를 줘서 화면이 거른다 — `/api/v1` 목록 API 가 생기면 검색 값을 그대로 요청에 싣는다.
- `Typography` 의 `color` 는 팔레트 이름(primary·error)만 쓴다 — 점 경로(text.secondary)는 먹지 않는다.
- 표 몸통 칸은 `ui/table.ts bodyCellFix` 를 붙인다(줄 높이 1.5 · 휴대폰 td 여백). 화면이 td 여백을 따로 주면 `'&&&& td.MuiTableCell-root'`.
