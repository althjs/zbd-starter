/**
 * 업무 보드(업무관리) 코드·설정 — React pages/admin/work/constatns.ts 와 같은 규칙.
 * 보드마다 COMMON_CODES 의 `group == {groupKey}` 항목 하나: 그 아래 대분류(c1) 들 + `group == "COLUMNS"` (열 · STATUS · SETTINGS).
 */
import { getMeta, isSuperAdmin } from '../../meta.svelte.ts';
import type { AdminProfile } from '#lib/shared/types/admin.ts';
import type { CodeItem, Category, WorkItem, WorkSetting, WorkSearch } from '#lib/shared/types/work.ts';

export const workSearchDefault: WorkSearch = {
	category: { c1: 0, c2: 0, c3: 0 },
	isQuestion: false,
	isGroupping: true,
	showChild: true,
	viewMode: '',
	adminIds: [],
	schedule: { from: '', to: '' }
};

const board = (groupKey: string): CodeItem | undefined => (getMeta('COMMON_CODES')?.items || []).find((x: CodeItem) => x.group == groupKey);
const columnsOf = (groupKey: string): CodeItem[] => board(groupKey)?.items?.find((o) => o.group == 'COLUMNS')?.items || [];

/** 대분류(c1) 목록 */
export function getWorkCodes(groupKey: string): CodeItem[] {
	return (board(groupKey)?.items || []).filter((o) => o.group != 'COLUMNS');
}
/** [보드 코드 id, COLUMNS 코드 id] — 라벨 고치기 경로 */
export function getWorkCodeIds(groupKey: string): number[] {
	const ids: number[] = [];
	const b = board(groupKey);
	if (!b) return ids;
	ids.push(b.id);
	const c = b.items?.find((x) => x.group == 'COLUMNS');
	if (c) ids.push(c.id);
	return ids;
}
/** 보드가 정한 열 — STATUS·SETTINGS 는 빼고(includeKeys 에 든 key 는 넣음) */
export function getWorkColumns(groupKey: string, includeKeys: string[] = [], excludeKeys: string[] = []): CodeItem[] {
	return columnsOf(groupKey).filter((o) => !excludeKeys.includes(o.key || '') && (includeKeys.includes(o.key || '') || (o.group !== 'STATUS' && o.group != 'SETTINGS')));
}
export function getWorkSettings(groupKey: string): WorkSetting | null {
	const s = columnsOf(groupKey).find((o) => o.group === 'SETTINGS');
	return s ? s.value || {} : null;
}
export function isWorkSharedFromMaster(groupKey: string) {
	return !!(board(groupKey) as any)?._sharedFromMaster;
}
export function isWorkWriteable(groupKey: string, profile: Partial<AdminProfile> | null) {
	if (isSuperAdmin(profile)) return true;
	let tf = true;
	const s = getWorkSettings(groupKey);
	if (s?.writeableAdminIds?.length && !s.writeableAdminIds.includes(Number(profile?.adminId))) tf = false;
	if (s?.writeableGroupIds?.length && !s.writeableGroupIds.includes(Number(profile?.groupId))) tf = false;
	else if (!tf && s?.writeableGroupIds?.includes(Number(profile?.groupId))) tf = true;
	return tf;
}
export function getWorkColumnCode(groupKey: string, codeId: number) {
	return columnsOf(groupKey).find((o) => o.id === codeId) || null;
}

