/** 웹하드(파일 탐색기) — API `admin/file-explorer/{root}` 응답 모양. 저장 위치는 서버 `/srv/cdn/file-explorer/ZBD/{root}/` */

/** 폴더 또는 파일 하나(폴더는 items 로 아래 목록을 갖는다) */
export type FILE_NODE = {
	/** 파일·폴더 이름 */
	name: string;
	/** 루트 기준 경로(예 `즈비드 호스팅/02_제품구성/04.md`) — 읽기·쓰기 API 의 filePath 로 그대로 쓴다 */
	path: string;
	type: 'directory' | 'file';
	/** 폴더일 때 아래 항목들(트리로 받을 때만) */
	items?: FILE_NODE[];
	/** 바이트(폴더는 0) */
	size: number;
	/** 마지막 수정 시각 — 서버가 주는 그대로(ms 숫자 또는 날짜 문자열) */
	updatedAt: any;
	/** 내 권한(폴더 권한 규칙이 있을 때만) — RW 읽기·쓰기 · R 읽기만 · N 권한 없음(보이지만 열 수 없음) */
	acc?: WEBHARD_ACCESS;
	/** 이 경로에 권한 규칙이 걸려 있음(권한 관리자에게만) */
	acl?: boolean;
};

export type WEBHARD_ACCESS = 'RW' | 'R' | 'N';
/** 웹하드 폴더·파일 권한 규칙 — groups 에 없는 권한 그룹 = others */
export type WEBHARD_ACL_RULE = { groups: Record<string, WEBHARD_ACCESS>; others: WEBHARD_ACCESS; hide: boolean };
/** GET …/acl 응답 — rules·groups 는 권한 관리자(canEdit)에게만 */
export type WEBHARD_ACL_INFO = {
	canEdit: boolean;
	access: WEBHARD_ACCESS;
	hide: boolean;
	rules?: (WEBHARD_ACL_RULE & { path: string; updatedAt?: number; by?: string })[];
	groups?: { groupId: string; name: string }[];
};
