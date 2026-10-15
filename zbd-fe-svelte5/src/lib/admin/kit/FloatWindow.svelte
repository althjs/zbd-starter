<script lang="ts">
	/**
	 * 화면 위에 뜨는 창(React common/resizable-modal/ResizableWindowContainer 의 줄인 판) — 48px 제목줄을 끌어 옮기고, 가장자리·모서리로 크기 조절, 제목줄 두 번 = 최대화.
	 * 자리·크기는 localStorage `zbd:admin:window:{windowId}` {position,size}. 최소화 독(dock)·창 붙이기(snap)는 옮기지 않았다.
	 * 바탕 README 의 「창 시스템 — 아직 없음」 자리 — 공용 후보.
	 */
	import type { Snippet } from 'svelte';
	import Icon from '#lib/admin/ui/Icon.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	const HEADER = 48;
	let {
		windowId,
		windowTitle = '팝업',
		x = 100,
		y = 100,
		width = 400,
		height = 500,
		minWidth = 200,
		minHeight = 150,
		windowSx = null,
		onclose,
		actions,
		children
	}: {
		windowId: string;
		windowTitle?: string;
		x?: number;
		y?: number;
		width?: number;
		height?: number;
		minWidth?: number;
		minHeight?: number;
		windowSx?: Sx | null;
		/** 주면 제목줄에 ✕ (2026-10-03) */
		onclose?: () => void;
		/** 제목줄 오른쪽(최소화 앞) 버튼 자리 — 예: JSON 편집 */
		actions?: Snippet;
		children?: Snippet;
	} = $props();
	const KEY = SK.admin.window(windowId);
	let pos = $state({ x, y });
	let size = $state({ width, height });
	let view = $state<'normal' | 'maximized' | 'minimized'>('normal');
	try {
		const s = JSON.parse(localStorage.getItem(KEY) || 'null');
		if (s?.position) pos = s.position;
		if (s?.size) size = s.size;
	} catch {}
	const save = () => {
		try {
			if (view == 'normal') localStorage.setItem(KEY, JSON.stringify({ position: pos, size }));
		} catch {}
	};
	const frames = (on: boolean) => document.querySelectorAll('iframe').forEach((f) => (f.style.pointerEvents = on ? 'auto' : 'none'));
	function track(e: PointerEvent, fn: (dx: number, dy: number) => void) {
		if (view == 'maximized') return;
		e.preventDefault();
		const x0 = e.clientX,
			y0 = e.clientY;
		frames(false);
		const mv = (ev: PointerEvent) => fn(ev.clientX - x0, ev.clientY - y0);
		const up = () => {
			window.removeEventListener('pointermove', mv);
			window.removeEventListener('pointerup', up);
			frames(true);
			save();
		};
		window.addEventListener('pointermove', mv);
		window.addEventListener('pointerup', up);
	}
	function move(e: PointerEvent) {
		const p0 = { ...pos };
		track(e, (dx, dy) => (pos = { x: Math.max(0, Math.min(window.innerWidth - 100, p0.x + dx)), y: Math.max(0, Math.min(window.innerHeight - HEADER, p0.y + dy)) }));
	}
	function resize(e: PointerEvent, h: string) {
		e.stopPropagation();
		const p0 = { ...pos },
			s0 = { ...size };
		track(e, (dx, dy) => {
			let w = s0.width,
				hh = s0.height,
				nx = p0.x,
				ny = p0.y;
			if (h.includes('e')) w = s0.width + dx;
			if (h.includes('s')) hh = s0.height + dy;
			if (h.includes('w')) w = s0.width - dx;
			if (h.includes('n')) hh = s0.height - dy;
			w = Math.max(minWidth, w);
			hh = Math.max(minHeight, hh);
			if (h.includes('w')) nx = p0.x + (s0.width - w);
			if (h.includes('n')) ny = Math.max(0, p0.y + (s0.height - hh));
			pos = { x: nx, y: ny };
			size = { width: w, height: hh };
		});
	}
	const max = $derived(view == 'maximized');
	const HANDLES: [string, Sx][] = [
		['s', { bottom: 0, left: 0, height: 6, width: '100%', cursor: 's-resize' }],
		['n', { top: 0, left: 0, height: 6, width: '100%', cursor: 'n-resize' }],
		['e', { right: 0, top: 0, width: 6, height: '100%', cursor: 'e-resize' }],
		['w', { left: 0, top: 0, width: 6, height: '100%', cursor: 'w-resize' }],
		['se', { right: 0, bottom: 0, width: 16, height: 16, cursor: 'se-resize', zIndex: 101 }],
		['sw', { left: 0, bottom: 0, width: 16, height: 16, cursor: 'sw-resize', zIndex: 101 }],
		['ne', { right: 0, top: 0, width: 16, height: 16, cursor: 'ne-resize', zIndex: 101 }],
		['nw', { left: 0, top: 0, width: 16, height: 16, cursor: 'nw-resize', zIndex: 101 }]
	];
	const btn = cx(
		'MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeSmall',
		toClass({
			padding: 4,
			border: 0,
			background: 'none',
			cursor: 'pointer',
			borderRadius: '50%',
			color: 'var(--adm-action-active, rgba(0,0,0,0.54))',
			display: 'inline-flex',
			'&:hover': { backgroundColor: 'rgba(0,0,0,0.04)' }
		})
	);
