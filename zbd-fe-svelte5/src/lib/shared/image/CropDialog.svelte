<script lang="ts">
	/**
	 * 사진 자르기 — 정해진 크기(width×height) 로. 끌어서 옮기기 · 확대 · 90° 돌리기. 예전 ImageCropDialog.tsx(Jcrop) 자리.
	 * autoCrop 이면 묻지 않고 가운데를 잘라 바로 돌려준다.
	 */
	import Dialog from '../mui/Dialog.svelte';
	import MuiButton from '../mui/MuiButton.svelte';
	import { blobToDataUrl, canvasBlob, loadImage } from './files';

	let {
		file,
		width,
		height,
		mime = 'image/webp',
		autoCrop = false,
		ondone
	}: { file: Blob | null; width: number; height: number; mime?: string; autoCrop?: boolean; ondone: (b: Blob | null) => void } = $props();

	let img = $state<HTMLImageElement | null>(null);
	let zoom = $state(1);
	let deg = $state(0);
	let off = $state({ x: 0, y: 0 });
	let box = $state<HTMLElement>();
	const VIEW = 320; // 미리보기 틀의 긴 변
	const vw = $derived(width >= height ? VIEW : Math.round((VIEW * width) / height));
	const vh = $derived(width >= height ? Math.round((VIEW * height) / width) : VIEW);

	$effect(() => {
		if (!file) return;
		blobToDataUrl(file)
			.then(loadImage)
			.then((i) => {
				img = i;
				zoom = 1;
				deg = 0;
				off = { x: 0, y: 0 };
				if (autoCrop) done();
			});
	});

	/** 회전 반영한 원본 크기 · 틀을 꽉 채우는 기본 배율 */
	const rot = $derived(deg % 180 != 0);
	const iw = $derived(img ? (rot ? img.height : img.width) : 1);
	const ih = $derived(img ? (rot ? img.width : img.height) : 1);
	const base = $derived(Math.max(vw / iw, vh / ih));

	function clamp(o: { x: number; y: number }) {
		const s = base * zoom;
		const mx = Math.max(0, (iw * s - vw) / 2);
		const my = Math.max(0, (ih * s - vh) / 2);
		return { x: Math.max(-mx, Math.min(mx, o.x)), y: Math.max(-my, Math.min(my, o.y)) };
	}

	let drag: { x: number; y: number; ox: number; oy: number } | null = null;
	function down(e: PointerEvent) {
		drag = { x: e.clientX, y: e.clientY, ox: off.x, oy: off.y };
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
	}
	function move(e: PointerEvent) {
		if (!drag) return;
		off = clamp({ x: drag.ox + e.clientX - drag.x, y: drag.oy + e.clientY - drag.y });
	}

	async function done() {
		if (!img) return;
		const c = document.createElement('canvas');
		c.width = width;
		c.height = height;
		const ctx = c.getContext('2d')!;
		const k = width / vw; // 미리보기 → 결과 배율
		const s = base * zoom * k;
		ctx.translate(width / 2 + off.x * k, height / 2 + off.y * k);
		ctx.rotate((deg * Math.PI) / 180);
		ctx.drawImage(img, (-img.width * s) / 2, (-img.height * s) / 2, img.width * s, img.height * s);
		const b = await canvasBlob(c, mime);
		img = null;
		ondone(b);
	}
</script>

<Dialog open={!!img && !autoCrop} onclose={() => ((img = null), ondone(null))} maxWidth={420} paperSx={{ width: vw + 48, maxWidth: 'calc(100% - 32px)' }} fitWidth={false}>
	{#snippet title()}{width}x{height} 로 자르기{/snippet}
	<div class="frame" bind:this={box} style="width:{vw}px;height:{vh}px" role="presentation" onpointerdown={down} onpointermove={move} onpointerup={() => (drag = null)}>
		{#if img}
			<img
				src={img.src}
				alt=""
				draggable="false"
				style="width:{img.width * base * zoom}px;height:{img.height * base * zoom}px;transform:translate(-50%,-50%) translate({off.x}px,{off.y}px) rotate({deg}deg)"
			/>
		{/if}
	</div>
	<div class="tools">
		<input type="range" min="1" max="4" step="0.01" bind:value={zoom} oninput={() => (off = clamp(off))} aria-label="확대" />
		<MuiButton size="small" onclick={() => ((deg = (deg + 90) % 360), (off = { x: 0, y: 0 }))}>돌리기</MuiButton>
	</div>
	{#snippet actions()}
		<MuiButton onclick={() => ((img = null), ondone(null))}>취소</MuiButton>
		<MuiButton variant="contained" onclick={done}>자르기</MuiButton>
	{/snippet}
</Dialog>

<style>
	/* 창 폭은 틀 폭 + 좌우 여백 24 로 정한다(fitWidth 끔) — 내용에 맞추는 창이 틀보다 좁게 재어 오른쪽이 잘렸다 */
	.frame {
		position: relative;
		flex-shrink: 0;
		overflow: hidden;
		margin: 0 auto;
		background: #222;
		touch-action: none;
		cursor: grab;
	}
	.frame img {
		position: absolute;
		left: 50%;
		top: 50%;
		max-width: none;
		user-select: none;
		pointer-events: none;
	}
	.tools {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-top: 12px;
	}
	.tools input {
		flex: 1;
	}
</style>
