<script lang="ts">
	/**
	 * 상태 고르기(React StatusItemSelectPopover) — 색 버튼을 누르면 3칸 격자로 상태들이 뜬다.
	 * 「라벨 수정」(2026-10-04 — 옮길 때 빠져 있었다): codeIds(공통코드 경로 — 표의 상태 칸 [보드, COLUMNS, STATUS] · 열 칸 [보드, COLUMNS, 열])를 받은 곳에서만 보인다.
	 *   편집기는 공통코드 관리와 같은 StatusLabelEditor, 저장은 COMMON_CODES 최신본을 읽어 그 코드의 items 만 바꿔 PUT → 메타를 다시 읽는다.
	 */
	import { api } from '#lib/shared/api.ts';
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import Popover from '../../ui/Popover.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { refreshToken } from '../../profile.svelte.ts';
	import StatusLabelEditor from '../../settings/codes/StatusLabelEditor.svelte';
	import { newStatusLabel } from '../../settings/codes/codes.ts';
	import type { CodeItem } from '#lib/shared/types/work.ts';

	let {
		codes = [],
		value,
		onchange,
		disabled = false,
		noMinWidth = false,
		label = '',
		sx = null,
		codeIds = null
	}: { codes?: CodeItem[]; value: any; onchange: (v: any) => void; disabled?: boolean; noMinWidth?: boolean; label?: string; sx?: Sx | null; codeIds?: number[] | null } = $props();
	const PALETTE = ['#FFD700', '#696969', '#DC143C', '#FF7F50', '#EE82EE', '#2F4F4F', '#D3D3D3', '#FF1493', '#6495ED', '#ADFF2F', '#FFA07A', '#32CD32', '#5D4037'];
	let anchor = $state<HTMLElement | null>(null);
	let editing = $state(false);
	let draft = $state<any[]>([]);
	let saving = $state(false);
	const canEdit = $derived(!!codeIds && codeIds.length >= 2);
	function openPop(el: HTMLElement) {
		draft = JSON.parse(JSON.stringify(codes || []));
		editing = false;
		anchor = el;
	}
	async function apply() {
		if (!canEdit || saving) return;
		saving = true;
		try {
			const cur = await api.get('/service/admin/meta/COMMON_CODES');
			if (!cur.ok || !cur.data) return void alert('공통 코드를 읽지 못했습니다.');
			const all = cur.data;
			// 경로 깊이는 화면마다 다르다 — codeIds 를 따라 내려가 마지막 코드의 items 를 바꾼다
			let target: any = { items: all.items };
			for (const id of codeIds || []) {
				target = (target?.items || []).find((x: any) => x.id === id);
				if (!target) break;
			}
			if (!target) return void alert('코드를 찾지 못했습니다. 목록을 새로고침한 뒤 다시 시도해 주세요.');
			const now = new Date().toISOString();
			target.items = draft.map((c) => ({ ...c, updatedAt: now }));
			const o = await api.put('/service/admin/meta/COMMON_CODES', all);
			if (!o.ok) return void alert('저장 도중 오류가 발생하였습니다.');
			toast('정상적으로 저장되었습니다.');
			editing = false;
			anchor = null;
			await refreshToken(); // 메타 다시 읽기 — 실시간 설정과 상관없이 바로 반영
		} finally {
			saving = false;
		}
	}
	const sel = $derived(codes.find((c) => (c.value || c.id) === value) || codes.find((c) => c.name === value));
	const bgOf = (c: CodeItem, i: number) => c.bgColor || c.color || PALETTE[i % PALETTE.length];
	const selBg = $derived(sel ? bgOf(sel, codes.indexOf(sel)) : 'var(--zc-f5f5f5, #f5f5f5)');
	const selFont = $derived(sel ? sel.color || '#fff' : 'var(--adm-text2)');
</script>

<Button
	variant="contained"
	size="small"
	onclick={(e) => !disabled && openPop(e.currentTarget as HTMLElement)}
	sx={{
		minWidth: noMinWidth ? 0 : 80,
		fontSize: 11,
		fontWeight: 'bold',
		backgroundColor: selBg,
		color: selFont,
		'&:hover': { backgroundColor: selBg, opacity: 0.9 },
		textShadow: sel ? '0px 1px 2px rgba(0,0,0,0.2)' : 'none',
		boxShadow: 'none',
		...(sx || {})
	}}>{label || (sel ? sel.name : '미설정')}</Button
>
<Popover {anchor} onclose={() => (anchor = null)} align="center" paperSx={{ padding: 16, borderRadius: 16, boxShadow: '0px 4px 20px rgba(0,0,0,0.15)', maxWidth: 400, minWidth: 320 }}>
	{#if editing}
		<!-- 편집 UI 는 공통코드 관리와 같은 부품 — 두 곳의 색·순서 규칙이 갈리지 않게 -->
		<StatusLabelEditor value={draft} onChange={(v) => (draft = v)} />
		<div style="margin-top:16px;display:flex;justify-content:center;gap:8px">
			<Button size="small" onclick={() => (editing = false)} sx={{ fontSize: 13 }}>취소</Button>
			<Button variant="outlined" size="small" disabled={saving} onclick={apply} sx={{ fontSize: 13, paddingLeft: 32, paddingRight: 32 }}>적용하기</Button>
		</div>
	{:else}
		<div class="MuiBox-root" style="display:grid;grid-template-columns:repeat(3,1fr);gap:8px">
			{#each codes as c, i (c.id)}
				{@const v = c.value || c.id}
				{@const on = v === value || c.name === value}
				<Button
					size="small"
					variant="contained"
					onclick={() => {
						onchange(v);
						anchor = null;
					}}
					sx={{
						fontSize: 11,
						fontWeight: 'bold',
						backgroundColor: bgOf(c, i),
						color: c.color || '#fff',
						minWidth: 100,
						whiteSpace: 'nowrap',
						'&:hover': { backgroundColor: bgOf(c, i), opacity: 0.8 },
						border: on ? '2px solid #000' : 'none',
						boxShadow: on ? '0px 0px 5px rgba(0,0,0,0.5)' : 'none',
						textShadow: '0px 1px 2px rgba(0,0,0,0.2)'
					}}>{c.name}</Button
				>
			{:else}
				<div class="MuiTypography-root MuiTypography-body2" style="grid-column:span 3;padding:32px 0;text-align:center;color:var(--adm-text2)">
					등록된 상태값이 없습니다.
					{#if canEdit}
						<div style="margin-top:12px">
							<Button
								variant="outlined"
								size="small"
								onclick={() => {
									draft = [...draft, newStatusLabel(draft)];
									editing = true;
								}}><Icon name="add" sx={{ fontSize: 16, marginRight: 4 }} />상태값 등록하기</Button
							>
						</div>
					{/if}
				</div>
			{/each}
		</div>
		{#if canEdit && codes.length}
			<div style="margin-top:16px;display:flex;justify-content:center">
				<Button size="small" onclick={() => (editing = true)} sx={{ fontSize: 13, color: 'var(--adm-text2)' }}><Icon name="edit" sx={{ fontSize: 16, marginRight: 4 }} />라벨 수정</Button>
			</div>
		{/if}
	{/if}
</Popover>
