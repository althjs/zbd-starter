/**
 * 글·게시판·댓글 — 서버 응답 모양(posts · post_comments · post_reactions · 메타 POST_GROUPS).
 * 관리자·사용자 화면이 같이 쓴다. json·jsonUser 는 풀어 둔 객체.
 */

/** 글의 json(관리자 칸)·jsonUser(사용자 칸) — 게시판 설정 dataColumn 이 어느 쪽에 쓰는지 정한다 */
export type PostJson = {
	/** 본문(HTML) */
	body?: string;
	/** 글쓴이 표시 이름 */
	name?: string;
	subTitle?: string;
	/** 대표 이미지 — imageId 폴더의 name 파일 */
	image?: { imageId?: string; name?: string; hostingId?: string } | null;
	files?: PostFile[];
	links?: PostLink[];
	/** 목록 위 고정 */
	isFixTop?: boolean;
	/** 글마다 켜는 댓글·별점·좋아요 */
	useComment?: boolean;
	useStarRating?: boolean;
	useLike?: boolean;
	isBest?: boolean;
	sortNo?: number;
	/** 관리자 메모(사용자에게 안 보임) */
	memo?: string;
	/** 검색엔진 설명 */
	seoDescription?: string;
	/** 처리 상태 코드 — 게시판 component.processMap 의 키 */
	processStatus?: number;
	/** 답글이면 위 글들의 postId(가장 위부터) */
	parentPostIds?: string[];
	schedule?: { from?: any; to?: any };
	/** 1:1 문의 관리자 답변들 */
	replies?: CsReply[];
	/** 관리자 답변(한 개만 쓰는 게시판) */
	adminReply?: string;
	viewCount?: number;
	commentCount?: number;
	/** 반응 수 — 반응 값 → 개수 */
	reaction?: Record<string, number>;
	/** 작성일을 고쳐 쓰는 게시판에서 실제로 쓴 시각(ms) */
	createdAtReal?: number;
	/** 업무 카드에서 만든 글이면 그 카드 */
	workId?: number;
	workGroupKey?: string;
	hostingId?: string;
	/** 답변 알림 받을 이메일(문의) */
	email?: string;
	/** 글을 쓴 관리자 화면 주소 — 답변 메일의 링크에 쓴다 */
	adminHost?: string;
	[k: string]: any;
};

/** 첨부 파일 — 보안 폴더 파일명 */
export type PostFile = { fileName: string; hostingId?: string; [k: string]: any };

/** 글에 붙인 링크 — isModal 이면 새 창 대신 모달로 */
export type PostLink = { id: number; url: string; target?: string; title?: string; isModal?: boolean; [k: string]: any };

/** 글 한 개 — 관리자(콘텐츠)·사용자(게시판) 가 같이 쓴다 */
export type Post = {
	/** 글 번호(8자 — Date.now().toString(36)) */
	postId: string;
	hostingId?: string;
	/** 관리자 글이면 쓴 관리자 */
	adminId?: number;
	/** 사용자 게시판은 번호·빈 값을 섞어 넘긴다(옛 사용자 타입이 any) */
	userId?: any;
	/** 게시판 key(POST_GROUPS) */
	groupKey: string;
	/** 하위 분류(게시판 items) — 라이프는 날짜, 앱 기록은 YYYYMM */
	groupSubKey: string;
	title: string;
	/** 관리자 칸 · 사용자 칸 */
	json: PostJson;
	jsonUser: PostJson;
	/** 0 공개 · 1 비공개 · -1 지움(숨김) · 99 내부용(설치 표시 등) */
	status: number;
	sortNo?: number;
	createdAt?: any;
	updatedAt?: any;
	/** 글쓴 회원 닉네임(응답에서 붙임) */
	nickname?: string;
	/** 답글 목록(묶어 보기) */
	replies?: Post[];
	/** 신고 상태 — isPostReported 등 */
	reportMap?: any;
	/** 화면에서만: 새 글(아직 저장 안 함) · 지움 표시 */
	isNew?: boolean;
	DELETED?: boolean;
	/** 사용자 게시판 — 글 번호·위 고정 순서 */
	postNo?: any;
	top?: number;
	[k: string]: any;
};

