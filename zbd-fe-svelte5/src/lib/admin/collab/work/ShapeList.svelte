<script lang="ts">
	/**
	 * 문서의 메모·마크다운·도형 목록(React ShapeListTable) — 답글은 부모 아래 들여 씀.
	 * 칸: 도형(누르면 편집) · 제목(작성자·고친 사람·일정·「n분 전」·접기·상태) · 설명(마크다운/HTML) · 답글 · 좋아요(+개발자 JSON) · 삭제·공유 링크.
	 * 접힘은 이 브라우저 `shape_{보드}_{문서}_{도형}` (React 와 같은 자리). `?shapeId=` 면 그 줄을 노랗게 하고 가운데로.
	 */
	import { page } from '$app/state';
	import { date, shortDate } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { admin } from '../../profile.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { toast } from '../../ui/modal.svelte.ts';
	import Tooltip from '../../ui/Tooltip.svelte';
	import Av from '../ui/Av.svelte';
	import Badge from '../ui/Badge.svelte';
	import MarkdownView from '../ui/MarkdownView.svelte';
	import { HTML_TABLE_SX, MD_BOX_SX } from '../ui/md.ts';
	import { getWorkSettings, getWorkStatusCodes } from './codes.ts';
	import type { NotiMember, Shape } from '#lib/shared/types/work.ts';
	import JsonDrawer from './JsonDrawer.svelte';
	import ScheduleField from './ScheduleField.svelte';
	import StatusPopover from './StatusPopover.svelte';
	import { SK } from '#lib/shared/storage-keys.ts';

	let {
		shapes,
		groupKey,
		workId,
		admins = [],
		likeNames = [],
		openEditor,
		openReply,
		deleteShape,
		toggleLike,
		updateShape
	}: {
		shapes: Shape[];
		groupKey: string;
		workId: number;
		admins?: NotiMember[];
		likeNames?: { adminId: number; name: string }[];
		openEditor: (s: Shape) => void;
		openReply: (s: Shape) => void;
		deleteShape: (s: Shape) => void;
		toggleLike: (s: Shape) => void;
		updateShape: (s: Shape) => void;
	} = $props();

	const profile = $derived(admin.profile);
	const setting = $derived(getWorkSettings(groupKey));
	const target = $derived(page.url.searchParams.get('shapeId'));
	const GT = (n: number, f: string) => `var(--gt-${n}, ${f})`;
	const ckey = (id: string) => SK.work.shapeOpen(groupKey, workId, id);
	let collapsed = $state<Record<string, boolean>>({});
	let preview = $state<string | null>(null);
	let jsonShape = $state<Shape | null>(null);
	$effect(() => {
		const m: Record<string, boolean> = {};
		try {
			for (const s of shapes) if (localStorage.getItem(ckey(s.id)) === '1') m[s.id] = true;
		} catch {}
		collapsed = m;
	});
	function toggleCollapse(id: string) {
		const n = { ...collapsed };
		try {
			if (n[id]) (delete n[id], localStorage.removeItem(ckey(id)));
			else ((n[id] = true), localStorage.setItem(ckey(id), '1'));
		} catch {}
		collapsed = n;
	}
	const rows = $derived.by(() => {
		const map = new Map<string, Shape[]>();
		for (const s of shapes) {
			const p = s.parentShapeId ? String(s.parentShapeId) : 'root';
			map.set(p, [...(map.get(p) || []), s]);
		}
		const out: (Shape & { _depth: number })[] = [];
		const add = (pid: string, d: number) => (map.get(pid) || []).forEach((c) => (out.push({ ...c, _depth: d }), add(String(c.id), d + 1)));
		add('root', 0);
		for (const s of shapes) if (!out.find((r) => String(r.id) === String(s.id))) out.push({ ...s, _depth: 0 });
		return out;
	});
	let scrolledFor = '';
	$effect(() => {
		const t = target;
		if (!t || !shapes.length || scrolledFor === t) return;
		scrolledFor = t;
		setTimeout(() => {
			const el = document.getElementById(`shape_${t}`);
			if (!el) return;
			const y = el.getBoundingClientRect().top + window.scrollY - window.innerHeight / 2;
			window.scrollTo({ top: Math.max(0, y), behavior: 'smooth' });
		}, 300);
	});
	const icon = (t: string) =>
		({ pen: 'create', arrow: 'near_me', rect: 'crop_square', circle: 'radio_button_unchecked', text: 'text_fields', 'no-shape': 'text_fields', markdown: 'markdown' })[t] || 'help_outline';
	function ago(ts: number) {
		const diff = Date.now() - ts;
		if (diff < 0) return '';
		const m = Math.floor(diff / 60000);
		const h = Math.floor(m / 60);
		if (h >= 24) return '';
		if (h > 0) return m % 60 > 0 ? `${h}시간 ${m % 60}분 전` : `${h}시간 전`;
		return m > 0 ? `${m}분 전` : '방금 전';
	}
	async function share(s: Shape) {
		try {
			await navigator.clipboard.writeText(`${location.origin}/admin/work/${groupKey}/${workId}?shapeId=${s.id}`);
		} catch {}
		toast('클립보드로 공유링크가 복사되었습니다.');
	}
	const noReply = $derived(!!setting?.disableShapeReply);
	const noLike = $derived(!!setting?.disableShapeLike);
	const CAP = 'MuiTypography-root MuiTypography-caption';
