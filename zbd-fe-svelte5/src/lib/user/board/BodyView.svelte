<script lang="ts">
	/**
	 * 글 본문 — 예전 board/BodyView.tsx. 글마다 CSS(`html_{postId}_css`)·스크립트(json.css·javascript, 관리자 글) 를 붙인다.
	 * ⚠ 회원 게시판(isUserBoard) 글은 회원이 쓴 HTML 이라 DOMPurify 로 거른다(React 판은 그대로 넣었다 — 2026-10-03 바꿈).
	 *   관리자 게시판 글(공지·약관) 은 운영자가 쓴 것이라 그대로.
	 */
	import DOMPurify from 'dompurify';
	import { onDestroy } from 'svelte';
	import { applyCSS, clearCSS, runScript } from '#lib/shared/css.ts';
	import { date } from '#lib/shared/format.ts';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let {
		post,
		dataColumn = 'json',
		trusted = false,
		className = '',
		sx = null,
		hideAdminReply = false
	}: { post: Post; dataColumn?: string; trusted?: boolean; className?: string; sx?: Sx | null; hideAdminReply?: boolean } = $props();

	const cssId = $derived(`html_${post?.postId}`);
	const html = $derived.by(() => {
		const body = post?.[dataColumn]?.body || '';
		return trusted ? body : DOMPurify.sanitize(body);
	});

	$effect(() => {
		const id = cssId;
		if (post?.json?.css) applyCSS(post.json.css, id);
		else clearCSS(id);
		const js = trusted ? post?.json?.javascript : '';
		const t = js ? setTimeout(() => runScript(js, id), 500) : 0;
		return () => {
			clearTimeout(t);
			clearCSS(id);
		};
	});
	onDestroy(() => clearCSS(cssId));
</script>

<div
	class={cx(
		`html_${post?.postId}`,
		post?.json?.className,
		className,
		'MuiBox-root zbd-raw',
		toClass({ '& .emoji': { display: 'inline-block', width: '1.3em', height: '1.3em', marginBottom: '-5px' }, ...(sx || {}) })
	)}
>
	{@html html}
</div>

{#if !hideAdminReply && post?.json?.adminReply}
	<Stack class="admin-reply" sx={{ gap: 8, padding: 24, border: '1px solid #eee', borderRadius: 8, marginTop: 24, backgroundColor: '#f9f9f9' }}>
		<Stack class="admin-reply-hd" direction="row" sx={{ gap: 8, alignItems: 'center' }}>
			<h5 class="MuiTypography-root MuiTypography-h5 admin-reply-dt">{post.json.adminReplyTS ? date(post.json.adminReplyTS) : '답변 내용'}</h5>
			<p class="MuiTypography-root MuiTypography-body1 admin-reply-name">{post.json.adminReplyAdminName || ''}</p>
		</Stack>
		<div class="admin-reply-body MuiBox-root">{@html DOMPurify.sanitize(String(post.json.adminReply || '').replaceAll('\n', '<br/>'))}</div>
	</Stack>
{/if}
