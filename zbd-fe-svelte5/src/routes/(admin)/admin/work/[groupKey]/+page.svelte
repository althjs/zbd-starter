<script lang="ts">
	import { page } from '$app/state';
	import WorkList from '#lib/admin/collab/work/WorkList.svelte';
	const groupKey = $derived((page.params as Record<string, string>).groupKey || '');
</script>

{#key groupKey}<WorkList {groupKey} />{/key}
