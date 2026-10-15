<script lang="ts">
	/**
	 * 저장 전 바뀐 값 미리보기(React DiffPreview 자리) — 경로 · 전 → 후. 빨강 = 지운 값, 초록 = 새 값.
	 *   <DiffPreview items={diffList(orig, cur)} />   (관리자·사용자 화면 공용 — 테두리는 그 화면의 구분선 색)
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import type { DiffItem } from '#lib/shared/diff.ts';
	let { items }: { items: DiffItem[] } = $props();
	const show = (v: any) => (v === undefined ? '' : typeof v == 'string' ? v : JSON.stringify(v));
</script>

{#if items.length}
	<div
		class={cx(
			'MuiStack-root',
			toClass({
				gap: 4,
				fontFamily: 'monospace',
				fontSize: 12,
				padding: 12,
				border: '1px solid var(--adm-divider, var(--mui-divider))',
				borderRadius: 8,
				overflow: 'auto',
				backgroundColor: 'var(--gt-100, #f8f8f8)'
			})
		)}
	>
		{#each items as d (d.path)}
			<div class={toClass({ wordBreak: 'break-all' })}>
				<b>{d.path}</b>:
				{#if d.kind != 'add'}<span class={toClass({ backgroundColor: '#ffbbbb', textDecoration: 'line-through', padding: '0 2px' })}>{show(d.before)}</span>{/if}
				{#if d.kind != 'remove'}<span class={toClass({ backgroundColor: '#bbffbb', padding: '0 2px' })}>{show(d.after)}</span>{/if}
			</div>
		{/each}
	</div>
{/if}
