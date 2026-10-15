<script lang="ts">
	/** MUI Checkbox(small) + 이름표 호환 */
	import SvgIcon from './SvgIcon.svelte';
	let { checked = $bindable(false), label = '', onchange }: { checked?: boolean; label?: string; onchange?: (v: boolean) => void } = $props();
</script>

<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
	<span class="MuiButtonBase-root MuiCheckbox-root MuiCheckbox-colorPrimary MuiCheckbox-sizeSmall PrivateSwitchBase-root{checked ? ' Mui-checked' : ''}" style="position:relative">
		<input class="PrivateSwitchBase-input" type="checkbox" bind:checked onchange={() => onchange?.(checked)} />
		<SvgIcon name={checked ? 'CheckBox' : 'CheckBoxOutlineBlank'} size="Small" />
	</span>
	<span class="MuiTypography-root MuiTypography-caption" style="min-width:fit-content">{label}</span>
</label>
