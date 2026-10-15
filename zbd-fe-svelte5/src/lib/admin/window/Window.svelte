<script lang="ts">
	/**
	 * 화면 위 창 하나(React resizable-modal/ResizableModal + ResizableWindowContainer) — 제목줄 끌기 · 8방향 크기 조절 · 최소화/최대화/닫기.
	 * 제목줄 두 번 누르기 = 최대화 토글 · Ctrl+Shift+화살표 = 반쪽 붙이기(맨 앞 창만).
	 * 끌거나 크기를 바꾸는 동안 모든 iframe 의 pointer-events 를 끈다(iframe 위로 지나가면 이벤트가 끊긴다).
	 * 휴대폰(<900)은 늘 화면 전체(끌기·크기 조절 없음), 뒤로가기 = 닫기.
	 */
	import type { Snippet } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { useBackToClose } from '../backToClose.svelte.ts';
	import { media } from '../media.svelte.ts';
	import Icon from '../ui/Icon.svelte';
	import IconButton from '../ui/IconButton.svelte';
	import Tooltip from '../ui/Tooltip.svelte';
	import { bringToFront, HEADER_HEIGHT, MIN_SIZE, minimize, requestCloseWindow, saveGeom, snap, toggleMaximize, topId, unsnap, zOf, type Win } from './windows.svelte.ts';

	let { win, children }: { win: Win; children: Snippet } = $props();

	const mobile = $derived(media.mdDown);
	const maxed = $derived(win.view === 'maximized' || mobile);
	const minimized = $derived(win.view === 'minimized');
	const z = $derived(zOf(win.id));

	useBackToClose(
		() => media.mdDown && win.view !== 'minimized',
		() => requestCloseWindow(win.id)
	);

	const setIframes = (on: boolean) => document.querySelectorAll('iframe').forEach((f) => (f.style.pointerEvents = on ? '' : 'none'));

	/** 제목줄 끌기 — pointerdown 뒤 window 에서 move/up 을 받는다 */
	function startMove(e: PointerEvent) {
		if (e.button !== 0 || maxed || (e.target as HTMLElement).closest('button,a')) return;
		if (unsnap(win)) return; // 스냅 상태면 원래 자리로만(React 와 같게 — 이번 끌기는 취소)
		e.preventDefault();
		const sx0 = e.clientX;
		const sy0 = e.clientY;
		const p0 = { ...win.pos };
		setIframes(false);
		const mv = (ev: PointerEvent) => {
			const W = window.innerWidth;
			const H = window.innerHeight;
			const x = Math.min(W - 50, Math.max(-win.size.width + 50, p0.x + ev.clientX - sx0));
			const y = Math.min(H - 50, Math.max(-win.size.height + 50, p0.y + ev.clientY - sy0));
			win.pos = { x, y };
		};
		const end = () => {
			window.removeEventListener('pointermove', mv);
			window.removeEventListener('pointerup', end);
			window.removeEventListener('pointercancel', end);
			setIframes(true);
			saveGeom(win);
		};
		window.addEventListener('pointermove', mv);
		window.addEventListener('pointerup', end);
		window.addEventListener('pointercancel', end);
	}

	type Dir = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw';
	const DIRS: Dir[] = ['s', 'w', 'e', 'n', 'sw', 'nw', 'se', 'ne'];
	function startResize(e: PointerEvent, d: Dir) {
		if (e.button !== 0 || maxed) return;
		e.preventDefault();
		e.stopPropagation();
		if (unsnap(win)) return;
		const sx0 = e.clientX;
		const sy0 = e.clientY;
		const p0 = { ...win.pos };
		const s0 = { ...win.size };
		setIframes(false);
		const clampW = (v: number) => Math.min(3000, Math.max(MIN_SIZE.width, v));
		const clampH = (v: number) => Math.min(3000, Math.max(MIN_SIZE.height, v));
		const mv = (ev: PointerEvent) => {
			const dx = ev.clientX - sx0;
			const dy = ev.clientY - sy0;
			let { width, height } = s0;
			let { x, y } = p0;
			if (d.includes('e')) width = clampW(s0.width + dx);
			if (d.includes('w')) ((width = clampW(s0.width - dx)), (x = p0.x + (s0.width - width)));
			if (d.includes('s')) height = clampH(s0.height + dy);
			if (d.includes('n')) ((height = clampH(s0.height - dy)), (y = p0.y + (s0.height - height)));
			if (y < 0) y = 0; // 제목줄이 화면 위로 나가지 않게
			win.size = { width, height };
			win.pos = { x, y };
		};
		const end = () => {
			window.removeEventListener('pointermove', mv);
			window.removeEventListener('pointerup', end);
			window.removeEventListener('pointercancel', end);
			setIframes(true);
			saveGeom(win);
		};
		window.addEventListener('pointermove', mv);
		window.addEventListener('pointerup', end);
		window.addEventListener('pointercancel', end);
	}

	// 맨 앞 창만 Ctrl+Shift+화살표
	$effect(() => {
		const h = (e: KeyboardEvent) => {
			if (!(e.ctrlKey && e.shiftKey) || mobile || topId() !== win.id) return;
			if (e.key === 'ArrowLeft' || e.key === 'ArrowRight' || e.key === 'ArrowUp' || e.key === 'ArrowDown') {
				e.preventDefault();
				snap(win.id, e.key);
			}
		};
		window.addEventListener('keydown', h);
		return () => window.removeEventListener('keydown', h);
	});

	const box = $derived(
		maxed
			? `left:0;top:0;width:100vw;height:100dvh;border-radius:0;box-shadow:none;border:none`
			: `left:${win.pos.x}px;top:${win.pos.y}px;width:${win.size.width}px;height:${win.size.height + HEADER_HEIGHT}px`
	);
