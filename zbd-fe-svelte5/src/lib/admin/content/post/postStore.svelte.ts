/**
 * 게시글 목록 상태 — React store/zustandPostStore 자리(관리자 게시판·포스트 관리가 같이 쓴다).
 *   posts.groupMap[groupKey] — `GET /service/admin/post/group/{groupKey}` 를 한 번 읽어 둔다(isForce 면 다시)
 *   getEsPosts(search) — 검색 엔진(ES) 게시판은 서버가 나눠 준다(`GET /service/board/{호스팅}?…`)
 */
import { api } from '#lib/shared/api.ts';
import { getHostingId, getMeta } from '../../meta.svelte.ts';
import type { Post } from '#lib/shared/types/post.ts';
import { convertPostItems, getPostGroup } from './postGroup.ts';

export const posts = $state({ groupMap: {} as Record<string, Post[] | undefined>, postMap: {} as Record<string, Post> });
const loading = new Set<string>();

function setGroupItems(groupKey: string, items: Post[]) {
	posts.groupMap[groupKey] = items;
	for (const p of items) posts.postMap[p.postId] = p;
}

export async function getPostGroupItems(groupKey: string, isForce = false) {
	if (!groupKey) return null;
	if (loading.has(groupKey) && !isForce) return posts.groupMap[groupKey] || null;
	loading.add(groupKey);
	const o = await api.get(`/service/admin/post/group/${groupKey}`, { headers: { cache: 'false' } });
	if (!o.ok) {
		setGroupItems(groupKey, []);
		return [];
	}
	const summary = o.meta?.summary;
	const list: Post[] = (o.data || []).map((p: any) => {
		if (typeof p.json == 'string' && p.json) p.json = JSON.parse(p.json);
		p.json ||= {};
		if (typeof p.json.sortNo == 'number') p.sortNo = p.json.sortNo;
		if (summary) {
			p.json.viewCount = summary.pvMap?.[p.postId] || 0;
			p.json.reaction = summary.reactionMap?.[p.postId] || {};
			p.json.commentCount = summary.commentMap?.[p.postId] || 0;
		}
		p.jsonUser = typeof p.jsonUser == 'string' ? JSON.parse(p.jsonUser || '{}') : p.jsonUser || {};
		return p;
	});
	setGroupItems(groupKey, list);
	return list;
}

/** 저장 뒤 목록 한 줄 바꾸기(없으면 더하고, DELETED 면 뺀다) */
export function setPostGroupItem(groupKey: string, post: Post) {
	let found = false;
	const list: Post[] = [];
	for (const o of posts.groupMap[groupKey] || []) {
		if (o.postId == post.postId) {
			if (typeof post.json == 'string') post.json = JSON.parse(post.json);
			if (typeof post.jsonUser == 'string') post.jsonUser = JSON.parse(post.jsonUser);
			if (typeof post.json?.sortNo == 'number') post.sortNo = post.json.sortNo;
			if (!post.DELETED) list.push(post);
			found = true;
		} else list.push(o);
	}
	if (!found && !post.DELETED) list.push(post);
	setGroupItems(groupKey, list);
}

/** 답글 게시판이면 하위 글을 붙인다 */
async function getReplyPosts(list: Post[]) {
	const ids = list.filter((p) => getPostGroup(p.groupKey)?.feature?.useReply).map((p) => p.postId);
	if (!ids.length) return [];
	const o = await api.post(`/service/board/${getHostingId()}/replies`, { postIds: ids });
	return o.ok ? convertPostItems(o.data || [], o.meta?.summary) : [];
}

function assignReplies(list: Post[], replies: Post[]) {
	if (!replies.length) return list;
	const all = [...replies];
	const rmap = new Map(replies.map((r) => [r.postId, r]));
	const nested = new Set<string>();
	for (const r of replies) {
		const ids = r.json?.parentPostIds || [];
		const parent = ids.length ? rmap.get(ids[ids.length - 1]) : null;
		if (parent) {
			(parent.replies ||= []).push(r);
			nested.add(r.postId);
		}
	}
	const top = replies.filter((r) => !nested.has(r.postId));
	const pmap = new Map(list.map((p) => [p.postId, p]));
	for (const r of top) {
		const ids = r.json?.parentPostIds || [];
		const parent = ids.length ? pmap.get(ids[0]) : null;
		if (parent) (parent.replies ||= []).push(r);
	}
	const drop = new Set(all.filter((r) => (r.json?.parentPostIds || []).length && pmap.has(r.json.parentPostIds![0])).map((r) => String(r.postId)));
	return drop.size ? list.filter((p) => !drop.has(String(p.postId))) : list;
}

export async function getEsPosts(search: Record<string, any>) {
	const params = new URLSearchParams();
	for (const k of Object.keys(search)) {
		if (search[k] !== null && search[k] !== undefined) params.append(k, k == 'groupKey' ? String(search[k]).toUpperCase() : search[k]);
	}
	const o = await api.get(`/service/board/${getHostingId()}?${params.toString()}`);
	if (!o.ok) return { totalCount: 0, posts: [] as Post[] };
	const summary = o.meta?.summary || null;
	let list = [...convertPostItems(o.meta?.fixedPosts || [], summary), ...convertPostItems(o.data || [], summary)];
	list = assignReplies(list, await getReplyPosts(list));
	for (const p of list) posts.postMap[p.postId] = p;
	return { totalCount: Number(o.meta?.totalCount || 0), posts: list };
}

/** 글 하나(신고 정보 포함) — `GET /service/board/{호스팅}/{postId}` */
export async function getPost(postId: string, force = false): Promise<Post | null> {
	if (!force && posts.postMap[postId]?.reportMap) return posts.postMap[postId];
	const o = await api.get(`/service/board/${getHostingId()}/${postId}`);
	if (!o.ok) return null;
	const p: Post = o.data;
	if (typeof p.json == 'string') p.json = JSON.parse(p.json || '{}');
	if (typeof p.jsonUser == 'string') p.jsonUser = JSON.parse(p.jsonUser || '{}');
	p.json ||= {};
	const summary = o.meta?.summary;
	p.json.viewCount = summary?.pvMap?.[p.postId] || 0;
	p.json.reaction = summary?.reactionMap?.[p.postId] || {};
	p.json.commentCount = summary?.commentMap?.[p.postId] || 0;
	p.reportMap = summary?.reportMap || {};
	const codes = (getMeta('COMMON_CODES')?.items || []).find((x: any) => x.group == 'POST_ISSUE_REPORT')?.items || [];
	for (const x of p.reportMap?.postReport || []) x.description = codes.find((c: any) => c.id == x.reaction)?.name || '알수없는 신고사유';
	for (const xs of Object.values(p.reportMap?.reportedCommentIdMap || {}) as any[]) for (const x of xs) x.description = codes.find((c: any) => c.id == x.reaction)?.name || '알수없는 신고사유';
	posts.postMap[p.postId] = { ...(posts.postMap[p.postId] || {}), ...p };
	return posts.postMap[p.postId];
}

/** 관리자 글 한 개(편집용) — `GET /service/admin/post/{postId}` */
export async function getAdminPost(postId: string) {
	const o = await api.get(`/service/admin/post/${postId}`);
	return o.ok ? (o.data as Post) : null;
}
