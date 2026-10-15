<script lang="ts">
	/**
	 * 메모(텍스트) 편집(React ModalMemoEditor) — 제목 · 연관 업무(여러 개, 찾아서 고르기) · 본문(질문 표시는 2026-10-04 뺌)(TinyMCE, 그림은 `work/work_{id}/{메모}`) · 색 · 상태.
	 * 연관 업무를 넣고 빼면 두 카드 모두의 Json.targetWorkIds 를 맞춘다(store linkWorks — 서로 바로가기).
	 * ⚠ React 는 크기 바꾸는 창(ResizableModal) · 바뀐 곳 비교(diff) 미리보기 · @멘션 알림이 있었다 — 여기는 일반 창.
	 */
	import { untrack } from 'svelte';
	import TinyMceEditor from '#lib/shared/editor/TinyMceEditor.svelte';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import { adminImagePrefix } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import { media } from '../../media.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import Popover from '../../ui/Popover.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import Toggle from '../ui/Toggle.svelte';
	import Autocomplete from '#lib/shared/ui/Autocomplete.svelte';
	import { clone, getWorkSettings, getWorkStatusCodes } from './codes.ts';
	import type { WorkItem, Shape } from '#lib/shared/types/work.ts';
	import StatusPopover from './StatusPopover.svelte';
	import { linkWorks, shapeTargets as targetsOf } from './store.svelte.ts';

	let {
		shape,
		groupKey,
		workId,
		items = [],
		onsave,
		onclose
	}: { shape: Shape | null; groupKey: string; workId: number; items?: WorkItem[]; onsave: (s: Shape) => Promise<void> | void; onclose: (s: Shape) => void } = $props();
	let orig = $state<Shape | null>(null);
	let item = $state<Shape | null>(null);
	let ver = $state(0);
	let ed = $state<any>();
	let picker = $state<any>();
	let colorAnchor = $state<HTMLElement | null>(null);
	$effect(() => {
		if (!shape) return;
		orig = clone(shape);
		item = clone(shape);
		// ver++ 는 ver 를 읽기도 해서 그대로 두면 이 효과가 자기를 다시 부른다(무한 반복 → 500) — 읽기는 추적하지 않는다
		untrack(() => ver++);
	});
	const changed = $derived(!!orig && !!item && JSON.stringify(orig) !== JSON.stringify(item));
	const setting = $derived(getWorkSettings(groupKey));
	const st = $derived(getWorkStatusCodes(groupKey));
	const folder = $derived('work_' + workId);
	const COLORS = [
		['#000000', '검정'],
		['#ef4444', '빨강'],
		['#f59e0b', '주황'],
		['#eab308', '노랑'],
		['#10b981', '초록'],
		['#06b6d4', '청록'],
		['#3b82f6', '파랑'],
		['#8b5cf6', '보라'],
		['#ec4899', '분홍'],
		['#6b7280', '회색']
	];
	async function save() {
		if (!item || !orig) return;
		const x = { ...clone(item), ts: Date.now() };
		const oldIds = targetsOf(orig);
		const newIds = targetsOf(x);
		await onsave(x);
		item = x;
		orig = clone(x);
		// 메모를 저장한 뒤 두 카드 모두에 연관 업무를 걸고 푼다(메모가 먼저 저장돼야 「아직 가리키는 메모」 판단이 맞다)
		for (const id of newIds.filter((v) => !oldIds.includes(v))) await linkWorks(workId, id, true);
		for (const id of oldIds.filter((v) => !newIds.includes(v))) await linkWorks(workId, id, false);
	}
	async function close(force = false) {
		if (!orig) return;
		if (changed && !force) {
			if (!(await confirm('변경된 내용이 있습니다. 취소하시겠습니까?'))) return;
			onclose(orig);
		} else onclose(item!);
		orig = null;
		item = null;
	}
	export function hasChanges() {
		return changed;
	}
	export function forceClose() {
		return close(true);
	}
	const others = $derived(items.filter((x) => x.groupKey === groupKey && x.workId !== workId).sort((a, b) => b.workId - a.workId));
	const picked = $derived(item ? targetsOf(item) : []);
	const targetOptions = $derived(others.filter((o) => !picked.includes(o.workId)).map((o) => ({ label: `#${o.workId} ${o.Json.title || ''}`, value: String(o.workId) })));
	// 고친 뒤엔 새 칸(targetWorkIds)만 — 옛 하나짜리 칸은 지운다
	function setTargets(ids: number[]) {
		if (!item) return;
		const { targetWorkId: _old, ...rest } = item;
		item = { ...rest, targetWorkIds: ids };
	}
	const titleOf = (id: number) => items.find((x) => x.workId === id)?.Json.title || '';
	// 휴대폰은 전체 화면 — 본문 칸이 남는 높이를 채운다(보이는 높이 기준이라 키보드가 뜨면 줄어든다)
	let vh = $state(typeof window != 'undefined' ? window.innerHeight : 800);
	$effect(() => {
		const vv = window.visualViewport;
		const on = () => (vh = vv?.height || window.innerHeight);
		on();
		(vv || window).addEventListener('resize', on);
		return () => (vv || window).removeEventListener('resize', on);
	});
	const editorHeight = $derived(media.mdDown ? Math.max(240, Math.round(vh - (item?.parentShapeId ? 230 : 290))) : 420);
</script>

