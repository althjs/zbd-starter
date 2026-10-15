<script lang="ts">
	/**
	 * 라이프 라벨 한 개 — 라이프 메모·미니홈피 메모·관리자 라이프로그가 같이 쓴다.
	 * 라벨 이미지가 있으면 그림, 없으면 글자 칩(바탕 80% 불투명).
	 */
	import { imageUrl } from '#lib/shared/config.ts';

	type Label = { name: string; color?: string; bgColor?: string; image?: { name: string } | null };
	let { label, imageId = '', onclick, style = '' }: { label: Label | null; imageId?: string; onclick?: (l: Label) => void; style?: string } = $props();

	function alpha80(hex: string) {
		const h = (hex || '#ffffff').replace('#', '');
		const f = h.length == 3 ? h.replace(/(.)/g, '$1$1') : h.slice(0, 6);
		const n = parseInt(f, 16);
		if (Number.isNaN(n)) return hex;
		return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, 0.8)`;
	}
</script>

{#if label}
	<div class="label-item" class:clickable={!!onclick} role={onclick ? 'button' : undefined} onclick={() => onclick?.(label)}>
		{#if label.image?.name}
			<img class="img" src={imageUrl(imageId, label.image.name)} title={label.name} alt={label.name} />
		{:else}
			<div class="chip" style="border:1px solid {label.bgColor || '#ddd'};background:{alpha80(label.bgColor || '#ffffff')};color:{label.color || '#000'};{style}">
				{label.name}
			</div>
		{/if}
	</div>
{/if}

<style>
	.label-item {
		display: flex;
		flex-direction: column;
	}
	.clickable {
		cursor: pointer;
	}
	.img {
		width: 50px;
		height: 25px;
		object-fit: contain;
	}
	.chip {
		font-size: 13px;
		text-align: center;
		width: max-content;
		padding: 0 8px;
	}
</style>
