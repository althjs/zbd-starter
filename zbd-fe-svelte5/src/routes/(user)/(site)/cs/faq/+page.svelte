<script lang="ts">
	/** 자주 묻는 질문 — 예전 pages/cs/faq (·/faq/{검색어}) */
	import { page } from '$app/state';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import { getGroupPosts } from '#lib/user/board/board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import FAQ from '#lib/user/cs/FAQ.svelte';

	let faq = $state<Post[] | null>(null);
	getGroupPosts('FAQ').then((p) => (faq = p));
	const keyword = $derived((page.params as any).keyword || '');
</script>

<Container sx={{ marginTop: viewport.isMobile ? 40 : 80, marginBottom: viewport.isMobile ? 40 : 80 }}>
	{#if faq?.length}
		{#key keyword}<FAQ items={faq} defaultKeyword={keyword} isMobile={viewport.isMobile} />{/key}
	{/if}
</Container>
