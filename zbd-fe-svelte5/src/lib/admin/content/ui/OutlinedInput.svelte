<script lang="ts">
	/**
	 * 한 줄 outlined 입력칸 + 앞 글자(startAdornment) — 공용 shared/mui/TextField 에 앞 자리가 없어 콘텐츠 묶음이 둔다.
	 *   <OutlinedInput value={name} oninput={(v) => …} size="small">{#snippet start()}<span …>작성자</span>{/snippet}</OutlinedInput>
	 *   end={snippet} 이면 칸 오른쪽(지우기 단추 등)
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value = '',
		placeholder = '',
		size = 'small',
		disabled = false,
		type = 'text',
		sx = null,
		oninput,
		start,
		end
	}: {
		value?: string | number;
		placeholder?: string;
		size?: 'small' | 'medium';
		disabled?: boolean;
		type?: string;
		sx?: Sx | null;
		oninput: (v: string) => void;
		start?: Snippet;
		end?: Snippet;
	} = $props();
	let focused = $state(false);
</script>

<div
	class={cx(
		'MuiFormControl-root MuiTextField-root',
		toClass({
			...(start ? { '& .MuiInputBase-adornedStart': { paddingLeft: 14 }, '& .MuiInputBase-inputAdornedStart': { paddingLeft: 0 } } : {}),
			...(end ? { '& .MuiInputBase-adornedEnd': { paddingRight: 14 }, '& .MuiInputBase-adornedEnd > input': { paddingRight: 0 } } : {}),
			...(sx || {})
		})
	)}
>
	<div
		class={cx(
			'MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-formControl',
			start && 'MuiInputBase-adornedStart',
			end && 'MuiInputBase-adornedEnd',
			size == 'small' && 'MuiInputBase-sizeSmall',
			disabled && 'Mui-disabled',
			focused && 'Mui-focused'
		)}
	>
		{@render start?.()}
		<input
			{type}
			{placeholder}
			{disabled}
			class={cx('MuiInputBase-input MuiOutlinedInput-input', size == 'small' && 'MuiInputBase-inputSizeSmall', start && 'MuiInputBase-inputAdornedStart')}
			value={value ?? ''}
			oninput={(e) => oninput((e.currentTarget as HTMLInputElement).value)}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
		/>
		{@render end?.()}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
	</div>
</div>
