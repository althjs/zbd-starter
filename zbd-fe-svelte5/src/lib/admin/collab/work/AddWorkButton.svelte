<script lang="ts">
	/**
	 * 새 문서 추가 · 문서 수정 · 이미지 추가/수정(React AddWorkModalButton).
	 *   work 모드: 제목 + 붙여 넣은 그림(그림판 바탕) — 보드 설정 disableWork 면 그림 대신 열 값 + 본문(TinyMCE)
	 *   subImage 모드: 추가 이미지(탭) 하나
	 * 저장: 새 문서는 POST 로 만들고 그 번호로 그림을 `work/work_{id}` 에 올린 뒤 PUT.
	 */
	import TinyMceEditor from '#lib/shared/editor/TinyMceEditor.svelte';
	import { blobToDataUrl, uploadFiles } from '#lib/shared/image/files.ts';
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import { adminImagePrefix } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import { clone, getWorkCodeIds, getWorkCodes, getWorkColumns, getWorkSettings } from './codes.ts';
	import type { WorkItem, Category, Schedule } from '#lib/shared/types/work.ts';
	import RenderCell from './RenderCell.svelte';
	import { addItem, updateItem } from './store.svelte.ts';

	let {
		groupKey,
		category,
		mode = 'work',
		subImageIndex = undefined,
		item = undefined,
		defaultTitle = '',
		defaultSchedule = undefined,
		onsuccess,
		variant = 'contained',
		size = 'medium',
		disabled = false,
		label = '',
		startIcon: startIconName = '',
		sx = null,
		onclick
	}: {
		groupKey: string;
		category: Category | undefined;
		mode?: 'work' | 'subImage';
		subImageIndex?: number;
		item?: WorkItem;
		defaultTitle?: string;
		defaultSchedule?: Schedule;
		onsuccess?: (name?: string) => void;
		variant?: 'text' | 'outlined' | 'contained';
		size?: 'small' | 'medium';
		disabled?: boolean;
		label?: string;
		startIcon?: string;
		sx?: Sx | null;
		onclick?: (e: MouseEvent) => void;
	} = $props();

	const isEdit = $derived(!!item && mode === 'work');
	const isSub = $derived(mode === 'subImage');
	const isSubEdit = $derived(isSub && subImageIndex !== undefined);
	const setting = $derived(getWorkSettings(groupKey));
	const codes = $derived(getWorkCodes(groupKey));
	const codeIds = $derived(getWorkCodeIds(groupKey));
	const cols = $derived(
		getWorkColumns(groupKey, ['status'], ['createdAt', 'updatedAt']).map((c) => ({
			key: c.key || `column_${c.id}`,
			name: c.name,
			width: Number(c.value) || 150,
			id: c.id,
			items: c.items,
			inputType: c.inputType,
			sticky: false
		}))
	);

	let open = $state(false);
	let title = $state('');
	let desc = $state('');
	let formJson = $state<any>({});
	let file = $state<{ blob: Blob; name: string } | null>(null);
	let preview = $state('');
	let loading = $state(false);
	const today = () => {
		const d = new Date();
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	};
	const imgUrl = (name: string) => `${adminImagePrefix(admin.profile?.hostingId)}/work/work_${item!.workId}/${name}`;

	function show(e: MouseEvent) {
		onclick?.(e);
		file = null;
		preview = '';
		if (isSubEdit) {
			const s = item!.Json.images?.[subImageIndex!];
			title = s?.title || '';
			desc = s?.desc || '';
			formJson = clone(s || {});
			if (s?.name) preview = imgUrl(s.name);
		} else if (isEdit) {
			title = item!.Json.title || '';
			desc = item!.Json.desc || '';
			formJson = clone(item!.Json || {});
			if (item!.Json.image) preview = imgUrl(item!.Json.image);
		} else {
			title = defaultTitle || '';
			desc = '';
			const t = today();
			formJson = { title: defaultTitle || '', category, status: 1, schedule: defaultSchedule || { from: t, to: t } };
		}
		open = true;
	}
	async function paste(e: ClipboardEvent) {
		for (const it of Array.from(e.clipboardData?.items || [])) {
			if (it.type.indexOf('image') === -1) continue;
			const f = it.getAsFile();
			if (!f) continue;
			file = { blob: f, name: `${Date.now()}.png` };
			preview = await blobToDataUrl(f);
			if (!title) title = f.name.split('.')[0];
		}
	}
	async function save() {
		if (!title) return;
		loading = true;
		try {
			let cur: WorkItem;
			if (isSub) cur = clone(item!);
			else if (isEdit) cur = { ...clone(item!), Json: { ...item!.Json, ...formJson, title, desc, category } };
			else {
				const made = await addItem({ adminId: 0, hostingId: '', Json: { ...formJson, title, desc, mentionedMembers: [], category }, createdAt: '', updatedAt: '', parentWorkId: 0, groupKey } as any);
				if (!made) throw new Error('add');
				cur = made;
			}
			if (setting?.disableWork) {
				if (isSub) {
					if (isSubEdit) Object.assign(cur.Json.images![subImageIndex!], { title, desc });
					else cur.Json.images = [...(cur.Json.images || []), { name: '', title, desc, mentionedMembers: [], shapes: [] }];
				} else Object.assign(cur.Json, { title, desc });
			} else if (file) {
				const up = await uploadFiles([file], `work_${cur.workId}`, { hostingId: admin.profile?.hostingId, preFolder: 'work' });
				if (!up?.length) throw new Error('upload');
				const name = up[0].name;
				if (isSub) {
					if (isSubEdit) Object.assign(cur.Json.images![subImageIndex!], { name, title });
					else cur.Json.images = [...(cur.Json.images || []), { name, title, desc: '', shapes: [] }];
				} else cur.Json.image = name;
			} else if (isSubEdit) Object.assign(cur.Json.images![subImageIndex!], { title, desc });
			else if (isEdit) Object.assign(cur.Json, { title, desc });
			if (!(await updateItem(cur))) throw new Error('update');
			toast(isSubEdit ? '이미지 정보가 수정되었습니다.' : isSub ? '이미지가 성공적으로 추가되었습니다.' : isEdit ? '성공적으로 수정되었습니다.' : '성공적으로 저장되었습니다.');
			onsuccess?.(file?.name);
			open = false;
		} catch {
			alert('처리 중 오류가 발생했습니다.');
		} finally {
			loading = false;
		}
	}
	const text = $derived(label || (isSubEdit ? '이미지 수정' : isSub ? '이미지 추가' : isEdit ? '수정' : '추가'));
