<script lang="ts">
	/** 고객센터 홈 — 예전 pages/cs/index.tsx (자주 묻는 질문 5개 + 공지 2개) */
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { getGroupPosts } from '#lib/user/board/board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import FAQ from '#lib/user/cs/FAQ.svelte';
	import Notice from '#lib/user/cs/Notice.svelte';

	const isMobile = $derived(viewport.isMobile);
	let faq = $state<Post[] | null>(null);
	let notice = $state<Post[] | null>(null);
	getGroupPosts('FAQ').then((p) => (faq = p));
	getGroupPosts('NOTICE').then((p) => (notice = p));
</script>

<Container sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: isMobile ? 40 : 80 }}>
	<h2 class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ textAlign: 'center', marginBottom: 24 }))}>고객센터</h2>
	{#if faq?.length}<FAQ isCSHome items={faq} {isMobile} />{/if}
	<Notice sx={{ marginTop: 24 }} items={notice} {isMobile} />
</Container>
