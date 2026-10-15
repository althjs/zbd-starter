<script lang="ts">
	/**
	 * 쓰기 권한(RW) 이 있어야 누를 수 있는 버튼(React components/common/AuthButton) — 지금 메뉴의 권한(activeNav().auth) 을 본다.
	 * 읽기 권한(R) 뿐이면 막히고, 글자 버튼이면 「권한이 없습니다」 로 바뀐다. 메뉴 권한이 아직 없으면 그리지 않는다.
	 *   <AuthButton variant="contained" onclick={save}>저장</AuthButton>
	 */
	import type { Snippet } from 'svelte';
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import { activeNav } from '../nav.svelte.ts';
	import Button from './Button.svelte';

	let {
		variant = 'contained',
		size = 'medium',
		color = 'primary',
		disabled = false,
		fullWidth = false,
		sx = null,
		text = '',
		onclick,
		startIcon,
		children
	}: {
		variant?: 'text' | 'outlined' | 'contained';
		size?: 'small' | 'medium' | 'large';
		color?: 'primary' | 'inherit' | 'error' | 'warning' | 'success';
		disabled?: boolean;
		fullWidth?: boolean;
		sx?: Sx | null;
		text?: string;
		onclick?: (e: MouseEvent) => void;
		startIcon?: Snippet;
		children?: Snippet;
	} = $props();
	const auth = $derived(activeNav()?.auth || '');
</script>

{#if auth}
	<Button {variant} {size} {color} {fullWidth} {sx} {onclick} {startIcon} disabled={auth != 'RW' || disabled}>
		{#if text}{auth != 'RW' ? '권한이 없습니다' : text}{:else}{@render children?.()}{/if}
	</Button>
{/if}
