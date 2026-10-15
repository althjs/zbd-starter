/**
 * 게시판 설정(POST_GROUPS) 읽기 — React common-constants getPostGroup · post/constants · board/constants 자리.
 */
import { getMeta } from '../../meta.svelte.ts';
import type { Post, PostGroup } from '#lib/shared/types/post.ts';

export const statusMapDefault: Record<string, string> = { 0: '공개', 99: '미공개' };

export const adminPostGroups = (): PostGroup[] => getMeta('POST_GROUPS', 'adminPostGroups') || [];
export const userPostGroups = (): PostGroup[] => getMeta('POST_GROUPS', 'userPostGroups') || [];

/** 관리자 게시판이면 dataColumn=json, 회원 게시판이면 jsonUser */
export function getPostGroup(groupKey: string | null | undefined): PostGroup {
	let o = adminPostGroups().find((x) => x.groupKey == groupKey);
	if (o) return { ...o, dataColumn: 'json' };
	o = userPostGroups().find((x) => x.groupKey == groupKey);
	if (o) return { ...o, dataColumn: 'jsonUser' };
	return { groupKey: groupKey || '', dataColumn: 'json' };
}

export const dataColumnOf = (g: PostGroup | null | undefined) => (g?.dataColumn || 'json') as 'json' | 'jsonUser';

/**
 * 관리자 화면의 작성자 이름 — 회원 글이고 저장된 이름에 「*」 가 있으면 회원 실명으로(React getPostAuthorInfo).
 */
export function getPostAuthorInfo(post: Post, userMap: Record<number, any> | null) {
	const raw = post?.jsonUser?.name || '';
	if (!post?.userId || raw.indexOf('*') === -1) return { name: raw, displayName: raw, unmasked: false };
	const real = userMap?.[post.userId]?.name;
	if (!real || real === raw) return { name: raw, displayName: raw, unmasked: false };
	return { name: real, displayName: raw, unmasked: true };
}

/** 문자열로 온 json·jsonUser 를 풀고 요약(조회수·반응·댓글)을 json 에 붙인다(React convertPostItems) */
export function convertPostItems(list: any[] | null | undefined, summary: any = null): Post[] {
	return (list || []).map((p) => {
		if (p.json && typeof p.json == 'string') p.json = JSON.parse(p.json);
		else if (!p.json) p.json = {};
		if (p.jsonUser && typeof p.jsonUser == 'string') p.jsonUser = JSON.parse(p.jsonUser);
		if (!p.jsonUser) p.jsonUser = {};
		if (p.nickname) {
			if (p.jsonUser?.name) p.jsonUser.name = p.nickname;
			if (p.json?.name) p.json.name = p.nickname;
		}
		p.json.viewCount = summary?.pvMap?.[p.postId] || 0;
		p.json.reaction = summary?.reactionMap?.[p.postId] || {};
		p.json.commentCount = summary?.commentMap?.[p.postId] || 0;
		return p as Post;
	});
}

export const postDefault = (): Post => ({ postId: '', groupKey: '', groupSubKey: '', title: '', json: { body: '', name: '' }, jsonUser: { body: '', name: '' }, status: 0 });
