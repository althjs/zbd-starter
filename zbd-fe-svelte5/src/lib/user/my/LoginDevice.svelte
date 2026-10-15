<script lang="ts">
	/** 인증 기기 — 로그인한 브라우저 목록·지우기(지금 기기는 못 지움). 예전 my/components/LoginDevice.tsx */
	import { api, getBrowserUid } from '#lib/shared/api.ts';
	import { shortDate } from '#lib/shared/format.ts';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import IconButton from '#lib/shared/mui/IconButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import { alert, confirm, toast } from '#lib/shared/ui/dialog.svelte.ts';

	let { isMobile }: { isMobile: boolean } = $props();
	let devices = $state<any[] | null>(null);
	const me = getBrowserUid();
	const OS: Record<string, string> = { A: 'Android', W: 'Windows', M: 'Macintosh', I: 'iOS', Z: 'Etc' };
	const BR: Record<string, string> = { C: 'Chrome', S: 'Safari', E: 'Edge', F: 'Firefox', W: 'Whale', Z: 'Etc' };

	async function load() {
		const r = await api.get('/service/user/my/device');
		if (!r.ok) return alert('데이터 로딩 도중 오류가 발생하였습니다.');
		devices = [...(r.data || [])].sort((a, b) => String(b.lastLoginedAt || '').localeCompare(String(a.lastLoginedAt || '')) || String(b.createdAt).localeCompare(String(a.createdAt)));
	}
	$effect(() => {
		load();
	});
	async function remove(uid: string) {
		if (!(await confirm('인증기기를 삭제하시겠습니까?'))) return;
		const r = await api.delete(`/service/user/my/device/${uid}`);
		if (!r.ok) return alert('처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 제거되었습니다.');
		load();
	}
	// MUI sx 의 「& th, td」 와 같은 구체성(표 클래스 + 칸) — 칸에 직접 주면 .MuiTable-root td 기본 여백에 진다
	const tableSx = toClass({ '& th, & td': { paddingLeft: 4, paddingRight: 4, fontSize: '13px' }, [up('sm')]: { '& th, & td': { paddingLeft: 8, paddingRight: 8 } } });
	const rowSx = toClass({ '&:last-child td, &:last-child th': { border: 0 } });
	const cell = '';
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ marginTop: 24, padding: 16, [up('md')]: { padding: 24 } }))}>
	<div class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h5' : 'h4'}`, toClass({ display: 'flex', marginBottom: 16, flexDirection: 'column' }))}>
		<Stack sx={{ marginBottom: 4 }}>인증 기기</Stack>
		<h6 class="MuiTypography-root MuiTypography-subtitle2" style="color:#999">인증기기로 등록되지 않은 경우 로그아웃 처리됩니다.</h6>
	</div>
	<table class={cx('MuiTable-root', tableSx)} aria-label="simple table">
		<thead class="MuiTableHead-root">
			<tr class="MuiTableRow-root MuiTableRow-head">
				<th class={cx('MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium', cell)}>OS</th>
				<th class={cx('MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium', cell)}>브라우저</th>
				<th class={cx('MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium', cell)}>마지막 로그인</th>
				<th class={cx('MuiTableCell-root MuiTableCell-head MuiTableCell-alignCenter MuiTableCell-sizeMedium', cell)} style="width:40px">삭제</th>
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each devices || [] as o (o.browserUid)}
				<tr class={cx('MuiTableRow-root', rowSx)}>
					<th class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium', cell)} scope="row">
						{OS[o.browserUid?.[0]] || 'UNKNOWN'}
						{#if me == o.browserUid}<Stack><span class="MuiTypography-root MuiTypography-overline" style="color:#f04438">현재기기</span></Stack>{/if}
					</th>
					<td class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium', cell)}>{BR[o.browserUid?.[1]] || 'UNKNOWN'}</td>
					<td class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium', cell)}>{shortDate(o.lastLoginedAt || o.createdAt)}</td>
					<td class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-alignCenter MuiTableCell-sizeMedium', cell)}>
						<IconButton disabled={me == o.browserUid} onclick={() => remove(o.browserUid)}><Icon name="auto_delete" /></IconButton>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
