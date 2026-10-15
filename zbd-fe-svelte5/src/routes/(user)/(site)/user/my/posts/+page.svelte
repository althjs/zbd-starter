<script lang="ts">
	/** 내 게시글 — 예전 pages/user/my/posts (board ?type=my) */
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import { convertPost } from '#lib/user/board/board.svelte.ts';
	import MyPostsPage from '#lib/user/my/MyPostsPage.svelte';

	async function load() {
		const o = await api.get(`/service/board/${HOSTING_ID}?type=my`);
		if (!o.ok) {
			alert('게시글 조회 도중 오류가 발생하였습니다.');
			return [];
		}
		return (o.data || []).map((x: any) => convertPost(x, o.meta.summary || null));
	}
</script>

<MyPostsPage title="내 게시글" {load} />
