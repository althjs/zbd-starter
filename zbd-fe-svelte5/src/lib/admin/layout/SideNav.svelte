<script lang="ts">
	/**
	 * 왼쪽 메뉴(React layouts/dashboard/SideNav) — 권한 메뉴 트리 · 메뉴 검색 · 프로필 · 보조 메뉴(고객센터).
	 * PC(≥1280) 는 고정 메뉴(펼침 280 / 축약 88 / 숨김), 그보다 좁으면 서랍(☰) — 주소가 바뀌면 서랍을 닫는다.
	 * 지금 메뉴(페이지 제목·이동 경로) 맞추기도 여기서(syncCurrentNav).
	 */
	import { untrack } from 'svelte';
	import { page } from '$app/state';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { SIDE_NAV_MINI_WIDTH, SIDE_NAV_WIDTH, TOP_NAV_HEIGHT } from '../config.ts';
	import { media } from '../media.svelte.ts';
	import { getCompanyName, isSuperAdmin } from '../meta.svelte.ts';
	import { nav, restoreSideNav, setOpenSideNav, syncCurrentNav, toggleSideNavByUser } from '../nav.svelte.ts';
	import { admin } from '../profile.svelte.ts';
	import type { NavItem } from '../types.ts';
	import SvgIcon from '../ui/SvgIcon.svelte';
	import Profile from './Profile.svelte';
	import SideNavItem from './SideNavItem.svelte';
	import SideNavMini from './SideNavMini.svelte';

	const pathname = $derived(page.url.pathname);
	const sup = $derived(isSuperAdmin(admin.profile));

	const all = $derived(
		(admin.menus || []).map((m): NavItem => ({
			menuId: m.menuId,
			title: m.name,
			desc: m.description || '',
			auth: m.auth || '',
			path: m.path,
			special: m.special || '',
			icon: m.icon || '',
			items: m.items ? m.items.map((s) => ({ menuId: s.menuId, title: s.name, desc: s.description || '', auth: s.auth || '', path: s.path, icon: s.icon || '' })) : undefined
		}))
	);
	const items = $derived(all.filter((o) => (o.menuId || 0) > 0));
	const subMenus = $derived(all.filter((o) => (o.menuId || 0) < 0 && (o.menuId || 0) > -90000));

	const miniItems = $derived(items);

	// 지금 메뉴 맞추기 — 주소·메뉴가 바뀔 때
	$effect(() => {
		if (!admin.menus) return;
		syncCurrentNav(items, subMenus, pathname, sup ? getCompanyName() : '');
	});
	// 휴대폰 서랍은 주소가 바뀌면 닫는다
	$effect(() => {
		void pathname;
		if (!media.lgUp) setOpenSideNav(false);
	});
	// PC 폭이 되면(미리 띄운 화면이 보일 때·창을 넓힐 때) 저장된 상태로
	$effect(() => {
		if (media.lgUp) untrack(restoreSideNav);
	});

	let searchInput = $state('');
	let keyword = $state('');
	function onKey(e: KeyboardEvent) {
		if (e.key == 'b' && (e.metaKey || e.ctrlKey)) toggleSideNavByUser(media.lgUp);
	}

	const closeOnMobile = () => !media.lgUp && setOpenSideNav(false);
	const btnSx = (on: boolean) => ({
		backgroundColor: on ? 'var(--theme-color-main)' : '#404040',
		color: '#fff',
		borderRadius: 6,
		padding: 4.8,
		'&:hover': { backgroundColor: on ? 'var(--theme-color-dark)' : '#2d2d2d' }
	});
</script>

<svelte:window onkeydown={onKey} />

