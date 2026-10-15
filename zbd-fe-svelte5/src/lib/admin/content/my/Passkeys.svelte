<script lang="ts">
	/** 지문·얼굴 로그인(패스키) 기기 — `GET/POST/DELETE /service/admin/my/passkey`(2026-10-04). 등록은 이 기기의 지문·얼굴·화면 잠금으로 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { cancelled, listPasskeys, passkeySupported, registerPasskey, removePasskey } from '#lib/shared/passkey.ts';
	import Button from '../../ui/Button.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';

	let items = $state<any[] | null>(null);
	let busy = $state(false);
	const supported = passkeySupported();
	$effect(() => {
		listPasskeys('admin').then((v) => (items = v));
	});
	async function add() {
		busy = true;
		try {
			items = await registerPasskey('admin');
			toast('이 기기를 등록했습니다. 다음부터 로그인 화면에서 「지문·얼굴로 로그인」을 누르세요.');
		} catch (e: any) {
			if (e?.name == 'InvalidStateError') alert('이 기기는 이미 등록되어 있습니다.');
			else if (!cancelled(e)) alert(e?.message || '등록하지 못했습니다.');
		} finally {
			busy = false;
		}
	}
	async function remove(it: any) {
		if (!(await confirm(`「${it.name || '이름 없는 기기'}」 를 지울까요?\n그 기기에서는 더 이상 지문으로 로그인할 수 없습니다.`))) return;
		try {
			await removePasskey('admin', it.id);
		} catch (e: any) {
			return alert(e.message);
		}
		items = (items || []).filter((x) => x.id != it.id);
		toast('지웠습니다. 기기에 남은 패스키는 기기 설정(비밀번호 관리자)에서 지울 수 있습니다.');
	}
	const day = (v: any) => (v ? String(v).slice(0, 16).replace('T', ' ') : '');
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ marginTop: 24, padding: 24 }))}>
	<div class={cx('MuiTypography-root MuiTypography-h4', toClass({ display: 'flex', marginBottom: 16, flexDirection: 'column' }))}>
		<div class={cx('MuiStack-root', toClass({ marginBottom: 4 }))}>지문·얼굴 로그인</div>
		<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ color: '#999' }))}>
			등록한 기기에서는 비밀번호·2단계 인증 없이 지문·얼굴(또는 화면 잠금)로 로그인합니다. 지문 정보는 기기 밖으로 나가지 않습니다.
		</h6>
	</div>
	<div class={cx('MuiStack-root', toClass({ gap: 12 }))}>
		{#if items?.length}
			{#each items as it (it.id)}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 12, padding: '10px 12px', border: '1px solid var(--adm-divider, #eee)', borderRadius: 10 }))}>
					<div class={toClass({ flex: 1, minWidth: 0 })}>
						<div class="MuiTypography-root MuiTypography-body2" style="font-weight:600">{it.name || '이름 없는 기기'}</div>
						<div class="MuiTypography-root MuiTypography-caption" style="color:#999">
							등록 {day(it.createdAt)}{it.lastUsedAt ? ` · 마지막 사용 ${day(it.lastUsedAt)}` : ' · 아직 안 씀'}
						</div>
					</div>
					<Button size="small" variant="outlined" color="error" onclick={() => remove(it)}>지우기</Button>
				</div>
			{/each}
		{:else if items}
			<p class="MuiTypography-root MuiTypography-body2" style="color:#999">등록한 기기가 없습니다.</p>
		{/if}
		{#if supported}
			<Button size="large" variant="contained" disabled={busy} onclick={add}>이 기기 등록하기</Button>
		{:else}
			<p class="MuiTypography-root MuiTypography-caption" style="color:#999">이 브라우저에서는 쓸 수 없습니다(https 주소·최신 브라우저 필요).</p>
		{/if}
	</div>
</div>
