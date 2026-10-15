<script lang="ts">
	/**
	 * 사용자 화면(zbd.kr) 묶음 — API 는 사용자 토큰(zbd:auth:user).
	 * 예전 App.tsx·Router.tsx 의 LayoutSelector 자리: 로그인 상태 확인, window.zbdSite 등록,
	 * #root 에 경로 클래스(`/app/store` → `_app_store` — 전역 CSS 가 `#root._app_store` 처럼 겨냥) · 언어 감싸개(`.ko`).
	 */
	import { page } from '$app/state';
	import { setRealm } from '#lib/shared/api.ts';
	import { HOST_REALM } from '#lib/shared/config.ts';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { loadSite } from '#lib/shared/site.svelte.ts';
	import { registerZbdSite } from '#lib/user/zbdSite.ts';
	import { loadProfile } from '#lib/user/session.svelte.ts';
	import PostEditorDrawer from '#lib/user/post/PostEditorDrawer.svelte';
	import { ensureUid } from '#lib/user/visit.ts';

	// 관리자 주소(backyard·chat·mail)에서는 사용자 화면을 그리지 않는다 — 관리자로(2026-10-03 — backyard / 가 사용자 홈으로 그려졌다, nginx 도 같게)
	const adminHost = HOST_REALM == 'admin';
	if (adminHost) location.replace('/admin');

	setRealm('user');
	if (typeof document != 'undefined') ensureUid(); // 브라우저 id(zbd_uid) — 로그인 토큰이 이 값에 묶인다
	loadSite();
	loadProfile();
	registerZbdSite();

	let { children } = $props();

	$effect(() => {
		const root = document.getElementById('root');
		if (root) root.className = page.url.pathname.replaceAll('/', '_');
	});
</script>

{#if !adminHost}
	<Stack class="ko">{@render children()}</Stack>
{/if}
<!-- 글쓰기·글 수정 창(하나) — 게시판·푸터 문의(window.zbdSite.openDrawer) 가 연다 -->
<PostEditorDrawer />
