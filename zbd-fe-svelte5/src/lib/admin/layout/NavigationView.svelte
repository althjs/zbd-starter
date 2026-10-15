<script lang="ts">
	/**
	 * 이동 경로(🏠 › 메뉴 › 하위) — 머리줄·페이지 제목 위(React layouts/dashboard/components/NavigationView). 스타터는 북마크(즐겨찾기 메뉴) 없음.
	 */
	import { page } from '$app/state';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { media } from '../media.svelte.ts';
	import { nav } from '../nav.svelte.ts';
	import SvgIcon from '../ui/SvgIcon.svelte';

	let { sx = null, linkSx = null, class: cls = '' }: { sx?: Sx | null; linkSx?: Sx | null; class?: string } = $props();

	const linkCls = $derived(toClass(linkSx ?? { '&:hover': { color: '#fff' } }));
	const cur = $derived(nav.current);
</script>

<div class={cx('MuiStack-root', cls, toClass({ flexDirection: 'row', alignItems: 'center', gap: '8px', ...(sx || {}) }))}>
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeSmall" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
			><a class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone', toClass({ color: 'inherit' }), linkCls)} href="/admin" aria-label="관리자 홈"
				><svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium" focusable="false" aria-hidden="true" viewBox="0 0 24 24"><path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"></path></svg></a
			></svg
		>
		{#if cur?.title}<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ marginLeft: 4 }))}>{cur.title}</p>{/if}
		{#if cur?.sub}
			<SvgIcon name="KeyboardArrowRight" />
			<a class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone', toClass({ color: 'inherit' }), linkCls)} href={cur.sub.path}
				><p class="MuiTypography-root MuiTypography-body1">{cur.sub.title}</p></a
			>
		{/if}
		{#if !media.mdDown && cur?.subDepthTitle && cur?.sub && (!page.url.pathname.endsWith(String(cur.sub.path)) || cur.sub.path === '/')}
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
				<SvgIcon name="KeyboardArrowRight" />
				<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ maxWidth: 150, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }))}>{cur.subDepthTitle}</p>
			</div>
		{/if}
	</div>
</div>
