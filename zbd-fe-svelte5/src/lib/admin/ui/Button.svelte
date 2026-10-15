<script lang="ts">
	/**
	 * 관리자 버튼 — 공용 shared/mui/MuiButton 을 감싸 관리자 쪽 칸(color·fullWidth)을 이름표로 더한다. 크기·모양은 admin.css.
	 *   <Button variant="contained" onclick={save}>저장</Button>
	 *   <Button variant="contained" color="inherit" fullWidth>취소</Button>   color: primary|inherit|error|warning|success
	 * 권한(RW) 이 없으면 막는 버튼은 AuthButton.
	 */
	import type { Snippet } from 'svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import { cx, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		variant = 'text',
		size = 'medium',
		color = 'primary',
		fullWidth = false,
		href = undefined,
		disabled = false,
		class: cls = '',
		sx = null,
		onclick,
		startIcon,
		children
	}: {
		variant?: 'text' | 'outlined' | 'contained';
		size?: 'small' | 'medium' | 'large';
		color?: 'primary' | 'inherit' | 'error' | 'warning' | 'success';
		fullWidth?: boolean;
		href?: string;
		disabled?: boolean;
		class?: string;
		sx?: Sx | null;
		onclick?: (e: MouseEvent) => void;
		startIcon?: Snippet;
		children?: Snippet;
	} = $props();
	const C = $derived(color[0].toUpperCase() + color.slice(1));
	const S = $derived(size[0].toUpperCase() + size.slice(1));
</script>

<MuiButton
	{variant}
	{size}
	{href}
	{disabled}
	{sx}
	{onclick}
	{startIcon}
	class={cx(color != 'primary' && `MuiButton-${variant}${C} MuiButton-color${C}`, `MuiButton-${variant}Size${S}`, fullWidth && 'MuiButton-fullWidth', cls)}>{@render children?.()}</MuiButton
>
