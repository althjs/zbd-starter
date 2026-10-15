<script lang="ts">
	/** 인증 기기(React my/components/LoginDevice) — `GET /service/user/my/device` · 지우기 `DELETE /service/user/my/device/{browserUid}`(지금 기기는 못 지움) */
	import { api, getBrowserUid } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { shortDate } from '../../format.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { bodyCellFix } from '../ui/table.ts';

	const osMap: Record<string, string> = { A: 'Android', W: 'Windows', M: 'Macintosh', I: 'iOS', Z: 'Etc' };
	const brMap: Record<string, string> = { C: 'Chrome', S: 'Safari', E: 'Edge', F: 'Firefox', W: 'Whale', Z: 'Etc' };
	const parse = (u: string) => (u ? { os: osMap[u.charAt(0)] || 'UNKNOWN', browser: brMap[u.charAt(1)] || 'UNKNOWN' } : { os: 'UNKNOWN', browser: 'UNKNOWN' });
	let devices = $state<any[] | null>(null);
	const mine = getBrowserUid();

	async function load() {
		const o = await api.get(`/service/user/my/device`);
		if (!o.ok) return alert('데이터 로딩 도중 오류가 발생하였습니다.');
		devices = o.data;
	}
	$effect(() => {
		load();
	});
	async function del(uid: string) {
		if (!(await confirm('인증기기를 삭제하시겠습니까?'))) return;
		const o = await api.delete(`/service/user/my/device/${uid}`);
		if (!o.ok) return alert('로그아웃 처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 제거되었습니다.');
		load();
	}
	const head = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const cell = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ marginTop: 24, padding: 24 }))}>
	<div class={cx('MuiTypography-root MuiTypography-h4', toClass({ display: 'flex', justifyItems: 'center', marginBottom: 16, flexDirection: 'column' }))}>
		<div class={cx('MuiStack-root', toClass({ marginBottom: 4 }))}>인증 기기</div>
		<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ color: '#999' }))}>인증기기로 등록되지 않은 경우 로그아웃 처리됩니다.</h6>
	</div>
	<table class={cx('MuiTable-root', toClass({ ...bodyCellFix, '& tbody > tr:last-child > td, & tbody > tr:last-child > th': { border: 0 } }))} aria-label="simple table">
		<thead class="MuiTableHead-root"
			><tr class="MuiTableRow-root MuiTableRow-head"
				><th class={head}>OS</th><th class={head}>브라우저</th><th class={head}>마지막 로그인</th><th class={cx(head, 'MuiTableCell-alignCenter')} style="width:80px">삭제</th></tr
			></thead
		>
		<tbody class="MuiTableBody-root">
			{#each devices || [] as o (o.browserUid)}
				{@const u = parse(o.browserUid)}
				<tr class="MuiTableRow-root">
					<th class={cell} scope="row">
						{u.os}
						{#if mine == o.browserUid}<div class="MuiStack-root"><span class={cx('MuiTypography-root MuiTypography-overline', toClass({ color: 'var(--adm-error-main)' }))}>현재기기</span></div>{/if}
					</th>
					<td class={cell}>{u.browser}</td>
					<td class={cell}>{shortDate(o.lastLoginedAt || o.createdAt)}</td>
					<td class={cx(cell, 'MuiTableCell-alignCenter')}><IconButton aria-label="삭제" disabled={mine == o.browserUid} onclick={() => del(o.browserUid)}><Icon name="auto_delete" /></IconButton></td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
