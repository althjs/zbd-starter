<script lang="ts">
	/**
	 * 「글 쓰기」 버튼 — 예전 UserPostManageDrawer(useCreateButton). 로그인했고 이 게시판 쓰기 권한(RW) 이 있는 회원에게만 보인다.
	 * 회원가입(추가 정보) 이 안 끝났으면 가입 화면으로.
	 */
	import { goto } from '$app/navigation';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import { toast } from '#lib/shared/ui/dialog.svelte.ts';
	import { openPostEditor } from '../post/editor.svelte.ts';
	import { getPostGroupAuth } from '../post/postAuth.ts';
	import { doLogin, session } from '../session.svelte.ts';
	import { updatePost } from './board.svelte.ts';

	let { postGroup, groupSubKey = '' }: { postGroup: any; groupSubKey?: string } = $props();

	const canWrite = $derived(!!session.profile?.userId && getPostGroupAuth(session.profile, postGroup?.groupKey).AUTH == 'RW');

	function click() {
		if (!session.profile?.userId) return doLogin();
		if (session.profile.isAdditionalInfoRequired) {
			toast('회원가입을 완료해 주세요');
			goto('/user/join');
			return;
		}
		openPostEditor({ groupKey: postGroup.groupKey, groupSubKey, onSave: updatePost });
	}
</script>

{#if canWrite && !session.profile?.isAdditionalInfoRequired}
	<MuiButton class="write-btn" variant="outlined" size="small" onclick={click}>{postGroup?.feature?.writeButtonText || '글 쓰기'}</MuiButton>
{/if}
