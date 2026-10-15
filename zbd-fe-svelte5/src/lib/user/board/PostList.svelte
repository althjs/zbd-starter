<script lang="ts">
	/**
	 * 게시판 목록 — 제목 · 도구줄(분류 선택 · 총 건수 · 글쓰기) · 고정글 표 · 목록 표 · 빈 안내. 예전 board/PostList.tsx 중 즈비드가 쓰는 길
	 * (카드형·분류 탭·검색 폼·베스트 상자는 즈비드 게시판이 안 써서 옮기지 않았다).
	 */
	import { number } from '#lib/shared/format.ts';
	import Select from '#lib/shared/mui/Select.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import { findPostGroup } from './board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import ListCellsDefault from './ListCellsDefault.svelte';
	import ListCellsThumb from './ListCellsThumb.svelte';
	import WriteButton from './WriteButton.svelte';

	let {
		items,
		postGroup,
		path,
		totalCount,
		search,
		getData,
		isMobile,
		spacing = 8,
		dateFormat = '',
		fixedPosts = [],
		hiddenTitle = false
	}: {
		items: Post[];
		postGroup: any;
		path: string;
		totalCount: number;
		search: any;
		getData: (s: any) => void;
		isMobile: boolean;
		spacing?: number;
		dateFormat?: string;
		fixedPosts?: Post[];
		hiddenTitle?: boolean;
	} = $props();

	const f = $derived(postGroup?.feature || {});
	const listTheme = $derived(f.listTheme || '');

	/** 답글을 원글 바로 아래로(useReply) */
	const ordered = $derived.by(() => {
		const list = items || [];
		if (!f.useReply || !list.length) return list;
		const pid = (o: any) => {
			const ids = o?.json?.parentPostIds || [];
			return ids.length ? String(ids[ids.length - 1]) : '';
		};
		const byId = new Map(list.map((o) => [String(o.postId), o]));
		const kids = new Map<string, Post[]>();
		const roots: Post[] = [];
		for (const o of list) {
			const p = pid(o);
			if (p && byId.has(p)) (kids.get(p) || kids.set(p, []).get(p)!).push(o);
			else roots.push(o);
		}
		const out: Post[] = [];
		const walk = (o: Post, d: number) => {
			out.push(d > 0 ? { ...o, replyDepth: d } : o);
			for (const c of kids.get(String(o.postId)) || []) walk(c, d + 1);
		};
		roots.forEach((o) => walk(o, 0));
		return out;
	});

	const tableSx = toClass({
		borderTop: '1px solid #ebebeb',
		height: 'fit-content',
		width: '100%',
		borderCollapse: 'collapse',
		display: 'block',
		'& tbody': { display: 'block' },
		'& tr': { width: '100%', display: 'block' },
		'& th': { paddingTop: 8, paddingBottom: 8, paddingLeft: 4, paddingRight: 4, fontSize: '16px !important', height: 24, display: 'none' },
		'& td': { paddingTop: 4, paddingBottom: 4, paddingLeft: 4, paddingRight: 4, fontSize: 16, border: 0, verticalAlign: 'middle', height: 'auto' },
		[up('sm')]: {
			borderTop: '0',
			display: 'table',
			'& tbody': { display: 'table-cell' },
			'& tr': { display: 'table' },
			'& th': { display: 'table-cell' },
			'& td': { paddingTop: 8, paddingBottom: 8, height: '40px' }
		},
		[up('md')]: { '& th': { height: 40 } }
	});
	const rowSx = toClass({ borderBottom: '1px solid #eee', position: 'relative', marginTop: 8, [up('sm')]: { marginTop: 0 } });
	const groupOf = (o: Post) => {
		let g = findPostGroup(o.groupKey) || postGroup;
		if (postGroup?.feature) g = { ...g, feature: { ...g.feature, ...postGroup.feature } };
		return g;
	};
</script>

