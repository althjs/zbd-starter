<script lang="ts">
	/**
	 * 사이트 머리(고정) — 로고 · 메뉴 · 로그인/내 정보. 스타터에는 페이지 디자인(GNB) 이 없어 코드로 그린다 — 메뉴는 아래 MENU 를 고친다.
	 */
	import Logo from '#lib/admin/layout/Logo.svelte';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import { siteMetaOf } from '#lib/user/site.svelte.ts';
	import { doLogin, session } from '#lib/user/session.svelte.ts';
	import { doLogout } from '#lib/user/zbdSite.ts';

	const MENU = [
		{ name: '공지사항', href: '/board/notice' },
		{ name: '자유게시판', href: '/board/free' },
		{ name: '고객센터', href: '/cs' }
	];
	const siteName = $derived(siteMetaOf()?.title || '');
	// 환경 설정 › 국내 홈페이지 의 로고(없으면 기본 스타터 로고) · 로고 링크주소
	const logo = $derived(siteMetaOf()?.design?.logo);
	const logoHref = $derived(siteMetaOf()?.design?.logoHref || '/');
	let open = $state(false);
</script>

<header class="zs-head">
	<div class="zs-wrap">
		<a class="zs-brand" href={logoHref} aria-label="홈">
			{#if logo?.name}
				<img class="zs-logo" src={`/upload/${logo.hostingId || HOSTING_ID}/site/${logo.name}`} alt={siteName || '홈'} />
			{:else}
				<Logo height={24} color="#1f2937" />
				{#if siteName}<span class="zs-site">{siteName}</span>{/if}
			{/if}
		</a>
		<button class="zs-burger" aria-label="메뉴" aria-expanded={open} onclick={() => (open = !open)}>☰</button>
		<nav class="zs-nav" class:open>
			{#each MENU as m (m.href)}<a href={m.href} onclick={() => (open = false)}>{m.name}</a>{/each}
			<span class="zs-gap"></span>
			{#if session.profile}
				<a href="/user/my" onclick={() => (open = false)}>내 정보</a>
				<button class="zs-link" onclick={() => doLogout()}>로그아웃</button>
			{:else if session.checked}
				<button class="zs-link" onclick={() => doLogin()}>로그인</button>
				<a class="zs-join" href="/user/join">회원가입</a>
			{/if}
		</nav>
	</div>
</header>

<style>
	.zs-head {
		position: sticky;
		top: 0;
		z-index: 100;
		background: #fff;
		border-bottom: 1px solid #e5e7eb;
	}
	.zs-wrap {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 20px;
		height: 60px;
		display: flex;
		align-items: center;
		gap: 24px;
	}
	.zs-brand {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
		color: #1f2937;
	}
	.zs-logo {
		height: 26px;
		width: auto;
		display: block;
	}
	.zs-site {
		font-weight: 700;
		font-size: 15px;
		padding-left: 10px;
		border-left: 1px solid #e5e7eb;
	}
	.zs-nav {
		display: flex;
		align-items: center;
		gap: 20px;
		flex: 1;
	}
	.zs-nav a,
	.zs-link {
		color: #374151;
		text-decoration: none;
		font-size: 15px;
		background: none;
		border: 0;
		padding: 0;
		cursor: pointer;
		font: inherit;
	}
	.zs-nav a:hover,
	.zs-link:hover {
		color: var(--theme-color-main);
	}
	.zs-gap {
		flex: 1;
	}
	.zs-nav .zs-join {
		padding: 6px 14px;
		border-radius: 8px;
		background: var(--theme-color-main);
		color: var(--theme-color-contrastText);
	}
	.zs-burger {
		display: none;
		margin-left: auto;
		background: none;
		border: 0;
		font-size: 22px;
		cursor: pointer;
	}
	@media (max-width: 899px) {
		.zs-burger {
			display: block;
		}
		.zs-nav {
			display: none;
			position: absolute;
			top: 60px;
			left: 0;
			right: 0;
			flex-direction: column;
			align-items: flex-start;
			gap: 14px;
			padding: 16px 20px 20px;
			background: #fff;
			border-bottom: 1px solid #e5e7eb;
		}
		.zs-nav.open {
			display: flex;
		}
		.zs-gap {
			display: none;
		}
	}
</style>
