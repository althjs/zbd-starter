<script lang="ts">
	/**
	 * 분류 하나의 문서들을 한 화면에(React canvas/[groupKey]/category/[c1]/[c2]/[c3]) — 문서마다 관리 폼 카드.
	 * 「하위페이지 보기」 = 아래 분류 문서까지 · 「순서 변경」 = 그 깊이의 sortNo 로 저장(quick-update).
	 * ?workId= 면 그 문서로 내려간다. ⚠ 떠 있는 목차 버튼(WorkTableOfContentsButton)은 안 옮겼다.
	 */
	import { goto, replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { media } from '../../media.svelte.ts';
	import { getHostingId } from '../../meta.svelte.ts';
	import { nav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import SortDialog from '../../ui/SortDialog.svelte';
	import { loadMembers, members } from '../members.svelte.ts';
	import '../ui/collab.css';
	import { CARD } from '../ui/cls.ts';
	import Toggle from '../ui/Toggle.svelte';
	import AddWorkButton from './AddWorkButton.svelte';
	import CategoryBreadcrumbs from './CategoryBreadcrumbs.svelte';
	import { workSearchDefault, getWorkCodes, getWorkSettings, getCategoryLabel } from './codes.ts';
	import type { WorkSearch } from '#lib/shared/types/work.ts';
	import ManageForm from './ManageForm.svelte';
	import { toNotiMembers } from './members.ts';
	import { work, doQuickUpdate, getItems } from './store.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	let { groupKey, c1, c2 = 0, c3 = 0 }: { groupKey: string; c1: number; c2?: number; c3?: number } = $props();
	const codes = $derived(getWorkCodes(groupKey));
	const setting = $derived(getWorkSettings(groupKey));
	const KEY = $derived(SK.work.search(groupKey));
	const actionType = $derived(c3 ? 'sortNo3' : c2 ? 'sortNo2' : 'sortNo1') as 'sortNo1' | 'sortNo2' | 'sortNo3';
	const read = (): WorkSearch => {
		try {
			return JSON.parse(localStorage.getItem(KEY) || 'null') || structuredClone(workSearchDefault);
		} catch {
			return structuredClone(workSearchDefault);
		}
	};
	let showChild = $state(true);
	$effect(() => {
		showChild = read().showChild !== false;
		getItems({ groupKey });
		loadMembers();
	});
	function setShowChild(v: boolean) {
		showChild = v;
		try {
			localStorage.setItem(KEY, JSON.stringify({ showChild: v }));
		} catch {}
	}
	const list = $derived(
		work.items
			.filter((o) => {
				if (o.groupKey != groupKey) return false;
				const c = o.Json?.category;
				if (showChild) return c?.c1 === c1 && (c2 > 0 ? !!c?.c2 && c.c2 === c2 : true) && (c3 > 0 ? !!c?.c3 && c.c3 === c3 : true);
				return c?.c1 === c1 && (c?.c2 ? c.c2 === c2 : true) && (c?.c3 ? c.c3 === c3 : true);
			})
			.sort((a, b) => (a.Json[actionType] ?? 0) - (b.Json[actionType] ?? 0) || a.workId - b.workId)
	);
	$effect(() => {
		const id = page.url.searchParams.get('workId');
		if (!id || !list.length) return;
		setTimeout(() => {
			const el = document.getElementById(`work_id_${id}`);
			if (!el) return;
			window.scrollTo({ top: el.getBoundingClientRect().top + window.pageYOffset - 110, behavior: 'smooth' });
			const u = new URL(location.href);
			u.searchParams.delete('workId');
			replaceState(u.pathname + u.search, page.state);
		}, 500);
	});
	function toList(category = { c1: 0, c2: 0, c3: 0 }) {
		try {
			localStorage.setItem(KEY, JSON.stringify({ ...read(), category }));
		} catch {}
		goto(`/admin/work/${groupKey}`);
	}
	const memberList = $derived(toNotiMembers(members.list, (c) => c === '' || c === getHostingId() || c === admin.profile?.hostingId));
	const cat = $derived({ c1, c2, c3 });
</script>

<PageContainer subtitle=" ">
	{#snippet typography()}
		<h3 class="MuiTypography-root {media.mdDown ? 'MuiTypography-h5' : 'MuiTypography-h3'}">{nav.current?.sub?.title || nav.current?.title || ''}</h3>
		<div class="MuiTypography-root MuiTypography-body1">
			<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;align-items:center">
				<a
					class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover"
					style="font-size:inherit"
					href={`/admin/work/${groupKey}`}
					onclick={(e) => (e.preventDefault(), toList())}>전체 목록</a
				>
				&gt;
				<CategoryBreadcrumbs category={cat} {codes} {groupKey} onselect={(c) => toList(c)} linkType="LIST" />
			</div>
		</div>
	{/snippet}
	<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;align-items:center">
		{#if !setting?.disableWrite}<AddWorkButton {groupKey} category={cat} />{/if}
		<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
			<Toggle size="medium" sx={{ marginLeft: 8 }} checked={showChild} onchange={setShowChild} />
			<span class="MuiTypography-root MuiTypography-body1 MuiFormControlLabel-label">하위페이지 보기</span>
		</label>
		<SortDialog
			items={list.map((o) => ({ id: o.workId, name: o.Json.title, type: getCategoryLabel(o.Json.category, codes) }))}
			onChange={async (os) => {
				if (
					await doQuickUpdate(
						os.map((o) => o.id),
						actionType
					)
				) {
					toast('순서가 변경되었습니다.');
					getItems({ groupKey, workIds: os.map((o) => o.id) });
				} else alert('저리 도중 오류가 발생하였습니다.');
			}}
		/>
	</div>
	<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:24px">
		{#each list as o (o.workId)}
			<div class={cx(CARD, toClass({ padding: 16 }))}><ManageForm item={o} items={list} members={memberList} /></div>
		{:else}
			<div class={cx(CARD, toClass({ marginTop: 8 }))}>
				<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;justify-content:center;align-items:center">
					<p class="MuiTypography-root MuiTypography-body2" style="margin:32px 0">등록된 컨텐츠가 없습니다.</p>
				</div>
			</div>
		{/each}
	</div>
	{#if !setting?.disableWrite}<AddWorkButton {groupKey} category={cat} sx={{ marginTop: 16 }} />{/if}
</PageContainer>
