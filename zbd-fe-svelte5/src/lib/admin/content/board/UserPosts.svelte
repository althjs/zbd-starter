<script lang="ts">
	/**
	 * 사용자 글 관리(React board/user-posts) — 검색 엔진 게시판 글을 서버 쪽 나눔으로(`GET /service/board/{호스팅}?…`, 그룹 비면 ES 게시판 전부).
	 * 검색값은 sessionStorage `zbd:admin:userPostSearch`.
	 */
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { getHostingId } from '../../meta.svelte.ts';
	import { activeNav } from '../../nav.svelte.ts';
	import Tabs from '../../ui/Tabs.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import { getYMD } from '../dates.ts';
	import PostManageDrawer from '../post/PostManageDrawer.svelte';
	import { convertPostItems, userPostGroups } from '../post/postGroup.ts';
	import type { Post } from '#lib/shared/types/post.ts';
	import BoardListTable from './BoardListTable.svelte';
	import BoardSearchForm from './BoardSearchForm.svelte';
	import { SK } from '#lib/shared/storage-keys.ts';

	const KEY = SK.admin.userPostSearch;
	const init = () => {
		try {
			const s = sessionStorage.getItem(KEY);
			if (s) return JSON.parse(s);
		} catch {}
		return { groupKey: '', groupSubKey: '', orderBy: 'postId', orderByDesc: 'desc', keyword: '', pageSize: 10, pageNo: 1, from: '2025-01-01', to: getYMD() };
	};
	const esGroups = userPostGroups().filter((g) => g.feature?.useES);
	let search = $state<any>(init());
	let items = $state<Post[] | null>(null);
	let total = $state(0);
	let pm = $state<any>();

	async function getData(s: any) {
		search = s;
		try {
			sessionStorage.setItem(KEY, JSON.stringify(s));
		} catch {}
		const all = esGroups.map((g) => g.groupKey).join(',');
		const q = Object.keys(s)
			.map((k) => `${k}=${encodeURIComponent(k == 'groupKey' && !s[k] ? all : s[k] || '')}`)
			.join('&');
		const o = await api.get(`/service/board/${getHostingId()}?${q}`);
		if (!o.ok) return;
		items = convertPostItems(o.data || [], o.meta?.summary);
		total = Number(o.meta?.totalCount || 0);
	}
	$effect(() => {
		getData({ ...search });
		const un = () => sessionStorage.removeItem(KEY);
		window.addEventListener('beforeunload', un);
		return () => window.removeEventListener('beforeunload', un);
	});
</script>

<PageContainer>
	<BoardSearchForm useES isUserPost {search} {getData} sx={{ marginBottom: 8 }} />
	<Tabs
		items={[{ value: '', label: '전체' }, ...esGroups.map((g) => ({ value: g.groupKey, label: g.name || g.groupKey }))]}
		value={search.groupKey || ''}
		onchange={(v) => getData({ ...search, groupKey: v })}
	/>
	{#if items}
		<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16 }))}>
			<BoardListTable itemList={items} isUserPost onmanage={(o) => pm?.openDrawer(o)} />
		</div>
		<TablePagination sx={{ marginTop: 16 }} {search} getData={(s) => getData({ ...s })} pageCount={Math.ceil((total || 1) / search.pageSize)} />
	{/if}
	{#if activeNav()?.auth == 'RW'}
		<PostManageDrawer
			bind:this={pm}
			onSave={(p) => {
				const list = [...(items || [])];
				const i = list.findIndex((o) => o.postId == p.postId);
				if (i >= 0) {
					if (p.DELETED) list.splice(i, 1);
					else list[i] = p;
				} else list.push(p);
				items = list;
			}}
		/>
	{/if}
</PageContainer>
