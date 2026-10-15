<script lang="ts">
	/**
	 * 글 제목 — 예전 board/TitleView.tsx. 제목은 운영자가 쓴 HTML 일 수 있어 그대로(목록·상세 공용).
	 * isWidget 이면 분류(groupSubKey) 를 앞에 · 위에 고정한 글은 「공지」 칩.
	 */
	import Chip from '#lib/shared/mui/Chip.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let { post, variant = '', sx = null, isWidget = false }: { post: Post; variant?: string; sx?: Sx | null; isWidget?: boolean } = $props();
</script>

<Stack class="title">
	<div
		class={cx(
			'MuiTypography-root',
			variant ? `MuiTypography-${variant}` : 'MuiTypography-body1',
			'title-text',
			toClass({ display: 'flex', flexDirection: 'row', alignItems: 'center', ...(sx || {}) })
		)}
	>
		{#if post?.groupSubKey && isWidget}
			<p class="MuiTypography-root MuiTypography-body1 groupSubKey" style="margin-right:8px">{post.groupSubKey}</p>
		{/if}
		{#if post?.json?.isFixTop}
			<Chip class="state-chip" size="small" label="공지" sx={{ marginRight: 8 }} />
		{/if}
		<Stack class="text" direction="row" sx={{ flex: 1, alignItems: 'center' }}>
			<div class={cx('title-html MuiBox-root', toClass({ '& .emoji': { display: 'inline-block', width: '1.3em', height: '1.3em', marginBottom: '-5px' }, ...(sx || {}) }))}>
				{@html post?.title || ''}
			</div>
		</Stack>
	</div>
</Stack>
