/** 업무 보드(업무관리) — `work_items` 행과 COMMON_CODES 보드 설정 모양 */

/** 공통코드 항목 — 보드 분류(대·중·소)·상태 목록·보드 설정이 이 모양으로 COMMON_CODES 에 있다 */
export type CodeItem = {
	id: number;
	name: string;
	group?: string;
	key?: string;
	value?: any;
	/** 글자색 · 바탕색(상태 칩 등) */
	color?: string;
	bgColor?: string;
	/** 하위 항목(분류 단계) */
	items?: CodeItem[];
	inputType?: string;
	/** 이 분류의 기본 담당 */
	adminIds?: number[];
	/** 이 분류 카드에서 보일 칸 */
	workFields?: string[];
	/** 항목 종류(SEO·CANVAS 등) */
	component?: string;
	[k: string]: any;
};

/** 카드 분류 — 대·중·소 CodeItem.id (0 = 없음) */
export type Category = { c1: number; c2: number; c3: number };

/** 일정 — 날짜 문자열(YYYY-MM-DD) */
export type Schedule = { from: string; to: string };

/** 멘션·알림 받은 사람 — notiAt = 알림 보낸 시각(ms) */
export type NotiMember = { adminId: number; name: string; profileImage?: string; notiAt?: number };

/** 첨부 — 보안 폴더(`secure/…`) 파일, 받기는 늘 API 로 */
export type SecureFile = { secureImageId?: string; fileName?: string; [k: string]: any };

/**
 * 코멘트·서브태스크(shape) — 카드 Json.shapes 또는 이미지별 images[].shapes.
 * 처음엔 그림 위 도형이라 color·strokeWidth 가 필수로 남아 있다(코멘트는 `#000000`·1).
 */
export type Shape = {
	/** 만든 시각 문자열(Date.now()) */
	id: string;
	/** markdown(코멘트) · no-shape(답글) · 옛 도형 종류 */
	type: string;
	color: string;
	strokeWidth: number;
	title?: string;
	/** 본문 — 이름은 html 이지만 **마크다운 원문** */
	html?: string;
	/** 쓴 사람 · 마지막에 고친 사람 */
	adminId?: number;
	lastEditAdminId?: number;
	/** 지금 편집 중인 사람 이름(다른 화면에 「편집 중」 표시) */
	activeEditorName?: string | null;
	isDeleted?: boolean;
	/** 다른 카드를 가리키는 코멘트 — 누르면 그 카드로(옛 값 · 하나) */
	targetWorkId?: number;
	/** 연관 업무 여러 개(2026-10-04~, 있으면 targetWorkId 대신) */
	targetWorkIds?: number[];
	/** 질문 표시(목록 「질문만」 거르기) */
	isQuestion?: boolean;
	/** 속한 이미지 번호 — images[i].shapes 이면 i */
	subImageIndex?: number;
	mentionedMembers?: NotiMember[];
	/** 쓴 시각(ms) */
	ts?: number;
	/** 답글이면 원 코멘트 id */
	parentShapeId?: number;
	likeAdminIds?: number[];
	/** 서브태스크로 쓰기 — status·schedule 을 갖는다 */
	useStatus?: boolean;
	status?: any;
	schedule?: Schedule | null;
	[k: string]: any;
};

/** 카드에 붙인 이미지 한 장 — 이미지마다 설명·코멘트를 따로 갖는다 */
export type WorkImage = {
	name: string;
	title: string;
	desc?: string;
	descActiveEditorName?: string;
	shapes: Shape[];
	mentionedMembers?: NotiMember[];
	ts?: number;
	files?: SecureFile[];
	likeAdminIds?: number[];
};

/** 카드 한 장(work_items) — 고치기는 늘 API(`PUT /service/admin/work`, 소켓 알림·권한 검사 때문에) */
export type WorkItem = {
	workId: number;
	/** 만든 관리자(API 를 부른 계정) — 담당은 Json.adminIds */
	adminId: number;
	hostingId: string;
	/** 보드 이름(TASK 등) */
	groupKey: string;
	/** 하위 카드면 부모 카드 */
	parentWorkId?: number;
	createdAt: string;
	updatedAt: string;
	/** 보내기 전용: true 면 updatedAt 을 바꾸지 않는다(목록 순서 유지 — 편집 중 표시 등) */
	ignoreUpdatedAt?: boolean;
	Json: {
		title?: string;
		/** 본문(마크다운) */
		desc?: string;
		/** 본문을 편집 중인 사람 이름 */
		descActiveEditorName?: string;
		/** 대표 이미지 파일명 */
		image?: string;
		shapes?: Shape[];
		category?: Category;
		/** 상태 — 보드 상태 목록의 id(TASK: 1 백로그 · 5 대기 · 10 진행중 · 30 완료). 없는 값은 「미설정」 */
		status?: number | '';
		schedule?: Schedule;
		mentionedMembers?: NotiMember[];
		images?: WorkImage[];
		/** 연결한 다른 카드 */
		targetWorkIds?: number[];
		/** 관련 관리자 → 알림 시각(ms) — 목록의 「담당·관련자」 거르기가 이 키를 본다 */
		adminNotiMap?: Record<number, number>;
		files?: SecureFile[];
		likeAdminIds?: number[];
		[k: string]: any;
	};
};

/** 보드 설정(보드마다) — 쓸 수 있는 사람과 끌 기능 */
export type WorkSetting = {
	/** 글을 쓸 수 있는 관리자·권한 그룹(비면 모두) */
	writeableAdminIds?: number[];
	writeableGroupIds?: number[];
	/** 처음 열 때 검색 조건 */
	defaultSearchOption?: WorkSearch;
	disableSearchForm?: boolean;
	disableWork?: boolean;
	disableText?: boolean;
	disableMarkdown?: boolean;
	disableFileAttach?: boolean;
	disableLike?: boolean;
	disableShapeLike?: boolean;
	disableShapeReply?: boolean;
	disableShapeStatus?: boolean;
	disableWrite?: boolean;
	disableDetail?: boolean;
	[k: string]: any;
};

/** 보드 목록 검색 조건(화면 상태 · 보드 기본값) */
export type WorkSearch = {
	category: Category;
	/** 질문 코멘트가 있는 카드만 */
	isQuestion: boolean;
	/** 대분류별로 묶어 보기(목록 보기에서만) — 철자는 저장된 키라 그대로 */
	isGroupping: boolean;
	/** 하위 카드도 보기 */
	showChild?: boolean;
	/** '' 목록 · CANVAN 칸반(철자 그대로) · CALENDAR 달력 */
	viewMode: '' | 'CANVAN' | 'CALENDAR';
	/** 담당·관련자 거르기 */
	adminIds: number[];
	schedule?: Schedule;
	/** 최근 N일 안에 바뀐 카드만 */
	days?: number;
	/** 보일 상태 — 보기 방식별로 따로 기억 */
	statusLabels?: string[];
	statusLabelsByViewMode?: Record<string, string[]>;
};
