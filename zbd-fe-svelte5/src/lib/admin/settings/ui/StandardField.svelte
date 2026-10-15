<script lang="ts">
	/**
	 * 밑줄 입력칸(MUI TextField variant="standard") — 관리자 편집 서랍(관리자 회원 등)이 쓴다.
	 *   <StandardField bind:value placeholder="…" fullWidth sx={{ mt: 8 }} />   label 을 주면 위에 작은 라벨(shrink)
	 *   multiline · minRows · type · end(오른쪽 끝) · size(기본 small — React 관리자 서랍은 거의 small: 위 여백 1px)
	 */
	import type { Snippet } from 'svelte';
	import '../settings.css';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		type = 'text',
		multiline = false,
		minRows = 1,
		maxLength = 0,
		fullWidth = false,
		size = 'small',
		disabled = false,
		sx = null,
		end,
		oninput,
		onenter,
		...rest
	}: {
		value?: string | number | null;
		label?: string;
		placeholder?: string;
		type?: string;
		multiline?: boolean;
		minRows?: number;
		maxLength?: number;
		fullWidth?: boolean;
		size?: 'small' | 'medium';
		disabled?: boolean;
		sx?: Sx | null;
		end?: Snippet;
		oninput?: (v: string) => void;
		onenter?: (v: string) => void;
		[k: string]: any;
	} = $props();
	let ta = $state<HTMLTextAreaElement>();
	const id = `adm-sf-${Math.random().toString(36).slice(2, 8)}`;
	function input(e: Event) {
		let v = (e.currentTarget as HTMLInputElement).value;
		if (maxLength && v.length > maxLength) {
			v = v.substring(0, maxLength);
			(e.currentTarget as HTMLInputElement).value = v;
		}
		value = v;
		oninput?.(v);
	}
	$effect(() => {
		void value;
		if (!multiline || !ta) return;
		ta.style.height = '0px';
		ta.style.height = `${Math.max(ta.scrollHeight, minRows * 23)}px`;
	});
</script>

<div class={cx('MuiFormControl-root MuiTextField-root', fullWidth && 'MuiFormControl-fullWidth', label && 'adm-sf-labeled', toClass(sx))}>
	{#if label}<label for={id} class="MuiFormLabel-root MuiInputLabel-root MuiInputLabel-formControl MuiInputLabel-animated MuiInputLabel-shrink MuiInputLabel-standard adm-sf-label">{label}</label>{/if}
	<div
		class={cx(
			'MuiInputBase-root MuiInput-root MuiInput-underline MuiInputBase-colorPrimary adm-std-input',
			fullWidth && 'MuiInputBase-fullWidth',
			size == 'small' && 'MuiInputBase-sizeSmall',
			multiline && 'MuiInputBase-multiline adm-std-multi',
			end && 'MuiInputBase-adornedEnd',
			disabled && 'Mui-disabled'
		)}
	>
		{#if multiline}
			<textarea bind:this={ta} {id} rows={minRows} {placeholder} {disabled} class="MuiInputBase-input MuiInput-input MuiInputBase-inputMultiline" value={value ?? ''} oninput={input} {...rest}
			></textarea>
		{:else}
			<input
				{id}
				{type}
				{placeholder}
				{disabled}
				class="MuiInputBase-input MuiInput-input"
				value={value ?? ''}
				oninput={input}
				onkeydown={(e) => e.key == 'Enter' && !e.isComposing && onenter?.(String(value ?? ''))}
				{...rest}
			/>
		{/if}
		{#if end}{@render end()}{/if}
	</div>
</div>
