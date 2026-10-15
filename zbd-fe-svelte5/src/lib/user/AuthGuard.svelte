<script lang="ts">
	/** 로그인해야 보는 화면 — 확인이 끝나 손님이면 로그인으로(돌아올 주소 쿠키, 뒤로가기 기록 남기지 않음). 예전 UserAuthGuard.tsx */
	import type { Snippet } from 'svelte';
	import { doLogin, session } from './session.svelte.ts';

	let { children }: { children: Snippet } = $props();
	$effect(() => {
		if (session.checked && !session.profile?.userId) doLogin(true);
	});
</script>

{#if session.profile?.userId}{@render children()}{/if}
