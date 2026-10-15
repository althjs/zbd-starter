<script lang="ts">
	/**
	 * 칸반 보기(React WorkListCanvan) — 상태마다 칸, 카드를 끌어 다른 칸에 놓으면 상태를 바꾼다. 빈 곳을 끌면 옆으로 밀린다.
	 * 상태가 없으면 1(백로그 자리)로 본다(React 와 같게).
	 */
	import { shortDate } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { adminImageUrl } from '../../meta.svelte.ts';
	import { members } from '../members.svelte.ts';
	import Av from '../ui/Av.svelte';
	import { CARD } from '../ui/cls.ts';
	import { clone, getWorkCodes, getWorkStatusCodes, getCategoryLabel, SHAPE_LABELS } from './codes.ts';
	import type { WorkItem } from '#lib/shared/types/work.ts';
	import { updateItem } from './store.svelte.ts';

	let { groupKey, items }: { groupKey: string; items: WorkItem[] } = $props();
	const codes = $derived(getWorkCodes(groupKey));
	const status = $derived(getWorkStatusCodes(groupKey).codes);
	const GT = (n: number, f: string) => `var(--gt-${n}, ${f})`;
	const img = (id: number) => {
		const m = members.list.find((x) => x.adminId === id);
		return m ? adminImageUrl(`admin_profile_${id}`, m.Json?.image?.name, m.hostingId) : null;
	};
	async function drop(workId: string, s: number) {
		const it = items.find((o) => String(o.workId) === workId);
		if (!it || (it.Json.status || 0) === s) return;
		const n = clone(it);
		n.Json.status = s === 0 ? '' : s;
		await updateItem(n);
	}
	let box = $state<HTMLElement>();
	let drag: { x: number; left: number } | null = null;
	const summary = (o: WorkItem) => {
		const m: Record<string, number> = {};
		(o.Json.shapes || []).forEach((s) => (m[s.type] = (m[s.type] || 0) + 1));
		return Object.entries(m)
			.map(([t, n]) => `${SHAPE_LABELS[t] || t} ${n}`)
			.join(', ');
	};
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	bind:this={box}
	class="adm-cv-kanban"
	style="display:flex;gap:16px;overflow-x:auto;min-height:400px;padding-bottom:16px;align-items:flex-start;cursor:grab;user-select:none"
	onmousedown={(e) => {
		if ((e.target as HTMLElement).closest('[draggable]') || !box) return;
		drag = { x: e.pageX - box.offsetLeft, left: box.scrollLeft };
		box.style.cursor = 'grabbing';
	}}
	onmouseleave={() => ((drag = null), box && (box.style.cursor = 'grab'))}
	onmouseup={() => ((drag = null), box && (box.style.cursor = 'grab'))}
	onmousemove={(e) => {
		if (!drag || !box) return;
		e.preventDefault();
		box.scrollLeft = drag.left - (e.pageX - box.offsetLeft - drag.x) * 1.5;
	}}
>
	{#each status as code (code.id)}
		{@const col = items.filter((o) => {
			const s = o.Json.status || 1;
			return s === code.id || (!s && code.id === 0);
		})}
		<div
			style="min-width:320px;max-width:320px;background-color:{GT(100, '#f8f8f8')};border-radius:16px;padding:12px;display:flex;flex-direction:column;gap:12px"
			ondragover={(e) => e.preventDefault()}
			ondrop={(e) => {
				e.preventDefault();
				const id = e.dataTransfer?.getData('workId');
				if (id) drop(id, code.id);
			}}
		>
			<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:8px;padding:0 8px">
				<div style="width:12px;height:12px;border-radius:50%;background-color:{code.bgColor || GT(400, '#d0d0d0')}"></div>
				<h6 class="MuiTypography-root MuiTypography-subtitle2" style="font-weight:bold">{code.name}</h6>
				<span class="MuiTypography-root MuiTypography-caption" style="color:{GT(600, '#a0a0a0')}">{col.length}</span>
			</div>
			<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:12px;min-height:50px">
				{#each col as o (o.workId)}
					{@const shapes = o.Json.shapes || []}
					{@const owner = img(o.adminId)}
					{@const assigned = Object.keys(o.Json.adminNotiMap || {})
						.map(Number)
						.filter((id) => members.list.find((m) => m.adminId === id))}
					<div
						class={cx(CARD, 'adm-cv-kcard', toClass({ padding: 16, display: 'flex', flexDirection: 'column', gap: 8, cursor: 'grab' }))}
						draggable="true"
						role="listitem"
						ondragstart={(e) => e.dataTransfer?.setData('workId', String(o.workId))}
					>
						<div class="MuiStack-root" style="display:flex;flex-direction:row;justify-content:space-between;align-items:center;gap:8px">
							<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px">
								{#if owner !== null}<Av src={owner} size={18} sx={{ backgroundColor: GT(300, '#e0e0e0') }} />{/if}
								<span
									class="MuiTypography-root MuiTypography-caption"
									style="color:{GT(600, '#a0a0a0')};background-color:{GT(
										100,
										'#f8f8f8'
									)};padding:1.6px 6.4px;border-radius:8px;font-weight:500;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:100%">{getCategoryLabel(o.Json.category, codes)}</span
								>
							</div>
						</div>
						<a
							class="MuiTypography-root MuiTypography-body1 adm-cv-ktitle"
							href={`/admin/work/${groupKey}/${o.workId}`}
							style="text-decoration:none;color:var(--adm-text);font-weight:bold;font-size:14px;line-height:1.4;word-break:break-word">{o.Json.title || `Canvas #${o.workId}`}</a
						>
						{#if o.Json.desc}<p
								class="MuiTypography-root MuiTypography-body2"
								style="color:{GT(700, '#909090')};display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;text-overflow:ellipsis;font-size:12px"
							>
								{o.Json.desc.replace(/<[^>]+>/g, ' ')}
							</p>{/if}
						{#if shapes.length || o.createdAt}
							<div
								class="MuiStack-root"
								style="display:flex;flex-direction:row;align-items:center;justify-content:space-between;margin-top:8px;padding-top:8px;border-top:1px solid {GT(100, '#f8f8f8')}"
							>
								<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px">
									{#if shapes.length}<span class="MuiTypography-root MuiTypography-caption" style="color:{GT(600, '#a0a0a0')}">{shapes.length}개 ({summary(o)})</span>{:else}<div></div>{/if}
								</div>
								{#if o.createdAt}
									<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px">
										<span class="MuiTypography-root MuiTypography-caption" style="color:{GT(500, '#b0b0b0')}">{shortDate(o.createdAt)}</span>
										{#if assigned.length}
											<div class="MuiStack-root" style="display:flex;flex-direction:row;margin-left:8px">
												{#each assigned as id, i (id)}<Av
														src={img(id) || ''}
														size={20}
														sx={{ backgroundColor: GT(300, '#e0e0e0'), border: '1px solid #fff', ...(i ? { marginLeft: -8 } : {}) }}
													/>{/each}
											</div>
										{/if}
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/each}
</div>