<Dialog open={!!item} onclose={() => close()} maxWidth="md" fullWidth mobileFullScreen title={item?.title || '제목업음'}>
	{#if item}
		<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:8px">
			{#if !item.parentShapeId}<OutlinedField placeholder="제목을 입력하세요" value={item.title || ''} oninput={(v: string) => item && (item = { ...item, title: v })} />{/if}
			<div class="targets">
				<Icon name="link" size="Small" color="Action" />
				{#each picked as id (id)}
					<span class="tchip" title={titleOf(id)}>
						#{id}
						{titleOf(id)}
						<button type="button" aria-label="빼기" onclick={() => setTargets(picked.filter((x) => x !== id))}><Icon name="close" size="Small" sx={{ fontSize: 14 }} /></button>
					</span>
				{/each}
				<Autocomplete
					class="tac"
					options={targetOptions}
					freeSolo={false}
					placeholder={picked.length ? '연관 업무 더 넣기' : '연관 업무 찾기 (번호·제목)'}
					onselect={(v) => v && setTargets([...picked, Number(v)])}
				/>
			</div>
			{#key ver}
				<TinyMceEditor
					bind:this={ed}
					body={item.html || ''}
					height={editorHeight}
					onchange={(h) => item && item.html !== h && (item = { ...item, html: h })}
					onImageUI={() => picker?.open_()}
					imageId={folder}
				/>
			{/key}
		</div>
		<div style="display:none">
			<ImagePicker
				bind:this={picker}
				title="본문에 사용할 이미지를 선택하세요"
				size={{ width: 'all', height: 'auto' }}
				{folder}
				preFolder="work"
				subFolder={String(item.id)}
				isMultiple
				useOriginalFileName
				maxWidth={1200}
				onSelect={(imgs: any[]) => {
					ed?.insert('<br/>' + imgs.map((i) => `<img style="max-width: 100%;" src="${adminImagePrefix(admin.profile?.hostingId)}/work/${folder}/${item?.id}/${i?.name}"/><br/>`).join(''));
					picker?.close();
				}}
			/>
		</div>
	{/if}
	{#snippet actions()}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;justify-content:space-between;width:100%">
			<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px">
				<Button variant="outlined" onclick={() => close()}>취소</Button>
				<Button variant="contained" disabled={!changed} onclick={save}>저장</Button>
			</div>
			{#if item && !setting?.disableShapeStatus}
				<div class="MuiStack-root" style="display:flex;flex-direction:row;flex-grow:1;justify-content:end;align-items:center;gap:8px">
					<Tooltip title={`색상: ${item.color || '#000000'}`} placement="top">
						<IconButton
							size="Small"
							sx={{ width: 28, height: 28, padding: 0, backgroundColor: item.color || '#000000', border: '1px solid #d1d5db', borderRadius: '50%' }}
							onclick={(e: MouseEvent) => (colorAnchor = e.currentTarget as HTMLElement)}><span></span></IconButton
						>
					</Tooltip>
					<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
						<Toggle
							checked={!!item.useStatus}
							onchange={(v) => {
								if (!item) return;
								if (!v) {
									const { status, ...rest } = item;
									item = { ...rest, useStatus: false, color: 'var(--zc-000000, #000000)' } as Shape;
								} else item = { ...item, useStatus: true };
							}}
						/>
						<span class="MuiTypography-root MuiTypography-body1 MuiFormControlLabel-label">상태 사용</span>
					</label>
					{#if item.useStatus}<StatusPopover codes={st.codes} value={item.status || ''} onchange={(v) => item && (item = { ...item, status: v })} />{/if}
				</div>
			{/if}
		</div>
	{/snippet}
</Dialog>
<Popover anchor={colorAnchor} onclose={() => (colorAnchor = null)} align="center">
	<div style="padding:8px;width:200px">
		<span class="MuiTypography-root MuiTypography-caption" style="font-weight:700;display:block;margin-bottom:4px">shape 색상</span>
		<div style="display:grid;grid-template-columns:repeat(5,1fr);gap:4px;margin-bottom:8px">
			{#each COLORS as [v, l] (v)}
				<button
					type="button"
					title={l}
					aria-label={l}
					style="width:28px;height:28px;padding:0;cursor:pointer;background:{v};border:{item?.color === v ? '2px solid #0369a1' : '1px solid #d1d5db'};border-radius:50%"
					onclick={() => (item && (item = { ...item, color: v }), (colorAnchor = null))}
				></button>
			{/each}
		</div>
		<div style="display:flex;align-items:center;gap:4px">
			<span class="MuiTypography-root MuiTypography-caption" style="color:#6b7280">직접 선택:</span>
			<input
				type="color"
				value={item?.color?.startsWith('#') ? item.color : '#000000'}
				oninput={(e) => item && (item = { ...item, color: (e.currentTarget as HTMLInputElement).value })}
				style="width:32px;height:24px;border:none;padding:0;cursor:pointer"
			/>
			<span class="MuiTypography-root MuiTypography-caption" style="color:#6b7280;flex:1;text-align:right">{item?.color}</span>
		</div>
	</div>
</Popover>

<style>
	.targets {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}
	.tchip {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		max-width: 100%;
		padding: 3px 4px 3px 10px;
		border: 1px solid var(--adm-divider, #e0e0e0);
		border-radius: 999px;
		font-size: 13px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tchip button {
		display: inline-flex;
		padding: 2px;
		border: 0;
		border-radius: 50%;
		background: none;
		color: inherit;
		cursor: pointer;
	}
	.tchip button:hover {
		background: var(--gt-200, #f2f2f2);
	}
	.targets :global(.tac) {
		flex: 1 1 240px;
		min-width: 200px;
	}
</style>
