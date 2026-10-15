<script lang="ts">
	/** 별점 보기(MUI Rating small · disabled · 0.5 단위) */
	let { value = 0 }: { value?: number } = $props();
	const stars = [1, 2, 3, 4, 5];
	const STAR = 'M12 17.27 18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z';
	const BORDER =
		'M22 9.24l-7.19-.62L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21 12 17.27 18.18 21l-1.63-7.03L22 9.24zM12 15.4l-3.76 2.27 1-4.28-3.32-2.88 4.38-.38L12 6.1l1.71 4.04 4.38.38-3.32 2.88 1 4.28L12 15.4z';
</script>

<span class="MuiRating-root MuiRating-sizeSmall Mui-disabled Mui-readOnly rating" role="img" aria-label="{value} Stars">
	{#each stars as s (s)}
		{@const fill = Math.max(0, Math.min(1, value - s + 1))}
		<span class="MuiRating-icon {fill >= 1 ? 'MuiRating-iconFilled' : 'MuiRating-iconEmpty'}" style="position:relative">
			{#if fill > 0 && fill < 1}
				<span style="position:absolute;inset:0;width:{fill * 100}%;overflow:hidden;color:#faaf00"
					><svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeInherit" viewBox="0 0 24 24"><path d={STAR}></path></svg></span
				>
			{/if}
			<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeInherit" viewBox="0 0 24 24"><path d={fill >= 1 ? STAR : BORDER}></path></svg>
		</span>
	{/each}
</span>

<style>
	.MuiRating-root {
		display: inline-flex;
		position: relative;
		font-size: 1.125rem;
		color: #faaf00;
		text-align: left;
		opacity: 0.5;
		pointer-events: none;
	}
	.MuiRating-icon {
		display: flex;
	}
	.MuiRating-iconEmpty {
		color: rgba(0, 0, 0, 0.26);
	}
</style>
