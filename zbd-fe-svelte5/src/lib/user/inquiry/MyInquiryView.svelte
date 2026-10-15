<script lang="ts">
	/**
	 * 1:1 문의 상세 `/user/my/inquiry/{postId}` — 예전 pages/user/my/inquiry/[postId].tsx.
	 * 로그인 필요 · 내 글을 GET /service/user/post/{postId} 로 읽어 CS 게시판 설정으로 글 보기(수정 가능) · 목록 = /user/my/inquiry.
	 * 게시판 CSS(BOARD_CSS) 를 이 화면에 씌운다.
	 */
	import { onDestroy } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { applyCSS, clearCSS } from '#lib/shared/css.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import { site } from '#lib/shared/site.svelte.ts';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import AuthGuard from '../AuthGuard.svelte';
	import { convertPost, findPostGroup } from '../board/board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostView from '../board/PostView.svelte';

	let { postId }: { postId: string } = $props();

	const isMobile = $derived(viewport.isMobile);
	const postGroup = $derived(site.base ? findPostGroup('cs') : null);
	let post = $state<Post | null>(null);

	$effect(() => {
		const id = postId;
		post = null;
		api.get(`/service/user/post/${id}`).then((o) => {
			if (!o.ok) {
				alert('조회 도중 오류가 발생하였습니다.');
				return;
			}
			if (postId == id && o.data) post = convertPost(o.data);
		});
	});
	$effect(() => {
		const css = site.base?.CSS?.BOARD_CSS;
		if (css) applyCSS(css, 'BOARD_CSS');
	});
	onDestroy(() => clearCSS('BOARD_CSS'));
</script>

<AuthGuard>
	<Container sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: isMobile ? 40 : 80 }}>
		{#if post && postGroup}
			{#key post.postId}
				<PostView useEdit spacing={24} {post} {postGroup} path="/user/my/inquiry" {isMobile} />
			{/key}
		{/if}
	</Container>
</AuthGuard>
