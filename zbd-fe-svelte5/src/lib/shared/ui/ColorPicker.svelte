<script lang="ts">
	/**
	 * 색 고르기 팝오버 — React mui-color-input 의 팝오버(채도·명도 판 · 색상 막대 · 투명도 막대 · 스포이드 · R/G/B).
	 * 브라우저 기본 고르기 창(input type=color)은 투명도가 없어서 만들었다(2026-10-03 대표).
	 *
	 *   anchor   붙일 엘리먼트(색 견본) — 그 아래 왼쪽에 뜬다(화면 밖이면 위로)
	 *   value    #rgb · #rrggbb · #rrggbbaa (그 밖이면 검정에서 시작)
	 *   format   'hex8' = 늘 #rrggbbaa · 'auto' = 불투명이면 #rrggbb, 투명하면 #rrggbbaa
	 *   alpha    투명도 막대를 보일지(기본 true)
	 * 바깥을 누르거나 Esc 면 onclose.
	 */
	import { portal } from './portal.ts';

	let {
		anchor,
		value = '',
		format = 'auto',
		alpha: useAlpha = true,
		onchange,
		onclose
	}: { anchor: HTMLElement | null | undefined; value?: string; format?: 'hex8' | 'auto'; alpha?: boolean; onchange: (v: string) => void; onclose: () => void } = $props();

	type HSV = { h: number; s: number; v: number; a: number };

	function parse(v: string): HSV {
		let h = (v || '').trim().replace('#', '');
		if (h.length == 3 || h.length == 4) h = h.replace(/(.)/g, '$1$1');
		if (!/^[0-9a-f]{6}([0-9a-f]{2})?$/i.test(h)) h = '000000';
		const r = parseInt(h.slice(0, 2), 16) / 255;
		const g = parseInt(h.slice(2, 4), 16) / 255;
		const b = parseInt(h.slice(4, 6), 16) / 255;
		const a = h.length == 8 ? parseInt(h.slice(6), 16) / 255 : 1;
		const max = Math.max(r, g, b);
		const d = max - Math.min(r, g, b);
		let hue = 0;
		if (d) hue = max == r ? ((g - b) / d) % 6 : max == g ? (b - r) / d + 2 : (r - g) / d + 4;
		return { h: (hue * 60 + 360) % 360, s: max ? d / max : 0, v: max, a };
	}
	function rgb({ h, s, v }: HSV) {
		const f = (n: number) => {
			const k = (n + h / 60) % 6;
			return Math.round((v - v * s * Math.max(0, Math.min(k, 4 - k, 1))) * 255);
		};
		return [f(5), f(3), f(1)];
	}
	const hx = (n: number) => n.toString(16).padStart(2, '0');
	function toHex(c: HSV) {
		const [r, g, b] = rgb(c);
		const a = Math.round(c.a * 255);
		return '#' + hx(r) + hx(g) + hx(b) + (format == 'hex8' || a < 255 ? hx(a) : '');
	}

	// 판을 움직이는 동안 값이 바뀌어도 색상(h)이 튀지 않게 — 내가 낸 값이면 다시 읽지 않는다
	let c = $state<HSV>({ h: 0, s: 0, v: 0, a: 1 });
	let last = '';
	$effect(() => {
		if (value !== last) c = parse(value);
	});
	function emit(next: HSV) {
		c = next;
		last = toHex(next);
		onchange(last);
	}
	const [cr, cg, cb] = $derived(rgb(c));
	const pure = $derived(`hsl(${c.h}, 100%, 50%)`);
	const solid = $derived(`rgb(${cr}, ${cg}, ${cb})`);

	// 자리 — 견본 아래(모자라면 위)
	let panel = $state<HTMLDivElement>();
	let pos = $state({ x: -9999, y: -9999 });
	$effect(() => {
		if (!anchor || !panel) return;
		const a = anchor.getBoundingClientRect();
		const w = panel.offsetWidth;
		const h = panel.offsetHeight;
		const x = Math.max(8, Math.min(a.left, window.innerWidth - w - 8));
		const below = a.bottom + 4;
		pos = { x, y: below + h > window.innerHeight - 8 && a.top - h - 4 > 8 ? a.top - h - 4 : below };
	});
	$effect(() => {
		const down = (e: PointerEvent) => {
			const t = e.target as Node;
			if (panel?.contains(t) || anchor?.contains(t)) return;
			onclose();
		};
		const key = (e: KeyboardEvent) => e.key == 'Escape' && onclose();
		window.addEventListener('pointerdown', down, true);
		window.addEventListener('keydown', key);
		return () => {
			window.removeEventListener('pointerdown', down, true);
			window.removeEventListener('keydown', key);
		};
	});

	/** 판·막대 끌기 — 누른 곳부터 바로 반영, 손을 떼면 끝 */
	function drag(e: PointerEvent, el: HTMLElement, at: (x: number, y: number) => void) {
		e.preventDefault();
		const go = (ev: PointerEvent) => {
			const r = el.getBoundingClientRect();
			at(Math.max(0, Math.min(1, (ev.clientX - r.left) / r.width)), Math.max(0, Math.min(1, (ev.clientY - r.top) / r.height)));
		};
		go(e);
		const up = () => {
			window.removeEventListener('pointermove', go);
			window.removeEventListener('pointerup', up);
		};
		window.addEventListener('pointermove', go);
		window.addEventListener('pointerup', up);
	}
	function keyStep(e: KeyboardEvent, axis: 'x' | 'y' | 'xy', get: () => [number, number], set: (x: number, y: number) => void) {
		const d = e.shiftKey ? 0.1 : 0.01;
		let [x, y] = get();
		if (e.key == 'ArrowLeft' && axis != 'y') x -= d;
		else if (e.key == 'ArrowRight' && axis != 'y') x += d;
		else if (e.key == 'ArrowUp' && axis != 'x') y -= d;
		else if (e.key == 'ArrowDown' && axis != 'x') y += d;
		else return;
		e.preventDefault();
		set(Math.max(0, Math.min(1, x)), Math.max(0, Math.min(1, y)));
	}

	let svEl = $state<HTMLDivElement>();
	let hueEl = $state<HTMLDivElement>();
	let alphaEl = $state<HTMLDivElement>();
	const setSV = (x: number, y: number) => emit({ ...c, s: x, v: 1 - y });
	const setHue = (x: number) => emit({ ...c, h: Math.min(359.9, x * 360) });
	const setAlpha = (x: number) => emit({ ...c, a: Math.round(x * 100) / 100 });

	function setRGB(i: number, raw: string) {
		const n = Math.max(0, Math.min(255, parseInt(raw, 10) || 0));
		const arr = [cr, cg, cb];
		arr[i] = n;
		emit({ ...parse('#' + arr.map(hx).join('')), a: c.a });
	}
	const canEye = typeof window != 'undefined' && 'EyeDropper' in window;
	async function eyedrop() {
		try {
			const r = await new (window as any).EyeDropper().open();
			emit({ ...parse(r.sRGBHex), a: c.a });
		} catch {}
	}
