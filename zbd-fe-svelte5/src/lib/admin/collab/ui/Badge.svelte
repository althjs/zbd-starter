<script lang="ts">
	/** 숫자 배지(MUI Badge) — 0 이면 숨김(MUI 기본). children 이 없으면 배지만 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { content = 0, color = 'error', max = 99, sx = null, children }: { content?: number; color?: 'error' | 'primary'; max?: number; sx?: Sx | null; children?: Snippet } = $props();
</script>

<span class={cx('MuiBadge-root', toClass(sx))}
	>{@render children?.()}<span
		class={cx('MuiBadge-badge MuiBadge-standard MuiBadge-anchorOriginTopRight', color == 'error' ? 'MuiBadge-colorError' : 'MuiBadge-colorPrimary', !content && 'MuiBadge-invisible')}
		>{content > max ? `${max}+` : content || ''}</span
	></span
>
