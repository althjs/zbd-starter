<script lang="ts">
	/**
	 * 내 글 관리 `/user/post` — 예전 pages/user/post/index.tsx.
	 * 내가 쓴 글 전부(GET /service/user/post) 를 최근 것부터 카드로: 제목 · 작성 시각 · 본문(회원 HTML → DOMPurify) · 「수정」(내 글).
	 * 「수정」 은 공용 글쓰기 창(openPostEditor) — 저장하면 목록의 그 글을 바꾼다.
	 * 예전 오른쪽 아래 「+」(새 글) 는 groupKey 가 비어 권한 확인을 못 넘어 그려지지 않았다 — 옮기지 않았다.
	 * ⚠ v1 목록 대기 — 지금 API 가 내 글 전체(라이프 기록 포함)를 준다(CLAUDE.md 4). 쪽 나누기 없음(React 판 그대로).
	 */
	import DOMPurify from 'dompurify';
	import { api } from '#lib/shared/api.ts';
	import { shortDate } from '#lib/shared/format.ts';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { openPostEditor } from '../post/editor.svelte.ts';
	import { session } from '../session.svelte.ts';

	type Item = { postId: string; title: string; createdAt: any; userId: any; jsonUser: any; body: string; [k: string]: any };

	let posts = $state<Item[] | null>(null);
	const userId = $derived((session.profile as any)?.userId || null);

	const parse = (o: any): Item => {
		if (o.jsonUser && typeof o.jsonUser == 'string') {
			try {
				o.jsonUser = JSON.parse(o.jsonUser);
			} catch {
				o.jsonUser = null;
			}
		}
		o.body = o.jsonUser?.body || '';
		return o;
	};
	const byNewest = (l: Item[]) => [...l].sort((a, b) => Number(b.createdAt || 0) - Number(a.createdAt || 0));

	$effect(() => {
		// 로그인 확인이 끝난 뒤 한 번(토큰이 붙어야 내 글이 온다)
		if (!session.checked) return;
		api.get('/service/user/post').then((o) => {
			if (!o.ok) return;
			posts = byNewest((o.data || []).map(parse));
		});
	});

	function onSave(p: any) {
		if (!posts) return;
		const q = { ...p };
		if (q.jsonUser?.body) q.body = q.jsonUser.body;
		const has = posts.some((o) => o.postId == q.postId);
		posts = byNewest(has ? posts.map((o) => (o.postId == q.postId ? q : o)) : [...posts, q]);
	}

	const card = 'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root';
</script>

<Stack class="user-post-list">
	{#if posts}
		{#each posts as o (o.postId)}
			<div class={cx(card, toClass({ padding: 24, margin: 8 }))}>
				<Stack direction="row" sx={{ justifyContent: 'space-between' }}>
					<h3 class="MuiTypography-root MuiTypography-h3">{o.title}</h3>
					<h6 class="MuiTypography-root MuiTypography-subtitle2">{shortDate(o.createdAt)}</h6>
				</Stack>
				<Stack sx={{ position: 'relative' }}>
					{#if userId && userId == o.userId}
						<MuiButton sx={{ position: 'absolute', right: 0 }} onclick={() => openPostEditor({ post: o, onSave })}>수정</MuiButton>
					{/if}
					<div class={cx('zbd-raw', toClass({ '& img': { maxWidth: '100%' } }))}>{@html DOMPurify.sanitize(o.body || '')}</div>
				</Stack>
			</div>
		{/each}
		{#if posts.length == 0}
			<div class={cx(card, toClass({ padding: 24 }))}>
				<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ textAlign: 'center' }))}>등록된 글이 없습니다.</p>
			</div>
		{/if}
	{/if}
</Stack>
