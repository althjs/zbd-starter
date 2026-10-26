<script lang="ts">
	/** MUI Menu(Popover) 호환 — anchor 요소 아래에 목록을 띄운다(body 에 그린다) */
	import type { Snippet } from 'svelte';
	import { tick } from 'svelte';
	import { portal } from '../ui/portal';
	import { cx, sx as toClass, type Sx } from './sx';

	let { anchor, open, onclose, paperSx = null, children }: { anchor: HTMLElement | null; open: boolean; onclose: () => void; paperSx?: Sx | null; children: Snippet } = $props();
	let paper = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0 });

	// 연 순간의 자리 — 올렸을 때만 보이는 단추(⋯)는 메뉴가 뜨면 hover 가 풀려 사라져 크기 0 이 된다
	let opened: DOMRect | null = null;
	$effect.pre(() => {
		opened = open && anchor ? anchor.getBoundingClientRect() : null;
	});
	function place() {
		if (!anchor) return;
		let r = anchor.getBoundingClientRect();
		if (!r.width && !r.height && opened) r = opened;
		const h = paper?.offsetHeight || 0;
		const w = paper?.offsetWidth || 0;
		pos = {
			top: Math.max(16, Math.min(r.bottom, window.innerHeight - h - 16)),
			left: Math.max(16, Math.min(r.left, window.innerWidth - w - 16))
		};
	}
	$effect(() => {
		if (!open || !anchor) return;
		tick().then(place);
	});
	// 안 내용이 바뀌어 크기가 달라지면(예: 반응 메뉴 → 이모지 고르기) 다시 화면 안으로
	$effect(() => {
		if (!open || !paper) return;
		const ro = new ResizeObserver(() => place());
		ro.observe(paper);
		return () => ro.disconnect();
	});
</script>

{#if open}
	<div use:portal role="presentation" class="MuiPopover-root MuiMenu-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={onclose}></div>
		<div
			bind:this={paper}
			class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper MuiMenu-paper', toClass(paperSx))}
			style="top:{pos.top}px;left:{pos.left}px"
		>
			<ul class="MuiList-root MuiList-padding MuiMenu-list" role="menu">
				{@render children()}
			</ul>
		</div>
	</div>
{/if}
<svelte:window onkeydown={(e) => open && e.key == 'Escape' && onclose()} />
