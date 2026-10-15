<script lang="ts">
	/**
	 * 목록 아래 줄(React components/common/table/TablePagenation) — 왼쪽 쪽당 개수 · 가운데 쪽 번호 · 오른쪽 자리(children).
	 *   <TablePagination search={search} pageCount={Math.ceil(total / search.pageSize)} getData={(s) => load(s)} />
	 * search = { pageNo, pageSize, … } — 바뀐 search 를 getData 로 넘긴다(쪽당 개수는 localStorage zbd:admin:pageSize 에 기억).
	 * 목록 패턴: 이 줄은 화면(+page) 이 들고, 표 부품에는 이미 자른 rows 와 startNo 만 넘긴다.
	 */
	import type { Snippet } from 'svelte';
	import Pagination from '#lib/shared/mui/Pagination.svelte';
	import Select from '#lib/shared/mui/Select.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { LS } from '../config.ts';

	let {
		search,
		getData,
		pageCount = 0,
		sizeList = [5, 10, 20, 50, 100, 500, 1000],
		sx = null,
		children
	}: { search: { pageNo: number; pageSize: number; [k: string]: any }; getData: (s: any) => void; pageCount?: number; sizeList?: number[]; sx?: Sx | null; children?: Snippet } = $props();
</script>

<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', ...(sx || {}) }))}>
	<Select
		value={search.pageSize}
		options={sizeList.map((n) => ({ value: n, label: String(n) }))}
		sx={{ alignSelf: 'start' }}
		onchange={(v) => {
			try {
				localStorage.setItem(LS.pageSize, String(v));
			} catch {}
			getData({ ...search, pageNo: 1, pageSize: v });
		}}
	/>
	<Pagination page={search.pageNo} count={pageCount} siblingCount={3} onchange={(p) => getData({ ...search, pageNo: p })} />
	<div class={cx('MuiStack-root', toClass({ width: children ? 'unset' : 100 }))}>{@render children?.()}</div>
</div>
