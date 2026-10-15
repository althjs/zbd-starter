<script lang="ts">
	/**
	 * 사용자 회원 표(React pages/admin/user/member/components/MemberListTable) — 컬럼 보기 옵션(톱니) · 선택 약관 동의 · 관리(톱니 → 편집 서랍).
	 * 이메일 = 회원 상세(/admin/user/member/{id}) 링크.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { shortDate } from '../../format.ts';
	import { getHostingId } from '../../meta.svelte.ts';
	import { getTermPostGroup, socialUserGroupIdMap } from '../common.ts';
	import ColumnOptionDialog, { getTableColumnOptions, getTableColumnOptionsDefault, type ColumnOption } from '../ui/ColumnOptionDialog.svelte';

	let { userList, onManage, isPromotion = false }: { userList: any[]; onManage?: (u: any) => void; isPromotion?: boolean } = $props();

	const postGroup = getTermPostGroup();
	const useESPost = true;
	let optionalTerms = $state<any[] | null>(null);
	const keyOf = () => `member${isPromotion ? '_promotion' : ''}`;
	const optionKey = keyOf();
	let colOpts = $state<ColumnOption[]>(getTableColumnOptions(optionKey));
	const isCol = (k: string) => colOpts?.find((x) => x.key === k)?.use !== false;

	const configs = $derived([
		{ key: 'userId', name: '회원ID' },
		{ key: 'name', name: '이름' },
		{ key: 'nickname', name: '닉네임' },
		{ key: 'joinType', name: '가입방식' },
		{ key: 'email', name: '이메일' },
		{ key: 'mobile', name: '핸드폰번호' },
		...(optionalTerms && optionalTerms.length > 0 ? [{ key: 'optionalTerms', name: '선택약관' }] : []),
		...(!isPromotion ? [{ key: 'createdAt', name: '가입일' }] : []),
		...(!isPromotion ? [{ key: 'lastLoginAt', name: '마지막 접속일' }] : [])
	]);
	const defaults = $derived.by(() => {
		const saved = getTableColumnOptionsDefault(optionKey);
		return configs.map((c) => {
			const f = saved?.find((m) => m.key === c.key);
			return f ? { ...c, use: true, ...f } : { ...c, use: true };
		});
	});

	onMount(async () => {
		const o = await api.get(`/service/user/terms?hostingId=${getHostingId()}`);
		if (!o.ok) return;
		const keys = (postGroup?.items || []).filter((x: any) => x.agreementOptional).map((x: any) => x.groupSubKey);
		const os: any[] = [];
		for (const t of o.data || []) if (keys.indexOf(t.groupSubKey) !== -1 && !os.find((x) => x.groupSubKey == t.groupSubKey)) os.push(t);
		optionalTerms = os;
	});
	const showTerms = $derived(!!optionalTerms && optionalTerms.length > 0 && isCol('optionalTerms'));
	const th = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const td = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
	const agreed = (o: any, k: string) => {
		const x = o.Json?.agreeMap?.[k];
		return x == true || (typeof x == 'object' && x && Object.values(x).find((v) => v == true));
	};
</script>

<div class="MuiStack-root">
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'flex-end' }))}>
		<ColumnOptionDialog {optionKey} options={colOpts} defaultOptions={defaults} onChange={(v) => (colOpts = v)} useCommonCodes />
	</div>
	<table class={cx('MuiTable-root', toClass({ '--adm-td-px': '8px', '& td.MuiTableCell-root': { paddingLeft: 8, paddingRight: 8, paddingTop: 0, paddingBottom: 0 } }))}>
		<thead class="MuiTableHead-root">
			<tr class="MuiTableRow-root MuiTableRow-head">
				{#if isCol('userId')}<th class="{th} MuiTableCell-alignCenter" style="width:80px">회원ID</th>{/if}
				{#if isCol('name')}<th class={th}>이름</th>{/if}
				{#if isCol('nickname')}<th class={th}>닉네임</th>{/if}
				{#if isCol('joinType')}<th class="{th} MuiTableCell-alignCenter">가입방식</th>{/if}
				{#if isCol('email')}<th class={th}>이메일</th>{/if}
				{#if isCol('mobile')}<th class={th}>핸드폰번호</th>{/if}
				{#if showTerms}<th class={th}>선택약관</th>{/if}
				{#if !isPromotion && isCol('createdAt')}<th class={th}>가입일</th>{/if}
				{#if !isPromotion && isCol('lastLoginAt')}<th class={th}>마지막 접속일</th>{/if}
				{#if onManage}<th class="{th} MuiTableCell-alignCenter" style="width:60px">관리</th>{/if}
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each userList as o (o.guestKey || o.userId)}
				<tr class="MuiTableRow-root">
					{#if isCol('userId')}<td class="{td} MuiTableCell-alignCenter">{o.guestKey || o.userId}</td>{/if}
					{#if isCol('name')}<td class={td}>{o.name ?? ''}</td>{/if}
					{#if isCol('nickname')}<td class={td}>{o.nickname || ''}</td>{/if}
					{#if isCol('joinType')}<td class="{td} MuiTableCell-alignCenter">{socialUserGroupIdMap[o.groupId] || ''}{o.guestKey ? '비회원' : ''}</td>{/if}
					{#if isCol('email')}
						<td class={td}>
							{#if useESPost && o.userId > 0}<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover" href="/admin/user/member/{o.userId}">{o.email}</a
								>{:else}{o.email ?? ''}{/if}
						</td>
					{/if}
					{#if isCol('mobile')}<td class={td}>{o.mobile ?? ''}</td>{/if}
					{#if showTerms}
						<td class={td}>
							<div class="MuiStack-root">
								{#each optionalTerms || [] as term (term.groupSubKey)}
									{@const sub = postGroup?.items?.find((x: any) => x.groupSubKey == term.groupSubKey)}
									<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: agreed(o, term.groupSubKey) ? 'var(--theme-color-main)' : 'var(--gt-400, #d9d9d9)' }))}
										>{term.groupSubKey}
										{#if sub?.agreementItems?.length > 0}
											<span class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 4 }))}>
												{#each sub.agreementItems.filter((it: string) => o.Json?.agreeMap?.[term.groupSubKey]?.[it]) as it (it)}
													<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ fontSize: 10, fontWeight: 700 }))}>{it}</span>
												{/each}
											</span>
										{/if}
									</span>
								{/each}
							</div>
						</td>
					{/if}
					{#if !isPromotion && isCol('createdAt')}<td class={td}>{shortDate(o.createdAt)}</td>{/if}
					{#if !isPromotion && isCol('lastLoginAt')}<td class={td}>{shortDate(o.updatedAt) || '-'}</td>{/if}
					{#if onManage}
						<td class="{td} MuiTableCell-alignCenter">
							<IconButton onclick={() => onManage(o)} aria-label="관리"><Icon name="settings" /></IconButton>
						</td>
					{/if}
				</tr>
			{/each}
		</tbody>
	</table>
	{#if userList && userList.length == 0}
		<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', marginTop: 40, marginBottom: 40 }))}>등록된 회원이 없습니다.</p>
	{/if}
</div>
