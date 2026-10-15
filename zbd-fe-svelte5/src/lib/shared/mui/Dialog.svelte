<script lang="ts">
	/**
	 * MUI Dialog 호환 — body 에 그리는 대화 창. title·actions 는 snippet.
	 * fullScreen 이면 화면 전체(휴대폰 편집 창 등). Esc·바탕 누르기로 onclose(막으려면 disableBackdropClose).
	 * 본문이 넘치면 Scrollbar(simplebar) 막대로 — 제목·버튼 줄은 그대로 두고 본문만 스크롤(MUI scroll="paper" 와 같은 자리).
	 */
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
	import { portal } from '../ui/portal';
	import Scrollbar from '../ui/Scrollbar.svelte';
	import { cx, sx as toClass, type Sx } from './sx';

	let {
		open,
		onclose,
		fullScreen = false,
		maxWidth = 600,
		paperSx = null,
		class: cls = '',
		disableBackdropClose = false,
		fitWidth = true,
		title,
		actions,
		children,
		onpaste
	}: {
		open: boolean;
		onclose: () => void;
		fullScreen?: boolean;
		maxWidth?: number | string;
		paperSx?: Sx | null;
		class?: string;
		disableBackdropClose?: boolean;
		/** 창 폭을 내용에 맞춘다(기본). 창 폭을 paperSx 로 정하는 창(사진 고르기 등)은 false — 켜 두면 내용 폭을 처음 잰 값으로 고정해 넓힌 창 안에서 내용이 좁게 잘린다 */
		fitWidth?: boolean;
		title?: Snippet;
		actions?: Snippet;
		children?: Snippet;
		onpaste?: (e: ClipboardEvent) => void;
	} = $props();

	$effect(() => {
		if (!open) return;
		const prev = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		const key = (e: KeyboardEvent) => e.key == 'Escape' && onclose();
		window.addEventListener('keydown', key);
		return () => {
			document.body.style.overflow = prev;
			window.removeEventListener('keydown', key);
		};
	});
</script>

{#if open}
	<div use:portal role="presentation" class={cx('MuiDialog-root MuiModal-root', cls)} {onpaste} transition:fade={{ duration: 200 }}>
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true"></div>
		<div class="MuiDialog-container MuiDialog-scrollPaper" role="presentation" onclick={(e) => e.target === e.currentTarget && !disableBackdropClose && onclose()}>
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation24 MuiDialog-paper MuiDialog-paperScrollPaper zbd-sb-paper',
					fullScreen && 'MuiDialog-paperFullScreen',
					toClass(fullScreen ? paperSx : { maxWidth: typeof maxWidth == 'number' ? `${maxWidth}px` : maxWidth, ...(paperSx || {}) })
				)}
				role="dialog"
				aria-modal="true"
			>
				{#if title}<h2 class="MuiTypography-root MuiTypography-h6 MuiDialogTitle-root">{@render title()}</h2>{/if}
				<Scrollbar class="zbd-sb-fill zbd-sb-dialog" fitWidth={!fullScreen && fitWidth}><div class="MuiDialogContent-root">{@render children?.()}</div></Scrollbar>
				{#if actions}<div class="MuiDialogActions-root MuiDialogActions-spacing">{@render actions()}</div>{/if}
			</div>
		</div>
	</div>
{/if}
