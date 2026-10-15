/**
 * 관리자 회원 목록(React store/zustandAdminGroupMember) — GET /service/admin/member 를 받아 json 을 풀고 번호 큰 순.
 *   adminMembers.list (null = 아직) · loadAdminMembers(force) · setAdminMembers(list)
 * 관리자(운영자)는 수십 명이라 전체를 받아 화면이 거른다(LIST_AUDIT 3-1 「전체를 받아도 됨」).
 */
import { api } from '#lib/shared/api.ts';

export const adminMembers = $state({ list: null as any[] | null });

export async function loadAdminMembers() {
	const o = await api.get('/service/admin/member', { headers: { cache: 'false' } });
	if (!o.ok) return;
	adminMembers.list = (o.data || [])
		.map((m: any) => {
			m.json = typeof m.json == 'string' ? JSON.parse(m.json || '{}') : m.json || {};
			m.Json = m.json;
			return m;
		})
		.sort((a: any, b: any) => b.adminId - a.adminId);
}
export function setAdminMembers(list: any[]) {
	adminMembers.list = list;
}

/** 관리자 레코드의 사이트 설정(React getAdminSiteMeta) — META_ITEMS 의 SITE 또는 json.SITE */
export function getAdminSiteMeta(json: any): any {
	if (!json) return {};
	return (json.META_ITEMS || []).find((x: any) => x?.metaKey === 'SITE')?.json || json.SITE || {};
}
