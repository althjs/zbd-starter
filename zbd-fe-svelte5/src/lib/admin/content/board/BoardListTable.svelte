<script lang="ts">
	/**
	 * 게시판 목록 표(React board/components/BoardListTable) — 게시판 설정에 따라 칸이 늘고 준다.
	 * 답글(replies) 은 원글 아래 들여 쓴 줄로. 「관리」(⚙) = onmanage, 「답글」 = onreply.
	 */
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { page } from '$app/state';
	import { number } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { shortDate } from '../../format.ts';
	import { getHostingItems } from '../../meta.svelte.ts';
	import { activeNav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import StateChip from '../../ui/StateChip.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import { date } from '#lib/shared/format.ts';
	import { loadAdminMembers, loadAllUsers, members } from '../members.svelte.ts';
	import { getPostAuthorInfo, getPostGroup, statusMapDefault } from '../post/postGroup.ts';
	import EsMarkTitle from '../ui/EsMarkTitle.svelte';
	import { bodyCellFix } from '../ui/table.ts';
	import type { Post, PostGroup } from '#lib/shared/types/post.ts';

	let {
		itemList,
		postGroup = null,
		useDetail = false,
		path = '',
		masterGroupIds = null,
		groupKey = '',
		isUserPost = false,
		onmanage,
		onreply,
		onestimate
	}: {
		itemList: Post[] | null;
		postGroup?: PostGroup | null;
		useDetail?: boolean;
		path?: string;
		masterGroupIds?: number[] | null;
		groupKey?: string;
		isUserPost?: boolean;
		onmanage?: (p: Post) => void;
		onreply?: (p: Post) => void;
		onestimate?: (p: Post) => void;
	} = $props();

	const profile = $derived(admin.profile);
	const auth = $derived(activeNav()?.auth);
	const isPartnerInfo = $derived(postGroup?.component?.name == 'PARTNER');
	const isPoll = $derived(postGroup?.component?.name == 'POLL' || !!postGroup?.feature?.useSchedule);
	const useSubGroup = $derived(!postGroup || (postGroup?.items?.length || 0) > 0);
	const isMaster = $derived(!masterGroupIds || masterGroupIds.indexOf(profile?.groupId as number) !== -1);
	const showInquiryHosting = $derived(groupKey == 'INQUIRY' && !!masterGroupIds && masterGroupIds.indexOf(profile?.groupId as number) !== -1);
	// 호스팅 이름(HOSTINGS.items) — 예전엔 설정 객체를 맵처럼 읽어 이름이 늘 비었다
	const hostingNames = $derived(Object.fromEntries(getHostingItems().map((o: any) => [o.hostingId || '', o.name])) as Record<string, string>);
	const ts = Date.now();

	$effect(() => {
		if (isUserPost) loadAllUsers();
		else loadAdminMembers();
	});

	const gray = toClass({ color: 'var(--gt-700, #909090)' });
	const stat = toClass({ flexDirection: 'row', gap: 8, '& .MuiTypography-root': { fontSize: 12, color: 'var(--gt-700, #909090)' } });
	const unmasked = toClass({ display: 'inline-flex', alignItems: 'center', gap: '2px', cursor: 'help', borderBottom: '1px dashed var(--gt-500, #c4c4c4)', lineHeight: 1.2 });
	const p0 = toClass({ padding: '0 !important' });
	const cell = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
	const head = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
</script>

{#snippet row(o: Post)}
	{@const pg = page.url.pathname.indexOf('/admin/cs') == -1 ? getPostGroup(o.groupKey) : postGroup}
	{@const useEditorPage = !!pg?.feature?.useEditerPage}
	{@const r = o?.json?.reaction || {}}
	{@const pv = o?.json?.viewCount || 0}
	{@const cc = o?.json?.commentCount || 0}
	{@const adm = members.admins?.find((x) => x.adminId == o.adminId)}
	{@const from = o.json?.schedule?.from ? new Date(o.json.schedule.from).getTime() : null}
	{@const to = o.json?.schedule?.to ? new Date(o.json.schedule.to).getTime() : null}
	{@const scheduleStatus = from && to ? (from < ts && to > ts ? '진행중' : ts < to ? '대기중' : '종료') : '미정'}
	{@const depth = o?.json?.parentPostIds?.length || 0}
	<tr class="MuiTableRow-root">
		<td class={cell} style="width:50px">
			{#if profile?.isDeveloper || useEditorPage}
				<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover" href={`/admin/board/${o.groupKey}/editor/${o.postId}`} title={`포스트 : ${o.title}`}>{o.postId}</a>
			{:else}{o.postId}{/if}
		</td>
		<th class={cell} scope="row">
			{#if useDetail}
				<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineAlways" href={`${path}/${o.postId}`} title={`포스트 : ${o.title}`}>
					{#if o.json?.isFixTop}<StateChip sx={{ marginRight: 4 }} label="상단고정" size="small" />{/if}
					<EsMarkTitle title={o.title} />
				</a>
			{:else}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
					{#if o.jsonUser?.isBest}<StateChip sx={{ marginRight: 4 }} label="베스트" size="small" />{/if}
					{#if o.json?.isFixTop}<StateChip sx={{ marginRight: 4 }} label="상단고정" size="small" />{/if}
					<a
						class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ display: 'flex', flexDirection: 'row' }))}
						href={`/admin/board/${o.groupKey}/${o.postId}`}
						title={`포스트 : ${o.title}`}
					>
						{#each o.json?.parentPostIds || [] as ppid, ii (ppid)}
							{#if ii < depth - 1}<Icon name="" size="Small" sx={{ width: 16 }} />{:else}<Icon name="subdirectory_arrow_right" size="Small" sx={{ width: 20 }} />{/if}
						{/each}
						<EsMarkTitle title={o.title} />
					</a>
				</div>
			{/if}
			<div class={cx('MuiStack-root', stat, depth ? toClass({ marginLeft: depth * 16 - 14 }) : '')}>
				{#if pv > 0}<p class="MuiTypography-root MuiTypography-body1">PV: {number(pv)}</p>{/if}
				{#if cc > 0}<p class="MuiTypography-root MuiTypography-body1">댓글: {number(cc)}</p>{/if}
				{#if r.L > 0}<p class="MuiTypography-root MuiTypography-body1">좋아요: {number(r.L)}</p>{/if}
				{#if r.R > 0}<p class="MuiTypography-root MuiTypography-body1">별점: {number(r.R)}건 ({r.RA || 0}점)</p>{/if}
				{#if r.A > 0}<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ color: 'var(--adm-error-main) !important' }))}>신고수: {number(r.A)}</p>{/if}
				{#if r.AC > 0}<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ color: 'var(--adm-secondary-main) !important' }))}>신고수(댓글): {number(r.AC)}</p>{/if}
			</div>
		</th>
		<td class={cell}>{pg?.name || o.groupKey}</td>
		{#if useSubGroup}<td class={cell}>{o.groupSubKey}</td>{/if}
		<td class={cell}>
			{#if isUserPost}
				{@const au = getPostAuthorInfo(o, members.userMap)}
				{#if !au.unmasked}<EsMarkTitle title={au.name} />{:else}
					<Tooltip title={`표시이름: ${au.displayName}`} placement="top"
						><span class={unmasked}><Icon name="person" sx={{ fontSize: 13, color: 'var(--gt-700, #909090)' }} /><EsMarkTitle title={au.name} /></span></Tooltip
					>
				{/if}
				<br /><span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{members.userMap?.[o.userId as number]?.email || '-'}</span>
			{:else}
				{adm?.name || o.json?.name || o.adminId || ''}
				<br /><span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{adm?.email || '-'}</span>
			{/if}
		</td>
		{#if isPartnerInfo}
			<td class={cell}>{o.json?.ceoName || ''}<br /><span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{`${o.json?.text1 || ''} ${o.json?.text2 || ''}`}</span></td>
		{/if}
		{#if isPoll}
			<td class={cell}>{date(from)}<br /> <span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{date(to)}</span></td>
			<td class={cx(cell, 'MuiTableCell-alignCenter')}><StateChip label={scheduleStatus} size="small" isAutoColor /></td>
		{/if}
		{#if auth == 'RW' && pg?.component?.processMap}
			<td class={cx(cell, 'MuiTableCell-alignCenter')}>{pg.component.processMap[(o as any)?.[pg?.dataColumn || 'json']?.processStatus || 0]}</td>
		{/if}
		{#if !postGroup?.disableOpenStatus}
			<td class={cx(cell, 'MuiTableCell-alignCenter')}>{(pg?.statusMap || statusMapDefault)[o.status]}</td>
		{/if}
		<td class={cell}>
			<div class="MuiStack-root">
				{shortDate(o.createdAt)}
				{#if o?.json?.createdAtReal}<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{date(o.json.createdAtReal)}</span>{/if}
			</div>
		</td>
		<td class={cell}>{shortDate(o.updatedAt) || '-'}</td>
		{#if showInquiryHosting}<td class={cx(cell, 'MuiTableCell-alignCenter')}>{hostingNames[hostingIdOf(o.json) as string] || hostingIdOf(o.json)}</td>{/if}
		{#if auth == 'RW'}
			{#if isMaster}
				<td class={cx(cell, 'MuiTableCell-alignCenter', p0)}><IconButton aria-label="관리" onclick={() => onmanage?.(o)}><Icon name="settings" /></IconButton></td>
			{/if}
			{#if postGroup?.feature?.useReply}
				<td class={cx(cell, 'MuiTableCell-alignCenter', p0)}><Button size="small" sx={{ minWidth: 'unset' }} onclick={() => onreply?.(o)}>답글</Button></td>
			{/if}
		{/if}
		{#if pg?.feature?.useEstimation}
			<td class={cx(cell, 'MuiTableCell-alignCenter')}><Button size="small" onclick={() => onestimate?.(o)}>견적관리</Button></td>
		{/if}
	</tr>
	{#each o.replies || [] as rp (rp.postId)}{@render row(rp)}{/each}
{/snippet}

<table class={cx('MuiTable-root', toClass({ minWidth: 650, ...bodyCellFix, '& tbody > tr:last-child > td, & tbody > tr:last-child > th': { borderBottom: 0 } }))}>
	<thead class="MuiTableHead-root">
		<tr class="MuiTableRow-root MuiTableRow-head">
			<th class={head} scope="col">ID</th>
			<th class={head} scope="col">제목</th>
			<th class={cx(head, toClass({ minWidth: 80 }))} scope="col">그룹</th>
			{#if useSubGroup}<th class={head} scope="col">서브그룹</th>{/if}
			<th class={head} scope="col">작성자 (이메일)</th>
			{#if isPartnerInfo}<th class={head} scope="col">파트너정보</th>{/if}
			{#if isPoll}<th class={head} scope="col">기간</th><th class={cx(head, 'MuiTableCell-alignCenter')} style="width:80px" scope="col">진행상태</th>{/if}
			{#if auth == 'RW' && postGroup?.component?.processMap}<th class={cx(head, 'MuiTableCell-alignCenter', toClass({ minWidth: 90 }))} scope="col">처리상태</th>{/if}
			{#if !postGroup?.disableOpenStatus}<th class={cx(head, 'MuiTableCell-alignCenter')} scope="col">공개상태</th>{/if}
			<th class={head} style="width:150px" scope="col">등록일</th>
			<th class={head} style="width:140px" scope="col">수정일</th>
			{#if showInquiryHosting}<th class={cx(head, 'MuiTableCell-alignCenter')} scope="col">업체</th>{/if}
			{#if auth == 'RW' && isMaster}
				<th class={cx(head, 'MuiTableCell-alignCenter', p0)} style="width:60px" scope="col">관리</th>
				{#if postGroup?.feature?.useReply}<th class={cx(head, 'MuiTableCell-alignCenter', p0)} style="width:60px" scope="col">답글</th>{/if}
			{/if}
			{#if postGroup?.feature?.useEstimation}<th class={cx(head, 'MuiTableCell-alignCenter')} style="width:100px" scope="col">견적관리</th>{/if}
		</tr>
	</thead>
	<tbody class="MuiTableBody-root">
		{#each itemList || [] as o (o.postId)}{@render row(o)}{/each}
	</tbody>
</table>
{#if itemList && itemList.length == 0}
	<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', margin: '40px 0' }))}>등록된 글이 없습니다.</p>
{/if}
