<script lang="ts">
	/**
	 * 환경 설정 › 게시판 관리(React setting/meta/common/PostGroupManageForm) — POST_GROUPS 메타의 관리자/사용자 포스트 그룹.
	 * 관리자·사용자 포스트 그룹 = 묶음(SettingGroup — 맨 위 카드·접기) 둘(2026-10-04 대표 — 예전엔 하위 탭).
	 * 묶음마다 그룹 카드(글 수 배지·바로가기·삭제)·순서 변경·추가·저장(지금 서버 값을 다시 읽어 그 묶음만 바꿔 PUT).
	 * 톱니·「추가」 = 게시판 설정 창(post-group/PostGroupSettingModal — groupKey·구분·게시판 명), { } = 그룹 JSON 직접 고치기(개발자만).
	 */
	import { onMount, untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import SortDialog from '../../ui/SortDialog.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { isSuperAdmin } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import SettingGroup from '../ui/SettingGroup.svelte';
	import PostGroupSettingModal from './post-group/PostGroupSettingModal.svelte';

	let { meta, hostingId, onChange }: { meta: any; hostingId: string; onChange: (m: any) => void } = $props();
	const groupDefault = { groupKey: '', name: '', dataColumn: 'jsonUser', feature: { useWrite: true, useEstimation: false } };
	type K = 'adminPostGroups' | 'userPostGroups';
	const GROUPS: { k: K; title: string; desc: string }[] = [
		{ k: 'adminPostGroups', title: '관리자 포스트 그룹', desc: '운영자가 쓰는 게시판 — 공지·고객센터 등' },
		{ k: 'userPostGroups', title: '사용자 포스트 그룹', desc: '회원이 쓰는 게시판 — 기록·게임 기록 등' }
	];
	// 지금 고치는 묶음 — 추가·설정 창·JSON 창이 이 묶음에 넣는다(버튼을 누를 때 맞춘다)
	let tab = $state<K>('adminPostGroups');
	let item = $state<any>({ adminPostGroups: [], userPostGroups: [] });
	let summary = $state<Record<string, Record<string, number>>>({});
	let edit = $state<{ group: any; isNew: boolean } | null>(null);
	let setting = $state<{ group: any; isNew: boolean } | null>(null);
	let settingOpen = $state(false);
	const openSetting = (group: any, isNew: boolean) => ((setting = { group, isNew }), (settingOpen = true));
	let sortOpen = $state<Record<K, boolean>>({ adminPostGroups: false, userPostGroups: false });
	const sup = $derived(isSuperAdmin(admin.profile));
	$effect(() => {
		const m = meta;
		untrack(() => (item = clone(m?.json || { adminPostGroups: [], userPostGroups: [] })));
	});
	const diffsOf = (k: K) => diffList(meta?.json?.[k] || [], item?.[k] || []);
	onMount(async () => {
		const o = await api.get('/service/admin/post/summary');
		if (!o.ok) return alert('데이터 로딩 도중 오류가 발생하였습니다.');
		const m: Record<string, Record<string, number>> = {};
		for (const x of o.data || []) (m[x.hostingId] ||= {})[x.groupKey] = x.cnt;
		summary = m;
	});
	const setTab = (list: any[], k: K = tab) => (item = { ...item, [k]: list });
	async function save(tab: K) {
		const g = await api.get(`/service/admin/meta/${meta.metaKey}?hostingId=${hostingId}`);
		if (!g.ok) return alert('처리 도중 오류가 발생하였습니다.');
		const oo = g.data;
		oo[tab] = item[tab];
		const o = await api.put(`/service/admin/meta/${meta.metaKey}?hostingId=${hostingId}`, oo);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		onChange(clone({ ...meta, json: oo }));
		toast('정상적으로 저장되었습니다.');
	}
	async function remove(o: any, k: K) {
		if (!(await confirm({ body: `${o.groupKey} ${o.name} 그룹을 삭제하시겠습니까?` }))) return;
		setTab(
			item[k].filter((x: any) => x.groupKey != o.groupKey),
			k
		);
	}
	const gray = 'var(--gt-700, #909090)';
	const badge = (n: number) =>
		cx('MuiBadge-badge MuiBadge-standard MuiBadge-anchorOriginTopRight MuiBadge-anchorOriginTopRightRectangular MuiBadge-overlapRectangular MuiBadge-colorPrimary', !n && 'MuiBadge-invisible');
</script>

<div class="grps">
	{#each GROUPS as g (g.k)}
		{@const k = g.k}
		{@const diffs = diffsOf(k)}
		<SettingGroup title={g.title} desc={g.desc} foldKey="postGroups:{k}">
			<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
				{#each item?.[k] || [] as o (o.groupKey)}
					{@const cnt = summary?.[hostingId]?.[o.groupKey] || 0}
					<Card sx={{ padding: 16 }}>
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 4, justifyContent: 'space-between', position: 'relative' }))}>
							<div class={cx('MuiTypography-root MuiTypography-h6', toClass({ display: 'flex', flexDirection: 'row', alignItems: 'center', color: gray }))}>
								{#if o.gubun}<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ marginRight: 8 }))}>{o.gubun}</span>{/if}
								{#if admin.profile?.isDeveloper}
									<IconButton sx={{ width: 'fit-content', padding: 0, marginRight: 4 }} aria-label="JSON" onclick={() => ((tab = k), (edit = { group: o, isNew: false }))}
										><Icon name="data_object" size="Small" /></IconButton
									>
								{/if}
								{#if cnt > 0}<span class={cx('MuiBadge-root', toClass({ marginRight: 16 }))}><Icon name="article" sx={{ color: 'var(--gt-500, #c7c7c7)' }} /><span class={badge(cnt)}>{cnt}</span></span
									>{/if}
								<div class={cx('MuiStack-root', toClass({ color: o.isBoardStore ? 'color-mix(in srgb, var(--theme-color-main) 80%, transparent)' : 'inherit' }))}>{o.name}</div>
								<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ marginLeft: 4 }))}>{o.groupKey} </span>
								<IconButton size="Small" href={`/admin/${admin.profile?.hostingId ? 'board' : 'post'}/${o.groupKey}${hostingId ? `?hostingId=${hostingId}` : ''}`} target="_blank"
									><Icon name="open_in_new" size="Small" sx={{ color: 'var(--gt-600, #b0b0b0)' }} /></IconButton
								>
							</div>
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}>
								<p class="MuiTypography-root MuiTypography-body2">{o.name || '설명없음'}</p>
								<span class="MuiBadge-root">
									<IconButton size="Small" aria-label="게시판 설정" onclick={() => ((tab = k), openSetting(o, false))}><Icon name="settings" size="Small" /></IconButton>
									<span class={badge(o.items?.length || 0)}>{o.items?.length || ''}</span>
								</span>
								<IconButton size="Small" aria-label="그룹 삭제" onclick={() => remove(o, k)}><Icon name="close" size="Small" /></IconButton>
							</div>
						</div>
					</Card>
				{/each}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', marginTop: 16 }))}>
					<div class="MuiStack-root">
						{#if sup}<SortDialog objectKey="groupKey" items={item?.[k] || []} bind:open={sortOpen[k]} onChange={(os) => setTab(os, k)} />{/if}
					</div>
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8 }))}>
						{#if sup}<Button variant="outlined" onclick={() => ((tab = k), openSetting({ ...clone(groupDefault), dataColumn: k == 'adminPostGroups' ? 'json' : 'jsonUser' }, true))}>추가</Button>{/if}
						<Button variant="contained" disabled={!diffs.length} onclick={() => save(k)}>저장</Button>
					</div>
				</div>
			</div>
			{#if diffs.length && admin.profile?.isDeveloper}<div class={cx('MuiStack-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>{/if}
		</SettingGroup>
	{/each}
</div>

<PostGroupSettingModal
	bind:open={settingOpen}
	postGroup={setting?.group}
	isSuperAdmin={sup}
	onChange={(o) => {
		if (!setting) return;
		if (setting.isNew) setTab([...(item[tab] || []), o]);
		else {
			const key = setting.group.groupKey;
			setTab(item[tab].map((x: any) => (x.groupKey == key ? o : x)));
		}
	}}
/>

<JsonEditDialog
	open={!!edit}
	onclose={() => (edit = null)}
	value={edit?.group}
	title={edit?.isNew ? '게시판 그룹 추가(JSON)' : `${edit?.group?.name || ''} 게시판 설정(JSON)`}
	onChange={(o) => {
		if (!edit) return;
		if (edit.isNew) setTab([...(item[tab] || []), o]);
		else {
			const key = edit.group.groupKey;
			setTab(item[tab].map((x: any) => (x.groupKey == key ? o : x)));
		}
		edit = null;
	}}
/>

<style>
	.grps {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
</style>
