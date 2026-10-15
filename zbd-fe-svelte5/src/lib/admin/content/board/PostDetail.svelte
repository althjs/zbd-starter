<script lang="ts">
	/**
	 * 관리자 글 보기(React board/[groupKey]/[postId]) — 제목·홈페이지 열기·수정·⚙ + 요약 + 신고 처리 + 본문 + 댓글(관리자 모드).
	 * 회원 보기(👤)는 회원 상세 화면을 화면 위 창(lib/admin/window, iframe ?popup=1)으로 연다 — 회원 서랍(UserMemberManageDrawer)은 회원 묶음 것.
	 */
	import { openRouteWindow } from '../../window/windows.svelte.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { siteHomepage } from '../../meta.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { toast } from '../../ui/modal.svelte.ts';
	import PostItemView from '../post/PostItemView.svelte';
	import PostManageDrawer from '../post/PostManageDrawer.svelte';
	import { getPostGroup } from '../post/postGroup.ts';
	import { getPost } from '../post/postStore.svelte.ts';
	import EsMarkTitle from '../ui/EsMarkTitle.svelte';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostSummaryInfo from './PostSummaryInfo.svelte';
	// 게임 기록 다시보기 — 사용자 글 보기와 같은 부품(프리셀 게임판이 사용자 쪽에 있어서 거기 둔다)

	let { groupKey, postId }: { groupKey: string; postId: string } = $props();
	const postGroup = $derived(getPostGroup(groupKey));
	let post = $state<Post | null>(null);
	let pm = $state<any>();

	$effect(() => {
		getPost(postId, true).then((p) => (post = p));
	});

	const homepage = siteHomepage();
	const member = (userId: number) => userId && openRouteWindow(`/admin/user/member/${userId}`, `회원 ${userId}`);
	const f = $derived(postGroup?.feature || {});
	const fixTop = $derived(!!post?.json?.isFixTop);
</script>

<PageContainer title={postGroup?.partnerName ? `[${postGroup.partnerName}] ${post?.title || ''}` : ''}>
	{#if post}
		<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16 }))}>
			<div class="MuiStack-root">
				<div
					class={cx(
						'MuiTypography-root MuiTypography-h3',
						toClass({
							display: 'flex',
							flexDirection: 'row',
							alignItems: 'center',
							justifyContent: 'space-between',
							marginBottom: 16,
							color: post.reportMap?.isPostReported ? 'var(--adm-error-main)' : 'var(--adm-text)'
						})
					)}
				>
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
						{#if post.reportMap?.isPostReported}<Icon name="error" color="Error" sx={{ marginRight: 8 }} />{/if}
						<EsMarkTitle title={post.title} />
					</div>
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
						<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover" href={`${homepage}board/${post.groupKey}/${post.postId}`} target="_blank" rel="noopener">
							<span class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}><Icon name="open_in_new" size="Small" sx={{ marginRight: 4 }} />홈페이지 오픈</span>
						</a>
						{#if postGroup?.feature?.useEditerPage}
							<a
								class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ marginLeft: 8 }))}
								href={`/admin/board/${postGroup.groupKey}/editor/${post.postId}`}
							>
								<span class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}><Icon name="edit" size="Small" sx={{ marginRight: 4 }} />수정</span>
							</a>
						{/if}
						<IconButton aria-label="관리" onclick={() => pm?.openDrawer(post)}><Icon name="settings" /></IconButton>
					</div>
				</div>
				<PostSummaryInfo o={post} {postGroup} onmember={member} />
				<PostItemView {post} />
			</div>
		</div>
	{/if}
	<PostManageDrawer bind:this={pm} {groupKey} groupSubKey={postGroup?.groupKey} subGroups={postGroup?.items || null} onSave={(p) => (post = { ...(post as Post), ...p })} />
</PageContainer>
