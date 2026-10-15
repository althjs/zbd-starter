/**
 * 회원 게시판 쓰기 권한 — 예전 pages/user/post/constants.ts getPostGroupAuth 그대로.
 *   1. groupKey 가 권한 코드(AUTH_GROUP) 어디에도 없으면 RW
 *   2. POST_GROUP_AUTO 코드 id == 회원 auth.authGroupId 면 RW
 *   3. POST_GROUP_MANUAL 이고 회원 auth.authGroups 에 그 groupKey 가 있으면 RW
 * groupSubKeys = 쓸 수 있는 하위 분류
 */
import { site } from '#lib/shared/site.svelte.ts';
import { findPostGroup } from '../board/board.svelte.ts';

export type PostGroupAuth = { AUTH: 'R' | 'RW'; groupSubKeys: string[] };

export function getPostGroupAuth(profile: any, groupKey: string): PostGroupAuth {
	const authCodes: any[] = (site.base?.commonCodes || []).find((x: any) => x.group == 'AUTH_GROUP')?.items || [];
	const postGroup = findPostGroup(groupKey) || {};
	const auth = profile?.auth || {};
	let AUTH: 'R' | 'RW' = 'R';
	let groupSubKeys: string[] = [];
	const allSub = () => (postGroup.items || []).map((i: any) => i.groupSubKey).filter(Boolean);

	if (!authCodes.find((code) => code.items?.some((item: any) => item.name === groupKey))) AUTH = 'RW';
	let manualRW = false;
	for (const code of authCodes) {
		if (code.group === 'POST_GROUP_AUTO' && auth?.authGroupId === code.id) AUTH = 'RW';
		if (code.group === 'POST_GROUP_MANUAL' && auth?.authGroups?.some((g: any) => g?.groupKey === groupKey)) {
			AUTH = 'RW';
			manualRW = true;
		}
	}
	if (AUTH === 'RW') {
		if (manualRW) {
			const matched = (auth.authGroups || []).filter((g: any) => g.groupKey === groupKey);
			groupSubKeys = matched.map((g: any) => g.groupSubKey).filter(Boolean);
			if (!groupSubKeys.length) groupSubKeys = allSub();
		} else {
			const autoCode = authCodes.find((code) => code.group === 'POST_GROUP_AUTO' && code.id === auth?.authGroupId)?.items?.find((x: any) => x.name == groupKey);
			groupSubKeys = autoCode?.items?.length ? autoCode.items.map((i: any) => i.name) : allSub();
		}
	}
	return { AUTH, groupSubKeys };
}
