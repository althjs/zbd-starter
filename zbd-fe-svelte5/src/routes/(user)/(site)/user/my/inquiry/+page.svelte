<script lang="ts">
	/** 내 문의 — 예전 pages/user/my/inquiry (CS 게시판 ?type=my). */
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import { convertPost } from '#lib/user/board/board.svelte.ts';
	import MyPostsPage from '#lib/user/my/MyPostsPage.svelte';

	async function load() {
		const o = await api.get(`/service/board/${HOSTING_ID}?groupKey=CS&type=my`);
		if (!o.ok) {
			alert('조회 도중 오류가 발생하였습니다.');
			return [];
		}
		return (o.data || []).map((x: any) => convertPost(x, o.meta.summary || null));
	}
</script>

<MyPostsPage title="내 문의" className="my-inquiry" cssKey="BOARD_CSS" emptyText="등록하신 문의사항이 없습니다." {load} />