</script>

{#snippet iconSnip()}<span class="material-symbols-outlined notranslate MuiIcon-root MuiIcon-fontSizeSmall" aria-hidden="true">{startIconName}</span>{/snippet}
<Button {variant} {size} {disabled} {sx} onclick={show} startIcon={startIconName ? iconSnip : undefined}>{text}</Button>

<Dialog {open} onclose={() => (open = false)} title={isSubEdit ? '이미지 정보 수정' : isSub ? '추가 이미지 등록' : isEdit ? '문서 수정' : '새 문서 추가'} maxWidth="sm" fullWidth>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="MuiStack-root" style="display:flex;flex-direction:column;margin-top:8px" onpaste={paste}>
		<OutlinedField label="제목" bind:value={title} />
		{#if setting?.disableWork}
			<div style="margin-top:16px">
				{#if cols.length}
					<table class="MuiTable-root" style="width:fit-content;margin-bottom:16px">
						<tbody class="MuiTableBody-root">
							<tr class="MuiTableRow-root">
								{#each cols as col (col.key)}
									<td
										class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium"
										style="padding:0 16px 0 0;border:0;white-space:{col.key === 'summary' || col.key.startsWith('column_') || col.key === 'status' ? 'nowrap' : 'normal'}"
									>
										<RenderCell
											{col}
											item={{ ...(item || ({} as WorkItem)), Json: formJson } as WorkItem}
											shapes={[]}
											summaryText=""
											{groupKey}
											{codes}
											{codeIds}
											onValueChange={async (_i, k, v) => void (formJson = { ...formJson, [k]: v })}
										/>
									</td>
								{/each}
							</tr>
						</tbody>
					</table>
				{/if}
				<TinyMceEditor body={desc} height={400} onchange={(v) => (desc = v)} imageId={`work_add_${item?.workId || 'new'}`} />
			</div>
		{:else}
			<div
				class="MuiBox-root"
				style="margin-top:16px;border:2px dashed #ccc;border-radius:8px;padding:16px;text-align:center;min-height:150px;display:flex;flex-direction:column;align-items:center;justify-content:center;background-color:var(--zc-f9f9f9, #f9f9f9)"
			>
				{#if preview}
					<img src={preview} alt="Preview" style="max-width:100%;max-height:300px;object-fit:contain" />
				{:else}
					<p class="MuiTypography-root MuiTypography-body1" style="color:var(--adm-text2)">이미지를 여기에 붙여넣으세요 (Ctrl+V / Cmd+V)</p>
				{/if}
			</div>
		{/if}
	</div>
	{#snippet actions()}
		<Button color="inherit" onclick={() => (open = false)}>취소</Button>
		<Button variant="contained" disabled={!title || loading} onclick={save}>{loading ? '처리 중...' : isEdit ? '수정' : '저장'}</Button>
	{/snippet}
</Dialog>
