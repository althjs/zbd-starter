<script lang="ts">
	/**
	 * MUI Drawer(temporary) 호환 — 화면 가장자리에서 미끄러져 나오는 서랍. body 에 그린다(portal).
	 * DOM·클래스는 MUI 와 같게(`#id.MuiDrawer-root.MuiModal-root > .MuiBackdrop-root + .MuiDrawer-paper`) — 운영자 CSS 가 `#mobile-drawer .MuiDrawer-paper` 등을 겨냥한다.
	 * 열리면 문서 스크롤을 잠그고, Esc·바탕을 누르면 onclose.
	 * 내용이 넘치면 Scrollbar(simplebar) 막대로 — 브라우저 기본 막대 없음. 단 `scrollbar={false}` 면 종이(paper) 가 직접 스크롤
	 *   (운영 CSS 가 `#mobile-drawer .MuiDrawer-paper > .MuiStack-root` 처럼 종이 바로 아래를 겨냥하는 모바일 메뉴 서랍).
	 */
	import type { Snippet } from 'svelte';
	import { cubicOut } from 'svelte/easing';
	import { fade } from 'svelte/transition';
	import { portal } from '../ui/portal';
	import Scrollbar from '../ui/Scrollbar.svelte';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		open,
		anchor = 'right',
		id = undefined,
		class: cls = '',
		paperSx = null,
		scrollbar = true,
		onclose,
		children
	}: {
		open: boolean;
		anchor?: 'right' | 'left' | 'bottom';
		id?: string;
		class?: string;
		paperSx?: Sx | null;
		scrollbar?: boolean;
		onclose: () => void;
		children?: Snippet;
	} = $props();

	const A = $derived(anchor == 'right' ? 'Right' : anchor == 'bottom' ? 'Bottom' : 'Left');

	function slide(node: Element, { duration = 225 }) {
		if (anchor == 'bottom') return { duration, easing: cubicOut, css: (t: number) => `transform: translateY(${(1 - t) * 100}%)` };
		const sign = anchor == 'right' ? 1 : -1;
		return { duration, easing: cubicOut, css: (t: number) => `transform: translateX(${(1 - t) * 100 * sign}%)` };
	}

	$effect(() => {
		if (!open) return;
		const html = document.documentElement;
		const prev = { overflow: document.body.style.overflow, pr: document.body.style.paddingRight };
		const gap = window.innerWidth - html.clientWidth;
		document.body.style.overflow = 'hidden';
		if (gap > 0) document.body.style.paddingRight = `${gap}px`;
		const key = (e: KeyboardEvent) => e.key == 'Escape' && onclose();
		window.addEventListener('keydown', key);
		return () => {
			document.body.style.overflow = prev.overflow;
			document.body.style.paddingRight = prev.pr;
			window.removeEventListener('keydown', key);
		};
	});
</script>

{#if open}
	<div use:portal {id} role="presentation" class={cx(`MuiDrawer-root MuiDrawer-anchor${A} MuiDrawer-modal`, cls, 'MuiModal-root')}>
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true" onclick={onclose} transition:fade={{ duration: 225 }}></div>
		<div
			class={cx(`MuiPaper-root MuiPaper-elevation MuiPaper-elevation16 MuiDrawer-paper MuiDrawer-paperAnchor${A}`, scrollbar && 'zbd-sb-paper', toClass(paperSx))}
			transition:slide={{ duration: 225 }}
		>
			{#if scrollbar}
				<Scrollbar class="zbd-sb-fill">{@render children?.()}</Scrollbar>
			{:else}
				{@render children?.()}
			{/if}
		</div>
	</div>
{/if}
