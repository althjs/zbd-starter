<script lang="ts">
	/** 상태 라벨 여러 개 고르기(React StatusLabelFilter) — 지금 목록에 있는 상태 이름만, 건수와 함께. 없으면 안 그린다 */
	import { tick } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Check from '../ui/Check.svelte';
	import Icon from '../../ui/Icon.svelte';
	import { getWorkItemStatusLabel } from './codes.ts';
	import type { WorkItem } from '#lib/shared/types/work.ts';

	let { items, value = [], onchange, pill = false }: { items: WorkItem[]; value?: string[]; onchange: (v: string[]) => void; pill?: boolean } = $props();
	const counts = $derived.by(() => {
		const m = new Map<string, number>();
		for (const o of items) {
			const l = getWorkItemStatusLabel(o);
			if (l) m.set(l, (m.get(l) || 0) + 1);
		}
		return m;
	});
	const labels = $derived([...counts.keys()].sort());
	let open = $state(false);
	let root = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0, w: 0 });
	async function show() {
		open = true;
		await tick();
		const r = root!.getBoundingClientRect();
		pos = { top: r.bottom, left: r.left, w: r.width };
	}
	const toggle = (l: string) => onchange(value.includes(l) ? value.filter((x) => x !== l) : [...value, l]);
</script>

{#if labels.length}
	{#if pill}
		<span bind:this={root} class="zwf-pill" class:on={value.length > 0}>
			<button type="button" class="zwf-pill-main" aria-expanded={open} onclick={show}>
				<Icon name="flag" sx={{ fontSize: 16 }} /><span class="zwf-pill-k">상태</span><span class="zwf-pill-v"
					>{!value.length ? '전체' : value.length <= 2 ? value.join(', ') : `${value.length}개`}</span
				>{#if !value.length}<Icon name="expand_more" sx={{ fontSize: 16 }} />{/if}
			</button>
			{#if value.length}<button type="button" class="zwf-pill-x" aria-label="상태 해제" onclick={() => onchange([])}><Icon name="close" sx={{ fontSize: 14 }} /></button>{/if}
		</span>
	{:else}
		<div class="MuiFormControl-root" style="min-width:140px">
			<div bind:this={root} class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-sizeSmall MuiSelect-root" class:Mui-focused={open} style="height:34px">
				<div
					role="combobox"
					aria-controls="adm-cv-status-list"
					aria-expanded={open}
					tabindex="0"
					class="MuiSelect-select MuiSelect-outlined MuiSelect-multiple MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall"
					style="padding-top:0;padding-bottom:0;display:flex;align-items:center"
					onclick={show}
					onkeydown={(e) => e.key == 'Enter' && show()}
				>
					{#if !value.length}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2)">상태 전체</span>
					{:else if value.length <= 2}<span class="MuiTypography-root MuiTypography-caption" style="font-weight:600">상태: {value.join(', ')}</span>
					{:else}<span class="MuiTypography-root MuiTypography-caption" style="font-weight:600">상태 {value.length}건</span>{/if}
				</div>
				<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSelect-icon MuiSelect-iconOutlined" focusable="false" aria-hidden="true" viewBox="0 0 24 24"><path d="M7 10l5 5 5-5z"></path></svg>
				<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
			</div>
		</div>
	{/if}
	{#if open}
		<div use:portal role="presentation" class="MuiPopover-root MuiMenu-root MuiModal-root">
			<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (open = false)}></div>
			<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper MuiMenu-paper" style="top:{pos.top}px;left:{pos.left}px;min-width:{pos.w}px">
				<ul class="MuiList-root MuiList-padding MuiMenu-list" role="listbox" id="adm-cv-status-list">
					{#each labels as l (l)}
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<li
							class="MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters"
							class:Mui-selected={value.includes(l)}
							role="option"
							aria-selected={value.includes(l)}
							style="padding-top:4px;padding-bottom:4px"
							onclick={() => toggle(l)}
						>
							<Check checked={value.includes(l)} sx={{ padding: 4, marginRight: 4 }} onclick={(e) => e.preventDefault()} />
							<div class="MuiListItemText-root MuiListItemText-multiline" style="display:flex;flex-direction:column">
								<span class="MuiTypography-root MuiTypography-body2">{l}</span>
								<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2)">{counts.get(l) || 0}건</span>
							</div>
						</li>
					{/each}
					{#if value.length}
						<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
						<div style="border-top:1px solid #eee;padding:4px 8px">
							<span class="MuiTypography-root MuiTypography-caption" style="cursor:pointer;color:#1d4ed8;font-weight:600" onclick={(e) => (e.stopPropagation(), onchange([]))}>전체 해제</span>
						</div>
					{/if}
				</ul>
			</div>
		</div>
	{/if}
{/if}
<svelte:window onkeydown={(e) => open && e.key == 'Escape' && (open = false)} />
