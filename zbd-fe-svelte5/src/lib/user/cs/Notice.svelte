<script lang="ts">
	/** 고객센터 홈의 공지사항 두 개 — 예전 pages/cs/components/Notice.tsx */
	import { date } from '#lib/shared/format.ts';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up, type Sx } from '#lib/shared/mui/sx.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let { items, isMobile, sx = null }: { items: Post[] | null; isMobile: boolean; sx?: Sx | null } = $props();
</script>

<Stack {sx}>
	<Stack direction="row" sx={{ justifyContent: 'space-between', marginBottom: 16, borderBottom: '1px solid #000', paddingTop: 12, paddingBottom: 12 }}>
		<h5 class="MuiTypography-root MuiTypography-{isMobile ? 'h6' : 'h5'}">공지사항</h5>
		<a class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ fontSize: '14px', [up('md')]: { fontSize: '16px' } }))} href="/board/NOTICE">더 보기 →</a>
	</Stack>
	<Stack direction="column" sx={{ gap: 8, justifyContent: 'space-between', [up('sm')]: { flexDirection: 'row', gap: 16 } }}>
		{#each (items || []).slice(0, 2) as o (o.postId)}
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
					toClass({
						width: 'calc(100% - 32px)',
						borderRadius: '8px',
						paddingLeft: 16,
						paddingRight: 16,
						paddingTop: 8,
						paddingBottom: 8,
						marginBottom: 8,
						[up('sm')]: { width: '50%', paddingTop: 16, paddingBottom: 16 }
					})
				)}
			>
				<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover" style="display:block;width:100%" href={`/board/NOTICE/${o.postId}`}>
					<h6 class="MuiTypography-root MuiTypography-{isMobile ? 'subtitle2' : 'subtitle1'}" style="color:#111927">{o.title}</h6>
					<span class="MuiTypography-root MuiTypography-caption" style="color:#aaa">{date(o.createdAt)}</span>
				</a>
			</div>
		{/each}
	</Stack>
</Stack>
