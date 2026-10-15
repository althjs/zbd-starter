/**
 * 관리자 목록(이름·프로필 사진) — React store/zustandAdminGroupMember 의 「전체 한 번 읽기」 몫. `GET /service/admin/member`.
 * 권한 관리·업무 보드 담당자가 같이 쓴다.
 */
import { api } from '#lib/shared/api.ts';
import { adminImagePrefix, getHostingId } from '../meta.svelte.ts';
import type { AdminMember } from '#lib/shared/types/admin.ts';

export const members = $state({ list: [] as AdminMember[], loaded: false });
let started = false;
export function loadMembers() {
	if (started) return;
	started = true;
	api.get('/service/admin/member', { headers: { cache: 'false' } }).then((o) => {
		if (!o.ok) return;
		const rows: AdminMember[] = (o.data || []).map((x: any) => {
			const j = typeof x.json == 'string' ? JSON.parse(x.json || '{}') : x.json || {};
			return { ...x, json: j, Json: j };
		});
		members.list = rows.sort((a, b) => b.adminId - a.adminId);
		members.loaded = true;
	});
}
/** adminId → 이름·사진 */
export function memberOf(aid?: number | null) {
	loadMembers();
	const m = members.list.find((x) => x.adminId == aid);
	const file = m?.Json?.image?.name || '';
	return { name: m?.name || (aid ? `관리자 ${aid}` : ''), image: file ? `${adminImagePrefix(m?.hostingId || getHostingId())}/admin_profile_${aid}/${file}` : '' };
}
