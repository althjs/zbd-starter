<script lang="ts">
	/**
	 * 관리자 아이콘 — 머티리얼 심볼(Outlined). React 관리자의 `<Icon>`(MUI baseClassName material-symbols-outlined) 과 같은 DOM.
	 *   <Icon name="settings" />  ·  <Icon name="mail" size="Small" color="Primary" sx={{ fontSize: 18 }} />
	 * 사용자 화면 아이콘(shared/mui/Icon = Material Icons 채움) 과 모양이 달라 관리자는 이걸 쓴다.
	 */
	import type { HTMLAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		name,
		size = 'Medium',
		color = '',
		sx = null,
		class: cls = '',
		...rest
	}: HTMLAttributes<HTMLSpanElement> & {
		name: string;
		size?: 'Small' | 'Medium' | 'Large' | 'Inherit';
		color?: '' | 'Primary' | 'Action' | 'Error' | 'Inherit';
		sx?: Sx | null;
		class?: string;
	} = $props();
</script>

<span class={cx(`material-symbols-outlined notranslate MuiIcon-root`, color && `MuiIcon-color${color}`, `MuiIcon-fontSize${size}`, cls, toClass(sx))} aria-hidden="true" {...rest}>{name}</span>
