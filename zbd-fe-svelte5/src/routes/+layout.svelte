<script lang="ts">
	import { migrateStorage } from '#lib/shared/storage-keys.ts';
	import '../app.css';
	import Dialogs from '#lib/shared/ui/Dialogs.svelte';
	import { setupScrollRestore } from '#lib/shared/scroll-restore.ts';

	// 브라우저 저장소 옛 키 → zbd:{영역}:{이름} 옮겨 담기 — 맨 처음 한 번(storage-keys 를 import 하는 순간에도 돈다, 두 번째는 아무 일 없음)
	migrateStorage();

	// 뒤로 가기 때 스크롤 자리(데이터가 늦게 와도) — scroll-restore.ts
	setupScrollRestore();

	// 뒤로가기 캐시(bfcache)로 되살아난 화면은 새로 읽는다 — 바깥 주소(소셜 로그인 등)에 다녀오면 옛 화면이 그대로 남는다
	if (typeof window != 'undefined') addEventListener('pageshow', (e) => e.persisted && location.reload());

	let { children } = $props();
</script>

{@render children()}
<Dialogs />
