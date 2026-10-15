<script lang="ts">
	/**
	 * 사이트 틀 — 고정 머리 · 본문 · 고정 바닥(lib/user/layout). 스타터에는 페이지 디자인(GNB·위젯) 이 없다.
	 */
	import { afterNavigate } from '$app/navigation';
	import { applyTheme } from '#lib/shared/theme/theme.ts';
	import { siteMetaOf } from '#lib/user/site.svelte.ts';
	import SiteFooter from '#lib/user/layout/SiteFooter.svelte';
	import SiteHeader from '#lib/user/layout/SiteHeader.svelte';

	let { children } = $props();

	const siteMeta = $derived(siteMetaOf());
	$effect(() => {
		applyTheme(siteMeta?.design?.theme);
	});
	afterNavigate((nav) => {
		const a = nav.from?.url?.pathname;
		const b = nav.to?.url?.pathname;
		if (a && b && a != b) window.scrollTo(0, 0);
	});
</script>

{#if siteMeta}
	<SiteHeader />
	<main class="zs-main">{@render children()}</main>
	<SiteFooter />
{/if}

<style>
	.zs-main {
		min-height: calc(100vh - 260px);
		position: relative;
	}
</style>
