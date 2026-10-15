<script lang="ts">
	/**
	 * 화면 도움말(TIP 게시판) — 지금 주소(groupSubKey) 의 글이 있으면 제목 아래 안내 상자(React components/common/HintPostView).
	 * 편집(PostManageDrawer)은 묶음 B(콘텐츠) 가 옮긴 뒤 붙인다 — 지금은 보기만.
	 */
	import { page } from '$app/state';
	import DOMPurify from 'dompurify';
	import { api } from '#lib/shared/api.ts';
	import Icon from '../ui/Icon.svelte';
	import IconButton from '../ui/IconButton.svelte';

	let { postId = '', groupKey = 'TIP' }: { postId?: string; groupKey?: string } = $props();
	const cache: Record<string, Promise<any[]>> = ((globalThis as any).__admHintCache ||= {});
	let posts = $state<any[]>([]);
	let detail = $state(false);
	$effect(() => {
		(cache[groupKey] ||= api.get(`/service/admin/post/group/${groupKey}`).then((o) => (o.ok && Array.isArray(o.data) ? o.data : []))).then((v) => (posts = v));
	});
	const post = $derived(
		posts.map((o) => (typeof o.json == 'string' && o.json ? { ...o, json: JSON.parse(o.json) } : o)).find((o) => (postId ? o.postId == postId : o.groupSubKey == page.url.pathname))
	);
	const body = $derived(post?.json?.body || '');
	const more = $derived(!!body && ((body.match(/<p>/g) || []).length > 2 || body.indexOf('<li>') != -1));
</script>

{#if body}
	<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation0 MuiAlert-root MuiAlert-colorSuccess MuiAlert-standardSuccess MuiAlert-standard adm-hint" role="alert">
		<div class="MuiAlert-message">
			<div class="adm-hint-row">
				<div class="adm-hint-body" class:closed={!detail}>{@html DOMPurify.sanitize(body)}</div>
				{#if more}<IconButton size="Small" onclick={() => (detail = !detail)}><Icon name="more_horiz" sx={{ fontSize: 11, color: 'var(--zc-000, #000)', fontWeight: 700 }} /></IconButton>{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.adm-hint {
		padding: 8px 16px;
		position: relative;
		margin-bottom: 16px;
		background-color: rgb(237, 247, 237);
		color: rgb(30, 70, 32);
	}
	.adm-hint-row {
		display: flex;
		align-items: flex-end;
	}
	.adm-hint-body :global(p) {
		margin: 0;
	}
	.adm-hint-body.closed {
		max-height: 40px;
		overflow: hidden;
	}
</style>
