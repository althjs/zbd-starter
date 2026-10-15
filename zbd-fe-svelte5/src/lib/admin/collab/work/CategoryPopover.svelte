<script lang="ts">
	/**
	 * 분류 고르기(React CategorySelectorPopover) — 「전체 목록 >」 + 지금 분류 경로 + 「분류 선택」 버튼 → 검색 + 분류 나무(건수·선택·추가).
	 * onselect 가 없으면 분류 화면(/admin/work/{보드}/category/…)으로 이동.
	 */
	import { goto } from '$app/navigation';
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Popover from '../../ui/Popover.svelte';
	import AddWorkButton from './AddWorkButton.svelte';
	import CategoryBreadcrumbs from './CategoryBreadcrumbs.svelte';
	import { getWorkSettings } from './codes.ts';
	import type { WorkItem, Category, CodeItem } from '#lib/shared/types/work.ts';

	let {
		category,
		codes,
		groupKey,
		onselect = null,
		btnText = '분류 선택',
		items = null,
		linkType = '',
		workId = 0,
		isWriteable = true,
		compact = false,
		variant = 'outlined',
		pill = false,
		btnSx = null
	}: {
		category: Category | undefined;
		codes: CodeItem[];
		groupKey: string;
		onselect?: ((c: Category) => void) | null;
		btnText?: string;
		items?: WorkItem[] | null;
		linkType?: '' | 'LIST';
		workId?: number;
		isWriteable?: boolean;
		compact?: boolean;
		variant?: 'text' | 'outlined' | 'contained';
		btnSx?: Sx | null;
		pill?: boolean; // 검색 줄의 「분류」 알약 하나로(업무 목록 거르기, 2026-10-12)
	} = $props();

	let anchor = $state<HTMLElement | null>(null);
	let q = $state('');
	let expanded = $state<string[]>([]);
	type Node = { item: CodeItem; path: string; label: string; cat: Category; depth: number; children: Node[] };

	function build(list: CodeItem[], c1 = 0, c2 = 0, pName = '', pId = '', depth = 0): Node[] {
		const kw = q.toLowerCase();
		const out: Node[] = [];
		for (const item of list) {
			const isC1 = c1 === 0;
			const isC2 = !isC1 && c2 === 0;
			const n1 = isC1 ? item.id : c1;
			const n2 = isC2 ? item.id : c2;
			const path = pId ? `${pId}-${item.id}` : `${item.id}`;
			const name = pName ? `${pName} > ${item.name}` : item.name;
			const children = item.items?.length ? build(item.items, n1, n2, name, path, depth + 1) : [];
			if (kw && !item.name.toLowerCase().includes(kw) && !children.length) continue;
			out.push({ item, path, label: q ? name : item.name, cat: { c1: n1, c2: n2, c3: !isC1 && !isC2 ? item.id : 0 }, depth, children });
		}
		return out;
	}
	const tree = $derived(build(codes));
	$effect(() => {
		// 검색 중이면 맞는 가지 다 펼치기, 대분류가 하나면 전부 펼치기
		const all: string[] = [];
		const walk = (ns: Node[]) => ns.forEach((n) => n.children.length && (all.push(n.path), walk(n.children)));
		walk(tree);
		expanded = q || codes.length === 1 ? all : [];
	});
	function count(n: Node) {
		if (!items) return 0;
		return items.filter((o) => {
			if (o.groupKey !== groupKey) return false;
			const c = o.Json?.category;
			if (!c) return false;
			const [a, b, d] = [Number(c.c1), Number(c.c2), Number(c.c3)];
			if (n.depth === 0) return a === n.item.id;
			if (n.depth === 1) return a === n.cat.c1 && b === n.item.id && d === 0;
			return a === n.cat.c1 && b === n.cat.c2 && d === n.item.id;
		}).length;
	}
	function pick(n: Node) {
		if (onselect) onselect(n.cat);
		else {
			let p = `/admin/work/${groupKey}/category/${n.cat.c1}`;
			if (n.cat.c2) p += `/${n.cat.c2}`;
			if (n.cat.c3) p += `/${n.cat.c3}`;
			goto(p);
		}
		close();
	}
	function toggle(n: Node) {
		if (expanded.includes(n.path)) expanded = expanded.filter((x) => x !== n.path);
		else {
			// 대분류를 펼치면 아래 가지도 모두
			const extra: string[] = [];
			if (n.depth === 0) {
				const walk = (ns: Node[]) => ns.forEach((c) => (extra.push(c.path), walk(c.children)));
				walk(n.children);
			}
			expanded = [...new Set([...expanded, n.path, ...extra])];
		}
	}
	const close = () => ((anchor = null), (q = ''));
	let qEl = $state<HTMLInputElement>();
	let qFocus = $state(false);
	$effect(() => {
		if (anchor && qEl) setTimeout(() => (qEl?.focus(), (qFocus = document.activeElement === qEl)), 0);
	});
	const canAdd = $derived(!compact && btnText != '분류변경' && !getWorkSettings(groupKey)?.disableWrite);
	const pillLabel = $derived.by(() => {
		const c = category || { c1: 0, c2: 0, c3: 0 };
		const a = c.c1 ? codes.find((x) => x.id === c.c1) : null;
		const b = a && c.c2 ? a.items?.find((x) => x.id === c.c2) : null;
		const d = b && c.c3 ? b.items?.find((x) => x.id === c.c3) : null;
		return [a, b, d]
			.filter(Boolean)
			.map((x) => x!.name)
			.join(' › ');
	});
