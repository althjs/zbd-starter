<script lang="ts">
	/**
	 * 이미지 위에 그리기(React ImageCanvasEditor) — 펜·형광펜·화살표·사각형·원·텍스트, 선택해 옮기기·크기 바꾸기·삭제, 이미지+도형을 클립보드로 복사.
	 * 좌표는 이미지 원본 크기(viewBox) 기준이라 화면 크기가 달라도 같은 자리.
	 *
	 *   업무 보드: <ImageCanvasEditor {imageUrl} {shapes} persist={(s) => updateShape(…)} openEditor={…} />
	 *     도형을 바꿀 때마다 persist(도형 하나, 지우면 {id, isDeleted:true}) — 보드 저장소가 화면·서버·소켓을 맞춘다.
	 *     편집(더블클릭·메모 아이콘·오른쪽 메뉴) 은 openEditor 로 넘긴다(메모·마크다운 창은 ManageForm 이 갖고 있다).
	 *   간단 편집 도구(GNB 이미지 아이콘): transient — 서버 저장 없이 onchange(전체 목록), 텍스트는 그 자리에서 바로 입력(여러 줄, 2026-10-04).
	 * 텍스트 상자 = 글자에 맞춘 크기(fitText — 글자 폭을 실제로 잰다). 모서리 핸들을 끌면 글자 크기가 바뀌고 상자는 다시 글자에 맞춘다.
	 * 다른 사람이 편집 중인 도형(activeEditorName)은 고를 수 없다.
	 */
	import { untrack } from 'svelte';
	import ColorInput from '#lib/admin/kit/ColorInput.svelte';
	import Menu from '#lib/shared/mui/Menu.svelte';
	import { cx } from '#lib/shared/mui/sx.ts';
	import { admin } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import Popover from '../../ui/Popover.svelte';
	import TextField from '../../ui/TextField.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import type { Shape } from '#lib/shared/types/work.ts';
	import ShapeRenderer, { TEXT_PAD, textBoldExtra, textLines, textWeight } from './ShapeRenderer.svelte';
	import '../ui/collab.css';
	import { SK } from '#lib/shared/storage-keys.ts';

	type Pt = { x: number; y: number };
	type Tool = 'select' | 'pen' | 'arrow' | 'rect' | 'circle' | 'text' | 'highlighter';
	type Handle = 'start' | 'end';

	let {
		imageUrl,
		shapes: shapesProp = [],
		transient = false,
		onchange,
		persist,
		openEditor
	}: {
		imageUrl: string;
		shapes?: Shape[];
		transient?: boolean;
		onchange?: (shapes: Shape[]) => void;
		persist?: (shape: Shape) => void | Promise<any>;
		openEditor?: (shape: Shape) => void;
	} = $props();

	const TOOLS: { v: Tool; icon: string; t: string }[] = [
		{ v: 'select', icon: 'mouse', t: '선택 수정' },
		{ v: 'pen', icon: 'create', t: '펜' },
		{ v: 'highlighter', icon: 'ink_highlighter', t: '형광펜' },
		{ v: 'arrow', icon: 'near_me', t: '화살표' },
		{ v: 'rect', icon: 'crop_square', t: '사각형' },
		{ v: 'circle', icon: 'radio_button_unchecked', t: '원' },
		{ v: 'text', icon: 'text_fields', t: '텍스트' }
	];
	const WIDTHS = [2, 5, 10, 15, 25];

	const myName = $derived(admin.profile?.name || '');
	let list = $state<Shape[]>([]);
	let tool = $state<Tool>('select');
	let color = $state(localStorage.getItem(SK.canvas.currentColor) || '#ff4444');
	let stroke = $state(Number(localStorage.getItem(SK.canvas.strokeWidth)) || 5);
	let selectedId = $state<string | null>(null);
	let drawing = $state<Shape | null>(null);
	let dragging = false;
	let moved = false; // 누르기만 하고 안 움직였으면 저장하지 않는다(소켓 알림·서버 쓰기 줄이기)
	let dragFrom: Pt | null = null;
	let handle: Handle | null = null;
	let initial: Shape | null = null;
	let dims = $state({ naturalWidth: 0, naturalHeight: 0 });
	let img = $state<HTMLImageElement>();
	let svg = $state<SVGSVGElement>();
	let widthAnchor = $state<HTMLElement | null>(null);
	let copying = $state(false);
	let menuAt = $state<{ x: number; y: number; shape: Shape } | null>(null);
	let menuAnchor = $state<HTMLElement | null>(null);
	let simple = $state<Shape | null>(null);
	let simpleText = $state('');
	/** 그 자리에서 고치는 텍스트 도형 id(transient) */
	let typingId = $state<string | null>(null);
	/** 화면 1px = 이미지 몇 px — 입력칸 위치·핸들 크기 */
	let unit = $state(1);
	$effect(() => {
		const el = svg;
		const nw = dims.naturalWidth;
		if (!el || !nw) return;
		const ro = new ResizeObserver(() => {
			const w = el.getBoundingClientRect().width;
			if (w) unit = nw / w;
		});
		ro.observe(el);
		return () => ro.disconnect();
	});

	// 텍스트 상자 크기 = 실제 글자 폭(굵게) — 예전엔 글자 수 × 0.8 로 짐작해 한글·영문 폭이 어긋났다
	let measureCtx: CanvasRenderingContext2D | null = null;
	function textWidth(lines: string[], font: number, weight = 700) {
		measureCtx ||= document.createElement('canvas').getContext('2d');
		if (!measureCtx) return lines.reduce((m, l) => Math.max(m, l.length), 1) * font;
		measureCtx.font = `${weight} ${font}px ${svg ? getComputedStyle(svg).fontFamily : 'sans-serif'}`;
		return Math.max(font * 0.5, ...lines.map((l) => measureCtx!.measureText(l || ' ').width));
	}
	const boxOf = (s: Shape) => ({ L: Math.min(s.start.x, s.end.x), T: Math.min(s.start.y, s.end.y), R: Math.max(s.start.x, s.end.x), B: Math.max(s.start.y, s.end.y) });
	const lineHOf = (s: Shape) => (s.start && s.end ? Math.max(4, (boxOf(s).B - boxOf(s).T) / textLines(s.title).length) : 40);
	/** 줄 높이 lineH 로 글자에 맞춘 상자 — tl: 왼쪽 위 고정, br: 오른쪽 아래 고정 */
	function fitText(s: Shape, lineH: number, anchor: 'tl' | 'br' = 'tl'): Shape {
		const lines = textLines(s.title);
		const w = textWidth(lines, lineH * 0.8, textWeight(s.strokeWidth)) + lineH * (TEXT_PAD * 2 + 0.8 * textBoldExtra(s.strokeWidth));
		const h = lineH * lines.length;
		const b = boxOf(s);
		return anchor === 'tl' ? { ...s, start: { x: b.L, y: b.T }, end: { x: b.L + w, y: b.T + h } } : { ...s, start: { x: b.R - w, y: b.B - h }, end: { x: b.R, y: b.B } };
	}
	/** 새 텍스트의 기본 줄 높이 — 이미지 폭에 비례(화면에서 20px 안팎) */
	const defaultLineH = () => Math.max(16, Math.round(30 * unit));

	// 바깥 도형 목록 → 화면(그리거나 끄는 중에는 덮지 않는다)
	$effect(() => {
		const next = shapesProp;
		untrack(() => {
			if (!dragging && !drawing) list = (next || []).filter((s) => !s.isDeleted);
		});
	});

	const busyMap = $derived(Object.fromEntries(list.filter((s) => s.activeEditorName).map((s) => [s.id, s.activeEditorName as string])));
	// 남이 편집을 시작한 도형은 선택 해제
	$effect(() => {
		if (selectedId && busyMap[selectedId] && busyMap[selectedId] !== myName) selectedId = null;
	});

	const clone = <T,>(v: T): T => JSON.parse(JSON.stringify(v));
	function commit(next: Shape[]) {
		list = next;
		if (transient) onchange?.(next);
	}
	function save(s: Shape) {
		if (!transient) persist?.(s);
	}
	function patchSelected(p: Partial<Shape>) {
		if (!selectedId) return;
		let changed: Shape | null = null;
		// 텍스트는 굵기(선 두께)가 바뀌면 글자 폭도 바뀌니 상자를 다시 맞춘다
		commit(list.map((s) => (s.id === selectedId ? (changed = s.type === 'text' && p.strokeWidth ? fitText({ ...s, ...p }, lineHOf(s)) : { ...s, ...p }) : s)));
		if (changed) save(changed);
	}
	function setColor(c: string) {
		color = c;
		localStorage.setItem(SK.canvas.currentColor, c);
		patchSelected({ color: c });
	}
	function setStroke(w: number) {
		stroke = w;
		localStorage.setItem(SK.canvas.strokeWidth, String(w));
		patchSelected({ strokeWidth: w });
	}

	function onImageLoad() {
		if (img) dims = { naturalWidth: img.naturalWidth, naturalHeight: img.naturalHeight };
	}
	function coords(e: MouseEvent): Pt {
		if (!svg || !dims.naturalWidth) return { x: 0, y: 0 };
		const r = svg.getBoundingClientRect();
		return { x: ((e.clientX - r.left) / r.width) * dims.naturalWidth, y: ((e.clientY - r.top) / r.height) * dims.naturalHeight };
	}
	const shapeIdOf = (e: Event) => (e.target as Element)?.getAttribute?.('data-shape-id') || null;
	/** 두 점이 min 보다 가까우면 min 만큼 띄운다(너무 작은 도형 방지) */
	const keep = (p: number, base: number, min: number) => (Math.abs(p - base) < min ? (p > base ? base + min : base - min) : p);

	function startHandle(s: Shape, h: Handle) {
		handle = h;
		dragging = true;
		initial = clone(s);
	}

	function down(e: MouseEvent) {
		if (e.button !== 0) return;
		if (typingId) return; // 입력 중 바깥 누르기 = 입력 끝(blur) 만 — 새 텍스트를 또 만들지 않는다
		const pos = coords(e);
		const id = shapeIdOf(e);
		if (handle) return;
		if (id) {
			if (busyMap[id]) return;
			selectedId = id;
			tool = 'select';
			dragging = true;
			dragFrom = pos;
			return;
		}
		if (tool === 'select') {
			selectedId = null;
			return;
		}
		selectedId = null;
		const hl = tool === 'highlighter';
		const n: Shape = { id: Date.now().toString(), type: hl ? 'pen' : tool, color: hl ? `${color}80` : color, strokeWidth: hl ? 25 : stroke };
		if (tool === 'pen' || hl) n.points = [pos, pos];
		else {
			n.start = pos;
			n.end = pos;
		}
		drawing = n;
	}

	function move(e: MouseEvent) {
		const pos = coords(e);
		if (dragging && selectedId) {
			moved = true;
			list = list.map((s) => {
				if (s.id !== selectedId) return s;
				if (handle) {
					if (s.type === 'pen' && initial?.points) {
						const ixs = initial.points.map((p: Pt) => p.x);
						const iys = initial.points.map((p: Pt) => p.y);
						const [mnX, mnY, mxX, mxY] = [Math.min(...ixs), Math.min(...iys), Math.max(...ixs), Math.max(...iys)];
						const fixed = handle === 'start' ? { x: mxX, y: mxY } : { x: mnX, y: mnY };
						const moving = handle === 'start' ? { x: mnX, y: mnY } : { x: mxX, y: mxY };
						const sx = Math.max(20, Math.abs(pos.x - fixed.x)) / Math.max(1, Math.abs(moving.x - fixed.x));
						const sy = Math.max(20, Math.abs(pos.y - fixed.y)) / Math.max(1, Math.abs(moving.y - fixed.y));
						return { ...s, points: initial.points.map((p: Pt) => ({ x: fixed.x + (p.x - fixed.x) * sx, y: fixed.y + (p.y - fixed.y) * sy })) };
					}
					// 텍스트: 반대 모서리에서 끈 곳까지 거리 ÷ 상자 대각선 비율로 글자 크기를 바꾸고, 상자는 글자에 맞춘다(반대 모서리 고정)
					if (s.type === 'text' && initial?.start && initial.end) {
						const b = boxOf(initial);
						const fixed = handle === 'end' ? { x: b.L, y: b.T } : { x: b.R, y: b.B };
						const k = Math.hypot(pos.x - fixed.x, pos.y - fixed.y) / Math.max(1, Math.hypot(b.R - b.L, b.B - b.T));
						return fitText({ ...s, start: initial.start, end: initial.end }, Math.max(8 * unit, lineHOf(initial) * k), handle === 'end' ? 'tl' : 'br');
					}
					if (handle === 'start' && s.end) return { ...s, start: { x: keep(pos.x, s.end.x, 20), y: keep(pos.y, s.end.y, 20) } };
					if (handle === 'end' && s.start) return { ...s, end: { x: keep(pos.x, s.start.x, 20), y: keep(pos.y, s.start.y, 20) } };
					return s;
				}
				if (dragFrom) {
					const dx = pos.x - dragFrom.x;
					const dy = pos.y - dragFrom.y;
					if (s.type === 'pen' && s.points) return { ...s, points: s.points.map((p: Pt) => ({ x: p.x + dx, y: p.y + dy })) };
					if (s.start && s.end) return { ...s, start: { x: s.start.x + dx, y: s.start.y + dy }, end: { x: s.end.x + dx, y: s.end.y + dy } };
				}
				return s;
			});
			if (!handle) dragFrom = pos;
			return;
		}
		if (!drawing) return;
		if (drawing.type === 'pen') drawing = { ...drawing, points: [...(drawing.points || []), pos] };
		else {
			const min = drawing.type === 'text' ? 0 : 20;
			drawing = { ...drawing, end: { x: keep(pos.x, drawing.start.x, min), y: keep(pos.y, drawing.start.y, min) } };
		}
	}

	function up() {
		if (drawing) {
			let n = drawing;
			if (n.type === 'text') {
				// 누르기만 하면 기본 크기, 끌었으면 끈 높이가 글자 높이(한 줄)
				const dragH = Math.abs(n.end.y - n.start.y);
				const lineH = dragH > 12 * unit ? dragH : defaultLineH();
				const at = { x: Math.min(n.start.x, n.end.x), y: dragH > 12 * unit ? Math.min(n.start.y, n.end.y) : n.start.y - lineH / 2 };
				n = fitText({ ...n, title: transient ? '' : '제목이 텍스트', start: at, end: at }, lineH);
			}
			drawing = null;
			commit([...list, n]);
			selectedId = n.id;
			save(n);
			if (n.type === 'text') {
				const created = n;
				setTimeout(() => edit(created), transient ? 0 : 100);
			}
		}
		if (dragging && moved && selectedId) {
			const s = list.find((x) => x.id === selectedId);
			if (s) {
				if (transient) onchange?.(list);
				save(s);
			}
		}
		dragging = false;
		moved = false;
		dragFrom = null;
		handle = null;
		initial = null;
	}

	function dblclick(e: MouseEvent) {
		const id = shapeIdOf(e);
		const s = id && list.find((x) => x.id === id);
		if (s) edit(s);
	}
	function contextmenu(e: MouseEvent) {
		e.preventDefault();
		const id = shapeIdOf(e);
		const s = id && list.find((x) => x.id === id);
		if (!s) return;
		selectedId = s.id;
		menuAt = { x: e.clientX, y: e.clientY, shape: s };
	}

	async function remove(id: string) {
		if (!(await confirm('선택된 도형을 삭제하시겠습니까?'))) return;
		commit(list.filter((s) => s.id !== id));
		save({ id, isDeleted: true } as Shape);
		selectedId = null;
	}

	function edit(s: Shape) {
		if (transient) {
			if (s.type === 'text') {
				selectedId = s.id;
				typingId = s.id;
				return;
			}
			if (s.type === 'markdown') {
				simple = s;
				simpleText = s.title || '';
			}
			return;
		}
		if (busyMap[s.id] && busyMap[s.id] !== myName) return void alert(`${busyMap[s.id]}님이 이미 편집 중입니다.`);
		selectedId = s.id;
		openEditor?.(s);
	}
	function saveSimple() {
		if (!simple) return;
		const s: Shape = { ...simple, title: simpleText.trim() };
		commit(list.map((x) => (x.id === s.id ? s : x)));
		simple = null;
		simpleText = '';
	}

	// ── 그 자리 텍스트 입력 ── Enter = 줄바꿈, Esc·Ctrl/⌘+Enter·바깥 누르기 = 끝. 비우고 끝내면 그 텍스트를 지운다.
	const typing = $derived(typingId ? list.find((x) => x.id === typingId) || null : null);
	function typeInput(value: string) {
		if (!typing) return;
		const t = typing;
		list = list.map((x) => (x.id === t.id ? fitText({ ...x, title: value }, lineHOf(x)) : x));
	}
	function finishTyping() {
		const t = typing;
		typingId = null;
		if (!t) return;
		const title = (t.title || '').replace(/^\n+|\s+$/g, '');
		if (!title.trim()) {
			commit(list.filter((x) => x.id !== t.id));
			selectedId = null;
			return;
		}
		commit(list.map((x) => (x.id === t.id ? fitText({ ...x, title }, lineHOf(x)) : x)));
	}
	function focusEnd(el: HTMLTextAreaElement) {
		requestAnimationFrame(() => {
			el.focus();
			el.setSelectionRange(el.value.length, el.value.length);
		});
	}

	// Delete·Backspace = 선택 도형 지우기(입력칸·편집 창이 아닐 때)
	function keydown(e: KeyboardEvent) {
		const t = e.target as HTMLElement;
		if (t?.tagName === 'INPUT' || t?.tagName === 'TEXTAREA' || t?.isContentEditable || simple) return;
		if ((e.key === 'Delete' || e.key === 'Backspace') && selectedId) remove(selectedId);
	}

	async function copyImage() {
		if (copying || !img || !svg) return;
		const { naturalWidth: w, naturalHeight: h } = dims;
		if (!w || !h) return void alert('이미지 정보를 가져올 수 없습니다.');
		const prev = selectedId;
		selectedId = null; // 핸들이 찍히지 않게
		copying = true;
		try {
			await new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(r)));
			const canvas = document.createElement('canvas');
			canvas.width = w;
			canvas.height = h;
			const ctx = canvas.getContext('2d');
			if (!ctx) throw new Error('canvas context 생성 실패');
			const local = imageUrl.startsWith('blob:') || imageUrl.startsWith('data:');
			const base = new Image();
			if (!local) base.crossOrigin = 'anonymous';
			await new Promise<void>((ok, ng) => {
				base.onload = () => ok();
				base.onerror = () => ng(new Error('이미지 로드 실패'));
				base.src = imageUrl; // 같은 출처(/upload/…) 그림 — 다른 출처 그림은 CORS 허용이 있어야 저장된다
			});
			ctx.drawImage(base, 0, 0, w, h);
			const c = svg.cloneNode(true) as SVGSVGElement;
			c.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
			c.setAttribute('width', String(w));
			c.setAttribute('height', String(h));
			c.removeAttribute('style');
			c.querySelectorAll('[data-memo-icon]').forEach((el) => el.remove());
			const url = URL.createObjectURL(new Blob([new XMLSerializer().serializeToString(c)], { type: 'image/svg+xml;charset=utf-8' }));
			try {
				const si = new Image();
				await new Promise<void>((ok, ng) => {
					si.onload = () => ok();
					si.onerror = () => ng(new Error('SVG 변환 실패'));
					si.src = url;
				});
				ctx.drawImage(si, 0, 0, w, h);
			} finally {
				URL.revokeObjectURL(url);
			}
			const blob = await new Promise<Blob>((ok, ng) => canvas.toBlob((b) => (b ? ok(b) : ng(new Error('canvas → blob 변환 실패 (CORS taint 가능성)'))), 'image/png'));
			await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]);
			toast('클립보드로 복사되었습니다.');
		} catch (e: any) {
			console.error('[copyToClipboard]', e);
			alert(`복사 실패: ${e?.message || e}`);
		} finally {
			copying = false;
			if (prev) selectedId = prev;
		}
	}
