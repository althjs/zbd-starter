<script lang="ts">
	/**
	 * 게시글 `/board/{groupKey}/{postId}` — 예전 pages/board/[groupKey]/[postId]/index.tsx.
	 * 로그인해야 보는 게시판(useLogInView) 은 로그인으로.
	 */
	import { setContext } from 'svelte';
	import { page } from '$app/state';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { site } from '#lib/shared/site.svelte.ts';
	import { board, findPostGroup, getPost } from '#lib/user/board/board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostView from '#lib/user/board/PostView.svelte';
	import { doLogin, session } from '#lib/user/session.svelte.ts';

	const groupKey = $derived((page.params.groupKey || '').toLowerCase());
	const postId = $derived(page.params.postId || '');
	const group = $derived(site.base ? findPostGroup(groupKey) : undefined);

	const ctx = $state({ post: null as Post | null, postGroup: null as any });
	setContext('zbd:board', ctx);

	$effect(() => {
		const id = postId;
		const g = groupKey;
		ctx.post = null;
		getPost(g, id).then((p) => {
			if (postId == id) ctx.post = p;
		});
	});
	$effect(() => {
		ctx.postGroup = group;
		if (group?.feature?.useLogInView && session.checked && !session.profile) doLogin(true);
	});
</script>

{#if board.unknown[postId]}
	<Stack class="unknown-post-item" sx={{ marginTop: 80 }}>
		<p class="MuiTypography-root MuiTypography-body2" style="text-align:center">해당 게시글은 존재하지 않거나, 비공개 상태입니다.</p>
	</Stack>
{/if}

{#if ctx.post}
	{#if group}
		<Container sx={{ paddingTop: viewport.isMobile ? 40 : 80, paddingBottom: viewport.isMobile ? 40 : 80 }}>
			<PostView spacing={24} post={ctx.post} postGroup={group} path={`/board/${groupKey}`} isMobile={viewport.isMobile} />
		</Container>
	{/if}
{/if}
