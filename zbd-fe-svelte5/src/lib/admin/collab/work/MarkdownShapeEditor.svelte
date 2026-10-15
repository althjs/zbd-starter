<script lang="ts">
	/**
	 * 마크다운 메모 편집(React ModalMarkdownEditor) — 화면을 덮는 창: 글(왼쪽)·미리보기(오른쪽) + 제목 · 상태 사용 · 취소/저장(Ctrl+S).
	 * 닫으면 onclose(편집 중 표시 풀기) — 바꾼 게 남아 있으면 먼저 묻는다.
	 * ⚠ React 는 크기 바꾸는 창(ResizableModal, 처음엔 최대화) 이었다 · Monaco 대신 CodeArea · @멘션 알림 없음.
	 */
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import MdEditor from '../explorer/MdEditor.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import Toggle from '../ui/Toggle.svelte';
	import { clone, getWorkSettings, getWorkStatusCodes } from './codes.ts';
	import type { Shape } from '#lib/shared/types/work.ts';
	import StatusPopover from './StatusPopover.svelte';

	let { shape, groupKey, onsave, onclose }: { shape: Shape | null; groupKey: string; onsave: (s: Shape) => Promise<void> | void; onclose: (s: Shape) => void } = $props();
	let orig = $state<Shape | null>(null);
	let item = $state<Shape | null>(null);
	let html = $state('');
	let title = $state('');
	$effect(() => {
		if (!shape) return;
		orig = clone(shape);
		item = clone(shape);
		html = shape.html || '';
		title = shape.title || '';
	});
	const strip = (s: any) => {
		const { ts, ...r } = s || {};
		return JSON.stringify(r);
	};
	const changed = $derived(!!orig && !!item && strip(orig) !== strip({ ...item, html, title }));
	const setting = $derived(getWorkSettings(groupKey));
	const st = $derived(getWorkStatusCodes(groupKey));
	async function save() {
		if (!item || !changed) return;
		const x = { ...clone(item), html, title, ts: Date.now() };
		await onsave(x);
		item = x;
		orig = clone(x);
	}
	async function close(force = false) {
		if (!orig) return;
		if (changed && !force && !(await confirm('변경된 내용이 있습니다. 취소하시겠습니까?'))) return;
		onclose(changed ? orig : ({ ...item, html, title } as Shape));
		orig = null;
		item = null;
	}
	export function hasChanges() {
		return changed;
	}
	export function forceClose() {
		return close(true);
	}
</script>

<Dialog open={!!item} onclose={() => close()} fullScreen title={item?.title || '마크다운 에디터'}>
	{#if item}
		<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:8px">
			<MdEditor markdown={html} onChange={(v) => (html = v)} doSave={() => save()} height="calc(100vh - 220px)" minHeight="300px" resizeKey="work" />
		</div>
	{/if}
	{#snippet actions()}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;justify-content:space-between;align-items:center;width:100%">
			<div style="flex:1;margin-right:16px"><OutlinedField label="제목" bind:value={title} /></div>
			<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;align-items:center">
				{#if item && !setting?.disableShapeStatus}
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
				{/if}
				<Button variant="outlined" onclick={() => close()}>취소</Button>
				<Button variant="contained" disabled={!changed} onclick={save}>저장</Button>
			</div>
		</div>
	{/snippet}
</Dialog>
<svelte:window
	onkeydown={(e) => {
		if (item && (e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
			e.preventDefault();
			save();
		}
	}}
/>
