<script lang="ts">
	/**
	 * MUI Select(outlined) 호환 — 누르면 목록(Menu) 이 body 에 뜬다. DOM·클래스는 MUI 와 같게(운영자 CSS 가 `.MuiOutlinedInput-root` 등을 겨냥).
	 * 목록 위치: 고른 항목이 상자 위에 겹치게(MUI 와 같은 방식).
	 */
	import { tick, type Snippet } from 'svelte';
	import { portal } from '../ui/portal';
	import { cx, sx as toClass, type Sx } from './sx';

	type Opt = { value: any; label: string; disabled?: boolean; caption?: string };
	let {
		value,
		options,
		size = 'small',
		class: cls = '',
		sx = null,
		onchange,
		optionIcon = undefined
	}: {
		value: any;
		options: Opt[];
		size?: 'small' | 'medium';
		class?: string;
		sx?: Sx | null;
		onchange: (v: any) => void;
		/** 항목 앞 아이콘(부르는 쪽이 그린다 — 관리자·사용자 아이콘 글꼴이 달라서) */
		optionIcon?: Snippet<[Opt]>;
	} = $props();

	let open = $state(false);
	let root = $state<HTMLElement>();
	let paper = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0, minWidth: 0 });
	const current = $derived(options.find((o) => o.value === value) || options.find((o) => String(o.value) === String(value)));

	async function show() {
		if (!root) return;
		open = true;
		await tick();
		const r = root.getBoundingClientRect();
		const sel = paper?.querySelector('.Mui-selected') as HTMLElement | null;
		const offset = sel ? sel.offsetTop + sel.offsetHeight / 2 - r.height / 2 : 0;
		const h = paper?.offsetHeight || 0;
		let top = r.top - offset;
		top = Math.max(16, Math.min(top, window.innerHeight - h - 16));
		pos = { top, left: Math.max(16, r.left), minWidth: r.width };
	}
	function pick(v: any) {
		open = false;
		if (v !== value) onchange(v);
	}
</script>

<div
	bind:this={root}
	class={cx(`MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary${size == 'small' ? ' MuiInputBase-sizeSmall' : ''} MuiSelect-root`, open && 'Mui-focused', cls, toClass(sx))}
>
	<div
		role="combobox"
		aria-expanded={open}
		aria-haspopup="listbox"
		tabindex="0"
		class="MuiSelect-select MuiSelect-outlined MuiInputBase-input MuiOutlinedInput-input{size == 'small' ? ' MuiInputBase-inputSizeSmall' : ''}"
		onclick={show}
		onkeydown={(e) => (e.key == 'Enter' || e.key == ' ' || e.key == 'ArrowDown') && (e.preventDefault(), show())}
	>
		{#if current?.label}{#if optionIcon}{@render optionIcon(current)}{/if}{current.label}{:else}<span class="notranslate" aria-hidden="true">​</span>{/if}
	</div>
	<input aria-invalid="false" aria-hidden="true" tabindex="-1" class="MuiSelect-nativeInput" value={value ?? ''} readonly />
	<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSelect-icon MuiSelect-iconOutlined" focusable="false" aria-hidden="true" viewBox="0 0 24 24"><path d="M7 10l5 5 5-5z"></path></svg>
	<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
</div>

{#if open}
	<div use:portal role="presentation" class="MuiPopover-root MuiMenu-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (open = false)}></div>
		<div
			bind:this={paper}
			class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper MuiMenu-paper"
			style="top:{pos.top}px;left:{pos.left}px;min-width:{pos.minWidth}px"
		>
			<ul class="MuiList-root MuiList-padding MuiMenu-list" role="listbox" tabindex="-1">
				{#each options as o (o.value)}
					<li
						class="MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters{o.value === value ? ' Mui-selected' : ''}{o.disabled ? ' Mui-disabled' : ''}"
						role="option"
						aria-selected={o.value === value}
						aria-disabled={o.disabled || undefined}
						tabindex={o.disabled ? -1 : 0}
						onclick={() => !o.disabled && pick(o.value)}
						onkeydown={(e) => e.key == 'Enter' && !o.disabled && pick(o.value)}
					>
						{#if optionIcon}{@render optionIcon(o)}{/if}{o.label}{#if o.caption}<span class="MuiTypography-root MuiTypography-caption" style="margin-left:8px;color:var(--zbd-gray-700, #6c737f)"
								>{o.caption}</span
							>{/if}
					</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}
<svelte:window onkeydown={(e) => open && e.key == 'Escape' && (open = false)} />
