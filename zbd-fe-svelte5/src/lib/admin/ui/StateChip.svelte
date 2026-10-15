<script lang="ts">
	/**
	 * 상태 이름표(React components/common/StateChip) — MUI Chip 모양. isAutoColor 면 글자로 색을 고른다
	 * (진행중·승인·할당·연결됨·리뷰완료 = 초록 칠함 / 대기중·리뷰중 = 주황 / 그 밖 = 회색).
	 *   <StateChip label="진행중" isAutoColor size="small" />   <StateChip label="공개" color="primary" variant="filled" />
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	type C = 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning';
	let {
		label,
		color = 'primary',
		variant = 'outlined',
		size = 'medium',
		isAutoColor = false,
		onclick,
		sx = null
	}: { label: string; color?: C; variant?: 'filled' | 'outlined'; size?: 'small' | 'medium'; isAutoColor?: boolean; onclick?: () => void; sx?: Sx | null } = $props();

	const v = $derived.by(() => {
		if (!isAutoColor) return { c: color, v: variant };
		if (['진행중', '승인', '할당', '연결됨', '리뷰완료'].includes(label)) return { c: 'success' as C, v: 'filled' as const };
		if (['대기중', '리뷰중'].includes(label)) return { c: 'warning' as C, v: variant };
		return { c: 'default' as C, v: variant };
	});
	const cap = (s: string) => s[0].toUpperCase() + s.slice(1);
	const S = $derived(size == 'small' ? 'Small' : 'Medium');
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
	class={cx(`state-chip MuiChip-root MuiChip-${v.v} MuiChip-size${S} MuiChip-color${cap(v.c)} MuiChip-${v.v}${cap(v.c)}`, onclick && 'MuiChip-clickable', toClass(sx))}
	role={onclick ? 'button' : undefined}
	tabindex={onclick ? 0 : undefined}
	{onclick}
	onkeydown={(e) => onclick && e.key == 'Enter' && onclick()}
>
	<span class="MuiChip-label MuiChip-label{S}">{label}</span>
</div>