{#snippet content()}
	<div class="adm-sidenav-scroll" style:height="calc(100% - {TOP_NAV_HEIGHT}px)" style:margin-top="{TOP_NAV_HEIGHT}px">
		<div class={cx('MuiBox-root', toClass({ display: 'flex', flexDirection: 'column', height: '100%' }))}>
			<nav class={cx('MuiBox-root', toClass({ flexGrow: 1, padding: '24px 16px' }))}>
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 4, marginBottom: 16 }))}>
					<div class={cx('MuiFormControl-root MuiTextField-root', toClass({ flex: 1 }))}>
						<div
							class="MuiInputBase-root MuiFilledInput-root MuiInputBase-colorPrimary MuiInputBase-formControl MuiInputBase-sizeSmall MuiInputBase-hiddenLabel"
							class:MuiInputBase-adornedEnd={!!searchInput}
						>
							<input
								class={cx('MuiInputBase-input MuiFilledInput-input MuiInputBase-inputSizeSmall MuiInputBase-inputHiddenLabel', toClass({ paddingTop: 4.8, paddingBottom: 4.8, fontSize: 13 }))}
								placeholder="메뉴검색"
								bind:value={searchInput}
								onkeydown={(e) => {
									if (e.key === 'Enter') keyword = searchInput.trim();
									if (e.key === 'Escape') searchInput = keyword = '';
								}}
							/>
							{#if searchInput}
								<div class="MuiInputAdornment-root MuiInputAdornment-positionEnd">
									<button type="button" class={cx('MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeSmall', toClass({ padding: 2 }))} onclick={() => (searchInput = keyword = '')}
										><SvgIcon name="Close" sx={{ fontSize: 16 }} /></button
									>
								</div>
							{/if}
						</div>
					</div>
					<button type="button" class={cx('MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeSmall', toClass(btnSx(false)))} aria-label="메뉴 검색" onclick={() => (keyword = searchInput.trim())}
						><SvgIcon name="Search" sx={{ fontSize: 18 }} /></button
					>
				</div>
				<nav class={cx('MuiList-root MuiList-padding', toClass({ width: '100%', maxWidth: 360, display: 'flex', flexDirection: 'column', gap: 8 }))} aria-labelledby="nested-list-subheader">
					{#each items as it (it.title)}
						<SideNavItem item={it} searchKeyword={keyword} superAdmin={sup} onclick={closeOnMobile} />
					{/each}
				</nav>
				<hr class={cx('MuiDivider-root MuiDivider-fullWidth', toClass({ borderColor: 'var(--theme-color-alpha12)', paddingTop: 16 }))} />
				<div class={cx('MuiStack-root', toClass({ marginTop: 16, marginLeft: 16 }))}><Profile useLogout /></div>
				{#if subMenus.length}
					<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ marginTop: 24, background: 'var(--gt-200, #f2f2f2)' }))}>
						{#each subMenus as it (it.title)}
							<SideNavItem item={it} searchKeyword={keyword} onclick={closeOnMobile} />
						{/each}
					</div>
				{/if}
			</nav>
		</div>
	</div>
{/snippet}

{#if admin.profile?.name}
	{#if media.lgUp}
		<div class={cx('MuiDrawer-root MuiDrawer-anchorLeft MuiDrawer-docked', toClass({ zIndex: 1099, flex: '0 0 auto' }))}>
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-elevation0 MuiDrawer-paper MuiDrawer-paperAnchorLeft MuiDrawer-paperAnchorDockedLeft',
					toClass({
						width: nav.mini ? SIDE_NAV_MINI_WIDTH : SIDE_NAV_WIDTH,
						borderRight: '1px solid var(--theme-color-alpha12)',
						transition: 'margin 195ms cubic-bezier(0.4, 0, 0.6, 1) 0ms',
						...(nav.open ? {} : { marginLeft: `${-SIDE_NAV_WIDTH}px !important` })
					})
				)}
			>
				{#if nav.mini}
					<div class="adm-sidenav-scroll" style:height="calc(100% - {TOP_NAV_HEIGHT}px)" style:margin-top="{TOP_NAV_HEIGHT}px">
						<SideNavMini items={miniItems} {subMenus} />
					</div>
				{:else}
					{@render content()}
				{/if}
			</div>
		</div>
	{:else}
		<Drawer open={nav.open} anchor="left" class="adm-sidenav-modal" paperSx={{ width: 280, borderColor: 'var(--theme-color-alpha12)' }} onclose={() => setOpenSideNav(false)}>
			{@render content()}
		</Drawer>
	{/if}
{/if}
