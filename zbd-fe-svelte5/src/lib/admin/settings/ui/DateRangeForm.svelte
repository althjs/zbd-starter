<script lang="ts">
	/**
	 * 「작성일」 기간 줄 — React pages/admin/board/components/BoardSearchForm 의 DateRangeForm 모양 그대로
	 * (날짜 칸 2개 + 전체·오늘·이번주·이번달·지난 7일·지난 30일. 지금 기간과 같은 버튼은 outlined).
	 *   <DateRangeForm {search} setSearch={(s) => load(s)} />   search.from · search.to · pageNo 를 고친 새 객체를 준다
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import { getThisMonthYMD, getThisWeekYMD, getYesterdayYMD, getYMD } from '../ymd.ts';
	import DatePickerField from './DatePickerField.svelte';

	let { search, setSearch, label = '작성일' }: { search: { from: string; to: string; [k: string]: any }; setSearch: (s: any) => void; label?: string } = $props();

	const today = getYMD();
	const ranges = [
		{ label: '전체', r: { from: '2024-01-01', to: today } },
		{ label: '오늘', r: { from: today, to: today } },
		{ label: '이번주', r: getThisWeekYMD() },
		{ label: '이번달', r: getThisMonthYMD() },
		{ label: '지난 7일', r: { from: getYesterdayYMD(null, 7), to: today } },
		{ label: '지난 30일', r: { from: getYesterdayYMD(null, 30), to: today } }
	];
	const set = (r: { from?: string; to?: string }) => setSearch({ ...search, ...r, pageNo: 1 });
</script>

<div class="MuiStack-root">
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
		<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ textAlign: 'center', minWidth: 0, '@media (min-width:600px)': { minWidth: 100 } }))}>{label}</span>
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
			<DatePickerField value={getYMD(search.from)} onchange={(v) => set({ from: v })} />
			~
			<DatePickerField value={getYMD(search.to)} onchange={(v) => set({ to: v })} />
		</div>
		{#each ranges as b (b.label)}
			<Button size="small" variant={search.from === b.r.from && search.to === b.r.to ? 'outlined' : 'text'} onclick={() => set(b.r)}>{b.label}</Button>
		{/each}
	</div>
</div>
