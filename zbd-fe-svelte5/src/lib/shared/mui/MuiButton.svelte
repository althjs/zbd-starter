<script lang="ts">
	/**
	 * MUI Button 호환(사이트 테마) — `.MuiButton-root.MuiButton-{variant}` 이름표. 운영자 CSS 가 `.post-btns .MuiButton-root` 등을 겨냥한다.
	 * href 가 있으면 링크로 그린다.
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		variant = 'text',
		size = 'medium',
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
		href?: string;
		disabled?: boolean;
		class?: string;
		sx?: Sx | null;
		onclick?: (e: MouseEvent) => void;
		startIcon?: Snippet;
		children?: Snippet;
	} = $props();

	const S = $derived(size[0].toUpperCase() + size.slice(1));
	const V = $derived(variant[0].toUpperCase() + variant.slice(1));
	const klass = $derived(cx(`MuiButtonBase-root MuiButton-root MuiButton-${variant} MuiButton-size${S} MuiButton-colorPrimary`, disabled && 'Mui-disabled', cls, toClass(sx)));
	void V;
</script>

{#if href}
	<a class={klass} {href} {onclick}>
		{#if startIcon}<span class="MuiButton-icon MuiButton-startIcon MuiButton-iconSize{S}">{@render startIcon()}</span>{/if}{@render children?.()}
	</a>
{:else}
	<button type="button" class={klass} {disabled} {onclick}>
		{#if startIcon}<span class="MuiButton-icon MuiButton-startIcon MuiButton-iconSize{S}">{@render startIcon()}</span>{/if}{@render children?.()}
	</button>
{/if}
