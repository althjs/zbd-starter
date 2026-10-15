<script lang="ts">
	/**
	 * 관리자 게시판 목록(React board/[groupKey]/index) — 이용약관 등 게시판과 고객센터(cs)가 같이 쓴다.
	 *
	 * ⚠ v1 목록 대기(CLAUDE.md 4 · LIST_AUDIT 1-3): 비ES 게시판은 지금 API(`GET /service/admin/post/group/{groupKey}`)가
	 *   그룹 글 **전부**를 주고 화면이 기간·검색어·하위분류·정렬·쪽을 거른다(React 판과 같음).
	 *   Go `GET /api/v1/admin/posts` 가 생기면 search 를 그대로 요청 값으로 넘기고 이 거르기를 지운다.
	 *   검색 엔진(feature.useES) 게시판은 이미 서버가 나눈다(getEsPosts).
	 */
	import { page } from '$app/state';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { activeNav } from '../../nav.svelte.ts';
	import SortDialog from '../../ui/SortDialog.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { getYesterdayYMD, getYMD } from '../dates.ts';
	import PostManageDrawer from '../post/PostManageDrawer.svelte';
	import { getPostGroup, postDefault } from '../post/postGroup.ts';
	import { getEsPosts, getPostGroupItems, posts, setPostGroupItem } from '../post/postStore.svelte.ts';
	import ScrollTabs from '../ui/ScrollTabs.svelte';
	import type { BoardSearch } from '../types.ts';
	import type { Post, PostGroup } from '#lib/shared/types/post.ts';
	import BoardListTable from './BoardListTable.svelte';
	import BoardSearchForm from './BoardSearchForm.svelte';
	import { SK } from '#lib/shared/storage-keys.ts';

	let {
		groupKey,
		customItems = null,
		customPostGroup = null,
		onChangePost,
		useDetail = false,
		path = '',
		masterGroupIds = null,
		buttonText = '글 작성'
	}: {
		groupKey: string;
		customItems?: Post[] | null;
		customPostGroup?: PostGroup | null;
		onChangePost?: (p: Post) => void;
		useDetail?: boolean;
		path?: string;
		masterGroupIds?: number[] | null;
		buttonText?: string;
	} = $props();

	const LS_SIZE = SK.admin.boardPageSize;
	function searchDefault(gk: string): BoardSearch {
		const g = getPostGroup(gk);
		let size = 10;
		try {
			size = parseInt(localStorage.getItem(LS_SIZE) || '10', 10) || 10;
		} catch {}
		return {
			groupKey: gk,
			groupSubKey: '',
			orderBy: 'postId',
			orderByDesc: 'desc',
			keyword: '',
			pageSize: size,
			pageNo: 1,
			from: g?.feature?.useCreatedAtEdit ? '2020-01-01' : '2024-01-01',
			to: getYesterdayYMD(new Date(), -30),
			nickname: '',
			status: null
		};
	}

	const postGroup = $derived(customPostGroup || getPostGroup(groupKey));
	const meta = $derived(postGroup);
	const useES = $derived(!!postGroup?.feature?.useES);
	let search = $state<BoardSearch>(searchDefault(''));
	let itemList = $state<Post[] | null>(null);
	let filteredTotal = $state(0);
	let pm = $state<any>();

	const items = $derived(customItems ? customItems : posts.groupMap[groupKey] || null);
	const dynamicSubGroups = $derived(meta?.isDynamicSubGroupKey && items ? [...new Set(items.map((o) => o.groupSubKey))].map((g) => ({ groupSubKey: g, name: '' })) : null);

	// 게시판이 바뀌면 검색 초기화 + 목록 읽기
	let lastKey = '';
	$effect(() => {
		if (groupKey == lastKey) return;
		lastKey = groupKey;
		search = searchDefault(groupKey);
		itemList = null;
		if (!customItems && !useES) getPostGroupItems(groupKey);
		if (useES) getData(search);
	});
	$effect(() => {
		if (!items || useES) return;
		getData(search);
	});

	function ordered(list: Post[], s: BoardSearch) {
		const dir = s.orderByDesc == 'asc' ? 1 : -1;
		return [...list].sort((a: any, b: any) => {
			const sa = a.sortNo ?? Infinity;
			const sb = b.sortNo ?? Infinity;
			if (sa !== sb) return sa < sb ? -1 : 1;
			const x = a[s.orderBy];
			const y = b[s.orderBy];
			if (x == y) return 0;
			return (x > y ? 1 : -1) * dir;
		});
	}

	async function getData(s: BoardSearch) {
		try {
			localStorage.setItem(LS_SIZE, String(s.pageSize || 10));
		} catch {}
		if (useES) {
			const o = await getEsPosts({ ...s, groupKey });
			filteredTotal = o.totalCount;
			itemList = o.posts;
			search = s;
			return;
		}
		let ps = [...(items || [])];
		if (s.groupSubKey) ps = ps.filter((o) => o.groupSubKey == s.groupSubKey);
		const kw = (s.keyword || '').trim();
		if (kw) {
			const isPartner = postGroup?.component?.name == 'PARTNER';
			const col = (postGroup?.dataColumn || 'json') as 'json' | 'jsonUser';
			ps = ps.filter(
				(o) => o.title.indexOf(kw) !== -1 || (o[col]?.body || '').indexOf(kw) !== -1 || (isPartner && ['ceoName', 'companyName', 'text1', 'text2'].some((k) => (o.json?.[k] || '').indexOf(kw) !== -1))
			);
		}
		if (s.from && s.to) ps = ps.filter((o) => getYMD(o.createdAt) >= s.from && getYMD(o.createdAt) <= s.to);
		if (typeof s.status == 'number') ps = ps.filter((o) => o.status == s.status);
		if (s.scheduleStatus) {
			const ts = Date.now();
			ps = ps.filter((o) => {
				const f = o.json?.schedule?.from ? new Date(o.json.schedule.from).getTime() : 0;
				const t = o.json?.schedule?.to ? new Date(o.json.schedule.to).getTime() : 0;
				return (ts < f ? '대기' : ts < t ? '진행중' : '종료') == s.scheduleStatus;
			});
		}
		ps = ordered(ps, s);
		filteredTotal = ps.length;
		const from = s.pageSize * s.pageNo - s.pageSize;
		itemList = ps.slice(from, from + s.pageSize);
		search = s;
	}

	// ?postId= 로 들어오면 그 글 편집 서랍을 연다
	$effect(() => {
		const pid = page.url.searchParams.get('postId');
		if (!pid || !items || !pm) return;
		const it = items.find((o) => o.postId == pid);
		if (it) setTimeout(() => pm?.openDrawer(it), 300);
	});

	const tabs = $derived(
		meta && (meta.items || dynamicSubGroups)
			? [{ value: '', label: '전체', name: '' }, ...(meta.items || dynamicSubGroups || []).map((g) => ({ value: g.groupSubKey, label: g.groupSubKey, name: g.name || '' }))]
			: null
	);
	const pageCount = $derived(Math.ceil((filteredTotal || 1) / search.pageSize));
	const canSort = $derived(!!meta?.useSort && (!postGroup?.items?.length || !!search.groupSubKey || !!postGroup?.feature?.useES) && (itemList?.length || 0) > 1 && pageCount == 1);

	async function saveSort(os: Post[]) {
		const o = await api.post(`/service/admin/post/group/multi-update`, { type: 'sort', postIds: os.map((x) => x.postId) });
		if (!o.ok) return alert('업데이트 도중 오류가 발생하였습니다.');
		for (const p of o.data || []) setPostGroupItem(p.groupKey, p);
		toast('정상적으로 저장되었습니다.');
	}

	function onSaved(p: Post) {
		if (onChangePost) return onChangePost(p);
		setTimeout(() => setPostGroupItem(groupKey, p), postGroup?.feature?.useES ? 1000 : 0);
	}

	function reply(o: Post) {
		const parentPostIds = (o?.json?.parentPostIds?.length || 0) > 0 ? [...o.json.parentPostIds!, o.postId] : [o.postId];
		const d = postDefault();
		pm?.openDrawer({ ...d, groupKey: postGroup.groupKey, groupSubKey: o.groupSubKey || '', json: { ...d.json, parentPostIds }, title: `RE: ${o.title}`, postId: Date.now().toString(36), isNew: true });
	}
