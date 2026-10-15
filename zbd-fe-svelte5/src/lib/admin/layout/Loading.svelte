<script lang="ts">
	/** 가운데 원형 로딩(React components/Loading) — 화면 높이만큼 */
	let { message = '', full = true }: { message?: string; full?: boolean } = $props();
</script>

<div class="MuiBox-root adm-loading" class:full>
	<span class="MuiCircularProgress-root MuiCircularProgress-indeterminate MuiCircularProgress-colorPrimary" role="progressbar">
		<svg class="MuiCircularProgress-svg" viewBox="22 22 44 44"
			><circle class="MuiCircularProgress-circle MuiCircularProgress-circleIndeterminate" cx="44" cy="44" r="20.2" fill="none" stroke-width="3.6"></circle></svg
		>
	</span>
	{#if message}<p class="MuiTypography-root MuiTypography-body1">{message}</p>{/if}
</div>

<style>
	.adm-loading {
		display: flex;
		justify-content: center;
		align-items: center;
		flex-direction: column;
		padding: 24px 0;
	}
	.adm-loading.full {
		min-height: 100vh;
		padding: 0;
	}
</style>
