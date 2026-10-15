<script lang="ts">
	import { page } from '$app/state';
	import PostDetail from '#lib/admin/content/board/PostDetail.svelte';
</script>

{#key `${page.params.groupKey}/${page.params.postId}`}<PostDetail groupKey={page.params.groupKey || ''} postId={page.params.postId || ''} />{/key}
