/** 사용자 회원 목록 거르기(React pages/admin/user/member/constants filterUserMember 그대로) */
export function filterUserMember(s: any, us: any[]): any[] {
	if ((s.keyword || '').trim()) {
		const k = s.keyword.trim();
		us = us.filter((o) => (o.name || '').indexOf(k) !== -1 || (o.nickname || '').indexOf(k) !== -1 || (o.email || '').indexOf(k) !== -1 || (o.mobile || '').indexOf(k) !== -1);
	}
	const userStatus = typeof s.userStatus == 'undefined' ? 0 : s.userStatus;
	us = us.filter((o) => o.status == userStatus);
	if (s.agreeMap && Object.keys(s.agreeMap).length > 0) {
		us = us.filter((o) =>
			Object.keys(s.agreeMap).every((sub) => {
				const cond = s.agreeMap[sub];
				const ua = o.Json?.agreeMap?.[sub];
				if (typeof cond === 'boolean') {
					if (!cond) return true;
					if (typeof ua === 'boolean') return ua === true;
					if (typeof ua === 'object' && ua !== null) return Object.values(ua).some((v) => v === true);
					return false;
				}
				if (typeof cond === 'object' && cond !== null) {
					if (!Object.values(cond).some((v) => v === true)) return true;
					if (typeof ua !== 'object' || ua === null) return false;
					return Object.keys(cond).every((k) => (cond[k] === true ? ua[k] === true : true));
				}
				return true;
			})
		);
	}
	if (s.type) us = us.filter((o) => (s.type == 'GENERAL' ? typeof o.Json?.type == 'undefined' : o.Json?.type == s.type));
	if (s.authGroupId) us = us.filter((o) => (o.Json?.auth?.authGroupId || 1) == s.authGroupId);
	if (s.excludeUserIds?.length > 0) us = us.filter((o) => s.excludeUserIds.indexOf(o.userId) === -1);
	return us;
}

/** 정렬(lodash orderBy 하나짜리) */
export function orderBy<T>(list: T[], key: string, dir: 'asc' | 'desc'): T[] {
	const m = dir == 'desc' ? -1 : 1;
	return [...list].sort((a: any, b: any) => (a?.[key] > b?.[key] ? m : a?.[key] < b?.[key] ? -m : 0));
}
