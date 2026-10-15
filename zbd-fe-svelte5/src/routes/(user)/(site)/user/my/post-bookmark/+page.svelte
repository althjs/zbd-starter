<script lang="ts">
	/** 저장한 게시글 — 예전 pages/user/my/post-bookmark (회원 postIds 를 최근 저장 순으로) */
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import { convertPost } from '#lib/user/board/board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import MyPostsPage from '#lib/user/my/MyPostsPage.svelte';
	import { session } from '#lib/user/session.svelte.ts';

	async function load() {
		const ids: string[] = session.profile?.postIds || [];
		const o = await api.get(`/service/board/${HOSTING_ID}?postIds=${ids.join(',')}`);
		if (!o.ok) {
			alert('게시글 조회 도중 오류가 발생하였습니다.');
			return [];
		}
		const posts: Post[] = (o.data || []).map((x: any) => convertPost(x, o.meta.summary || null));
		return [...ids]
			.reverse()
			.map((id) => posts.find((p) => p.postId === id))
			.filter(Boolean) as Post[];
	}
</script>

<MyPostsPage title="저장한 게시글" {load} />
