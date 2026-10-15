<script lang="ts">
	/** 사람 사진(MUI Avatar) — 사진이 없거나 못 읽으면 MUI 처럼 사람 모양(Person) 아이콘 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { src = '', size = 40, sx = null }: { src?: string; size?: number; sx?: Sx | null } = $props();
	let failed = $state(false);
	$effect(() => {
		void src;
		failed = false;
	});
</script>

<div class={cx('MuiAvatar-root MuiAvatar-circular', (!src || failed) && 'MuiAvatar-colorDefault', toClass({ width: size, height: size, flexShrink: 0, ...(sx || {}) }))}>
	{#if src && !failed}
		<img {src} alt="" class="MuiAvatar-img" onerror={() => (failed = true)} />
	{:else}
		<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiAvatar-fallback" style="width:75%;height:75%;fill:currentColor" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
			><path d="M12 12c2.7 0 4.8-2.1 4.8-4.8S14.7 2.4 12 2.4 7.2 4.5 7.2 7.2 9.3 12 12 12zm0 2.4c-3.2 0-9.6 1.6-9.6 4.8v2.4h19.2v-2.4c0-3.2-6.4-4.8-9.6-4.8z"></path></svg
		>
	{/if}
</div>
