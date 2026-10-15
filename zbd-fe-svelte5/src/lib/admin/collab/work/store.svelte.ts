/**
 * 업무 보드 저장소 — React zustandWorkStore 와 같은 API.
 *   GET /service/admin/work?groupKey=  (보드 전체 — v1 목록 대기: 쪽 나누기 없이 다 받는다)
 *   GET/PUT/DELETE /service/admin/work/{id} · POST /service/admin/work {item} · PUT /service/admin/work/{id}/{shapeId}
 * 공유받은 보드는 응답 meta.workCodes 로 분류·열 설정이 함께 온다 → 내 COMMON_CODES 에 붙여 둔다(React 와 같게).
 */
import { api } from '#lib/shared/api.ts';
import { globalMeta } from '../../meta.svelte.ts';
import type { WorkItem, Shape } from '#lib/shared/types/work.ts';

export const work = $state({ items: [] as WorkItem[], loading: false, adminMembers: [] as any[] });
const loadingMap: Record<string, boolean> = {};

function mergeBoardCodes(groupKey: string, cc: any) {
	if (!cc || !groupKey) return;
	const codes = globalMeta.v.COMMON_CODES || { items: [] };
	if ((codes.items || []).find((x: any) => x.group == groupKey)) return;
	const id = Math.max(0, ...(codes.items || []).map((x: any) => Number(x.id) || 0)) + 1;
	// 화면 메타에만(저장 안 함) — 다음 메뉴 갱신 때 사라져도 다시 받는다
	globalMeta.v = { ...globalMeta.v, COMMON_CODES: { ...codes, items: [...(codes.items || []), { id, group: groupKey, name: cc.name || groupKey, items: cc.items || [], _sharedFromMaster: true }] } };
}
function upsert(list: WorkItem[]) {
	const next = [...work.items];
	for (const o of list) {
		const i = next.findIndex((x) => x.workId === o.workId);
		if (i > -1) next[i] = o;
		else next.push(o);
	}
	work.items = next;
}

export async function getItems(p: { groupKey?: string; workIds?: number[] }) {
	const key = p.groupKey || '__all__';
	if (loadingMap[key] && !p.workIds) return;
	loadingMap[key] = true;
	work.loading = true;
	const q = new URLSearchParams();
	if (p.groupKey) q.set('groupKey', p.groupKey);
	(p.workIds || []).forEach((id) => q.append('workIds[]', String(id)));
	const o = await api.get(`/service/admin/work?${q}`);
	work.loading = false;
	if (!p.workIds) loadingMap[key] = false;
	if (!o.ok) return;
	if (p.groupKey) mergeBoardCodes(p.groupKey, o.meta?.workCodes);
	upsert((o.data?.items || []) as WorkItem[]);
}

export async function getItem(workId: number | string): Promise<WorkItem | null> {
	const o = await api.get(`/service/admin/work/${workId}`);
	if (!o.ok) return null;
	const item = o.data as WorkItem | null;
	if (item?.groupKey) mergeBoardCodes(item.groupKey, o.meta?.workCodes);
	if (o.meta?.adminMembers) work.adminMembers = o.meta.adminMembers.map((m: any) => (typeof m.json == 'string' ? { ...m, Json: JSON.parse(m.json || '{}') } : m));
	if (item) upsert([item]);
	else work.items = work.items.filter((x) => x.workId !== Number(workId));
	return item;
}

export async function addItem(item: Omit<WorkItem, 'workId'>): Promise<WorkItem | null> {
	const o = await api.post('/service/admin/work', { item });
	return o.ok ? (o.data as WorkItem) : null;
}

/** 화면 먼저 바꾸고 저장 */
export async function updateItem(item: WorkItem): Promise<boolean> {
	work.items = work.items.map((x) => (x.workId === item.workId ? item : x));
	const o = await api.put(`/service/admin/work/${item.workId}`, { item });
	return o.ok ? (o.data ?? true) : false;
}

/** 메모가 가리키는 업무 — 새 칸(targetWorkIds) 또는 옛 하나(targetWorkId) */
export const shapeTargets = (s: Shape) => s.targetWorkIds ?? (s.targetWorkId ? [s.targetWorkId] : []);
/** 이 카드의 메모(본 그림·추가 이미지) 중 그 업무를 가리키는 것이 남아 있나 */
const shapesPointTo = (it: WorkItem, id: number) => [...(it.Json.shapes || []), ...(it.Json.images || []).flatMap((m) => m.shapes || [])].some((s) => !s.isDeleted && shapeTargets(s).includes(id));

/**
 * 연관 업무를 **양쪽 카드에** 걸고 푼다(2026-10-04 대표 — 서로 왔다 갔다) — 카드 Json.targetWorkIds.
 * 풀 때는 그 카드의 다른 메모가 아직 상대를 가리키면 남겨 둔다. 다른 보드 카드라 목록에 없으면 읽어 와서.
 */
export async function linkWorks(a: number, b: number, on: boolean) {
	for (const [me, other] of [
		[a, b],
		[b, a]
	]) {
		const it = work.items.find((x) => x.workId === me) || (await getItem(me));
		if (!it) continue;
		const ids = it.Json.targetWorkIds || [];
		if (on ? ids.includes(other) : !ids.includes(other) || shapesPointTo(it, other)) continue;
		await updateItem({ ...it, Json: { ...it.Json, targetWorkIds: on ? [...ids, other] : ids.filter((x) => x !== other) } });
	}
}

export async function deleteItem(workId: number) {
	const o = await api.delete(`/service/admin/work/${workId}`);
	return o.ok;
}

export async function updateShape(workId: number, shape: Shape, subImageIndex?: number | null) {
	syncShape(workId, shape, subImageIndex);
	const o = await api.put(`/service/admin/work/${workId}/${shape.id}`, { ...shape, subImageIndex: subImageIndex ?? undefined });
	return o.ok;
}

export function syncShape(workId: number, shape: Shape, subImageIndex?: number | null) {
	work.items = work.items.map((item) => {
		if (item.workId !== workId) return item;
		const sub = typeof subImageIndex === 'number';
		const shapes = (sub ? item.Json.images?.[subImageIndex!]?.shapes : item.Json.shapes) || [];
		const next = shape.isDeleted ? shapes.filter((s) => s.id !== shape.id) : shapes.find((s) => s.id === shape.id) ? shapes.map((s) => (s.id === shape.id ? shape : s)) : [...shapes, shape];
		if (sub) {
			const images = [...(item.Json.images || [])];
			images[subImageIndex!] = { ...images[subImageIndex!], shapes: next };
			return { ...item, Json: { ...item.Json, images } };
		}
		return { ...item, Json: { ...item.Json, shapes: next } };
	});
}

/** 분류 화면 순서 저장 — actionType = sortNo1|2|3 (그 분류 깊이) */
export async function doQuickUpdate(workIds: number[], actionType: 'sortNo1' | 'sortNo2' | 'sortNo3') {
	const o = await api.post('/service/admin/work/quick-update', { workIds, actionType });
	return o.ok ? !!(o.data ?? true) : false;
}
