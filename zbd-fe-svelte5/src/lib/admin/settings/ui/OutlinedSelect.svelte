<script lang="ts">
	/**
	 * outlined 선택 상자(React 관리자 `<Select size=small>`) — 공용 shared/mui/Select 와 같은 DOM·목록에
	 * 「테두리 안 작은 라벨」(React 회원 검색의 sx 꾸밈: 라벨 12px 회색 + 값 아래로 16px) 을 더할 수 있다.
	 *   <OutlinedSelect value={v} options={[{ value: '0', label: '정상회원' }]} label="회원 상태" onchange={(v) => …} sx={{ minWidth: 150 }} />
	 * item 스니펫을 주면 값·목록 칸을 직접 그린다(색 칸 + 이름 등).
	 * labelStyle="common" 은 React components/common/Select 의 isLabeled 모양(높이 46.73 · 값 위 17 아래 2).
	 * variant="standard" 면 밑줄 상자(MUI Select variant=standard).
	 */
	import { tick, type Snippet } from 'svelte';
	import '../settings.css';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';

	type Opt = { value: any; label: string; disabled?: boolean };
	let {
		value,
		options,
		label = '',
		labelStyle = 'search',
		size = 'small',
		variant = 'outlined',
		disabled = false,
		fullWidth = false,
		placeholder = '',
		sx = null,
		item,
		onchange
	}: {
		value: any;
		options: Opt[];
		label?: string;
		labelStyle?: 'search' | 'common' | 'area';
		size?: 'small' | 'medium';
		variant?: 'outlined' | 'standard' | 'filled';
		disabled?: boolean;
		fullWidth?: boolean;
		placeholder?: string;
		sx?: Sx | null;
		item?: Snippet<[Opt]>;
		onchange: (v: any) => void;
	} = $props();

	let open = $state(false);
	const listId = `adm-sel-${Math.random().toString(36).slice(2, 8)}`;
	let root = $state<HTMLElement>();
	let paper = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0, minWidth: 0 });
	const current = $derived(options.find((o) => o.value === value) || options.find((o) => String(o.value) === String(value ?? '')));

	async function show() {
		if (!root || disabled) return;
		open = true;
		await tick();
		const r = root.getBoundingClientRect();
		const sel = paper?.querySelector('.Mui-selected') as HTMLElement | null;
		const offset = sel ? sel.offsetTop + sel.offsetHeight / 2 - r.height / 2 : 0;
		const h = paper?.offsetHeight || 0;
		const top = Math.max(16, Math.min(r.top - offset, window.innerHeight - h - 16));
		pos = { top, left: Math.max(16, r.left), minWidth: r.width };
	}
	function pick(v: any) {
		open = false;
		if (v !== value) onchange(v);
	}
	const V = $derived(variant == 'outlined' ? 'Outlined' : variant == 'filled' ? 'Filled' : 'Input');
</script>

<div
	bind:this={root}
	class={cx(
		`MuiInputBase-root Mui${V === 'Input' ? 'Input' : V + 'Input'}-root MuiInputBase-colorPrimary MuiSelect-root`,
		size == 'small' && 'MuiInputBase-sizeSmall',
		variant == 'standard' && 'MuiInput-underline adm-sel-std',
		fullWidth && 'MuiInputBase-fullWidth',
		label && (labelStyle == 'common' ? 'adm-sel-labeled adm-sel-lab2' : labelStyle == 'area' ? 'adm-sel-labeled adm-sel-lab3' : 'adm-sel-labeled'),
		open && 'Mui-focused',
		disabled && 'Mui-disabled',
		toClass(sx)
	)}
>
	<div
		role="combobox"
		aria-expanded={open}
		aria-haspopup="listbox"
		aria-controls={listId}
		aria-disabled={disabled}
		tabindex={disabled ? -1 : 0}
		class={cx(
			`MuiSelect-select MuiSelect-${variant} MuiInputBase-input`,
			variant == 'outlined' && 'MuiOutlinedInput-input',
			variant == 'standard' && 'MuiInput-input',
			size == 'small' && 'MuiInputBase-inputSizeSmall'
		)}
		onclick={show}
		onkeydown={(e) => (e.key == 'Enter' || e.key == ' ' || e.key == 'ArrowDown') && (e.preventDefault(), show())}
	>
		{#if current && item}{@render item(current)}{:else if current?.label}{current.label}{:else if placeholder}<span class="adm-sel-ph">{placeholder}</span>{:else}<span
				class="notranslate"
				aria-hidden="true">​</span
			>{/if}
	</div>
	<input aria-invalid="false" aria-hidden="true" tabindex="-1" class="MuiSelect-nativeInput" value={value ?? ''} readonly />
	<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSelect-icon MuiSelect-icon{V}" focusable="false" aria-hidden="true" viewBox="0 0 24 24"><path d="M7 10l5 5 5-5z"></path></svg>
	{#if variant == 'outlined'}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline">
			<legend
				>{#if label}<span>{label}</span>{:else}<span class="notranslate" aria-hidden="true">​</span>{/if}</legend
			>
		</fieldset>
	{/if}
</div>

{#if open}
	<div use:portal role="presentation" class="MuiPopover-root MuiMenu-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (open = false)}></div>
		<div
			bind:this={paper}
			class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper MuiMenu-paper"
			style="top:{pos.top}px;left:{pos.left}px;min-width:{pos.minWidth}px"
		>
			<ul id={listId} class="MuiList-root MuiList-padding MuiMenu-list" role="listbox" tabindex="-1">
				{#each options as o (o.value)}
					<li
						class={cx('MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters', String(o.value) === String(value ?? '') && 'Mui-selected', o.disabled && 'Mui-disabled')}
						role="option"
						aria-selected={String(o.value) === String(value ?? '')}
						tabindex="0"
						onclick={() => !o.disabled && pick(o.value)}
						onkeydown={(e) => e.key == 'Enter' && !o.disabled && pick(o.value)}
					>
						{#if item}{@render item(o)}{:else}{o.label}{/if}
					</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}
<svelte:window onkeydown={(e) => open && e.key == 'Escape' && (open = false)} />
