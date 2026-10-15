<script lang="ts">
	import { page } from '$app/state';
	import BoardGroup from '#lib/admin/content/board/BoardGroup.svelte';
</script>

{#key page.params.groupKey}<BoardGroup groupKey={page.params.groupKey || ''} />{/key}
