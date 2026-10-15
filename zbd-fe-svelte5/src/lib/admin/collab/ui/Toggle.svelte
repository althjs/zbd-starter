<script lang="ts">
	/** 스위치(MUI Switch, 기본 small) — 이름표 없이. size="medium" 이면 큰 것 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let {
		checked = false,
		disabled = false,
		size = 'small',
		sx = null,
		onchange
	}: { checked?: boolean; disabled?: boolean; size?: 'small' | 'medium'; sx?: Sx | null; onchange?: (v: boolean) => void } = $props();
</script>

<span class={cx('MuiSwitch-root', size == 'small' ? 'MuiSwitch-sizeSmall' : 'MuiSwitch-sizeMedium', toClass(sx))}>
	<span class="MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root" class:Mui-checked={checked} class:Mui-disabled={disabled}>
		<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" {checked} {disabled} onchange={(e) => onchange?.((e.currentTarget as HTMLInputElement).checked)} />
		<span class="MuiSwitch-thumb"></span>
	</span>
	<span class="MuiSwitch-track"></span>
</span>
