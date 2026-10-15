<script lang="ts">
	/** 원형 로딩(MUI CircularProgress) — 크기만 바꿔 쓴다 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { size = 40, sx = null }: { size?: number; sx?: Sx | null } = $props();
</script>

<span class={cx('MuiCircularProgress-root MuiCircularProgress-indeterminate MuiCircularProgress-colorPrimary', toClass(sx))} role="progressbar" style="width:{size}px;height:{size}px">
	<svg class="MuiCircularProgress-svg" viewBox="22 22 44 44"
		><circle class="MuiCircularProgress-circle MuiCircularProgress-circleIndeterminate" cx="44" cy="44" r="20.2" fill="none" stroke-width="3.6"></circle></svg
	>
</span>
