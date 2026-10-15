<script lang="ts">
	/** 라이트 / 다크 / 시스템 고르기 (React components/theme/DarkModeToggle) */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { dark, setDarkMode, type ColorScheme } from '../dark.svelte.ts';
	import SvgIcon from '../ui/SvgIcon.svelte';
	import Tooltip from '../ui/Tooltip.svelte';

	let { iconColor = 'inherit' }: { iconColor?: string } = $props();
	let anchor = $state<HTMLElement | null>(null);
	let pos = $state({ top: 0, right: 0 });
	const title = $derived(dark.scheme == 'auto' ? `자동 (현재: ${dark.resolved == 'dark' ? '다크' : '라이트'})` : dark.resolved == 'dark' ? '다크 모드' : '라이트 모드');
	const icon = $derived(dark.scheme == 'auto' ? 'SettingsBrightness' : dark.resolved == 'dark' ? 'DarkMode' : 'LightMode');
	function open(e: MouseEvent) {
		anchor = e.currentTarget as HTMLElement;
		const r = anchor.getBoundingClientRect();
		pos = { top: r.bottom + 12, right: window.innerWidth - r.right };
	}
	function pick(m: ColorScheme) {
		setDarkMode(m);
		anchor = null;
	}
	const ITEMS: { m: ColorScheme; icon: string; label: string; sub?: string }[] = [
		{ m: 'light', icon: 'LightMode', label: '라이트 모드' },
		{ m: 'dark', icon: 'DarkMode', label: '다크 모드' },
		{ m: 'auto', icon: 'SettingsBrightness', label: '시스템 설정', sub: '자동으로 전환' }
	];
</script>

<Tooltip {title}>
	<button type="button" class={cx('MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeSmall', toClass({ marginLeft: 8, color: iconColor }))} aria-haspopup="true" aria-label={title} onclick={open}
		><SvgIcon name={icon} /></button
	>
</Tooltip>
{#if anchor}
	<div use:portal role="presentation" class="MuiPopover-root MuiMenu-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (anchor = null)}></div>
		<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation0 MuiPopover-paper MuiMenu-paper adm-dm-paper" style="top:{pos.top}px;right:{pos.right}px">
			<ul class="MuiList-root MuiList-padding MuiMenu-list" role="menu">
				{#each ITEMS as it (it.m)}
					<li class="MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters" role="menuitem" tabindex="0" onclick={() => pick(it.m)} onkeydown={(e) => e.key == 'Enter' && pick(it.m)}>
						<div class="MuiListItemIcon-root"><SvgIcon name={it.icon} size="Small" /></div>
						{#if it.sub}
							<div class="MuiBox-root">
								<div class="MuiListItemText-root"><span class="MuiTypography-root MuiTypography-body1">{it.label}</span></div>
								<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2)">{it.sub}</span>
							</div>
						{:else}
							<div class="MuiListItemText-root"><span class="MuiTypography-root MuiTypography-body1">{it.label}</span></div>
						{/if}
						{#if dark.scheme == it.m}<SvgIcon name="Check" size="Small" sx={{ marginLeft: 8 }} />{/if}
					</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}
