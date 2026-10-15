<script lang="ts">
	/**
	 * MUI Stack 호환 — `.MuiStack-root`(flex 세로) 를 붙인 상자.
	 * 운영자 CSS 가 `.MuiStack-root` 를 겨냥하므로 모양 부품을 새로 만들 때도 이걸 쓴다(styles/mui.css).
	 * direction·sx 는 클래스 규칙으로(sx.ts) — 인라인 style 은 운영자 CSS 가 덮지 못한다.
	 */
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		tag = 'div',
		direction = '',
		sx = null,
		class: cls = '',
		el = $bindable(),
		children,
		...rest
	}: HTMLAttributes<HTMLElement> & {
		tag?: string;
		direction?: string;
		sx?: Sx | null;
		class?: string;
		el?: HTMLElement;
		children?: Snippet;
	} = $props();

	const klass = $derived(cx('MuiStack-root', cls, toClass(direction ? { flexDirection: direction, ...(sx || {}) } : sx)));
</script>

<svelte:element this={tag} bind:this={el} class={klass} {...rest}>{@render children?.()}</svelte:element>
