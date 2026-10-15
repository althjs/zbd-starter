<script lang="ts">
	/**
	 * 글 본문 보기(React post/components/PostItemView) — DOMPurify 로 거른 본문 + 관리자 답변(json.adminReply).
	 * useImageModal 이면 본문 사진을 누를 때 크게 본다. 글의 json.css 는 `.html_{postId}` 범위로 붙인다.
	 */
	import DOMPurify from 'dompurify';
	import { date } from '#lib/shared/format.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let { post, useImageModal = false }: { post: Post | null; useImageModal?: boolean } = $props();
	let preview = $state<{ src: string; w: number; h: number } | null>(null);
	const cls = $derived(`html_${post?.postId}${post?.json?.className ? ` ${post.json.className}` : ''}`);
	const body = $derived(DOMPurify.sanitize(post?.json?.body || post?.jsonUser?.body || ''));

	$effect(() => {
		const css = post?.json?.css;
		if (!css || !post?.postId) return;
		const el = document.createElement('style');
		el.id = `__post_css_${post.postId}`;
		el.textContent = css;
		document.head.appendChild(el);
		return () => el.remove();
	});

	function click(e: MouseEvent) {
		if (!useImageModal) return;
		const t = e.target as HTMLElement;
		if (!(t instanceof HTMLImageElement) || t.getAttribute('onclick')) return;
		e.preventDefault();
		preview = { src: t.currentSrc || t.src, w: t.naturalWidth || 0, h: t.naturalHeight || 0 };
	}
</script>

<div class={cx('MuiStack-root', toClass({ position: 'relative' }))}>
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div class={cx('MuiBox-root zbd-raw', cls, toClass({ '& img': { maxWidth: '100%', ...(useImageModal ? { cursor: 'pointer' } : {}) } }))} onclick={click}>{@html body}</div>
	{#if post?.json?.adminReply}
		<div class={cx('MuiStack-root', toClass({ gap: 8, padding: 24, border: '1px solid #eee', borderRadius: 4, marginTop: 24, backgroundColor: 'var(--zc-f9f9f9, #f9f9f9)' }))}>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
				<h5 class="MuiTypography-root MuiTypography-h5">{post.json?.adminReplyTS ? date(post.json.adminReplyTS) : '답변 내용'}</h5>
				<p class="MuiTypography-root MuiTypography-body1">{post.json?.adminReplyAdminName || ''}</p>
			</div>
			<div class="MuiBox-root zbd-raw">{@html DOMPurify.sanitize((post?.json?.adminReply || '').replaceAll('\n', '<br/>'))}</div>
		</div>
	{/if}
</div>

{#if preview}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div use:portal class="MuiModal-root" style="position:fixed;inset:0;z-index:1300" onclick={() => (preview = null)}>
		<div class="MuiBackdrop-root" style="position:fixed;inset:0;background:rgba(0,0,0,0.5)"></div>
		<div style="position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);max-width:95vw;max-height:95vh;cursor:zoom-out">
			<img
				src={preview.src}
				alt=""
				style="display:block;width:auto;height:auto;max-width:{preview.w > 0 ? `min(${preview.w}px, 95vw)` : '95vw'};max-height:{preview.h > 0 ? `min(${preview.h}px, 95vh)` : '95vh'}"
			/>
		</div>
	</div>
{/if}
