<script lang="ts">
	/**
	 * 메타 기본정보 탭(React setting/meta/components/MetaKeyTable) — 이 호스팅 메타 목록(metaKey·이름·설명·수정일) · metaKey 를 누르면 RAW 편집 ·
	 * 「추가하기」(RW) = 새 메타키를 물어 RAW 편집으로. 즈비드 한 호스팅이라 호스팅 칸(isMaster)은 없다.
	 */
	import { goto } from '$app/navigation';
	import { date } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import { prompt } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';

	let { meta, hostingId, target = 'setting/meta' }: { meta: any[] | null; hostingId: string; target?: string } = $props();
	const list = $derived((meta || []).filter((o) => o.hostingId == (admin.profile?.hostingId || hostingId)));
	const th = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const td = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
	const editUrl = (cid: string, key: string) => `/admin/${target}/setting/${cid ? cid : 'edit'}/${key}`;
</script>

<table class="MuiTable-root">
	<thead class="MuiTableHead-root">
		<tr class="MuiTableRow-root MuiTableRow-head">
			<th class={th} style="width:150px">metaKey</th>
			<th class={th}>이름</th>
			<th class={th}>설명</th>
			<th class="{th} MuiTableCell-alignCenter" style="width:200px">수정일</th>
		</tr>
	</thead>
	<tbody class="MuiTableBody-root">
		{#each list as o (o.metaKey)}
			<tr class="MuiTableRow-root">
				<td class={td}><a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover" href={editUrl(o.hostingId, o.metaKey)}>{o.metaKey}</a></td>
				<td class={td}>{o.name}</td>
				<td class={td}>{o.memo || ''}</td>
				<td class="{td} MuiTableCell-alignCenter">{date(o.updatedAt) || '-'}</td>
			</tr>
		{/each}
	</tbody>
</table>
{#if meta && list.length == 0}<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', margin: '40px 0' }))}>등록된 정보가 없습니다.</p>{/if}
<div class={cx('MuiStack-root', toClass({ marginTop: 8 }))}>
	<AuthButton
		variant="outlined"
		sx={{ width: 'fit-content' }}
		onclick={async () => {
			const key = await prompt({ title: '신규 메타키를 추가합니다.', body: '추가할 메타키를 영문으로 입력하세요' });
			if (!key) return;
			goto(editUrl(hostingId, key));
		}}>추가하기</AuthButton
	>
</div>
