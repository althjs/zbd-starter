<script lang="ts">
	/**
	 * 라벨 있는 outlined 입력칸(MUI TextField variant="outlined" — 라벨이 테두리에 걸린다). 공용 shared/mui/TextField 는 라벨이 없다.
	 *   <OutlinedField label="제목" bind:value={t} size="small" fullWidth />   multiline minRows maxRows · start 스니펫 · onkeydown
	 * 라벨은 늘 위에 붙어 있다(관리자 테마 shrink).
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		size = 'small',
		multiline = false,
		minRows = 1,
		maxRows = 0,
		maxLength = 0,
		fullWidth = true,
		disabled = false,
		autofocus = false,
		sx = null,
		inputSx = null,
		start,
		onkeydown,
		oninput
	}: {
		value?: string;
		label?: string;
		placeholder?: string;
		size?: 'small' | 'medium';
		multiline?: boolean;
		minRows?: number;
		maxRows?: number;
		maxLength?: number;
		fullWidth?: boolean;
		disabled?: boolean;
		autofocus?: boolean;
		sx?: Sx | null;
		inputSx?: Sx | null;
		start?: Snippet;
		onkeydown?: (e: KeyboardEvent) => void;
		oninput?: (v: string) => void;
	} = $props();
	let focused = $state(false);
	let el = $state<HTMLInputElement | HTMLTextAreaElement>();
	const id = `adm-of-${Math.random().toString(36).slice(2, 8)}`;
	const S = $derived(size == 'small');
	$effect(() => {
		if (autofocus) el?.focus();
	});
	function input(e: Event) {
		let v = (e.currentTarget as HTMLInputElement).value;
		if (maxLength && v.length > maxLength) v = v.substring(0, maxLength);
		value = v;
		oninput?.(v);
	}
	function fit() {
		if (!multiline || !el) return;
		el.style.height = '0px';
		const lh = 23;
		let h = Math.max(el.scrollHeight, minRows * lh);
		if (maxRows) h = Math.min(h, maxRows * lh);
		el.style.height = `${h}px`;
		el.style.overflow = maxRows && el.scrollHeight > maxRows * lh ? 'auto' : 'hidden';
	}
	$effect(() => {
		void value;
		fit();
	});
</script>

<div class={cx('MuiFormControl-root MuiTextField-root', fullWidth && 'MuiFormControl-fullWidth', toClass(sx))}>
	{#if label}
		<label
			for={id}
			class={cx(
				'MuiFormLabel-root MuiInputLabel-root MuiInputLabel-formControl MuiInputLabel-animated MuiInputLabel-shrink MuiInputLabel-outlined',
				S && 'MuiInputLabel-sizeSmall',
				focused && 'Mui-focused',
				disabled && 'Mui-disabled'
			)}>{label}</label
		>
	{/if}
	<div
		class={cx(
			'MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-formControl',
			fullWidth && 'MuiInputBase-fullWidth',
			S && 'MuiInputBase-sizeSmall',
			multiline && 'MuiInputBase-multiline',
			start && 'MuiInputBase-adornedStart',
			focused && 'Mui-focused',
			disabled && 'Mui-disabled'
		)}
	>
		{#if start}<div class="MuiInputAdornment-root MuiInputAdornment-positionStart">{@render start()}</div>{/if}
		{#if multiline}
			<textarea
				bind:this={el}
				{id}
				rows={minRows}
				{placeholder}
				{disabled}
				class={cx('MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputMultiline', S && 'MuiInputBase-inputSizeSmall', toClass(inputSx))}
				value={value ?? ''}
				oninput={input}
				onfocus={() => (focused = true)}
				onblur={() => (focused = false)}
				{onkeydown}></textarea>
		{:else}
			<input
				bind:this={el}
				{id}
				{placeholder}
				{disabled}
				class={cx('MuiInputBase-input MuiOutlinedInput-input', S && 'MuiInputBase-inputSizeSmall', start && 'MuiInputBase-inputAdornedStart', toClass(inputSx))}
				value={value ?? ''}
				oninput={input}
				onfocus={() => (focused = true)}
				onblur={() => (focused = false)}
				{onkeydown}
			/>
		{/if}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline">
			<legend style={label ? 'font-size:0.75em;height:11px;visibility:hidden;max-width:100%;white-space:nowrap' : ''}>
				{#if label}<span style="padding-left:5px;padding-right:5px;display:inline-block;opacity:0;visibility:visible">{label}</span>{:else}<span class="notranslate" aria-hidden="true">​</span>{/if}
			</legend>
		</fieldset>
	</div>
</div>
