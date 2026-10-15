<script lang="ts">
	/**
	 * 코드 표 한 단계(React codes/CodeCategoryListTable) — ID·이름(누르면 고치기)·전용 칸(설명·값·라벨 색·SEO·상태)·순서·전시·사용·하위분류·관리(⚙ 서랍)·삭제.
	 * onChange(코드) 가 Promise 를 돌려주면 저장된 값으로 열린 서랍도 맞춘다.
	 * 즈비드는 국내/해외(design_outbound) 가 없어 IN/OUT 칸·탭은 옮기지 않았다.
	 */
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import LabelPreview from '#lib/shared/ui/LabelPreview.svelte';
	import { media } from '../../media.svelte.ts';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Popover from '../../ui/Popover.svelte';
	import SortDialog from '../../ui/SortDialog.svelte';
	import StateChip from '../../ui/StateChip.svelte';
	import Button from '../../ui/Button.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import { adminImagePrefix, isSuperAdmin } from '../../meta.svelte.ts';
	import { activeNav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone, getNextId } from '../common.ts';
	import ColorField from '../ui/ColorField.svelte';
	import InlineText from '../ui/InlineText.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import CodeCategoryDrawer from './CodeCategoryDrawer.svelte';
	import StatusLabelEditor from './StatusLabelEditor.svelte';
	import { getParentCode, READ_ONLY_CODE_IDS } from './codes.ts';
	import '../settings.css';

	let {
		codeItems,
		onChange,
		doSortUpdate = null,
		title = '',
		caption = '',
		depth,
		maxDepth = null,
		componentName = '',
		onChangeBySubComponent = null,
		isSpread = false,
		spreadParentId = 0,
		inDrawer = false,
		onDelete = null,
		meta = null,
		parentIds = [],
		readonly = false
	}: {
		codeItems: any[] | null;
		onChange: (o: any) => Promise<any> | void;
		doSortUpdate?: ((os: any[], depth: number) => any) | null;
		title?: string;
		caption?: string;
		depth: number;
		maxDepth?: number | null;
		componentName?: string;
		onChangeBySubComponent?: ((o: any) => void) | null;
		isSpread?: boolean;
		spreadParentId?: any;
		inDrawer?: boolean;
		onDelete?: ((base: any, code: any) => any) | null;
		meta?: any;
		parentIds?: any[];
		readonly?: boolean;
	} = $props();

	let manage = $state<any>();
	let jsonItem = $state<any>(null);
	let jsonOpen = $state(false);
	let statusAnchor = $state<{ el: HTMLElement; o: any } | null>(null);

	const nextId = $derived(getNextId(codeItems));
	const isSuperAuth = $derived(isSuperAdmin(admin.profile));
	const isDev = $derived(!!admin.profile?.isDeveloper);
	const isCodeManage = /dom-codes|CODE_DOMESTIC|COMMON_CODES/.test(location.href);
	const componentNames = $derived(getParentCode(parentIds, meta)?.components || []);
	const showKeyColumn = $derived(isDev && inDrawer);
	const canWrite = $derived(activeNav()?.auth == 'RW');
	const lockFirst = $derived(!isSuperAuth && isCodeManage && depth == 1);
	const rows = $derived((codeItems || []).filter((o) => !(lockFirst && (READ_ONLY_CODE_IDS.includes(parseInt(o.id, 10)) || o.isSuperAuth))));
	/** 전용 칸 — 이 표 자체(componentName) + 부모가 정한 칸(components, 펼친 표 제외) */
	const extraCols = $derived([...(componentName ? [componentName] : []), ...(!isSpread ? componentNames.filter((c: string) => ['STATUS_LABEL', 'DESC', 'VALUE'].includes(c)) : [])]);

	function save(o: any) {
		const df = onChange(o);
		if (df) df.then((x: any) => manage?.updateModel(x));
	}
	const sub = (o: any) => onChangeBySubComponent?.(o);
	const fmt = (n: any) => (n ? Number(n).toLocaleString() : '');

	async function remove(code: any) {
		const ok = await confirm({
			title: `"${code.name || code.id}" 항목을 삭제하시겠습니까?`,
			body: `${meta?.metaKey == 'CODE_DOMESTIC' ? '상품에 코드값이 설정된 경우, 삭제시 문제가 발생할 수 있습니다.\n' : ''}삭제된 코드는 복구하실 수 없습니다.\n삭제를 진행하시겠습니까?`
		});
		if (!ok) return;
		const ids = [...parentIds, code.id];
		const item = clone(meta.json.items.find((x: any) => x.id == ids[0]));
		const deep = (p: any, now: number) => {
			if (now > item.maxDepth || !p?.items) return;
			if (depth == 1) {
				p._DELETED = true;
				return;
			}
			p.items = p.items.reduce((arr: any[], oo: any) => {
				if (!(depth == now && code.id == oo.id)) arr.push(oo);
				if (oo.id == parentIds[now - 1] && oo.items) deep(oo, now + 1);
				return arr;
			}, []);
		};
		deep(item, 2);
		onDelete?.(item, code);
	}

	const th = 'MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium';
	const td = 'MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium';
	const c = 'MuiTableCell-alignCenter';
	const r = 'MuiTableCell-alignRight';
	const px0 = toClass({ '&&&&': { paddingLeft: 0, paddingRight: 0 } });
