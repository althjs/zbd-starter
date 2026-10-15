<script lang="ts">
	/** 체크박스(MUI Checkbox small) — 이름표 없이. 이름표가 있으면 label */
	import SvgIcon from '../../ui/SvgIcon.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let {
		checked = false,
		disabled = false,
		label = '',
		labelSx = null,
		sx = null,
		ariaLabel = undefined,
		onchange,
		onclick
	}: { checked?: boolean; disabled?: boolean; label?: string; labelSx?: Sx | null; sx?: Sx | null; ariaLabel?: string; onchange?: (v: boolean) => void; onclick?: (e: MouseEvent) => void } = $props();
</script>

{#snippet box()}
	<span
		class={cx('MuiButtonBase-root MuiCheckbox-root MuiCheckbox-colorPrimary MuiCheckbox-sizeSmall PrivateSwitchBase-root', checked && 'Mui-checked', disabled && 'Mui-disabled', toClass(sx))}
		style="position:relative;display:inline-flex"
	>
		<input
			class="PrivateSwitchBase-input"
			type="checkbox"
			{checked}
			{disabled}
			aria-label={ariaLabel}
			onclick={(e) => onclick?.(e)}
			onchange={(e) => onchange?.((e.currentTarget as HTMLInputElement).checked)}
		/>
		<SvgIcon name={checked ? 'CheckBox' : 'CheckBoxOutlineBlank'} size="Small" />
	</span>
{/snippet}

{#if label}
	<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
		{@render box()}
		<span class={cx('MuiTypography-root MuiTypography-body1 MuiFormControlLabel-label', toClass(labelSx))}>{label}</span>
	</label>
{:else}
	{@render box()}
{/if}
