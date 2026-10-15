<script lang="ts">
	/** MUI Icon 호환 — 머티리얼 아이콘 글자(`<Icon>close</Icon>`) */
	import type { HTMLAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		name,
		size = 'Medium',
		sx = null,
		class: cls = '',
		...rest
	}: HTMLAttributes<HTMLSpanElement> & {
		name: string;
		size?: 'Small' | 'Medium' | 'Large' | 'Inherit';
		sx?: Sx | null;
		class?: string;
	} = $props();
</script>

<span class={cx(`material-icons notranslate MuiIcon-root MuiIcon-fontSize${size}`, cls, toClass(sx))} aria-hidden="true" {...rest}>{name}</span>