</script>

<!-- 창 한 겹 — body 바로 아래, z-index 1250 + 순번(1299 까지) -->
<div use:portal class="zw-layer" style="z-index:{z};display:{minimized ? 'none' : 'block'}" data-window-id={win.id} onpointerdowncapture={() => bringToFront(win.id)}>
	<div class="zw-win" class:maxed role="dialog" aria-label={win.title} style={box}>
		<div class="zw-head modal-draggable-header" role="toolbar" tabindex="-1" style:cursor={maxed ? 'default' : 'move'} onpointerdown={startMove} ondblclick={() => !mobile && toggleMaximize(win.id)}>
			<div class="zw-title-wrap"><h6 class="MuiTypography-root MuiTypography-h6 zw-title">{win.title}</h6></div>
			<div class="zw-btns">
				{#each win.actions || [] as a (a.title)}
					<Tooltip placement="top" title={a.title}><IconButton size="Small" sx={{ padding: '4px' }} aria-label={a.title} onclick={a.onclick}><Icon name={a.icon} size="Small" /></IconButton></Tooltip>
				{/each}
				<Tooltip placement="top" title="최소화"
					><IconButton size="Small" sx={{ padding: '4px' }} aria-label="최소화" onclick={() => minimize(win.id)}><Icon name="remove" size="Small" /></IconButton></Tooltip
				>
				{#if !mobile}
					<Tooltip placement="top" title={win.view === 'maximized' ? '이전 크기로 복원' : '최대화'}
						><IconButton size="Small" sx={{ padding: '4px' }} aria-label={win.view === 'maximized' ? '이전 크기로 복원' : '최대화'} onclick={() => toggleMaximize(win.id)}
							><Icon name={win.view === 'maximized' ? 'filter_none' : 'crop_square'} size="Small" /></IconButton
						></Tooltip
					>
				{/if}
				<Tooltip placement="top" title="닫기"
					><IconButton size="Small" sx={{ padding: '4px' }} aria-label="닫기" onclick={(e: MouseEvent) => (e.stopPropagation(), requestCloseWindow(win.id))}
						><Icon name="close" size="Small" /></IconButton
					></Tooltip
				>
			</div>
		</div>
		<div class="zw-body">{@render children()}</div>
		{#if !maxed}
			{#each DIRS as d (d)}
				<span class="zw-h zw-h-{d}" role="presentation" onpointerdown={(e) => startResize(e, d)}></span>
			{/each}
		{/if}
	</div>
</div>

<style>
	.zw-layer {
		position: fixed;
		top: 0;
		left: 0;
		width: 0;
		height: 0;
		overflow: visible;
	}
	.zw-win {
		position: fixed;
		box-sizing: content-box; /* React 처럼 내용 크기 + 테두리 */
		display: flex;
		flex-direction: column;
		box-shadow: 0px 4px 20px rgba(0, 0, 0, 0.15);
		border: 1px solid var(--zc-e0e0e0, #e0e0e0);
		border-radius: 8px;
		background-color: var(--zc-white, white);
		overflow: hidden;
		pointer-events: auto;
	}
	.zw-head {
		flex: 0 0 auto;
		box-sizing: border-box;
		height: 48px;
		padding: 8px 16px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--zc-f5f5f5, #f5f5f5);
		border-bottom: 1px solid var(--zc-ddd, #ddd);
		user-select: none;
		touch-action: none;
		position: relative;
		z-index: 2;
	}
	.zw-title-wrap {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.zw-title {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.zw-btns {
		display: flex;
		gap: 4px;
		flex-shrink: 0;
	}
	.zw-body {
		flex: 1 1 auto;
		min-height: 0;
		overflow: auto;
		position: relative;
	}
	/* 크기 조절 손잡이 — React react-resizable 과 같은 두께(가장자리 6 · 모서리 16) */
	.zw-h {
		position: absolute;
		z-index: 3;
		touch-action: none;
	}
	.zw-h-n,
	.zw-h-s {
		left: 0;
		width: 100%;
		height: 6px;
	}
	.zw-h-n {
		top: 0;
		cursor: n-resize;
	}
	.zw-h-s {
		bottom: 0;
		cursor: s-resize;
	}
	.zw-h-e,
	.zw-h-w {
		top: 0;
		height: 100%;
		width: 6px;
	}
	.zw-h-e {
		right: 0;
		cursor: e-resize;
	}
	.zw-h-w {
		left: 0;
		cursor: w-resize;
	}
	.zw-h-ne,
	.zw-h-nw,
	.zw-h-se,
	.zw-h-sw {
		width: 16px;
		height: 16px;
		z-index: 4;
	}
	.zw-h-ne {
		top: 0;
		right: 0;
		cursor: ne-resize;
	}
	.zw-h-nw {
		top: 0;
		left: 0;
		cursor: nw-resize;
	}
	.zw-h-se {
		bottom: 0;
		right: 0;
		cursor: se-resize;
	}
	.zw-h-sw {
		bottom: 0;
		left: 0;
		cursor: sw-resize;
	}
</style>
