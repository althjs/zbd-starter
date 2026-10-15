<script lang="ts">
	/**
	 * 글자로 보이다가 누르면 고치는 칸(React components/common/input/TextInputInline).
	 * 보일 때: 값(없으면 회색 placeholder), hint 면 밑줄 · 누르면 outlined 작은 칸 → 엔터·벗어나기로 끝(onBlur), Esc 로 되돌림.
	 *   <InlineText value={o.name} hint placeholder="메뉴 이름을 입력하세요" onChange={(v) => …} />
	 *   useButton 이면 「취소·적용」 단추로만 바꾼다. multiline · isNumber · formatter · hintText
	 */
	import { tick } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import { getNumber } from '../common.ts';

	let {
		value,
		placeholder = '',
		multiline = false,
		minRows = 1,
		onChange = () => {},
		onBlur = null,
		hint = false,
		useButton = false,
		fullWidth = false,
		readonly = false,
		isNumber = false,
		hintText = '',
		sx = null,
		previewSx = null,
		formatter = (v: string) => v
	}: {
		value: any;
		placeholder?: string;
		multiline?: boolean;
		minRows?: number;
		onChange?: (v: any) => void;
		onBlur?: ((v: any) => void) | null;
		hint?: boolean;
		useButton?: boolean;
		fullWidth?: boolean;
		readonly?: boolean;
		isNumber?: boolean;
		hintText?: string;
		sx?: Sx | null;
		previewSx?: Sx | null;
		formatter?: (v: string) => string;
	} = $props();

	let edit = $state(false);
	let text = $state<any>('');
	let el = $state<HTMLInputElement | HTMLTextAreaElement>();
	$effect(() => {
		if (!edit) text = value ?? '';
	});
	async function start() {
		if (readonly) return;
		text = value ?? '';
		edit = true;
		await tick();
		setTimeout(() => (el?.focus(), el?.select()), 100);
	}
	function input(e: Event) {
		let v: any = (e.currentTarget as HTMLInputElement).value;
		v = isNumber ? getNumber(v) : formatter(v);
		text = v;
		if (!useButton) onChange(v);
	}
	function finish() {
		if (useButton) return;
		onBlur?.(text);
		edit = false;
	}
	const html = $derived(
		String(value || placeholder || '')
			.replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' })[c] as string)
			.replaceAll('\n', '<br/>')
	);
</script>

{#if !edit}
	<div class={cx('MuiStack-root', toClass({ position: 'relative', ...(sx || {}) }))}>
		<div
			class={cx('MuiBox-root', toClass({ width: '100%', borderBottom: hint && !readonly ? '1px solid #eee' : undefined, color: !value ? '#bbb' : undefined, ...(previewSx || {}) }))}
			role="button"
			tabindex="0"
			onclick={start}
			onkeydown={(e) => e.key == 'Enter' && start()}
		>
			{@html html}
		</div>
	</div>
{:else}
	<div class={cx('MuiFormControl-root MuiTextField-root', fullWidth && 'MuiFormControl-fullWidth', toClass(sx))}>
		<div
			class={cx(
				'MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary Mui-focused MuiInputBase-formControl MuiInputBase-sizeSmall',
				fullWidth && 'MuiInputBase-fullWidth',
				multiline && 'MuiInputBase-multiline',
				toClass({ borderRadius: 0 })
			)}
		>
			{#if multiline}
				<textarea
					bind:this={el}
					rows={minRows}
					class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputMultiline MuiInputBase-inputSizeSmall"
					{placeholder}
					value={text}
					oninput={input}
					onblur={finish}
					onkeyup={(e) => e.key == 'Escape' && ((text = value), (edit = false))}></textarea>
			{:else}
				<input
					bind:this={el}
					class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall"
					{placeholder}
					value={isNumber ? getNumber(text) : text}
					oninput={input}
					onblur={finish}
					onkeyup={(e) => {
						if (e.key == 'Enter' && !useButton) finish();
						else if (e.key == 'Escape') ((text = value), (edit = false));
					}}
				/>
			{/if}
			{#if useButton}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8 }))}>
					<Button size="small" color="warning" variant="outlined" onclick={() => ((text = value), (edit = false))}>취소</Button>
					<Button size="small" color="success" variant="outlined" onclick={() => (onChange(text), (edit = false))}>적용</Button>
				</div>
			{/if}
			<fieldset aria-hidden="true" class={cx('MuiOutlinedInput-notchedOutline', toClass({ borderRadius: 0 }))}><legend><span class="notranslate">​</span></legend></fieldset>
		</div>
	</div>
	{#if hintText}
		<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ display: 'flex', flexDirection: 'row', alignItems: 'center', color: 'var(--gt-700, #909090)', marginTop: 4, minWidth: 120 }))}
			><Icon name="info" sx={{ marginRight: 4, fontSize: 13 }} />{hintText}</span
		>
	{/if}
{/if}
