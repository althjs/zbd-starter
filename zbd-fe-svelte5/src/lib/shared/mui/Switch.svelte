<script lang="ts">
	/** MUI Switch + FormControlLabel 호환(작은 크기) */
	let { checked = $bindable(false), label = '', class: cls = '', onchange }: { checked?: boolean; label?: string; class?: string; onchange?: (v: boolean) => void } = $props();
</script>

<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd {cls}">
	<span class="MuiSwitch-root MuiSwitch-sizeSmall">
		<span class="MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary{checked ? ' Mui-checked' : ''} PrivateSwitchBase-root">
			<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" bind:checked onchange={() => onchange?.(checked)} />
			<span class="MuiSwitch-thumb"></span>
		</span>
		<span class="MuiSwitch-track"></span>
	</span>
	<span class="MuiTypography-root MuiTypography-caption">{label}</span>
</label>
