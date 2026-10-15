/**
 * 관리자 바탕 화면 타입(왼쪽 메뉴·지금 메뉴 위치) — React 판 types/*.d.ts(전역) 에서 옮겼다.
 * 서버 데이터 모양(프로필·권한 메뉴 등)은 lib/shared/types/ 에 둔다. 여기는 화면용만.
 */

/** 왼쪽 메뉴 항목(화면용) */
export type NavItem = {
	menuId?: number;
	title: string;
	desc?: string;
	auth?: string;
	path?: string;
	special?: string;
	icon?: string;
	items?: NavItem[];
	active?: boolean;
	disabled?: boolean;
	external?: boolean;
	/** 축약 메뉴 팝업용 */
	miniIcon?: string;
	miniDepth?: number;
};

/** 지금 화면의 메뉴 위치 — 페이지 제목·이동 경로·권한(AuthButton) */
export type CurrentNav = {
	title: string;
	desc?: string;
	auth?: string;
	path?: string;
	sub?: NavItem | null;
	subDepthTitle?: string | null;
};
