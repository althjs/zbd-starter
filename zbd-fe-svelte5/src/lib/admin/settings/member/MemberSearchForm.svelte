<script lang="ts">
	/**
	 * 회원 검색 줄(React pages/admin/setting/member/components/MemberSearchForm) — 관리자·사용자 회원 공용.
	 * 관리자(isAdmin): 위에 그룹 탭 + 검색어. 사용자: 검색어 + 선택 약관 + 기본 그룹 + 회원 상태.
	 * 검색어는 칠 때마다 getData(목록이 화면에서 거른다).
	 */
	import { onMount } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Card from '../../ui/Card.svelte';
	import Tabs from '../../ui/Tabs.svelte';
	import { admin } from '../../profile.svelte.ts';
	import { getMeta, isSuperAdmin } from '../../meta.svelte.ts';
	import { adminGroups, ensureAdminGroups } from '../adminGroups.svelte.ts';
	import { getMemberBasicAuths, getNumber, getTermPostGroup, USER_STATUS_MAP } from '../common.ts';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import OutlinedSelect from '../ui/OutlinedSelect.svelte';
	import OptionalTermsFilterForm from '../user/OptionalTermsFilterForm.svelte';

	let { search, getData, sx = null, isAdmin = true, memberType = 'ADMIN' }: { search: any; getData: (s: any) => void; sx?: Sx | null; isAdmin?: boolean; memberType?: string } = $props();

	onMount(() => {
		if (isAdmin && memberType == 'ADMIN') ensureAdminGroups();
	});
	const isSuperAuth = $derived(isSuperAdmin(admin.profile));
	const masterGroupIds: number[] = getMeta('COMMON', 'masterGroupIds') || [];
	const basicAuths = getMemberBasicAuths();
	const hasOptionalTerms = !!(getTermPostGroup()?.items || []).find((x: any) => x.agreementOptional || x.agreementItems?.length > 0);

	const groupTabs = $derived([
		{ value: '', label: '전체' },
		...(adminGroups.list || [])
			.filter((o) => !(masterGroupIds.indexOf(o.groupId) !== -1 && !isSuperAuth))
			.filter((o) => o.hostingId == search.hostingId || (!search.hostingId && !!admin.profile?.hostingId))
			.map((o) => ({ value: o.groupId, label: o.name }))
	]);
</script>

<div class={cx('MuiStack-root', toClass(sx))}>
	{#if isAdmin && memberType == 'ADMIN'}
		<Tabs items={groupTabs} value={search.groupId ?? ''} onchange={(tab) => getData({ ...search, pageNo: 1, groupId: tab })} />
	{/if}

	<Card sx={{ padding: 16 }}>
		<div class={cx('MuiStack-root', !isAdmin && 'msf-row', toClass({ flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 16 }))}>
			<OutlinedField
				value={search.keyword || ''}
				placeholder="이름/이메일/휴대폰"
				sx={{ minWidth: '100%', '@media (min-width:600px)': { minWidth: 300 } }}
				oninput={(v) => getData({ ...search, keyword: v, pageNo: 1 })}
			/>

			{#if !isAdmin && hasOptionalTerms}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
					<OptionalTermsFilterForm agreeMap={search.agreeMap || {}} onChange={(am) => getData({ ...search, pageNo: 1, agreeMap: am })} />
				</div>
			{/if}

			{#if !isAdmin}
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 8 }))}>
					{#if basicAuths.length > 0}
						<div class="MuiStack-root">
							<OutlinedSelect
								label="기본 그룹"
								value={search.authGroupId === undefined ? '' : search.authGroupId + ''}
								options={[{ value: '', label: '전체' }, ...basicAuths.map((o: any) => ({ value: o.id + '', label: o.name }))]}
								sx={{ minWidth: 150 }}
								onchange={(v) => getData({ ...search, pageNo: 1, authGroupId: getNumber(v, null, true) })}
							/>
						</div>
					{/if}
					<OutlinedSelect
						label="회원 상태"
						size="medium"
						value={search.userStatus === undefined ? '0' : search.userStatus + ''}
						options={Object.keys(USER_STATUS_MAP).map((k) => ({ value: k, label: USER_STATUS_MAP[k] }))}
						sx={{ minWidth: 150 }}
						onchange={(v) => getData({ ...search, pageNo: 1, userStatus: getNumber(v, null, true) })}
					/>
				</div>
			{/if}
		</div>
	</Card>
</div>

<style>
	/* 사용자 회원 검색 줄 — 검색칸·선택 상자(라벨 있음) 높이를 48 로 같게(두 부품의 기본 높이가 40·51 로 달랐고, 선택 상자는 테두리가 아래로 밀려 있었다) */
	.msf-row :global(.MuiInputBase-root) {
		height: 48px;
		box-sizing: border-box;
	}
	.msf-row :global(.MuiOutlinedInput-input) {
		height: 100%;
		box-sizing: border-box;
		padding-top: 0;
		padding-bottom: 0;
	}
	.msf-row :global(.adm-sel-labeled .MuiSelect-outlined) {
		margin-top: 0;
		padding-top: 20px;
		line-height: 23px;
	}
	/* 라벨(legend)이 테두리 선을 아래로 밀어 테두리가 칸보다 12px 내려가 있었다 — 라벨을 띄우고 테두리를 칸에 붙인다 */
	.msf-row :global(.adm-sel-labeled fieldset) {
		top: 0;
	}
	.msf-row :global(.adm-sel-labeled legend) {
		position: absolute;
		top: 0;
		left: 8px;
	}
	.msf-row :global(.adm-sel-labeled legend > span) {
		margin-top: 6px;
		line-height: 12px;
	}
</style>
