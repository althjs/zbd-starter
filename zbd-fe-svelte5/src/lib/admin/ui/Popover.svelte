<script lang="ts">
	/**
	 * 팝오버 — 누른 요소(anchor) 아래에 종이를 띄운다(body 에 그림). 바깥 누르기·Esc 로 닫힌다.
	 *   <button onclick={(e) => (anchor = e.currentTarget)}>…</button>
	 *   <Popover {anchor} onclose={() => (anchor = null)} align="right"> 내용 </Popover>
	 */
	import type { Snippet } from 'svelte';
	import { tick } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';

	let {
		anchor,
		onclose,
		align = 'left',
		paperSx = null,
		children
	}: { anchor: HTMLElement | null; onclose: () => void; align?: 'left' | 'right' | 'center'; paperSx?: Sx | null; children?: Snippet } = $props();
	let paper = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0 });
	$effect(() => {
		if (!anchor) return;
		tick().then(() => {
			const r = anchor!.getBoundingClientRect();
			const w = paper?.offsetWidth || 0;
			const h = paper?.offsetHeight || 0;
			let left = align == 'right' ? r.right - w : align == 'center' ? r.left + r.width / 2 - w / 2 : r.left;
			left = Math.max(16, Math.min(left, window.innerWidth - w - 16));
			const top = Math.max(16, Math.min(r.bottom, window.innerHeight - h - 16));
			pos = { top, left };
		});
	});
</script>

<svelte:window onkeydown={(e) => anchor && e.key == 'Escape' && onclose()} />
{#if anchor}
	<div use:portal role="presentation" class="MuiPopover-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={onclose}></div>
		<div bind:this={paper} class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper', toClass(paperSx))} style="top:{pos.top}px;left:{pos.left}px">
			{@render children?.()}
		</div>
	</div>
{/if}
