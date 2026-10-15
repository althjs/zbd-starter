<script lang="ts">
	/**
	 * 게시판 다중 답변(React board/components/PostReplyThread) — post.json.replies[] 에 쌓는다.
	 * 옛 단일 답변(json.adminReply) 은 id 0 가상 항목으로 맨 앞에 세운다(수정은 adminReply 로).
	 * 본문은 일반 글자(pre-wrap) — 사용자 화면에 그대로 나가는 값이라 HTML 을 받지 않는다.
	 */
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { date } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import type { AdminProfile } from '#lib/shared/types/admin.ts';
	import { getNextId } from '../dates.ts';
	import type { Post } from '#lib/shared/types/post.ts';

	let { post, profile, readonly = false, onSave }: { post: Post; profile: AdminProfile; readonly?: boolean; onSave: (patch: Record<string, any>) => Promise<void> | void } = $props();
	let body = $state('');
	let editId = $state(-1);
	let editBody = $state('');
	let saving = $state(false);

	const json = $derived(post?.json || {});
	const replies = $derived((json.replies || []) as any[]);
	const thread = $derived.by(() => {
		const list = [...replies];
		if (json.adminReply) list.unshift({ id: 0, type: 'ANSWER', adminId: json.adminReplyAdminId, adminName: json.adminReplyAdminName, body: json.adminReply, ts: json.adminReplyTS });
		return list;
	});

	async function add() {
		const v = body.trim();
		if (!v || saving) return;
		saving = true;
		try {
			await onSave({ replies: [...replies, { id: getNextId(replies), type: 'ANSWER', adminId: profile.adminId, adminName: profile.name || profile.nickname || '관리자', body: v, ts: Date.now() }] });
			body = '';
		} finally {
			saving = false;
		}
	}
	async function saveEdit(o: any) {
		const v = editBody.trim();
		if (!v || saving) return;
		saving = true;
		try {
			if (o.id == 0) await onSave({ adminReply: v, adminReplyTS: Date.now() });
			else await onSave({ replies: replies.map((x) => (x.id == o.id ? { ...x, body: v, updatedAt: Date.now() } : x)) });
			editId = -1;
			editBody = '';
		} finally {
			saving = false;
		}
	}
	async function del(o: any) {
		if (!(await confirm('해당 답변을 삭제하시겠습니까?'))) return;
		if (o.id == 0) await onSave({ adminReply: '' });
		else await onSave({ replies: replies.filter((x) => x.id != o.id) });
	}
	const chip = toClass({ height: 20, fontSize: '0.7rem' });
	const gray = toClass({ color: 'var(--gt-700, #909090)' });
</script>

<div class={cx('MuiStack-root', toClass({ gap: 16 }))}>
	{#if thread.length == 0}<p class={cx('MuiTypography-root MuiTypography-body2', gray)}>등록된 답변이 없습니다.</p>{/if}
	{#each thread as o (o.id)}
		{@const isUser = !!o?.userId}
		{@const canEdit = !readonly && !isUser && (o.id == 0 || o.adminId == profile.adminId || profile.isMaster)}
		<div
			class={cx(
				'MuiStack-root',
				toClass({ gap: 4, padding: 12, borderRadius: 4, border: '1px solid var(--gt-300, #e5e5e5)', backgroundColor: isUser ? 'var(--gt-150, #f5f5f5)' : 'var(--zc-fff, #fff)' })
			)}
		>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
				<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ fontWeight: 700 }))}>{isUser ? o.userName || '작성자' : o.adminName || '관리자'}</h6>
				<div
					class={cx(`MuiChip-root MuiChip-sizeSmall ${isUser ? 'MuiChip-outlined MuiChip-colorDefault MuiChip-outlinedDefault' : 'MuiChip-filled MuiChip-colorPrimary MuiChip-filledPrimary'}`, chip)}
				>
					<span class="MuiChip-label MuiChip-labelSmall">{isUser ? '재문의' : '답변'}</span>
				</div>
				<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{date(o.ts)}{o.updatedAt ? ` (수정 ${date(o.updatedAt)})` : ''}</span>
				{#if canEdit}
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 4, marginLeft: 'auto' }))}>
						<Button size="small" sx={{ minWidth: 'unset', padding: 4, fontSize: 13 }} onclick={() => ((editId = o.id), (editBody = o.body || ''))}>수정</Button>
						<Button size="small" color="error" sx={{ minWidth: 'unset', padding: 4, fontSize: 13 }} onclick={() => del(o)}>삭제</Button>
					</div>
				{/if}
			</div>
			{#if editId == o.id}
				<div class={cx('MuiStack-root', toClass({ gap: 8, marginTop: 4 }))}>
					<TextField size="small" multiline minRows={3} bind:value={editBody} />
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8 }))}>
						<Button size="small" variant="contained" disabled={!editBody.trim() || saving} onclick={() => saveEdit(o)}>저장</Button>
						<Button size="small" variant="outlined" onclick={() => ((editId = -1), (editBody = ''))}>취소</Button>
					</div>
				</div>
			{:else}
				<div class={cx('MuiBox-root', toClass({ whiteSpace: 'pre-wrap', fontSize: 14 }))}>{o.body}</div>
			{/if}
		</div>
	{/each}
	{#if !readonly}
		<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
			<TextField size="small" multiline minRows={3} placeholder="답변을 입력하세요" bind:value={body} />
			<Button variant="contained" size="small" sx={{ alignSelf: 'flex-start' }} disabled={!body.trim() || saving} onclick={add}>{saving ? '저장 중...' : '답변 등록'}</Button>
		</div>
	{/if}
</div>
