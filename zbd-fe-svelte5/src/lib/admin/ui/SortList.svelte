<script lang="ts">
	/**
	 * 끌어서 순서 바꾸기 목록 — 왼쪽 손잡이를 끌거나 ▲▼·⤒ 버튼. 바뀐 배열을 onchange 로 준다(원본은 건드리지 않는다).
	 *   <SortList items={rows} label={(o) => o.name} onchange={(next) => (rows = next)} />
	 * 끌기는 공용 액션 lib/shared/ui/sortable.ts(줄이 손가락을 따라 움직이고 다른 줄이 비켜 준다, 2026-10-04).
	 */
	import type { Snippet } from 'svelte';
	import { sortable } from '#lib/shared/ui/sortable.ts';
	import Icon from './Icon.svelte';
	import IconButton from './IconButton.svelte';

	let { items, label, onchange, row, buttons = true }: { items: any[]; label?: (o: any) => string; onchange: (next: any[]) => void; row?: Snippet<[any, number]>; buttons?: boolean } = $props();

	function move(from: number, to: number) {
		if (from == to || to < 0 || to >= items.length) return;
		const next = [...items];
		const [x] = next.splice(from, 1);
		next.splice(to, 0, x);
		onchange(next);
	}
</script>

<div class="adm-sort" use:sortable={{ handle: '.adm-sort-handle', onmove: move }}>
	{#each items as o, i (o)}
		<div class="adm-sort-row">
			<span class="adm-sort-handle" role="button" tabindex="-1" aria-label="끌어서 옮기기"><Icon name="drag_indicator" /></span>
			<div class="adm-sort-body">
				{#if row}{@render row(o, i)}{:else}{label?.(o) ?? String(o)}{/if}
			</div>
			{#if buttons}
				<IconButton size="Small" disabled={i == 0} onclick={() => move(i, 0)} aria-label="맨 위로"><Icon name="vertical_align_top" size="Small" /></IconButton>
				<IconButton size="Small" disabled={i == 0} onclick={() => move(i, i - 1)} aria-label="위로"><Icon name="arrow_upward" size="Small" /></IconButton>
				<IconButton size="Small" disabled={i == items.length - 1} onclick={() => move(i, i + 1)} aria-label="아래로"><Icon name="arrow_downward" size="Small" /></IconButton>
			{/if}
		</div>
	{/each}
</div>

<style>
	.adm-sort {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.adm-sort-row {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		border: 1px solid var(--gt-300, #ebebeb);
		border-radius: 8px;
		background: var(--adm-paper);
		font-size: 14px;
	}
	.adm-sort-handle {
		cursor: grab;
		color: var(--gt-700, #909090);
		display: inline-flex;
		touch-action: none;
	}
	.adm-sort-body {
		flex: 1;
		min-width: 0;
	}
</style>
