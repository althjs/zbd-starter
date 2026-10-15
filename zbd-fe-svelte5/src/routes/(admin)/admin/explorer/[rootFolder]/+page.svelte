<script lang="ts">
	import { page } from '$app/state';
	import Explorer from '#lib/admin/collab/explorer/Explorer.svelte';
</script>

{#key (page.params as Record<string, string>).rootFolder}<Explorer rootFolder={(page.params as Record<string, string>).rootFolder || ''} />{/key}
