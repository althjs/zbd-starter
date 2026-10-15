<script lang="ts">
	import { page } from '$app/state';
	import PostEditPage from '#lib/admin/content/board/PostEditPage.svelte';
</script>

{#key `${page.params.groupKey}/${page.params.postId}`}<PostEditPage groupKey={page.params.groupKey || ''} postId={page.params.postId || ''} />{/key}