</script>

<div
	class={toClass({ display: 'inline-block', position: 'absolute', top: 0, left: 0, ...(windowSx || {}), ...(max ? { borderRadius: 0, boxShadow: 'none', border: 'none' } : {}) })}
	style="transform:translate({max ? 0 : pos.x}px, {max ? 0 : pos.y}px)"
>
	<div
		role="presentation"
		class={toClass({
			background: 'var(--zc-f5f5f5, #f5f5f5)',
			borderBottom: '1px solid #ddd',
			padding: '8px 16px',
			display: 'flex',
			alignItems: 'center',
			justifyContent: 'space-between',
			height: HEADER,
			position: 'relative',
			zIndex: 1302,
			userSelect: 'none',
			boxSizing: 'border-box'
		})}
		style="cursor:{max ? 'default' : 'move'}"
		onpointerdown={move}
		ondblclick={() => (view = max ? 'normal' : 'maximized')}
	>
		<div style="display:flex;align-items:center;gap:8px"><h6 class={cx('MuiTypography-root MuiTypography-h6', toClass({ fontSize: '1.1rem', fontWeight: 500 }))}>{windowTitle}</h6></div>
		<div style="display:flex;gap:4px" role="presentation" onpointerdown={(e) => e.stopPropagation()}>
			{@render actions?.()}
			<button type="button" class={btn} title="최소화" onclick={() => (view = view == 'minimized' ? 'normal' : 'minimized')}><Icon name="remove" sx={{ fontSize: 20 }} /></button>
			<button type="button" class={btn} title={max ? '이전 크기로 복원' : '최대화'} onclick={() => (view = max ? 'normal' : 'maximized')}
				><Icon name={max ? 'filter_none' : 'crop_square'} sx={{ fontSize: 20 }} /></button
			>
			{#if onclose}<button type="button" class={btn} title="닫기" onclick={onclose}><Icon name="close" sx={{ fontSize: 20 }} /></button>{/if}
		</div>
	</div>
	<div class={toClass({ position: 'relative' })} style="display:{view == 'minimized' ? 'none' : 'block'};width:{max ? innerWidth : size.width}px;height:{max ? innerHeight - HEADER : size.height}px">
		<div style="height:100%;overflow:auto">{@render children?.()}</div>
		{#if !max}{#each HANDLES as [h, s] (h)}<span role="presentation" class={toClass({ position: 'absolute', zIndex: 100, touchAction: 'none', ...s })} onpointerdown={(e) => resize(e, h)}
				></span>{/each}{/if}
	</div>
</div>
