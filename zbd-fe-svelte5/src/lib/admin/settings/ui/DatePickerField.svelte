<script lang="ts">
	/**
	 * 날짜 칸 — React 의 MUI X DatePicker(작은 outlined, 높이 32, yyyy-MM-dd) 모양. 누르면 브라우저 날짜 고르기를 연다.
	 *   <DatePickerField value="2026-10-03" onchange={(ymd) => …} width={150} />
	 * 값은 'YYYY-MM-DD'. (달력 창 대신 브라우저 것 — 관리 화면 쓰임새엔 충분)
	 */
	import '../settings.css';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let { value = '', onchange, width = 150, sx = null }: { value?: string; onchange?: (ymd: string) => void; width?: number | string; sx?: Sx | null } = $props();
	let native = $state<HTMLInputElement>();
	let focused = $state(false);
	const parts = $derived(String(value || '').split('-'));

	function open() {
		if (!native) return;
		try {
			native.showPicker();
		} catch {
			native.focus();
			native.click();
		}
	}
</script>

<div class={cx('MuiFormControl-root MuiPickersTextField-root adm-dp', focused && 'focused', toClass({ width, ...(sx || {}) }))}>
	<div class="MuiPickersInputBase-root MuiPickersOutlinedInput-root adm-dp-root" role="presentation" onclick={open}>
		<div class="MuiPickersSectionList-root adm-dp-sections" role="group">
			{#if parts.length == 3}<span>{parts[0]}</span><span>-</span><span>{parts[1]}</span><span>-</span><span>{parts[2]}</span>{:else}<span>{value}</span>{/if}
		</div>
		<div class="MuiInputAdornment-root MuiInputAdornment-positionEnd adm-dp-adorn">
			<button type="button" class="MuiButtonBase-root MuiIconButton-root MuiIconButton-edgeEnd adm-dp-btn" aria-label="날짜 고르기" onclick={(e) => (e.stopPropagation(), open())}>
				<svg focusable="false" aria-hidden="true" viewBox="0 0 24 24"
					><path d="M17 12h-5v5h5v-5zM16 1v2H8V1H6v2H5c-1.11 0-1.99.9-1.99 2L3 19c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2h-1V1h-2zm3 18H5V8h14v11z"></path></svg
				>
			</button>
		</div>
		<fieldset aria-hidden="true"><legend><span class="notranslate">&#8203;</span></legend></fieldset>
		<input
			bind:this={native}
			class="adm-dp-native"
			type="date"
			tabindex="-1"
			{value}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			onchange={(e) => {
				const v = (e.currentTarget as HTMLInputElement).value;
				if (v) onchange?.(v);
			}}
		/>
	</div>
</div>
