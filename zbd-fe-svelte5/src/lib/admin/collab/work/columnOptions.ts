/**
 * 표 열 보기 옵션 — React ModalTableColumnOption 의 저장 규칙 그대로.
 *   이 브라우저: localStorage `zbd:admin:tableColumn:{optionKey}` · 호스팅 기본값: COMMON_CODES `TABLE_COLUMN_OPTIONS` 의 key == optionKey
 */
import { getMeta } from '../../meta.svelte.ts';
import { SK } from '#lib/shared/storage-keys.ts';

export type ColumnOption = { key: string; name: string; use?: boolean; sticky?: boolean };

function defaults(): any[] {
	const out: any[] = [];
	const seen = new Set<string>();
	for (const c of (getMeta('COMMON_CODES')?.items || []).filter((x: any) => x.group == 'TABLE_COLUMN_OPTIONS'))
		for (const it of c.items || []) {
			if (!it?.key || seen.has(it.key)) continue;
			seen.add(it.key);
			out.push(it);
		}
	return out;
}
// 옛 이름 canvas — 2026-10 옮기는 기간: 업무관리 표의 키 work_{보드} 가 호스팅 설정에 아직 canvas_{보드} 로 있으면 그것을 쓴다(배포 뒤 설정 키를 바꾼다)
const legacyKey = (k: string) => (k.startsWith('work_') ? 'canvas_' + k.slice(5) : '');
const findDefault = (k: string) => defaults().find((x) => x.key == k) || (legacyKey(k) ? defaults().find((x) => x.key == legacyKey(k)) : null);

export function getColumnOptionsDefault(optionKey: string): ColumnOption[] | null {
	const v = findDefault(optionKey)?.value;
	return v ? JSON.parse(JSON.stringify(v)) : null;
}
export function getColumnOptions(optionKey: string): ColumnOption[] {
	try {
		const s = localStorage.getItem(SK.admin.tableColumn(optionKey));
		if (s) return JSON.parse(s) || [];
	} catch {}
	return findDefault(optionKey)?.value || [];
}
export function saveColumnOptions(optionKey: string, v: ColumnOption[] | null) {
	try {
		if (v) localStorage.setItem(SK.admin.tableColumn(optionKey), JSON.stringify(v));
		else localStorage.removeItem(SK.admin.tableColumn(optionKey));
	} catch {}
}
