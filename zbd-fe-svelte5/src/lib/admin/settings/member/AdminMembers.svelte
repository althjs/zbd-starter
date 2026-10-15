<script lang="ts">
	/**
	 * 관리자 회원 관리(React pages/admin/setting/member) — 그룹 탭·검색·표(웹메일 칸)·쪽 나누기 · 조직도 · 「관리자 추가」.
	 * 관리자(운영자)는 수십 명이라 전체를 받아 화면이 거른다(LIST_AUDIT 3-1).
	 * 조직도 고치기 = 이 메뉴의 쓰기 권한(RW). 로컬 개발 서버 로그인 주소 복사(개발자) = GET /service/admin/member/token.
	 * (여러 호스팅을 보는 isMaster 의 호스팅 탭·사이트빌더 접속은 즈비드에 없어 옮기지 않았다)
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Card from '../../ui/Card.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { getMeta, isSuperAdmin } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { ensureAdminGroups } from '../adminGroups.svelte.ts';
	import MemberListTable from './MemberListTable.svelte';
	import MemberManageDrawer from './MemberManageDrawer.svelte';
	import MemberSearchForm from './MemberSearchForm.svelte';
	import { adminMembers, getAdminSiteMeta, loadAdminMembers, setAdminMembers } from './adminMembers.svelte.ts';

	const profile = admin.profile!;
	const hostingId = profile?.hostingId || '';
	const masterGroupIds: number[] = getMeta('COMMON', 'masterGroupIds') || [];
	const isSuperAuth = isSuperAdmin(profile);
	let search = $state<any>({ hostingId, groupId: '', keyword: '', pageNo: 1, pageSize: 10, orderBy: 'adminId', orderByDesc: 'desc' });
	let memberList = $state<any[] | null>(null);
	let totalCount = $state(0);
	let menus = $state<any[] | null>(null);
	let drawer = $state<any>(null);

	const canEditOrg = $derived(!!(admin.menus || []).flatMap((m: any) => [m, ...(m.items || [])]).find((m: any) => m.path == '/admin/setting/member' && m.auth == 'RW'));

	function getData(s: any) {
		let ms = (adminMembers.list || []).filter((o) => o.hostingId == s.hostingId && !((masterGroupIds || []).indexOf(o.groupId) !== -1 && !isSuperAuth));
		if (s.groupId) ms = ms.filter((o) => o.groupId == s.groupId);
		totalCount = ms.length;
		const k = String(s.keyword || '').trim();
		if (k)
			ms = ms.filter(
				(o) =>
					o.adminId == k ||
					(o.name || '').indexOf(k) !== -1 ||
					(o.email || '').indexOf(k) !== -1 ||
					(o.mobile || '').indexOf(k) !== -1 ||
					(o?.json?.companyName || '').indexOf(k) !== -1 ||
					(getAdminSiteMeta(o?.json)?.companyName || '').indexOf(k) !== -1
			);
		const from = s.pageSize * s.pageNo - s.pageSize;
		search = s;
		memberList = ms.slice(from, from + s.pageSize);
	}

	onMount(async () => {
		ensureAdminGroups();
		const [, mo] = await Promise.all([loadAdminMembers(), api.get('/service/admin/menus')]);
		if (!mo.ok) alert('메뉴정보 로딩 도중 오류가 발생하였습니다.');
		else menus = mo.data;
		getData({ ...search });
	});

	async function doCopyDevLogin(m: any) {
		const o = await api.get(`/service/admin/member/token?adminId=${m.adminId}`);
		if (!o.ok) return alert('로그인 주소 생성 도중 오류가 발생하였습니다.');
		// 개발 서버(vite, 4503)의 개발 로그인이 # 뒤 토큰을 받아 그 관리자로 들어간다 — # 뒤는 서버로 가지 않는다
		const url = `http://localhost:4503/dev/login#admin=${encodeURIComponent(o.data.token)}`;
		try {
			await navigator.clipboard.writeText(url);
			toast(`${m.name} 로컬 로그인 주소가 복사되었습니다.`);
		} catch {
			alert(url);
		}
	}
	function replace(list: any[]) {
		setAdminMembers(list);
		getData(search);
	}
</script>

{#if memberList}
	<MemberSearchForm sx={{ marginBottom: 8 }} {search} {getData} />
	<Card sx={{ padding: 16 }}>
		<MemberListTable {memberList} {doCopyDevLogin} onManage={(m) => (drawer = m)}>
			{#snippet emailExtra(o)}

			{/snippet}
		</MemberListTable>
		<TablePagination sx={{ marginTop: 16 }} {search} getData={(s) => getData(s)} pageCount={Math.ceil(totalCount / search.pageSize)} />
	</Card>
{/if}

<div class={cx('MuiBox-root', toClass({ marginTop: 16 }))}>
	<AuthButton variant="contained" onclick={() => (drawer = { adminId: 0 })}>관리자 추가</AuthButton>
</div>

<MemberManageDrawer
	member={drawer}
	{hostingId}
	{menus}
	onclose={() => (drawer = null)}
	onUpdated={(m) => replace((adminMembers.list || []).flatMap((o) => (o.adminId == m.adminId ? (m.status >= 0 ? [{ ...o, ...m }] : []) : [o])))}
	onAdded={(m) => replace([{ ...m, json: typeof m.json == 'string' ? JSON.parse(m.json || '{}') : m.json }, ...(adminMembers.list || [])])}
/>
