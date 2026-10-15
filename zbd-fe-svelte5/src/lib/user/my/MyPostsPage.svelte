<script lang="ts">
	/**
	 * 마이페이지 글 목록 틀 — 「내 게시글」(my/posts) · 「저장한 게시글」(my/post-bookmark) · 「내 문의」(my/inquiry).
	 * 예전 화면들이 같은 모양(제목 · 카드 안 사진형 목록 · 쪽 번호)이라 하나로. 운영에선 전역 CSS 로 숨긴 메뉴들이다.
	 * ⚠ v1 목록 대기 — 지금 API 가 전체를 주어 화면에서 쪽을 나눈다(CLAUDE.md 4). Go v1 목록이 생기면 page·size 를 넘긴다.
	 */
	import { onDestroy } from 'svelte';
	import { applyCSS, clearCSS } from '#lib/shared/css.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import TablePagination from '#lib/shared/mui/TablePagination.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { site } from '#lib/shared/site.svelte.ts';
	import AuthGuard from '../AuthGuard.svelte';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostList from '../board/PostList.svelte';
	import { session } from '../session.svelte.ts';

	let {
		title,
		className = 'my-posts',
		cssKey = 'MY_POSTS_CSS',
		emptyText = '',
		load
	}: { title: string; className?: string; cssKey?: string; emptyText?: string; load: () => Promise<Post[]> } = $props();

	const isMobile = $derived(viewport.isMobile);
	let all = $state<Post[] | null>(null);
	let pageNo = $state(1);
	let pageSize = $state(5);
	const list = $derived((all || []).slice((pageNo - 1) * pageSize, pageNo * pageSize));

	$effect(() => {
		if (session.profile?.userId) load().then((ps) => (all = ps));
	});
	$effect(() => {
		const css = site.base?.CSS?.[cssKey];
		if (css) applyCSS(css, cssKey);
	});
	onDestroy(() => clearCSS(cssKey));
</script>

<AuthGuard>
	<Container class={className} sx={{ marginTop: isMobile ? 40 : 80, marginBottom: isMobile ? 40 : 80 }}>
		<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ textAlign: 'center' }))}>{title}</svelte:element>
		{#if all && emptyText && all.length == 0}
			<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 40, marginBottom: 40 }))}>{emptyText}</p>
		{:else if all}
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
					toClass({ padding: 16, marginTop: 16, '& table tr:nth-last-of-type(1)': { borderBottom: 'none' } })
				)}
			>
				<PostList spacing={16} path="" items={list} postGroup={{ feature: { listTheme: 'img-table' } }} hiddenTitle totalCount={0} search={{}} getData={() => {}} {isMobile} />
			</div>
			<TablePagination
				{pageNo}
				{pageSize}
				pageCount={Math.ceil((all.length || 1) / pageSize)}
				sx={{ height: '30px', marginTop: 16 }}
				onchange={(s) => ((pageNo = s.pageNo), (pageSize = s.pageSize), window.scrollTo(0, 0))}
			/>
		{/if}
	</Container>
</AuthGuard>
