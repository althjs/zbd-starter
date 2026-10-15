<script lang="ts">
	import { page } from '$app/state';
	import MetaEdit from '#lib/admin/settings/meta/MetaEdit.svelte';
	const p = $derived(page.params as Record<string, string | undefined>);
</script>

{#key `${p.hostingId}/${p.metaKey}`}<MetaEdit hostingId={p.hostingId || 'edit'} metaKey={p.metaKey || ''} />{/key}
