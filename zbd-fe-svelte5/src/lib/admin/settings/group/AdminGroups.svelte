<script lang="ts">
	/**
	 * 관리자 그룹 관리(React pages/admin/setting/group) — 그룹 표(구성원 수 배지)·관리(톱니 → 편집 서랍)·권한(메뉴 권한 서랍)·「그룹 추가」.
	 * GET /service/admin/group · /service/admin/menus · 구성원 수는 관리자 회원 목록에서.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import { shortDate } from '../../format.ts';
	import { getMeta, isSuperAdmin } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import AuthSettingDrawer from '../member/AuthSettingDrawer.svelte';
	import { adminMembers, loadAdminMembers } from '../member/adminMembers.svelte.ts';
	import GroupManageDrawer from './GroupManageDrawer.svelte';

	const hostingId = admin.profile?.hostingId || '';
	const masterGroupIds: number[] = getMeta('COMMON', 'masterGroupIds') || [];
	const isSuperAuth = isSuperAdmin(admin.profile);
	let groups = $state<any[] | null>(null);
	let menus = $state<any[] | null>(null);
	let edit = $state<any>(null);
	let auth = $state<any>(null);
	const groupList = $derived((groups || []).filter((o) => o.hostingId == hostingId && !((masterGroupIds || []).indexOf(o.groupId) !== -1 && !isSuperAuth)));
	const count = (id: number) => (adminMembers.list || []).filter((m) => m.groupId == id).length;

	onMount(async () => {
		if (!adminMembers.list) loadAdminMembers();
		const [go, mo] = await Promise.all([api.get('/service/admin/group'), api.get('/service/admin/menus')]);
		if (!go.ok) alert('데이터 로딩 도중 오류가 발생하였습니다.');
		else groups = go.data;
		if (!mo.ok) alert('메뉴정보 로딩 도중 오류가 발생하였습니다.');
		else menus = mo.data;
	});
	const th = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const td = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
</script>

<Card sx={{ padding: 16, marginBottom: 16 }}>
	<table class={cx('MuiTable-root', toClass({ minWidth: 650 }))}>
		<thead class="MuiTableHead-root">
			<tr class="MuiTableRow-root MuiTableRow-head">
				<th class={th} style="width:80px">그룹ID</th>
				<th class={th}>이름</th>
				<th class={th}>설명</th>
				<th class={th}>등록일</th>
				<th class={th}>수정일</th>
				<th class="{th} MuiTableCell-alignCenter" style="width:60px">관리</th>
				<th class="{th} MuiTableCell-alignCenter" style="width:60px">권한</th>
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each groupList as o (o.groupId)}
				{@const n = count(o.groupId)}
				<tr class={cx('MuiTableRow-root', toClass({ '&:last-child td, &:last-child th': { border: 0 } }))}>
					<td class={td} style="width:50px">{o.groupId}</td>
					<th class={td} scope="row">
						<span class="MuiBadge-root">
							<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ marginRight: 8 }))}>{o.name}</p>
							<span
								class={cx(
									'MuiBadge-badge MuiBadge-standard MuiBadge-anchorOriginTopRight MuiBadge-anchorOriginTopRightRectangular MuiBadge-overlapRectangular MuiBadge-colorPrimary',
									!n && 'MuiBadge-invisible'
								)}>{n || ''}</span
							>
						</span>
					</th>
					<td class={td}>{o.description ?? ''}</td>
					<td class={td}>{shortDate(o.createdAt)}</td>
					<td class={td}>{shortDate(o.updatedAt) || '-'}</td>
					<td class="{td} MuiTableCell-alignCenter"><IconButton aria-label="관리" onclick={() => (edit = o)}><Icon name="settings" /></IconButton></td>
					<td class="{td} MuiTableCell-alignCenter"><IconButton aria-label="권한" disabled={!menus} onclick={() => (auth = o)}><Icon name="admin_panel_settings" /></IconButton></td>
				</tr>
			{/each}
		</tbody>
	</table>
	{#if groups && groupList.length == 0}
		<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 40, marginBottom: 40 }))}>등록된 그룹이 없습니다.</p>
	{/if}
</Card>

<AuthButton variant="contained" onclick={() => (edit = { groupId: 0 })}>그룹 추가</AuthButton>

<GroupManageDrawer
	group={edit}
	{hostingId}
	onclose={() => (edit = null)}
	onUpdated={(m) => (groups = (groups || []).map((o) => (o.groupId == m.groupId ? { ...o, ...m } : o)))}
	onAdded={(m) => (groups = [...(groups || []), m])}
	onDelete={(id) => (groups = (groups || []).filter((o) => o.groupId != id))}
/>
<AuthSettingDrawer group={auth} {menus} {hostingId} onclose={() => (auth = null)} />
