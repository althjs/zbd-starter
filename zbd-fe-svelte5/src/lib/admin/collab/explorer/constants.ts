/** 문서 보기 모드 기억 (2026-10-03) — 「편집」 으로 열 때 마지막에 본 모드(viewMode 값, "" = 좌우 나란히) 로 연다. 이 브라우저에만 */
import { SK } from '#lib/shared/storage-keys.ts';
const LAST_VIEW_MODE_KEY = SK.explorer.mdViewMode;
export function saveLastViewMode(vm: string) {
	try {
		localStorage.setItem(LAST_VIEW_MODE_KEY, vm);
	} catch {}
}
/** 마크다운 파일이면 마지막 모드를 붙인 주소 */
export function withLastViewMode(url: string) {
	if (!/\.md$/i.test(url)) return url;
	let vm = '';
	try {
		vm = localStorage.getItem(LAST_VIEW_MODE_KEY) || '';
	} catch {}
	return /^[123]$/.test(vm) ? `${url}?viewMode=${vm}` : url;
}
