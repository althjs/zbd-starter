<script lang="ts">
	/**
	 * 목록 아래 줄 — [쪽 크기 선택] [쪽 번호] [오른쪽 자리]. 예전 components/common/table/TablePagenation.tsx (관리자·사용자 공용)
	 * 쪽 크기가 5·10 의 배수가 아니면 선택 상자 대신 빈 자리.
	 */
	import type { Snippet } from 'svelte';
	import Pagination from './Pagination.svelte';
	import Select from './Select.svelte';
	import Stack from './Stack.svelte';
	import type { Sx } from './sx';

	let {
		pageNo,
		pageSize,
		pageCount,
		onchange,
		hiddenSelectBox = false,
		sx = null,
		children
	}: { pageNo: number; pageSize: number; pageCount: number; onchange: (s: { pageNo: number; pageSize: number }) => void; hiddenSelectBox?: boolean; sx?: Sx | null; children?: Snippet } = $props();

	const sizes = [5, 10, 20, 50, 100, 500, 1000];
</script>

<Stack class="paging" direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between', ...(sx || {}) }}>
	{#if pageSize % 10 != 0 && pageSize % 5 != 0}
		<Stack />
	{:else}
		<Select
			class="page-size-select"
			value={pageSize || 10}
			options={sizes.map((v) => ({ value: v, label: String(v) }))}
			sx={{ width: 90, alignSelf: 'start', visibility: hiddenSelectBox ? 'hidden' : 'visible' }}
			onchange={(v) => onchange({ pageNo: 1, pageSize: v })}
		/>
	{/if}
	<Pagination page={pageNo} count={pageCount} siblingCount={3} onchange={(p) => onchange({ pageNo: p, pageSize })} />
	<Stack sx={{ width: children ? 'unset' : 90 }}>{@render children?.()}</Stack>
</Stack>
