<script lang="ts">
	/**
	 * 보드 내용 검색(React WorkKeywordSearch) — 목록을 거르지 않고 「그 글자가 어디 적혀 있나」 를 팝오버로.
	 * ↑/↓ 로 고르고 Enter(또는 누르기) = 새 탭으로 그 문서(메모면 ?shapeId=).
	 * Shift 를 누르고 고르면 새 탭 대신 화면 위 창(React 와 같게 — 창 id 는 경로 기반이라 같은 곳은 같은 창).
	 */
	import { tick } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Icon from '../../ui/Icon.svelte';
	import { openRouteWindow } from '../../window/windows.svelte.ts';
	import Chip from '../ui/Chip.svelte';
	import type { WorkItem } from '#lib/shared/types/work.ts';
	import { FIELD_LABEL, findMatches } from './match.ts';

	let { items, groupKey, full = false }: { items: WorkItem[]; groupKey: string; full?: boolean } = $props();
	let keyword = $state('');
	let focused = $state(false);
	let active = $state(-1);
	let root = $state<HTMLElement>();
	let pos = $state({ top: 0, left: 0 });
	const matches = $derived(findMatches(items, keyword));
	const grouped = $derived.by(() => {
		const m = new Map<number, { title: string; list: typeof matches }>();
		for (const x of matches) {
			const g = m.get(x.workId) || { title: x.workTitle, list: [] };
			g.list.push(x);
			m.set(x.workId, g);
		}
		return [...m.entries()];
	});
	const flat = $derived(grouped.flatMap(([id, g]) => g.list.map((x) => ({ workId: id, x }))));
	const open = $derived(focused && !!keyword.trim());
	$effect(() => {
		void keyword;
		active = -1;
	});
	$effect(() => {
		if (!open || !root) return;
		tick().then(() => {
			const r = root!.getBoundingClientRect();
			pos = { top: r.bottom, left: Math.max(8, Math.min(r.left, window.innerWidth - 528)) };
		});
	});
	function go(workId: number, shapeId?: string | number, asWindow = false) {
		focused = false;
		const p = `/admin/work/${groupKey}/${workId}${shapeId ? `?shapeId=${shapeId}` : ''}`;
		if (asWindow) openRouteWindow(p, flat.find((f) => f.workId === workId)?.x.workTitle || p);
		else window.open(p, '_blank');
	}
	function key(e: KeyboardEvent) {
		if (!flat.length) return;
		if (e.key === 'ArrowDown') (e.preventDefault(), (active = (active + 1) % flat.length));
		else if (e.key === 'ArrowUp') (e.preventDefault(), (active = active <= 0 ? flat.length - 1 : active - 1));
		else if (e.key === 'Enter') {
			const t = flat[active >= 0 ? active : 0];
			if (t) (e.preventDefault(), go(t.workId, t.x.shapeId, e.shiftKey));
		} else if (e.key === 'Escape') focused = false;
	}
</script>

<div bind:this={root} class="MuiFormControl-root MuiTextField-root" style="width:{full ? '100%' : '240px'}">
	<div
		class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-formControl MuiInputBase-sizeSmall MuiInputBase-adornedStart"
		class:MuiInputBase-adornedEnd={!!keyword}
		class:Mui-focused={focused}
		style="padding-left:14px;padding-right:{keyword ? 14 : 0}px"
	>
		<div class="MuiInputAdornment-root MuiInputAdornment-positionStart MuiInputAdornment-outlined MuiInputAdornment-sizeSmall">
			<Icon name="search" sx={{ fontSize: 18, color: 'var(--gt-700, #909090)' }} />
		</div>
		<input
			class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall MuiInputBase-inputAdornedStart"
			style="padding:6px 0"
			placeholder="내용 검색 (제목/본문/메모)"
			bind:value={keyword}
			onfocus={() => (focused = true)}
			onkeydown={key}
		/>
		{#if keyword}
			<div class="MuiInputAdornment-root MuiInputAdornment-positionEnd MuiInputAdornment-outlined MuiInputAdornment-sizeSmall">
				<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090);margin-right:4px">{matches.length}건</span>
				<button type="button" class="MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeSmall" style="padding:2px" aria-label="지우기" onclick={() => (keyword = '')}
					><Icon name="close" sx={{ fontSize: 16 }} /></button
				>
			</div>
		{/if}
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
	</div>
</div>
{#if open}
	<div use:portal role="presentation" class="MuiPopover-root MuiModal-root">
		<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (focused = false)}></div>
		<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper" style="top:{pos.top}px;left:{pos.left}px;width:520px;padding:8px;overflow:hidden">
			<div style="max-height:420px;overflow-y:auto">
				{#if !grouped.length}<p class="MuiTypography-root MuiTypography-body2" style="padding:16px;color:var(--gt-700, #909090);text-align:center">일치하는 내용이 없습니다.</p>{/if}
				{#each grouped as [cid, g] (cid)}
					<div style="margin-bottom:8px">
						<div
							class="MuiStack-root"
							style="display:flex;flex-direction:row;align-items:center;gap:4px;position:sticky;top:0;z-index:1;padding:4px 8px;background-color:var(--gt-150, #f5f5f5);border-radius:4px"
						>
							<span class="MuiTypography-root MuiTypography-subtitle2 MuiTypography-noWrap" style="flex-grow:1;min-width:0">{g.title}</span>
							<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">{g.list.length}건</span>
						</div>
						{#each g.list as x, i (i)}
							{@const fi = flat.findIndex((f) => f.workId === cid && f.x === x)}
							<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions, a11y_mouse_events_have_key_events -->
							<div
								class="MuiStack-root"
								style="display:flex;flex-direction:row;align-items:flex-start;gap:6px;padding:5px 8px;margin:1px 2px;cursor:pointer;border-radius:4px;background-color:{fi === active
									? 'var(--gt-150, #f5f5f5)'
									: 'transparent'};outline:{fi === active ? '1px solid var(--theme-color-main)' : 'none'};outline-offset:-1px"
								onmouseover={() => (active = fi)}
								onclick={(e) => go(cid, x.shapeId, e.shiftKey)}
							>
								<Chip
									size="small"
									variant="outlined"
									label={FIELD_LABEL[x.field] + (x.shapeNo ? ` #${x.shapeNo}` : '')}
									sx={{ height: 18, flexShrink: 0, borderColor: 'var(--gt-300, #e0e0e0)', '& .MuiChip-label': { paddingLeft: 6, paddingRight: 6, fontSize: 11 } }}
								/>
								<p class="MuiTypography-root MuiTypography-body2" style="font-size:13px;line-height:1.5;min-width:0;flex-grow:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">
									{x.snippet.slice(0, x.at)}<mark style="background-color:#fde047;color:inherit;padding:0 2px;border-radius:4px">{x.snippet.slice(x.at, x.at + x.len)}</mark>{x.snippet.slice(
										x.at + x.len
									)}
								</p>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}
