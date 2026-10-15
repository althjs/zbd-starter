<script lang="ts">
	import { page } from '$app/state';
	import WorkItemPage from '#lib/admin/collab/work/WorkItemPage.svelte';
	const p = $derived(page.params as Record<string, string>);
</script>

{#key p.workId}<WorkItemPage groupKey={p.groupKey || ''} workId={Number(p.workId) || 0} />{/key}