/** 게시판 하위 분류 */
export type PostGroupSub = { groupSubKey: string; name?: string; id?: number };

/** 게시판 설정 — 메타 POST_GROUPS 의 adminPostGroups·userPostGroups 한 항목 */
export type PostGroup = {
	groupKey: string;
	name?: string;
	/** 내용을 쓰는 칸 — 관리자 게시판 json · 사용자 게시판 jsonUser */
	dataColumn?: 'json' | 'jsonUser' | string;
	/** 기능 스위치(useComment·useWrite·useSitemap …) — 게시판 기능 설정 창 */
	feature?: Record<string, any> | null;
	/** 처리 상태(processMap 코드 → 이름) 등 화면 구성 */
	component?: { name?: string; title?: string; processMap?: Record<string, string>; condition?: string } | null;
	/** 글 상태 이름 바꾸기(status → 이름) */
	statusMap?: Record<string, string> | null;
	/** 하위 분류 */
	items?: PostGroupSub[] | null;
	/** 순서 정렬(sortNo) 사용 */
	useSort?: boolean;
	/** 공개 상태 칸 숨김 */
	disableOpenStatus?: boolean;
	/** 작성 후 수정 불가(masterGroupIds 관리자는 예외) */
	isEditDisabled?: boolean;
	masterGroupIds?: number[];
	/** 하위 분류를 글의 groupSubKey 값에서 만든다(설정 목록 대신) */
	isDynamicSubGroupKey?: boolean;
	/** 사용자 공개 게시판 */
	isBoardStore?: boolean;
	/** 관리자 답변 — 한 개 / 여러 개 / 답변하면 글쓴이에게 메일 */
	isAdminReply?: boolean;
	isAdminMultipleReply?: boolean;
	isAdminReplyEmailNoti?: boolean;
	useAttachFile?: boolean;
	/** 첨부 허용 형식(input accept) */
	fileAttachAccept?: string;
	/** 이 게시판 글을 다른 호스팅 id 로 쓸 때 */
	customHostingId?: string;
	/** 제목 앞에 붙는 「[이름]」 */
	partnerName?: string;
	/** 새 글 알림 받을 메일 · 알림 링크 앞부분 */
	notiEmail?: string;
	notiLinkPrefix?: string;
	[k: string]: any;
};

/** 1:1 문의 답변 한 개 — parentId 가 있으면 답변의 답 */
export type CsReply = {
	id: number;
	parentId?: number;
	adminId?: number;
	adminName?: string;
	body: string;
	ts?: number;
	files?: PostFile[];
	type?: string;
};

/** 게시판 글 목록의 조회수·반응·댓글 수(postId → 값) */
export type PostSummary = { pvMap?: Record<string, number>; reactionMap?: Record<string, any>; commentMap?: Record<string, number> } | null;

/** 댓글(post_comments) — 글에 hostingId + postId 로 붙는다 */
export type Comment = {
	commentId: number;
	postId: string;
	/** 대댓글이면 원 댓글 */
	parentCommentId?: number;
	userId?: number;
	userNickname?: string;
	/** 프로필 사진 파일명 */
	userImage?: string;
	comment: string;
	status?: number;
	/** 비밀 댓글(글쓴이·댓글쓴이만) */
	isSecret?: 0 | 1;
	isDeleted?: boolean;
	createdAt?: any;
	updatedAt?: any;
};

/** 반응(post_reactions) — reactionType L = 좋아요(reaction 100)·싫어요(99) / R = 별점(1~5). 신고는 'A' 로 저장된다(타입 밖) */
export type Reaction = { reactionId: number; postId: string; commentId: number; reactionType: 'L' | 'R'; reaction: number; userId: number; createdAt?: any; status?: number };
