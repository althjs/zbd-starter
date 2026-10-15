<script lang="ts">
	/**
	 * 업무 문서 하나(React canvas/[groupKey]/[workId]) — 제목 아래 「전체 목록 > 분류 > 제목」, 본문은 문서 관리 폼.
	 * 초대된 문서(work-invite) 면 제목 = 문서 제목, 설명 = 「초대된 업무입니다.」.
	 * 실시간: 다른 곳에서 고친 문서·도형을 따라온다(onWorkUpdate).
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { media } from '../../media.svelte.ts';
	import { getHostingId } from '../../meta.svelte.ts';
	import { nav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { onAdminSocket } from '../../socket.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import { toast } from '../../ui/modal.svelte.ts';
	import { loadMembers, members } from '../members.svelte.ts';
	import '../ui/collab.css';
	import { CARD } from '../ui/cls.ts';
	import CategoryBreadcrumbs from './CategoryBreadcrumbs.svelte';
	import { workSearchDefault, getWorkCodes, getWorkSettings, getWorkSharedHostingIds } from './codes.ts';
	import type { WorkItem } from '#lib/shared/types/work.ts';
	import ManageForm from './ManageForm.svelte';
	import { toNotiMembers } from './members.ts';
	import { work, getItem, getItems, syncShape } from './store.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	let { groupKey, workId, inviteWorkId = 0 }: { groupKey: string; workId: number; inviteWorkId?: number } = $props();
	const codes = $derived(getWorkCodes(groupKey));
	const setting = $derived(getWorkSettings(groupKey));
	const item = $derived<WorkItem | null>(work.items.find((x) => x.workId === workId) || null);
	const SEARCH_KEY = $derived(SK.work.search(groupKey));
	const subImageIndex = $derived(page.url.searchParams.get('subImageIndex') ? Number(page.url.searchParams.get('subImageIndex')) : null);
	const ymd = $derived(page.url.searchParams.get('ymd'));

	$effect(() => {
		const id = workId;
		loadMembers();
		getItem(id).then((x) => {
			if (!x) {
				goto('/admin', { replaceState: true });
				toast('권한이 없습니다.');
			}
		});
		// 연관 문서 제목(📚)·연관 고르기에 보드 목록이 필요하다
		getItems({ groupKey });
	});
	$effect(() => {
		return onAdminSocket('onWorkUpdate', (d: any) => {
			if (d.groupKey && groupKey !== d.groupKey) return;
			if (d.shape) syncShape(d.workId, d.shape, d.shape.subImageIndex);
			else if (d.workId) getItem(d.workId);
		});
	});
	function toList(category = { c1: 0, c2: 0, c3: 0 }) {
		let s: any = null;
		try {
			s = JSON.parse(localStorage.getItem(SEARCH_KEY) || 'null');
		} catch {}
		try {
			localStorage.setItem(SEARCH_KEY, JSON.stringify({ ...(s || workSearchDefault), category }));
		} catch {}
		goto(`/admin/work/${groupKey}`);
	}
	const memberList = $derived.by(() => {
		const master = item?.hostingId === '';
		const allowed = new Set(['', admin.profile?.hostingId || '', getHostingId(), ...(master ? getWorkSharedHostingIds(groupKey) : [])]);
		const ms = toNotiMembers([...members.list, ...work.adminMembers.filter((a) => !members.list.find((m) => m.adminId === a.adminId))], (c) => allowed.has(c));
		for (const x of item?.Json?.invitedAdmins || []) if (!ms.find((m) => m.adminId == x.adminId)) ms.push(x);
		return ms;
	});
</script>

<PageContainer title={inviteWorkId ? item?.Json?.title || '' : ''} subtitle=" ">
	{#snippet typography()}
		<h3 class="MuiTypography-root {media.mdDown ? 'MuiTypography-h5' : 'MuiTypography-h3'}">{inviteWorkId ? item?.Json?.title || '' : nav.current?.sub?.title || nav.current?.title || ''}</h3>
		{#if !setting?.disableSearchForm}
			<div class="MuiTypography-root MuiTypography-body1">
				{#if inviteWorkId}
					초대된 업무입니다.
				{:else}
					<div class="MuiStack-root" style="display:flex;flex-direction:row;flex-wrap:wrap;gap:4px 8px;align-items:center">
						<a
							class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover"
							style="font-size:inherit;white-space:nowrap"
							href={`/admin/work/${groupKey}`}
							onclick={(e) => (e.preventDefault(), toList())}>전체 목록</a
						>
						&gt;
						<CategoryBreadcrumbs
							category={{ c1: Number(item?.Json?.category?.c1), c2: Number(item?.Json?.category?.c2) || 0, c3: Number(item?.Json?.category?.c3) || 0 }}
							{codes}
							{groupKey}
							onselect={(c) => toList(c)}
							linkType="LIST"
						/>
						{#if item?.Json?.title}<p class="MuiTypography-root MuiTypography-body2">{' '}&gt; {item.Json.title}</p>{/if}
					</div>
				{/if}
			</div>
		{/if}
	{/snippet}
	{#if item}
		<div class={cx(CARD, toClass({ padding: 16 }))}>
			<ManageForm paramWorkId={workId} {subImageIndex} {item} items={work.items} {inviteWorkId} members={memberList} />
		</div>
	{/if}
	{#if ymd}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;margin-top:16px">
			<Button size="small" onclick={() => goto(`/admin/work/${groupKey}?ymd=${ymd}`)}>{ymd.substring(0, 7).replace('-', '.')}월 달력 보기</Button>
		</div>
	{/if}
</PageContainer>
