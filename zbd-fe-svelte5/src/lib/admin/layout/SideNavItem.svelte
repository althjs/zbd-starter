<script lang="ts">
	/**
	 * 왼쪽 메뉴 한 줄(React layouts/dashboard/components/SideNavItem) — 하위가 있으면 눌러 펼치고, 없으면 링크.
	 * 검색어가 하위 이름에 있으면 펼치고 그 이름을 빨갛게. 하위 줄 오른쪽에 「새 창」 버튼.
	 */
	import { page } from '$app/state';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { slide } from 'svelte/transition';
	import type { NavItem } from '../types.ts';
	import Icon from '../ui/Icon.svelte';
	import IconButton from '../ui/IconButton.svelte';
	import MenuIcon from './MenuIcon.svelte';
	import { bestSubPath, pathUnder } from '../nav.svelte.ts';
	import { openRouteWindow } from '../window/windows.svelte.ts';

	let { item, searchKeyword = '', badge = 0, superAdmin = false, onclick = () => {} }: { item: NavItem; searchKeyword?: string; badge?: number; superAdmin?: boolean; onclick?: () => void } = $props();

	const pathname = $derived(page.url.pathname);
	const kw = $derived(String(searchKeyword || '').trim());
	const items = $derived(item.items || []);
	// 같은 묶음에 아이콘이 있으면 아이콘 없는 줄도 같은 자리를 비워 글자 줄을 맞춘다
	const anyIcon = $derived(items.some((it) => !!it.icon));
	const hasSubMatch = $derived(!!kw && items.some((it) => String(it.title || '').includes(kw)));
	const bestSub = $derived(bestSubPath(items, pathname));
	const subActive = (s: NavItem) => !!bestSub && String(s.path || '').split('?')[0] === bestSub;
	const active = $derived.by(() => {
		let a = (item.path ? pathname === item.path : false) || pathname == item.path || (!items.length && item.path !== '/' && pathUnder(pathname, item.path));
		if (pathname.indexOf('/admin') !== 0) a = false; // React 와 같은 순서 — 하위 메뉴(/dev/… 개발자 도구)는 /admin 밖이어도 맞으면 활성
		return a || items.some(subActive);
	});
	let open = $state(false);
	// 주소가 바뀌면 지금 메뉴만 펼친다 · 검색어가 바뀌면 맞는 것만
	$effect(() => {
		void pathname;
		open = active || hasSubMatch;
	});
	$effect(() => {
		void kw;
		open = kw ? hasSubMatch : active;
	});

	const textSx = $derived({ fontWeight: active ? '800' : '500', color: active ? 'var(--theme-color-main)' : 'var(--adm-n500)', fontSize: '14px' });
</script>