</script>

{#snippet nodes(ns: Node[])}
	<ul role="group" style="margin:0;padding:0;list-style:none">
		{#each ns as n (n.path)}
			{@const open = expanded.includes(n.path)}
			{@const cnt = count(n)}
			<li class="MuiTreeItem-root" role="treeitem" aria-selected="false" aria-expanded={n.children.length ? open : undefined}>
				<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
				<div
					class="MuiTreeItem-content adm-cv-tree"
					style="display:flex;align-items:center;gap:8px;padding:4px 8px 4px {8 + n.depth * 12}px;border-radius:4px;cursor:pointer"
					onclick={() => (n.depth > 0 || !n.children.length ? pick(n) : toggle(n))}
				>
					<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
					<div style="width:16px;display:flex;justify-content:center" onclick={(e) => (e.stopPropagation(), n.children.length && toggle(n))}>
						{#if n.children.length}<Icon name={open ? 'expand_more' : 'chevron_right'} sx={{ fontSize: 18 }} />{/if}
					</div>
					<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;justify-content:space-between;width:100%;padding:4px 8px 4px 0">
						<p class="MuiTypography-root MuiTypography-body2" style="flex-grow:1;font-size:13px;cursor:pointer">
							{n.label}{#if cnt > 0}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2);margin-left:4px;font-size:11px">({cnt})</span>{/if}
						</p>
						<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px">
							{#if n.children.length}<Button size="small" sx={{ minWidth: 'auto', height: 24, fontSize: 11, paddingLeft: 8, paddingRight: 8 }} onclick={(e) => (e.stopPropagation(), pick(n))}
									>선택</Button
								>{/if}
							{#if canAdd}
								<AddWorkButton
									{groupKey}
									category={n.cat}
									defaultTitle={n.item.name}
									disabled={!isWriteable}
									size="small"
									sx={{
										minWidth: 'auto',
										height: 24,
										fontSize: 11,
										paddingLeft: 8,
										paddingRight: 8,
										boxShadow: 'none',
										backgroundColor: 'var(--adm-success-main, #2e7d32)',
										'&:hover': { backgroundColor: 'var(--adm-success-dark, #1b5e20)', boxShadow: 'none' }
									}}
									onclick={(e) => (e.preventDefault(), e.stopPropagation())}
								/>
							{/if}
						</div>
					</div>
				</div>
				{#if n.children.length && open}{@render nodes(n.children)}{/if}
			</li>
		{/each}
	</ul>
{/snippet}

{#if pill}
	<span class="zwf-pill" class:on={!!pillLabel}>
		<button type="button" class="zwf-pill-main" onclick={(e) => (anchor = e.currentTarget as HTMLElement)}>
			<Icon name="folder_open" sx={{ fontSize: 16 }} /><span class="zwf-pill-k">분류</span><span class="zwf-pill-v">{pillLabel || '전체'}</span>{#if !pillLabel}<Icon
					name="expand_more"
					sx={{ fontSize: 16 }}
				/>{/if}
		</button>
		{#if pillLabel && onselect}<button type="button" class="zwf-pill-x" aria-label="분류 해제" onclick={() => onselect?.({ c1: 0, c2: 0, c3: 0 })}><Icon name="close" sx={{ fontSize: 14 }} /></button
			>{/if}
	</span>
{:else}
	<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center">
		{#if !onselect || linkType == 'LIST'}
			<a
				class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover"
				href={`/admin/work/${groupKey}`}
				onclick={(e) => {
					if (onselect) {
						e.preventDefault();
						e.stopPropagation();
						onselect({ c1: 0, c2: 0, c3: 0 });
					}
				}}>전체 목록</a
			>
			<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2);margin-left:8px">&gt;</span>
		{/if}
		{#if workId > 0}<IconButton href={`/admin/work/${groupKey}/${workId}`} sx={{ padding: 0, ...(!onselect || linkType == 'LIST' ? { marginLeft: 8 } : {}) }}
				><Icon name="keyboard_double_arrow_right" size="Small" /></IconButton
			>{/if}
		{#if !compact}<div style="margin-left:{!onselect || linkType == 'LIST' || workId > 0 ? 8 : 0}px"><CategoryBreadcrumbs {category} {codes} {groupKey} {onselect} {linkType} /></div>{/if}
		<Button {variant} size="small" onclick={(e) => (anchor = e.currentTarget as HTMLElement)} sx={{ marginLeft: 8, ...(btnSx || {}) }}>{btnText}</Button>
	</div>
{/if}
<Popover {anchor} onclose={close}>
	<div class="MuiStack-root" style="display:flex;flex-direction:column;padding:16px;min-width:350px;max-width:600px">
		<div class="MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root" style="margin-bottom:8px">
			<div class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-sizeSmall" class:Mui-focused={qFocus}>
				<input
					bind:this={qEl}
					class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall"
					placeholder="카테고리 검색..."
					bind:value={q}
					onfocus={() => (qFocus = true)}
					onblur={() => (qFocus = false)}
				/>
				<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
			</div>
		</div>
		<div class="MuiBox-root" role="tree" style="max-height:calc(100vh - 200px);overflow:auto">{@render nodes(tree)}</div>
	</div>
</Popover>
