<script lang="ts">
	import { page } from '$app/state';
	import CategoryView from '#lib/admin/collab/work/CategoryView.svelte';
	const p = $derived(page.params as Record<string, string>);
	const c = $derived((p.cats || '').split('/').map((x) => Number(x) || 0));
</script>

{#key p.cats}<CategoryView groupKey={p.groupKey || ''} c1={c[0] || 0} c2={c[1] || 0} c3={c[2] || 0} />{/key}
