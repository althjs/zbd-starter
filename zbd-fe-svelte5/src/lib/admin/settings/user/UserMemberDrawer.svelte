<script lang="ts">
	/**
	 * 회원 편집 서랍(React components/common/user/UserMemberManageDrawer) — 오른쪽, 폭 50%. 저장 안 한 내용이 있으면 닫기 전에 묻는다.
	 *   <UserMemberDrawer user={u} onclose={…} onSaved={(u) => …} onDelete={(id, un) => …} />   user 가 있으면 열린다
	 */
	import { untrack } from 'svelte';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import { clone } from '../common.ts';
	import UserMemberDetail from './UserMemberDetail.svelte';

	let { user, onclose, onSaved, onDelete }: { user: any; onclose: () => void; onSaved?: (u: any) => void; onDelete?: (id: number, un: boolean | number) => void } = $props();
	let cur = $state<any>(null);
	let orig = '';
	$effect(() => {
		const u = user;
		untrack(() => {
			cur = u ? clone(u) : null;
			orig = u ? JSON.stringify(u) : '';
		});
	});
	async function close() {
		if (JSON.stringify(cur) != orig && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		onclose();
	}
</script>

<Drawer open={!!user && !!cur} anchor="right" onclose={close} paperSx={{ borderRadius: 0, width: '50%' }}>
	{#if cur}
		{#key user}
			<UserMemberDetail
				user={cur}
				isDrawer
				onclose={close}
				onChange={(u) => (cur = u)}
				onSave={(u) => {
					orig = JSON.stringify(u);
					onSaved?.(u);
				}}
				{onDelete}
			/>
		{/key}
	{/if}
</Drawer>
