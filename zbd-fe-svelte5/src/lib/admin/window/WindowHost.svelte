<script lang="ts">
	/**
	 * 화면 위 창 렌더러(React window-control/GlobalWindowRenderer) — 관리자 껍데기(AdminShell)에 한 번. 페이지를 옮겨도 창이 남는다.
	 * + 오른쪽 가운데 창 목록 손잡이(WindowDock) + 아래 오른쪽 최소화 줄(#minimized-dock).
	 * + 최고 관리자는 관리자 링크를 Shift+누르기 = 그 화면을 창으로(React RouterLink).
	 * 메뉴 없는 틀(PopupLayout — ?popup=1 새 창·단독 앱·경로 창의 iframe)에도 둔다: 그 안에서 연 부품 창은 그 문서 안에 뜬다.
	 * 창이 최대화돼 있으면 뒤 페이지 스크롤을 잠근다(React GlobalStyles html/body overflow hidden).
	 */
	import { isSuperAdmin } from '../meta.svelte.ts';
	import { admin } from '../profile.svelte.ts';
	import MinimizedDock from './MinimizedDock.svelte';
	import RouteFrame from './RouteFrame.svelte';
	import Window from './Window.svelte';
	import WindowDock from './WindowDock.svelte';
	import { openRouteWindow, wm } from './windows.svelte.ts';

	const anyMax = $derived(wm.list.some((w) => w.view === 'maximized'));
	$effect(() => {
		document.documentElement.classList.toggle('zbd-win-max', anyMax);
		return () => document.documentElement.classList.remove('zbd-win-max');
	});

	// Shift+누르기(최고 관리자) — /admin 링크를 창으로
	$effect(() => {
		const h = (e: MouseEvent) => {
			if (!e.shiftKey || e.button !== 0 || !isSuperAdmin(admin.profile)) return;
			const a = (e.target as HTMLElement)?.closest?.('a[href]') as HTMLAnchorElement | null;
			if (!a || a.target == '_blank' || a.closest('.zw-layer')) return;
			const u = new URL(a.href, location.href);
			if (u.origin !== location.origin || !u.pathname.startsWith('/admin')) return;
			e.preventDefault();
			e.stopPropagation();
			openRouteWindow(u.pathname + u.search, a.title || a.textContent?.trim() || u.pathname);
		};
		document.addEventListener('click', h, true);
		return () => document.removeEventListener('click', h, true);
	});
</script>

{#each wm.list as win (win.id)}
	<Window {win}>
		{#if win.path}
			<RouteFrame id={win.id} path={win.path} title={win.title} />
		{:else if win.component}
			{@const C = win.component}
			<C {...win.props || {}} windowId={win.id} />
		{/if}
	</Window>
{/each}
<WindowDock />
<MinimizedDock />

<style>
	:global(html.zbd-win-max),
	:global(html.zbd-win-max body) {
		overflow: hidden !important;
	}
</style>
