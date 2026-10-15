<script lang="ts">
	/** 없는 주소·오류 화면 — 예전 fe-service pages/NotFound.tsx(영문 기본 문구였던 것을 우리 말로) */
	import { page } from '$app/state';
	const notFound = $derived(page.status == 404);
	const home = $derived(page.url.pathname.startsWith('/admin') ? '/admin' : '/');
</script>

<svelte:head><title>{notFound ? '페이지를 찾을 수 없어요' : '오류'} - 즈비드</title></svelte:head>

<main class="zbd-error">
	<span class="material-icons" aria-hidden="true">{notFound ? 'construction' : 'error_outline'}</span>
	<h3>{notFound ? '찾는 페이지가 없어요' : '화면을 그리다 문제가 생겼어요'}</h3>
	<p>{notFound ? '주소가 바뀌었거나 지워진 페이지예요. 메뉴에서 다시 찾아 주세요.' : page.error?.message || '잠시 뒤 다시 시도해 주세요.'}</p>
	<a href={home}><span class="material-icons" aria-hidden="true">home</span>처음으로</a>
</main>

<style>
	.zbd-error {
		min-height: 70vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 16px;
		padding: 24px 16px;
		text-align: center;
	}
	.zbd-error > .material-icons {
		font-size: 70px;
		color: var(--theme-color-main, #31726c);
	}
	h3 {
		font-size: 24px;
		font-weight: 700;
	}
	p {
		color: #666;
	}
	a {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		margin-top: 8px;
		padding: 8px 20px;
		border-radius: 8px;
		background: var(--theme-color-main, #31726c);
		color: #fff;
		text-decoration: none;
		font-weight: 600;
	}
	a .material-icons {
		font-size: 18px;
	}
</style>
