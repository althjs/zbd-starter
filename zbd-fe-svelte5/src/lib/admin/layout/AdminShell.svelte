<script lang="ts">
	/**
	 * 관리자 대시보드 틀(React layouts/dashboard/layout) — 머리줄 + 왼쪽 메뉴 + 본문(main).
	 * 단독 앱(chat/mail.zbd.kr)·`?popup=1` 창은 메뉴 없이 본문만(PopupLayout).
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { SIDE_NAV_MINI_WIDTH, SIDE_NAV_WIDTH, TOP_NAV_HEIGHT } from '../config.ts';
	import { media } from '../media.svelte.ts';
	import { getPageBgColor, isIgnoreBorder } from '../meta.svelte.ts';
	import { nav, toggleSideNavByUser } from '../nav.svelte.ts';
	import SideNav from './SideNav.svelte';
	import TopNav from './TopNav.svelte';
	import WindowHost from '../window/WindowHost.svelte';
	import PluginLicenseBanner from './PluginLicenseBanner.svelte';

	let { children }: { children?: Snippet } = $props();
	const border = isIgnoreBorder();
	const bg = $derived(getPageBgColor());
	const ml = $derived(nav.open && media.lgUp ? (nav.mini ? SIDE_NAV_MINI_WIDTH : SIDE_NAV_WIDTH) : 0);
</script>

<div class={cx('MuiStack-root', toClass({ height: '100%' }))} id="main_area">
	<TopNav onNavOpen={() => toggleSideNavByUser(media.lgUp)} />
	<SideNav />
	<div style="display:flex;flex:1 1 auto;max-width:100%;transition:margin 195ms cubic-bezier(0.4, 0, 0.6, 1) 0ms;margin-left:{ml}px">
		<main
			class={cx(
				'MuiBox-root',
				toClass({
					width: '100%',
					minHeight: `calc(100vh - ${TOP_NAV_HEIGHT}px)`,
					marginTop: TOP_NAV_HEIGHT,
					backgroundColor: bg || (border ? 'unset' : 'color-mix(in srgb, var(--theme-color-dark) 4%, transparent)'),
					...(border ? {} : { '@media (min-width:1280px)': { padding: '48px 40px' }, '@media (min-width:1440px)': { padding: '48px 80px' } })
				})
			)}
		>
			<PluginLicenseBanner />
			<div class="MuiStack-root">{@render children?.()}</div>
			<!-- 화면 위 창(React GlobalWindowRenderer) — 페이지를 옮겨도 남는다 -->
			<WindowHost />
		</main>
	</div>
</div>