</script>

<svelte:window onkeydown={keydown} />

{#if imageUrl}
	<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:16px;width:100%;position:relative">
		<div class="MuiStack-root" style="display:flex;flex-direction:row;flex-wrap:wrap;gap:8px;align-items:center;padding:8px;background-color:var(--zc-f5f5f5, #f5f5f5);border-radius:4px">
			<div class="MuiToggleButtonGroup-root adm-tbg" role="group" aria-label="drawing tools">
				{#each TOOLS as o (o.v)}
					<Tooltip title={o.t} placement="top">
						<button
							type="button"
							class={cx('MuiButtonBase-root MuiToggleButton-root MuiToggleButton-sizeSmall MuiToggleButtonGroup-grouped', tool === o.v && 'Mui-selected')}
							style="padding:7px"
							aria-label={o.t}
							aria-pressed={tool === o.v}
							onclick={() => {
								tool = o.v;
								if (o.v !== 'select') selectedId = null;
							}}><Icon name={o.icon} /></button
						>
					</Tooltip>
				{/each}
			</div>
			<ColorInput value={color} onChange={setColor} size="small" sx={{ width: 120 }} />
			<IconButton sx={{ border: '1px solid #ccc', borderRadius: 4, padding: 7 }} aria-label="선 두께" onclick={(e: MouseEvent) => (widthAnchor = e.currentTarget as HTMLElement)}>
				<svg width="20" height="20" viewBox="0 0 24 24"><line x1="4" y1="20" x2="20" y2="4" stroke={color} stroke-width={Math.max(2, stroke / 2)} stroke-linecap="round" /></svg>
			</IconButton>
			<div style="flex-grow:1"></div>
			<Tooltip title="이미지+도형을 클립보드로 복사" placement="top">
				<Button variant="outlined" size="small" onclick={copyImage} disabled={copying || !dims.naturalWidth}>
					{#snippet startIcon()}<Icon name="content_copy" size="Small" />{/snippet}
					{copying ? '복사 중...' : '이미지 복사'}
				</Button>
			</Tooltip>
			{#if selectedId}
				<Button variant="outlined" color="warning" size="small" onclick={() => selectedId && remove(selectedId)}>
					{#snippet startIcon()}<Icon name="delete" size="Small" />{/snippet}
					선택 삭제
				</Button>
			{/if}
		</div>

		<div style="position:relative;display:inline-block;line-height:0;width:fit-content">
			<img bind:this={img} src={imageUrl} alt="" onload={onImageLoad} draggable="false" style="max-width:100%;height:auto;display:block;user-select:none;border:1px solid var(--gt-300, #e0e0e0)" />
			<svg
				bind:this={svg}
				role="application"
				onmousedown={down}
				onmousemove={move}
				onmouseup={up}
				onmouseleave={up}
				ondblclick={dblclick}
				oncontextmenu={contextmenu}
				style="position:absolute;top:0;left:0;width:100%;height:100%;touch-action:none;cursor:{tool === 'select' ? 'default' : 'crosshair'}"
				viewBox={`0 0 ${dims.naturalWidth || 100} ${dims.naturalHeight || 100}`}
			>
				{#each list as s (s.id)}
					<ShapeRenderer
						shape={s}
						selected={s.id === selectedId}
						editingByMe={!!myName && s.activeEditorName === myName}
						hideText={s.id === typingId}
						{unit}
						onhandle={(h) => startHandle(s, h)}
						onmemo={() => !busyMap[s.id] && edit(s)}
					/>
				{/each}
				{#if drawing}<ShapeRenderer shape={drawing} {unit} />{/if}
			</svg>
			{#if typing?.start && typing.end}
				{@const b = boxOf(typing)}
				{@const lh = lineHOf(typing)}
				<textarea
					class="ice-text"
					use:focusEnd
					value={typing.title || ''}
					placeholder="텍스트 입력"
					rows={textLines(typing.title).length}
					style="font-weight:{textWeight(typing.strokeWidth)};left:{b.L / unit}px;top:{b.T / unit}px;font-size:{(lh * 0.8) / unit}px;line-height:{lh / unit}px;padding:0 {(lh * TEXT_PAD) /
						unit}px;color:{typing.color};outline-color:{typing.color}"
					oninput={(e) => typeInput((e.currentTarget as HTMLTextAreaElement).value)}
					onblur={finishTyping}
					onkeydown={(e) => {
						if (e.isComposing) return;
						if (e.key === 'Escape' || (e.key === 'Enter' && (e.metaKey || e.ctrlKey))) {
							e.preventDefault();
							e.stopPropagation();
							(e.currentTarget as HTMLTextAreaElement).blur();
						}
					}}></textarea>
			{/if}
		</div>

		{#if !transient && dims.naturalWidth > 0}
			<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2);padding:0 8px">
				{selectedId ? '도형이 선택되었습니다. 드래그하여 이동하거나 핸들을 잡아 크기를 조절하세요. (Del 키로 삭제)' : '도구를 선택하여 그리거나 이미 그려진 도형을 선택하세요.'}
			</span>
		{/if}
	</div>
{/if}

<Popover anchor={widthAnchor} onclose={() => (widthAnchor = null)}>
	<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:8px;padding:8px">
		{#each WIDTHS as w (w)}
			<IconButton
				sx={{ border: stroke === w ? `2px solid ${color}` : '1px solid #eee', padding: 8, borderRadius: 4 }}
				aria-label={`두께 ${w}`}
				onclick={() => {
					setStroke(w);
					widthAnchor = null;
				}}><svg width="24" height="24" viewBox="0 0 24 24"><line x1="4" y1="20" x2="20" y2="4" stroke={color} stroke-width={Math.max(2, w / 2)} stroke-linecap="round" /></svg></IconButton
			>
		{/each}
	</div>
</Popover>

<!-- 오른쪽 클릭 메뉴 — 마우스 자리에 둔 보이지 않는 점을 기준으로 -->
{#if menuAt}<div bind:this={menuAnchor} style="position:fixed;left:{menuAt.x}px;top:{menuAt.y}px;width:0;height:0"></div>{/if}
<Menu anchor={menuAnchor} open={!!menuAt && !!menuAnchor} onclose={() => (menuAt = null)}>
	{@const s = menuAt?.shape}
	{@const busy = !!(s && busyMap[s.id])}
	<li
		class={cx('MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters', busy && 'Mui-disabled')}
		role="menuitem"
		tabindex="0"
		onclick={() => (s && !busy && edit(s), (menuAt = null))}
		onkeydown={() => {}}
	>
		편집
	</li>
	<li
		class={cx('MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters', busy && 'Mui-disabled')}
		role="menuitem"
		tabindex="0"
		onclick={() => (s && !busy && remove(s.id), (menuAt = null))}
		onkeydown={() => {}}
	>
		삭제
	</li>
</Menu>

{#if transient}
	<Dialog open={!!simple} onclose={saveSimple} title="텍스트 입력" maxWidth="sm" fullWidth zIndex={1500}>
		<TextField
			bind:value={simpleText}
			autofocus
			fullWidth
			multiline
			placeholder="텍스트를 입력하세요"
			sx={{ marginTop: 8 }}
			onkeydown={(e: KeyboardEvent) => {
				if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
					e.preventDefault();
					saveSimple();
				}
			}}
		/>
	</Dialog>
{/if}

<style>
	.ice-text {
		position: absolute;
		margin: 0;
		border: 0;
		outline: 1.5px dashed;
		outline-offset: 0;
		background: rgba(255, 255, 255, 0.85);
		font-family: inherit;
		white-space: pre;
		overflow: hidden;
		resize: none;
		min-width: 2em;
		field-sizing: content;
		box-sizing: content-box;
		z-index: 2;
	}
	.ice-text::placeholder {
		color: currentColor;
		opacity: 0.4;
	}
</style>
