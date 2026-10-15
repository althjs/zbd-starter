/**
 * 게시판 데이터 — 예전 store/zustandBoardStore.ts 중 즈비드가 쓰는 길.
 *
 * 게시판 종류 두 가지(게시판 설정 feature.useES):
 *   - 검색엔진(ES) 게시판(자유게시판 등) — 쪽마다 서버에 묻는다(getPosts)
 *   - 그 밖(공지·약관 등) — 그 게시판 글을 한 번에 받아 화면에서 거르고 쪽을 나눈다(getGroupPosts)
 * 글의 json·jsonUser 는 문자열로 온다 — 객체로 풀고, 조회수·반응·댓글 수(summary) 를 채운다.
 */
import { api } from '#lib/shared/api.ts';
import { HOSTING_ID } from '#lib/shared/config.ts';
import { site } from '#lib/shared/site.svelte.ts';
import type { Post, PostSummary } from '#lib/shared/types/post.ts';

function parse(v: any) {
	if (!v) return {};
	if (typeof v == 'string') {
		try {
			return JSON.parse(v) || {};
		} catch {
			return {};
		}
	}
	return v;
}

export function convertPost(o: any, summary: PostSummary = null): Post {
	o.json = parse(o.json);
	o.jsonUser = parse(o.jsonUser);
	if (o.nickname) {
		if (o.json?.name) o.json.name = o.nickname;
		if (o.jsonUser?.name) o.jsonUser.name = o.nickname;
	}
	o.json.viewCount = summary?.pvMap?.[o.postId] || o.json.viewCount || 0;
	o.json.reaction = summary?.reactionMap?.[o.postId] || o.json.reaction || {};
	o.json.commentCount = summary?.commentMap?.[o.postId] || o.json.commentCount || 0;
	return o as Post;
}

/** 게시판 설정 — 관리자 게시판 + 회원 게시판(isUserBoard). groupKey 대소문자 무시 */
export function findPostGroup(groupKey?: string | null): any {
	const key = String(groupKey || '').toLowerCase();
	if (!key) return null;
	const meta = site.base?.postGroupsMeta || {};
	const list = [...(meta.adminPostGroups || []), ...(meta.userPostGroups || []).map((o: any) => ({ ...o, isUserBoard: true }))];
	return list.find((x: any) => String(x?.groupKey || '').toLowerCase() == key) || null;
}

export const board = $state({
	/** 한 번에 받은 게시판 글(ES 아닌 게시판) — groupKey(대문자) → 글 */
	postMap: {} as Record<string, Post[]>,
	fixedPostMap: {} as Record<string, Post[]>,
	summary: null as PostSummary,
	/** 없는(또는 비공개) 글 */
	unknown: {} as Record<string, boolean>,
	/** 글을 쓰거나 고치거나 지우면 오른다 — 목록이 다시 읽는다 */
	version: 0
});

/** 글쓰기 창이 저장한 글을 받은 목록에 반영(새 글은 맨 앞, 지운 글은 뺀다) */
export function updatePost(p: any) {
	const key = String(p?.groupKey || '').toUpperCase();
	const list = board.postMap[key];
	if (list) {
		const i = list.findIndex((x) => x.postId == p.postId);
		if (p.DELETED) {
			if (i != -1) list.splice(i, 1);
		} else {
			const o = convertPost({ ...p }, board.summary);
			if (i != -1) list[i] = o;
			else list.splice(list.findIndex((x) => !x.top) == -1 ? list.length : list.findIndex((x) => !x.top), 0, o);
		}
	}
	board.version++;
}

const groupLoading = new Map<string, Promise<Post[]>>();

/** ES 아닌 게시판의 글 전부(위에 고정한 글 먼저, 최신순) */
export function getGroupPosts(groupKey: string, force = false): Promise<Post[]> {
	const key = groupKey.toUpperCase();
	if (!force && board.postMap[key]) return Promise.resolve(board.postMap[key]);
	if (!force && groupLoading.has(key)) return groupLoading.get(key)!;
	const p = api.get(`/service/board/${HOSTING_ID}?groupKey=${key}`).then((o) => {
		groupLoading.delete(key);
		if (!o.ok) return [];
		const summary = o.meta.summary || null;
		board.summary = { ...(board.summary || {}), ...(summary || {}) };
		const posts = (o.data || []).map((x: any) => convertPost(x, summary));
		posts.forEach((x: Post) => (x.top = x.json?.isFixTop ? 1 : 0));
		posts.sort((a: Post, b: Post) => (b.top || 0) - (a.top || 0) || Number(b.createdAt) - Number(a.createdAt));
		board.postMap[key] = posts;
		const fixed = (o.meta.fixedPosts || []).map((x: any) => convertPost(x, summary));
		if (fixed.length) board.fixedPostMap[key] = fixed;
		return posts;
	});
	groupLoading.set(key, p);
	return p;
}

