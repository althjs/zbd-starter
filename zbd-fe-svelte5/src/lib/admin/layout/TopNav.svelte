<script lang="ts">
	/**
	 * 관리자 머리줄 — ☰ · 로고 · 즐겨찾기 메뉴 | 도움말 · 홈페이지 · 다크모드 · 내 사진.
	 */
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { TOP_NAV_HEIGHT } from '../config.ts';
	import { media } from '../media.svelte.ts';
	import { adminImagePrefix, getCompanyName, getMeta, isSuperAdmin, siteHomepage } from '../meta.svelte.ts';
	import { setCurrentNav } from '../nav.svelte.ts';
	import { admin, getFirstAuthMenuPath } from '../profile.svelte.ts';
	import Icon from '../ui/Icon.svelte';
	import IconButton from '../ui/IconButton.svelte';
	import Tooltip from '../ui/Tooltip.svelte';
	import DarkModeToggle from './DarkModeToggle.svelte';
	import Logo from './Logo.svelte';
	import NavigationView from './NavigationView.svelte';
	import Profile from './Profile.svelte';

	let { onNavOpen }: { onNavOpen: () => void } = $props();

	const logo = $derived(getMeta('SITE', 'logo'));
	const first = $derived(getFirstAuthMenuPath(admin.menus));
	// 홈페이지 바로가기 — 관리자 페이지 탭의 홈페이지 → 국내 홈페이지 탭의 홈페이지 → 같은 주소의 /
	const homepage = $derived(siteHomepage());
	const lightest = 'var(--theme-color-lightest)';
	const sup = $derived(isSuperAdmin(admin.profile));

	function goHome() {
		const x: any = first.homeNav;
		if (x.sub) {
			x.sub.title = x.sub.name;
			x.sub.active = true;
		}
		setCurrentNav(x, sup ? getCompanyName() : '');
	}
</script>

{#if admin.profile?.name}
	<header class="MuiBox-root adm-topnav">
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', minHeight: TOP_NAV_HEIGHT, padding: '0 16px' }))}>
			<div class={cx('MuiStack-root adm-sp2', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
				<div class={cx('MuiBox-root', toClass({ display: 'flex', width: media.mdDown ? 'auto' : 280 }))}>
					<IconButton sx={{ cursor: 'pointer' }} onclick={onNavOpen}><Icon name="menu" sx={{ color: '#fff' }} /></IconButton>
					<a
						class={cx(
							'MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone',
							toClass({ display: media.mdDown ? 'none' : 'flex', flexDirection: 'row', color: '#fff', '&:hover': { color: '#fff' }, maxHeight: 60 })
						)}
						href={first.firstPath || '/'}
						onclick={goHome}
					>
						{#if logo?.name}
							<!-- svelte-ignore a11y_missing_attribute (React 판처럼 alt 없음 — 못 읽으면 같은 깨진 그림) -->
							<img
								class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img', toClass({ maxWidth: 180, height: 'auto', objectFit: 'contain' }))}
								src={`${adminImagePrefix(hostingIdOf(logo) || '')}/site/${logo.name}`}
							/>
						{:else}
							<div class={cx('MuiStack-root', toClass({ paddingLeft: 8, justifyContent: 'center' }))}><Logo height={26} /></div>
						{/if}
					</a>
				</div>
				{#if !media.mdDown}
					<div class={cx('MuiStack-root adm-sp2', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
						<NavigationView linkSx={{ fontSize: 16 }} />
					</div>
				{/if}
			</div>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
				{#if !media.mdDown}
					<!-- 도움말 주소가 없으면 버튼을 숨긴다(예전엔 oye 노션 안내서로 갔다) -->
					{#if getMeta('SITE', 'helpUrl')}
						<Tooltip title="도움말" placement="left">
							<IconButton sx={{ marginLeft: 8 }} size="Small" href={getMeta('SITE', 'helpUrl')} target="_blank" aria-label="도움말"
								><Icon name="help" sx={{ color: lightest, marginLeft: -4 }} /></IconButton
							>
						</Tooltip>
					{/if}
					<Tooltip title="홈페이지" placement="right">
						<IconButton size="Small" href={homepage} target="_blank" aria-label="홈페이지"><Icon name="homepage" sx={{ color: lightest, marginLeft: -4 }} /></IconButton>
					</Tooltip>
					<DarkModeToggle iconColor={lightest} />
				{/if}
				<Profile sx={{ marginLeft: 8 }} />
			</div>
		</div>
	</header>
{/if}

<style>
	.adm-topnav {
		flex-grow: 1;
		backdrop-filter: blur(6px);
		background-color: var(--adm-header-bg);
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 1100;
		color: #fff;
	}
	/* MUI Stack spacing={2} — 둘째 자식부터 왼쪽 16 */
	:global(.adm-sp2 > :not(style) ~ :not(style)) {
		margin-left: 16px;
	}
</style>