<Stack class={cx('bd-list', postGroup?.groupKey, listTheme)} sx={{ gap: spacing }}>
	{#if !hiddenTitle && !(postGroup?.design?.useHeaderBanenr && postGroup?.design?.headerBanner?.useTitle)}
		<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'} board-title`, toClass({ textAlign: 'center', marginBottom: 24 }))}>
			{postGroup?.name}
		</svelte:element>
	{/if}

	{#if (postGroup?.items?.length > 0 && f.useES) || totalCount > 0}
		<Stack class="list-toolbar" direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
			{#if postGroup?.items?.length > 0 && !f.useCategoryTab}
				<Stack class="sub-group-select">
					<Select
						value={search?.groupSubKey || ''}
						options={[{ value: '', label: '전체' }, ...(postGroup.items || []).map((s: any) => ({ value: s.groupSubKey, label: s.name || s.groupSubKey }))]}
						sx={{ minWidth: 150 }}
						onchange={(v) => getData({ ...search, groupSubKey: v, pageNo: 1 })}
					/>
				</Stack>
			{/if}
			<Stack class="list-actions" direction="row" sx={{ gap: 8, alignItems: 'center', marginLeft: 'auto' }}>
				{#if !f.useCategoryTab}
					<Stack class="summary-cnt" direction="row" sx={{ justifyContent: 'end' }}>
						<p class="MuiTypography-root MuiTypography-body1">총: {number(totalCount)}건</p>
					</Stack>
				{/if}
				{#if f.useWrite == true}
					<Stack direction="row" class="write-btn-area" sx={{ justifyContent: 'end' }}>
						<WriteButton {postGroup} groupSubKey={search?.groupSubKey || ''} />
					</Stack>
				{/if}
			</Stack>
		</Stack>
	{/if}

	{#if fixedPosts.length > 0}
		<Stack class="fixed-posts">
			<table class={cx('MuiTable-root fixed-table', tableSx)}>
				<tbody class="MuiTableBody-root">
					{#each fixedPosts as o (o.postId)}
						<tr class={cx('MuiTableRow-root post-row fixed-row', rowSx)}>
							<ListCellsDefault {o} {postGroup} isFixedPost={true} {path} {dateFormat} {isMobile} />
						</tr>
					{/each}
				</tbody>
			</table>
		</Stack>
	{/if}

	<table class={cx('MuiTable-root list-table', tableSx)}>
		<thead class={cx('MuiTableHead-root list-head', toClass({ fontSize: '16px' }))}>
			<tr class="MuiTableRow-root MuiTableRow-head">
				{#if !listTheme}
					<ListCellsDefault isHead={true} {dateFormat} {postGroup} {isMobile} />
				{/if}
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each ordered as o (o.postId)}
				<tr class={cx('MuiTableRow-root post-row', rowSx)}>
					{#if !listTheme}
						<ListCellsDefault {o} postGroup={groupOf(o)} path={path || `/board/${o.groupKey.toLowerCase()}`} {dateFormat} {isMobile} />
					{:else if listTheme == 'img-table'}
						<ListCellsThumb {o} postGroup={groupOf(o)} path={path || `/board/${o.groupKey.toLowerCase()}`} {isMobile} />
					{/if}
				</tr>
			{/each}
		</tbody>
	</table>

	{#if items && items.length == 0}
		<p class={cx('MuiTypography-root MuiTypography-body2 no-items', toClass({ textAlign: 'center', paddingTop: 40, paddingBottom: 40 }))}>등록된 게시글이 없습니다.</p>
	{/if}
</Stack>

<style>
	/* 기본 모양(게시판마다 디자인 CSS 로 덮는다 — 그래서 최소만) */
	:global(.bd-list .list-table) {
		border: 1px solid #e3e8e7;
		border-radius: 14px;
		border-collapse: separate;
		border-spacing: 0;
		overflow: hidden;
	}
	:global(.bd-list .list-table .cell-date) {
		white-space: nowrap;
	}
	:global(.bd-list .list-table th) {
		background: #f6f9f8;
		color: #8a9796;
		font-size: 12px;
		padding: 10px 16px;
		white-space: nowrap;
	}
	:global(.bd-list .list-table td) {
		padding: 16px;
		border-bottom: 1px solid #f1f4f3;
		font-size: 13px;
		color: #8a9796;
	}
	:global(.bd-list .list-table .cell-title a),
	:global(.bd-list .list-table .cell-title a *) {
		font-size: 15px;
		font-weight: 600;
		color: #1f2a2b;
	}
</style>
