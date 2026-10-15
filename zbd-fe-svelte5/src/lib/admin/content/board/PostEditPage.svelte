<script lang="ts">
	/** 글 편집 페이지(React board/[groupKey]/editor/[postId]) — 편집 폼을 서랍 없이 화면 안에(카드 p 24 + 「홈페이지 오픈」) */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { siteHomepage } from '../../meta.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import PostManageDrawer from '../post/PostManageDrawer.svelte';
	import { getPostGroup } from '../post/postGroup.ts';
	import { getAdminPost } from '../post/postStore.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let { groupKey, postId }: { groupKey: string; postId: string } = $props();
	const postGroup = $derived(getPostGroup(groupKey.toUpperCase()));
	let post = $state<Post | null>(null);
	let pm = $state<any>();
	const homepage = siteHomepage();

	$effect(() => {
		if (!pm) return;
		getAdminPost(postId).then((p) => {
			if (!p) return alert('데이터 조회 도중 오류가 발생하였습니다.');
			post = p;
			pm.init(p, postGroup);
		});
	});
</script>

<PageContainer title={postGroup?.partnerName ? `[${postGroup.partnerName}] ${post?.title || ''}` : ''}>
	<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 24 }))}>
		{#if post}
			<a
				class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ position: 'absolute', right: 0, top: 14 }))}
				href={`${homepage}board/${post.groupKey}/${post.postId}`}
				target="_blank"
				rel="noopener"
			>
				<span class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}><Icon name="open_in_new" size="Small" sx={{ marginRight: 4 }} />홈페이지 오픈</span>
			</a>
		{/if}
		<PostManageDrawer bind:this={pm} inline {groupKey} statusMap={postGroup?.statusMap || undefined} onSave={(p) => (post = p)} oncancel={() => history.back()} />
	</div>
</PageContainer>
