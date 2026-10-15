<script lang="ts">
	/**
	 * 자주 묻는 질문 — 예전 pages/cs/components/FAQ.tsx. 홈(isCSHome) 은 최근 5개 + 검색창(누르면 /cs/faq/{검색어}),
	 * FAQ 화면은 분류 탭 · 검색 · 질문을 누르면 답이 펼쳐진다.
	 * ⚠ v1 목록 대기 — FAQ 게시판 전체를 받아 화면에서 거른다(CLAUDE.md 4).
	 */
	import { goto } from '$app/navigation';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { cx, sx as toClass, up, type Sx } from '#lib/shared/mui/sx.ts';
	import BodyView from '../board/BodyView.svelte';
	import type { Post } from '#lib/shared/types/post.ts';
	import TitleView from '../board/TitleView.svelte';

	let { items, isMobile, isCSHome = false, defaultKeyword = '', sx = null }: { items: Post[]; isMobile: boolean; isCSHome?: boolean; defaultKeyword?: string; sx?: Sx | null } = $props();

	let keyword = $state(defaultKeyword);
	let tab = $state('');
	let opened = $state<Record<string, boolean>>({});
	const tabs = $derived([...new Set(items.map((o) => o.groupSubKey || ''))]);
	const list = $derived.by(() => {
		const k = isCSHome ? '' : keyword.trim();
		const out = items
			.filter((o) => !tab || o.groupSubKey == tab)
			.filter((o) => !k || o.title.includes(k) || String(o.json?.body || '').includes(k))
			.sort((a, b) => (String(b.postId) > String(a.postId) ? 1 : -1));
		return isCSHome ? out.slice(0, 5) : out;
	});
	const search = () => goto(`/cs/faq/${encodeURIComponent(keyword)}`);
	const badge = toClass({ width: '24px', height: '24px', [up('md')]: { width: '28px', height: '28px' } });
	const qa = toClass({ fontSize: '14px', color: '#fff', [up('md')]: { fontSize: '16px' } });
</script>

<Stack {sx}>
	{#if !isCSHome}
		<h2 class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ textAlign: 'center', marginBottom: 24 }))}>자주 묻는 질문</h2>
	{/if}
	<Stack
		direction="column"
		sx={{
			justifyContent: 'center',
			alignItems: 'center',
			padding: 16,
			marginBottom: 8,
			backgroundColor: '#f9f9f9',
			gap: 16,
			[up('sm')]: { flexDirection: 'row', gap: 8 },
			[up('md')]: { marginBottom: 24 }
		}}
	>
		<h5 class="MuiTypography-root MuiTypography-{isMobile ? 'h6' : 'h5'}">무엇을 도와드릴까요?</h5>
		<Stack direction="row" sx={{ gap: 8, width: '100%', [up('sm')]: { width: 'auto' } }}>
			<div role="presentation" onkeyup={(e) => e.key == 'Enter' && isCSHome && search()}>
				<TextField
					bind:value={keyword}
					placeholder="궁금한 점을 입력해보세요."
					sx={{ width: '100%', [up('sm')]: { width: '300px' }, '& .MuiInputBase-input': { paddingTop: '11px', paddingBottom: '11px', backgroundColor: '#fff' } }}
				/>
			</div>
			<MuiButton variant="contained" disabled={!keyword.trim()} onclick={search}>검색</MuiButton>
		</Stack>
	</Stack>
	{#if isCSHome}
		<Stack direction="row" sx={{ justifyContent: 'space-between', paddingTop: 12, paddingBottom: 12, borderBottom: '1px solid #000' }}>
			<h5 class="MuiTypography-root MuiTypography-{isMobile ? 'h6' : 'h5'}">자주묻는 질문</h5>
			<a class={cx('MuiTypography-root MuiLink-root MuiLink-underlineHover', toClass({ fontSize: '14px', [up('md')]: { fontSize: '16px' } }))} href="/cs/faq">더 보기 →</a>
		</Stack>
	{:else}
		<div class="MuiTabs-root faq-tabs" role="tablist">
			{#each ['', ...tabs] as name (name)}
				<button type="button" role="tab" class="MuiButtonBase-root MuiTab-root{tab == name ? ' Mui-selected' : ''}" aria-selected={tab == name} onclick={() => (tab = name)}>{name || '전체'}</button>
			{/each}
		</div>
	{/if}
	{#each list as o (o.postId)}
		<Stack sx={{ borderBottom: '1px solid #ddd' }}>
			<Stack
				direction="row"
				sx={{ justifyContent: 'flex-start', gap: 8, padding: 8, alignItems: 'center', cursor: 'pointer', [up('sm')]: { gap: 16 } }}
				role="button"
				tabindex={0}
				onclick={() => (opened[o.postId] = !opened[o.postId])}
				onkeydown={(e) => e.key == 'Enter' && (opened[o.postId] = !opened[o.postId])}
			>
				<span class={cx('MuiIconButton-root faq-badge', badge)} style="background:{opened[o.postId] ? '#000' : '#ddd'}"><span class={cx('MuiTypography-root MuiTypography-h6', qa)}>Q</span></span>
				<Stack>
					<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ color: 'var(--theme-color-primary)', fontSize: '12px', [up('md')]: { fontSize: '14px' } }))}>{o.groupSubKey}</h6>
					<TitleView variant={isMobile ? 'subtitle2' : 'subtitle1'} post={o} />
				</Stack>
				<Icon name={opened[o.postId] ? 'keyboard_arrow_up' : 'keyboard_arrow_down'} sx={{ marginLeft: 'auto' }} />
			</Stack>
			{#if opened[o.postId]}
				<Stack
					direction="row"
					sx={{
						gap: 8,
						alignItems: 'flex-start',
						paddingTop: 8,
						paddingBottom: 8,
						paddingLeft: 8,
						paddingRight: 8,
						backgroundColor: '#f9f9f9',
						[up('sm')]: { gap: 16, paddingLeft: 16, paddingRight: 16 },
						[up('md')]: { paddingTop: 16, paddingBottom: 16 }
					}}
				>
					<span class={cx('MuiIconButton-root faq-badge', badge)} style="background:var(--theme-color-light, #97cfc8)"><span class={cx('MuiTypography-root MuiTypography-h6', qa)}>A</span></span>
					<BodyView post={o} dataColumn="json" trusted sx={{ '& h1, & h2, & h3, & h4, & h5, & h6, & p, & ol': { margin: '0' }, '& p': { fontSize: '13px', [up('md')]: { fontSize: '15px' } } }} />
				</Stack>
			{/if}
		</Stack>
	{/each}
</Stack>

<style>
	.faq-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.faq-tabs {
		display: flex;
		gap: 16px;
		overflow-x: auto;
		width: 100%;
		scrollbar-width: none; /* React 판 MUI Tabs(scrollable) 처럼 막대 없이 넘긴다 */
	}
	.faq-tabs::-webkit-scrollbar {
		display: none;
	}
	.faq-tabs :global(.MuiTab-root) {
		padding: 8px 0;
		font-size: 14px;
		font-weight: 500;
		opacity: 0.4;
		cursor: pointer;
		white-space: nowrap;
		border-bottom: 2px solid transparent;
	}
	@media (min-width: 600px) {
		.faq-tabs :global(.MuiTab-root) {
			font-size: 13px;
		}
	}
	.faq-tabs :global(.MuiTab-root.Mui-selected) {
		opacity: 1;
		border-bottom-color: var(--theme-color-primary);
		color: var(--theme-color-primary);
	}
</style>