</script>

<PageContainer title={postGroup?.partnerName ? `[${postGroup.partnerName}] ${postGroup?.name} ` : ''}>
	<div class={cx('MuiStack-root', toClass({ justifyContent: 'space-between' }))}>
		{#if tabs}
			<ScrollTabs items={tabs} value={search.groupSubKey || ''} onchange={(v) => getData({ ...search, groupSubKey: v })}>
				{#snippet label(t)}
					{#if t.value === ''}{t.label}{:else}<Tooltip title={t.name || ''} placement="top"><h6 class="MuiTypography-root MuiTypography-subtitle2">{t.label}</h6></Tooltip>{/if}
				{/snippet}
			</ScrollTabs>
		{/if}
	</div>

	<BoardSearchForm
		{useES}
		isUserPost={postGroup?.dataColumn == 'jsonUser'}
		{search}
		{getData}
		sx={{ marginBottom: 8 }}
		isPoll={postGroup?.component?.name == 'POLL' || !!postGroup?.feature?.useSchedule}
	/>

	<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16 }))}>
		<BoardListTable {itemList} {postGroup} {masterGroupIds} {groupKey} {useDetail} {path} isUserPost={postGroup?.dataColumn == 'jsonUser'} onmanage={(o) => pm?.openDrawer(o)} onreply={reply} />
	</div>

	{#if itemList}
		{#if canSort}
			<TablePagination sx={{ marginTop: 16 }} {search} getData={(s) => getData({ ...search, ...s })} {pageCount}>
				<SortDialog objectKey="postId" okText="저장" items={[...(itemList || [])].sort((a, b) => (a.json?.sortNo ?? 0) - (b.json?.sortNo ?? 0))} onChange={(os: any[]) => saveSort(os)} />
			</TablePagination>
		{:else}
			<TablePagination sx={{ marginTop: 16 }} {search} getData={(s) => getData({ ...search, ...s })} {pageCount} />
		{/if}
	{/if}

	{#if meta && activeNav()?.auth == 'RW'}
		<PostManageDrawer
			bind:this={pm}
			useCreateButton
			buttonText={groupKey == 'MAIL-TEMPLATE' ? '메일 작성' : postGroup?.feature?.writeButtonText || buttonText}
			{groupKey}
			groupSubKey={search.groupSubKey}
			subGroups={meta?.items || null}
			onSave={onSaved}
			{customPostGroup}
			statusMap={meta?.statusMap || undefined}
			{masterGroupIds}
		/>
	{/if}
</PageContainer>
