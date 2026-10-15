<script lang="ts">
	/**
	 * 게시판 목록 화면 — 예전 board/BoardListPage.tsx(페이지 판). 검색 조건은 주소(?groupSubKey·keyword·pageNo·pageSize) 에 둔다(뒤로가기·공유).
	 * 쪽·검색·분류는 모두 서버가 한다 — 검색엔진(ES) 게시판은 getPosts, 그 밖은 Go v1 게시판 목록(getBoardPage, 2026-10-03).
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Pagination from '#lib/shared/mui/Pagination.svelte';
	import Select from '#lib/shared/mui/Select.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { board, findPostGroup, getBoardPage, getPosts } from './board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostList from './PostList.svelte';
	import WriteButton from './WriteButton.svelte';

	let { groupKey, path = '/board', isMobile, dateFormat = '' }: { groupKey: string; path?: string; isMobile: boolean; dateFormat?: string } = $props();

	const postGroup = $derived(findPostGroup(groupKey));
	const f = $derived(postGroup?.feature || {});

	type Search = { pageNo: number; pageSize: number; groupKey: string; groupSubKey: string; keyword: string; orderBy: string; orderByDesc: string; type: string };
	const search = $derived.by<Search>(() => {
		const q = page.url.searchParams;
		return {
			pageNo: parseInt(q.get('pageNo') || '', 10) || 1,
			pageSize: parseInt(q.get('pageSize') || '', 10) || f.pageSize || 10,
			groupKey: postGroup?.groupKey || groupKey,
			groupSubKey: (q.get('groupSubKey') || '').replace(/\+/g, ' '),
			keyword: q.get('keyword') || '',
			orderBy: q.get('orderBy') || 'postId',
			orderByDesc: q.get('orderByDesc') || 'desc',
			type: ''
		};
	});

	let list = $state<Post[] | null>(null);
	let totalCount = $state(0);
	let boardEl = $state<HTMLElement>();

	$effect(() => {
		const s = search;
		const g = postGroup;
		void board.version;
		if (!g) return;
		let stop = false;
		if (g.feature?.useES) {
			getPosts({ ...s, groupKey: g.groupKey }).then((o) => {
				if (stop) return;
				list = o.posts;
				totalCount = o.total;
			});
		} else {
			getBoardPage({ ...s, groupKey: g.groupKey }).then((o) => {
				if (stop) return;
				list = o.posts;
				totalCount = o.total;
			});
		}
		return () => (stop = true);
	});

	/** 조건 바꾸기 = 주소 바꾸기(목록은 주소를 보고 다시 읽는다) */
	function getData(s: Search, replace = true) {
		const q = new URLSearchParams();
		for (const [k, v] of Object.entries(s)) if (k != 'groupKey' && v) q.set(k, String(v));
		goto(`${page.url.pathname}?${q}`, { replace, reset: false });
	}

	function scrollToBoard() {
		if (!boardEl) return;
		window.scrollTo({ top: boardEl.getBoundingClientRect().top + window.scrollY - 100, behavior: 'smooth' });
	}
	const pageCount = $derived(Math.ceil((totalCount || 1) / search.pageSize));
	const pageSizes = [5, 10, 20, 50, 100]; // v1 목록은 한 쪽 최대 100
</script>

<Stack bind:el={boardEl} class="board" sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: 80 }}>
	{#if list}
		<PostList
			spacing={8}
			path={`${path}${groupKey ? `/${groupKey}` : ''}`}
			items={list}
			{postGroup}
			{totalCount}
			{search}
			{getData}
			{isMobile}
			{dateFormat}
			fixedPosts={board.fixedPostMap[String(groupKey).toUpperCase()] || []}
		/>
		<Stack class="paging" direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between', marginTop: 16 }}>
			{#if search.pageSize % 10 != 0 && search.pageSize % 5 != 0}
				<Stack />
			{:else}
				<Select
					class="page-size-select"
					value={search.pageSize || 10}
					options={pageSizes.map((v) => ({ value: v, label: String(v) }))}
					sx={{ width: 90, alignSelf: 'start', visibility: f.listTheme == 'card' || f.pageSize >= 1000 ? 'hidden' : 'visible' }}
					onchange={(v) => getData({ ...search, pageNo: 1, pageSize: v })}
				/>
			{/if}
			<Pagination
				page={search.pageNo}
				count={pageCount}
				siblingCount={3}
				onchange={(p) => {
					getData({ ...search, pageNo: p }, false);
					setTimeout(scrollToBoard, 100);
				}}
			/>
			<Stack sx={{ width: f.useWrite == true ? 'unset' : 90 }}>
				{#if f.useWrite == true}
					<Stack direction="row" class="write-btn-area" sx={{ justifyContent: 'end' }}>
						<WriteButton {postGroup} groupSubKey={search.groupSubKey} />
					</Stack>
				{/if}
			</Stack>
		</Stack>
	{/if}
</Stack>

<style>
	/* 페이지 번호 — 선택만 사이트 색(기본 모양, 게시판 디자인 CSS 로 덮는다) */
	:global(.board .MuiPaginationItem-root.Mui-selected) {
		background: #215953;
		color: #fff;
	}
</style>
