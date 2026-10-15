<script lang="ts">
	/**
	 * 관리자 입력칸 — React 관리자 TextField 기본값(filled · 라벨은 늘 위에 붙음) 모양.
	 *   <TextField label="이름" bind:value={name} fullWidth />
	 *   <TextField bind:value={q} size="small" placeholder="검색어" />     라벨이 없으면 hiddenLabel(얇은 칸)
	 *   <TextField label="메모" multiline minRows={3} helperText="500자까지" error={bad} />
	 *   {#snippet end()}<Icon name="search" />{/snippet}                  칸 오른쪽(start 도 있음)
	 * 공용 shared/mui/TextField 는 outlined(사용자 화면·로그인) — 관리자 폼은 이걸 쓴다.
	 */
	import type { Snippet } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		size = 'medium',
		type = 'text',
		multiline = false,
		minRows = 1,
		maxLength = 0,
		fullWidth = false,
		disabled = false,
		readonly = false,
		error = false,
		helperText = '',
		autofocus = false,
		align = '',
		inputmode = undefined,
		class: cls = '',
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
		autofocus?: boolean;
		align?: '' | 'right' | 'center';
		inputmode?: HTMLInputAttributes['inputmode'];
		class?: string;
		sx?: Sx | null;
		inputSx?: Sx | null;
		start?: Snippet;
		end?: Snippet;
		oninput?: (v: string, e: Event) => void;
		onblur?: (e: FocusEvent) => void;
		onkeydown?: (e: KeyboardEvent) => void;
		/** 엔터(여러 줄이 아니면) */
		onenter?: (v: string) => void;
		[k: string]: any;
	} = $props();

	let focused = $state(false);
	let el = $state<HTMLInputElement | HTMLTextAreaElement>();
	const id = `adm-tf-${Math.random().toString(36).slice(2, 8)}`;
	const hidden = $derived(!label);
	const S = $derived(size == 'small');

	$effect(() => {
		if (autofocus) el?.focus();
	});
	function input(e: Event) {
		let v = (e.currentTarget as HTMLInputElement).value;
		if (maxLength && v.length > maxLength) v = v.substring(0, maxLength);
		value = v;
		oninput?.(v, e);
	}
	function fit() {
		if (!multiline || !el) return;
		el.style.height = '0px';
		el.style.height = `${Math.max(el.scrollHeight, minRows * 24)}px`;
	}
	$effect(() => {
		void value;
		fit();
	});
</script>

<div class={cx('MuiFormControl-root MuiTextField-root', fullWidth && 'MuiFormControl-fullWidth', cls, toClass(sx))}>
	{#if label}
		<label
			for={id}
			class={cx(
				'MuiFormLabel-root MuiInputLabel-root MuiInputLabel-formControl MuiInputLabel-animated MuiInputLabel-shrink MuiInputLabel-filled',
				S && 'MuiInputLabel-sizeSmall',
				focused && 'Mui-focused',
				error && 'Mui-error',
				disabled && 'Mui-disabled'
			)}>{label}</label
		>
	{/if}
	<div
		class={cx(
			'MuiInputBase-root MuiFilledInput-root MuiInputBase-colorPrimary MuiInputBase-formControl',
			fullWidth && 'MuiInputBase-fullWidth',
			S && 'MuiInputBase-sizeSmall',
			hidden && 'MuiInputBase-hiddenLabel',
			multiline && 'MuiInputBase-multiline',
			start && 'MuiInputBase-adornedStart',
			end && 'MuiInputBase-adornedEnd',
			focused && 'Mui-focused',
			error && 'Mui-error',
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
				{readonly}
				class={cx('MuiInputBase-input MuiFilledInput-input MuiInputBase-inputMultiline', S && 'MuiInputBase-inputSizeSmall', toClass(inputSx))}
				value={value ?? ''}
				oninput={input}
				onfocus={() => (focused = true)}
				onblur={(e) => ((focused = false), onblur?.(e))}
				{onkeydown}
				{...rest}></textarea>
		{:else}
			<input
				bind:this={el}
				{id}
				{type}
				{placeholder}
				{disabled}
				{readonly}
				{inputmode}
				class={cx(
					'MuiInputBase-input MuiFilledInput-input',
					S && 'MuiInputBase-inputSizeSmall',
					hidden && 'MuiInputBase-inputHiddenLabel',
					toClass({ ...(align ? { textAlign: align } : {}), ...(inputSx || {}) })
				)}
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
		{#if end}<div class="MuiInputAdornment-root MuiInputAdornment-positionEnd">{@render end()}</div>{/if}
	</div>
	{#if helperText}<p class={cx('MuiFormHelperText-root MuiFormHelperText-filled', error && 'Mui-error')}>{helperText}</p>{/if}
</div>
