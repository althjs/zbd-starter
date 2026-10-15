<script lang="ts">
	/** 글 요약 줄(React board/components/PostSummaryInfo) — 작성자·등록일·조회수·댓글·좋아요·싫어요·별점 */
	import { date, number } from '#lib/shared/format.ts';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import type { Post, PostGroup } from '#lib/shared/types/post.ts';

	let { o, postGroup, onmember, sx = null }: { o: Post; postGroup: PostGroup | null; onmember?: (userId: number) => void; sx?: Sx | null } = $props();
	const writer = $derived((o as any)?.[postGroup?.dataColumn || 'json']?.name || '');
	const t = cx('MuiTypography-root MuiTypography-body2', toClass({ color: 'var(--adm-text)', letterSpacing: 'normal' }));
	const ti = cx('MuiTypography-root MuiTypography-body2', toClass({ color: 'var(--adm-text)', letterSpacing: 'normal', display: 'flex', flexDirection: 'row', alignItems: 'center' }));
	const ic = { fontSize: 18, color: '#aaa', marginRight: 1.6 };
	const ra = $derived(o.json?.reaction?.RA || 0);
</script>

<div class={cx('MuiStack-root summary', toClass({ flexDirection: 'row', flexWrap: 'wrap', gap: 8, alignItems: 'center', ...(sx || {}) }))}>
	{#if onmember}
		<IconButton class="member-manage-btn" size="Small" sx={{ marginRight: -8 }} aria-label="회원 보기" onclick={() => onmember(o.userId as number)}
			><Icon name="person_search" size="Small" /></IconButton
		>
	{/if}
	<p class={cx(t, 'writer')}>
		<span>{'작성자: '}</span>{#if writer.includes('<mark>')}{@html o.jsonUser.name}{:else}{writer || '***'}{/if}
	</p>
	<p class={cx(t, 'created-at')}><span>{'등록일: '}</span>{date(o.createdAt)}</p>
	<p class={cx(ti, 'view-cnt')}><Icon name="remove_red_eye" sx={ic} /><span>{'조회수: '}</span>{number(o.json?.viewCount || 0)}</p>
	{#if postGroup?.feature?.useComment && (o.json?.commentCount || 0) > 0}
		<p class={cx(ti, 'comment-cnt')}><Icon name="chat" sx={ic} /><span>{'코맨트: '}</span>{number(o.json?.commentCount || 0)}</p>
	{/if}
	{#if postGroup?.feature?.useLike && (o.json?.reaction?.LL || 0) > 0}
		<p class={cx(ti, 'like-cnt')}><Icon name="thumb_up" sx={ic} /><span>{'좋아요: '}</span>{number(o.json?.reaction?.LL || 0)}</p>
	{/if}
	{#if postGroup?.feature?.useComment && (o.json?.reaction?.LH || 0) > 0}
		<p class={cx(ti, 'like-cnt dislike-cnt')}><Icon name="thumb_down" sx={ic} /><span>{'싫어요: '}</span>{number(o.json?.reaction?.LH || 0)}</p>
	{/if}
	{#if postGroup?.feature?.useStarRating && ra > 0}
		<span class="MuiRating-root MuiRating-sizeSmall Mui-disabled rating" aria-label={`${ra}점`} style="display:inline-flex;font-size:1.125rem;color:#faaf00;opacity:0.5">
			{#each [1, 2, 3, 4, 5] as i (i)}<span>{ra >= i ? '★' : ra >= i - 0.5 ? '⯪' : '☆'}</span>{/each}
		</span>
	{/if}
</div>
