<script lang="ts">
	/**
	 * 메뉴에 묶인 API(권한 검사 대상) 편집(React setting/menu/components/MenuApiDrawer) — 줄마다 저장·삭제.
	 * GET·PUT /service/admin/menus/apis/{menuId} · DELETE …/{menuId}/{apiId}. menu 를 넘기면 열린다.
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import OutlinedField from '../ui/OutlinedField.svelte';

	let { menu, hostingId, onChange, onclose }: { menu: any; hostingId: string; onChange: (menuId: number, apis: any[]) => void; onclose: () => void } = $props();
	let apis = $state<any[] | null>(null);
	let orig: any[] = [];
	const q = () => (admin.profile?.isMaster ? `?hostingId=${hostingId}` : '');

	$effect(() => {
		const m = menu;
		untrack(() => {
			apis = null;
			if (m) load(m.menuId);
		});
	});
	async function load(menuId: number) {
		const o = await api.get(`/service/admin/menus/apis/${menuId}${q()}`, { headers: { cache: 'false' } });
		if (!o.ok) return alert('상세정보 로딩 도중 오류가 발생하였습니다.');
		orig = clone(o.data || []);
		apis = o.data || [];
	}
	const edited = $derived(!!apis && JSON.stringify(orig) != JSON.stringify(apis));
	function set(idx: number, k: 'api' | 'description', v: string) {
		const os = [...apis!];
		const o = { ...os[idx], [k]: v };
		o.edited = orig[idx] ? orig[idx].api != o.api || orig[idx].description != o.description : true;
		os[idx] = o;
		apis = os;
	}
	async function save(item: any, idx: number) {
		const o = await api.put(`/service/admin/menus/apis/${menu.menuId}${q()}`, item);
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 저장되었습니다.');
		const os = [...apis!];
		os[idx] = { ...o.data, edited: false };
		orig = clone(os);
		apis = os;
		onChange(menu.menuId, os);
	}
	async function remove(item: any, idx: number) {
		if (!(await confirm('삭제하시겠습니까?'))) return;
		if (item.apiId) {
			const o = await api.delete(`/service/admin/menus/apis/${menu.menuId}/${item.apiId}${q()}`);
			if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
			toast('정상적으로 삭제되었습니다.');
		}
		const os = [...apis!];
		os.splice(idx, 1);
		orig = clone(os);
		apis = os;
		onChange(menu.menuId, os);
	}
	async function close() {
		if (edited && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		onclose();
	}
</script>

<Drawer open={!!menu} anchor="right" onclose={close} paperSx={{ borderRadius: 0, userSelect: 'none', width: '50%' }}>
	<h4 class={cx('MuiTypography-root MuiTypography-h4', toClass({ textAlign: 'center', height: 48, marginTop: 16 }))}>{menu?.name || ''} API 설정</h4>
	<Card class="max" sx={{ padding: 24, paddingBottom: 16, marginLeft: 24, marginRight: 24 }}>
		{#each apis || [] as o, idx (o.apiId || -idx - 1)}
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8, '& > :not(style) ~ :not(style)': { marginLeft: 8 } }))}>
				<OutlinedField placeholder="ex) /service/admin/menus" size="small" sx={{ minWidth: 200 }} value={o.api || ''} maxLength={50} oninput={(v) => set(idx, 'api', v)} />
				<OutlinedField fullWidth placeholder="API 설명" size="small" value={o.description || ''} maxLength={100} oninput={(v) => set(idx, 'description', v)} />
				<Button variant="contained" disabled={!o.edited || !o.api} onclick={() => save(o, idx)}>저장</Button>
				<Button variant="outlined" color="error" onclick={() => remove(o, idx)}>삭제</Button>
			</div>
		{/each}
		{#if apis && apis.length == 0}<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center' }))}>등록된 API가 없습니다.</p>{/if}
	</Card>
	<Button
		variant="outlined"
		sx={{ width: 'fit-content', marginTop: 16, marginLeft: 24 }}
		onclick={() => (apis = [...(apis || []), { apiId: '', menuId: menu.menuId, api: '', description: '', edited: true }])}>API 추가</Button
	>
</Drawer>
