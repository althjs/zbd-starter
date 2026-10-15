<script lang="ts">
	/** MUI IconButton 호환 — 둥근 아이콘 버튼 */
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		size = 'Medium',
		sx = null,
		disabled = false,
		class: cls = '',
		children,
		...rest
	}: HTMLButtonAttributes & { size?: 'Small' | 'Medium' | 'Large'; sx?: Sx | null; class?: string; children?: Snippet } = $props();
</script>

<button type="button" class={cx('MuiButtonBase-root', disabled && 'Mui-disabled', `MuiIconButton-root MuiIconButton-size${size}`, cls, toClass(sx))} {disabled} {...rest}>{@render children?.()}</button
>
