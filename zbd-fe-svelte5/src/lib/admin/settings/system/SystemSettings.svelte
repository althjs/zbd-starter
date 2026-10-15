<script lang="ts">
	/**
	 * 환경 설정(React pages/admin/setting/system/[tab]) — 홈페이지(design) · 해외 홈페이지 · 관리자 페이지(SITE) · 게시판 관리(POST_GROUPS) · 도메인(사용자·관리자 주소 — DomainManage).
	 * GET /service/admin/meta 로 메타를 다 받아 탭마다 해당 메타를 고친다. 저장하면 전역 메타(SITE)도 바꾼다(관리자 페이지 탭은 새로고침).
	 */
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import Tabs from '../../ui/Tabs.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { getHostingId, isSuperAdmin, setGlobalMeta } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import DomainManage from './DomainManage.svelte';
	import PostGroupManageForm from './PostGroupManageForm.svelte';
	import SiteDesignMetaManage from './SiteDesignMetaManage.svelte';
	import SiteMetaManage from './SiteMetaManage.svelte';

	/** onTab 을 주면(메타 관리 화면 안에 넣을 때 — React DesingSetting 의 setTab) 주소를 바꾸지 않고 그 함수로 탭을 바꾼다 */
	let { tab, onTab }: { tab: string; onTab?: (t: string) => void } = $props();
	const hostingTab = admin.profile?.hostingId || '';
	const hostingId = getHostingId();
	let metaItems = $state<any[] | null>(null);
	async function loadMeta() {
		const o = await api.get('/service/admin/meta');
		if (!o.ok) return alert('메타정보 로딩 도중 오류가 발생하였습니다.');
		metaItems = o.data;
	}
	onMount(loadMeta);
	const site = $derived((metaItems || []).find((o) => o.hostingId == hostingTab && o.metaKey == 'SITE'));
	const postGroups = $derived((metaItems || []).find((o) => o.hostingId == hostingTab && o.metaKey == 'POST_GROUPS'));
	const sup = $derived(isSuperAdmin(admin.profile));
	const tabs = $derived([
		...(hostingTab && hostingTab != 'SITE' ? [{ value: 'design', label: '홈페이지' }] : []),
		{ value: 'SITE', label: '관리자 페이지' },
		{ value: 'POST_GROUPS', label: '게시판 관리' },
		// 도메인 — 사용자·관리자 주소(바꾸면 로그인·주소 판단이 바뀐다 — 개발자·최고 관리자만)
		...(sup || admin.profile?.isDeveloper ? [{ value: 'DOMAIN', label: '도메인' }] : [])
	]);
	function updateMetaItem(m: any, reload: boolean) {
		if (m.metaKey == 'SITE' || m.json?.design) setGlobalMeta({ SITE: m.json });
		if (reload) return location.reload();
		toast('정상적으로 저장되었습니다.');
		metaItems = (metaItems || []).map((o) => (o.metaKey == m.metaKey && o.hostingId == m.hostingId ? m : o));
	}
</script>

<Tabs items={tabs} value={tab} onchange={(t) => (onTab ? onTab(t) : goto(`/admin/setting/system/${t}`, { replaceState: true }))} />
<!-- 탭마다 묶음(SettingGroup)이 저마다 맨 위 카드다 — 큰 카드로 다시 감싸지 않는다(2026-10-04 대표) -->
<div class="body">
	{#if metaItems}
		{#if tab == 'design' && site}
			<SiteDesignMetaManage meta={site} subKey={tab} onChange={(o) => updateMetaItem(o, false)} />
		{:else if tab == 'SITE' && site}
			<SiteMetaManage meta={site} hostingId={hostingTab} onChange={(o) => updateMetaItem(o, hostingId == o.hostingId)} />
		{:else if tab == 'DOMAIN'}
			<!-- 저장하면 SITE 가 바뀐다 — 다른 탭이 옛 SITE 를 통째로 다시 저장하지 않게 메타를 다시 읽는다 -->
			<DomainManage onSaved={loadMeta} />
		{:else if tab == 'POST_GROUPS' && postGroups}
			<PostGroupManageForm meta={postGroups} hostingId={hostingTab} onChange={(m) => updateMetaItem(m, false)} />
		{/if}
	{/if}
</div>

<style>
	/* 탭과 첫 카드 사이(2026-10-04 대표) */
	.body {
		margin-top: 16px;
	}
</style>
