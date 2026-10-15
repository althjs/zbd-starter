<script lang="ts">
	/**
	 * 사진 목록 줄(img-table) — 글 정보 + 오른쪽 대표 사진(휴대폰은 글 옆 작은 사진). 예전 list-theme/PostListThumbnailCells.tsx
	 * 본문 미리보기는 태그를 지운 글자만.
	 */
	import DOMPurify from 'dompurify';
	import { hostingIdOf, imagePrefix } from '#lib/shared/config.ts';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import PostSummaryInfo from './PostSummaryInfo.svelte';
	import TitleView from './TitleView.svelte';

	let { o, path = '', postGroup, isMobile }: { o: Post; path?: string; postGroup: any; isMobile: boolean } = $props();

	const col = $derived(postGroup?.dataColumn || 'json');
	const json = $derived(o?.[col] || {});
	const image = $derived(json?.image || null);
	const imageUrl = $derived(image ? `${imagePrefix(hostingIdOf(image) || '')}/${image.imageId}/${image.name}` : '/no-image.svg');
	const desc = $derived(DOMPurify.sanitize(json?.body || '', { USE_PROFILES: { html: false } }).replaceAll('&nbsp;', ' '));
	const disabled = $derived(!!postGroup?.feature?.listThemeOptions?.isDisabledDetailLInk);
	const href = $derived(disabled ? undefined : `${path}${path.includes('?') ? `&postId=${o.postId}` : `/${o.postId}`}`);
	const link = 'MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone post-link';
</script>

<td class={cx('MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium cell-content', toClass({ display: 'block', height: 'auto', [up('sm')]: { display: 'table-cell' } }))}>
	<Stack sx={{ height: '100%' }}>
		<Stack class="post-group-name" direction="row" sx={{ gap: 8 }}>
			<p class="MuiTypography-root MuiTypography-body1 groupKey">{postGroup?.name}</p>
			{#if o?.groupSubKey}
				<span>{' > '}</span>
				<p class="MuiTypography-root MuiTypography-body1 groupSubKey">{o.groupSubKey}</p>
			{/if}
		</Stack>
		<a class={cx(link, toClass({ fontSize: '14px', fontWeight: '500', display: 'flex', flexGrow: 1, borderRadius: isMobile ? 8 : 0, [up('md')]: { fontSize: '16px' } }))} {href}>
			<Stack class="img-table-item" sx={{ flexGrow: 1, justifyContent: 'space-between', paddingTop: 8, paddingBottom: 8 }}>
				<Stack class="item-main" direction="row" sx={{ gap: 16, justifyContent: 'space-between' }}>
					<Stack class="text-info" sx={{ flexGrow: 1 }}>
						<TitleView sx={{ fontSize: 20, fontWeight: 700 }} post={o} />
						{#if o?.json?.subTitle}
							<p class={cx('MuiTypography-root MuiTypography-body1 sub-title', toClass({ color: '#909090', fontSize: 12 }))}>{o.json.subTitle}</p>
						{/if}
						<p class={cx('MuiTypography-root MuiTypography-body1 desc-txt', toClass({ '& .emoji': { display: 'inline-block', width: '1.3em', height: '1.3em', marginBottom: '-5px' } }))}>
							{@html desc}
						</p>
					</Stack>
					{#if isMobile}
						<Stack class="thumbnail-cell">
							<img
								class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img thumbnail', toClass({ aspectRatio: 1, objectFit: 'cover', width: 100, height: 100, borderRadius: 8 }))}
								src={imageUrl}
								alt=""
							/>
						</Stack>
					{/if}
				</Stack>
				<PostSummaryInfo {postGroup} isList={true} {o} />
			</Stack>
		</a>
	</Stack>
</td>
{#if !isMobile}
	<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium thumbnail-cell" width="250">
		<a class={link} {href}>
			<img class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img thumbnail', toClass({ aspectRatio: 1.333, objectFit: 'cover' }))} src={imageUrl} alt="" />
		</a>
	</td>
{/if}
