<script lang="ts">
	/** 지문·얼굴 로그인(패스키) — 이 기기 등록 · 등록 기기 목록·지우기(`/service/user/my/passkey`, 2026-10-04). 다음부터 로그인 화면 「지문·얼굴로 로그인」 */
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { cancelled, listPasskeys, passkeySupported, registerPasskey, removePasskey } from '#lib/shared/passkey.ts';
	import { alert, confirm, toast } from '#lib/shared/ui/dialog.svelte.ts';
	import AuthGuard from '#lib/user/AuthGuard.svelte';
	import { session } from '#lib/user/session.svelte.ts';

	const isMobile = $derived(viewport.isMobile);
	const supported = passkeySupported();
	let items = $state<any[] | null>(null);
	let busy = $state(false);
	$effect(() => {
		if (session.profile?.userId) listPasskeys('user').then((v) => (items = v));
	});
	async function add() {
		busy = true;
		try {
			items = await registerPasskey('user');
			toast('이 기기를 등록했어요');
		} catch (e: any) {
			if (e?.name == 'InvalidStateError') alert('이 기기는 이미 등록되어 있어요.');
			else if (!cancelled(e)) alert(e?.message || '등록하지 못했어요.');
		} finally {
			busy = false;
		}
	}
	async function remove(it: any) {
		if (!(await confirm(`「${it.name || '이름 없는 기기'}」 를 지울까요?\n그 기기에서는 더 이상 지문으로 로그인할 수 없어요.`))) return;
		try {
			await removePasskey('user', it.id);
		} catch (e: any) {
			return alert(e.message);
		}
		items = (items || []).filter((x) => x.id != it.id);
		toast('지웠어요');
	}
	const day = (v: any) => (v ? String(v).slice(0, 16).replace('T', ' ') : '');
	const card = 'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root';
</script>

<AuthGuard>
	<Container class="my-passkey" sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: isMobile ? 40 : 80 }}>
		<Stack sx={{ width: '100%', maxWidth: 600, margin: '0 auto' }}>
			<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center', marginBottom: 24 }}>
				<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ fontWeight: 700 }))}>지문·얼굴 로그인</svelte:element>
			</Stack>
			<div class={cx(card, toClass({ padding: 20 }))}>
				<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ color: '#666', marginBottom: 16 }))}>
					이 기기를 등록하면 다음부터 소셜 로그인 없이 지문·얼굴(또는 화면 잠금)로 바로 로그인해요. 지문 정보는 기기 밖으로 나가지 않아요.
				</p>
				<Stack sx={{ gap: 10 }}>
					{#each items || [] as it (it.id)}
						<Stack direction="row" sx={{ alignItems: 'center', gap: 12, padding: '10px 12px', border: '1px solid #eee', borderRadius: 10 }}>
							<Icon name="fingerprint" sx={{ color: 'var(--theme-color-main)' }} />
							<div class={toClass({ flex: 1, minWidth: 0 })}>
								<p class="MuiTypography-root MuiTypography-body2" style="font-weight:600">{it.name || '이름 없는 기기'}</p>
								<p class="MuiTypography-root MuiTypography-caption" style="color:#999">등록 {day(it.createdAt)}{it.lastUsedAt ? ` · 마지막 사용 ${day(it.lastUsedAt)}` : ''}</p>
							</div>
							<MuiButton size="small" variant="outlined" sx={{ color: '#d32f2f', borderColor: '#d32f2f80' }} onclick={() => remove(it)}>지우기</MuiButton>
						</Stack>
					{/each}
					{#if items && !items.length}
						<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ color: '#999' }))}>등록한 기기가 없어요.</p>
					{/if}
					{#if supported}
						<MuiButton variant="contained" disabled={busy} onclick={add} sx={{ width: '100%', marginTop: 8 }}>이 기기 등록하기</MuiButton>
					{:else}
						<p class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: '#999' }))}>이 브라우저에서는 쓸 수 없어요(최신 브라우저 필요).</p>
					{/if}
				</Stack>
			</div>
		</Stack>
	</Container>
</AuthGuard>
