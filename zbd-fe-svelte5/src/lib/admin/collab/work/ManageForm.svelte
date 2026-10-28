<script lang="ts">
	/**
	 * 문서 관리(React WorkManageForm) — 분류 바꾸기 · 제목 · 상태 · 설명 추가/수정/문서 삭제 · 보드 열 값 · 설명(본문) · 이미지 탭 · 메모 목록 · 텍스트/마크다운 추가.
	 * 설명을 고치는 동안 그 칸에 「○○ 님이 설명을 수정중입니다」 가 다른 사람에게 보인다(descActiveEditorName).
	 * 이미지 위에 그리기 = ImageCanvasEditor(도형 그리기·옮기기·복사) — 도형 하나씩 updateShape 로 저장, 편집은 아래 메모·마크다운 창(openEditor).
	 * 「수정」 = 설명(본문) 편집(2026-10-04 대표 — 예전엔 제목·그림 창이 열렸다, 그 창은 「이미지 추가」 로).
	 * 「파일 첨부」 = 본문 편집과 따로(2026-10-04 — Svelte 로 옮기며 빠져 있었다. React 는 편집기 아래 「파일 추가」). 파일 종류 제한 없음(대표), 한 개 1GB(서버 상한) · 보안 폴더 `secure/{호스팅}/work_{번호}` → Json.files(이미지 탭이면 images[i].files).
	 */
	import { goto } from '$app/navigation';
	import TinyMceEditor from '#lib/shared/editor/TinyMceEditor.svelte';
	import HtmlView from '#lib/shared/html/HtmlView.svelte';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import Select from '#lib/shared/mui/Select.svelte';
	import { api } from '#lib/shared/api.ts';
	import { secureDelete, secureUpload } from '../../content/files.ts';
	import { adminImagePrefix, getMeta } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import Tabs from '../../ui/Tabs.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import Badge from '../ui/Badge.svelte';
	import AddWorkButton from './AddWorkButton.svelte';
	import CategoryPopover from './CategoryPopover.svelte';
	import { clone, getWorkCodeIds, getWorkCodes, getWorkColumns, getWorkSettings, getWorkStatusMap, getCategoryLabel, isWorkWriteable, toDescHtml } from './codes.ts';
	import type { WorkItem, NotiMember, Shape } from '#lib/shared/types/work.ts';
	import ImageCanvasEditor from './ImageCanvasEditor.svelte';
	import InlineTitle from './InlineTitle.svelte';
	import JsonDrawer from './JsonDrawer.svelte';
	import MarkdownShapeEditor from './MarkdownShapeEditor.svelte';
	import MemoShapeEditor from './MemoShapeEditor.svelte';
	import RenderCell, { type Col } from './RenderCell.svelte';
	import { withIssuer } from './codes.ts';
	import ShapeList from './ShapeList.svelte';
	import { deleteItem, updateItem, updateShape } from './store.svelte.ts';

	let {
		item,
		items = [],
		subImageIndex = null,
		paramWorkId = 0,
		inviteWorkId = 0,
		members = []
	}: { item: WorkItem; items?: WorkItem[]; subImageIndex?: number | null; paramWorkId?: number; inviteWorkId?: number; members?: NotiMember[] } = $props();

	const profile = $derived(admin.profile);
	const groupKey = $derived(item.groupKey);
	const codes = $derived(getWorkCodes(groupKey));
	const codeIds = $derived(getWorkCodeIds(groupKey));
	const setting = $derived(getWorkSettings(groupKey));
	const writeable = $derived(isWorkWriteable(groupKey, profile));
	const cols = $derived(
		withIssuer(
			getWorkColumns(groupKey, ['status']).map(
				(c) => ({ key: c.key || `column_${c.id}`, name: c.name, width: Number(c.value) || 150, id: c.id, items: c.items, inputType: c.inputType, sticky: false }) as Col
			)
		)
	);
	const statusItems = $derived(
		Object.entries(getWorkStatusMap(groupKey))
			.map(([key, value]) => ({ key, value }))
			.sort((a, b) => (a.key < b.key ? -1 : a.key > b.key ? 1 : 0))
	);
	let tab = $state(0);
	$effect(() => {
		tab = subImageIndex !== null && subImageIndex !== undefined ? subImageIndex + 1 : 0;
	});
	let isEdit = $state(false);
	let draft = $state('');
	let jsonItem = $state<WorkItem | null>(null);
	let picker = $state<any>();
	let ed = $state<any>();

	const isMain = $derived(tab === 0);
	const cur = $derived(isMain ? null : item.Json.images?.[tab - 1] || null);
	const content = $derived(toDescHtml((isMain ? item.Json.desc : cur?.desc) || ''));
	const curFiles = $derived((isMain ? item.Json.files : cur?.files) || []);
	const editorName = $derived(isMain ? item.Json.descActiveEditorName : cur?.descActiveEditorName);
	const likes = $derived((isMain ? item.Json.likeAdminIds : cur?.likeAdminIds) || []);
	const editingMe = $derived(editorName === profile?.name);
	const latest = () => items.find((x) => x.workId === item.workId) || item;
	const siteMeta = $derived(getMeta('SITE') || {});

	async function onValueChange(it: WorkItem, key: string, v: any) {
		const n = clone(it);
		n.Json[key] = v;
		await updateItem(n);
	}
	async function changeCategory(category: any) {
		const ok = await confirm({ title: '분류를 변경하시겠습니까?', body: `기존: ${getCategoryLabel(item.Json.category, codes)}\n변경: ${getCategoryLabel(category, codes)}` });
		if (ok) updateItem({ ...item, Json: { ...item.Json, category } });
	}
	function addDesc() {
		if (isMain) updateItem({ ...item, Json: { ...item.Json, desc: '', descActiveEditorName: profile?.name } });
		else {
			const images = clone(item.Json.images || []);
			images[tab - 1].desc = '';
			updateItem({ ...item, Json: { ...item.Json, images, descActiveEditorName: profile?.name } });
		}
		draft = '';
		isEdit = true;
	}
	function startEdit() {
		if (editorName && profile?.name !== editorName) return;
		const n = clone(item);
		if (isMain) n.Json.descActiveEditorName = profile?.name;
		else n.Json.images![tab - 1].descActiveEditorName = profile?.name;
		updateItem({ ...n, ignoreUpdatedAt: true });
		draft = content || '';
		isEdit = true;
	}
	function finishEdit() {
		const l = clone(latest());
		const same = content === draft;
		if (isMain) updateItem({ ...l, Json: { ...l.Json, desc: draft || content, descActiveEditorName: '' }, ignoreUpdatedAt: same });
		else {
			const im = l.Json.images![tab - 1];
			im.desc = draft || content;
			im.descActiveEditorName = '';
			if (!same) im.ts = Date.now();
			updateItem({ ...l, ignoreUpdatedAt: same });
		}
		isEdit = false;
	}
	async function toggleDescLike() {
		const n = clone(item);
		const me = Number(profile?.adminId);
		const target = isMain ? n.Json : n.Json.images![tab - 1];
		const ids = target.likeAdminIds || [];
		target.likeAdminIds = ids.includes(me) ? ids.filter((x: number) => x !== me) : [...ids, me];
		await updateItem({ ...n, ignoreUpdatedAt: true });
	}
	async function removeDoc() {
		const ok = await confirm({ title: `"${item.Json.title || '제목없음'}" 문서를 삭제하시겠습니까?`, body: '삭제된 데이터는 복구하실 수 없습니다.\n삭제를 진행하시겠습니까?' });
		if (!ok) return;
		await deleteItem(item.workId);
		toast('정상적으로 삭제되었습니다.');
		if (paramWorkId) goto(`/admin/work/${groupKey}`);
	}
	async function removeImage(idx: number) {
		if (!(await confirm('이미지를 삭제하시겠습니까? 관련 도형 정보도 모두 삭제됩니다.'))) return;
		const images = [...(item.Json.images || [])];
		images.splice(idx, 1);
		updateItem({ ...item, Json: { ...item.Json, images } });
		if (tab === idx + 1) tab = 0;
		else if (tab > idx + 1) tab = tab - 1;
	}
	let fileInput = $state<HTMLInputElement>();
	let attaching = $state(false);
	/** 지금 탭(본문·추가 이미지)의 첨부 목록을 최신 카드에 바꿔 넣어 저장 */
	async function saveFiles(fs: any[]) {
		const l = clone(latest());
		if (isMain) l.Json.files = fs;
		else l.Json.images![tab - 1].files = fs;
		await updateItem(l);
	}
	async function attachFiles(list: FileList | null | undefined) {
		const picked = Array.from(list || []);
		if (fileInput) fileInput.value = '';
		if (!picked.length || attaching) return;
		attaching = true;
		toast(picked.length > 1 ? `파일 ${picked.length}개 올리는 중...` : '파일 올리는 중...');
		const next = [...curFiles];
		let fail = 0;
		for (const f of picked) {
			const o = await secureUpload(f, `work_${item.workId}`, '', 'work');
			const d: any = o.ok ? o.data : null;
			if (!d?.fileName) {
				fail++;
				continue;
			}
			const one = { secureImageId: d.secureImageId, fileName: d.fileName, hostingId: d.hostingId };
			const i = next.findIndex((x) => x.fileName === one.fileName);
			if (i >= 0) next[i] = one;
			else next.push(one);
		}
		if (picked.length > fail) await saveFiles(next);
		attaching = false;
		if (fail) alert(`${fail}개 파일을 올리지 못했습니다.`);
		else toast('첨부했습니다.');
	}
	async function removeFile(f: any) {
		if (!(await confirm(`${f.fileName} 파일을 삭제하시겠습니까?`))) return;
		const o = await secureDelete(`work_${item.workId}`, f.fileName);
		if (!o.ok) return void alert('삭제 도중 오류가 발생하였습니다.');
		await saveFiles(curFiles.filter((x) => x.fileName !== f.fileName));
		toast('정상적으로 삭제되었습니다.');
	}
	async function downloadFile(f: any) {
		const cid = profile?.hostingId != item.hostingId ? item.hostingId || 'OCTOPUS' : profile?.hostingId || 'OCTOPUS';
		const o = await api.post<Blob>('/service/user/file/secure-download', { imageId: `${cid ? `${cid}/` : ''}work_${item.workId}`, fileName: f.fileName }, { blob: true });
		if (!o.ok) return void alert('파일을 받지 못했습니다.');
		const a = document.createElement('a');
		a.href = URL.createObjectURL(o.data as Blob);
		a.download = f.fileName;
		a.click();
	}

	// 메모·마크다운 열기·저장(React ImageCanvasEditor 의 도형 처리 몫)
	const sub = $derived(isMain ? undefined : tab - 1);
	const shapes = $derived((isMain ? item.Json.shapes : cur?.shapes) || []);
	let mdShape = $state<Shape | null>(null);
	let memoShape = $state<Shape | null>(null);
	let mdEd = $state<any>();
	let memoEd = $state<any>();
	const persist = (s: Shape) => updateShape(item.workId, s, sub);
	async function openEditor(s: Shape) {
		if (s.activeEditorName && s.activeEditorName !== profile?.name) return void alert(`${s.activeEditorName}님이 이미 편집 중입니다.`);
		if (mdShape || memoShape) {
			if ((mdEd?.hasChanges?.() || memoEd?.hasChanges?.()) && !(await confirm('수정 중인 내용이 있습니다. 닫고 새로 여시겠습니까?'))) return;
			await mdEd?.forceClose?.();
			await memoEd?.forceClose?.();
		}
		const x = { ...s, activeEditorName: profile?.name, adminId: s.adminId || profile?.adminId };
		if (x.type === 'markdown') mdShape = x;
		else memoShape = x;
		await persist(x);
	}
	async function saveShape(s: Shape) {
		if (s.adminId) s.lastEditAdminId = profile?.adminId;
		await persist({ ...s });
	}
	function addShape(type: 'no-shape' | 'markdown') {
		const n: Shape = { id: Date.now().toString(), type, title: type === 'markdown' ? '새 마크다운' : '새 텍스트', color: 'var(--zc-000000, #000000)', strokeWidth: 1 };
		const l = clone(item);
		if (isMain) l.Json.shapes = [...(l.Json.shapes || []), n];
		else l.Json.images![tab - 1].shapes = [...(l.Json.images![tab - 1].shapes || []), n];
		updateItem(l);
		toast(type === 'markdown' ? '새 마크다운이 추가되었습니다.' : '새 텍스트가 추가되었습니다.');
		setTimeout(() => openEditor(n), 100);
	}
	const imageUrl = $derived(
		isMain
			? item.Json.image
				? `${adminImagePrefix(profile?.hostingId)}/work/work_${item.workId}/${item.Json.image}`
				: ''
			: cur?.name
				? `${adminImagePrefix(profile?.hostingId)}/work/work_${item.workId}/${cur.name}`
				: ''
	);
	const DESC_BOX = 'margin-bottom:8px;border:1px solid var(--gt-400, #d0d0d0);padding:0 16px;border-radius:8px';
