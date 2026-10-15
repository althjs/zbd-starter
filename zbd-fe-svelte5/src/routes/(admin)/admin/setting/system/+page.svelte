<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	onMount(() => goto('/admin/setting/system/design', { replaceState: true }));
</script>
