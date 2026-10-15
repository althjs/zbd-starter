<script lang="ts">
	/**
	 * 기본 목록 줄(표 칸) — No · 제목 · (작성자) · 등록일 · (댓글수) · 조회수(또는 처리 상태). 예전 list-theme/PostListDefaultCells.tsx
	 * 휴대폰은 칸을 세로로 쌓는다(번호·조회수 숨김).
	 */
	import { date, number } from '#lib/shared/format.ts';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import { session } from '../session.svelte.ts';
	import { board, postHref } from './board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import TitleView from './TitleView.svelte';

	let {
		isHead = false,
		o = null as any,
		path = '',
		postGroup,
		isFixedPost = false,
		dateFormat = '',
		isMobile
	}: { isHead?: boolean; o?: Post | null; path?: string; postGroup: any; isFixedPost?: boolean; dateFormat?: string; isMobile: boolean } = $props();

	const f = $derived(postGroup?.feature || {});
	const blocked = $derived(!!f.useOwnerOnly && (!session.profile?.userId || session.profile.userId != o?.userId));
	const showNo = $derived(!isMobile && !f.useES);
	const shortDate = $derived(isFixedPost || dateFormat == 'yyyy.MM.dd');
	const head = (extra = '') => cx('MuiTableCell-root MuiTableCell-head', extra, 'MuiTableCell-sizeMedium');
	const body = (extra = '') => cx('MuiTableCell-root MuiTableCell-body', extra, 'MuiTableCell-sizeMedium');
	const h6 = 'MuiTypography-root MuiTypography-subtitle2';
	const fs16 = toClass({ fontSize: '16px' });
	/** 칸 너비 — MUI TableCell 처럼 width 속성으로(구체성 없는 표시 힌트라 운영자 CSS 가 덮는다) */
	const w = (n: number) => ({ width: n }) as Record<string, any>;
	const block = toClass({ display: 'block', [up('sm')]: { display: 'table-cell' } });
</script>

{#if isHead}
	{#if showNo}
		<th class={cx(head('MuiTableCell-alignCenter'), 'cell-no', fs16)} scope="col" {...w(100)}><h6 class={h6}>No</h6></th>
	{/if}
	<th class={cx(head(), 'cell-title', toClass({ textAlign: 'center', fontSize: '16px' }))} scope="col"><h6 class={h6}>제목</h6></th>
	{#if !isFixedPost && f.listThemeOptions?.useWriter}
		<th class={cx(head('MuiTableCell-alignCenter'), 'cell-writer', fs16)} scope="col" {...w(120)}><h6 class={h6}>작성자</h6></th>
	{/if}
	{#if !isFixedPost && f.listThemeOptions?.usePostSummary}
		<th class={cx(head(), 'cell-summary')} scope="col"></th>
	{/if}
	{#if isFixedPost || !f.listThemeOptions?.usePostSummary}
		<th class={cx(head('MuiTableCell-alignCenter'), 'cell-date', fs16)} scope="col" {...w(shortDate ? 100 : 150)}><h6 class={h6}>등록일</h6></th>
		{#if f.useReaction && f.useComment}
			<th class={cx(head('MuiTableCell-alignCenter'), 'cell-comment-cnt', fs16)} scope="col" {...w(60)}><h6 class={h6}>댓글수</h6></th>
		{/if}
		<th class={cx(head('MuiTableCell-alignCenter'), 'cell-view-cnt', fs16)} scope="col" {...w(60)}><h6 class={h6}>{postGroup?.component?.title || '조회수'}</h6></th>
	{/if}
{:else if o}
	{#if showNo}
		<td class={cx(body('MuiTableCell-alignCenter'), 'cell-no')} {...w(100)}>{o.postNo || o.postId}</td>
	{/if}
	<td class={cx(body(), 'cell-title', block)}>
		{#if blocked}
			<div class={cx('post-link blocked MuiBox-root', toClass({ fontSize: '14px', fontWeight: 500, color: '#7e7e7e', [up('md')]: { fontSize: '16px' } }))}>
				<TitleView isWidget={!isFixedPost} post={o} />
			</div>
		{:else}
			<a
				class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone post-link', toClass({ fontSize: '14px', fontWeight: '500', [up('md')]: { fontSize: '16px' } }))}
				href={postHref(path, o)}><TitleView isWidget={!isFixedPost} post={o} /></a
			>
		{/if}
	</td>
	{#if !isFixedPost && f.listThemeOptions?.useWriter}
		<td class={cx(body('MuiTableCell-alignCenter'), 'cell-writer', toClass({ display: 'none', [up('sm')]: { display: 'table-cell' } }))} {...w(120)}>
			<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ fontSize: '13px', [up('md')]: { fontSize: '14px' } }))}>
				{o?.nickname || o?.[postGroup?.dataColumn || 'json']?.name || '-'}
			</p>
		</td>
	{/if}
	{#if isFixedPost || !f.listThemeOptions?.usePostSummary}
		<td
			class={cx(
				body(shortDate ? 'MuiTableCell-alignCenter' : 'MuiTableCell-alignLeft'),
				'cell-date',
				toClass({ color: '#999', display: 'block', [up('sm')]: { color: '#000', display: 'table-cell' } })
			)}
			width={shortDate ? 100 : 150}
		>
			<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ fontSize: '13px', [up('md')]: { fontSize: '14px' } }))}>
				{shortDate ? String(date(o.createdAt)).split('(')[0] : date(o.createdAt)}
			</p>
		</td>
		{#if f.useReaction && f.useComment}
			<td class={cx(body('MuiTableCell-alignCenter'), 'cell-comment-cnt', fs16)} {...w(60)}>
				<h6 class={'MuiTypography-root MuiTypography-subtitle2'}>{number(board.summary?.commentMap?.[o.postId] || o?.json?.commentCount || 0)}</h6>
			</td>
		{/if}
		<td class={cx(body(), 'cell-view-cnt', toClass({ display: 'none', textAlign: 'left', [up('sm')]: { display: 'table-cell', textAlign: 'center' } }))} {...w(60)}>
			<p class={'MuiTypography-root MuiTypography-body2'}>
				{#if postGroup?.component?.processMap}
					{postGroup.component.processMap[o?.[postGroup?.dataColumn]?.processStatus || 0]}
				{:else}
					{number(board.summary?.pvMap?.[o.postId] || o?.json?.viewCount || 0)}
				{/if}
			</p>
		</td>
	{/if}
{/if}