</script>

<div class="MuiStack-root" id={`work_id_${item.workId}`} style="display:flex;flex-direction:column">
	<div class="MuiStack-root" style="display:flex;flex-direction:row;justify-content:space-between">
		{#if !setting?.disableSearchForm && !inviteWorkId}
			<CategoryPopover
				category={item.Json.category}
				{codes}
				workId={paramWorkId ? 0 : item.workId}
				{groupKey}
				onselect={changeCategory}
				variant="text"
				btnText="분류변경"
				btnSx={{ color: 'var(--gt-500, #b0b0b0) !important' }}
			/>
		{/if}
		{#if item.Json.targetWorkIds?.length}
			<div class="related">
				<span class="rl"><Icon name="link" size="Small" sx={{ fontSize: 16 }} />연관 업무</span>
				{#each item.Json.targetWorkIds as tid (tid)}
					{@const t = items.find((x) => x.workId === tid)}
					<a class="rc" href={`/admin/work/${t?.groupKey || groupKey}/${tid}`} title={t?.Json?.title || `#${tid}`}>#{tid} {t?.Json?.title || ''}</a>
				{/each}
			</div>
		{/if}
	</div>

	<div class="MuiStack-root mf-head" style="display:flex;flex-direction:row;gap:8px;justify-content:space-between;margin-bottom:8px">
		{#if !inviteWorkId}
			<h3 class="MuiTypography-root MuiTypography-h3 mf-title" style="width:100%">
				<InlineTitle
					placeholder="제목을 입력하세요"
					value={item.Json.title || ''}
					onsave={(v) => (latest().Json.title || '') !== v && updateItem({ ...latest(), Json: { ...latest().Json, title: v } })}
				/>
			</h3>
		{/if}
		<div class="MuiStack-root mf-tools" style="display:{inviteWorkId ? 'none' : 'flex'};flex-direction:row;gap:8px;min-width:fit-content">
			{#if profile?.isDeveloper}<IconButton onclick={() => (jsonItem = item)}><Icon name="data_object" size="Small" /></IconButton>{/if}
			{#if !cols.find((x) => x.key == 'status')}
				<Select
					value={String(item.Json.status ?? '')}
					options={statusItems.map((s) => ({ value: s.key, label: s.value }))}
					onchange={(v) => updateItem({ ...item, Json: { ...item.Json, status: Number(String(v).replace(/[^0-9]/g, '')) || '' } })}
				/>
			{/if}
			{#if (isMain ? !item.Json.desc : !cur?.desc) && !item.Json.chatSummary}
				<Button variant="contained" size="small" onclick={addDesc}>설명 추가</Button>
			{/if}
			{#if !setting?.disableWork && !item.Json.chatSummary}
				<AddWorkButton
					mode={item.Json.image ? 'subImage' : 'work'}
					{item}
					{groupKey}
					category={item.Json.category}
					onsuccess={() => (tab = item.Json.images?.length || 0)}
					variant="contained"
					size="small"
					startIcon="add_a_photo"
					label={item.Json.image ? '' : '이미지 추가'}
				/>
			{/if}
			<Button variant="contained" size="small" color="error" disabled={!writeable || (item.Json.shapes?.length || 0) > 0 || (item.Json.images?.length || 0) > 0} onclick={removeDoc}
				><Icon name="delete" size="Small" sx={{ marginRight: 4 }} />문서 삭제</Button
			>
			{#if item.Json.claudeChat?.sessionId}
				<Button variant="outlined" size="small" href={`/admin/claude/${parseInt(item.Json.claudeChat.bridgeId) || 340}?session=${item.Json.claudeChat.sessionId}`}
					><Icon name="auto_awesome" size="Small" sx={{ marginRight: 8 }} />AI 채팅 바로가기</Button
				>
			{/if}
		</div>
	</div>

	{#if cols.length}
		<div class="MuiStack-root mf-cols" style="display:flex;flex-direction:column;margin-bottom:16px">
			<table class="MuiTable-root adm-cv-table" style="width:fit-content">
				<tbody class="MuiTableBody-root">
					<tr class="MuiTableRow-root">
						{#each cols as col (col.key)}
							<td
								class="MuiTableCell-root MuiTableCell-body MuiTableCell-sizeMedium"
								style="padding:0 16px 0 0;border:0;position:static;background-color:var(--zc-fff, #fff);white-space:{col.key === 'summary' || col.key.startsWith('column_') || col.key === 'status'
									? 'nowrap'
									: 'normal'}"
							>
								<RenderCell {col} {item} shapes={[]} summaryText="" {groupKey} {codes} {codeIds} admins={members} {onValueChange} />
							</td>
						{/each}
					</tr>
				</tbody>
			</table>
		</div>
	{/if}

	<div class="MuiStack-root" style="display:flex;flex-direction:column">
		{#if editorName && profile?.name !== editorName}
			<div class="MuiStack-root" style="display:flex;flex-direction:column">
				<p class="MuiTypography-root MuiTypography-body2" style="color:var(--theme-color-secondary, #9c27b0)">{editorName} 님이 설명을 수정중입니다.</p>
				<div class="MuiStack-root" style="display:flex;flex-direction:column;cursor:pointer;{DESC_BOX}"><HtmlView html={content} {siteMeta} /></div>
			</div>
		{:else if isEdit || editingMe}
			<div class="MuiStack-root" style="display:flex;flex-direction:column">
				<TinyMceEditor bind:this={ed} body={draft || ''} height={400} onchange={(v) => (draft = v)} onImageUI={() => picker?.open()} imageId={`work_${item.workId}`} />
				<div style="margin-top:8px"><Button variant="contained" sx={{ width: 'fit-content' }} onclick={finishEdit}>{content === draft ? '취소' : '편집 완료 (저장)'}</Button></div>
				<div style="display:none">
					<ImagePicker
						bind:this={picker}
						title="본문에 사용할 이미지를 선택하세요"
						size={{ width: 'all', height: 'auto' }}
						folder={`work_${item.workId}`}
						isMultiple
						useOriginalFileName
						maxWidth={1200}
						onSelect={(imgs: any[]) => {
							ed?.insert('<br/>' + imgs.map((i) => `<img style="max-width: 100%;" src="${adminImagePrefix(profile?.hostingId)}/work_${item.workId}/${i?.name}"/><br/>`).join(''));
							picker?.close();
						}}
					/>
				</div>
			</div>
		{:else}
			<div class="MuiStack-root mf-desc" style="display:flex;flex-direction:column;position:relative;{DESC_BOX}">
				{#if content}<HtmlView html={content} {siteMeta} />{:else}<p class="MuiTypography-root MuiTypography-body2" style="color:var(--adm-text2);padding:16px 0">
						설명이 없습니다. 클릭하여 추가하세요.
					</p>{/if}
				<div class="MuiStack-root" style="display:flex;flex-direction:row;justify-content:flex-end;position:absolute;right:8px;top:8px">
					<IconButton disabled={!writeable} onclick={startEdit}><Icon name="edit" size="Small" /></IconButton>
					{#if !setting?.disableLike}
						<IconButton onclick={toggleDescLike}
							><Badge content={likes.length} color="primary"><Icon name="thumb_up" size="Small" color={likes.includes(Number(profile?.adminId)) ? 'Primary' : 'Inherit'} /></Badge></IconButton
						>
					{/if}
				</div>
			</div>
		{/if}
		{#if curFiles.length}
			<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;flex-wrap:wrap;margin-bottom:8px">
				{#each curFiles as f (f.fileName)}
					<span class="att">
						<Button size="small" onclick={() => downloadFile(f)}><Icon name="download" size="Small" sx={{ marginRight: 4 }} />{f.fileName}</Button>
						{#if writeable}<IconButton size="Small" sx={{ padding: 2 }} aria-label="첨부 삭제" onclick={() => removeFile(f)}><Icon name="close" sx={{ fontSize: 16 }} /></IconButton>{/if}
					</span>
				{/each}
			</div>
		{/if}
	</div>

	{#if item.Json.images?.length}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:8px;border-bottom:1px solid var(--adm-divider);margin-bottom:16px">
			<Tabs
				items={[{ value: 0, label: '기본 이미지' }, ...item.Json.images.map((im, i) => ({ value: i + 1, label: im.title || `추가 ${i + 1}`, disabled: isEdit }))]}
				value={tab}
				onchange={(v) => (tab = v)}
			/>
			{#if tab > 0}<IconButton size="Small" sx={{ padding: 0 }} onclick={() => removeImage(tab - 1)} aria-label="이미지 삭제"><Icon name="close" sx={{ fontSize: 16 }} /></IconButton>{/if}
		</div>
	{/if}

	<div class="MuiStack-root" style="display:flex;flex-direction:column;width:100%;position:relative">
		{#if imageUrl}
			<div style="margin-bottom:16px">
				<ImageCanvasEditor {imageUrl} shapes={shapes.filter((s) => s.type !== 'no-shape' && s.type !== 'markdown')} {persist} {openEditor} />
			</div>
		{/if}
		<ShapeList
			{shapes}
			{groupKey}
			workId={item.workId}
			admins={members}
			likeNames={members}
			{openEditor}
			openReply={(s) => {
				const n: Shape = { id: Date.now().toString(), type: 'no-shape', parentShapeId: Number(s.id), color: '#000000', strokeWidth: 3 };
				persist(n);
				setTimeout(() => openEditor(n), 100);
			}}
			deleteShape={async (s) => {
				if (!(await confirm('선택된 도형을 삭제하시겠습니까?'))) return;
				persist({ id: s.id, isDeleted: true } as Shape);
			}}
			toggleLike={(s) => {
				const me = Number(profile?.adminId);
				const ids = s.likeAdminIds || [];
				persist({ ...s, likeAdminIds: ids.includes(me) ? ids.filter((x) => x !== me) : [...ids, me] });
			}}
			updateShape={(s) => persist(s)}
		/>
	</div>

	<div class="MuiStack-root mf-acts" style="display:flex;flex-direction:row;gap:8px;margin-top:{!setting?.disableText || !setting?.disableMarkdown ? 8 : 0}px">
		{#if !isEdit && !editingMe}
			<Button
				variant="outlined"
				disabled={!writeable || (!!editorName && editorName !== profile?.name)}
				sx={{ color: 'var(--theme-color-secondary, #9c27b0)', borderColor: 'var(--theme-color-secondary, #9c27b0)' }}
				onclick={() => ((isMain ? item.Json.desc : cur?.desc) ? startEdit() : addDesc())}>수정</Button
			>
		{/if}
		{#if tab > 0 && !setting?.disableWork && !item.Json.chatSummary}
			<AddWorkButton
				variant="outlined"
				size="small"
				{groupKey}
				category={item.Json.category}
				{item}
				mode="subImage"
				subImageIndex={tab - 1}
				sx={{ color: 'var(--theme-color-secondary, #9c27b0)', borderColor: 'var(--theme-color-secondary, #9c27b0)' }}
			/>
		{/if}
		{#if !setting?.disableFileAttach}
			<Button variant="outlined" disabled={!writeable || attaching} onclick={() => fileInput?.click()}
				><Icon name="attach_file" size="Small" sx={{ marginRight: 4 }} />{attaching ? '올리는 중...' : '파일 첨부'}</Button
			>
			<input bind:this={fileInput} type="file" multiple style="display:none" onchange={(e) => attachFiles((e.currentTarget as HTMLInputElement).files)} />
		{/if}
		{#if !setting?.disableText}<Button variant="outlined" onclick={() => addShape('no-shape')}>텍스트 추가</Button>{/if}
		{#if !setting?.disableMarkdown}<Button variant="outlined" onclick={() => addShape('markdown')}><Icon name="markdown" size="Small" sx={{ marginRight: 4 }} />마크다운 추가</Button>{/if}
	</div>
</div>

<MarkdownShapeEditor
	bind:this={mdEd}
	shape={mdShape}
	{groupKey}
	onsave={saveShape}
	onclose={async (s) => {
		mdShape = null;
		await persist({ ...s, activeEditorName: null });
	}}
/>
<MemoShapeEditor
	bind:this={memoEd}
	shape={memoShape}
	{groupKey}
	workId={item.workId}
	{items}
	onsave={saveShape}
	onclose={async (s) => {
		memoShape = null;
		await persist({ ...s, activeEditorName: null });
	}}
/>
{#if profile?.isDeveloper}<JsonDrawer value={jsonItem} onclose={() => (jsonItem = null)} onsave={(v) => updateItem(v)} />{/if}

<style>
	.related {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 4px 6px;
		min-width: 0;
	}
	.rl {
		display: inline-flex;
		align-items: center;
		gap: 2px;
		font-size: 12px;
		color: var(--gt-700, #909090);
	}
	.rc {
		max-width: 260px;
		padding: 2px 10px;
		border-radius: 999px;
		background: var(--adm-hover, #f5f5f5);
		color: var(--theme-color-primary, #1976d2);
		font-size: 13px;
		text-decoration: none;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.rc:hover {
		text-decoration: underline;
	}
	.att {
		display: inline-flex;
		align-items: center;
		border: 1px solid var(--adm-divider, #e0e0e0);
		border-radius: 6px;
		padding-right: 2px;
	}
	/* 휴대폰(2026-10-05 대표) — 제목 아래로 도구 줄, 칸·버튼은 줄바꿈. 예전엔 한 줄에 다 넣어 제목이 세로로 쪼개지고 버튼이 화면 밖으로 나갔다 */
	.mf-tools :global(.MuiButton-root) {
		white-space: nowrap;
	}
	.mf-acts :global(.MuiButton-root) {
		white-space: nowrap;
	}
	@media (max-width: 899px) {
		.mf-head {
			flex-direction: column !important;
		}
		.mf-title {
			font-size: 1.5rem;
			line-height: 1.35;
		}
		.mf-tools {
			flex-wrap: wrap;
			align-items: center;
			min-width: 0 !important;
		}
		.mf-cols table,
		.mf-cols tbody {
			display: block;
			width: 100% !important;
		}
		.mf-cols tr {
			display: flex;
			flex-wrap: wrap;
			align-items: center;
			gap: 8px 0;
		}
		.mf-cols td {
			display: block;
		}
		.mf-acts {
			flex-wrap: wrap;
		}
		/* 연필·좋아요 버튼(오른쪽 위 고정)이 설명 글자를 덮지 않게 */
		.mf-desc {
			padding-top: 36px !important;
		}
	}
</style>
