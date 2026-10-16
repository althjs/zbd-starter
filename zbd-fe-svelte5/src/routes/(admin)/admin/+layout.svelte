<script lang="ts">
	/**
	 * 관리자 화면(/admin/…) — 로그인 확인·권한 메뉴(loadAdmin) 뒤 틀을 그린다.
	 * 단독 앱(chat/mail.zbd.kr)·새 창(?popup=1) 은 메뉴 없는 틀, 단독 앱은 그 앱 화면 밖이면 앱 첫 화면으로.
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { SINGLE_APP } from '#lib/admin/config.ts';
	import { adminImagePrefix, getMeta } from '#lib/admin/meta.svelte.ts';
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { admin, authCheckAndRedirect, loadAdmin } from '#lib/admin/profile.svelte.ts';
	import { startAdminSocket } from '#lib/admin/socket.svelte.ts';
	import AdminShell from '#lib/admin/layout/AdminShell.svelte';
	import PopupLayout from '#lib/admin/layout/PopupLayout.svelte';
	import { IN_WIN_FRAME } from '#lib/admin/window/frame.ts';
	import { isStandalone, markInstalled } from '#lib/admin/app-installed.ts';

	let { children } = $props();

	loadAdmin(page.url.pathname).then(() => admin.profile && startAdminSocket(!!admin.profile.isDeveloper));

	// 메뉴를 받은 뒤 주소가 바뀔 때마다 권한 확인(React 는 첫 진입·/admin 에서)
	$effect(() => {
		const p = page.url.pathname;
		if (admin.loaded && (p == '/admin' || p == '/admin/')) authCheckAndRedirect(admin.menus, p);
	});
	$effect(() => {
		if (SINGLE_APP && !SINGLE_APP.paths.some((p) => page.url.pathname.indexOf(p) == 0)) goto(SINGLE_APP.home, { replaceState: true });
	});
	// 파비콘 = 사이트 설정(SITE 메타 favicon)이 있으면 그것, 없으면 app.html 의 기본 즈비드 아이콘(2026-10-03 — 설정값이 어디에도 쓰이지 않았다)
	$effect(() => {
		void admin.metaVersion;
		const name = admin.loaded ? getMeta('SITE', 'favicon')?.name : '';
		const el = document.getElementById('zbd-favicon') as HTMLLinkElement | null;
		if (el && name) el.href = `${adminImagePrefix(hostingIdOf(getMeta('SITE', 'favicon')) || '')}/site/${name}`;
	});
	// 설치 앱 표시 — 설치 창으로 열렸거나 방금 설치되면 이 앱(단독 앱 키, 아니면 관리자 앱)을 「설치됨」 으로(대시보드 「웹앱 설치」 가 숨긴다)
	$effect(() => {
		const key = SINGLE_APP?.key || 'admin';
		if (isStandalone()) markInstalled(key);
		const on = () => markInstalled(key);
		addEventListener('appinstalled', on);
		return () => removeEventListener('appinstalled', on);
	});
	const popup = $derived(!!SINGLE_APP || page.url.searchParams.get('popup') === '1' || IN_WIN_FRAME); // 화면 위 창(iframe) 안은 옮겨 다녀도 메뉴 없는 틀
</script>

{#if admin.loaded}
	{#if popup}
		<PopupLayout>{@render children()}</PopupLayout>
	{:else}
		<AdminShell>{@render children()}</AdminShell>
	{/if}
{/if}