</script>

<div bind:this={panel} use:portal class="zcp" role="dialog" aria-label="색 고르기" style="left:{pos.x}px;top:{pos.y}px">
	<div
		bind:this={svEl}
		class="sv"
		style="background-color:{pure}"
		role="slider"
		tabindex="0"
		aria-label="채도·명도"
		aria-valuenow={Math.round(c.s * 100)}
		onpointerdown={(e) => drag(e, svEl!, setSV)}
		onkeydown={(e) => keyStep(e, 'xy', () => [c.s, 1 - c.v], setSV)}
	>
		<div class="sv-w"></div>
		<div class="sv-b"></div>
		<span class="dot" style="left:{c.s * 100}%;top:{(1 - c.v) * 100}%;background:{solid}"></span>
	</div>
	<div class="row">
		{#if canEye}<button type="button" class="eye" aria-label="스포이드" onclick={eyedrop}
				><svg viewBox="0 0 24 24" width="20" height="20"
					><path
						fill="currentColor"
						d="m20.71 5.63-2.34-2.34a.996.996 0 0 0-1.41 0l-3.12 3.12-1.93-1.91-1.41 1.41 1.42 1.42L3 16.25V21h4.75l8.92-8.92 1.42 1.42 1.41-1.41-1.92-1.92 3.12-3.12c.4-.4.4-1.03.01-1.42M6.92 19 5 17.08l8.06-8.06 1.92 1.92z"
					/></svg
				></button
			>{/if}
		<span class="preview"><span style="background:{solid};opacity:{c.a}"></span></span>
		<div class="bars">
			<div
				bind:this={hueEl}
				class="bar hue"
				role="slider"
				tabindex="0"
				aria-label="색상"
				aria-valuenow={Math.round(c.h)}
				onpointerdown={(e) => drag(e, hueEl!, (x) => setHue(x))}
				onkeydown={(e) =>
					keyStep(
						e,
						'x',
						() => [c.h / 360, 0],
						(x) => setHue(x)
					)}
			>
				<span class="knob" style="left:{(c.h / 360) * 100}%;background:{pure}"></span>
			</div>
			{#if useAlpha}
				<div
					bind:this={alphaEl}
					class="bar alpha"
					role="slider"
					tabindex="0"
					aria-label="투명도"
					aria-valuenow={Math.round(c.a * 100)}
					onpointerdown={(e) => drag(e, alphaEl!, (x) => setAlpha(x))}
					onkeydown={(e) =>
						keyStep(
							e,
							'x',
							() => [c.a, 0],
							(x) => setAlpha(x)
						)}
				>
					<span class="alpha-fill" style="background:linear-gradient(to right, rgba({cr},{cg},{cb},0), {solid})"></span>
					<span class="knob" style="left:{c.a * 100}%;background:{solid}"></span>
				</div>
			{/if}
		</div>
	</div>
	<div class="nums">
		{#each ['R', 'G', 'B'] as n, i (n)}
			<label><input type="number" min="0" max="255" value={[cr, cg, cb][i]} onchange={(e) => setRGB(i, (e.currentTarget as HTMLInputElement).value)} /><span>{n}</span></label>
		{/each}
		{#if useAlpha}
			<label
				><input type="number" min="0" max="100" value={Math.round(c.a * 100)} onchange={(e) => setAlpha(Math.max(0, Math.min(100, +(e.currentTarget as HTMLInputElement).value || 0)) / 100)} /><span
					>A %</span
				></label
			>
		{/if}
	</div>
</div>

<style>
	.zcp {
		position: fixed;
		z-index: 1500;
		width: 236px;
		padding: 0 0 12px;
		background: #fff;
		color: rgba(0, 0, 0, 0.87);
		border-radius: 4px;
		box-shadow:
			0 5px 5px -3px rgba(0, 0, 0, 0.2),
			0 8px 10px 1px rgba(0, 0, 0, 0.14),
			0 3px 14px 2px rgba(0, 0, 0, 0.12);
		font:
			14px/1.4 -apple-system,
			BlinkMacSystemFont,
			'Segoe UI',
			Roboto,
			sans-serif;
		user-select: none;
	}
	.sv {
		position: relative;
		height: 140px;
		border-radius: 4px 4px 0 0;
		cursor: crosshair;
		touch-action: none;
		outline: 0;
	}
	.sv-w,
	.sv-b {
		position: absolute;
		inset: 0;
		border-radius: inherit;
	}
	.sv-w {
		background: linear-gradient(to right, #fff, rgba(255, 255, 255, 0));
	}
	.sv-b {
		background: linear-gradient(to top, #000, rgba(0, 0, 0, 0));
	}
	.dot {
		position: absolute;
		width: 12px;
		height: 12px;
		margin: -8px 0 0 -8px;
		border: 2px solid #fff;
		border-radius: 50%;
		box-shadow: 0 0 0 1.5px rgba(0, 0, 0, 0.6);
		pointer-events: none;
	}
	.sv:focus-visible .dot {
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.6);
	}
	.row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 12px 0;
	}
	.eye {
		display: inline-flex;
		padding: 4px;
		border: 0;
		background: none;
		color: rgba(0, 0, 0, 0.7);
		border-radius: 50%;
		cursor: pointer;
	}
	.eye:hover {
		background: rgba(0, 0, 0, 0.06);
	}
	.preview,
	.alpha {
		background-image:
			linear-gradient(45deg, #ccc 25%, transparent 25%), linear-gradient(135deg, #ccc 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #ccc 75%),
			linear-gradient(135deg, transparent 75%, #ccc 75%);
		background-size: 8px 8px;
		background-position:
			0 0,
			4px 0,
			4px -4px,
			0 4px;
	}
	.preview {
		flex: none;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		overflow: hidden;
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.1);
	}
	.preview span {
		display: block;
		width: 100%;
		height: 100%;
	}
	.bars {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.bar {
		position: relative;
		height: 10px;
		border-radius: 5px;
		cursor: pointer;
		touch-action: none;
		outline: 0;
	}
	.hue {
		background: linear-gradient(to right, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%);
	}
	.alpha-fill {
		position: absolute;
		inset: 0;
		border-radius: inherit;
	}
	.knob {
		position: absolute;
		top: 50%;
		width: 14px;
		height: 14px;
		margin: -9px 0 0 -9px;
		border: 2px solid #fff;
		border-radius: 50%;
		box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.35);
		pointer-events: none;
	}
	.bar:focus-visible .knob {
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.5);
	}
	.nums {
		display: flex;
		gap: 6px;
		padding: 12px 12px 0;
	}
	.nums label {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		color: rgba(0, 0, 0, 0.6);
	}
	.nums input {
		width: 100%;
		box-sizing: border-box;
		height: 30px;
		padding: 0 4px;
		border: 1px solid rgba(0, 0, 0, 0.23);
		border-radius: 4px;
		font: inherit;
		font-size: 13px;
		color: rgba(0, 0, 0, 0.87);
		text-align: center;
		appearance: textfield;
		-moz-appearance: textfield;
	}
	.nums input::-webkit-inner-spin-button,
	.nums input::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}
	.nums input:focus {
		outline: 0;
		border-color: var(--theme-color-main, #1976d2);
	}
</style>
