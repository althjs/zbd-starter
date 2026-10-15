<script lang="ts">
	/**
	 * 이름표(MUI Chip) — 아이콘·지우기 ✕·누르기까지. 공용 shared/mui/Chip 은 글자만이라 협업 화면은 이걸.
	 *   <Chip size="small" variant="outlined" icon="attach_file" label="a.pdf (3KB)" onclick={…} ondelete={…} />
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Icon from '../../ui/Icon.svelte';
	import Av from './Av.svelte';
	type C = 'default' | 'primary' | 'error' | 'warning' | 'success';
	let {
		label,
		size = 'medium',
		variant = 'filled',
		color = 'default',
		icon = '',
		iconSx = null,
		avatar = undefined,
		avatarSx = null,
		sx = null,
		onclick,
		ondelete
	}: {
		label: string;
		size?: 'small' | 'medium';
		variant?: 'filled' | 'outlined';
		color?: C;
		icon?: string;
		iconSx?: Sx | null;
		avatar?: string;
		avatarSx?: Sx | null;
		sx?: Sx | null;
		onclick?: () => void;
		ondelete?: () => void;
	} = $props();
	const cap = (s: string) => s[0].toUpperCase() + s.slice(1);
	const S = $derived(size == 'small' ? 'Small' : 'Medium');
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
	class={cx(
		`MuiChip-root MuiChip-${variant} MuiChip-size${S} MuiChip-color${cap(color)} MuiChip-${variant}${cap(color)}`,
		onclick && 'MuiChip-clickable',
		ondelete && 'MuiChip-deletable',
		toClass(sx)
	)}
	role={onclick ? 'button' : undefined}
	tabindex={onclick ? 0 : undefined}
	onclick={() => onclick?.()}
	onkeydown={(e) => onclick && e.key == 'Enter' && onclick()}
>
	{#if avatar !== undefined}<Av src={avatar} size={18} class="MuiChip-avatar MuiChip-avatar{S}" sx={avatarSx} />{/if}
	{#if icon}<Icon name={icon} class="MuiChip-icon MuiChip-icon{S}" sx={iconSx} />{/if}
	<span class="MuiChip-label MuiChip-label{S}">{label}</span>
	{#if ondelete}
		<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiChip-deleteIcon MuiChip-deleteIcon{S}" viewBox="0 0 24 24" aria-hidden="true" onclick={(e) => (e.stopPropagation(), ondelete())}
			><path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2m5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12z"></path></svg
		>
	{/if}
</div>
