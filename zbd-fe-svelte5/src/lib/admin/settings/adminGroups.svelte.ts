/**
 * 관리자 그룹 목록(React store/zustandAdminGroup) — 처음 읽을 때 한 번 받는다. reloadAdminGroups() 는 강제로(force=1).
 *   adminGroups.list   (null = 아직)
 */
import { api } from '#lib/shared/api.ts';
import type { AdminGroup } from '#lib/shared/types/admin.ts';

export const adminGroups = $state({ list: null as AdminGroup[] | null });
let started = false;

async function load(force = false) {
	const o = await api.get(`/service/admin/group${force ? '?force=1' : ''}`);
	if (o.ok) adminGroups.list = o.data || [];
}
export function ensureAdminGroups() {
	if (started) return;
	started = true;
	load();
}
export const reloadAdminGroups = () => load(true);
