<script lang="ts">
	/** 회원 탈퇴 — 안내 동의 → 확인 → 탈퇴. 예전 pages/user/leave/index.tsx (즈비드는 따로 안내 글이 없다) */
	import { api, endSession } from '#lib/shared/api.ts';
	import Checkbox from '#lib/shared/mui/Checkbox.svelte';
	import Container from '#lib/shared/mui/Container.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { alert, confirm } from '#lib/shared/ui/dialog.svelte.ts';

	let agree = $state(false);
	async function leave() {
		if (!(await confirm('정말 탈퇴하시겠습니까?'))) return;
		const r = await api.delete('/service/user/member/leave');
		if (!r.ok) return alert(r.error.message || '탈퇴에 실패하였습니다.');
		await endSession('user');
		await alert('탈퇴가 완료되었습니다.');
		location.replace('/');
	}
</script>

<Container class="member-leave" sx={{ paddingTop: 80, paddingBottom: 80 }}>
	<Stack sx={{ width: '100%', maxWidth: 475, margin: '0 auto' }}>
		<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center' }}>
			<h2 class={cx('MuiTypography-root MuiTypography-h2', toClass({ fontWeight: 700 }))}>회원 탈퇴</h2>
		</Stack>
		<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 24, marginTop: 24 }))}>
			<Stack class="max">
				<Checkbox bind:checked={agree} label="안내 사항을 모두 확인하였으며, 이에 동의합니다." />
				<MuiButton variant="contained" disabled={!agree} onclick={leave}>탈퇴하기</MuiButton>
			</Stack>
		</div>
	</Stack>
</Container>
