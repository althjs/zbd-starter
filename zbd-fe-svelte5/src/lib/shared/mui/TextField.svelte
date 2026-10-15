<script lang="ts">
	/**
	 * MUI TextField(outlined) 호환 — 한 줄 또는 여러 줄(minRows 만큼, 내용 따라 늘어남).
	 * end 자리(endAdornment) 에 비밀글 체크박스 같은 것을 넣는다.
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		value = $bindable(''),
		placeholder = '',
		multiline = false,
		minRows = 1,
		size = 'medium',
		fullWidth = true,
		disabled = false,
		autofocus = false,
		maxLength = 0,
		type = 'text',
		class: cls = '',
		sx = null,
		onblur,
		oninput,
		onpaste,
		onkeydown,
		start,
		end
	}: {
		value?: string;
		placeholder?: string;
		multiline?: boolean;
		minRows?: number;
		size?: 'small' | 'medium';
		fullWidth?: boolean;
		disabled?: boolean;
		autofocus?: boolean;
		maxLength?: number;
		type?: string;
		class?: string;
		sx?: Sx | null;
		onblur?: (e: FocusEvent) => void;
		onpaste?: (e: ClipboardEvent) => void;
		/** 값이 바뀔 때(bind:value 대신 쓸 때) */
		oninput?: (v: string) => void;
		onkeydown?: (e: KeyboardEvent) => void;
		/** 앞 자리(startAdornment) — 검색 아이콘 등. `.MuiInputAdornment-root` 로 감싸서 준다 */
		start?: Snippet;
		end?: Snippet;
	} = $props();

	let focused = $state(false);
	let ta = $state<HTMLTextAreaElement>();
	const LINE = 24; // theme MuiOutlinedInput input lineHeight
	let height = $state(minRows * LINE);

	function fit() {
		if (!ta) return;
		const prev = ta.style.height;
		ta.style.height = '0px';
		const h = Math.max(ta.scrollHeight, minRows * LINE);
		ta.style.height = prev;
		height = h;
	}
	$effect(() => {
		void value;
		fit();
	});
	function input(e: Event) {
		let v = (e.currentTarget as HTMLTextAreaElement | HTMLInputElement).value;
		if (maxLength && v.length > maxLength) v = v.substring(0, maxLength);
		value = v;
		oninput?.(v);
	}
	$effect(() => {
		if (autofocus) ta?.focus();
	});
</script>

<div class={cx('MuiFormControl-root', fullWidth && 'MuiFormControl-fullWidth', 'MuiTextField-root', cls, toClass(sx))}>
	<div
		class={cx(
			'MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary',
			disabled && 'Mui-disabled',
			focused && 'Mui-focused',
			fullWidth && 'MuiInputBase-fullWidth',
			'MuiInputBase-formControl',
			size == 'small' && 'MuiInputBase-sizeSmall',
			multiline && 'MuiInputBase-multiline',
			start && 'MuiInputBase-adornedStart',
			end && 'MuiInputBase-adornedEnd'
		)}
	>
		{#if start}<div class="MuiInputAdornment-root MuiInputAdornment-positionStart MuiInputAdornment-outlined MuiInputAdornment-sizeMedium">{@render start()}</div>{/if}
		{#if multiline}
			<textarea
				bind:this={ta}
				rows={minRows}
				{placeholder}
				{disabled}
				class={cx(
					'MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputMultiline',
					size == 'small' && 'MuiInputBase-inputSizeSmall',
					disabled && 'Mui-disabled',
					end && 'MuiInputBase-inputAdornedEnd'
				)}
				style="height:{height}px"
				{value}
				oninput={input}
				{onpaste}
				onfocus={() => (focused = true)}
				onblur={(e) => {
					focused = false;
					onblur?.(e);
				}}></textarea>
			<!-- MUI 자동 높이 계산용 그림자(보이지 않음) — DOM 을 MUI 와 같게 -->
			<textarea
				aria-hidden="true"
				class="MuiInputBase-input MuiOutlinedInput-input{disabled ? ' Mui-disabled' : ''}"
				readonly
				tabindex="-1"
				style="visibility:hidden;position:absolute;overflow:hidden;height:0;top:0;left:0;transform:translateZ(0);padding-top:0;padding-bottom:0;width:100%"></textarea>
		{:else}
			<input
				{type}
				{placeholder}
				{disabled}
				class={cx('MuiInputBase-input MuiOutlinedInput-input', size == 'small' && 'MuiInputBase-inputSizeSmall', start && 'MuiInputBase-inputAdornedStart', disabled && 'Mui-disabled')}
				{value}
				oninput={input}
				{onkeydown}
				{onpaste}
				onfocus={() => (focused = true)}
				onblur={(e) => {
					focused = false;
					onblur?.(e);
				}}
			/>
		{/if}
		{@render end?.()}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
	</div>
</div>