</script>

{#if shapes.length}
	<div class="MuiStack-root" style="display:flex;flex-direction:column">
		<table
			class={cx(
				'MuiTable-root adm-cv-table sl-table',
				toClass({
					width: '100%',
					tableLayout: 'fixed',
					border: `1px solid ${GT(300, '#e0e0e0')}`,
					'& th': { paddingTop: 12, paddingBottom: 12, paddingLeft: 4, paddingRight: 4, background: 'transparent !important', fontWeight: 'normal !important' },
					'& td': { paddingTop: 0, paddingBottom: 0, paddingLeft: 8, paddingRight: 8, fontSize: 15, border: `1px solid ${GT(300, '#e0e0e0')}` }
				})
			)}
		>
			<thead class="MuiTableHead-root">
				<tr class="MuiTableRow-root MuiTableRow-head">
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium MuiTableCell-alignCenter" style="width:50px" scope="col">도형</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium" style="width:250px;min-width:250px" scope="col">제목</th>
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium" scope="col">설명</th>
					{#if !noReply}<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium MuiTableCell-alignCenter" style="width:50px" scope="col">답글</th>{/if}
					{#if !noLike}<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium MuiTableCell-alignCenter" style="width:50px" scope="col">좋아요</th>{/if}
					<th class="MuiTableCell-root MuiTableCell-head MuiTableCell-sizeMedium MuiTableCell-alignCenter" style="width:50px" scope="col">삭제</th>
				</tr>
			</thead>
			<tbody class="MuiTableBody-root">
				{#each rows as x (x.id)}
					{@const plain = x.type === 'no-shape' || x.type === 'markdown'}
					{@const isT = !!target && String(x.id) === target}
					{@const a = admins.find((m) => m.adminId === x.adminId)}
					{@const editor = x.lastEditAdminId && x.adminId != x.lastEditAdminId ? admins.find((m) => m.adminId === x.lastEditAdminId) : null}
					{@const val = x.schedule || null}
					{@const same = !!val?.from && val?.from == val?.to}
					{@const fold = !!collapsed[x.id]}
					<tr id={`shape_${x.id}`} class={cx('MuiTableRow-root', isT && 'adm-cv-target', !plain && 'adm-cv-shape-row')} style="cursor:{plain ? 'default' : 'pointer'}">
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium MuiTableCell-alignCenter" style="position:relative;vertical-align:top">
							{#if !plain}<Icon name="arrow_drop_up" sx={{ cursor: 'pointer', position: 'absolute', top: 0, left: 0, fontSize: '20px', color: 'var(--adm-action)' }} />{/if}
							<div class="MuiStack-root" style="display:flex;flex-direction:column;align-items:center">
								<IconButton disabled={!!x.activeEditorName && x.activeEditorName !== profile?.name} onclick={() => openEditor(x)}>
									{#if (x.parentShapeId || 0) > 0}<Icon name="subdirectory_arrow_right" size="Small" sx={{ color: GT(600, '#a0a0a0') }} />{:else}<Icon
											name={icon(x.type)}
											sx={{ color: x.color || 'inherit' }}
										/>{/if}
								</IconButton>
							</div>
						</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium" style="vertical-align:top">
							<div class="MuiStack-root" style="display:flex;flex-direction:column;margin-top:8px;margin-bottom:4px">
								{#if x.activeEditorName}<span class={CAP} style="border:1px solid {GT(700, '#909090')};background:{GT(300, '#e0e0e0')};padding:0 8px;margin-bottom:4px;width:fit-content"
										>{x.activeEditorName}</span
									>{/if}
								<div
									class="MuiStack-root"
									style="display:flex;flex-direction:{x._depth ? 'row' : 'column'};gap:{x._depth ? 4 : 0}px;padding-left:{x._depth * 12}px;{x._depth ? 'flex-wrap:wrap;align-items:center' : ''}"
								>
									<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;{x._depth ? 'flex-shrink:0;white-space:nowrap' : ''}">
										{#if x.isQuestion}<Icon name="contact_support" color="Error" />{/if}
										{#if a}<Tooltip title={a.name} placement="top"><Av src={a.profileImage || ''} text={a.name} size={15} /></Tooltip>{/if}
										{#if editor}<Tooltip title={editor.name} placement="top"><Av src={editor.profileImage || ''} text={editor.name} size={15} /></Tooltip>{/if}
										{#if !x._depth || x.useStatus}
											<ScheduleField schedule={val} onchange={(s) => updateShape({ ...x, schedule: s })}>
												{#snippet template()}
													<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;flex-shrink:0;white-space:nowrap">
														{#if !val}<IconButton size="Small"><Icon name="calendar_month" sx={{ fontSize: 16, color: GT(700, '#909090') }} /></IconButton>{/if}
														{#if val?.from}
															<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;flex-shrink:0;white-space:nowrap">
																<Icon name="calendar_month" sx={{ fontSize: 16, marginRight: 4, color: 'var(--theme-color-alpha50)' }} />
																{#if x._depth}
																	<span class={CAP} style="color:{GT(700, '#909090')};white-space:nowrap">{shortDate(val.from)}{val.to && !same ? ` ~ ${shortDate(val.to)}` : ''}</span>
																{:else}
																	<span class={CAP} style="color:{GT(700, '#909090')};white-space:nowrap">{date(val.from)}</span>
																	{#if val.to && !same}<span class={CAP} style="color:{GT(700, '#909090')};white-space:nowrap">{' ~ '} {shortDate(val.to)}</span>{/if}
																	{#if same}<span class={CAP} style="color:{GT(700, '#909090')};white-space:nowrap">{' ~ '} 당일</span>{/if}
																{/if}
															</div>
														{/if}
													</div>
												{/snippet}
											</ScheduleField>
										{/if}
										{#if !val?.from}<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;flex-wrap:wrap"><span>{x.title || ''}</span></div>{/if}
									</div>
									{#if val?.from}<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;margin-top:8px"><span>{x.title || ''}</span></div>{/if}
									{#if x.ts}
										{@const t = ago(x.ts)}
										{@const dt = String(date(x.ts)).split(' ')}
										<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;flex-grow:1;justify-content:space-between">
											<Tooltip title={String(date(x.ts))} placement="top">
												<div class="MuiBox-root" style="display:inline-flex;margin-top:{x._depth ? 0 : 4}px">
													{#if t}<p class="MuiTypography-root MuiTypography-body1" style="font-size:12px;color:{GT(700, '#909090')}">{t}</p>
													{:else}<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:4px">
															<p class="MuiTypography-root MuiTypography-body1" style="font-size:12px;color:{GT(700, '#909090')}">{dt[0]}</p>
															<span class={CAP} style="color:{GT(700, '#909090')}">{dt[1] || ''}</span>
														</div>{/if}
												</div>
											</Tooltip>
											<IconButton size="Small" sx={{ padding: 0 }} onclick={(e: MouseEvent) => (e.stopPropagation(), toggleCollapse(String(x.id)))}
												><Icon name="keyboard_arrow_down" size="Small" sx={{ transition: 'transform 0.2s', transform: fold ? 'rotate(-90deg)' : 'rotate(0deg)' }} /></IconButton
											>
										</div>
									{/if}
									{#if x.useStatus}
										{@const st = getWorkStatusCodes(groupKey)}
										<StatusPopover
											sx={{ marginTop: 8, width: 'fit-content' }}
											codes={st.codes}
											value={x.status || ''}
											onchange={(v) => updateShape({ ...x, status: v, color: st.codes.find((s) => s.id === v)?.bgColor || x.color })}
										/>
									{/if}
								</div>
							</div>
						</td>
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium" style="vertical-align:top;position:relative">
							{#if x.type === 'markdown'}
								<MarkdownView text={x.html || ''} sx={{ ...MD_BOX_SX, ...(fold ? { maxHeight: 110, overflow: 'hidden', position: 'relative' } : {}) }} />
							{:else}
								<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
								<div
									class={cx(
										'MuiBox-root',
										toClass({ ...(fold ? { maxHeight: 110, overflow: 'hidden', position: 'relative' } : {}), '& img': { cursor: 'pointer', maxWidth: '100%', height: 'auto' }, ...HTML_TABLE_SX })
									)}
									onclick={(e) => {
										const t = e.target as HTMLElement;
										if (t.tagName === 'IMG') (e.stopPropagation(), (preview = (t as HTMLImageElement).src));
									}}
								>
									{@html x.html || ''}
								</div>
							{/if}
							{#if fold}
								<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
								<div
									class="adm-cv-more"
									style="cursor:pointer;text-align:center;color:var(--zc-000, #000);font-size:12px;background:#f0f0f080;padding:2px 0;position:absolute;left:0;right:0;bottom:0"
									onclick={(e) => (e.stopPropagation(), toggleCollapse(String(x.id)))}
								>
									···
								</div>
							{/if}
						</td>
						{#if !noReply}
							<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium" style="vertical-align:top"
								><IconButton onclick={() => openReply(x)}><Icon name="reply" size="Small" /></IconButton></td
							>
						{/if}
						{#if !noLike}
							{@const likes = x.likeAdminIds || []}
							{@const names = likes
								.map((id) => likeNames.find((m) => m.adminId === id)?.name)
								.filter(Boolean)
								.join(', ')}
							<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium" style="vertical-align:top">
								<div class="MuiStack-root" style="display:flex;flex-direction:column;align-items:center;gap:2px">
									{#snippet likeBtn()}<IconButton onclick={() => toggleLike(x)}
											><Badge content={likes.length} color="primary"><Icon name="thumb_up" size="Small" color={likes.includes(Number(profile?.adminId)) ? 'Primary' : 'Inherit'} /></Badge></IconButton
										>{/snippet}
									{#if names}<Tooltip title={names} placement="top">{@render likeBtn()}</Tooltip>{:else}{@render likeBtn()}{/if}
									{#if profile?.isDeveloper}
										<Tooltip title="JSON 편집 (개발자 전용)" placement="right"
											><IconButton size="Small" onclick={() => (jsonShape = x)}><Icon name="data_object" size="Small" sx={{ color: GT(700, '#909090') }} /></IconButton></Tooltip
										>
									{/if}
								</div>
							</td>
						{/if}
						<td class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium" style="vertical-align:top">
							<IconButton onclick={() => deleteShape(x)}><Icon name="delete" size="Small" /></IconButton>
							<IconButton onclick={() => share(x)}><Icon name="share" size="Small" /></IconButton>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}

{#if preview}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div class="MuiModal-root" style="position:fixed;inset:0;z-index:1300;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.5)" onclick={() => (preview = null)}>
		<img src={preview} alt="preview" style="display:block;max-width:90vw;max-height:90vh;object-fit:contain" />
	</div>
{/if}
<JsonDrawer
	value={jsonShape}
	onclose={() => (jsonShape = null)}
	onsave={(v) => {
		if (!v?.id) return void toast('id 는 필수입니다 (누락됨).');
		updateShape(v);
		toast('shape 저장되었습니다.');
	}}
/>

<style>
	/* 휴대폰(2026-10-05 대표) — 표 대신 카드: 윗줄 = 도형·제목·답글·좋아요·삭제, 아랫줄 = 설명 전체 폭. 예전엔 설명 칸이 오른쪽에 글자 한두 개 폭으로 끼였다 */
	@media (max-width: 899px) {
		.sl-table,
		.sl-table :global(tbody) {
			display: block;
			border: 0 !important;
		}
		.sl-table :global(thead) {
			display: none;
		}
		.sl-table :global(tbody > tr) {
			display: grid;
			grid-template-columns: 40px minmax(0, 1fr) auto auto auto;
			align-items: start;
			margin-bottom: 8px;
			border: 1px solid var(--gt-300, #e0e0e0);
			border-radius: 8px;
			overflow: hidden;
		}
		.sl-table :global(tbody > tr > td) {
			display: block;
			border: 0 !important;
			padding: 0 2px !important;
			min-width: 0;
		}
		.sl-table :global(tbody > tr > td:nth-child(2)) {
			grid-row: 1;
			grid-column: 2 / -1;
		}
		.sl-table :global(tbody > tr > td:nth-child(3)) {
			grid-row: 2;
			grid-column: 1 / -1;
			padding: 0 10px 4px !important;
			overflow-x: auto;
		}
		/* 답글·좋아요·삭제는 아랫줄 오른쪽 끝에 */
		.sl-table :global(tbody > tr > td:nth-child(n + 4)) {
			grid-row: 3;
		}
		.sl-table :global(tbody > tr > td:nth-child(n + 4):last-child) {
			grid-column: 5;
		}
		.sl-table :global(tbody > tr > td:nth-child(n + 4):nth-last-child(2)) {
			grid-column: 4;
		}
		.sl-table :global(tbody > tr > td:nth-child(n + 4):nth-last-child(3)) {
			grid-column: 3;
		}
		.sl-table :global(tbody > tr > td:nth-child(n + 4) > .MuiStack-root) {
			flex-direction: row !important;
		}
	}
</style>