</script>

{#snippet sw(checked: boolean, disabled: boolean, onchange: (v: boolean) => void)}
	<span class="MuiSwitch-root MuiSwitch-sizeSmall">
		<span class={cx('MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root', checked && 'Mui-checked', disabled && 'Mui-disabled')}>
			<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" {checked} {disabled} onchange={(e) => onchange((e.currentTarget as HTMLInputElement).checked)} />
			<span class="MuiSwitch-thumb"></span>
		</span>
		<span class="MuiSwitch-track"></span>
	</span>
{/snippet}

{#snippet headCols(name: string)}
	{#if name == 'LABEL_COLOR'}
		<th class="{th} {c}" style="width:150px">배경색</th><th class="{th} {c}" style="width:150px">글자색</th><th class="{th} {c}" style="width:150px">미리보기</th>
	{:else if name == 'DESC'}
		<th class="{th} {r}" style="width:150px">설명</th>
	{:else if name == 'VALUE'}
		<th class="{th} {r}" style="width:100px">값</th>
	{:else if name == 'SEO'}
		<th class="{th} {c}" style="width:110px">메뉴 이름</th><th class="{th} {c}" style="width:150px">title</th><th class="{th} {c}" style="width:200px">description</th><th
			class="{th} {c}"
			style="width:100px">og:image</th
		><th class="{th} {c}" style="width:50px">priority</th><th class="{th} {c}" style="width:100px">사이트맵 제외</th>
	{:else if name == 'STATUS_LABEL'}
		<th class="{th} {c}" style="width:110px">상태</th>
	{/if}
{/snippet}

{#snippet bodyCols(name: string, o: any)}
	{#if name == 'LABEL_COLOR'}
		<td class="{td} {c}"><ColorField value={o.bgColor || ''} sx={{ background: 'var(--zc-fff, #fff)', height: 40 }} onChange={(v) => sub({ ...o, bgColor: v })} /></td>
		<td class="{td} {c}"><ColorField value={o.color || ''} sx={{ background: 'var(--zc-fff, #fff)', height: 40 }} onChange={(v) => sub({ ...o, color: v })} /></td>
		<td class="{td} {c}"><LabelPreview label={o} /></td>
	{:else if name == 'DESC' || name == 'VALUE'}
		{@const k = name == 'VALUE' ? 'value' : 'desc'}
		<td class="{td} {r}">
			<InlineText
				placeholder={name == 'VALUE' ? '값' : '설명'}
				value={o[k] || ''}
				onChange={(v) => {
					v = String(v).substring(0, 30);
					const x = { ...o, [k]: v };
					if (!v.trim()) delete x[k];
					sub(x);
				}}
			/>
		</td>
	{:else if name == 'SEO'}
		<td class="{td} {c}"><OutlinedField size="small" placeholder="비우면 메뉴 아님" value={o.navName || ''} oninput={(v: string) => sub({ ...o, navName: v.substring(0, 20) })} /></td>
		<td class="{td} {c}"><OutlinedField size="small" value={o.title || ''} oninput={(v: string) => sub({ ...o, title: v })} /></td>
		<td class="{td} {c}"><OutlinedField size="small" value={o.desc || ''} oninput={(v: string) => sub({ ...o, desc: v })} /></td>
		<td class="{td} {c}">
			<ImagePicker
				image={o.image || null}
				size={{ width: 1200, height: 630 }}
				hostingId={meta?.hostingId || ''}
				useUnsplash
				folder="seo"
				onSelect={(img: any) => sub({ ...o, image: { name: img.name, imageId: 'seo', hostingId: meta?.hostingId || '' } })}
			>
				{#snippet after()}<img
						class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img', toClass({ maxWidth: 100 }))}
						src={`${adminImagePrefix(hostingIdOf(o.image) || '')}/${o.image?.imageId || ''}/${o.image?.name}`}
						alt=""
					/>{/snippet}
			</ImagePicker>
		</td>
		<td class="{td} {c}"><OutlinedField size="small" value={o.priority || ''} oninput={(v: string) => sub({ ...o, priority: v })} /></td>
		<td class="{td} {c}">{@render sw(!!o.ignoreSitemap, false, (tf) => sub({ ...o, ignoreSitemap: tf }))}</td>
	{:else if name == 'STATUS_LABEL'}
		<td class="{td} {c}">
			<Button
				variant="outlined"
				size="small"
				disabled={readonly || !canWrite}
				sx={{ whiteSpace: 'nowrap', fontSize: 12 }}
				onclick={(e: MouseEvent) => (statusAnchor = { el: e.currentTarget as HTMLElement, o })}
			>
				{#snippet startIcon()}<Icon name="label" sx={{ fontSize: '16px !important' }} />{/snippet}
				{(o.items || []).length}개
			</Button>
		</td>
	{/if}
{/snippet}

{#snippet body()}
	{#if isSpread}
		<h6 class={cx('MuiTypography-root MuiTypography-h6', toClass({ display: 'flex', alignItems: 'center', height: 50 }))}>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
				{#if caption}<span class="MuiTypography-root MuiTypography-caption">{caption}</span>{/if}{title}
			</div>
		</h6>
	{/if}
	<div class={cx('MuiBox-root', toClass({ overflowX: 'auto' }))}>
		<table
			class={cx(
				'MuiTable-root',
				toClass({
					'--adm-td-px': '4px',
					'& th.MuiTableCell-root.MuiTableCell-head': { padding: '16px 4px' },
					'& td.MuiTableCell-root.MuiTableCell-body': { padding: '0 4px' },
					'& tr:hover': { background: 'var(--gt-200, #f2f2f2)' },
					// 휴대폰 — 분류명 칸이 눌려 글자가 한 자씩 꺾이지 않게(넘치면 표만 옆으로 밀림)
					...(media.mdDown ? { '& .cc-name': { minWidth: 220 }, '& .MuiChip-root': { whiteSpace: 'nowrap', flexShrink: 0 } } : {})
				})
			)}
		>
			<thead class="MuiTableHead-root">
				<tr class="MuiTableRow-root MuiTableRow-head">
					<th class="{th} {c}" style="width:40px">ID</th>
					<th class="{th} cc-name">분류명</th>
					{#each extraCols as n, i (n + i)}{@render headCols(n)}{/each}
					{#if showKeyColumn}<th class="{th} {c}" style="width:130px">key</th>{/if}
					<th class="{th} {c}" style="width:60px">순서</th>
					{#if isDev}<th class="{th} {c}" style="width:55px">전시</th>{/if}
					<th class="{th} {c}" style="width:55px">사용</th>
					<th class="{th} {c}" style="width:55px">하위분류</th>
					<th class="{th} {c} {px0}" style="width:30px">관리</th>
					{#if onDelete && (meta?.metaKey == 'CATEGORY_DOMESTIC' || depth > 1)}<th class="{th} {c} {px0}" style="width:30px">삭제</th>{/if}
				</tr>
			</thead>
			<tbody class="MuiTableBody-root">
				{#each rows as o, idx (o.id)}
					<tr class={cx('MuiTableRow-root', spreadParentId && spreadParentId == o.id && toClass({ '&&': { background: 'var(--theme-color-alpha12)' } }))}>
						<td class="{td} {c}">{o.id}</td>
						<td class={td}>
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 4, alignItems: 'center' }))}>
								{#if isDev}
									<IconButton size="Small" aria-label="JSON" sx={{ display: 'inline-block' }} onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), (jsonItem = o), (jsonOpen = true))}
										><Icon name="data_object" size="Small" /></IconButton
									>
								{/if}
								{#if isCodeManage && depth == 1 && (READ_ONLY_CODE_IDS.includes(parseInt(o.id, 10)) || o.isSuperAuth)}
									<div class={cx('MuiStack-root', toClass({ marginBottom: 4 }))}><Icon name="admin_panel_settings" size="Small" /></div>
								{/if}
								{#if o.group}<StateChip size="small" sx={{ marginRight: 8 }} label={o.group} />{/if}
								<InlineText fullWidth value={o.name} sx={{ '& input': { paddingTop: 4, paddingBottom: 4 } }} onBlur={(v) => v && v != o.name && save({ ...o, name: v })} />
							</div>
						</td>
						{#each extraCols as n, i (n + i)}{@render bodyCols(n, o)}{/each}
						{#if showKeyColumn}
							<td class="{td} {c}">
								<OutlinedField
									size="small"
									placeholder="key"
									value={o.key || ''}
									sx={{ width: '100%' }}
									inputSx={{ '&&': { paddingTop: 4, paddingBottom: 4, fontSize: 12, textAlign: 'center' } }}
									onblur={(e: FocusEvent) => {
										const v = String((e.target as HTMLInputElement).value).trim();
										if (v === String(o.key || '')) return;
										const next: any = { ...o };
										if (v) next.key = v;
										else delete next.key;
										save(next);
									}}
								/>
							</td>
						{/if}
						<td class="{td} {c}">{idx + 1}</td>
						{#if isDev}
							<td class="{td} {c}">
								{@render sw(!!o.display, lockFirst, (tf) => {
									const x = { ...o, display: tf };
									if (!tf) delete x.display;
									save(x);
								})}
							</td>
						{/if}
						<td class="{td} {c}">{@render sw(!!o.use, lockFirst, (tf) => save({ ...o, use: tf }))}</td>
						<td class="{td} {c}">{fmt(o.items?.length) || '-'}</td>
						<td class="{td} {c} {px0}">
							<IconButton
								aria-label="관리"
								sx={o.components?.length > 0 ? { color: 'var(--adm-error-main)' } : null}
								onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), manage?.openDrawer(o, o.component || '', parentIds))}><Icon name="settings" /></IconButton
							>
						</td>
						{#if onDelete && (meta?.metaKey == 'CATEGORY_DOMESTIC' || depth > 1)}
							<td class="{td} {c} {px0}" style="width:30px">
								<span class="MuiBadge-root">
									<IconButton aria-label="삭제" sx={{ padding: 0, color: 'var(--theme-color-main)' }} onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), remove(o))}
										><Icon name="delete" /></IconButton
									>
									<span
										class="MuiBadge-badge MuiBadge-standard MuiBadge-invisible MuiBadge-anchorOriginTopRight MuiBadge-anchorOriginTopRightRectangular MuiBadge-overlapRectangular MuiBadge-colorPrimary"
									></span>
								</span>
							</td>
						{/if}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
	{#if codeItems && codeItems.length == 0}
		<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ marginTop: 24, marginBottom: 24, textAlign: 'center' }))}>등록된 분류 정보가 없습니다.</p>
	{/if}

	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', marginTop: 16, display: lockFirst ? 'none' : 'flex', '& > :not(style) ~ :not(style)': { marginLeft: 8 } }))}>
		{#if doSortUpdate}
			<SortDialog disabled={(codeItems || []).length < 2} items={codeItems || []} okText="저장" onChange={(os) => doSortUpdate?.(os, depth)} />
		{/if}
		<CodeCategoryDrawer
			bind:this={manage}
			{title}
			{nextId}
			depth={depth + 1}
			{maxDepth}
			readonly={readonly || !canWrite}
			onChange={(code: any) => {
				const df = onChange(code);
				if (df) df.then((x: any) => manage?.updateModel(x));
			}}
			createParentIds={parentIds}
			{meta}
		/>
	</div>
{/snippet}

<Card sx={{ flexGrow: 1, padding: 16 }}>{@render body()}</Card>

<JsonEditDialog bind:open={jsonOpen} value={jsonItem} onChange={(o) => save(o)} />

<Popover
	anchor={statusAnchor?.el || null}
	align="right"
	onclose={() => (statusAnchor = null)}
	paperSx={{ padding: 16, borderRadius: 16, boxShadow: '0px 4px 20px rgba(0,0,0,0.15)', minWidth: 360, maxWidth: 460 }}
>
	{#if statusAnchor}
		{@const o = statusAnchor.o}
		<StatusLabelEditor
			value={o.items || []}
			disabled={readonly || !canWrite}
			onChange={(list) => {
				const x = { ...o, items: list };
				statusAnchor = statusAnchor ? { ...statusAnchor, o: x } : null;
				sub(x);
			}}
		/>
		<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ display: 'block', marginTop: 8, color: 'var(--adm-text2)' }))}>변경 내용은 [저장] 을 눌러야 반영됩니다.</span>
	{/if}
</Popover>
