<script lang="ts">
	/**
	 * 보통 크기 스위치 + 글자(MUI FormControlLabel + Switch medium) — 공용 shared/mui/Switch 는 작은 크기·caption 글자라 관리자 폼은 이걸.
	 *   <SwitchField checked={v} label="상단고정" variant="body2" onchange={(x) => …} />
	 */
	let {
		checked = $bindable(false),
		label = '',
		variant = 'body1',
		disabled = false,
		onchange
	}: { checked?: boolean; label?: string; variant?: 'body1' | 'body2'; disabled?: boolean; onchange?: (v: boolean) => void } = $props();
</script>

<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd" class:Mui-disabled={disabled}>
	<span class="MuiSwitch-root MuiSwitch-sizeMedium">
		<span class="MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root" class:Mui-checked={checked} class:Mui-disabled={disabled}>
			<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" bind:checked {disabled} onchange={() => onchange?.(checked)} />
			<span class="MuiSwitch-thumb"></span>
		</span>
		<span class="MuiSwitch-track"></span>
	</span>
	{#if variant == 'body1'}<span class="MuiTypography-root MuiTypography-body1 MuiFormControlLabel-label">{label}</span>{:else}<span class="MuiFormControlLabel-label"
			><p class="MuiTypography-root MuiTypography-body2">{label}</p></span
		>{/if}
</label>
