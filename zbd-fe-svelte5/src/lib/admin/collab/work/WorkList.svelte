<script lang="ts">
	/**
	 * 업무 보드 목록(React canvas/[groupKey]) — 위 거르기(분류·그룹핑·보기 방식·담당자·상태·기간 · 내용 검색) + 분류별 카드 표.
	 * 거르기 값은 이 브라우저 `zbd:work:search:{보드}`, 분류 접힘은 `zbd:work:collapsed:{보드}`.
	 * ⚠ v1 목록 대기 — 보드 글을 통째로 받아 화면에서 거른다(React 와 같음). Go `/api/v1` 목록이 생기면 서버 쪽으로.
	 * ⚠ 캘린더 보기는 아직 안 옮겼다(고르면 안내만).
	 */
	import { page } from '$app/state';
	import { replaceState } from '$app/navigation';
	import { date } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { media } from '../../media.svelte.ts';
	import { adminImageUrl, getHostingId, isSuperAdmin } from '../../meta.svelte.ts';
	import { admin, checkAuth } from '../../profile.svelte.ts';
	import { onAdminSocket } from '../../socket.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import { toast } from '../../ui/modal.svelte.ts';
	import { loadMembers, members } from '../members.svelte.ts';
	import { CARD } from '../ui/cls.ts';
	import Toggle from '../ui/Toggle.svelte';
	import AddWorkButton from './AddWorkButton.svelte';
	import CategoryPopover from './CategoryPopover.svelte';
	import {
		workSearchDefault,
		clone,
		getWorkCodes,
		getWorkColumns,
		getWorkItemStatusLabel,
		getWorkSettings,
		getStatusLabelsForView,
		isWorkSharedFromMaster,
		isWorkWriteable,
		setStatusLabelsForView
	} from './codes.ts';
	import type { WorkItem, WorkSearch } from '#lib/shared/types/work.ts';
	import Kanban from './Kanban.svelte';
	import WorkCalendar from './WorkCalendar.svelte';
	import KeywordSearch from './KeywordSearch.svelte';
	import '../ui/collab.css';
	import ListTable from './ListTable.svelte';
	import { toNotiMembers } from './members.ts';
	import ScheduleField from './ScheduleField.svelte';
	import StatusLabelFilter from './StatusLabelFilter.svelte';
	import { work, getItem, getItems, syncShape } from './store.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	let { groupKey }: { groupKey: string } = $props();
	const profile = $derived(admin.profile);
	const codes = $derived(getWorkCodes(groupKey));
	const columns = $derived(getWorkColumns(groupKey));
	const setting = $derived(getWorkSettings(groupKey));
	const useSchedule = $derived(!!columns.find((o) => o.inputType == 'SCHEDULE'));
	const writeable = $derived(isWorkWriteable(groupKey, profile));
	const SEARCH_KEY = $derived(SK.work.search(groupKey));
	const COLLAPSE_KEY = $derived(SK.work.collapsed(groupKey));
	const ls = {
		get(k: string) {
			try {
				return JSON.parse(localStorage.getItem(k) || 'null');
			} catch {
				return null;
			}
		},
		set(k: string, v: any) {
			try {
				localStorage.setItem(k, JSON.stringify(v));
			} catch {}
		}
	};

	let search = $state<WorkSearch>(clone(workSearchDefault));
	let collapsed = $state<string[]>([]);
	$effect(() => {
		const gk = groupKey;
		let s: WorkSearch = ls.get(SK.work.search(gk)) || clone(getWorkSettings(gk)?.defaultSearchOption || workSearchDefault);
		const qv = page.url.searchParams.get('viewMode');
		if (qv) {
			s = { ...s, viewMode: qv as any };
			const u = new URL(location.href);
			u.searchParams.delete('viewMode');
			replaceState(u.pathname + u.search, page.state);
		}
		if (gk === 'CALENDAR' && !isSuperAdmin(admin.profile)) s = { ...s, viewMode: 'CALENDAR' };
		search = s;
		collapsed = (ls.get(SK.work.collapsed(gk)) || []).map(String);
		getItems({ groupKey: gk });
		loadMembers();
	});
	function setSearch(s: WorkSearch) {
		search = s;
		ls.set(SEARCH_KEY, s);
	}

	// 실시간 — 다른 곳에서 고친 문서·도형을 따라온다
	$effect(() => {
		const gk = groupKey;
		return onAdminSocket('onWorkUpdate', (d: any) => {
			if (d.groupKey && gk !== d.groupKey) return;
			if (d.shape) syncShape(d.workId, d.shape, d.shape.subImageIndex);
			else if (d.workId) getItem(d.workId);
			else if (d.workIds?.length > 0) {
				getItems({ groupKey: gk, workIds: d.workIds });
				toast(`${admin.profile?.browserUid === (d.zbduid ?? d.oxuid) ? '' : '다른곳에서 '} ${d.workIds.length}건의 블럭이 업데이트 되었습니다.`);
			}
		});
	});

	const lastTs = (o: WorkItem) => Math.max(new Date(o.updatedAt || o.createdAt).getTime(), ...(o.Json?.shapes || []).map((s) => s.ts || 0));
	const filtered = $derived.by(() => {
		const s = search;
		let os = work.items.filter((o) => o.groupKey == groupKey);
		if (s.schedule?.from && s.schedule?.to) {
			const f = new Date(s.schedule.from).getTime();
			const t = new Date(s.schedule.to).getTime();
			os = os.filter((o) => lastTs(o) >= f && lastTs(o) <= t);
		}
		if (s.days) {
			const d = new Date();
			d.setDate(d.getDate() - s.days);
			const limit = new Date(date(d, true).slice(0, 10).replace(/\./g, '-')).getTime();
			os = os.filter((o) => lastTs(o) >= limit);
		}
		if (s.category.c1 > 0) os = os.filter((o) => o.Json?.category?.c1 == s.category.c1);
		if (s.category.c2 > 0) os = os.filter((o) => o.Json?.category?.c2 == s.category.c2);
		if (s.category.c3 > 0) os = os.filter((o) => o.Json?.category?.c3 == s.category.c3);
		const preAdmin = os;
		if (s.adminIds?.length)
			os = os.filter((o) =>
				s.adminIds.some((id) =>
					Object.keys(o.Json.adminNotiMap || {})
						.map(Number)
						.includes(id)
				)
			);
		const preStatus = os;
		const labels = getStatusLabelsForView(s);
		if (labels.length) os = os.filter((o) => labels.includes(getWorkItemStatusLabel(o)));
		os = [...os].sort((a, b) => b.workId - a.workId);
		return { list: os, preAdmin, preStatus };
	});
	const list = $derived(filtered.list);
	const groups = $derived.by(() => {
		const m = new Map<string, WorkItem[]>();
		for (const o of list) {
			const k = String(o.Json?.category?.c1);
			m.set(k, [...(m.get(k) || []), o]);
		}
		return m;
	});
	const visibleC1s = $derived([...groups.keys()]);
	const allCollapsed = $derived(visibleC1s.length > 0 && visibleC1s.every((k) => collapsed.includes(k)));
	function toggleC1(k: string) {
		collapsed = collapsed.includes(k) ? collapsed.filter((x) => x !== k) : [...collapsed, k];
		ls.set(COLLAPSE_KEY, collapsed);
	}
	function setAll(on: boolean) {
		collapsed = on ? [...visibleC1s] : [];
		ls.set(COLLAPSE_KEY, collapsed);
	}

	// 담당자 거르기 후보 — 지금 걸러진 글(담당자 거르기 전)의 담당자
	const filterAdmins = $derived.by(() => {
		const ids = new Set<number>();
		filtered.preAdmin.forEach((o) => Object.keys(o.Json.adminNotiMap || {}).forEach((k) => k && ids.add(Number(k))));
		return members.list.filter((m) => ids.has(m.adminId));
	});
	const notiMembers = $derived.by(() => {
		const cid = getHostingId();
		return toNotiMembers(members.list, (c) => c === cid);
	});
	const otherGroups = $derived([...groups.entries()].filter(([k]) => !codes.find((c) => String(c.id) === k)));
	const viewModes = $derived([
		{ value: '', label: '표', icon: 'table_rows' },
		{ value: 'CANVAN', label: '칸반', icon: 'view_kanban' },
		...(useSchedule ? [{ value: 'CALENDAR', label: '캘린더', icon: 'calendar_month' }] : [])
	]);
	// 걸어 둔 거르기 수 — 하나라도 있을 때만 「초기화」 를 보인다(보기 방식·묶기는 거르기가 아니라 그대로 둔다)
	const activeFilters = $derived(
		(search.category?.c1 > 0 ? 1 : 0) + (getStatusLabelsForView(search).length ? 1 : 0) + (search.schedule?.from || search.days ? 1 : 0) + (search.adminIds?.length ? 1 : 0)
	);
	let daysText = $state('');
	$effect(() => {
		daysText = search.days ? String(search.days) : '';
	});
	// 휴대폰은 웹메일 목록처럼 카드 없이 화면 끝까지(2026-10-05 대표) — 페이지 여백(16)만큼 밖으로, 모서리·그림자 없이 위아래 선만
	const bleed = $derived(
		media.mdDown ? { marginLeft: -16, marginRight: -16, borderRadius: 0, boxShadow: 'none', borderTop: '1px solid var(--adm-divider)', borderBottom: '1px solid var(--adm-divider)' } : {}
	);
	// 표 카드는 좌우 여백도 없앤다(머리 줄만 16)
	const listBleed = $derived(media.mdDown ? { ...bleed, paddingLeft: 0, paddingRight: 0 } : {});
