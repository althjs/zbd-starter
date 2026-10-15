<script lang="ts">
	/**
	 * 사용자 회원 상세 — 회원 정보 편집(UserMemberDetail) + 그 회원의 게시글(댓글은 고급 게시판 플러그인).
	 * 게시글 = ES 게시판 `GET /service/board/{호스팅}?userId=` — 쪽·게시판·세부 분류는 **서버에 넘겨** 받은 쪽만 그린다
	 * (React 는 10000개를 한 번에 받아 화면이 걸렀다). 쪽 크기는 localStorage `zbd:admin:userManagePostPageSize`·`…CommentPageSize`,
	 * 탭·쪽은 sessionStorage `zbd:admin:userMemberManageState`(새로고침하면 지움). 게시글 「관리」 = 글 편집 서랍(메뉴 RW 일 때).
	 * 표 부품은 콘텐츠 묶음(lib/admin/content/board) 것을 쓴다.
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import Select from '#lib/shared/mui/Select.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { activeNav } from '../../nav.svelte.ts';
	import Card from '../../ui/Card.svelte';
	import TablePagination from '../../ui/TablePagination.svelte';
	import Tabs from '../../ui/Tabs.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import BoardListTable from '../../content/board/BoardListTable.svelte';
	import PostManageDrawer from '../../content/post/PostManageDrawer.svelte';
	import { getEsPosts } from '../../content/post/postStore.svelte.ts';
	import { userPostGroups } from '../../content/post/postGroup.ts';
	import { socialUserGroupIdMap } from '../common.ts';
	import UserMemberDetail from './UserMemberDetail.svelte';
	import { SK } from '#lib/shared/storage-keys.ts';

	let { userId }: { userId: string } = $props();
	const KEY = SK.admin.userMemberManageState;
	const ls = (k: string) => {
		try {
			return Number(localStorage.getItem(k) || 5);
		} catch {
			return 5;
		}
	};
	function init() {
		try {
			const s = sessionStorage.getItem(KEY);
			if (s) return JSON.parse(s);
		} catch {}
		const page = (k: string) => ({ pageNo: 1, pageSize: ls(k), orderBy: 'createdAt', orderByDesc: 'desc', groupKey: '', groupSubKey: '' });
		return { tab: 'POSTS', postPage: page(SK.admin.userManagePostPageSize), commentPage: page(SK.admin.userManageCommentPageSize) };
	}
	let st = $state<any>(init());
	let user = $state<any>(null);
	let posts = $state<any[] | null>(null);
	let postTotal = $state(0);
	let pm = $state<PostManageDrawer>();
	const groups = userPostGroups();

	function save(next: any) {
		st = next;
		try {
			sessionStorage.setItem(KEY, JSON.stringify(next));
		} catch {}
	}
	$effect(() => {
		const id = userId;
		untrack(async () => {
			const o = await api.get(`/service/admin/user/${id}`);
			if (!o.ok) return alert('회원정보 로딩 도중 오류가 발생하였습니다.');
			const oo = o.data;
			oo.Json = JSON.parse(oo.json || '{}');
			delete oo.json;
			oo.Json.imageId = `u_${oo.userId}_my`;
			user = oo;
		});
	});
	$effect(() => {
		const p = st.postPage;
		const s: any = { userId, pageNo: p.pageNo, pageSize: p.pageSize };
		if (p.groupKey) s.groupKey = p.groupKey;
		if (p.groupSubKey) s.groupSubKey = p.groupSubKey;
		untrack(() => getEsPosts(s)).then((r) => {
			posts = r.posts;
			postTotal = r.totalCount || r.posts.length;
		});
	});
	$effect(() => {
		const un = () => sessionStorage.removeItem(KEY);
		window.addEventListener('beforeunload', un);
		return () => window.removeEventListener('beforeunload', un);
	});

	const social = $derived(socialUserGroupIdMap[user?.groupId]);
	const title = $derived(`웹회원${user?.name ? ` - ${user.name} ${social && social != '이메일' ? `(${social})` : ''}` : ''}`);
	const groupOptions = [{ value: '', label: '전체' }, ...groups.map((g: any) => ({ value: g.groupKey, label: g.name || g.groupKey }))];
	const subItems = $derived((groups.find((g: any) => g.groupKey == st.postPage.groupKey) as any)?.items || []);
</script>

<PageContainer {title}>
	<Card sx={{ padding: 24 }}
		>{#if user}<UserMemberDetail {user} onChange={(u) => (user = u)} />{/if}</Card
	>

	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'end', justifyContent: 'space-between' }))}>
		<Tabs
			sx={{ marginTop: 16 }}
			value={st.tab}
			items={[{ value: 'POSTS', label: '게시글' }]}
			onchange={(t) => save({ ...st, tab: t, ...(t == 'POSTS' ? { postPage: { ...st.postPage, pageNo: 1 } } : { commentPage: { ...st.commentPage, pageNo: 1 } }) })}
		/>
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, marginBottom: 8 }))}>
			{#if st.tab === 'POSTS'}
				<Select
					value={st.postPage.groupKey || ''}
					options={groupOptions}
					sx={{ height: 40, minWidth: 140 }}
					onchange={(v) => (st = { ...st, postPage: { ...st.postPage, groupKey: v, groupSubKey: '', pageNo: 1 } })}
				/>
				{#if st.postPage.groupKey && subItems.length > 0}
					<Select
						value={st.postPage.groupSubKey || ''}
						options={[{ value: '', label: '전체' }, ...subItems.map((g: any) => ({ value: g.groupSubKey, label: g.groupSubKey }))]}
						sx={{ height: 40, minWidth: 140 }}
						onchange={(v) => (st = { ...st, postPage: { ...st.postPage, groupSubKey: v, pageNo: 1 } })}
					/>
				{/if}
			{/if}
		</div>
	</div>

	{#if st.tab === 'POSTS'}
		<Card sx={{ padding: 16, marginBottom: 16 }}><BoardListTable itemList={posts || []} onmanage={(p) => pm?.openDrawer(p)} /></Card>
		<TablePagination
			search={st.postPage}
			pageCount={Math.ceil(postTotal / st.postPage.pageSize) || 1}
			getData={({ pageNo, pageSize }) => {
				try {
					localStorage.setItem(SK.admin.userManagePostPageSize, String(pageSize));
				} catch {}
				save({ ...st, postPage: { ...st.postPage, pageNo, pageSize } });
			}}
		/>
	{/if}

	{#if activeNav()?.auth == 'RW'}
		<PostManageDrawer bind:this={pm} useCreateButton={false} onSave={(p: any) => (posts = (posts || []).map((o) => (o.postId == p.postId ? p : o)))} />
	{/if}
</PageContainer>