/**
 * 게시판 한 쪽 — 서버가 나누고 찾는다(Go `/api/v1/boards/{게시판}/posts`, CLAUDE.md 4). 위 고정 글이 먼저(서버 기본 정렬).
 * 목록 줄 모양(Post)으로 바꿔 돌려준다 — 본문은 앞 100자(excerpt)만 온다.
 */
export async function getBoardPage(s: { groupKey: string; pageNo: number; pageSize: number; keyword?: string; groupSubKey?: string }): Promise<{ posts: Post[]; total: number }> {
	const q = new URLSearchParams({ page: String(s.pageNo || 1), size: String(Math.min(100, s.pageSize || 10)) });
	if (s.keyword?.trim()) q.set('q', s.keyword.trim());
	if (s.groupSubKey) q.set('groupSubKey', s.groupSubKey);
	const o = await api.get<any[]>(`/api/v1/boards/${encodeURIComponent(s.groupKey.toUpperCase())}/posts?${q}`);
	if (!o.ok) return { posts: [], total: 0 };
	const total = Number(o.meta?.total) || 0;
	const offset = (Math.max(1, s.pageNo || 1) - 1) * Math.min(100, s.pageSize || 10);
	const posts = (o.data || []).map((x: any, i: number) => {
		const name = x.writer?.nickname || '';
		return {
			postId: x.postId,
			// 목록 번호(NO) — 저장된 값이 아니라 자리 번호(전체 - 앞 쪽 - 줄). 예전 목록 API 와 같은 계산
			postNo: Math.max(total - offset - i, 1),
			groupKey: x.groupKey,
			groupSubKey: x.groupSubKey || '',
			title: x.title,
			userId: x.writer?.userId || 0,
			nickname: name,
			status: 0,
			createdAt: x.createdAt,
			updatedAt: x.updatedAt,
			top: x.isFixTop ? 1 : 0,
			json: { name, isFixTop: !!x.isFixTop, image: x.image, viewCount: x.viewCount || 0, commentCount: x.commentCount || 0, body: x.excerpt || '' },
			jsonUser: { name, image: x.image, body: x.excerpt || '' }
		} as Post;
	});
	return { posts, total };
}

/** ES 게시판 한 쪽 */
export async function getPosts(search: Record<string, any>): Promise<{ posts: Post[]; total: number }> {
	const params = new URLSearchParams();
	for (const [k, v] of Object.entries(search)) {
		if (v !== null && v !== undefined) params.append(k, k == 'groupKey' ? String(v).toUpperCase() : String(v));
	}
	const o = await api.get(`/service/board/${HOSTING_ID}?${params}`);
	if (!o.ok) return { posts: [], total: 0 };
	const summary = o.meta.summary || null;
	board.summary = { ...(board.summary || {}), ...(summary || {}) };
	return { posts: (o.data || []).map((x: any) => convertPost(x, summary)), total: o.meta.total || 0 };
}

/** 글 하나 — 받은 목록에 있으면 그것을, 없으면 서버에서 */
export async function getPost(groupKey: string, postId: string): Promise<Post | null> {
	const hit = (board.postMap[groupKey.toUpperCase()] || []).find((x) => x.postId == postId);
	if (hit) return hit;
	const o = await api.get(`/service/board/${HOSTING_ID}/${postId}`);
	if (!o.ok) {
		board.unknown[postId] = true;
		return null;
	}
	return convertPost(o.data, o.meta.summary || null);
}

/** 조회수 올리기(상세를 열 때 한 번) */
const counted = new Set<string>();
export function countUp(post: Post) {
	// 개발 서버는 운영 API 를 본다 — 시험하느라 운영 조회수를 올리지 않는다
	if (import.meta.env.DEV || counted.has(post.postId)) return;
	counted.add(post.postId);
	api.post(`/service/board/${HOSTING_ID}`, { postId: post.postId });
}

/** 목록 → 상세 주소. 답글은 원글로 보내고 답글 위치를 실어 보낸다(예전 utils/postLink.ts) */
export function postHref(path: string, post: Post) {
	const parents: any[] = post?.json?.parentPostIds || [];
	const postId = String(post?.postId || '');
	const root = parents.length ? String(parents[0]) : '';
	let href = path.includes('?') ? `${path}&postId=${root || postId}` : `${path}/${root || postId}`;
	if (root && postId) href += `${href.includes('?') ? '&' : '?'}replyPostId=${postId}`;
	return href;
}
