<script lang="ts">
	/**
	 * 마이페이지 — 예전 pages/user/my/index.tsx. 회원 정보 수정 · (작성한 글·댓글) · 푸시 알람 설정 · 문의하기.
	 * 「작성한 게시글·댓글」 칸은 전역 CSS 가 링크 주소로 숨긴다(.my-posts) — 마크업은 그대로 둔다.
	 */
	import { onDestroy } from 'svelte';
	import { endSession } from '#lib/shared/api.ts';
	import { applyCSS, clearCSS } from '#lib/shared/css.ts';
	import { number } from '#lib/shared/format.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { site } from '#lib/shared/site.svelte.ts';
	import { confirm } from '#lib/shared/ui/dialog.svelte.ts';
	import AuthGuard from '#lib/user/AuthGuard.svelte';
	import { loadSummary, summary } from '#lib/user/my/summary.svelte.ts';
	import { session } from '#lib/user/session.svelte.ts';
	import { passkeySupported } from '#lib/shared/passkey.ts';

	const isMobile = $derived(viewport.isMobile);
	const siteMeta = $derived(site.base?.siteMeta);
	$effect(() => {
		if (site.base?.CSS?.MY_HOME) applyCSS(site.base.CSS.MY_HOME, 'MY_HOME');
		if (session.profile) loadSummary();
	});
	onDestroy(() => clearCSS('MY_HOME'));

	async function logout() {
		if (!(await confirm('로그아웃 하시겠습니까?'))) return;
		await endSession('user');
		location.replace('/');
	}
	const hasCommentBoard = $derived(!!site.base?.postGroupsMeta?.userPostGroups?.some((x: any) => x.feature?.useComment));
	const card = 'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root';
	const link = 'MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover';
</script>

<!-- 칸 모양(컴포넌트 안에 두면 props 가 되므로 밖에) -->
{#snippet item(cls: string, href: string, label: string, right: import('svelte').Snippet)}
	<Stack class={cls} sx={{ width: isMobile ? '100%' : '49%' }}>
		<div class={cx(card, toClass({ paddingTop: 16, paddingBottom: 16, paddingLeft: 16, paddingRight: 16 }))}>
			<a class={cx(link, toClass({ width: '100%' }))} {href}>
				<Stack direction="row" sx={{ justifyContent: 'space-between' }}>
					<svelte:element this={isMobile ? 'h6' : 'h6'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'subtitle2' : 'subtitle1'}`, toClass({ color: '#000' }))}>{label}</svelte:element>
					{@render right()}
				</Stack>
			</a>
		</div>
	</Stack>
{/snippet}
{#snippet editIcon()}<h6 class="MuiTypography-root MuiTypography-h6" style="color:var(--theme-color-main);height:24px"><Icon name="edit" sx={{ fontSize: '20px' }} /></h6>{/snippet}
{#snippet posts()}<h5 class="MuiTypography-root MuiTypography-h5" style="color:var(--theme-color-main)">{number(summary.postCount || 0)}</h5>{/snippet}
{#snippet comments()}<h5 class="MuiTypography-root MuiTypography-h5" style="color:var(--theme-color-main)">{number(summary.commentCount || 0)}</h5>{/snippet}
{#snippet bell()}<h5 class="MuiTypography-root MuiTypography-h5" style="color:var(--theme-color-main)"><Icon name="edit_notifications" sx={{ fontSize: '20px' }} /></h5>{/snippet}
{#snippet fingerprint()}<h5 class="MuiTypography-root MuiTypography-h5" style="color:var(--theme-color-main)"><Icon name="fingerprint" sx={{ fontSize: '20px' }} /></h5>{/snippet}

<AuthGuard>
	<Container class="my-home" sx={{ paddingTop: isMobile ? 40 : 80, paddingBottom: isMobile ? 40 : 80 }}>
		<Stack sx={{ width: '100%', margin: '0 auto' }}>
			<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center', marginBottom: 24 }}>
				<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ fontWeight: 700 }))}>마이페이지</svelte:element>
			</Stack>
			<Stack>
				<Stack class="hd" direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center', marginBottom: 24, gap: 8 }}>
					<svelte:element this={isMobile ? 'h5' : 'h3'} class="MuiTypography-root MuiTypography-{isMobile ? 'h5' : 'h3'}"
						>{session.profile?.name}<span style="font-weight:400">님, 안녕하세요.</span></svelte:element
					>
					<MuiButton size={isMobile ? 'small' : 'medium'} variant="outlined" sx={{ marginLeft: 'auto' }} onclick={logout}>로그아웃</MuiButton>
				</Stack>
				<Stack class="links" direction="row" sx={{ flexWrap: 'wrap', gap: 16, justifyContent: 'space-between' }}>
					{@render item('my-info', '/user/my/account', '회원 정보 수정', editIcon)}
					{@render item('my-posts', '/user/my/posts', '작성한 게시글', posts)}
					{#if hasCommentBoard && summary.commentCount}
						{@render item('my-comments', '/user/my/comments', '작성한 댓글', comments)}
					{/if}
					{#if passkeySupported()}
						{@render item('my-passkey', '/user/my/passkey', '지문·얼굴 로그인', fingerprint)}
					{/if}
				</Stack>
			</Stack>
		</Stack>
	</Container>
</AuthGuard>
