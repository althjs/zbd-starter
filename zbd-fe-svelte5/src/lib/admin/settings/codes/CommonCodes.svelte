<script lang="ts">
	/**
	 * 공통 코드 관리(React pages/admin/common-codes, common-codes/[groupKey]) — COMMON_CODES 메타를 읽어 분류 표.
	 * groupKey 면 그 그룹 하위만, 제목·부제는 그 그룹 이름·설명.
	 */
	import { api } from '#lib/shared/api.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import CodeCategory from './CodeCategory.svelte';

	let { groupKey = '' }: { groupKey?: string } = $props();
	let meta = $state<any>(null);
	const metaItem = $derived(meta?.json?.items?.find((x: any) => x.group == groupKey));

	$effect(() => {
		groupKey;
		api.get('/service/admin/meta?metaKey=COMMON_CODES').then((o) => {
			if (!o.ok) return alert('등록된 메타 정보가 없습니다.');
			meta = o.data;
		});
	});
</script>

<PageContainer title={groupKey ? metaItem?.name || '' : ''} subtitle={groupKey ? metaItem?.desc || '' : ''}>
	{#if meta}
		{#key groupKey}<CodeCategory title="공콩 코드" {meta} {groupKey} onChange={(m) => (meta = m)} />{/key}
	{/if}
</PageContainer>
