<script lang="ts">
	/** 관리자 로그인 — React 경로 그대로(/login/admin-login). 짙은 청록 바탕(빛 번짐 두 개) + 가운데 흰 카드(2026-10-07 개편) */
	import AdminLogin from '#lib/admin/login/AdminLogin.svelte';
</script>

<svelte:head><title>관리자 로그인</title></svelte:head>

<div class="MuiStack-root" id="main_area" style="height:100%">
	<div class="adm-login-root">
		<main class="MuiBox-root adm-login-main"><AdminLogin /></main>
	</div>
</div>

<style>
	.adm-login-root {
		position: relative;
		overflow: hidden;
		display: flex;
		flex: 1;
		align-items: center;
		justify-content: center;
		min-height: 100dvh;
		padding: 40px 16px;
		box-sizing: border-box;
		background:
			radial-gradient(520px 420px at 85% 8%, rgba(255, 143, 135, 0.28), transparent 70%), radial-gradient(620px 520px at 8% 100%, rgba(111, 211, 184, 0.22), transparent 70%),
			linear-gradient(160deg, #1d433d 0%, #143430 55%, #0e2724 100%);
	}
	.adm-login-root::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image: radial-gradient(rgba(255, 255, 255, 0.07) 1px, transparent 1px);
		background-size: 22px 22px;
		mask-image: linear-gradient(180deg, #000 0%, transparent 70%);
		pointer-events: none;
	}
	.adm-login-main {
		position: relative;
		width: 100%;
		justify-content: center;
		display: flex;
	}
</style>
