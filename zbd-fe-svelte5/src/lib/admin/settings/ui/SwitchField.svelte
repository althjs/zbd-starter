<script lang="ts">
	/**
	 * 스위치 + 이름표(MUI FormControlLabel + Switch small) — 공용 shared/mui/Switch 는 이름표가 caption(12) 고정이라,
	 * React 에서 label 을 글자 그대로 준 곳(body1 → 관리자 14px)은 이걸 쓴다.
	 *   <SwitchField checked={v} label="자동 새로고침" onchange={(x) => …} />   variant="caption" 이면 공용과 같다
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let {
		checked = false,
		label = '',
		variant = 'body1',
		disabled = false,
		labelSx = null,
		onchange
	}: { checked?: boolean; label?: string; variant?: 'body1' | 'caption' | 'body2'; disabled?: boolean; labelSx?: Sx | null; onchange?: (v: boolean) => void } = $props();
</script>

<label class={cx('MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd', disabled && 'Mui-disabled')}>
	<span class="MuiSwitch-root MuiSwitch-sizeSmall">
		<span class={cx('MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root', checked && 'Mui-checked', disabled && 'Mui-disabled')}>
			<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" {checked} {disabled} onchange={(e) => onchange?.((e.currentTarget as HTMLInputElement).checked)} />
			<span class="MuiSwitch-thumb"></span>
		</span>
		<span class="MuiSwitch-track"></span>
	</span>
	<span
		class={cx(`MuiTypography-root MuiTypography-${variant} MuiFormControlLabel-label`, disabled && 'Mui-disabled', labelSx && toClass({ '&.MuiTypography-root.MuiFormControlLabel-label': labelSx }))}
		>{label}</span
	>
</label>
