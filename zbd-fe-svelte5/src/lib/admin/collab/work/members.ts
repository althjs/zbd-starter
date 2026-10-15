/** 보드에서 쓰는 관리자 목록(이름·사진) — React 의 ADMIN_NOTI_MEMBER 모양으로 */
import { adminImagePrefix } from '../../meta.svelte.ts';
import type { NotiMember } from '#lib/shared/types/work.ts';

export function toNotiMembers(list: any[], allowed: (hostingId: string) => boolean): NotiMember[] {
	return (list || [])
		.filter((m) => allowed(m.hostingId || ''))
		.map((m) => ({ adminId: m.adminId, name: m.name, profileImage: m?.Json?.image?.name ? `${adminImagePrefix(m.hostingId || 'MASTER')}/admin_profile_${m.adminId}/${m.Json.image.name}` : '' }));
}
