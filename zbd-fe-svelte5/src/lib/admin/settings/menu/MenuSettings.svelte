<script lang="ts">
	/**
	 * 메뉴 설정(React pages/admin/setting/menu) — 대메뉴 카드(손잡이로 순서·아이콘·이름·링크·설명·특수 메뉴·API) + 하위 메뉴 줄(순서·이름·경로·설명·삭제·API).
	 * 바꾼 것을 한꺼번에 저장: PUT /service/admin/menus(배열 그대로) · 대메뉴 삭제 DELETE ?menuId=. 저장 전 바뀐 값 미리보기.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import SvgIcon from '../ui/SvgIcon.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin, POST_MENU_SPECIALS } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/admin/kit/DiffPreview.svelte';
	import { diffHtml } from '#lib/admin/kit/diff.ts';
	import MenuIconSelector from '../ui/MenuIconSelector.svelte';
	import OutlinedSelect from '../ui/OutlinedSelect.svelte';
	import SwitchField from '../ui/SwitchField.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import { moveItem, sortable } from '#lib/shared/ui/sortable.ts';
	import MenuApiDrawer from './MenuApiDrawer.svelte';

	const MENU_SPECIALS = [
		{ value: '', label: '특수 메뉴 없음' },
		{ value: 'mail', label: '웹메일 폴더' },
		{ value: 'adminPosts', label: '관리자 포스트(게시판 목록)' },
		{ value: 'userPosts', label: '사용자 포스트(게시판 목록)' }
	];
	const hostingId = admin.profile?.hostingId || '';
	let itemList = $state<any[] | null>(null);
	let orig = $state<any[]>([]);
	let max = $state(0);
	let apiMenu = $state<any>(null);
	// 게시판 메뉴 노출 — POST_GROUPS 의 각 게시판 hideInMenu(특수 메뉴 「관리자/사용자 포스트」 의 하위 메뉴에서 뺄지). 메뉴 저장 때 같이 저장
	type PG = { adminPostGroups: any[]; userPostGroups: any[] };
	let postGroups = $state<PG>({ adminPostGroups: [], userPostGroups: [] });
	let postGroupsOrig = $state<PG>({ adminPostGroups: [], userPostGroups: [] });
	const hiddenKeys = (pg: PG) =>
		['adminPostGroups', 'userPostGroups']
			.flatMap((k) => ((pg as any)[k] || []).filter((g: any) => g?.hideInMenu).map((g: any) => `${k}:${g.groupKey}`))
			.sort()
			.join(',');
	const boardsDirty = $derived(hiddenKeys(postGroups) != hiddenKeys(postGroupsOrig));
	function setBoardShown(key: 'adminPostGroups' | 'userPostGroups', groupKey: string, shown: boolean) {
		postGroups = { ...postGroups, [key]: postGroups[key].map((g: any) => (g.groupKey == groupKey ? { ...g, hideInMenu: shown ? undefined : true } : g)) };
	}
	async function loadPostGroups() {
		const o = await api.get(`/service/admin/meta/POST_GROUPS?hostingId=${hostingId}`);
		if (!o.ok) return;
		const pg = { adminPostGroups: o.data?.adminPostGroups || [], userPostGroups: o.data?.userPostGroups || [] };
		postGroups = clone(pg);
		postGroupsOrig = clone(pg);
	}
	/** 지금 서버 값을 다시 읽어 hideInMenu 칸만 바꿔 저장(게시판 설정의 다른 칸을 옛 값으로 덮지 않게) */
	async function savePostGroups() {
		const g = await api.get(`/service/admin/meta/POST_GROUPS?hostingId=${hostingId}`);
		if (!g.ok) return false;
		const oo = g.data;
		for (const k of ['adminPostGroups', 'userPostGroups'] as const) {
			const hidden = new Set(postGroups[k].filter((x: any) => x.hideInMenu).map((x: any) => x.groupKey));
			oo[k] = (oo[k] || []).map((x: any) => {
				const { hideInMenu: _h, ...rest } = x;
				return hidden.has(x.groupKey) ? { ...rest, hideInMenu: true } : rest;
			});
		}
		const o = await api.put(`/service/admin/meta/POST_GROUPS?hostingId=${hostingId}`, oo);
		if (!o.ok) return false;
		postGroupsOrig = clone(postGroups);
		return true;
	}

	function onUpdateMenus(menus: any[]) {
		let mx = 0;
		for (const o of menus) {
			if (o.menuId > mx) mx = o.menuId;
			for (const s of o.items || []) if (s.menuId > mx) mx = s.menuId;
		}
		max = mx + 1;
	}
	onMount(async () => {
		const o = await api.get('/service/admin/menus');
		if (!o.ok) return alert('메뉴정보 로딩 도중 오류가 발생하였습니다.');
		const list = (o.data || [])
			.map((m: any) => {
				m.hostingId ||= '';
				for (const s of m.items || []) s.hostingId ||= '';
				return m;
			})
			.filter((m: any) => m.hostingId == hostingId)
			.map((m: any) => ({ ...m, items: m.items || [] }));
		orig = clone(list);
		itemList = list;
		onUpdateMenus(list);
		loadPostGroups();
	});

	const helpers = $derived.by(() => {
		const h: Record<number, { name?: string; path?: string }> = {};
		const add = (id: number, k: 'name' | 'path', v: string) => ((h[id] ||= {})[k] = v);
		for (const o of itemList || []) {
			if (!String(o.name || '').trim()) add(o.menuId, 'name', '메뉴명은 필수입니다.');
			if (o.items?.length) {
				for (const s of o.items) {
					if (!String(s.name || '').trim()) add(s.menuId, 'name', '메뉴명은 필수입니다.');
					if (!String(s.path || '').trim()) add(s.menuId, 'path', '하위 링크경로는 필수입니다.');
				}
			} else if (!o.path && !POST_MENU_SPECIALS[o.special || '']) add(o.menuId, 'path', '서브메뉴가 없는 경우 필수값 입니다.');
		}
		return h;
	});
	const isValid = $derived(!Object.values(helpers).some((x) => x.name || (x.path && x.path != '서브메뉴가 없는 경우 필수값 입니다.')));
	const diffs = $derived(itemList && isValid ? diffList(orig, itemList) : []);
	// 저장 전 바뀐 내용(React 판과 같은 jsondiffpatch 화면) — 메뉴 + 게시판 메뉴 노출. 화면 전용 칸(isNew)·API 개수(API 서랍에서 바로 저장됨)는 빼고 비교
	// 배열이면 이름을 바꾼 메뉴가 「지우고 새로 넣음」 으로 보여 메뉴 번호로 묶는다(순서는 order 칸)
	const byId = (list: any[], sub: boolean): Record<string, any> =>
		Object.fromEntries(
			(list || []).map(({ isNew: _n, api_count: _a, chosen: _c, selected: _s, items, ...o }: any, i: number) => [
				String(o.menuId),
				{ ...o, order: i, ...(sub ? {} : { items: byId(items || [], true) }) }
			])
		);
	const forDiff = (list: any[], pg: PG) => ({
		menus: byId(list, false),
		boardsHiddenInMenu: Object.fromEntries((hiddenKeys(pg) ? hiddenKeys(pg).split(',') : []).map((k) => [k, true]))
	});
	let diffHTML = $state('');
	$effect(() => {
		const left = forDiff(orig, postGroupsOrig);
		const right = forDiff($state.snapshot(itemList) || [], $state.snapshot(postGroups));
		if (!itemList || !isValid) return void (diffHTML = '');
		let stale = false;
		diffHtml(left, right).then((h) => !stale && (diffHTML = h));
		return () => (stale = true);
	});

	const setMenu = (id: number, patch: any) => (itemList = itemList!.map((o) => (o.menuId == id ? { ...o, ...patch } : o)));
	const setSubs = (id: number, items: any[]) => setMenu(id, { items });
	const setSub = (pid: number, sid: number, patch: any) =>
		setSubs(
			pid,
			itemList!.find((o) => o.menuId == pid)!.items.map((s: any) => (s.menuId == sid ? { ...s, ...patch } : s))
		);

	function afterSave(removeId = 0) {
		const list = itemList!.filter((o) => o.menuId != removeId).map((o) => ({ ...o, items: (o.items || []).filter((s: any) => !s.isDeleted).map(({ isNew: _n, ...s }: any) => s) }));
		orig = clone(list);
		itemList = clone(list);
		toast('정상적으로 저장되었습니다.');
	}
	async function save() {
		if (!(await confirm('저장하시겠습니까?'))) return;
		if (boardsDirty && !(await savePostGroups())) return alert('게시판 노출 저장 도중 오류가 발생하였습니다.');
		if (!diffs.length) return toast('정상적으로 저장되었습니다. 왼쪽 메뉴는 새로고침하면 바뀝니다.');
		const o = await api.put(`/service/admin/menus${admin.profile?.isMaster ? `?hostingId=${hostingId}` : ''}`, itemList);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		onUpdateMenus(o.data?.admin?.menus || []);
		afterSave();
	}
	async function removeMenu(o: any) {
		if (!(await confirm(`"${o.name || '이름없음'}" 대메뉴를 삭제하시겠습니까?`))) return;
		const r = await api.delete(`/service/admin/menus?menuId=${o.menuId}${admin.profile?.isMaster ? `&hostingId=${hostingId}` : ''}`);
		if (!r.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		onUpdateMenus(r.data?.admin?.menus || []);
		afterSave(o.menuId);
	}
	async function removeSub(pid: number, s: any) {
		if (!(await confirm({ body: '서브메뉴항목을 삭제하시겠습니까?' }))) return;
		const items = itemList!.find((o) => o.menuId == pid)!.items;
		if (s.isNew)
			setSubs(
				pid,
				items.filter((x: any) => x.menuId != s.menuId)
			);
		else {
			setSubs(
				pid,
				items.map((x: any) => (x.menuId == s.menuId ? { ...x, isDeleted: true } : x))
			);
			toast('저장시 삭제됩니다.');
		}
	}
	// 메뉴 옮기기(2026-10-05) — 대메뉴(하위 없음)를 다른 대메뉴 아래로 · 하위 메뉴를 다른 대메뉴 아래나 대메뉴로. 저장하면 서버가 보낸 트리대로 상위를 바꾼다
	const moveTargets = (selfId: number, fromParent: number) => [
		{ value: '', label: '옮기기…' },
		...(fromParent ? [{ value: '0', label: '대메뉴로' }] : []),
		...(itemList || []).filter((o) => o.menuId != selfId && o.menuId != fromParent && !o.special).map((o) => ({ value: String(o.menuId), label: `${o.name || '(이름 없음)'} 아래로` }))
	];
	function moveMenu(id: number, fromParent: number, to: string) {
		if (to === '' || !itemList) return;
		const toParent = Number(to);
		let node: any;
		let list = itemList.map((o) => ({ ...o, items: [...(o.items || [])] }));
		if (fromParent == 0) {
			node = list.find((o) => o.menuId == id);
			list = list.filter((o) => o.menuId != id);
		} else {
			const p = list.find((o) => o.menuId == fromParent)!;
			node = p.items.find((x: any) => x.menuId == id);
			p.items = p.items.filter((x: any) => x.menuId != id);
		}
		if (!node) return;
		const { items: _i, ...rest } = node;
		if (toParent == 0) list.push({ ...rest, parentId: 0, items: [], api_count: rest.api_count || 0 });
		else list.find((o) => o.menuId == toParent)!.items.push({ ...rest, parentId: toParent });
		itemList = list;
	}
	const setApiCount = (menuId: number, n: number) =>
		(itemList = itemList!.map((o) => ({ ...o, ...(o.menuId == menuId ? { api_count: n } : {}), items: (o.items || []).map((s: any) => (s.menuId == menuId ? { ...s, api_count: n } : s)) })));
	const err = 'MuiTypography-root MuiTypography-subtitle2 adm-err';
</script>

{#snippet apiBtn(o: any)}
	<Button size="small" onclick={() => (apiMenu = o)}>
		<Icon name="compare_arrows" sx={{ marginRight: 4 }} />API{#if o.api_count}<span class="api-n">{o.api_count}</span>{/if}
	</Button>
{/snippet}

{#if itemList}
	<div class="menus" use:sortable={{ handle: '.handle', onmove: (f, t) => (itemList = moveItem(itemList!, f, t)) }}>
		{#each itemList as o (o.menuId)}
			<Card sx={{ marginBottom: 16, padding: 16, '@media (min-width:900px)': { padding: '20px 24px' } }}>
				<div class="top">
					<div class="title">
						<SvgIcon name="DragIndicator" class="handle" sx={{ cursor: 'grab', color: '#999' }} />
						<MenuIconSelector icon={o.icon} {hostingId} onChange={(icon) => setMenu(o.menuId, { icon })} />
						<div class="name">
							<OutlinedField
								fullWidth
								value={o.name || ''}
								placeholder="대메뉴 이름"
								maxLength={30}
								error={!!helpers[o.menuId]?.name}
								inputSx={{ fontWeight: 600 }}
								oninput={(v) => setMenu(o.menuId, { name: v })}
							/>
						</div>
					</div>
					<div class="fields">
						<OutlinedField fullWidth value={o.path || ''} placeholder={o.items?.length ? '링크 (서브메뉴가 있으면 비워 둠)' : '링크 /admin/…'} oninput={(v) => setMenu(o.menuId, { path: v })} />
						<OutlinedField fullWidth value={o.description || ''} placeholder="메뉴 설명 (페이지 서브타이틀)" oninput={(v) => setMenu(o.menuId, { description: v })} />
						<div class="acts">
							{#if !o.items?.length}
								<IconButton aria-label="대메뉴 삭제" onclick={() => removeMenu(o)}><Icon name="delete" /></IconButton>
							{:else}<span></span>
							{/if}
							{#if !o.items?.length && !o.special}
								<div class="mv"><OutlinedSelect size="small" value="" options={moveTargets(o.menuId, 0)} onchange={(v) => moveMenu(o.menuId, 0, v)} /></div>
							{/if}
							{@render apiBtn(o)}
						</div>
					</div>
				</div>
				{#if helpers[o.menuId]?.name || helpers[o.menuId]?.path}
					<h6 class={err}>{helpers[o.menuId]?.name || helpers[o.menuId]?.path}</h6>
				{/if}
				{#if !o.items?.length}
					<!-- 특수 메뉴 — 하위 메뉴를 직접 두지 않은 대메뉴만. 링크 칸 아래 줄에 둬서 위 줄의 설명 칸이 좁아지지 않게 -->
					<div class="special">
						<span></span>
						<div>
							<OutlinedSelect size="small" fullWidth value={o.special || ''} options={MENU_SPECIALS} onchange={(v) => setMenu(o.menuId, { special: v || null })} />
							<p class="note">
								{#if POST_MENU_SPECIALS[o.special || '']}
									하위 메뉴 = 게시판 설정의 {o.special == 'adminPosts' ? '관리자' : '사용자'} 게시판 — 스위치로 메뉴 노출을 고릅니다.
								{:else if o.special == 'mail'}
									하위 메뉴 = 내 웹메일 폴더.
								{:else}
									특수 메뉴 — 하위 메뉴를 자동으로 채웁니다(웹메일 폴더·게시판 목록).
								{/if}
							</p>
						</div>
					</div>
				{/if}

				{#if POST_MENU_SPECIALS[o.special || '']}
					{@const pk = POST_MENU_SPECIALS[o.special || '']}
					<div class="boards">
						{#each postGroups[pk] as g (g.groupKey)}
							<div class={cx('board', (g.use === false || g.hideInMenu) && 'off')}>
								<SwitchField checked={!g.hideInMenu} disabled={g.use === false} label={g.name || g.groupKey} onchange={(v) => setBoardShown(pk, g.groupKey, v)} />
								<code>{g.groupKey}</code>
								{#if g.use === false}<span class="tag">사용 안 함</span>{/if}
							</div>
						{:else}
							<p class="note">게시판 설정에 게시판이 없습니다.</p>
						{/each}
					</div>
				{/if}
				{#if o.items?.length}
					<div class="subs" use:sortable={{ handle: '.subItemHandle', onmove: (f, t) => setSubs(o.menuId, moveItem(o.items, f, t)) }}>
						{#each o.items as s (s.menuId)}
							<div>
								<div class={cx('sub', s.isDeleted && 'deleted')}>
									<SvgIcon name="DragIndicator" class="subItemHandle" sx={{ cursor: 'grab', color: '#bbb' }} />
									<span class="sic"><MenuIconSelector compact btnText="하위 메뉴 아이콘(선택)" icon={s.icon || ''} {hostingId} onChange={(icon) => setSub(o.menuId, s.menuId, { icon })} /></span>
									<OutlinedField fullWidth value={s.name} placeholder="메뉴명" maxLength={30} error={!!helpers[s.menuId]?.name} oninput={(v) => setSub(o.menuId, s.menuId, { name: v })} />
									<OutlinedField fullWidth value={s.path} placeholder="/admin/…" maxLength={100} error={!!helpers[s.menuId]?.path} oninput={(v) => setSub(o.menuId, s.menuId, { path: v })} />
									<OutlinedField fullWidth value={s.description || ''} placeholder="메뉴 설명 (페이지 서브타이틀)" oninput={(v) => setSub(o.menuId, s.menuId, { description: v })} />
									<div class="acts">
										<IconButton aria-label="서브메뉴 삭제" onclick={() => removeSub(o.menuId, s)}><SvgIcon name="Clear" /></IconButton>
										{@render apiBtn(s)}
										{#if !s.isNew}<div class="mv"><OutlinedSelect size="small" value="" options={moveTargets(s.menuId, o.menuId)} onchange={(v) => moveMenu(s.menuId, o.menuId, v)} /></div>{/if}
									</div>
								</div>
								{#if helpers[s.menuId]?.name || helpers[s.menuId]?.path}<h6 class={cx(err, 'sub-err')}>{helpers[s.menuId]?.name || helpers[s.menuId]?.path}</h6>{/if}
							</div>
						{/each}
					</div>
				{/if}
				{#if !POST_MENU_SPECIALS[o.special || '']}
					<AuthButton
						variant="outlined"
						size="small"
						sx={{ marginTop: 12, width: 'fit-content' }}
						onclick={() => {
							const no = max;
							max = no + 1;
							setSubs(o.menuId, [...o.items, { menuId: no, name: '', parentId: o.menuId, path: '', sortNo: 0, isNew: true }]);
						}}><SvgIcon name="Add" size="Small" /> 서브메뉴 추가</AuthButton
					>
				{/if}
			</Card>
		{/each}
	</div>
	{#if itemList.length == 0}<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 48, marginBottom: 48 }))}>등록된 메뉴가 없습니다.</p>{/if}
{/if}

<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between' }))}>
	<Button
		variant="outlined"
		sx={{ width: 'fit-content' }}
		onclick={() => {
			const no = max;
			max = no + 1;
			itemList = [...(itemList || []), { menuId: no, parentId: 0, name: '', action: '', icon: '', items: [], api_count: 0 }];
		}}><SvgIcon name="Add" size="Small" /> 대 메뉴 추가</Button
	>
	<Button variant="contained" sx={{ width: 'fit-content' }} disabled={!diffs.length && !boardsDirty} onclick={save}>저장</Button>
</div>

<MenuApiDrawer menu={apiMenu} {hostingId} onclose={() => (apiMenu = null)} onChange={(id, apis) => setApiCount(id, apis.length)} />

{#if diffHTML && admin.profile?.isDeveloper}<div class={toClass({ marginTop: 16 })}><DiffPreview {diffHTML} /></div>{/if}

<style>
	/* 대메뉴 줄: [손잡이·아이콘·이름] [링크 · 설명 · 특수 메뉴 · 삭제/API] — 하위 줄도 같은 칸 폭을 써서 세로로 맞는다 */
	.top {
		display: grid;
		grid-template-columns: 280px 1fr;
		gap: 16px;
		align-items: center;
	}
	.title {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.name {
		flex: 1;
		min-width: 0;
	}
	.fields {
		display: grid;
		grid-template-columns: 220px 1fr auto;
		gap: 12px;
		align-items: center;
	}
	/* 삭제 · API — 칸 폭을 고정해 줄마다 같은 자리에(API 개수 표시가 있어도 삭제 버튼이 밀리지 않게) */
	.acts {
		display: grid;
		grid-template-columns: 40px 96px auto;
		align-items: center;
		justify-items: start;
		gap: 4px;
	}
	.special {
		display: grid;
		grid-template-columns: 280px 1fr;
		gap: 16px;
		margin-top: 8px;
	}
	.special > div {
		display: grid;
		grid-template-columns: 220px 1fr;
		gap: 12px;
		align-items: center;
	}
	.subs {
		margin-top: 16px;
		padding-top: 12px;
		border-top: 1px solid var(--zc-eee, #eee);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.sub {
		display: grid;
		grid-template-columns: 24px 28px 248px 220px 1fr auto;
		gap: 12px;
		align-items: center;
		border-radius: 6px;
	}
	.sic {
		display: inline-flex;
		justify-content: center;
	}
	.mv {
		width: 150px;
	}
	.sub.deleted {
		background: rgba(255, 165, 0, 0.25);
	}
	.sub-err {
		margin: -4px 0 0 36px;
	}
	.boards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 4px 16px;
		margin: 12px 0 0 296px;
		padding-top: 12px;
		border-top: 1px solid var(--zc-eee, #eee);
	}
	.board {
		display: flex;
		align-items: center;
		gap: 6px;
		min-width: 0;
	}
	.board code {
		font-size: 11px;
		color: var(--zc-999, #999);
	}
	.board.off :global(.MuiFormControlLabel-label) {
		color: var(--zc-999, #999);
	}
	.tag {
		font-size: 11px;
		color: var(--zc-999, #999);
		border: 1px solid var(--zc-ddd, #ddd);
		border-radius: 4px;
		padding: 0 4px;
	}
	.api-n {
		margin-left: 6px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 9px;
		background: var(--theme-color-primary, #1976d2);
		color: #fff;
		font-size: 11px;
		line-height: 18px;
		text-align: center;
	}
	.note {
		margin: 0;
		font-size: 13px;
		color: var(--zc-888, #888);
	}
	@media (max-width: 1199px) {
		.top,
		.special {
			grid-template-columns: 1fr;
		}
		.special > span {
			display: none;
		}
		.boards {
			margin-left: 0;
		}
		.sub {
			grid-template-columns: 24px 28px 1fr 1fr;
		}
		.sub > :global(:nth-child(5)) {
			grid-column: 3 / 5;
		}
		.sub > .acts {
			grid-column: 3 / 5;
		}
	}
	@media (max-width: 899px) {
		.fields,
		.special > div {
			grid-template-columns: 1fr;
		}
		.sub {
			grid-template-columns: 24px 1fr;
			padding-bottom: 8px;
			border-bottom: 1px dashed var(--zc-eee, #eee);
		}
		.sub > :global(*:not(:first-child)) {
			grid-column: 2;
		}
	}
</style>
