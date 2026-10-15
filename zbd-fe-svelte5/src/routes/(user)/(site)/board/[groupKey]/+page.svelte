<script lang="ts">
	/**
	 * 게시판 목록 `/board/{groupKey}` — 예전 pages/board/[groupKey]/index.tsx.
	 * 대문자 주소(/board/NOTICE) 는 소문자로 보낸다.
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import BoardListPage from '#lib/user/board/BoardListPage.svelte';
	import { findPostGroup } from '#lib/user/board/board.svelte.ts';
	import { site } from '#lib/shared/site.svelte.ts';

	const groupKey = $derived(page.params.groupKey || '');
	$effect(() => {
		if (groupKey != groupKey.toLowerCase()) goto(`/board/${groupKey.toLowerCase()}${page.url.search}`, { replace: true });
	});
	const group = $derived(site.base ? findPostGroup(groupKey) : undefined);
</script>

{#if group}
	<Container><BoardListPage path="/board" {groupKey} isMobile={viewport.isMobile} /></Container>
{/if}
