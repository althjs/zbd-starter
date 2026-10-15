<script lang="ts">
	/**
	 * 바탕 ui/SvgIcon 에 없는 @mui/icons-material 경로를 더한 것 — 이 묶음 화면이 쓰는 것만(경로는 그 패키지 값 그대로).
	 * 없는 이름은 바탕 SvgIcon 으로 넘긴다. (바탕 PATHS 에 합치면 이 파일은 지워도 된다 — 공용 후보)
	 */
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Base from '../../ui/SvgIcon.svelte';

	const PATHS: Record<string, string> = {
		DragIndicator:
			'M11 18c0 1.1-.9 2-2 2s-2-.9-2-2 .9-2 2-2 2 .9 2 2m-2-8c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m0-6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m6 4c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2m0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2',
		Clear: 'M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z',
		Add: 'M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6z',
		Refresh:
			'M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4z',
		ContentCopy: 'M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2m0 16H8V7h11z',
		Delete: 'M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6zM19 4h-3.5l-1-1h-5l-1 1H5v2h14z',
		Edit: 'M3 17.25V21h3.75L17.81 9.94l-3.75-3.75zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34a.996.996 0 0 0-1.41 0l-1.83 1.83 3.75 3.75z',
		ExpandMore: 'M16.59 8.59 12 13.17 7.41 8.59 6 10l6 6 6-6z',
		ChevronLeft: 'M15.41 7.41 14 6l-6 6 6 6 1.41-1.41L10.83 12z',
		ChevronRight: 'M10 6 8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z'
	};
	let { name, size = 'Medium', color = '', class: cls = '', sx = null }: { name: string; size?: 'Small' | 'Medium' | 'Large' | 'Inherit'; color?: string; class?: string; sx?: Sx | null } = $props();
</script>

{#if PATHS[name]}
	<svg class={cx(`MuiSvgIcon-root MuiSvgIcon-fontSize${size}`, color && `MuiSvgIcon-color${color}`, cls, toClass(sx))} focusable="false" aria-hidden="true" viewBox="0 0 24 24"
		><path d={PATHS[name]}></path></svg
	>
{:else}
	<Base {name} {size} {color} class={cls} {sx} />
{/if}
