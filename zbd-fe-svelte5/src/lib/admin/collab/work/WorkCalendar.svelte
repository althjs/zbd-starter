<script lang="ts">
	/**
	 * 업무 보드 캘린더 보기(React canvas/components/WorkListCalendar) — 일정(schedule.from~to)이 있는 카드와 그 안 메모(shape)를 달력에 막대로.
	 * 주마다 칸(slot)을 나눠 긴 일정이 이어지게, 막대 색은 그 카드 상태(status 코드의 bgColor/color), 메모 막대는 반투명.
	 * 막대에는 시작일(그 주 첫 칸)에만 담당자 얼굴 + 「대분류) 제목」. 누르면 `/admin/work/{보드}/{번호}?ymd=…[&shapeId=]`.
	 * 보드 설정: isCalendarStartWithMonday(월요일 시작) · disableCalendarHolyday(휴일 끔 — 메타 COMMON.holydays) · useCalendarWriteButton(칸마다 「글쓰기」) ·
	 * useCalendarPopoverTooltip(막대에 올리면 제목·본문 미리보기, 쓰기 권한 없으면 링크 없음).
	 * 처음 달은 주소 ?ymd= (읽은 뒤 지운다). 여러 보드가 섞인 목록(전체 일정)도 각 카드의 groupKey 로 상태·설정을 찾는다.
	 */
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount, untrack } from 'svelte';
	import Button from '#lib/admin/ui/Button.svelte';
	import Icon from '#lib/admin/ui/Icon.svelte';
	import IconButton from '#lib/admin/ui/IconButton.svelte';
	import Tooltip from '#lib/admin/ui/Tooltip.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { getMeta } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import AddWorkButton from './AddWorkButton.svelte';
	import { getWorkCodes, getWorkSettings, getWorkStatusCodes, isWorkWriteable } from './codes.ts';
	import type { WorkItem, NotiMember } from '#lib/shared/types/work.ts';

	let { groupKey, items, admins = [] }: { groupKey: string; items: WorkItem[]; admins?: NotiMember[] } = $props();

	// ── 날짜 도우미(date-fns 대신) ──
	const pad = (n: number) => String(n).padStart(2, '0');
	const ymd = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
	const addDays = (d: Date, n: number) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + n);
	const startOfMonth = (d: Date) => new Date(d.getFullYear(), d.getMonth(), 1);
	const endOfMonth = (d: Date) => new Date(d.getFullYear(), d.getMonth() + 1, 0);
	const startOfWeek = (d: Date, ws: number) => addDays(d, -((d.getDay() - ws + 7) % 7));
	const endOfWeek = (d: Date, ws: number) => addDays(startOfWeek(d, ws), 6);

	// 처음 달 = 주소 ?ymd= — 읽고 주소에서 지운다(한 번만, React 와 같게)
	const q = untrack(() => page.url.searchParams.get('ymd'));
	let current = $state(q ? new Date(q) : new Date());
	onMount(() => {
		if (!q) return;
		const u = new URL(location.href);
		u.searchParams.delete('ymd');
		replaceState(u.pathname + u.search, {});
	});

	const setting = $derived(getWorkSettings(groupKey) as any);
	const statusCodes = $derived(getWorkStatusCodes(groupKey).codes);
	const writeable = $derived(isWorkWriteable(groupKey, admin.profile));
	const ctxMap = $derived.by(() => {
		const m: Record<string, { statusCodes: any[]; setting: any; writeable: boolean }> = {};
		for (const it of items) {
			const gk = (it as any).groupKey || groupKey;
			if (gk && !m[gk]) m[gk] = { statusCodes: getWorkStatusCodes(gk).codes, setting: getWorkSettings(gk), writeable: isWorkWriteable(gk, admin.profile) };
		}
		return m;
	});
	const ctxOf = (it: any) => ctxMap[it.groupKey || groupKey] || { statusCodes, setting, writeable };

	const holydays = $derived(setting?.disableCalendarHolyday ? [] : ((getMeta('COMMON', 'holydays') || []) as any[]));
	const holyMap = $derived(holydays.reduce((m: Record<string, string>, o: any) => ((m[o.ymd] = o.name), m), {}));
	const monthHoly = $derived(holydays.filter((h: any) => String(h.ymd).startsWith(`${current.getFullYear()}-${pad(current.getMonth() + 1)}`)).sort((a: any, b: any) => a.ymd.localeCompare(b.ymd)));
	const ws = $derived(setting?.isCalendarStartWithMonday ? 1 : 0);
	const weekDays = $derived(ws === 1 ? ['월', '화', '수', '목', '금', '토', '일'] : ['일', '월', '화', '수', '목', '금', '토']);
	const isSun = (i: number) => (ws === 0 ? i === 0 : i === 6);
	const isSat = (i: number) => (ws === 0 ? i === 6 : i === 5);

	/** 일정이 있는 카드 + 그 안의 메모(카드 메모·이미지 메모)를 막대 하나씩으로 */
	function shapeEvent(item: any, shape: any) {
		return {
			...item,
			_eventId: `shape_${shape.id}`,
			_isShape: true,
			_shapeId: shape.id,
			Json: {
				...item.Json,
				schedule: shape.schedule,
				status: shape.status !== undefined && shape.status !== null ? shape.status : item.Json.status,
				title: `${item.Json.title || '제목없음'} > ${shape.title || '텍스트 없음'}`,
				adminIds: shape.adminId ? Array.from(new Set([shape.adminId, shape.lastEditAdminId].filter((v) => typeof v === 'number' && v > 0))) : item.Json.adminIds
			}
		};
	}
	const events = $derived.by(() => {
		const out: any[] = [];
		for (const item of items as any[]) {
			if (item.Json?.schedule?.from) out.push(item);
			for (const s of item.Json?.shapes || []) if (s.schedule?.from) out.push(shapeEvent(item, s));
			for (const img of item.Json?.images || []) for (const s of img.shapes || []) if (s.schedule?.from) out.push(shapeEvent(item, s));
		}
		return out;
	});

	const grid = $derived.by(() => {
		const monthStart = startOfMonth(current);
		const start = startOfWeek(monthStart, ws);
		const end = endOfWeek(endOfMonth(monthStart), ws);
		const rows: Date[][] = [];
		for (let d = start; d <= end; d = addDays(d, 7)) rows.push(Array.from({ length: 7 }, (_x, i) => addDays(d, i)));
		const rowSlots = rows.map((rowDays) => {
			const wS = ymd(rowDays[0]);
			const wE = ymd(rowDays[6]);
			const week = events
				.filter((it) => it.Json.schedule.from <= wE && (it.Json.schedule.to || it.Json.schedule.from) >= wS)
				.sort((a, b) => {
					const af = a.Json.schedule.from,
						bf = b.Json.schedule.from;
					if (af !== bf) return af.localeCompare(bf);
					const al = new Date(a.Json.schedule.to || af).getTime() - new Date(af).getTime();
					const bl = new Date(b.Json.schedule.to || bf).getTime() - new Date(bf).getTime();
					return bl - al;
				});
			const slots: (any | null)[][] = Array.from({ length: 7 }, () => []);
			for (const it of week) {
				const from = it.Json.schedule.from,
					to = it.Json.schedule.to || from;
				let si = -1,
					ei = 7;
				for (let i = 0; i < 7; i++) {
					const d = ymd(rowDays[i]);
					if (si === -1 && d >= from) si = i;
					if (d > to) {
						ei = i;
						break;
					}
				}
				if (si === -1) si = 0;
				let slot = 0;
				for (;;) {
					let ok = true;
					for (let i = si; i < ei; i++) if (slots[i][slot]) ok = false;
					if (ok) break;
					slot++;
				}
				for (let i = si; i < ei; i++) slots[i][slot] = it;
			}
			const max = Math.max(...slots.map((s) => s.length), 0);
			for (let i = 0; i < 7; i++) for (let j = 0; j < max; j++) if (slots[i][j] === undefined) slots[i][j] = null;
			return { rowDays, slots };
		});
		return { monthStart, rowSlots };
	});

	const titleOf = (item: any) => {
		const title = String(item?.Json?.title || '').trim() || '제목없음';
		const gk = groupKey || item?.groupKey || '';
		const c1 = Number(item?.Json?.category?.c1) || 0;
		if (!gk || !c1) return title;
		const codes = getWorkCodes(gk);
		if (codes.length <= 1) return title;
		const name = String(codes.find((c) => c.id === c1)?.name || '').trim();
		if (!name) return title;
		return (title.split(')')[0]?.trim() || '').toLowerCase() === name.toLowerCase() ? title : `${name}) ${title}`;
	};
	const faces = (ids: number[]) => (ids || []).map((id) => admins.find((m) => m.adminId === id)).filter(Boolean) as NotiMember[];

	// 막대에 올리면 미리보기(useCalendarPopoverTooltip) — 막대→창 사이를 지나갈 시간 200ms
	let hoverId = $state<any>(null);
	let pop = $state<{ item: any; x: number; y: number } | null>(null);
	let closeT: any = null;
	const cancelClose = () => (clearTimeout(closeT), (closeT = null));
	const scheduleClose = () => (cancelClose(), (closeT = setTimeout(() => ((pop = null), (hoverId = null)), 200)));
	const todayYmd = ymd(new Date());
	const isThisMonth = $derived(ymd(current) == todayYmd);
	const cap = 'MuiTypography-root MuiTypography-caption';
	const CELL_BORDER = '1px solid var(--adm-divider)';
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16, minHeight: 600 }))}>
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 16, marginBottom: 24 }))}>
		<IconButton onclick={() => (current = new Date(current.getFullYear(), current.getMonth() - 1, current.getDate()))}><Icon name="chevron_left" /></IconButton>
		<h5 class={cx('MuiTypography-root MuiTypography-h5', toClass({ fontWeight: 'bold', minWidth: 150, textAlign: 'center' }))}>{current.getFullYear()}년 {current.getMonth() + 1}월</h5>
		<IconButton onclick={() => (current = new Date(current.getFullYear(), current.getMonth() + 1, current.getDate()))}><Icon name="chevron_right" /></IconButton>
		<Button variant="outlined" size="small" disabled={isThisMonth} sx={{ marginLeft: 16 }} onclick={() => (current = new Date())}>이번달</Button>
	</div>

	<table
		class={cx(
			'MuiTable-root',
			toClass({
				tableLayout: 'fixed',
				border: CELL_BORDER,
				'& th, & td': { border: CELL_BORDER, padding: 8, verticalAlign: 'top' },
				'& th': { background: 'var(--adm-n50, #f8f9fa)', paddingTop: 12, paddingBottom: 12 }
			})
		)}
	>
		<thead class="MuiTableHead-root">
			<tr class="MuiTableRow-root MuiTableRow-head">
				{#each weekDays as d, i (d)}
					<th
						class="MuiTableCell-root MuiTableCell-head MuiTableCell-alignCenter"
						style="color:{isSun(i) ? 'var(--adm-error-main)' : isSat(i) ? 'var(--adm-info-main)' : 'var(--adm-text)'};font-weight:bold">{d}</th
					>
				{/each}
			</tr>
		</thead>
		<tbody class="MuiTableBody-root">
			{#each grid.rowSlots as row, ri (ri)}
				{@const wS = ymd(row.rowDays[0])}
				{@const wE = ymd(row.rowDays[6])}
				{@const gridStart = ymd(grid.rowSlots[0].rowDays[0])}
				<tr class="MuiTableRow-root" style="height:120px">
					{#each row.rowDays as day, idx (ymd(day))}
						{@const ds = ymd(day)}
						{@const holy = holyMap[ds]}
						{@const isToday = ds == todayYmd}
						{@const inMonth = day.getMonth() == grid.monthStart.getMonth()}
						<td class="MuiTableCell-root MuiTableCell-body" style="width:14.28%;position:relative;background:{isToday ? 'var(--theme-color-alpha4, rgba(99,102,241,0.04))' : 'var(--adm-paper)'}">
							<div class={cx('MuiStack-root', toClass({ alignItems: holy ? 'baseline' : 'center', marginBottom: 8, flexDirection: holy ? 'row' : 'column' }))}>
								<p
									class={cx(
										'MuiTypography-root MuiTypography-body2',
										toClass({ textAlign: 'center', borderRadius: '50%', width: 24, height: 24, lineHeight: '24px', fontWeight: isToday ? 'bold' : 'normal' })
									)}
									style="color:{isToday
										? '#fff'
										: !inMonth
											? 'var(--adm-text-disabled, rgba(17,25,39,0.38))'
											: holy || isSun(idx)
												? 'var(--adm-error-main)'
												: isSat(idx)
													? 'var(--adm-info-main)'
													: 'var(--adm-text)'};background:{isToday ? 'var(--theme-color-main)' : 'transparent'}"
								>
									{day.getDate()}
								</p>
								{#if holy}<span class={cx(cap, toClass({ color: 'var(--adm-error-main)', fontSize: '0.65rem', fontWeight: 'bold', textAlign: 'center', lineHeight: 1 }))}>{holy}</span>{/if}
							</div>
							{#if writeable && setting?.useCalendarWriteButton}
								<AddWorkButton
									{groupKey}
									category={setting?.defaultSearchOption?.category || { c1: 0, c2: 0, c3: 0 }}
									defaultSchedule={{ from: ds, to: ds }}
									size="small"
									variant="text"
									label="글쓰기"
									sx={{
										position: 'absolute',
										right: 8,
										bottom: 4,
										width: 'fit-content',
										padding: '0 !important',
										minWidth: 'unset',
										opacity: 0.5,
										fontSize: '0.75rem',
										'&:hover': { background: 'transparent', opacity: 1 },
										zIndex: 1000
									}}
								/>
							{/if}
							<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
								{#each row.slots[idx] as item, si (si)}
									{#if !item}
										<div class="MuiBox-root" style="height:24px"></div>
									{:else}
										{@const from = item.Json.schedule.from}
										{@const to = item.Json.schedule.to || from}
										{@const isStart = from === ds}
										{@const isEnd = to === ds}
										{@const multi = from !== to}
										{@const visStart = from < gridStart ? gridStart : from}
										{@const showTitle = ds === visStart || !multi}
										{@const segEnd = to > wE ? wE : to}
										{@const segLen = showTitle ? Math.round((new Date(segEnd).getTime() - new Date(ds).getTime()) / 86400000) + 1 : 1}
										{@const ctx = ctxOf(item)}
										{@const sc = ctx.statusCodes.find((c: any) => c.id === item.Json.status)}
										{@const id = item._eventId || item.workId}
										<a
											class={cx(
												'MuiStack-root',
												toClass({
													flexDirection: 'row',
													alignItems: 'center',
													padding: '4px 8px',
													textDecoration: 'none',
													position: 'relative',
													height: 24,
													boxSizing: 'border-box',
													'&:hover': { filter: 'brightness(0.9)' }
												})
											)}
											style="background:{sc?.bgColor || 'var(--theme-color-alpha16, rgba(99,102,241,0.15))'};color:{sc?.color ||
												(multi ? 'var(--theme-color-main)' : 'var(--adm-text)')};border-radius:{isStart && isEnd ? '8px' : isStart ? '4px 0 0 4px' : isEnd ? '0 4px 4px 0' : '0'};margin-left:{multi &&
											!isStart
												? -8
												: 0}px;margin-right:{multi && !isEnd ? -8 : 0}px;opacity:{item._isShape ? 0.5 : 1};filter:{hoverId === id ? 'brightness(0.9)' : 'none'};overflow:{multi && showTitle
												? 'visible'
												: 'hidden'};z-index:{showTitle ? 2 : 1}"
											href={ctx.setting?.useCalendarPopoverTooltip && !ctx.writeable
												? undefined
												: `/admin/work/${item.groupKey || groupKey}/${item.workId}?ymd=${ds}${item._shapeId ? `&shapeId=${item._shapeId}` : ''}`}
											onmouseenter={(e) => {
												cancelClose();
												hoverId = id;
												if (ctx.setting?.useCalendarPopoverTooltip) {
													const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
													pop = { item, x: r.left, y: r.top };
												}
											}}
											onmouseleave={() => (ctx.setting?.useCalendarPopoverTooltip ? scheduleClose() : (hoverId = null))}
										>
											{#if showTitle}
												<div
													class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center' }))}
													style={multi ? `width:calc(${segLen * 100}% + ${(segLen - 1) * 16}px);flex-shrink:0` : 'flex:1;min-width:0'}
												>
													{#if faces(item.Json.adminIds).length}
														<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', marginRight: 4, '& > :not(style) ~ :not(style)': { marginLeft: -8 } }))}>
															{#each faces(item.Json.adminIds) as m (m.adminId)}
																<div
																	class={cx(
																		'MuiAvatar-root MuiAvatar-circular',
																		toClass({
																			position: 'relative',
																			width: 20,
																			height: 20,
																			background: 'var(--gt-300, #ddd)',
																			color: '#fff',
																			border: '1px solid #fff',
																			borderRadius: '50%',
																			overflow: 'hidden',
																			flexShrink: 0,
																			display: 'flex',
																			alignItems: 'center',
																			justifyContent: 'center'
																		})
																	)}
																>
																	<!-- MUI Avatar 와 같게: 사진이 없거나 못 읽으면 사람 모양 -->
																	<svg class="MuiAvatar-fallback" style="width:75%;height:75%" viewBox="0 0 24 24" aria-hidden="true"
																		><path fill="currentColor" d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4m0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4"></path></svg
																	>
																	{#if m.profileImage}<img
																			src={m.profileImage}
																			alt=""
																			style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
																			onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
																		/>{/if}
																</div>
															{/each}
														</div>
													{/if}
													{#if setting?.useCalendarPopoverTooltip}
														<span class={cx(cap, toClass({ fontSize: 11, fontWeight: 'bold', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', minWidth: 0 }))}>{titleOf(item)}</span>
													{:else}
														<Tooltip title={titleOf(item)} placement="top"
															><span class={cx(cap, toClass({ fontSize: 11, fontWeight: 'bold', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', minWidth: 0, display: 'block' }))}
																>{titleOf(item)}</span
															></Tooltip
														>
													{/if}
												</div>
											{:else}
												<span class={cap} style="visibility:hidden">-</span>
											{/if}
										</a>
									{/if}
								{/each}
							</div>
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>

	{#if !setting?.disableCalendarHolyday && monthHoly.length > 0}
		<div class={cx('MuiStack-root', toClass({ paddingBottom: 16, paddingTop: 8, borderTop: CELL_BORDER, '& > :not(style) ~ :not(style)': { marginTop: 8 } }))}>
			<span class={cx('MuiTypography-root MuiTypography-overline', toClass({ fontWeight: 'bold', color: 'var(--adm-text2)' }))}>휴일 정보</span>
			<div class="MuiBox-root" style="display:flex;flex-wrap:wrap;gap:8px">
				{#each monthHoly as h (h.ymd)}
					{@const d = new Date(h.ymd)}
					<div
						class={cx(
							'MuiStack-root',
							toClass({
								flexDirection: 'row',
								alignItems: 'center',
								gap: 8,
								background: 'rgba(240,68,56,0.05)',
								color: 'var(--adm-error-main)',
								padding: '4px 12px',
								borderRadius: 8,
								border: '1px solid rgba(240,68,56,0.1)'
							})
						)}
					>
						<span class={cx(cap, toClass({ fontWeight: 'bold' }))}>{d.getMonth() + 1}/{d.getDate()}</span>
						<span class={cx(cap, toClass({ fontWeight: 'bold' }))}>{h.name}</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

{#if pop}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		use:portal
		class={cx(
			'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation8 MuiPopover-paper',
			toClass({
				position: 'fixed',
				zIndex: 1300,
				background: 'var(--adm-paper)',
				borderRadius: 8,
				boxShadow: '0px 5px 5px -3px rgba(0,0,0,0.2),0px 8px 10px 1px rgba(0,0,0,0.14),0px 3px 14px 2px rgba(0,0,0,0.12)',
				transform: 'translateY(-100%)'
			})
		)}
		style="left:{pop.x}px;top:{pop.y}px"
		onmouseenter={cancelClose}
		onmouseleave={scheduleClose}
	>
		<div class="MuiBox-root" style="padding:16px;max-width:350px">
			<h6 class={cx('MuiTypography-root MuiTypography-subtitle1 MuiTypography-gutterBottom', toClass({ fontWeight: 'bold', marginBottom: '0.35em' }))}>{pop.item.Json.title || '제목없음'}</h6>
			<div
				class={toClass({
					fontSize: '0.8125rem',
					color: 'var(--adm-text2)',
					maxHeight: 200,
					overflowY: 'auto',
					'& *': { margin: 0, padding: 0 },
					'& table': { borderCollapse: 'collapse' },
					'& td, & th': { border: '1px solid var(--adm-divider)', padding: '2px 4px' }
				})}
			>
				{@html pop.item.Json.desc || ''}
			</div>
		</div>
	</div>
{/if}
