<script lang="ts">
	/** 메모 기록(React setting/components/MemoHistory) — 최근 것부터, 이름·시각 + 내용 */
	import { date } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	let { memos }: { memos?: { dt: any; memo: string; name?: string }[] | null } = $props();
	const items = $derived([...(memos || [])].map((o) => ({ ...o, t: new Date(o.dt).getTime() })).sort((a, b) => b.t - a.t));
</script>

{#each items as o (o.t + o.memo)}
	<div class={cx('MuiStack-root', toClass({ marginBottom: 8 }))}>
		<span class="MuiTypography-root MuiTypography-overline">{o.name ? `${o.name} ` : ''}{date(o.dt)}</span>
		<p class="MuiTypography-root MuiTypography-body2">{o.memo}</p>
	</div>
{/each}
