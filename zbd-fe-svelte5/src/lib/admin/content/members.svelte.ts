/**
 * 관리자·회원 명단(이름·이메일 표시용) — React zustandAdminGroupMember · zustandUserMemberManageStore 의 읽기 부분.
 * 처음 부를 때 한 번 읽는다. 목록 줄의 「작성자 (이메일)」 칸이 쓴다.
 */
import { api } from '#lib/shared/api.ts';

export const members = $state({ admins: null as any[] | null, userMap: null as Record<number, any> | null });
let adminLoading = false;
let userLoading = false;

export function loadAdminMembers() {
	if (adminLoading) return;
	adminLoading = true;
	api.get('/service/admin/member', { headers: { cache: 'false' } }).then((o) => {
		if (!o.ok) return;
		members.admins = (o.data || []).map((m: any) => {
			if (typeof m.json == 'string') m.json = JSON.parse(m.json || '{}');
			m.Json = m.json;
			return m;
		});
	});
}

export function loadAllUsers() {
	if (userLoading) return;
	userLoading = true;
	api.get('/service/admin/user?isIncludeJson=1').then((o) => {
		if (!o.ok) return;
		const map: Record<number, any> = {};
		for (const u of o.data || []) {
			u.Json = JSON.parse(u.json || '{}');
			delete u.json;
			map[u.userId] = u;
		}
		members.userMap = map;
	});
}
