/**
 * 콘텐츠 묶음 화면 타입(게시판 검색 조건) — 글·게시판 같은 서버 데이터 모양은 lib/shared/types/post.ts.
 */

export type BoardSearch = {
	groupKey: string;
	groupSubKey: string;
	orderBy: string;
	orderByDesc: 'asc' | 'desc';
	keyword: string;
	keywordType?: string;
	nickname?: string;
	pageSize: number;
	pageNo: number;
	from: string;
	to: string;
	status?: number | null;
	scheduleStatus?: string;
	[k: string]: any;
};
