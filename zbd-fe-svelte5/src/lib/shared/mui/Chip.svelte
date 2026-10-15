<script lang="ts">
	/**
	 * MUI Chip 호환 — 작은 이름표(게시글 「공지」·스토어 분류·정렬 등).
	 * onclick 이 있으면 누를 수 있는 칩(MuiChip-clickable), icon 은 머티리얼 아이콘 이름.
	 */
	import { cx, sx as toClass, type Sx } from './sx';

	type Color = 'primary' | 'default' | 'info' | 'warning' | 'success' | 'error';
	let {
		label,
		variant = 'outlined',
		color = 'primary',
		size = 'medium',
		icon = '',
		iconSize = 0,
		class: cls = 'state-chip',
		sx = null,
		onclick = undefined
	}: {
		label: string;
		variant?: 'outlined' | 'filled';
		color?: Color;
		size?: 'small' | 'medium';
		icon?: string;
		iconSize?: number;
		class?: string;
		sx?: Sx | null;
		onclick?: (e: MouseEvent) => void;
	} = $props();

	const S = $derived(size == 'small' ? 'Small' : 'Medium');
	const C = $derived(color[0].toUpperCase() + color.slice(1));
</script>

{#if onclick}
	<div
		role="button"
		tabindex="0"
		class={cx(`MuiButtonBase-root MuiChip-root MuiChip-${variant} MuiChip-size${S} MuiChip-color${C} MuiChip-clickable MuiChip-clickableColor${C} MuiChip-${variant}${C}`, cls, toClass(sx))}
		{onclick}
		onkeydown={(e) => (e.key == 'Enter' || e.key == ' ') && (e.preventDefault(), onclick?.(e as any))}
	>
		{#if icon}<span
				class="material-icons notranslate MuiIcon-root MuiIcon-fontSizeMedium MuiChip-icon MuiChip-icon{S} MuiChip-iconColor{C}"
				aria-hidden="true"
				style={iconSize ? `font-size:${iconSize}px !important` : ''}>{icon}</span
			>{/if}
		<span class="MuiChip-label MuiChip-label{S}">{label}</span>
	</div>
{:else}
	<div class={cx(`MuiChip-root MuiChip-${variant} MuiChip-size${S} MuiChip-color${C} MuiChip-${variant}${C}`, cls, toClass(sx))}>
		{#if icon}<span
				class="material-icons notranslate MuiIcon-root MuiIcon-fontSizeMedium MuiChip-icon MuiChip-icon{S} MuiChip-iconColor{C}"
				aria-hidden="true"
				style={iconSize ? `font-size:${iconSize}px !important` : ''}>{icon}</span
			>{/if}
		<span class="MuiChip-label MuiChip-label{S}">{label}</span>
	</div>
{/if}
