<script lang="ts">
	/**
	 * 말풍선 — 감싼 것에 마우스를 올리면 title 을 띄운다(MUI Tooltip 모양). 휴대폰(터치)은 띄우지 않는다.
	 *   <Tooltip title="도움말" placement="left"><IconButton …/></Tooltip>
	 */
	import type { Snippet } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';

	let {
		title,
		placement = 'bottom',
		zIndex = undefined,
		children
	}: { title: string; placement?: 'top' | 'bottom' | 'left' | 'right'; /** 기본 1500 — 더 높은 층(창 목록 2100 등) 위에 띄울 때 */ zIndex?: number; children: Snippet } = $props();
	let host = $state<HTMLElement>();
	let tip = $state<HTMLElement>();
	let show = $state(false);
	let pos = $state({ top: 0, left: 0 });
	let timer: any;

	function enter(e: PointerEvent) {
		if (e.pointerType == 'touch' || !title) return;
		clearTimeout(timer);
		timer = setTimeout(async () => {
			show = true;
			await Promise.resolve();
			requestAnimationFrame(place);
		}, 100);
	}
	function leave() {
		clearTimeout(timer);
		show = false;
	}
	function place() {
		const el = host?.firstElementChild as HTMLElement | null;
		if (!el || !tip) return;
		const r = el.getBoundingClientRect();
		const w = tip.offsetWidth,
			h = tip.offsetHeight;
		const gap = 14;
		let top = r.bottom + gap,
			left = r.left + r.width / 2 - w / 2;
		if (placement == 'top') top = r.top - h - gap;
		if (placement == 'bottom') top = r.bottom + gap;
		if (placement == 'left') ((top = r.top + r.height / 2 - h / 2), (left = r.left - w - gap));
		if (placement == 'right') ((top = r.top + r.height / 2 - h / 2), (left = r.right + gap));
		pos = { top: Math.max(4, top), left: Math.max(4, Math.min(left, window.innerWidth - w - 4)) };
	}
</script>

<span bind:this={host} class="adm-tip-host" role="presentation" onpointerover={(e) => !show && enter(e)} onpointerout={(e) => !host?.contains(e.relatedTarget as Node) && leave()} onpointerdown={leave}
	>{@render children()}</span
>
{#if show}
	<div use:portal bind:this={tip} class="MuiTooltip-popper adm-tip" role="tooltip" style="top:{pos.top}px;left:{pos.left}px" style:z-index={zIndex}>
		<div class="MuiTooltip-tooltip">{title}</div>
	</div>
{/if}

<style>
	.adm-tip-host {
		display: contents;
	}
	.adm-tip {
		position: fixed;
		z-index: 1500;
		pointer-events: none;
	}
</style>
