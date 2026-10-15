<script lang="ts">
	/**
	 * 관리자 카드(MUI Card — 모서리 12 · 옅은 그림자 · 흰 바탕, 다크면 어두운 바탕). 공용 shared/ui/Card 는 사용자 화면 크기라 관리자는 이걸.
	 *   <Card sx={{ padding: 24 }}>…</Card>   <Card title="실시간 회원 현황">…</Card>
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { title = '', sx = null, class: cls = '', children, action }: { title?: string; sx?: Sx | null; class?: string; children?: Snippet; action?: Snippet } = $props();
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', cls, toClass(sx))}>
	{#if title || action}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }))}>
			<h6 class="MuiTypography-root MuiTypography-h6">{title}</h6>
			{@render action?.()}
		</div>
	{/if}
	{@render children?.()}
</div>
