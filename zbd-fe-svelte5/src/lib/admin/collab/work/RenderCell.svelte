<script lang="ts">
	/**
	 * 업무 보드 표의 칸 하나(React WorkListTableRenderCell) — id·분류·제목·요약·설명·상태·만든/고친 때 + 보드가 정한 열
	 * (inputType EMAILS·TAGS → 이름표, ADMINS → 담당자, SCHEDULE → 일정, items → 색 버튼, workFields → 칸 값 세로로).
	 */
	import { date, shortDate } from '#lib/shared/format.ts';
	import { admin } from '../../profile.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import Av from '../ui/Av.svelte';
	import AdminPick from './AdminPick.svelte';
	import CategoryBreadcrumbs from './CategoryBreadcrumbs.svelte';
	import { getWorkColumnCode, getWorkSettings, getWorkStatusCodes, getCategoryLabel, isWorkWriteable, toDescHtml } from './codes.ts';
	import type { WorkItem, CodeItem, NotiMember, Shape } from '#lib/shared/types/work.ts';
	import ExpandableText from './ExpandableText.svelte';
	import ScheduleField from './ScheduleField.svelte';
	import StatusPopover from './StatusPopover.svelte';
	import TagsCell from './TagsCell.svelte';

	export type Col = {
		key: string;
		name: string;
		width: number;
		align?: string;
		sticky?: boolean;
		id?: number;
		items?: CodeItem[];
		inputType?: string;
		left?: number;
		borderRight?: string;
		workFields?: string[];
		component?: string;
	};
	let {
		col,
		item: o,
		shapes = [],
		summaryText = '',
		groupKey,
		codes,
		codeIds = [],
		admins = [],
		onValueChange,
		isHiddenCategoryColumn = false,
		onjson
	}: {
		col: Col;
		item: WorkItem;
		shapes?: Shape[];
		summaryText?: string;
		groupKey: string;
		codes: CodeItem[];
		codeIds?: number[];
		admins?: NotiMember[];
		onValueChange: (item: WorkItem, key: string, v: any) => Promise<void> | void;
		isHiddenCategoryColumn?: boolean;
		onjson?: (item: WorkItem) => void;
	} = $props();

	const profile = $derived(admin.profile);
	const writeable = $derived(isWorkWriteable(groupKey, profile));
	const noDetail = $derived(!!getWorkSettings(groupKey)?.disableDetail);
	const href = $derived(`/admin/work/${groupKey}/${o.workId}`);
	const val = $derived((o.Json as any)?.[col.key] || '');
	const files = $derived([...(o.Json.images || []).flatMap((i) => i.files || []), ...(o.Json.files || [])]);
	const dt = (v: any) => String(date(v || '') || '').split(' ');
	const LINK = 'MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover';
</script>

