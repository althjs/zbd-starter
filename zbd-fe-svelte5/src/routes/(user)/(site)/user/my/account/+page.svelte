<script lang="ts">
	/** 회원 정보 수정 — 예전 pages/user/my/Account.tsx (회원 정보 · 인증 기기 · 로그아웃 · 탈퇴 링크) */
	import { goto } from '$app/navigation';
	import { endSession } from '#lib/shared/api.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { confirm, toast } from '#lib/shared/ui/dialog.svelte.ts';
	import AuthGuard from '#lib/user/AuthGuard.svelte';
	import LoginDevice from '#lib/user/my/LoginDevice.svelte';
	import UserAccount from '#lib/user/my/UserAccount.svelte';
	import { session } from '#lib/user/session.svelte.ts';

	const isMobile = $derived(viewport.isMobile);
	let sent = false; // 한 번만 — 프로필이 다시 읽힐 때마다 안내가 쌓였다
	$effect(() => {
		if (session.profile?.isAdditionalInfoRequired && !sent) {
			sent = true;
			toast('회원가입을 완료해 주세요');
			goto('/user/join', { replace: true });
		}
	});
	async function logout() {
		if (!(await confirm('로그아웃 하시겠습니까?'))) return;
		await endSession('user');
		location.replace('/');
	}
</script>

<AuthGuard>
	<Container sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: isMobile ? 40 : 80 }}>
		<Stack sx={{ width: '100%', maxWidth: 475, margin: '0 auto' }}>
			<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center', marginBottom: 24 }}>
				<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ fontWeight: 700 }))}>정보수정</svelte:element>
			</Stack>
			<Stack>
				<UserAccount />
				<LoginDevice {isMobile} />
				<MuiButton variant="outlined" sx={{ marginTop: 16 }} onclick={logout}>로그아웃</MuiButton>
				<a class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone', toClass({ marginTop: 8, fontWeight: 400, fontSize: 13 }))} href="/user/leave">회원탈퇴</a>
			</Stack>
		</Stack>
	</Container>
</AuthGuard>
