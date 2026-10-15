<script lang="ts">
	import { page } from '$app/state';
	import UserMemberManage from '#lib/admin/settings/user/UserMemberManage.svelte';
</script>

{#key page.params.userId}<UserMemberManage userId={page.params.userId || ''} />{/key}
