<script lang="ts">
	/**
	 * 관리자 그룹 추가·수정 서랍(React setting/group/components/GroupManageDrawer) — 이름 · 설명 · 메모(기록) · 삭제(구성원 0명일 때만).
	 * 저장: 수정 PUT /service/admin/group/{groupId} · 추가 POST /service/admin/group(json 문자열, hostingId) · 삭제 DELETE.
	 * group 을 넘기면 열린다(새로 = { groupId: 0 }).
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { number } from '#lib/shared/format.ts';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import StandardField from '../ui/StandardField.svelte';
	import MemoHistory from '../member/MemoHistory.svelte';
	import { adminMembers } from '../member/adminMembers.svelte.ts';

	let {
		group: input,
		hostingId,
		onclose,
		onUpdated,
		onAdded,
		onDelete
	}: { group: any; hostingId: string; onclose: () => void; onUpdated?: (g: any) => void; onAdded?: (g: any) => void; onDelete?: (groupId: number) => void } = $props();
	const groupDefault = { groupId: 0, name: '', description: '', status: 0, json: {} as any };
	let g = $state<any>(null);
	let orig = '';
	let memo = $state('');
	let jsonOpen = $state(false);

	$effect(() => {
		const x = input;
		untrack(() => {
			memo = '';
			if (!x) return (g = null);
			g = x.groupId ? clone(x) : clone(groupDefault);
			orig = JSON.stringify(g);
			if (x.groupId) load(x.groupId);
		});
	});
	async function load(id: number) {
		const o = await api.get(`/service/admin/group/${id}`);
		if (!o.ok) return alert('상세정보 로딩 도중 오류가 발생하였습니다.');
		const oo = o.data;
		oo.json = typeof oo.json == 'string' ? JSON.parse(oo.json) : oo.json || {};
		orig = JSON.stringify(oo);
		g = oo;
	}
	const edited = $derived(!!g && (JSON.stringify(g) != orig || !!memo));
	const memberCount = $derived((adminMembers.list || []).filter((m) => m.groupId == g?.groupId).length);

	async function save() {
		const data: any = clone(g);
		if (memo) (data.json.memos ||= []).push({ dt: new Date(), memo });
		data.json = JSON.stringify(data.json) || '{}';
		if (g.groupId) {
			const o = await api.put(`/service/admin/group/${g.groupId}`, data);
			if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
			toast('정상적으로 수정되었습니다.');
			onUpdated?.(o.data);
			return onclose();
		}
		data.hostingId = hostingId;
		const o = await api.post('/service/admin/group', data);
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 추가되었습니다.');
		onAdded?.(o.data);
		onclose();
	}
	async function remove() {
		if (!(await confirm('그룹을 삭제하시겠습니까?'))) return;
		const o = await api.delete(`/service/admin/group/${g.groupId}`);
		if (!o.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		toast('정상적으로 삭제되었습니다.');
		onDelete?.(g.groupId);
		onclose();
	}
	async function close() {
		if (edited && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		onclose();
	}
	const h6 = 'MuiTypography-root MuiTypography-h6';
</script>

{#snippet label(name: string)}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<div class={cx(h6, toClass({ marginBottom: 8 }))}>{name}</div>
		<span class={cx('MuiTypography-root MuiTypography-body2', toClass({ marginLeft: 0.8, color: '#FB3B3B' }))}>*</span>
	</div>
{/snippet}

<Drawer open={!!input && !!g} anchor="right" onclose={close} paperSx={{ borderRadius: 0, userSelect: 'none', width: '50%' }}>
	{#if g}
		<h4 class={cx('MuiTypography-root MuiTypography-h4', toClass({ textAlign: 'center', height: 48, marginTop: 16 }))}>그룹 {g.groupId ? '수정' : '추가'}</h4>
		<div class={cx('MuiStack-root max', toClass({ paddingLeft: 24, paddingRight: 24, '& > :not(style) ~ :not(style)': { marginTop: 16 } }))}>
			<div class="MuiStack-root">
				{@render label('그룹이름')}
				<div class={cx('MuiStack-root', toClass({ flex: 1 }))}>
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between' }))}>
						<StandardField value={g.name} fullWidth placeholder="관리자 성명을 입력하세요" sx={{ marginTop: 8 }} oninput={(v) => (g = { ...g, name: v })} id="group_name" />
					</div>
				</div>
			</div>
			<div class="MuiStack-root">
				{@render label('설명')}
				<div class={cx('MuiStack-root', toClass({ flex: 1 }))}>
					<StandardField value={g.description} fullWidth placeholder="그룹 설명을 작성해주세요" oninput={(v) => (g = { ...g, description: v })} id="group_description" />
				</div>
			</div>
			<div class="MuiStack-root">
				<div class={cx('MuiStack-root', toClass({ width: 200, marginTop: 16 }))}><div class={cx(h6, toClass({ marginBottom: 8 }))}>메모작성</div></div>
				<StandardField value={memo} multiline minRows={3} fullWidth maxLength={100} placeholder="메모를 입력하세요" oninput={(v) => (memo = v)} id="group_memo">
					{#snippet end()}<span class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignSelf: 'end', marginBottom: -32, color: '#BDBDBD', fontSize: 14 }))}
							><strong>{number(memo.length) || 0}</strong>/100</span
						>{/snippet}
				</StandardField>
			</div>
			<div class="MuiStack-root"><MemoHistory memos={g?.json?.memos} /></div>
			{#if onDelete && g.groupId > 0 && adminMembers.list && memberCount == 0}
				<Button variant="contained" color="error" sx={{ width: 'fit-content' }} onclick={remove}>그룹 삭제</Button>
			{/if}
		</div>
		<hr class={cx('MuiDivider-root', toClass({ height: 100, border: 0, margin: 0, flexShrink: 0 }))} />
		<div
			class={cx(
				'MuiStack-root',
				toClass({ position: 'fixed', zIndex: 1, background: 'var(--zc-fff, #fff)', bottom: 0, left: '50%', width: '50%', '@media (max-width:899.95px)': { left: 0, width: '100%' } })
			)}
		>
			<AuthButton variant="contained" fullWidth sx={{ height: 48, borderRadius: 0 }} disabled={!edited} onclick={save}>저장</AuthButton>
			{#if admin.profile?.isDeveloper}<IconButton sx={{ position: 'fixed', right: 10, bottom: 5, color: '#fff' }} aria-label="JSON 편집" onclick={() => (jsonOpen = true)}
					><Icon name="data_object" /></IconButton
				>{/if}
		</div>
		<JsonEditDialog bind:open={jsonOpen} value={g} onChange={(o) => (g = o)} title="그룹 JSON" />
	{/if}
</Drawer>
