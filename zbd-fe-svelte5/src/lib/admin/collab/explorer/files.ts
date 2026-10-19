/**
 * 웹하드(파일 탐색기) API — React explorer/zustandMarkdownStore. `/service/admin/file-explorer/{루트}`.
 * 저장은 불러온 내용의 지문(baseHash) 과 같이 — 그사이 다른 곳에서 바뀌었으면 서버가 code 4090 + 지금 내용으로 거절한다.
 */
import { api } from '#lib/shared/api.ts';
import type { FILE_NODE, WEBHARD_ACL_INFO, WEBHARD_ACL_RULE } from '#lib/shared/types/file.ts';

const base = (root: string) => `/service/admin/file-explorer/${encodeURIComponent(root)}`;
const q = (filePath: string) => `?filePath=${encodeURIComponent(filePath)}`;

export async function getFileTree(root: string): Promise<FILE_NODE> {
	const o = await api.get(base(root));
	const os = o.ok ? o.data : [];
	if (Array.isArray(os)) return { name: root, path: root, type: 'directory', items: os, size: 0, updatedAt: new Date() };
	return os || { name: root, path: root, type: 'directory', items: [], size: 0, updatedAt: new Date() };
}
export async function getTextFile(root: string, filePath: string) {
	const o = await api.get(base(root) + q(filePath));
	return { text: o.ok ? o.data : null, meta: o.meta };
}
export const isImageFile = (n: string) => /\.(png|jpe?g|gif|webp|svg|bmp)$/i.test(n || '');
export const isTextFile = (n: string) => /\.(md|txt|json|js|jsx|ts|tsx|css|html|xml|yml|yaml|csv)$/i.test(n || '');
/** 글 파일이면 글, 그림·그 밖은 Blob */
export async function getFile(root: string, filePath: string): Promise<any> {
	if (isTextFile(filePath)) {
		const o = await api.get(base(root) + q(filePath));
		return o.ok ? o.data : '';
	}
	const o = await api.blob(base(root) + q(filePath));
	return o.ok ? o.data : '';
}
/** 저장 — 결과를 옛 모양(code·data·message)으로 돌려준다: 4090(그사이 바뀜)이면 data = 지금 내용 */
export async function updateFile(root: string, filePath: string, text: string, baseHash?: string): Promise<{ ok: boolean; conflict?: boolean; data?: any; message?: string }> {
	const o = await api.put(base(root) + q(filePath), { text, ...(baseHash ? { baseHash } : {}) });
	if (o.ok) return { ok: true, data: o.data };
	// 4090 응답의 data(지금 내용)는 request() 가 실패 때 버린다(shared 요청) — 지금 내용을 다시 읽는다
	if (o.error.code == '4090') return { ok: false, conflict: true, data: (await getTextFile(root, filePath)).text, message: o.error.message };
	return { ok: false, message: o.error.message };
}
export const mv = (root: string, fromPath: string, toPath: string) => api.put(`${base(root)}/move`, { fromPath, toPath });
export const mkdir = (root: string, dirPath: string) => api.post(base(root) + q(dirPath), { isFolder: true });
export const mkfile = (root: string, filePath: string) => api.post(base(root) + q(filePath), {});
/** 내용을 담아 새 파일 만들기(같은 이름이 있으면 덮는다 — 이름은 부르는 쪽이 고른다) */
export const writeNewFile = (root: string, filePath: string, content: string) => api.post(base(root) + q(filePath), { content });
export const rm = (root: string, filePath: string) => api.delete(base(root) + q(filePath));
export function uploadFiles(root: string, folderPath: string, files: File[]) {
	const fd = new FormData();
	files.forEach((f) => fd.append('files[]', f));
	return api.post(`${base(root)}/upload?folderPath=${encodeURIComponent(folderPath || '')}`, fd);
}
/** 폴더·파일 권한 — 스타터 기본은 권한이 없어 늘 읽기·쓰기(「업무 웹하드」 플러그인이 이 파일을 서버에 묻는 판으로 바꾼다) */
export async function getAcl(_root: string, _path = ''): Promise<WEBHARD_ACL_INFO> {
	return { canEdit: false, access: 'RW', hide: false };
}
export const putAcl = async (_root: string, _path: string, _rule: WEBHARD_ACL_RULE | null) => ({ ok: false });
