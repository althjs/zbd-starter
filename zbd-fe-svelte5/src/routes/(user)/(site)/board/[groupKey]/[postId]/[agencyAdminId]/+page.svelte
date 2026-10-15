<script lang="ts">
	/**
	 * 대행사 글 주소 `/board/{groupKey}/{postId}/{agencyAdminId}` — 예전 pages/board/[groupKey]/[postId]/[agencyAdminId].tsx.
	 * 주소의 대행사 id 를 localStorage `zbd:site:agencyAdminId` 에 두고(다르면 새로고침), 글은 보통 글 보기 그대로.
	 */
	import { page } from '$app/state';
	import BoardPostPage from '../+page.svelte';
	import { SK } from '#lib/shared/storage-keys.ts';

	$effect(() => {
		const id = (page.params as Record<string, string>).agencyAdminId || '';
		try {
			if (id !== localStorage.getItem(SK.site.agencyAdminId)) {
				localStorage.setItem(SK.site.agencyAdminId, id);
				location.reload();
			}
		} catch {
			/* 저장소를 못 쓰면 그냥 글만 */
		}
	});
</script>

{#key page.params.postId}
	<BoardPostPage />
{/key}
