<script lang="ts">
	/**
	 * 메타 관리(React pages/admin/setting/meta/[metaKey]) — 탭: 기본정보(home, 최고 관리자) · SITE · COMMON · HOSTINGS(옛 COMPANIES) · POST_GROUPS · PRODUCT_LABLES · COMMON_CODES
	 * (이 호스팅 메타에 있는 것만 탭으로 — React 와 같은 순서). 메타 메뉴(/admin/setting/meta) 가 권한 메뉴에 있으면 Card 로 감싼다.
	 * SITE·PRODUCT_LABLES 는 안쪽 탭을 주소 대신 화면 상태로 바꾼다(React 와 같게). 즈비드 한 호스팅이라 호스팅 고르기 탭은 없다.
	 */
	import { goto } from '$app/navigation';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import Card from '../../ui/Card.svelte';
	import Tabs from '../../ui/Tabs.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { isSuperAdmin } from '../../meta.svelte.ts';
	import PostGroupManageForm from '../system/PostGroupManageForm.svelte';
	import SystemSettings from '../system/SystemSettings.svelte';
	import HostingMetaJson from '../meta/HostingMetaJson.svelte';
	import MetaKeyTable from '../meta/MetaKeyTable.svelte';
	import CodeCategory from './CodeCategory.svelte';

	let { metaKey }: { metaKey: string } = $props();
	// 옛 이름 COMPANIES — 2026-10 옮기는 기간(설정 키가 HOSTINGS 로 바뀌기 전 행도 탭으로)
	const CUSTOM = ['SITE', 'COMMON', 'HOSTINGS', 'COMPANIES', 'POST_GROUPS', 'PRODUCT_LABLES', 'COMMON_CODES'];
	const SUB_DEFAULT: Record<string, string> = { SITE: 'SITE', PRODUCT_LABLES: 'all' };
	let metaItems = $state<any[] | null>(null);
	let subTab = $state('');
	$effect(() => {
		subTab = SUB_DEFAULT[metaKey] || '';
	});
	const hostingTab = $derived(admin.profile?.hostingId || '');

	$effect(() => {
		api.get('/service/admin/meta').then((o) => {
			if (!o.ok) return alert('메타정보 로딩 도중 오류가 발생하였습니다.');
			metaItems = o.data;
		});
	});
	function update(m: any) {
		metaItems = (metaItems || []).map((o) => (o.metaKey == m.metaKey && o.hostingId == m.hostingId ? m : o));
	}
	const find = (k: string) => (metaItems || []).find((o) => o.hostingId == hostingTab && o.metaKey == k);
	const useTab = $derived((admin.menus || []).some((o: any) => (o.path || '') == '/admin/setting/meta' || (o.items || []).some((mm: any) => (mm.path || '') == '/admin/setting/meta')));
	const tabs = $derived([
		...(admin.profile?.isMaster || isSuperAdmin(admin.profile) ? [{ value: 'home', label: '기본정보' }] : []),
		...CUSTOM.map((k) => find(k))
			.filter(Boolean)
			.map((o: any) => ({ value: o.metaKey, label: o.name || o.metaKey }))
	]);
</script>

{#snippet form()}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', flexGrow: 1, justifyContent: 'space-between' }))}>
		<p class="MuiTypography-root MuiTypography-body2"></p>
		{#if useTab}<Tabs value={metaKey} items={tabs} onchange={(v: string) => goto(`/admin/setting/meta/${v}`, { replaceState: true })} />{/if}
	</div>
	{#if metaKey == 'home'}
		<MetaKeyTable meta={metaItems} hostingId={hostingTab} />
	{:else if metaKey == 'SITE'}
		<SystemSettings tab={subTab || 'SITE'} onTab={(t) => (subTab = t)} />
	{:else if metaKey == 'COMMON'}
		<HostingMetaJson meta={find('COMMON')} metaKey="COMMON" onChange={update} />
	{:else if metaKey == 'HOSTINGS' || metaKey == 'COMPANIES'}
		<HostingMetaJson meta={find(metaKey)} {metaKey} onChange={update} />
	{:else if metaKey == 'COMMON_CODES'}
		{@const m = metaItems?.find((o) => o.metaKey == 'COMMON_CODES')}
		{#if m}<CodeCategory hostingId={hostingTab} title="공콩 코드" meta={m} onChange={update} />{/if}
	{:else if metaKey == 'POST_GROUPS' && find('POST_GROUPS')}
		<PostGroupManageForm meta={find('POST_GROUPS')} hostingId={hostingTab} onChange={update} />
	{/if}
{/snippet}

<PageContainer hintPostId="lug2tbcq">
	{#if metaItems}
		<div class={cx('MuiStack-root', toClass({ minWidth: 0, overflowX: 'auto', '@media (min-width:1200px)': { minWidth: 1200, overflowX: 'visible' } }))}>
			{#if useTab}<Card sx={{ padding: 16 }}>{@render form()}</Card>{:else}<div class="MuiStack-root">{@render form()}</div>{/if}
		</div>
	{/if}
</PageContainer>