<div class="MuiBox-root">
	<div
		class={cx(
			'MuiButtonBase-root MuiListItemButton-root MuiListItemButton-gutters',
			toClass({ paddingTop: 12, paddingBottom: 12, paddingRight: 8, borderRadius: 8, background: active ? 'var(--theme-color-alpha8)' : 'transparent' })
		)}
		role="button"
		tabindex="0"
		onclick={() => (open = !open)}
		onkeydown={(e) => e.key == 'Enter' && (open = !open)}
	>
		<div class={cx('MuiListItemIcon-root', toClass({ color: 'var(--adm-n500)' }))}>
			{#if item.icon}
				<span
					class={cx(
						'MuiBox-root',
						toClass({ alignItems: 'center', color: active ? 'var(--theme-color-main)' : 'var(--adm-n400)', display: 'inline-flex', justifyContent: 'center', marginRight: '0' })
					)}><MenuIcon icon={item.icon} /></span
				>
			{/if}
		</div>
		<div class={cx('MuiListItemText-root', toClass({ marginTop: 0, marginBottom: 0 }))}>
			<span class="MuiTypography-root MuiTypography-body1 MuiListItemText-primary">
				{#if items.length}
					<span class={cx('MuiTypography-root MuiTypography-body1', toClass(textSx))}>{item.title}</span>
				{:else}
					<a class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineAlways', toClass({ padding: 0 }))} href={item.path} {onclick}
						><p class={cx('MuiTypography-root MuiTypography-body1', toClass({ width: '100%', display: 'inline-block', ...textSx }))}>
							{item.title}{#if badge > 0}<span
									class={cx(
										'MuiBox-root menu-new-badge',
										toClass({
											display: 'inline-flex',
											alignItems: 'center',
											justifyContent: 'center',
											verticalAlign: 'top',
											marginLeft: 4,
											minWidth: 18,
											height: 18,
											padding: '0 4.8px',
											borderRadius: 9,
											backgroundColor: 'var(--theme-color-main)',
											color: '#fff',
											fontSize: 11,
											fontWeight: 700,
											lineHeight: '18px'
										})
									)}>{badge > 99 ? '99+' : badge}</span
								>{/if}
						</p></a
					>
				{/if}
			</span>
		</div>
		{#if items.length}
			<svg class={cx('MuiSvgIcon-root MuiSvgIcon-fontSizeMedium', toClass({ color: 'var(--adm-action)', fontSize: 16, marginLeft: 16 }))} focusable="false" aria-hidden="true" viewBox="0 0 24 24"
				><svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
					><path d={open || active ? 'm12 8-6 6 1.41 1.41L12 10.83l4.59 4.58L18 14z' : 'M16.59 8.59 12 13.17 7.41 8.59 6 10l6 6 6-6z'}></path></svg
				></svg
			>
		{/if}
	</div>
	{#if items.length && open}
		<div class="MuiCollapse-root MuiCollapse-vertical MuiCollapse-entered" transition:slide={{ duration: 200 }}>
			<div class="MuiCollapse-wrapper MuiCollapse-vertical">
				<div class="MuiCollapse-wrapperInner MuiCollapse-vertical">
					<div class={cx('MuiList-root', toClass({ padding: 0 }))}>
						{#each items as s (s.menuId ?? s.title)}
							{@const sa = subActive(s)}
							<div class={cx('MuiBox-root', toClass({ position: 'relative' }))}>
								<a class={cx('MuiButtonBase-root MuiListItemButton-root MuiListItemButton-gutters', toClass({ paddingLeft: 36, paddingRight: 64, borderRadius: 8 }))} href={s.path} {onclick}>
									<span
										class={cx(
											'MuiBox-root',
											toClass({
												cursor: 'pointer',
												flexGrow: 1,
												fontFamily: 'var(--mui-font)',
												fontSize: 14,
												lineHeight: '24px',
												whiteSpace: 'nowrap',
												color: sa ? 'var(--theme-color-main)' : 'var(--adm-n500)'
											})
										)}
									>
										<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', width: '100%' }))}>
											<p
												class={cx(
													'MuiTypography-root MuiTypography-body2',
													toClass({ fontWeight: sa ? 700 : 500, ...(kw && String(s.title || '').includes(kw) ? { color: 'var(--adm-error-main)', fontWeight: 700 } : {}) })
												)}
											>
												{#if s.icon}<span class="sub-ic"><MenuIcon icon={s.icon} sx={{ fontSize: 18 }} /></span>{:else if anyIcon}<span class="sub-ic blank"></span>{/if}{s.title}
											</p>
										</div>
									</span>
								</a>
								{#if (s.menuId || 0) > 0}
									<div class={cx('MuiBox-root', toClass({ position: 'absolute', right: 8, top: '50%', transform: 'translateY(-50%)', display: 'flex', alignItems: 'center', zIndex: 2 }))}>
										{#if superAdmin}
											<!-- 창으로 열기 — 화면 위 창(React WindowOpenerForRoutePath) -->
											<IconButton
												size="Small"
												sx={{ float: 'right', marginRight: '4px' }}
												onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), openRouteWindow(String(s.path), s.title))}
												onmousedown={(e: MouseEvent) => e.stopPropagation()}
											>
												<Icon name="web_asset" size="Small" sx={{ fontSize: 15, color: 'var(--gt-600, #bdbdbd)' }} />
											</IconButton>
										{/if}
										<IconButton size="Small" onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), window.open(s.path))} onmousedown={(e: MouseEvent) => e.stopPropagation()}>
											<Icon name="open_in_new" size="Small" sx={{ fontSize: 15, color: 'var(--gt-600, #bdbdbd)' }} />
										</IconButton>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.sub-ic {
		display: inline-flex;
		vertical-align: middle;
		margin-right: 6px;
	}
	.sub-ic.blank {
		width: 18px;
	}
	.sub-ic :global(img) {
		width: 18px !important;
		height: 18px !important;
	}
</style>
