<script lang="ts">
	/**
	 * 사용자 회원 관리(React pages/admin/user/member) — 회원 전체를 한 번 받아(GET /service/admin/user?isIncludeJson=1) 화면에서 거르고 자른다.
	 * 톱니 → 편집 서랍(UserMemberDrawer). 즈비드(ZBD)는 아래에 미니홈피 차단 목록·플러그인 구독.
	 * (호스팅 고르기 탭은 여러 호스팅을 보는 isMaster 계정만 — 즈비드엔 없다)
	 * v1 목록 대기: 지금은 회원 전체를 받아 화면이 거른다(CLAUDE.md 4) — Go GET /api/v1/admin/members(LIST_AUDIT 1-1) 가 생기면 page·size·q·status 를 서버로.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { number } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Card from '../../ui/Card.svelte';
	import Tabs from '../../ui/Tabs.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { getMeta } from '../../meta.svelte.ts';
	import MemberSearchForm from '../member/MemberSearchForm.svelte';
	import { filterUserMember, orderBy } from './filter.ts';
	import MemberListTable from './MemberListTable.svelte';
	import UserMemberDrawer from './UserMemberDrawer.svelte';

	const hostingId = admin.profile?.hostingId || '';
	let search = $state<any>({ hostingId, keyword: '', pageNo: 1, pageSize: 10, orderBy: 'userId', orderByDesc: 'desc', type: '', agreeMap: {}, authGroupId: '' });
	let users = $state<any[] | null>(null);
	let userList = $state<any[] | null>(null);
	let totalCount = $state(0);
	let filteredTotal = $state(0);
	let drawerUser = $state<any>(null);
	const userTypeMap = getMeta('SITE')?.userTypeMap;

	function getData(s: any) {
		let us = filterUserMember(
			s,
			(users || []).filter((o) => o.hostingId == s.hostingId)
		);
		filteredTotal = us.length;
		const from = s.pageSize * s.pageNo - s.pageSize;
		search = s;
		userList = us.slice(from, from + s.pageSize);
	}

	onMount(async () => {
		const params = [];
		if (admin.profile?.isMaster) params.push(`hostingId=${hostingId}`);
		params.push('isIncludeJson=1');
		const o = await api.get(`/service/admin/user?${params.join('&')}`);
		if (!o.ok) return alert('회원정보 로딩 도중 오류가 발생하였습니다.');
		const list = (o.data || []).map((x: any) => {
			if (typeof x.json == 'string') {
				x.Json = JSON.parse(x.json);
				delete x.json;
			}
			return x;
		});
		totalCount = list.length;
		users = orderBy(list, search.orderBy, search.orderByDesc);
		getData({ ...search, hostingId });
	});

	async function openDrawer(u: any) {
		const o = await api.get(`/service/admin/user/${u.userId}`);
		if (!o.ok) return alert('회원정보 로딩 도중 오류가 발생하였습니다.');
		const oo = o.data;
		oo.Json = JSON.parse(oo.json || '{}');
		delete oo.json;
		oo.Json.imageId = `u_${oo.userId}_my`;
		drawerUser = oo;
	}
	function replaceUsers(fn: (arr: any[]) => any[]) {
		users = fn(users || []);
		getData(search);
	}
</script>

<MemberSearchForm sx={{ marginBottom: 8 }} {search} {getData} isAdmin={false} />

{#if userTypeMap}
	<div class="MuiStack-root">
		<Tabs
			items={[{ value: '', label: '전체' }, ...Object.keys(userTypeMap).map((t) => ({ value: t || 'GENERAL', label: userTypeMap[t] }))]}
			value={search.type || ''}
			onchange={(t) => getData({ ...search, type: t })}
		/>
	</div>
{/if}

{#if totalCount > 0}
	<span class="MuiTypography-root MuiTypography-caption">총 회원수: {number(totalCount)}{totalCount != filteredTotal ? `, 검색 회원 수: ${number(filteredTotal)}` : ''}</span>
{/if}

<Card sx={{ padding: 16 }}>
	{#if userList}
		<MemberListTable {userList} onManage={openDrawer} />
		{#if filteredTotal > 0}
			<TablePagination sx={{ marginTop: 16 }} {search} getData={(s) => getData(s)} pageCount={Math.ceil(filteredTotal / search.pageSize)} />
		{/if}
	{:else}
		<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 40, marginBottom: 40 }))}>로딩중...</p>
	{/if}
</Card>

<!-- 미니홈피 플러그인이 이 빈 칸을 「주소 예약어」 칸으로 바꾼다 -->
<!-- svelte-ignore block_empty -->
{#if hostingId == 'ZBD'}

{/if}

<UserMemberDrawer
	user={drawerUser}
	onclose={() => (drawerUser = null)}
	onDelete={(userId, isUnDelete) => {
		replaceUsers((arr) => arr.flatMap((o) => (o.userId == userId ? (isUnDelete === -1000 ? [] : [{ ...o, status: isUnDelete ? 0 : -1 }]) : [o])));
		drawerUser = null;
	}}
	onSaved={(u) => replaceUsers((arr) => arr.map((o) => (o.userId == u.userId ? u : o)))}
/>
