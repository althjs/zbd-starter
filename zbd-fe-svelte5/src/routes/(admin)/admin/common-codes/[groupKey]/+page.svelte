<script lang="ts">
	import { page } from '$app/state';
	import CommonCodes from '#lib/admin/settings/codes/CommonCodes.svelte';
</script>

<CommonCodes groupKey={page.params.groupKey || ''} />