{#if col.key === 'id'}
	{#if noDetail}<span class="MuiTypography-root MuiTypography-caption">{o.workId}</span>{:else}<a class={LINK} {href} title={o.Json.title}>{o.workId}</a>{/if}
{:else if col.key === 'category'}
	<div class="MuiStack-root" style="display:flex;flex-direction:column">
		<CategoryBreadcrumbs category={o.Json.category} {codes} {groupKey} workId={o.workId} />
		{#each shapes.filter((s) => s.isQuestion && s.title) as s (s.id)}
			<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;margin-top:4px">
				<Icon name="contact_support" sx={{ fontSize: '14px !important', color: 'var(--adm-error-main)' }} />
				<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-error-main);font-weight:bold">{s.title}</span>
			</div>
		{/each}
	</div>
{:else if col.key === 'title'}
	<div class="MuiStack-root" style="display:flex;flex-direction:column">
		{#if noDetail}<p class="MuiTypography-root MuiTypography-body2" title={o.Json.title}>{o.Json.title}</p>{:else}<a class={LINK} {href} title={o.Json.title}>{o.Json.title}</a>{/if}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center">
			{#if isHiddenCategoryColumn}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">{getCategoryLabel(o.Json.category, codes)}</span>{/if}
			{#if profile?.isDeveloper}
				<div class="MuiStack-root adm-cv-dev" style="display:flex;flex-direction:row;margin-left:8px">
					{#if onjson}<IconButton size="Small" sx={{ width: 'fit-content' }} onclick={() => onjson(o)}><Icon name="data_object" sx={{ fontSize: 13 }} /></IconButton>{/if}
					{#if noDetail}<IconButton size="Small" sx={{ width: 'fit-content' }} {href}><Icon name="open_in_new" sx={{ fontSize: 13 }} /></IconButton>{/if}
				</div>
			{/if}
			{#each files as f, i (i)}
				<Tooltip title={f.fileName || ''} placement="top"><Icon name="file_present" sx={{ fontSize: 16, color: 'var(--gt-600, #a0a0a0)', cursor: 'default' }} /></Tooltip>
			{/each}
		</div>
	</div>
{:else if col.key === 'summary'}
	<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px">
		{#if shapes.length}
			<p class="MuiTypography-root MuiTypography-body2">{shapes.length}개</p>
			<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2)">({summaryText})</span>
		{/if}
	</div>
{:else if col.key === 'desc'}
	<ExpandableText text={o.Json.desc || ''} cashPurpose={o.Json.cashReceiptPurpose} cashNumber={o.Json.cashReceiptNumber} />
{:else if col.key === 'status'}
	{@const st = getWorkStatusCodes(groupKey)}
	<StatusPopover
		codes={st.codes}
		disabled={!writeable}
		value={o.Json.status || ''}
		codeIds={codeIds?.length && st.codeId > 0 ? [...codeIds, st.codeId] : null}
		onchange={(v) => onValueChange(o, 'status', v)}
	/>
{:else if col.key === 'createdAt'}
	{@const d = dt(o.createdAt)}
	{#if d[0]}
		{@const a = admins.find((m) => m.adminId === o.adminId)}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:8px">
			{#if a?.adminId}<Tooltip title={a.name} placement="top"><Av src={a.profileImage || ''} size={30} sx={{ backgroundColor: 'var(--zc-eee, #eee)', flexShrink: 0 }} /></Tooltip>{/if}
			<div class="MuiStack-root" style="display:flex;flex-direction:column;min-width:0">
				<p class="MuiTypography-root MuiTypography-body1" style="font-size:12px;white-space:nowrap">{d[0]}</p>
				<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090);white-space:nowrap">{a?.name ? `${a.name} · ` : ''}{d[1] || ''}</span>
			</div>
		</div>
	{/if}
{:else if col.key === 'updatedAt'}
	{@const d = dt(o.updatedAt)}
	{#if d[0]}
		<div class="MuiStack-root" style="display:flex;flex-direction:column">
			<p class="MuiTypography-root MuiTypography-body1" style="font-size:12px">{d[0]}</p>
			<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">{d[1] || ''}</span>
		</div>
	{/if}
{:else if col.workFields?.length}
	<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:4px">
		{#each col.workFields as f (f)}
			{@const v = (o.Json as any)?.[f]}
			<span class="MuiTypography-root MuiTypography-caption" style="white-space:nowrap">{v === undefined || v === null || v === '' ? '-' : typeof v === 'object' ? JSON.stringify(v) : String(v)}</span>
		{/each}
	</div>
{:else if col.inputType === 'EMAILS'}
	<TagsCell
		tags={val}
		onchange={(v) => onValueChange(o, col.key, v)}
		emptyText="이메일로 사용자 초대"
		modalTitle="이메일 초대"
		placeholder={'콤마 "," 로 이메일이 구분됩니다.'}
		readonly={!writeable}
	/>
{:else if col.inputType === 'TAGS'}
	<TagsCell tags={val} onchange={(v) => onValueChange(o, col.key, v)} readonly={!writeable} />
{:else if col.inputType === 'ADMINS'}
	<AdminPick targetAdminIds={getWorkColumnCode(groupKey, col.id || 0)?.adminIds || []} adminIds={val || []} onchange={(v) => onValueChange(o, col.key, v)} {admins} readonly={!writeable} />
{:else if col.inputType === 'SCHEDULE'}
	{@const same = !!val?.from && val?.from == val?.to}
	<div class="MuiStack-root" style="display:flex;flex-direction:column">
		<ScheduleField schedule={val || null} readonly={!writeable} onchange={(s) => onValueChange(o, col.key, s)}>
			{#snippet template()}
				<div class="MuiStack-root" style="display:flex;flex-direction:column">
					{#if !val}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">일정 선택</span>{/if}
					{#if val?.from}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--gt-700, #909090)">{date(val.from)}</span>{/if}
					{#if val?.to && !same}<span class="MuiTypography-root MuiTypography-caption">~ {shortDate(val.to)}</span>{/if}
					{#if same}<span class="MuiTypography-root MuiTypography-caption">~ 당일</span>{/if}
				</div>
			{/snippet}
		</ScheduleField>
	</div>
{:else if col.items}
	<StatusPopover codes={col.items} value={val} codeIds={codeIds?.length && col.id != null ? [...codeIds, col.id] : null} onchange={(v) => onValueChange(o, col.key, v)} />
{:else}
	{val}
{/if}
