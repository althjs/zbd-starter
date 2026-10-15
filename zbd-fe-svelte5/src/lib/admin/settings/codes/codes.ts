/**
 * 공통 코드(분류) 저장·찾기 — React components/common/codes/{CodeCategory.doSaveMeta, constants, statusLabel}.
 * 저장은 늘 「다시 읽고 → 그 항목만 바꿔 → 통째로 PUT」(다른 사람이 고쳤을 수 있어서).
 */
import { api } from '#lib/shared/api.ts';
import { alert } from '../../ui/modal.svelte.ts';

const q = (hostingId: string) => (hostingId ? `?hostingId=${hostingId}` : '');

/** 1단계 코드 하나를 저장(새로면 끝에 더함, _DELETED 면 뺌) — 새 메타를 돌려준다 */
export async function doSaveMeta(hostingId: string, meta: any, item: any): Promise<{ o: any; m: any } | null> {
	const g = await api.get(`/service/admin/meta/${meta.metaKey}${q(hostingId)}`);
	if (!g.ok) {
		alert('데이터를 다시 불러오는 도중 오류가 발생하였습니다.');
		return null;
	}
	const json = g.data;
	let isNew = true;
	json.items = (json.items || []).reduce((arr: any[], o: any) => {
		if (o.id == item.id) {
			isNew = false;
			if (!item._DELETED) arr.push(item);
		} else arr.push(o);
		return arr;
	}, []);
	if (isNew) json.items.push(item);
	const p = await api.put(`/service/admin/meta/${meta.metaKey}${q(hostingId)}`, json);
	if (!p.ok) alert('저장 도중 오류가 발생하였습니다.');
	return { o: item, m: { ...meta, json: { ...json } } };
}

/** 1단계 순서 저장 — React 와 같이 다시 읽을 때는 hostingId 를 붙이지 않는다 */
export async function doSortMeta(hostingId: string, meta: any, os: any[]): Promise<any | null> {
	const g = await api.get(`/service/admin/meta/${meta.metaKey}`);
	if (!g.ok) {
		alert('데이터를 다시 불러오는 도중 오류가 발생하였습니다.');
		return null;
	}
	const json = g.data;
	const items = os.reduce((arr: any[], o: any, sortNo: number) => {
		const oo = json.items.find((x: any) => x.id == o.id);
		if (oo) arr.push({ ...oo, sortNo: sortNo + 1 });
		return arr;
	}, []);
	if (items.length !== json.items.length) for (const o of json.items) if (!items.find((x: any) => x.id == o.id)) items.push(o);
	json.items = items;
	const p = await api.put(`/service/admin/meta/${meta.metaKey}${q(hostingId)}`, json);
	if (!p.ok) alert('저장 도중 오류가 발생하였습니다.');
	return { ...meta, json: { ...json } };
}

export function getParentCode(parentIds: any[], meta: any): any | null {
	let items: any[] = meta?.json?.items || [];
	let found: any = null;
	for (const id of parentIds) {
		found = items.find((x) => x.id === id) || null;
		if (!found) return null;
		items = found.items || [];
	}
	return found;
}

/** 1단계에서 최고 관리자만 보는 코드(React READ_ONLY_CODE_IDS) */
export const READ_ONLY_CODE_IDS = [17, 34, 10];

/** 그 1단계 코드가 허용하는 깊이 */
export const maxDepthOf = (metaKey: string, base: any) => (metaKey == 'CATEGORY_DOMESTIC' ? base?.maxDepth || 3 : base?.maxDepth >= 3 ? base.maxDepth : 2);

export function nextLabelId(list: any[]): number {
	return list.length > 0 ? Math.max(...list.map((c) => Number(c.id) || 0)) + 1 : 1;
}
export function newStatusLabel(list: any[]) {
	return { id: nextLabelId(list), name: '', bgColor: '#f5f5f5', color: 'var(--zc-000, #000)' };
}
/** 상태 라벨 색 테마 — 배경/글자 짝 */
export const STATUS_LABEL_THEMES: { bg: string; fg: string }[] = [
	{ bg: '#f1f5f9', fg: '#334155' },
	{ bg: '#e2e8f0', fg: '#1e293b' },
	{ bg: '#94a3b8', fg: '#ffffff' },
	{ bg: '#475569', fg: '#ffffff' },
	{ bg: '#1e293b', fg: '#ffffff' },
	{ bg: '#fee2e2', fg: '#991b1b' },
	{ bg: '#fca5a5', fg: '#7f1d1d' },
	{ bg: '#ef4444', fg: '#ffffff' },
	{ bg: '#b91c1c', fg: '#ffffff' },
	{ bg: '#ffedd5', fg: '#9a3412' },
	{ bg: '#fdba74', fg: '#7c2d12' },
	{ bg: '#f97316', fg: '#ffffff' },
	{ bg: '#fef3c7', fg: '#92400e' },
	{ bg: '#fcd34d', fg: '#78350f' },
	{ bg: '#eab308', fg: '#422006' },
	{ bg: '#ecfccb', fg: '#3f6212' },
	{ bg: '#bef264', fg: '#365314' },
	{ bg: '#84cc16', fg: '#1a2e05' },
	{ bg: '#dcfce7', fg: '#166534' },
	{ bg: '#86efac', fg: '#14532d' },
	{ bg: '#22c55e', fg: '#ffffff' },
	{ bg: '#ccfbf1', fg: '#115e59' },
	{ bg: '#5eead4', fg: '#134e4a' },
	{ bg: '#14b8a6', fg: '#ffffff' },
	{ bg: '#dbeafe', fg: '#1e40af' },
	{ bg: '#93c5fd', fg: '#1e3a8a' },
	{ bg: '#3b82f6', fg: '#ffffff' },
	{ bg: '#ede9fe', fg: '#5b21b6' },
	{ bg: '#c4b5fd', fg: '#4c1d95' },
	{ bg: '#a855f7', fg: '#ffffff' },
	{ bg: '#fce7f3', fg: '#9d174d' },
	{ bg: '#f9a8d4', fg: '#831843' },
	{ bg: '#ec4899', fg: '#ffffff' }
];
