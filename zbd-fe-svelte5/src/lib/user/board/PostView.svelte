<script lang="ts">
	/**
	 * 게시글 보기 — 예전 board/PostView.tsx 중 즈비드 게시판이 쓰는 길:
	 *   (게시판명) · 카드[제목 · 작성자·작성일 · 조회수 · 공유 · 신고] · 본문 · (앱 다시보기) · 댓글·반응 · [목록 · 글 수정]
	 * 상품·투표·프로모션·다중 답변은 즈비드 게시판이 안 써서 옮기지 않았다.
	 */
	import { siteBase } from '#lib/shared/config.ts';
	import { date } from '#lib/shared/format.ts';
	import Divider from '#lib/shared/mui/Divider.svelte';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import IconButton from '#lib/shared/mui/IconButton.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import ShareDialog from '#lib/shared/ui/ShareDialog.svelte';
	import { toast } from '#lib/shared/ui/dialog.svelte.ts';
	import { siteMetaOf } from '#lib/user/site.svelte.ts';
	import { doLogin, session } from '../session.svelte.ts';
	import BodyView from './BodyView.svelte';
	import { goto } from '$app/navigation';
	import { openPostEditor, postDefault } from '../post/editor.svelte.ts';
	import { convertPost, countUp, updatePost } from './board.svelte.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostSummaryInfo from './PostSummaryInfo.svelte';
	import TitleView from './TitleView.svelte';

	let {
		post,
		postGroup,
		path,
		isMobile,
		spacing = 24,
		useEdit = true,
		showGroupName = true,
		showSubCategory = false,
		className = ''
	}: {
		post: Post;
		postGroup: any;
		path: string;
		isMobile: boolean;
		spacing?: number;
		useEdit?: boolean;
		showGroupName?: boolean;
		showSubCategory?: boolean;
		className?: string;
	} = $props();

	const siteMeta = $derived(siteMetaOf());
	const f = $derived(postGroup?.feature || {});
	const col = $derived(postGroup?.dataColumn || 'json');
	const subName = $derived.by(() => {
		const key = String(post?.groupSubKey || '').trim();
		if (!key) return '';
		return String((postGroup?.items || []).find((x: any) => String(x?.groupSubKey || '') === key)?.name || key);
	});

	let viewEl = $state<HTMLElement>();

	function edit() {
		openPostEditor({
			post,
			onSave: (p) => {
				updatePost(p);
				if (p?.DELETED) {
					setTimeout(() => goto(path, { replace: true }), 1000);
					return;
				}
				Object.assign(post, convertPost({ ...p }, null));
			}
		});
	}

	function reply() {
		if (!session.profile?.userId) return doLogin();
		const parentPostIds = (post?.json?.parentPostIds?.length || 0) > 0 ? [...post.json.parentPostIds!, post.postId] : [post.postId];
		openPostEditor({
			post: {
				...structuredClone(postDefault),
				groupKey: postGroup.groupKey,
				groupSubKey: post.groupSubKey || '',
				json: { ...postDefault.json, parentPostIds },
				title: `RE: ${post.title}`,
				postId: Date.now().toString(36),
				isNew: true
			},
			onSave: updatePost
		});
	}

	// 조회수 — 1초 뒤 한 번(내 글·마이페이지 제외, React 판과 같은 조건)
	$effect(() => {
		if (!post?.postId || session.profile?.userId == post.userId) return;
		const t = setTimeout(() => countUp(post), 1000);
		return () => clearTimeout(t);
	});

	const sub2 = toClass({ fontSize: '12px', [up('sm')]: { fontSize: '14px' } });
	// Stack spacing 은 margin 으로 준다(MUI 와 같게 — gap 이 아니다)
	const spacingSx = $derived(toClass({ '& > :not(style) ~ :not(style)': { marginTop: spacing } }));
</script>