const GT700 = 'var(--gt-700, #909090)';
const DEFAULT_STATUS_MAP: Record<string, string> = { '': '미설정', 10: '리뷰중', 100: '리뷰완료' };
const DEFAULT_STATUS_CODES: CodeItem[] = [
	{ id: 0, name: '미설정', color: '#fff', bgColor: GT700 },
	{ id: 10, name: '리뷰중', color: '#fff', bgColor: '#fbcb00' },
	{ id: 100, name: '리뷰완료', color: '#fff', bgColor: '#9cd326' }
];
export function getWorkStatusMap(groupKey: string): Record<string, string> {
	const st = columnsOf(groupKey).find((o) => o.group === 'STATUS');
	if (!st) return DEFAULT_STATUS_MAP;
	const m: Record<string, string> = {};
	(st.items || []).forEach((x) => (m[typeof x.value == 'string' ? x.value : x.id] = x.name));
	return m;
}
export function getWorkStatusCodes(groupKey: string): { codeId: number; codes: CodeItem[] } {
	const st = columnsOf(groupKey).find((o) => o.group === 'STATUS');
	return st ? { codeId: st.id, codes: st.items || [] } : { codeId: 0, codes: [...DEFAULT_STATUS_CODES] };
}
/** 상태 이름 — 라벨 다중 필터가 쓴다(보드가 달라도 같은 이름이면 같이 묶이게) */
export function getWorkItemStatusLabel(item: WorkItem): string {
	const v = item?.Json?.status;
	if (v == null || v === '') return '';
	const { codes } = getWorkStatusCodes(item.groupKey || '');
	const c = codes.find((c) => c.id === v || String(c.id) === String(v) || (c.value != null && String(c.value) === String(v)));
	return c?.name || String(v);
}
export function getCategoryLabel(category: Category | undefined, codes: CodeItem[]) {
	const { c1 = 0, c2 = 0, c3 = 0 } = category || ({} as Category);
	if (!c1) return '미선택';
	const a = codes.find((c) => c.id === c1);
	if (!a) return '미선택';
	if (!c2) return a.name;
	const b = a.items?.find((c) => c.id === c2);
	if (!b) return a.name;
	if (!c3) return `${a.name} > ${b.name}`;
	const c = b.items?.find((c) => c.id === c3);
	return c ? `${a.name} > ${b.name} > ${c.name}` : `${a.name} > ${b.name}`;
}
export function getStatusLabelsForView(s: WorkSearch): string[] {
	const vm = s.viewMode || '';
	const map = s.statusLabelsByViewMode;
	if (map && Array.isArray(map[vm])) return map[vm];
	return Array.isArray(s.statusLabels) ? s.statusLabels : [];
}
export function setStatusLabelsForView(s: WorkSearch, next: string[]): WorkSearch {
	const vm = s.viewMode || '';
	const { statusLabels: _drop, ...rest } = s;
	return { ...rest, statusLabelsByViewMode: { ...(s.statusLabelsByViewMode || {}), [vm]: next } };
}
/** 옛 본문(그냥 글)은 줄바꿈을 <br> 로 — HTML 이면 그대로(TinyMCE 가 넣는 \n 을 <br> 로 바꾸면 저장할 때마다 빈 줄이 늘었다) */
export function toDescHtml(text: string | undefined, br = '<br>') {
	if (!text) return text || '';
	return /<(p|div|table|h[1-6]|ul|ol|li|br|blockquote|pre|img|span|figure)[\s/>]/i.test(text) ? text : `<p>${text.split('\n').join(br)}</p>`;
}
export const SHAPE_LABELS: Record<string, string> = { rect: '사각형', circle: '원', arrow: '화살표', pen: '펜', 'no-shape': '메모' };
export const clone = <T>(o: T): T => (o == null ? o : JSON.parse(JSON.stringify(o)));
/** 마스터 보드를 나눠 받은 호스팅들 */
export function getWorkSharedHostingIds(groupKey: string): string[] {
	const s = (board(groupKey) as any)?.setting;
	const ids = s?.sharedHostingIds ?? s?.sharedCompanyIds; // 옛 이름 sharedCompanyIds — 2026-10 옮기는 기간
	return Array.isArray(ids) ? ids.filter((x: any) => typeof x === 'string') : [];
}

/** 발행자 칸 — 담당자(ADMINS) 칸 바로 뒤, 담당자 칸이 없으면 맨 앞(목록·상세 머리 같이) */
export function withIssuer<T extends { key: string; inputType?: string }>(list: T[]): T[] {
	if (list.some((c) => c.key === 'issuer')) return list;
	const col = { key: 'issuer', name: '발행자', width: 120, sticky: false } as unknown as T;
	const i = list.findIndex((c) => c.inputType === 'ADMINS');
	return i === -1 ? [col, ...list] : [...list.slice(0, i + 1), col, ...list.slice(i + 1)];
}
