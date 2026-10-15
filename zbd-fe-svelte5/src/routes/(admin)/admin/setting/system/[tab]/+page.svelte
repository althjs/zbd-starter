<script lang="ts">
	import { page } from '$app/state';
	import PageContainer from '#lib/admin/layout/PageContainer.svelte';
	import SystemSettings from '#lib/admin/settings/system/SystemSettings.svelte';
</script>

<PageContainer
	>{#key page.params.tab}<SystemSettings tab={page.params.tab || 'design'} />{/key}</PageContainer
>
