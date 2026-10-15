<script lang="ts">
	/**
	 * 선택칸 — 관리자 입력칸(ui/TextField, filled · 라벨이 칸 안 위쪽)과 **같은 부품 위에** 그린다(2026-10-04 대표 — 「Select 쪽 스타일이 계속 틀어진다」).
	 * OutlinedSelect 는 테두리 모양이라 TextField 와 한 줄에 놓으면 높이·라벨 위치가 어긋났다 → 입력칸과 섞는 폼에서는 이걸 쓴다.
	 *   <SelectField label="상태" value={v} options={[{ value: 'plan', label: '기획' }]} onchange={(v) => …} />
	 * 목록은 OutlinedSelect 와 같은 메뉴(고른 줄이 칸 위치에 오게). ↑↓·Enter·Esc.
	 */
	import { tick } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { cx, type Sx } from '#lib/shared/mui/sx.ts';
	import TextField from '../../ui/TextField.svelte';

	type Opt = { value: any; label: string; disabled?: boolean };
	let {
		value,
		options,
		label = '',
		placeholder = '',
		size = 'small',
		fullWidth = true,
		disabled = false,
		sx = null,
		onchange
	}: { value: any; options: Opt[]; label?: string; placeholder?: string; size?: 'small' | 'medium'; fullWidth?: boolean; disabled?: boolean; sx?: Sx | null; onchange: (v: any) => void } = $props();

	let open = $state(false);
	let box = $state<HTMLElement>();
	let paper = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0, minWidth: 0 });
	const current = $derived(options.find((o) => String(o.value) === String(value ?? '')));
	const listId = `adm-sf-${Math.random().toString(36).slice(2, 8)}`;

	async function show() {
		if (!box || disabled) return;
		open = true;
		await tick();
		const r = box.getBoundingClientRect();
		const sel = paper?.querySelector('.Mui-selected') as HTMLElement | null;
		const offset = sel ? sel.offsetTop + sel.offsetHeight / 2 - r.height / 2 : 0;
		const h = paper?.offsetHeight || 0;
		pos = { top: Math.max(16, Math.min(r.top - offset, window.innerHeight - h - 16)), left: Math.max(16, r.left), minWidth: r.width };
	}
	function pick(v: any) {
		open = false;
		if (String(v) !== String(value ?? '')) onchange(v);
	}
	function keys(e: KeyboardEvent) {
		if (e.key == 'Enter' || e.key == ' ' || e.key == 'ArrowDown') {
			e.preventDefault();
			show();
		}
	}
</script>

<div bind:this={box} class="sf" role="combobox" aria-expanded={open} aria-controls={listId} tabindex="-1" onclick={show} onkeydown={keys}>
	<TextField {size} {fullWidth} {label} {placeholder} {disabled} {sx} readonly value={current?.label ?? ''} inputSx={{ cursor: disabled ? 'default' : 'pointer' }}>
		{#snippet end()}<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium sf-ic" class:up={open} focusable="false" aria-hidden="true" viewBox="0 0 24 24"><path d="M7 10l5 5 5-5z"></path></svg
			>{/snippet}
	</TextField>
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
						{o.label}
					</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}
<svelte:window onkeydown={(e) => open && e.key == 'Escape' && (open = false)} />

<style>
	.sf {
		min-width: 0;
		outline: none;
	}
	.sf-ic {
		color: var(--adm-n500, #667085);
		transition: transform 150ms;
		margin-right: -4px;
	}
	.sf-ic.up {
		transform: rotate(180deg);
	}
</style>
