<script lang="ts">
	/**
	 * 관리자 창(MUI Dialog) — 가운데 종이 + 제목·내용·버튼 줄. body 에 그린다. Esc·바탕 누르면 onclose(persistent 면 바탕은 무시).
	 *   <Dialog open={o} onclose={() => (o = false)} title="순서 변경" maxWidth="sm" fullWidth>
	 *     내용
	 *     {#snippet actions()}<Button …>저장</Button>{/snippet}
	 *   </Dialog>
	 * fullScreen=true(또는 휴대폰에서 mobileFullScreen) 면 화면을 덮는다 — 휴대폰 뒤로가기로 닫힌다(installAutoBackToClose).
	 */
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { media } from '../media.svelte.ts';

	let {
		open,
		onclose,
		title = '',
		maxWidth = 'sm',
		fullWidth = false,
		fullScreen = false,
		mobileFullScreen = false,
		persistent = false,
		paperSx = null,
		zIndex = 1300,
		children,
		actions,
		head
	}: {
		open: boolean;
		onclose: () => void;
		title?: string;
		maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | false;
		fullWidth?: boolean;
		fullScreen?: boolean;
		mobileFullScreen?: boolean;
		persistent?: boolean;
		paperSx?: Sx | null;
		zIndex?: number;
		children?: Snippet;
		actions?: Snippet;
		/** 제목 줄을 직접 그릴 때 */
		head?: Snippet;
	} = $props();

	const full = $derived(fullScreen || (mobileFullScreen && media.mdDown));
	$effect(() => {
		if (!open) return;
		const prev = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		return () => (document.body.style.overflow = prev);
	});
	const W = { xs: 444, sm: 600, md: 900, lg: 1200, xl: 1536 };
</script>

{#if open}
	<div use:portal role="presentation" class="MuiDialog-root MuiModal-root" style:z-index={zIndex} onkeydown={(e) => e.key == 'Escape' && onclose()}>
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true" transition:fade={{ duration: 225 }}></div>
		<div
			class="MuiDialog-container MuiDialog-scrollPaper adm-dlg-container"
			role="presentation"
			onclick={(e) => e.target == e.currentTarget && !persistent && onclose()}
			transition:fade={{ duration: 225 }}
		>
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-elevation24 MuiDialog-paper adm-dlg-paper',
					!full && 'MuiPaper-rounded',
					full && 'MuiDialog-paperFullScreen',
					fullWidth && 'MuiDialog-paperFullWidth',
					toClass(paperSx)
				)}
				style:max-width={full ? '100%' : maxWidth ? `${W[maxWidth]}px` : 'calc(100% - 64px)'}
				role="dialog"
				aria-modal="true"
			>
				{#if head}{@render head()}{:else if title}<h2 class="MuiTypography-root MuiTypography-h6 MuiDialogTitle-root adm-dlg-title">{title}</h2>{/if}
				<div class="MuiDialogContent-root adm-dlg-content">{@render children?.()}</div>
				{#if actions}<div class="MuiDialogActions-root MuiDialogActions-spacing adm-dlg-actions">{@render actions()}</div>{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.adm-dlg-container {
		height: 100%;
		outline: 0;
		display: flex;
		justify-content: center;
		align-items: center;
	}
	.adm-dlg-paper {
		margin: 32px;
		position: relative;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		max-height: calc(100% - 64px);
		box-shadow: 0px 9px 46px rgba(0, 0, 0, 0.08);
	}
	.adm-dlg-paper:global(.MuiDialog-paperFullWidth) {
		width: calc(100% - 64px);
	}
	.adm-dlg-paper:global(.MuiDialog-paperFullScreen) {
		margin: 0;
		width: 100%;
		max-height: none;
		height: 100%;
		border-radius: 0;
	}
	.adm-dlg-title {
		margin: 0;
		padding: 16px 24px;
		flex: 0 0 auto;
	}
	.adm-dlg-content {
		flex: 1 1 auto;
		overflow-y: auto;
		padding: 20px 24px;
	}
	.adm-dlg-title + .adm-dlg-content {
		padding-top: 0;
	}
	.adm-dlg-actions {
		display: flex;
		align-items: center;
		padding: 8px;
		justify-content: flex-end;
		flex: 0 0 auto;
		gap: 8px;
	}
</style>
