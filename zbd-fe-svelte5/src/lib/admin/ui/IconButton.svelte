<script lang="ts">
	/**
	 * 관리자 아이콘 버튼 — 공용 shared/mui/IconButton 을 감싸고, href 가 있으면 같은 이름표의 링크(<a>)로 그린다(MUI IconButton href).
	 *   <IconButton size="Small" href="https://…" target="_blank"><Icon name="help" /></IconButton>
	 */
	import type { Snippet } from 'svelte';
	import Base from '#lib/shared/mui/IconButton.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		size = 'Medium',
		sx = null,
		href = undefined,
		target = undefined,
		disabled = false,
		color = '',
		class: cls = '',
		children,
		...rest
	}: {
		size?: 'Small' | 'Medium' | 'Large';
		sx?: Sx | null;
		href?: string;
		target?: string;
		disabled?: boolean;
		color?: '' | 'Primary';
		class?: string;
		children?: Snippet;
		[k: string]: any;
	} = $props();
</script>

{#if href}
	<a
		class={cx('MuiButtonBase-root MuiIconButton-root', color && `MuiIconButton-color${color}`, `MuiIconButton-size${size}`, cls, toClass(sx))}
		{href}
		{target}
		rel={target == '_blank' ? 'noopener' : undefined}
		{...rest}>{@render children?.()}</a
	>
{:else}
	<Base {size} {sx} {disabled} class={cx(color && `MuiIconButton-color${color}`, cls)} {...rest}>{@render children?.()}</Base>
{/if}
