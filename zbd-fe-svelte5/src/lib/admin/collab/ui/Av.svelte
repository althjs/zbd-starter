<script lang="ts">
	/** MUI Avatar — 사진이 없으면 MUI 처럼 사람 모양(글자가 있으면 글자) */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { src = '', text = '', size = 40, class: cls = '', sx = null }: { src?: string; text?: string; size?: number; class?: string; sx?: Sx | null } = $props();
	let failed = $state(false);
	$effect(() => {
		void src;
		failed = false;
	});
	const img = $derived(!!src && !failed);
</script>

<div class={cx('MuiAvatar-root MuiAvatar-circular', !img && 'MuiAvatar-colorDefault', cls, toClass({ width: size, height: size, ...(sx || {}) }))}>
	{#if img}
		<img {src} alt="" class="MuiAvatar-img" onerror={() => (failed = true)} />
	{:else if text}
		{text}
	{:else}
		<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiAvatar-fallback" focusable="false" aria-hidden="true" viewBox="0 0 24 24" style="width:75%;height:75%;fill:currentColor"
			><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"></path></svg
		>
	{/if}
</div>