</script>

{#snippet groupCard(key: string, name: string, items: WorkItem[], addCode: any)}
	{@const fold = collapsed.includes(key)}
	<div class={cx(CARD, toClass({ padding: 16, marginTop: 16, marginBottom: 16, ...listBleed }))}>
		<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;justify-content:space-between;padding:0 {media.mdDown ? 16 : 0}px">
			<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
			<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;margin-bottom:{fold ? 0 : 16}px;cursor:pointer;user-select:none" onclick={() => toggleC1(key)}>
				<Icon name={fold ? 'chevron_right' : 'expand_more'} sx={{ color: 'var(--gt-700, #909090)' }} />
				<h6 class="MuiTypography-root MuiTypography-h6" style="font-weight:bold">{name}</h6>
				<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">({items.length})</span>
			</div>
			{#if addCode && !fold && !search.category.c2 && !setting?.disableWrite}
				<AddWorkButton {groupKey} disabled={!writeable} defaultTitle={addCode.name || ''} category={{ c1: addCode.id, c2: 0, c3: 0 }} size="small" variant="contained" />
			{/if}
		</div>
		{#if !fold}<ListTable {groupKey} {items} admins={notiMembers} />{/if}
	</div>
{/snippet}

<PageContainer>
	{#if !setting?.disableSearchForm || isSuperAdmin(profile)}
		<!-- 검색 줄(2026-10-12 대표 「검색 UI 전면 개편」) — 위: 내용 검색 · 보기 방식 · 분류 관리 / 아래: 거르기 알약(분류·상태·기간·담당) · 묶기·접기 · 초기화 -->
		<div class={cx(CARD, 'zwf', toClass({ ...bleed, ...(setting?.disableSearchForm ? { background: 'var(--gt-400, #d0d0d0)' } : {}) }))}>
			{#if setting?.disableSearchForm}<span class="MuiTypography-root MuiTypography-caption">최고 관리자만 보입니다.</span>{/if}
			<div class="zwf-top">
				<div class="zwf-q"><KeywordSearch items={list} {groupKey} full /></div>
				{#if !(groupKey === 'CALENDAR' && !isSuperAdmin(profile))}
					<div class="zwf-seg" role="radiogroup" aria-label="보기 방식">
						{#each viewModes as v (v.value)}
							{@const on = (search.viewMode || '') === v.value}
							<button type="button" role="radio" aria-checked={on} class:on title={v.label} onclick={() => setSearch({ ...search, viewMode: v.value as any })}>
								<Icon name={v.icon} sx={{ fontSize: 18 }} /><span>{v.label}</span>
							</button>
						{/each}
					</div>
				{/if}
				{#if !isWorkSharedFromMaster(groupKey) && checkAuth('/admin/setting/meta/COMMON_CODES')}
					<a
						class="zwf-icon"
						href="/admin/setting/meta/COMMON_CODES"
						title="분류 관리"
						aria-label="분류 관리"
						onclick={() => profile?.hostingId == '' && localStorage.setItem(SK.admin.hostingTab, '')}
					>
						<Icon name="settings" sx={{ fontSize: 20 }} />
					</a>
				{/if}
			</div>
			<div class="zwf-filters">
				<CategoryPopover pill category={search.category} {codes} {groupKey} onselect={(category) => setSearch({ ...search, category })} items={work.items} isWriteable={writeable} linkType="LIST" />
				<StatusLabelFilter pill items={filtered.preStatus} value={getStatusLabelsForView(search)} onchange={(n) => setSearch(setStatusLabelsForView(search, n))} />
				<span class="zwf-pill" class:on={!!search.schedule?.from || !!search.days}>
					<ScheduleField schedule={search.schedule} fromLabel="시작" toLabel="종료" onchange={(s) => setSearch({ ...search, schedule: s || { from: '', to: '' }, days: s?.from ? 0 : search.days })}>
						{#snippet template()}
							<span class="zwf-pill-main">
								<Icon name="event" sx={{ fontSize: 16 }} /><span class="zwf-pill-k">기간</span>
								{#if search.schedule?.from}<span class="zwf-pill-v">{search.schedule.from.slice(5)} ~ {search.schedule.to.slice(5)}</span>
								{:else}
									<span class="zwf-pill-v">최근</span>
									<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
									<input
										class="zwf-days"
										inputmode="numeric"
										placeholder="-"
										aria-label="최근 며칠"
										bind:value={daysText}
										onclick={(e) => (e.preventDefault(), e.stopPropagation())}
										onkeydown={(e) => e.key === 'Enter' && (e.currentTarget as HTMLInputElement).blur()}
										onblur={() => setSearch({ ...search, days: Number(String(daysText).replace(/[^0-9]/g, '')) || 0 })}
									/><span class="zwf-pill-v">일</span>
								{/if}
							</span>
						{/snippet}
					</ScheduleField>
					{#if search.schedule?.from || search.days}<button type="button" class="zwf-pill-x" aria-label="기간 해제" onclick={() => setSearch({ ...search, schedule: { from: '', to: '' }, days: 0 })}
							><Icon name="close" sx={{ fontSize: 14 }} /></button
						>{/if}
				</span>
				{#if filterAdmins.length}
					<span class="zwf-sep"></span>
					<div class="zwf-people" aria-label="담당자">
						{#each filterAdmins as m (m.adminId)}
							{@const on = search.adminIds?.includes(m.adminId)}
							{@const img = adminImageUrl(`admin_profile_${m.adminId}`, m.Json?.image?.name, m.hostingId)}
							<button
								type="button"
								class="zwf-person"
								class:on
								aria-pressed={on}
								title={m.name || String(m.adminId)}
								onclick={() => setSearch({ ...search, adminIds: on ? search.adminIds.filter((x) => x !== m.adminId) : [...(search.adminIds || []), m.adminId] })}
							>
								{#if img}<img src={img} alt="" />{:else}<span>{(m.name || '?').slice(0, 1)}</span>{/if}
								<em>{m.name || m.adminId}</em>
							</button>
						{/each}
					</div>
				{/if}
				<div class="zwf-grow"></div>
				{#if !search.viewMode}
					<label class="zwf-switch" title="분류별로 묶어 보기">
						<Toggle checked={search.isGroupping || false} onchange={(v) => setSearch({ ...search, isGroupping: v })} /><span>분류별로 묶기</span>
					</label>
					{#if search.isGroupping && visibleC1s.length > 1}
						<button type="button" class="zwf-text" onclick={() => setAll(!allCollapsed)}>
							<Icon name={allCollapsed ? 'unfold_more' : 'unfold_less'} sx={{ fontSize: 18 }} />{allCollapsed ? '전체 펼침' : '전체 접기'}
						</button>
					{/if}
				{/if}
				{#if activeFilters}
					<button type="button" class="zwf-text zwf-reset" onclick={() => setSearch({ ...clone(workSearchDefault), viewMode: search.viewMode, isGroupping: search.isGroupping })}>
						<Icon name="filter_alt_off" sx={{ fontSize: 18 }} />초기화 <b>{activeFilters}</b>
					</button>
				{/if}
			</div>
		</div>
	{/if}

	{#if search.viewMode === 'CANVAN'}
		<div class="MuiBox-root" style="margin:16px 0"><Kanban {groupKey} items={list} /></div>
	{:else if search.viewMode === 'CALENDAR'}
		<div class="MuiBox-root" style="margin:16px 0"><WorkCalendar {groupKey} items={list} admins={notiMembers} /></div>
	{:else if search.isGroupping}
		{#each codes as code (code.id)}
			{@const g = groups.get(String(code.id))}
			{#if g}{@render groupCard(String(code.id), code.name, g, code)}{/if}
		{/each}
		{#each otherGroups as [k, g] (k)}{@render groupCard(k, `기타 (${k})`, g, null)}{/each}
		{#if !groups.size}<p class="MuiTypography-root MuiTypography-body2" style="margin:40px 0;text-align:center">데이터가 없습니다.</p>{/if}
	{:else}
		<div class={cx(CARD, toClass({ padding: 16, marginTop: 16, marginBottom: 16, ...listBleed }))}><ListTable {groupKey} items={list} admins={notiMembers} /></div>
	{/if}

	{#if search.category.c1 > 0 && !setting?.disableWrite}
		<AddWorkButton disabled={!writeable} {groupKey} category={search.category} />
	{/if}
</PageContainer>
