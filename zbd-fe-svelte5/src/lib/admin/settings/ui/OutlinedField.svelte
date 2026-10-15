<script lang="ts">
	/**
	 * outlined 입력칸(React 관리자의 `<TextField variant="outlined">`) — 검색 줄 등. 관리자 폼 기본(filled)은 ui/TextField.
	 *   <OutlinedField bind:value size="small" placeholder="검색어" onenter={go} sx={{ minWidth: 326 }} />
	 *   multiline · minRows · readonly · inputSx · label(작은 라벨이 테두리에 걸림) · start/end
	 * 공용 shared/mui/TextField(outlined) 에는 입력·엔터 콜백이 없어 관리자용을 따로 둔다(공용 후보).
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		size = 'small',
		type = 'text',
		multiline = false,
		minRows = 1,
		maxLength = 0,
		fullWidth = false,
		disabled = false,
		readonly = false,
		error = false,
		helperText = '',
		inputmode = undefined,
		sx = null,
		inputSx = null,
		start,
		end,
		oninput,
		onblur,
		onkeydown,
		onenter,
		...rest
	}: {
		value?: string | number | null;
		label?: string;
		placeholder?: string;
		size?: 'small' | 'medium';
		type?: string;
		multiline?: boolean;
		minRows?: number;
		maxLength?: number;
		fullWidth?: boolean;
		disabled?: boolean;
		readonly?: boolean;
		error?: boolean;
		helperText?: string;
		inputmode?: any;
		sx?: Sx | null;
		inputSx?: Sx | null;
		start?: Snippet;
		end?: Snippet;
		oninput?: (v: string, e: Event) => void;
		onblur?: (e: FocusEvent) => void;
		onkeydown?: (e: KeyboardEvent) => void;
		onenter?: (v: string) => void;
		[k: string]: any;
	} = $props();

	let focused = $state(false);
	let ta = $state<HTMLTextAreaElement>();
	const id = `adm-of-${Math.random().toString(36).slice(2, 8)}`;
	const S = $derived(size == 'small');

	function input(e: Event) {
		let v = (e.currentTarget as HTMLInputElement).value;
		if (maxLength && v.length > maxLength) {
			v = v.substring(0, maxLength);
			(e.currentTarget as HTMLInputElement).value = v;
		}
		value = v;
		oninput?.(v, e);
	}
	function fit() {
		if (!multiline || !ta) return;
		ta.style.height = '0px';
		ta.style.height = `${Math.max(ta.scrollHeight, minRows * 24)}px`;
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
				error && 'Mui-error'
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
			end && 'MuiInputBase-adornedEnd',
			focused && 'Mui-focused',
			error && 'Mui-error',
			disabled && 'Mui-disabled'
		)}
	>
		{#if start}{@render start()}{/if}
		{#if multiline}
			<textarea
				bind:this={ta}
				{id}
				rows={minRows}
				{placeholder}
				{disabled}
				{readonly}
				class={cx('MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputMultiline', S && 'MuiInputBase-inputSizeSmall', toClass(inputSx))}
				value={value ?? ''}
				oninput={input}
				onfocus={() => (focused = true)}
				onblur={(e) => ((focused = false), onblur?.(e))}
				{onkeydown}
				{...rest}></textarea>
		{:else}
			<input
				{id}
				{type}
				{placeholder}
				{disabled}
				{readonly}
				{inputmode}
				class={cx('MuiInputBase-input MuiOutlinedInput-input', S && 'MuiInputBase-inputSizeSmall', start && 'MuiInputBase-inputAdornedStart', end && 'MuiInputBase-inputAdornedEnd', toClass(inputSx))}
				value={value ?? ''}
				oninput={input}
				onfocus={() => (focused = true)}
				onblur={(e) => ((focused = false), onblur?.(e))}
				onkeydown={(e) => {
					onkeydown?.(e);
					if (e.key == 'Enter' && !e.isComposing) onenter?.(String(value ?? ''));
				}}
				{...rest}
			/>
		{/if}
		{#if end}{@render end()}{/if}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline">
			<legend class={cx(label && 'adm-of-notch')}
				>{#if label}<span>{label}</span>{:else}<span class="notranslate" aria-hidden="true">​</span>{/if}</legend
			>
		</fieldset>
	</div>
	{#if helperText}<p class={cx('MuiFormHelperText-root MuiFormHelperText-sizeSmall MuiFormHelperText-contained', error && 'Mui-error')}>{helperText}</p>{/if}
</div>
