<script lang="ts">
	/** 글 요약 줄 — 작성자 · 등록일 · 조회수 · 댓글 · 좋아요 · 별점. 예전 list-theme/PostSummaryInfo.tsx */
	import { date, dateByToday, number } from '#lib/shared/format.ts';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import Rating from './Rating.svelte';

	let {
		o,
		postGroup,
		isList = false,
		hiddenWriter = false,
		hiddenCreatedAt = false,
		dateFormat = '',
		showEmpty = false,
		sx = null,
		onclick
	}: { onclick?: () => void; o: Post; postGroup: any; isList?: boolean; hiddenWriter?: boolean; hiddenCreatedAt?: boolean; dateFormat?: string; showEmpty?: boolean; sx?: Sx | null } = $props();

	const writer = $derived(o?.[postGroup?.dataColumn || 'json']?.name || '');
	const t = 'MuiTypography-root MuiTypography-body2';
	const gray = '';
	const row = toClass({ display: 'flex', flexDirection: 'row', alignItems: 'center' });
	const ic = { fontSize: '18px', color: '#aaa', marginRight: '1.6px' };
	const f = $derived(postGroup?.feature || {});
	const r = $derived(o.json?.reaction || {});
</script>

<Stack class="summary" direction="row" onclick={() => onclick?.()} sx={{ flexWrap: 'wrap', gap: 8, alignItems: 'center', ...(sx || {}) }}>
	{#if !hiddenWriter}
		<p class={cx(t, 'writer', gray)}>
			<span>{'작성자: '}</span>{#if writer.includes('<mark>')}<span style="display:inline">{@html o.jsonUser?.name}</span>{:else}{writer || '***'}{/if}
		</p>
	{/if}
	{#if !hiddenCreatedAt}
		<p class={cx(t, 'created-at', gray)}>
			<span>{'등록일: '}</span>{dateFormat === 'yyyy.MM.dd' ? String(date(o.createdAt)).split('(')[0] : isList ? dateByToday(o.createdAt) : date(o.createdAt)}
		</p>
	{/if}
	<p class={cx(t, 'view-cnt', row)}>
		<Icon name="remove_red_eye" sx={ic} /><span>{'조회수: '}</span>{number(o.json?.viewCount || 0)}
	</p>
	{#if f.useComment && ((o.json?.commentCount || 0) > 0 || showEmpty)}
		<p class={cx(t, 'comment-cnt', row)}><Icon name="chat" sx={ic} /><span>{'코맨트: '}</span>{number(o.json?.commentCount || 0)}</p>
	{/if}
	{#if f.useLike && (r.LL > 0 || showEmpty)}
		<p class={cx(t, 'like-cnt', row)}><Icon name="thumb_up" sx={ic} /><span>{'좋아요: '}</span>{number(r.LL || 0)}</p>
	{/if}
	{#if f.useComment && r.LH > 0}
		<p class={cx(t, 'like-cnt dislike-cnt', row)}><Icon name="thumb_down" sx={ic} /><span>{'싫어요: '}</span>{number(r.LH || 0)}</p>
	{/if}
	{#if f.useStarRating && (r.RA > 0 || showEmpty)}
		<Rating value={r.RA || 0} />
	{/if}
</Stack>
