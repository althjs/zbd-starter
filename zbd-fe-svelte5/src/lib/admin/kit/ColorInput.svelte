<script lang="ts">
	/**
	 * 색 입력칸(React mui-color-input MuiColorInput · format hex · hiddenLabel) — 왼쪽 색 견본을 누르면 색 고르기, 글자로도 입력.
	 * 견본은 비었을 때 바둑판 무늬. 색 고르기는 공용 팝오버(`#lib/shared/ui/ColorPicker` — 투명도 포함, 불투명이면 #rrggbb).
	 * 공용 후보: 디자인 위젯 편집기 대부분이 쓴다.
	 */
	import TextField from '#lib/admin/ui/TextField.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import ColorPicker from '#lib/shared/ui/ColorPicker.svelte';

	let {
		value = '',
		onChange,
		sx = null,
		size = 'medium',
		label = '',
		fullWidth = false,
		disabled = false
	}: { value?: string; onChange: (v: string) => void; sx?: Sx | null; size?: 'small' | 'medium'; label?: string; fullWidth?: boolean; disabled?: boolean } = $props();
	let btn = $state<HTMLButtonElement>();
	let opened = $state(false);
	const CHECK =
		'linear-gradient(45deg, #ccc 25%, transparent 25%), linear-gradient(135deg, #ccc 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #ccc 75%), linear-gradient(135deg, transparent 75%, #ccc 75%)';
	const swatch = $derived(
		toClass({
			minWidth: 0,
			width: 24,
			height: 24,
			padding: 0,
			borderRadius: 4,
			boxShadow: '0 4px 6px rgba(50, 50, 93, 0.11), 0 1px 3px rgba(0, 0, 0, 0.08)',
			...(value ? { backgroundColor: value } : { backgroundImage: CHECK, backgroundSize: '8px 8px', backgroundPosition: '0 0, 4px 0, 4px -4px, 0 4px' })
		})
	);
</script>

<TextField
	{value}
	{label}
	{size}
	{fullWidth}
	{disabled}
	class={cx('MuiColorInput-TextField')}
	sx={{ backgroundColor: 'var(--zc-fff, #fff)', ...(label ? { '& .MuiInputAdornment-positionStart': { marginTop: 16, height: 24 } } : {}), ...(sx || {}) }}
	inputSx={{ paddingLeft: 0 }}
	oninput={(v) => onChange(v)}
>
	{#snippet start()}
		<span class="notranslate" aria-hidden="true">​</span>
		<button
			type="button"
			class={cx('MuiButtonBase-root MuiButton-root MuiButton-text MuiButton-sizeMedium MuiColorInput-Button', swatch)}
			aria-label="색 고르기"
			{disabled}
			bind:this={btn}
			onclick={() => (opened = !opened)}
		></button>
		{#if opened}<ColorPicker anchor={btn} {value} onchange={onChange} onclose={() => (opened = false)} />{/if}
	{/snippet}
</TextField>
