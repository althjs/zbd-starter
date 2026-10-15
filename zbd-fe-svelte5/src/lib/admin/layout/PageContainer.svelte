<script lang="ts">
	/**
	 * 관리자 화면 틀(React components/PageContainer) — 이동 경로·북마크 + 제목·설명 + 본문.
	 *   <PageContainer>…</PageContainer>                      제목·설명은 지금 메뉴에서
	 *   <PageContainer title="회원 상세" subtitle="…">…</PageContainer>   제목을 주면 이동 경로 끝에도 붙는다
	 *   {#snippet typography()}…{/snippet}               제목 자리를 직접 그릴 때
	 * 휴대폰(<900)은 이동 경로·북마크를 제목 옆 ⋮ 안으로 접는다. 창(?popup=1) 안에서는 제목 없이 본문만.
	 */
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { media } from '../media.svelte.ts';
	import { getPageBgColor, isIgnoreBorder } from '../meta.svelte.ts';
	import { nav, setSubDepthTitle } from '../nav.svelte.ts';
	import Icon from '../ui/Icon.svelte';
	import { IN_WIN_FRAME } from '../window/frame.ts';
	import { SINGLE_APP } from '../config.ts';
	import HintPostView from './HintPostView.svelte';
	import NavigationView from './NavigationView.svelte';

	let {
		title = '',
		subtitle = '',
		sx = null,
		hintPostId = '',
		typography,
		children
	}: { title?: string; subtitle?: string; sx?: Sx | null; hintPostId?: string; typography?: Snippet; children?: Snippet } = $props();

	$effect(() => setSubDepthTitle(title));

	const windowMode = $derived(page.url.searchParams.get('popup') === '1' || IN_WIN_FRAME || !!SINGLE_APP); // 단독 앱(오피스 등)도 이동 경로·제목 없이 — 앱 머리는 그 화면이 그린다
	const fold = $derived(media.mdDown && !windowMode);
	const bg = $derived(getPageBgColor());
	const border = isIgnoreBorder();
	let navEl = $state<HTMLElement | null>(null);
	let navPos = $state({ top: 0, right: 0 });
	const cur = $derived(nav.current);

	const rootSx = $derived(
		border
			? {
					...(windowMode ? { padding: 16 } : { paddingLeft: media.mdDown ? 16 : 32, paddingRight: media.mdDown ? 16 : 32, paddingTop: media.mdDown ? 12 : 32, paddingBottom: 32 }),
					position: 'relative',
					backgroundColor: bg || 'var(--adm-bg)',
					...(bg ? { '& .MuiTable-root': { backgroundColor: 'var(--zc-fff, #fff)' } } : {}),
					...(sx || {})
				}
			: { overflow: 'inherit', position: 'relative', padding: media.mdDown ? 16 : 32, backgroundColor: bg || 'var(--adm-bg)', ...(sx || {}) }
	);
</script>

<div class={cx(border ? 'page-default MuiBox-root' : 'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass(rootSx))}>
	{#if !windowMode && !fold}
		<div class={cx('MuiStack-root', toClass({ flexGrow: 1, justifySelf: 'end', position: 'relative' }))}>
			<NavigationView
				linkSx={{}}
				sx={{
					position: 'static',
					flexGrow: 1,
					right: 0,
					width: '100%',
					justifyContent: 'end',
					marginBottom: 8,
					color: 'var(--gt-700, #909090)',
					'@media (min-width:1280px)': { position: 'absolute', top: -24, width: 'max-content', marginBottom: 0 }
				}}
			/>
		</div>
	{/if}

	<div class={cx('MuiBox-root', toClass({ marginTop: windowMode ? 0 : 8, position: 'relative' }))}>
		{#if fold}
			<button
				type="button"
				class={cx('MuiButtonBase-root MuiIconButton-root MuiIconButton-sizeMedium', toClass({ position: 'absolute', right: -8, top: -6, zIndex: 1, color: 'var(--gt-700, #909090)' }))}
				aria-label="이동 경로·북마크"
				onclick={(e) => {
					navEl = e.currentTarget as HTMLElement;
					const r = navEl.getBoundingClientRect();
					navPos = { top: r.bottom, right: window.innerWidth - r.right };
				}}><Icon name="more_vert" /></button
			>
			{#if navEl}
				<div use:portal role="presentation" class="MuiPopover-root MuiModal-root">
					<div class="MuiBackdrop-root MuiBackdrop-invisible MuiModal-backdrop" aria-hidden="true" onclick={() => (navEl = null)}></div>
					<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper" style="top:{navPos.top}px;right:{navPos.right}px">
						<NavigationView linkSx={{}} sx={{ padding: '8px 12px', color: 'var(--gt-700, #909090)', maxWidth: 'calc(100vw - 32px)', flexWrap: 'wrap' }} />
					</div>
				</div>
			{/if}
		{/if}
		{#if !windowMode}
			<div class={cx('MuiBox-root', fold && toClass({ paddingRight: 36 }))}>
				{#if typography}
					{@render typography()}
					<div class={cx('MuiTypography-root MuiTypography-body1', toClass({ marginBottom: 8 }))}>{subtitle || cur?.sub?.desc || ''}</div>
				{:else}
					<h3 class="MuiTypography-root {media.mdDown ? 'MuiTypography-h5' : 'MuiTypography-h3'}">{title || cur?.sub?.title || cur?.title || ''}</h3>
					<div class={cx('MuiTypography-root MuiTypography-body1', toClass({ marginBottom: 8 }))}>{subtitle || cur?.sub?.desc || cur?.desc || ''}</div>
				{/if}
			</div>
			<HintPostView postId={hintPostId} />
		{/if}
		{@render children?.()}
	</div>
</div>