{#if post?.postId}
	<Stack class={cx('post-view', className, spacingSx)}>
		{#if showGroupName || showSubCategory}
			<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'} post-title`, toClass({ textAlign: 'center' }))}>
				{#if postGroup?.groupKey == 'TERMS'}
					{post.groupSubKey}
				{:else}
					{#if showGroupName}<span class="post-group-name">{postGroup?.name}</span>{/if}
					{#if showGroupName && showSubCategory && subName}<span class="post-title-sep">{' · '}</span>{/if}
					{#if showSubCategory && subName}<span class="post-sub-category">{subName}</span>{/if}
				{/if}
			</svelte:element>
		{/if}

		<div bind:this={viewEl} class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root post-view-card', '')}>
			<!-- 머리 = 제목(전체 폭) + 한 줄(작성자·작성일 … 조회수 · 공유 · 더보기) + 구분선(.post-hd 안 — 운영자 CSS `.post-hd hr` 가 기대). 2026-10-03 아이콘이 머리 옆에 따로 떠 있던 것 정리 -->
			<Stack class="post-hd-container" direction="column" sx={{ gap: 8 }}>
				<Stack class="post-hd" direction="column" sx={{ gap: 8 }}>
					<TitleView variant={isMobile ? 'h6' : 'h4'} {post} />
					<Stack class="info" direction="row" sx={{ gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
						<Stack class="writer-dt" direction="row" sx={{ flexGrow: 1, gap: 8, alignItems: 'center' }}>
							<h6 class={cx('MuiTypography-root MuiTypography-subtitle2 writer', sub2)}><span>작성자:</span> {post?.[col]?.name || '***'}</h6>
							<h6 class={cx('MuiTypography-root MuiTypography-subtitle2 createdAt', sub2)}><span>작성일:</span> {date(post?.createdAt)}</h6>
						</Stack>
						<PostSummaryInfo hiddenCreatedAt={true} hiddenWriter={true} o={post} {postGroup} />
						<Stack direction="row" class="more-icons" sx={{ flexShrink: 0, alignItems: 'center', marginTop: -8, marginBottom: -8, marginRight: -8 }}>
							<ShareDialog asLink url={`${siteBase(siteMeta?.homepage)}board/${post.groupKey.toLowerCase()}/${post.postId}`} {isMobile} />
						</Stack>
					</Stack>
					<hr class="MuiDivider-root MuiDivider-fullWidth" />
				</Stack>
			</Stack>

			<Stack class="etc-info" sx={{ alignItems: 'center' }} />

			<BodyView
				className="post-body"
				{post}
				dataColumn={col}
				trusted={!postGroup?.isUserBoard}
				sx={{
					minHeight: postGroup?.appId || post.groupKey?.toUpperCase() == 'FREECELL' ? 0 : 300,
					'& img': { display: 'block', maxWidth: '100%', height: 'auto', [up('md')]: { maxWidth: '900px' }, [up('lg')]: { maxWidth: '1280px' } },
					'&::after': { content: '""', display: 'table', clear: 'both' }
				}}
			/>
		</div>

		<Stack class="post-btns" direction="row" sx={{ alignItems: 'center', justifyContent: 'center', '& > :not(style) ~ :not(style)': { marginLeft: 8 } }}>
			<MuiButton class="list-btn" href={path} variant="contained" sx={{ minWidth: 100 }}>목록</MuiButton>
			{#if useEdit && !postGroup?.isEditDisabled && session.profile?.userId && session.profile.userId == post.userId}
				<MuiButton class="edit-btn" variant="outlined" sx={{ minWidth: 100 }} onclick={edit}>글 수정</MuiButton>
			{/if}
			{#if f.useReply}
				<MuiButton class="reply-btn" variant="contained" onclick={reply}>답글 작성</MuiButton>
			{/if}
		</Stack>
	</Stack>
{/if}

<style>
	/* 기본 모양(게시판마다 디자인 CSS 로 덮는다 — 그래서 최소만) */
	:global(.post-view > .post-view-card) {
		border: 1px solid #e3e8e7;
		border-radius: 14px;
		box-shadow: none;
		padding: clamp(16px, 4vw, 32px);
	}
	:global(.post-view .post-hd .title > *) {
		font-size: clamp(21px, 3vw, 26px);
		font-weight: 800;
	}
	:global(.post-view .post-hd :is(.writer, .createdAt, .view-cnt)) {
		font-size: 13px;
		color: #8a9796;
	}
	:global(.post-view .post-hd hr) {
		border-color: #eef2f1;
	}
</style>
