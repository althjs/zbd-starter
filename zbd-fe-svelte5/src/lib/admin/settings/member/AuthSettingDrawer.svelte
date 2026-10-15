<script lang="ts">
	/**
	 * 그룹 메뉴 권한(React setting/components/AuthSettingDrawer) — 메뉴마다 읽기/쓰기·읽기·권한없음.
	 * 그룹 관리에서 열면 그룹 권한을 저장(PUT /service/admin/group/auth/{groupId}),
	 * 관리자 편집에서 열면(member) 그 사람의 개별 권한(authUser)만 골라 onChange(items, 그룹 원본) 로 돌려준다(저장은 관리자 편집이).
	 *   <AuthSettingDrawer group={g} {menus} hostingId={c} onclose={…} />   group 이 있으면 열린다
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import RadioGroup from '../ui/RadioGroup.svelte';

	let {
		group,
		menus,
		hostingId,
		member = null,
		onChange = null,
		onclose
	}: { group: any; menus: any[] | null; hostingId: string; member?: any; onChange?: ((items: any[], orig: any[]) => void) | null; onclose: () => void } = $props();

	const descMap: Record<string, string> = { R: '읽기', RW: '읽기/쓰기' };
	let items = $state<any[] | null>(null);
	let orig: any[] = [];
	let jsonOpen = $state(false);
	const k = $derived(member ? 'authUser' : 'auth');

	$effect(() => {
		const g = group;
		untrack(() => {
			items = null;
			if (g) load(g.groupId);
		});
	});
	async function load(groupId: number) {
		const o = await api.get(`/service/admin/group/auth/${groupId}?hostingId=${hostingId}`);
		if (!o.ok) return alert('상세정보 로딩 도중 오류가 발생하였습니다.');
		const authList: any[] = o.data || [];
		const list = clone(menus || [])
			.filter((m: any) => !(m.hostingId !== hostingId && m.hostingId !== admin.profile?._hostingId))
			.map((m: any) => {
				m.auth = authList.find((x) => x.menuId == m.menuId)?.auth || '';
				if (member) m.authUser = (member?.json?.authMenus || []).find((x: any) => x.menuId == m.menuId)?.auth || m.auth;
				for (const s of m.items || []) {
					s.auth = authList.find((x) => x.menuId == s.menuId)?.auth || '';
					if (member) s.authUser = (member?.json?.authMenus || []).find((x: any) => x.menuId == s.menuId)?.auth || s.auth;
				}
				return m;
			});
		orig = clone(list);
		items = list;
	}
	const diffs = $derived(items ? diffList(orig, items) : []);

	async function save() {
		if (!items) return;
		const out: any[] = [];
		for (const m of items) {
			out.push({ menuId: m.menuId, auth: m[k] || '' });
			for (const s of m.items || []) out.push({ menuId: s.menuId, auth: s[k] || '' });
		}
		if (onChange) {
			onChange(out, orig);
			onclose();
			return;
		}
		const o = await api.put(`/service/admin/group/auth/${group.groupId}`, out);
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 수정되었습니다.');
		onclose();
	}
	async function close() {
		if (diffs.length && !member && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		onclose();
	}
	function set(m: any, v: string) {
		m[k] = v;
		items = [...items!];
	}
	const rw = [
		{ value: 'RW', label: '읽기/쓰기' },
		{ value: 'R', label: '읽기' }
	];
</script>

<Drawer open={!!group} anchor="right" onclose={close} paperSx={{ borderRadius: 0, userSelect: 'none', width: '50%' }}>
	<h4 class={cx('MuiTypography-root MuiTypography-h4', toClass({ textAlign: 'center', height: 48, marginTop: 16 }))}>
		{group?.name || ''} 그룹{member?.name ? `을 포함한 ${member.name}님 개별` : ''} 권한 관리
	</h4>
	<div class={cx('MuiStack-root max', toClass({ paddingLeft: 24, paddingRight: 24, '& > :not(style) ~ :not(style)': { marginTop: 16 } }))}>
		{#each items || [] as m (m.menuId)}
			{@const sub = (m.items || []).length > 0}
			<Card sx={{ padding: 24 }}>
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }))}>
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 8 }))}>
						<Icon name={m.icon || ''} />
						<div class="MuiTypography-root MuiTypography-h6">{m.name}</div>
					</div>
					<div class="MuiFormControl-root">
						<RadioGroup
							value={m[k] || ''}
							disabled={!!member && m.auth == 'RW'}
							options={[...rw.map((r) => ({ ...r, disabled: sub })), { value: '', label: sub ? '서브권한' : '권한없음' }]}
							onchange={(v) => set(m, v)}
						/>
						{#if member && m.auth}<h6 class="MuiTypography-root MuiTypography-subtitle2">그룹 권한: {descMap[m.auth]}</h6>{/if}
					</div>
				</div>
				{#if sub}
					<div class={cx('MuiStack-root', toClass({ marginTop: 16 }))}>
						{#each m.items as s (s.menuId)}
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', gap: 8 }))}>
								<div class="MuiStack-root">{s.name}</div>
								<div class="MuiFormControl-root">
									<RadioGroup value={s[k] || ''} disabled={!!member && s.auth == 'RW'} options={[...rw, { value: '', label: '권한없음' }]} onchange={(v) => set(s, v)} />
									{#if member && s.auth}<h6 class="MuiTypography-root MuiTypography-subtitle2">그룹 권한: {descMap[s.auth]}</h6>{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</Card>
		{/each}
		{#if diffs.length && admin.profile?.isDeveloper}<div class={cx('MuiGrid-root', toClass({ margin: 16, maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>{/if}
	</div>
	<hr class={cx('MuiDivider-root', toClass({ height: 100, border: 0, margin: 0, flexShrink: 0 }))} />
	<div
		class={cx(
			'MuiStack-root',
			toClass({ position: 'fixed', zIndex: 1, background: 'var(--zc-fff, #fff)', bottom: 0, left: '50%', width: '50%', '@media (max-width:899.95px)': { left: 0, width: '100%' } })
		)}
	>
		<Button variant="contained" fullWidth sx={{ height: 48, borderRadius: 0 }} disabled={!diffs.length} onclick={save}>저장</Button>
		{#if admin.profile?.isDeveloper}
			<IconButton sx={{ position: 'fixed', right: 10, bottom: 5, color: '#fff' }} aria-label="JSON 편집" onclick={() => (jsonOpen = true)}><Icon name="data_object" /></IconButton>
		{/if}
	</div>
	<JsonEditDialog bind:open={jsonOpen} value={items} onChange={(o) => (items = o)} title="권한 JSON" />
</Drawer>
