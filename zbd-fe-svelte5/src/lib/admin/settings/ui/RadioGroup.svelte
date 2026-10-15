<script lang="ts">
	/**
	 * 라디오 묶음(MUI RadioGroup row + FormControlLabel + Radio) — 값은 문자열.
	 *   <RadioGroup value={v} options={[{ value: 'RW', label: '읽기/쓰기' }, …]} onchange={(v) => …} />
	 *   size="small" 이면 작은 라디오, disabled(전체) · option.disabled(하나)
	 */
	import '../settings.css';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value,
		options,
		onchange,
		row = true,
		size = 'medium',
		disabled = false,
		sx = null
	}: { value: any; options: { value: any; label: string; disabled?: boolean }[]; onchange: (v: any) => void; row?: boolean; size?: 'small' | 'medium'; disabled?: boolean; sx?: Sx | null } = $props();
	const name = `adm-rg-${Math.random().toString(36).slice(2, 8)}`;
</script>

<div role="radiogroup" class={cx('MuiFormGroup-root', row && 'MuiFormGroup-row', toClass({ display: 'flex', flexDirection: row ? 'row' : 'column', flexWrap: 'wrap', ...(sx || {}) }))}>
	{#each options as o (o.value)}
		{@const on = String(o.value ?? '') === String(value ?? '')}
		{@const dis = disabled || !!o.disabled}
		<label class={cx('MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd', dis && 'Mui-disabled')}>
			<span
				class={cx('MuiButtonBase-root MuiRadio-root MuiRadio-colorPrimary PrivateSwitchBase-root', size == 'small' && 'MuiRadio-sizeSmall adm-radio-sm', on && 'Mui-checked', dis && 'Mui-disabled')}
				style="position:relative;display:inline-flex"
			>
				<input class="PrivateSwitchBase-input" type="radio" {name} checked={on} disabled={dis} onchange={() => onchange(o.value)} />
				<span class="adm-radio-icon">
					<svg class="MuiSvgIcon-root" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
						><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z"></path></svg
					>
					<svg class="MuiSvgIcon-root adm-radio-dot" class:on focusable="false" aria-hidden="true" viewBox="0 0 24 24"
						><path d="M8.465 8.465C9.37 7.56 10.62 7 12 7C14.76 7 17 9.24 17 12C17 13.38 16.44 14.63 15.535 15.535C14.63 16.44 13.38 17 12 17C9.24 17 7 14.76 7 12C7 10.62 7.56 9.37 8.465 8.465Z"
						></path></svg
					>
				</span>
			</span>
			<span class={cx('MuiTypography-root MuiTypography-body1 MuiFormControlLabel-label', dis && 'Mui-disabled')}>{o.label}</span>
		</label>
	{/each}
</div>
