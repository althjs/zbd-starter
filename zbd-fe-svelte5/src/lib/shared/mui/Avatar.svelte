<script lang="ts">
	/** MUI Avatar 호환 — 사진이 없거나 못 읽으면 글자 */
	import { cx, sx as toClass, type Sx } from './sx';
	let { src = '', text = '', size = 40, class: cls = '', sx = null }: { src?: string; text?: string; size?: number; class?: string; sx?: Sx | null } = $props();
	let failed = $state(false);
	$effect(() => {
		void src;
		failed = false;
	});
</script>

<div class={cx('MuiAvatar-root MuiAvatar-circular', (!src || failed) && 'MuiAvatar-colorDefault', cls, toClass({ width: size, height: size, ...(sx || {}) }))}>
	{#if src && !failed}
		<img {src} alt="" class="MuiAvatar-img" onerror={() => (failed = true)} />
	{:else}
		{text}
	{/if}
</div>
