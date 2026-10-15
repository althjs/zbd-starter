<script lang="ts">
	/**
	 * 관리자 묶음(backyard·chat·mail) — 로그인 화면과 /admin 화면 공통.
	 *  - API 는 관리자 토큰(zbd:auth:admin) — 401 이면 로그인 화면으로
	 *  - 관리자 주소에서만 연다(CLAUDE.md 1) — 서버가 알려 준 HOST_REALM(관리자 「도메인」 설정, 개발 한 주소는 끔)
	 *  - <html class="zbd-admin"> + 테마 변수(사이트 대표색·다크모드) — 관리자 CSS 는 이 클래스 아래에만 걸린다
	 *  - 알림·확인 창(AdminDialogs), 휴대폰 뒤로가기 = 덮은 창 닫기
	 */
	import '#lib/admin/styles/admin.css';
	import { page } from '$app/state';
	import { setRealm } from '#lib/shared/api.ts';
	import { ADMIN_LOGIN_PATH, LS } from '#lib/admin/config.ts';
	import { HOST_REALM } from '#lib/shared/config.ts';
	import { installAutoBackToClose } from '#lib/admin/backToClose.svelte.ts';
	import { dark } from '#lib/admin/dark.svelte.ts';
	import { getMeta } from '#lib/admin/meta.svelte.ts';
	import { applyAdminTheme } from '#lib/admin/theme/theme.ts';
	import AdminDialogs from '#lib/admin/ui/AdminDialogs.svelte';

	let { children } = $props();

	// 관리자 주소가 아니면 그리지 않고 관리자 주소로
	const wrongHost = HOST_REALM == 'user';
	if (wrongHost) location.replace('/'); // 사용자 주소에서는 관리자 화면을 드러내지 않는다 — 홈으로(2026-10-03 대표, nginx 도 같게)

	setRealm('admin', () => {
		if (location.pathname.indexOf('/login') == 0) return;
		try {
			localStorage.setItem(LS.loginPath, location.pathname + location.search);
		} catch {}
		location.replace(ADMIN_LOGIN_PATH);
	});

	const isLogin = $derived(page.url.pathname.indexOf('/login') == 0);

	function themeName() {
		let opts: any = null;
		try {
			opts = JSON.parse(localStorage.getItem(LS.themeOpts) || 'null');
		} catch {}
		// 로그인 화면은 저장된 테마 선택만, 대시보드는 사이트 메타가 먼저(React 두 레이아웃과 같게)
		if (isLogin) return opts?.theme || 'blue';
		const name = getMeta('SITE', 'theme') || opts?.theme || 'blue';
		try {
			localStorage.setItem(LS.themeOpts, JSON.stringify({ ...(opts || { hide: true }), theme: name }));
		} catch {}
		return name;
	}

	$effect(() => {
		const html = document.documentElement;
		html.classList.add('zbd-admin');
		return () => {
			html.classList.remove('zbd-admin', 'zbd-admin-bare');
			html.removeAttribute('data-mode');
		};
	});
	$effect(() => {
		document.documentElement.classList.toggle('zbd-admin-bare', isLogin);
	});
	$effect(() => {
		document.documentElement.setAttribute('data-mode', dark.resolved);
		applyAdminTheme(themeName(), isLogin ? 'light' : dark.resolved);
	});
	$effect(() => installAutoBackToClose());
</script>

<svelte:head>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard/dist/web/static/pretendard-dynamic-subset.css" />
</svelte:head>

{#if !wrongHost}
	{@render children()}
	<AdminDialogs />
{/if}
