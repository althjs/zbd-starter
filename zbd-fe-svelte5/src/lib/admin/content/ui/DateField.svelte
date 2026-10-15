<script lang="ts">
	/**
	 * 날짜 칸 — MUI x DatePicker(outlined · small · 폭 150 · 높이 32) 모양. 달력 단추를 누르면 브라우저 날짜 고르기(showPicker)가 뜬다.
	 *   <DateField value="2026-10-03" onchange={(ymd) => …} />
	 * 값은 'YYYY-MM-DD'. 칸에 직접 쓰기도 된다(브라우저 date 입력이 칸 위에 투명하게 겹쳐 있다).
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value,
		onchange,
		width = 150,
		height = 32,
		label = '',
		sx = null
	}: { value: string; onchange: (ymd: string) => void; width?: number | string; height?: number; label?: string; sx?: Sx | null } = $props();
	let input = $state<HTMLInputElement>();
	const parts = $derived(String(value || '').split('-'));
	function pick() {
		try {
			input?.showPicker();
		} catch {
			input?.focus();
		}
	}
	const root = $derived(toClass({ display: 'inline-flex', flexDirection: 'column', position: 'relative', minWidth: 0, width, ...(sx || {}) }));
	const base = $derived(
		toClass({
			display: 'flex',
			alignItems: 'center',
			position: 'relative',
			height,
			padding: '0 14px',
			borderRadius: 8,
			fontFamily: 'var(--mui-font)',
			fontSize: 14,
			fontWeight: 500,
			lineHeight: '24px',
			letterSpacing: '0.13132px',
			color: 'var(--adm-text)',
			cursor: 'text',
			'&:hover fieldset': { borderColor: 'var(--adm-text)' }
		})
	);
	const sec = toClass({ display: 'flex', flex: '1 1 auto', overflow: 'hidden', lineHeight: '20.125px', whiteSpace: 'nowrap' });
	const fs = toClass({
		position: 'absolute',
		top: -5,
		left: 0,
		right: 0,
		bottom: 0,
		margin: 0,
		padding: '0 8px',
		border: '1px solid var(--adm-n200)',
		borderRadius: 8,
		pointerEvents: 'none',
		minWidth: '0%',
		overflow: 'hidden',
		textAlign: 'left'
	});
	const adorn = toClass({ position: 'relative', zIndex: 1, display: 'flex', alignItems: 'center', marginLeft: 8, height: '0.01em', maxHeight: '2em', color: 'var(--adm-text2)' });
	const ib = toClass({
		display: 'inline-flex',
		alignItems: 'center',
		justifyContent: 'center',
		padding: 8,
		marginRight: -12,
		borderRadius: '50%',
		color: 'inherit',
		border: 0,
		background: 'transparent',
		cursor: 'pointer',
		'&:hover': { backgroundColor: 'var(--adm-hover)' }
	});
	const native = toClass({
		position: 'absolute',
		inset: 0,
		opacity: 0,
		width: '100%',
		height: '100%',
		border: 0,
		padding: 0,
		cursor: 'text',
		'&::-webkit-calendar-picker-indicator': { display: 'none' }
	});
</script>

<div class={cx('MuiFormControl-root MuiPickersTextField-root', root)}>
	{#if label}<label class="MuiFormLabel-root MuiInputLabel-root MuiInputLabel-shrink">{label}</label>{/if}
	<div class={cx('MuiPickersInputBase-root MuiPickersOutlinedInput-root MuiPickersInputBase-adornedEnd', base)}>
		<div class={cx('MuiPickersSectionList-root MuiPickersInputBase-sectionsContainer', sec)}>
			<span class="MuiPickersSectionList-section"><span class="MuiPickersSectionList-sectionContent">{parts[0] || 'YYYY'}</span><span class="MuiPickersInputBase-sectionAfter">-</span></span>
			<span class="MuiPickersSectionList-section"><span class="MuiPickersSectionList-sectionContent">{parts[1] || 'MM'}</span><span class="MuiPickersInputBase-sectionAfter">-</span></span>
			<span class="MuiPickersSectionList-section"><span class="MuiPickersSectionList-sectionContent">{parts[2] || 'DD'}</span></span>
		</div>
		<input
			bind:this={input}
			class={native}
			type="date"
			{value}
			aria-label={label || '날짜'}
			onchange={(e) => (e.currentTarget as HTMLInputElement).value && onchange((e.currentTarget as HTMLInputElement).value)}
		/>
		<div class={cx('MuiInputAdornment-root MuiInputAdornment-positionEnd MuiInputAdornment-outlined', adorn)}>
			<button type="button" class={cx('MuiButtonBase-root MuiIconButton-root MuiIconButton-edgeEnd MuiIconButton-sizeMedium', ib)} aria-label="날짜 고르기" onclick={pick}>
				<svg
					class={cx('MuiSvgIcon-root MuiSvgIcon-fontSizeMedium', toClass({ fontSize: 18, width: '1em', height: '1em', fill: 'currentColor' }))}
					focusable="false"
					aria-hidden="true"
					viewBox="0 0 24 24"><path d="M17 12h-5v5h5v-5zM16 1v2H8V1H6v2H5c-1.11 0-1.99.9-1.99 2L3 19c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2h-1V1h-2zm3 18H5V8h14v11z"></path></svg
				>
			</button>
		</div>
		<fieldset aria-hidden="true" class={cx('MuiPickersOutlinedInput-notchedOutline', fs)}>
			<legend style="float:unset;width:auto;overflow:hidden;padding:0;line-height:11px;height:11px;visibility:hidden;max-width:0.01px"><span class="notranslate">&#8203;</span></legend>
		</fieldset>
	</div>
</div>
