<script lang="ts">
	import { page } from '$app/state';
	import MetaKeyManage from '#lib/admin/settings/codes/MetaKeyManage.svelte';
</script>

{#key page.params.metaKey}<MetaKeyManage metaKey={page.params.metaKey || ''} />{/key}
