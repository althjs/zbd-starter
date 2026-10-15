<script lang="ts">
	/**
	 * PC 축약 메뉴 — 아이콘 + 짧은 이름(폭 88). 하위가 있으면 마우스를 올리면 오른쪽에 하위 목록, 누르면 첫 하위로(React SideNavMini).
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import type { NavItem } from '../types.ts';
	import Icon from '../ui/Icon.svelte';
	import MenuIcon from './MenuIcon.svelte';
	import { bestSubPath, pathUnder } from '../nav.svelte.ts';

	let { items, subMenus = [], badges = {} }: { items: NavItem[]; subMenus?: NavItem[]; badges?: Record<string, number> } = $props();
	const pathname = $derived(page.url.pathname);
	const fullPath = $derived(page.url.pathname + page.url.search);
	let hover = $state<{ key: string; rect: DOMRect; item: NavItem } | null>(null);
	let timer: any;
	const keep = () => clearTimeout(timer);
	const closeSoon = () => {
		clearTimeout(timer);
		timer = setTimeout(() => (hover = null), 150);
	};
	$effect(() => {
		void fullPath;
		hover = null;
	});

	function isActive(it: NavItem) {
		if (pathname.indexOf('/admin') !== 0) return false;
		if (it.items?.length)
			return it.items.some((s) => {
				const p = (s.path || '').split('?')[0];
				return pathUnder(pathname, p);
			});
		if (!it.path) return false;
		return it.path === pathname || (it.path !== '/' && pathUnder(pathname, it.path));
	}
	const subActive = (s: NavItem, sibs: NavItem[] = []) =>
		!!s.path && (s.path.includes('?') ? fullPath === s.path : s.path === pathname || (pathUnder(pathname, s.path) && bestSubPath(sibs.length ? sibs : [s], pathname) === s.path));
	function rowSx(it: NavItem, key: string) {
		const a = isActive(it);
		return toClass({
			position: 'relative',
			display: 'flex',
			flexDirection: 'column',
			alignItems: 'center',
			width: '100%',
			padding: '8px 4px',
			borderRadius: 12,
			textDecoration: 'none',
			'&:link, &:visited': { color: 'inherit' },
			cursor: it.disabled ? 'default' : 'pointer',
			opacity: it.disabled ? 0.5 : 1,
			backgroundColor: a || hover?.key === key ? 'var(--theme-color-alpha8)' : 'transparent',
			'&:hover': { backgroundColor: 'var(--theme-color-alpha8)' }
		});
	}
	function click(it: NavItem, key: string, e: MouseEvent) {
		const first = (it.items || []).find((s) => !!s.path && !s.disabled);
		if (first?.path) {
			hover = null;
			if (first.external) window.open(first.path, '_blank', 'noopener');
			else goto(first.path);
			return;
		}
		keep();
		hover = hover?.key === key ? null : { key, rect: (e.currentTarget as HTMLElement).getBoundingClientRect(), item: it };
	}
</script>

{#snippet body(it: NavItem)}
	{@const a = isActive(it)}
	{@const badge = it.path ? badges[it.path] || 0 : 0}
	<span class={cx('MuiBox-root', toClass({ position: 'relative', display: 'inline-flex', color: a ? 'var(--theme-color-main)' : 'var(--adm-n400)', '& svg, & .MuiIcon-root': { fontSize: 24 } }))}>
		<MenuIcon icon={it.icon} />
		{#if badge > 0}<span
				class={cx(
					'MuiBox-root',
					toClass({
						position: 'absolute',
						top: -6,
						right: -10,
						minWidth: 16,
						height: 16,
						padding: '0 3.2px',
						borderRadius: 8,
						backgroundColor: 'var(--theme-color-main)',
						color: '#fff',
						fontSize: 10,
						fontWeight: 700,
						lineHeight: '16px',
						textAlign: 'center'
					})
				)}>{badge > 99 ? '99+' : badge}</span
			>{/if}
	</span>
	<span
		class={cx(
			'MuiTypography-root MuiTypography-body1',
			toClass({
				marginTop: 4,
				fontSize: 11,
				lineHeight: 1.25,
				fontWeight: a ? 700 : 500,
				color: a ? 'var(--theme-color-main)' : 'var(--gt-700, #909090)',
				textAlign: 'center',
				wordBreak: 'keep-all',
				display: '-webkit-box',
				WebkitLineClamp: 2,
				WebkitBoxOrient: 'vertical',
				overflow: 'hidden'
			})
		)}>{it.title}</span
	>
	{#if it.items?.length}<span class={cx('MuiBox-root', toClass({ position: 'absolute', right: 4, top: 14, fontSize: 16, lineHeight: 1, color: 'var(--gt-500, #cfcfcf)' }))}>›</span>{/if}
{/snippet}

{#snippet row(it: NavItem, key: string)}
	{#if it.items?.length}
		<div
			class={cx('MuiBox-root', rowSx(it, key))}
			role="button"
			tabindex="0"
			onmouseenter={(e) => (keep(), (hover = { key, rect: (e.currentTarget as HTMLElement).getBoundingClientRect(), item: it }))}
			onmouseleave={closeSoon}
			onclick={(e) => click(it, key, e)}
			onkeydown={() => {}}
		>
			{@render body(it)}
		</div>
	{:else if it.disabled || !it.path}
		<div class={cx('MuiBox-root', rowSx(it, key))}>{@render body(it)}</div>
	{:else if it.external}
		<a class={cx('MuiBox-root', rowSx(it, key))} href={it.path} target="_blank" rel="noopener">{@render body(it)}</a>
	{:else}
		<a class={cx('MuiBox-root', rowSx(it, key))} href={it.path}>{@render body(it)}</a>
	{/if}
{/snippet}

<nav class={cx('MuiBox-root', toClass({ display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 4, padding: '16px 8px' }))}>
	{#each items as it, i (`m_${it.menuId ?? i}_${it.title}`)}{@render row(it, `m_${it.menuId ?? i}_${it.title}`)}{/each}
	{#if subMenus.length}
		<div class={cx('MuiBox-root', toClass({ borderTop: '1px solid var(--theme-color-alpha12)', marginTop: 8, paddingTop: 8 }))}>
			{#each subMenus as it, i (`s_${it.menuId ?? i}_${it.title}`)}{@render row(it, `s_${it.menuId ?? i}_${it.title}`)}{/each}
		</div>
	{/if}
</nav>

{#if hover}
	<div use:portal class="adm-mini-pop" style="top:{hover.rect.top}px;left:{hover.rect.right + 6}px" role="menu" tabindex="-1" onmouseenter={keep} onmouseleave={closeSoon}>
		<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 adm-mini-paper">
			{#each hover.item.items || [] as s, i (`${s.menuId ?? i}_${s.title}`)}
				{@const sa = subActive(s, hover.item.items || [])}
				<a class="adm-mini-sub" class:on={sa} href={s.path} style:padding-left="{16 + (s.miniDepth || 0) * 12}px" onclick={() => (hover = null)}>
					{#if s.miniIcon}<Icon name={s.miniIcon} sx={{ fontSize: 18, marginRight: 8, color: sa ? 'var(--theme-color-main)' : 'var(--adm-n400)' }} />{:else if s.icon}<span
							style="display:inline-flex;margin-right:8px"><MenuIcon icon={s.icon} sx={{ fontSize: 18 }} /></span
						>{/if}{s.title}
				</a>
			{/each}
		</div>
	</div>
{/if}

<style>
	.adm-mini-pop {
		position: fixed;
		z-index: 1101;
	}
	.adm-mini-paper {
		padding: 8px 0;
		min-width: 180px;
		max-height: 70vh;
		overflow-y: auto;
		border-radius: 16px;
	}
	.adm-mini-sub {
		display: flex;
		align-items: center;
		padding: 8px 16px;
		font-size: 14px;
		white-space: nowrap;
		text-decoration: none;
		color: var(--adm-n500);
		font-weight: 500;
	}
	.adm-mini-sub.on {
		color: var(--theme-color-main);
		font-weight: 700;
	}
	.adm-mini-sub:hover {
		background-color: var(--theme-color-alpha8);
	}
</style>
