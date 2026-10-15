<script lang="ts">
	/** 메뉴 없는 관리자 틀 — 단독 앱(chat.zbd.kr·mail.zbd.kr) · 새 창(?popup=1). 로그인·소켓은 그대로(React layouts/admin-popup) */
	import type { Snippet } from 'svelte';
	import WindowHost from '../window/WindowHost.svelte';
	let { children }: { children?: Snippet } = $props();
</script>

<div class="MuiStack-root" id="main_area" style="height:100%">
	<main class="MuiBox-root" style="width:100%;min-height:100vh">{@render children?.()}</main>
</div>
<WindowHost />
