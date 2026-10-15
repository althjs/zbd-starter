<script lang="ts">
	/**
	 * 업무 보드 표(React WorkListTable) — 기본 열(ID·카테고리·제목·상태) + 보드가 정한 열 + 요약·설명.
	 * 보이기·고정·순서는 톱니(목록 보기 옵션) — 고정 열은 왼쪽에 붙고 마지막 고정 열 뒤에 굵은 선(휴대폰은 고정 없음).
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { admin } from '../../profile.svelte.ts';
	import { getWorkCodeIds, getWorkCodes, getWorkColumns, SHAPE_LABELS, clone } from './codes.ts';
	import type { WorkItem, NotiMember } from '#lib/shared/types/work.ts';
	import ColumnOptionButton from './ColumnOptionButton.svelte';
	import { getColumnOptions, getColumnOptionsDefault, type ColumnOption } from './columnOptions.ts';
	import Scrollbar from '#lib/shared/ui/Scrollbar.svelte';
	import { media } from '../../media.svelte.ts';
	import JsonDrawer from './JsonDrawer.svelte';
	import RenderCell, { type Col } from './RenderCell.svelte';
	import { updateItem } from './store.svelte.ts';

	let { items = [], groupKey, admins = [] }: { items?: WorkItem[]; groupKey: string; admins?: NotiMember[] } = $props();
	const codes = $derived(getWorkCodes(groupKey));
	const codeIds = $derived(getWorkCodeIds(groupKey));
	const optKey = $derived(`work_${groupKey}`);
	let options = $state<ColumnOption[]>([]);
	$effect(() => {
		options = getColumnOptions(optKey);
	});
	let jsonItem = $state<WorkItem | null>(null);

	const dyn = $derived(
		getWorkColumns(groupKey).map(
			(c) =>
				({
					key: c.key || `column_${c.id}`,
					name: c.name,
					width: Number(c.value) || 100,
					id: c.id,
					items: c.items,
					inputType: c.inputType,
					workFields: c.workFields,
					component: c.component,
					sticky: false
				}) as Col
		)
	);
	const configs = $derived<Col[]>([
		{ key: 'id', name: 'ID', width: 70, align: 'center', sticky: true },
		{ key: 'category', name: '카테고리', width: 300, sticky: true },
		{ key: 'title', name: '제목', width: 250, sticky: false },
		{ key: 'status', name: '상태', width: 100, align: 'center', sticky: false },
		...dyn,
		...(dyn.some((c) => c.key === 'summary' || c.name === '요약') ? [] : [{ key: 'summary', name: '요약', width: 200, sticky: false }]),
		...(dyn.some((c) => c.key === 'desc' || c.name === '설명') ? [] : [{ key: 'desc', name: '설명', width: 400, sticky: false }])
	]);
	const optOf = (k: string) => options.find((x) => x.key === k);
	const cols = $derived.by(() => {
		const rank = (k: string) => {
			const i = options.findIndex((x) => x.key === k);
			return i === -1 ? Number.MAX_SAFE_INTEGER : i;
		};
		const vis = configs.filter((c) => optOf(c.key)?.use !== false).sort((a, b) => rank(a.key) - rank(b.key));
		let left = 0;
		return vis.map((c, i) => {
			// 휴대폰(< 900)은 고정 열을 쓰지 않는다 — 좁은 화면에서 고정 열이 내용을 가렸다(2026-10-05 대표)
			const st = (x: Col) => !media.mdDown && (optOf(x.key)?.sticky !== undefined ? !!optOf(x.key)!.sticky : !!x.sticky);
			const sticky = st(c);
			const next = vis[i + 1];
			const out = { ...c, sticky, left: sticky ? left : undefined, borderRight: sticky && !(next && st(next)) ? '2px solid #ddd' : undefined };
			if (sticky) left += c.width;
			return out;
		});
	});
	const defaultOptions = $derived.by(() => {
		const cur = configs.map((c) => ({ name: c.name, key: c.key, sticky: c.sticky, use: true }));
		const saved = getColumnOptionsDefault(optKey);
		return saved ? cur.map((c) => ({ ...c, ...(saved.find((m) => m.key === c.key) || {}) })) : cur;
	});
	const hideCat = $derived(!optOf('category')?.use);
	async function onValueChange(item: WorkItem, key: string, v: any) {
		const n = clone(item);
		n.Json[key] = v;
		await updateItem(n);
	}
	const summaryOf = (shapes: any[]) => {
		const m: Record<string, number> = {};
		shapes.forEach((s) => (m[s.type] = (m[s.type] || 0) + 1));
		return Object.entries(m)
			.map(([t, n]) => `${SHAPE_LABELS[t] || t} ${n}`)
			.join(', ');
	};
	const cellSx = (c: any, head: boolean) => ({
		position: c.sticky ? 'sticky' : 'static',
		left: c.left,
		zIndex: head ? (c.sticky ? 2 : 1) : c.sticky ? 1 : 0,
		backgroundColor: head ? 'var(--zc-f5f5f5, #f5f5f5) !important' : 'var(--zc-fff, #fff) !important',
		borderRight: c.borderRight,
		...(head ? {} : { whiteSpace: c.key === 'summary' || c.key.startsWith('column_') || c.key === 'status' ? 'nowrap' : 'normal' })
	});
	const al = (c: Col) => (c.align === 'center' ? 'MuiTableCell-alignCenter' : c.align === 'right' ? 'MuiTableCell-alignRight' : '');
</script>

<!-- 톱니(목록 보기 옵션) — PC 는 표 위 줄, 휴대폰은 따로 한 줄을 차지하지 않게 머리 줄 첫 칸 안(2026-10-05 대표) -->
{#if !media.mdDown}<ColumnOptionButton optionKey={optKey} {options} {defaultOptions} onchange={(v) => (options = v)} />{/if}
<Scrollbar class="adm-cv-scroll" style="width:100%">
	<table
		class={cx(
			'MuiTable-root adm-cv-table lt-table',
			toClass({
				width: 'max-content',
				minWidth: '100%',
				borderCollapse: 'separate',
				'& th': {
					paddingTop: 12,
					paddingBottom: 12,
					paddingLeft: 8,
					paddingRight: 8,
					background: 'var(--zc-f5f5f5, #f5f5f5) !important',
					fontSize: 12,
					fontWeight: 'bold !important',
					borderBottom: '1px solid #ddd',
					whiteSpace: 'nowrap'
				},
				'& td': { paddingLeft: 8, paddingRight: 8, paddingTop: 0, paddingBottom: 0, fontSize: 12, whiteSpace: 'nowrap' }
			})
		)}
	>
		<thead class="MuiTableHead-root">
			<tr class="MuiTableRow-root MuiTableRow-head">
				{#each cols as c, i (c.key)}<th
						class={cx('MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium', al(c), toClass(cellSx(c, true)))}
						style:width={`${c.width}px`}
						data-k={c.key}
						scope="col"
						>{#if i === 0 && media.mdDown}<span class="lt-opt"><ColumnOptionButton optionKey={optKey} {options} {defaultOptions} onchange={(v) => (options = v)} /></span>{/if}{c.name}</th
					>{/each}
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each items as o (o.workId)}
				{@const shapes = o.Json.shapes || []}
				{@const sum = summaryOf(shapes)}
				<tr class="MuiTableRow-root">
					{#each cols as c (c.key)}
						<td class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium', al(c), toClass(cellSx(c, false)))} data-k={c.key}>
							<RenderCell col={c} item={o} {shapes} summaryText={sum} {groupKey} {codes} {codeIds} {admins} {onValueChange} isHiddenCategoryColumn={hideCat} onjson={(x) => (jsonItem = x)} />
						</td>
					{/each}
				</tr>
			{:else}
				<tr class="MuiTableRow-root">
					<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium MuiTableCell-alignCenter" colspan={cols.length}
						><p class="MuiTypography-root MuiTypography-body2" style="margin:40px 0">등록된 컨텐츠가 없습니다.</p></td
					>
				</tr>
			{/each}
		</tbody>
	</table>
</Scrollbar>
{#if admin.profile?.isDeveloper}<JsonDrawer value={jsonItem} onclose={() => (jsonItem = null)} onsave={(v) => updateItem(v)} />{/if}

<style>
	.lt-opt {
		display: inline-flex;
		vertical-align: middle;
		margin-right: 6px;
	}
	/* 휴대폰(2026-10-05 대표) — 제목 칸을 화면 폭에 맞춰 ID 와 함께 한눈에(두 줄까지 줄바꿈), 나머지 칸은 옆으로 밀어 본다 */
	@media (max-width: 899px) {
		.lt-table :global([data-k='title']) {
			width: calc(100vw - 150px) !important;
			min-width: 180px;
			max-width: 320px;
		}
		.lt-table :global(td[data-k='title']) {
			white-space: normal !important;
			padding-top: 6px !important;
			padding-bottom: 6px !important;
		}
		.lt-table :global(td[data-k='title'] a) {
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
			overflow: hidden;
			word-break: keep-all;
			overflow-wrap: anywhere;
		}
	}
</style>
