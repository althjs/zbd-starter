<script lang="ts">
	/**
	 * 접속통계(React pages/admin/login-devices) — 로그인 기기 기록. 웹회원(U)·관리자(A) 탭, 기간, 쪽 나누기.
	 * GET /service/admin/member/login-devices?{search} → { totalCount, uniqUserIdCount, items }
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { date, number } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Card from '../../ui/Card.svelte';
	import Tabs from '../../ui/Tabs.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import Loading from '../../layout/Loading.svelte';
	import DateRangeForm from '../ui/DateRangeForm.svelte';
	import { parseBrowserUID, socialUserGroupIdMap } from '../common.ts';
	import { getYesterdayYMD, getYMD } from '../ymd.ts';

	type Search = { pageNo: number; pageSize: number; orderBy: string; orderByDesc: string; keyword: string; tab: 'U' | 'A'; from: string; to: string };
	let search = $state<Search>({ pageNo: 1, pageSize: 10, orderBy: 'lastLoginedAt', orderByDesc: 'desc', keyword: '', tab: 'U', from: getYesterdayYMD(null, 30), to: getYMD() });
	let items = $state<any[]>([]);
	let totalCount = $state(0);
	let uniq = $state(0);
	let loading = $state(false);
	let error = $state('');

	async function getData(s: Search) {
		search = s;
		loading = true;
		error = '';
		const q = new URLSearchParams(s as any).toString();
		const o = await api.get(`/service/admin/member/login-devices?${q}`);
		loading = false;
		if (!o.ok) {
			error = o.error.message || 'API Error';
			items = [];
			totalCount = 0;
			uniq = 0;
			return;
		}
		items = (o.data?.items || []).map((x: any) => ({ ...x, env: parseBrowserUID(x.browserUid) }));
		totalCount = o.data?.totalCount || 0;
		uniq = o.data?.uniqUserIdCount || 0;
	}
	onMount(() => getData(search));
	const isAdmin = $derived(search.tab == 'A');
</script>

<Card sx={{ padding: 16 }}>
	<DateRangeForm {search} setSearch={(s) => getData(s)} />
</Card>

<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'end', flexWrap: 'wrap', justifyContent: 'space-between' }))}>
	<Tabs
		items={[
			{ value: 'U', label: '웹회원' },
			{ value: 'A', label: '관리자 회원' }
		]}
		value={search.tab}
		onchange={(tab) => getData({ ...search, tab, pageNo: 1 })}
	/>
	<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ marginBottom: 8 }))}>총 로그인 세션: {number(totalCount)}, 유저 수: {number(uniq)}</span>
</div>

<Card sx={{ padding: 16 }}>
	{#if loading}
		<div class={toClass({ textAlign: 'center', paddingTop: 32, paddingBottom: 32 })}><Loading full={false} /></div>
	{:else if error}
		<div class={toClass({ color: 'red', textAlign: 'center', paddingTop: 32, paddingBottom: 32 })}>{error}</div>
	{:else if !items.length}
		<div class={toClass({ textAlign: 'center', paddingTop: 32, paddingBottom: 32 })}>데이터 없음</div>
	{:else}
		<table class="MuiTable-root">
			<thead class="MuiTableHead-root">
				<tr class="MuiTableRow-root MuiTableRow-head">
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">ID</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">이메일</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">이름</th>
					{#if !isAdmin}
						<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">닉네임</th>
						<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall MuiTableCell-alignCenter">가입방식</th>
					{/if}
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">IP</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">브라우저</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">접속 기기</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">접속일</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">마지막 활동일</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeSmall">BrowserUID</th>
				</tr>
			</thead>
			<tbody class="MuiTableBody-root">
				{#each items as o, i (i)}
					<tr class="MuiTableRow-root">
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.userId}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.email}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.name}</td>
						{#if !isAdmin}
							<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.nickname}</td>
							<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall MuiTableCell-alignCenter">{socialUserGroupIdMap[o.groupId] ?? ''}</td>
						{/if}
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.ip || '-'}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.env?.browser || '-'}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.env?.os || '-'}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{date(o.lastLoginedAt)}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.lastActivedAt ? date(o.lastActivedAt) : '-'}</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeSmall">{o.browserUid || '-'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</Card>

<div class={cx('MuiStack-root', toClass({ marginTop: 16 }))}>
	<TablePagination {search} getData={(s) => getData(s)} pageCount={Math.ceil(totalCount / search.pageSize)} />
</div>
