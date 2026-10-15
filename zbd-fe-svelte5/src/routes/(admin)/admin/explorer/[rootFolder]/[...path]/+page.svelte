<script lang="ts">
	/** 웹하드 문서 — React 는 d1~d5 폴더 깊이였다. 여기선 깊이 제한 없음 */
	import { page } from '$app/state';
	import MarkdownFile from '#lib/admin/collab/explorer/MarkdownFile.svelte';
	const target = $derived(((page.params as Record<string, string>).path || '').split('/').filter(Boolean).join('/'));
</script>

{#key target}<MarkdownFile rootFolder={(page.params as Record<string, string>).rootFolder || ''} targetFile={target} />{/key}
