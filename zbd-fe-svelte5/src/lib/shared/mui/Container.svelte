<script lang="ts">
	/** MUI Container(maxWidth lg) 호환 — 헤더·푸터·페이지 폭의 틀(좌우 16/24, 최대 1280) */
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		maxWidth = 'lg',
		sx = null,
		class: cls = '',
		children,
		...rest
	}: HTMLAttributes<HTMLDivElement> & { maxWidth?: 'sm' | 'md' | 'lg'; sx?: Sx | null; class?: string; children?: Snippet } = $props();
</script>

<div class={cx(`MuiContainer-root MuiContainer-maxWidth${maxWidth[0].toUpperCase()}${maxWidth.slice(1)}`, cls, toClass(sx))} {...rest}>{@render children?.()}</div>
