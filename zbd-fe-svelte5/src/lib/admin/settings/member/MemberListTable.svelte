<script lang="ts">
	/**
	 * 관리자 회원 표(React setting/member/components/MemberListTable) — No.·이름·이메일(+ 웹메일 칸)·휴대폰·그룹·등록일·수정일·로그인(DEV)·관리.
	 * 고르기(selected) 모드는 이메일·문자 발송 대상 고르기에서 쓴다(그때는 등록일·관리 칸 없음).
	 * (사이트빌더 「홈피」·「접속」 칸은 여러 호스팅을 보는 isMaster 계정만 — 즈비드엔 없어 옮기지 않았다)
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import { shortDate } from '../../format.ts';
	import { admin } from '../../profile.svelte.ts';
	import { adminGroups } from '../adminGroups.svelte.ts';
	import { socialUserGroupIdMap } from '../common.ts';
	import { getAdminSiteMeta } from './adminMembers.svelte.ts';

	let {
		memberList,
		onManage = null,
		doCopyDevLogin = null,
		selected = null,
		onSelect = null,
		selectTarget = '',
		isAdmin = true,
		memberType = 'ADMIN',
		emailExtra
	}: {
		memberList: any[];
		onManage?: ((m: any) => void) | null;
		doCopyDevLogin?: ((m: any) => void) | null;
		selected?: number[] | null;
		onSelect?: ((ids: number[]) => void) | null;
		selectTarget?: string;
		isAdmin?: boolean;
		memberType?: string;
		emailExtra?: Snippet<[any]>;
	} = $props();

	const idOf = (o: any) => (isAdmin ? o.adminId : o.userId);
	const isValidEmail = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(v || '').trim());
	const isValidMobile = (v: string) => /^01[0-9]-?\d{3,4}-?\d{4}$/.test(String(v || '').trim());
	const valid = (o: any) => (selectTarget == 'EMAIL' ? isValidEmail(o.email) : selectTarget == 'MOBILE_NO' ? isValidMobile(o.mobile) : false);
	const th = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const td = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
	const gray = toClass({ color: 'var(--gt-700, #909090)' });
</script>

<table
	class={cx(
		'MuiTable-root',
		toClass({
			minWidth: 650,
			'--adm-td-px': '8px',
			'& th.MuiTableCell-root, & td.MuiTableCell-root': { paddingLeft: 8, paddingRight: 8, paddingTop: 4, paddingBottom: 4 },
			'& .MuiCheckbox-root': { padding: 0 }
		})
	)}
	aria-label="simple table"
>
	<thead class="MuiTableHead-root">
		<tr class="MuiTableRow-root MuiTableRow-head">
			{#if selected}
				<th class={th} style="width:50px">
					<input
						type="checkbox"
						onchange={(e) => {
							const tf = (e.currentTarget as HTMLInputElement).checked;
							let ids = [...selected!];
							if (tf) for (const o of memberList) valid(o) && ids.indexOf(idOf(o)) == -1 && ids.push(idOf(o));
							else ids = ids.filter((id) => !memberList.find((x) => idOf(x) == id));
							onSelect?.(ids);
						}}
					/>
				</th>
			{/if}
			<th class={th}>No.</th>
			{#if memberType == 'LAND' || memberType == 'AGENCY'}<th class={th}>회사명</th>{/if}
			<th class={th} style="width:120px">이름</th>
			{#if !isAdmin}<th class="{th} MuiTableCell-alignCenter">가입방식</th>{/if}
			<th class={th}>이메일</th>
			<th class={th} style="width:150px">휴대폰번호</th>
			{#if isAdmin && memberType == 'ADMIN'}<th class={th}>그룹</th>{/if}
			{#if !selected}<th class={th}>등록일</th><th class={th}>수정일</th>{/if}
			{#if admin.profile?.isDeveloper && doCopyDevLogin}<th class="{th} MuiTableCell-alignCenter" style="width:60px">로그인<br />(DEV)</th>{/if}
			{#if !selected && onManage}<th class="{th} MuiTableCell-alignCenter" style="width:60px">관리</th>{/if}
		</tr>
	</thead>
	<tbody class="MuiTableBody-root">
		{#each memberList as o (idOf(o))}
			{@const site = getAdminSiteMeta(o.json)}
			<tr class={cx('MuiTableRow-root', toClass({ '&:last-child td, &:last-child th': { border: 0 } }))}>
				{#if selected}
					<td class={td}>
						<input
							type="checkbox"
							disabled={!valid(o)}
							checked={selected.indexOf(idOf(o)) !== -1}
							onchange={(e) => {
								const ids = selected!.filter((id) => id != idOf(o));
								if ((e.currentTarget as HTMLInputElement).checked) ids.push(idOf(o));
								onSelect?.(ids);
							}}
						/>
					</td>
				{/if}
				<td class={td} style="width:50px">{idOf(o)}</td>
				{#if memberType == 'LAND' || memberType == 'AGENCY'}<td class={td}>{o.companyName ?? ''}</td>{/if}
				<th class={td} scope="row">
					<div class="MuiStack-root">
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>{o.name}</div>
						{#if site?.companyName}<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{site.companyName}</span>{/if}
					</div>
				</th>
				{#if !isAdmin}<td class="{td} MuiTableCell-alignCenter">{socialUserGroupIdMap[o.groupId] ?? ''}</td>{/if}
				<td class={td}>
					{o.email}
					{#if emailExtra}{@render emailExtra(o)}{/if}
					{#if o.json?.nickname}<div class="MuiStack-root"><span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{o.json.nickname}</span></div>{/if}
				</td>
				<td class={td}>{o.mobile ?? ''}</td>
				{#if isAdmin && memberType == 'ADMIN'}
					<td class={td}>
						{(adminGroups.list || []).find((g) => g.groupId == o.groupId)?.name || (o._hostingId == 'SITE' ? '사이트 마스터' : '-')}
						{#if o.json?.companyName}<div class="MuiTypography-root MuiTypography-overline">{o.json.companyName}</div>{/if}
					</td>
				{/if}
				{#if !selected}
					<td class={td}>{shortDate(o.createdAt)}</td>
					<td class={td}>{shortDate(o.updatedAt) || '-'}</td>
				{/if}
				{#if admin.profile?.isDeveloper && doCopyDevLogin}
					<td class="{td} MuiTableCell-alignCenter" style="width:60px">
						<Tooltip title="로컬 개발 서버(localhost:4503) 로그인 주소 복사" placement="top"
							><IconButton onclick={() => doCopyDevLogin?.(o)} aria-label="로컬 로그인 주소 복사"><Icon name="content_copy" size="Small" /></IconButton></Tooltip
						>
					</td>
				{/if}
				{#if !selected && onManage}
					<td class="{td} MuiTableCell-alignCenter"><IconButton onclick={() => onManage?.(o)} aria-label="관리"><Icon name="settings" /></IconButton></td>
				{/if}
			</tr>
		{/each}
	</tbody>
</table>
{#if memberList && memberList.length == 0}
	<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 40, marginBottom: 40 }))}>데이터가 없습니다.</p>
{/if}
