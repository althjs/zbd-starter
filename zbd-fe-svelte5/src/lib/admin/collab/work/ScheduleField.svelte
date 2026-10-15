<script lang="ts">
	/**
	 * 일정 고르기(React DateRangeModal 자리) — 보이는 모양(template)을 누르면 공용 기간 고르기 창(ui/DateRangeModal, 여러 달 달력).
	 * React 업무 보드와 같은 값: 시작일·종료일 · 「날짜 초기화」(확인 후 null).
	 */
	import type { Snippet } from 'svelte';
	import DateRangeModal from '../../ui/DateRangeModal.svelte';
	import type { Schedule } from '#lib/shared/types/work.ts';

	let {
		schedule,
		onchange,
		readonly = false,
		useReset = true,
		fromLabel = '시작일',
		toLabel = '종료일',
		template
	}: { schedule: Schedule | null | undefined; onchange: (s: Schedule | null) => void; readonly?: boolean; useReset?: boolean; fromLabel?: string; toLabel?: string; template: Snippet } = $props();
</script>

<DateRangeModal
	schedule={schedule?.from ? schedule : null}
	{readonly}
	{useReset}
	{fromLabel}
	{toLabel}
	{template}
	onchange={(s) => onchange(s ? ({ from: s.from, to: s.to || s.from } as Schedule) : null)}
/>
