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
};
