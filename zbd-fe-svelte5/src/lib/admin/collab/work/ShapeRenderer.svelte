<script lang="ts" module>
	/** 텍스트 상자 왼쪽 여백 = 줄 높이 × TEXT_PAD (편집기의 너비 맞추기와 같은 값) */
	export const TEXT_PAD = 0.15;
	export const textLines = (t?: string) => (t || '텍스트를 입력하세요').split('\n');
	/** 텍스트 굵기 = 선 두께 칸(2·5·10·15·25) — 기본 5 가 예전과 같은 bold(700) */
	export const textWeight = (sw?: number) => (!sw ? 700 : sw >= 15 ? 900 : sw >= 10 ? 800 : sw >= 5 ? 700 : 400);
	/** 글꼴 굵기 단계만으론 700~900 차이가 작다 → 10·15·25 는 같은 색 테두리를 글자 크기 비율만큼 더 얹는다 */
	export const textBoldExtra = (sw?: number) => (!sw ? 0 : sw >= 25 ? 0.05 : sw >= 15 ? 0.032 : sw >= 10 ? 0.018 : 0);
</script>

<script lang="ts">
	/**
	 * 도형 하나(React ShapeRenderer) — 펜·화살표·사각형·원·텍스트를 SVG 로. 이미지 원본 크기 좌표(viewBox) 기준.
	 * 누르기 판정은 `data-shape-id` 가 붙은 엘리먼트로 한다(편집기가 target 에서 읽는다). 선택되면 양 끝 핸들 2개.
	 * 📝·📚 메모 아이콘(`data-memo-icon`) 은 이미지 복사 때 빠진다.
	 * 텍스트는 여러 줄(title 의 줄바꿈) — 줄 높이 = 상자 높이 / 줄 수, 글자 = 줄 높이 × 0.8, 왼쪽 정렬.
	 * unit = 화면 1px 이 이미지 몇 px 인지 — 핸들·선택 점선을 화면에서 늘 같은 크기로(큰 캡처를 줄여 보면 핸들이 점처럼 작았다).
	 */
	import type { Shape } from '#lib/shared/types/work.ts';

	let {
		shape,
		selected = false,
		editingByMe = false,
		hideText = false,
		unit = 1,
		onhandle,
		onmemo
	}: {
		shape: Shape;
		selected?: boolean;
		editingByMe?: boolean;
		hideText?: boolean;
		unit?: number;
		onhandle?: (h: 'start' | 'end') => void;
		onmemo?: () => void;
	} = $props();

	const cursor = $derived(selected ? 'move' : 'pointer');
	const dash = $derived(editingByMe ? '5,5' : 'none');
	const pts = $derived(shape.type === 'pen' && shape.points ? (shape.points as { x: number; y: number }[]) : null);
	/** 핸들·아이콘 자리 — 펜은 점들의 바깥 상자 */
	const box = $derived.by(() => {
		if (pts) {
			const xs = pts.map((p) => p.x);
			const ys = pts.map((p) => p.y);
			return { s: { x: Math.min(...xs), y: Math.min(...ys) }, e: { x: Math.max(...xs), y: Math.max(...ys) } };
		}
		if (!shape.start || !shape.end) return null;
		return { s: shape.start, e: shape.end };
	});
	const rect = $derived(box ? { x: Math.min(box.s.x, box.e.x), y: Math.min(box.s.y, box.e.y), w: Math.abs(box.s.x - box.e.x), h: Math.abs(box.s.y - box.e.y) } : null);
	const iconAt = $derived(rect ? { x: rect.x + rect.w, y: rect.y } : null);
	const targetId = $derived(shape.targetWorkIds?.[0] ?? shape.targetWorkId); // 여러 개면 첫 업무로
	const hasTarget = $derived(!!targetId);
	const hasContent = $derived(!!(shape.title || shape.html || shape.activeEditorName));

	const pathData = $derived(pts ? pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ') : '');
	const arrow = $derived.by(() => {
		if (shape.type !== 'arrow' || !shape.start || !shape.end) return null;
		const { start: s, end: e } = shape;
		const angle = Math.atan2(e.y - s.y, e.x - s.x);
		const head = Math.max(20, shape.strokeWidth * 3);
		const ha = Math.PI / 6;
		const off = head * Math.cos(ha) * 0.9;
		return {
			lineEnd: { x: e.x - off * Math.cos(angle), y: e.y - off * Math.sin(angle) },
			p1: { x: e.x - head * Math.cos(angle - ha), y: e.y - head * Math.sin(angle - ha) },
			p2: { x: e.x - head * Math.cos(angle + ha), y: e.y - head * Math.sin(angle + ha) }
		};
	});
	const textH = $derived(rect ? Math.max(1, rect.h) : 1);
	const lines = $derived(textLines(shape.title));
	const lineH = $derived(textH / lines.length);
	const fontSize = $derived(lineH * 0.8);
	const boldExtra = $derived(fontSize * textBoldExtra(shape.strokeWidth));

	function iconDown(e: MouseEvent) {
		e.stopPropagation();
		if (hasTarget) {
			const el = document.getElementById(`work_id_${targetId}`);
			if (el) return void window.scrollTo({ top: el.getBoundingClientRect().top + window.pageYOffset - 110, behavior: 'smooth' });
		}
		onmemo?.();
	}
	function handleDown(e: MouseEvent, h: 'start' | 'end') {
		e.stopPropagation();
		onhandle?.(h);
	}
</script>

{#snippet handles()}
	{#if selected && box}
		<circle cx={box.s.x} cy={box.s.y} r={7 * unit} fill="white" stroke="#ff4444" stroke-width={1.5 * unit} style="cursor:nwse-resize" role="presentation" onmousedown={(e) => handleDown(e, 'start')} />
		<circle cx={box.e.x} cy={box.e.y} r={7 * unit} fill="white" stroke="#ff4444" stroke-width={1.5 * unit} style="cursor:nwse-resize" role="presentation" onmousedown={(e) => handleDown(e, 'end')} />
	{/if}
{/snippet}
{#snippet memo()}
	{#if iconAt && (hasTarget || hasContent)}
		<g data-memo-icon="1" transform={`translate(${iconAt.x}, ${iconAt.y})`} style="cursor:pointer">
			{#if shape.title}<title>{shape.title}</title>{/if}
			<rect
				x={-10}
				y={-24}
				width={shape.activeEditorName ? 63 : 28}
				height={28}
				fill="white"
				rx={shape.activeEditorName ? 4 : 40}
				stroke="#ddd"
				stroke-width={1}
				role="presentation"
				onmousedown={iconDown}
			/>
			<text x={shape.activeEditorName ? 20 : 4} y={-10} font-size="14" text-anchor="middle" dominant-baseline="central" style="pointer-events:none;user-select:none"
				>{hasTarget ? '📚' : shape.isQuestion ? '❓' : '📝'} {shape.activeEditorName || ''}</text
			>
		</g>
	{/if}
{/snippet}

{#if pts}
	<g id={`shape_id_${shape.id}`}>
		<path d={pathData} stroke={shape.color} stroke-width={shape.strokeWidth} stroke-dasharray={dash} fill="none" data-shape-id={shape.id} style:cursor stroke-linecap="round" stroke-linejoin="round" />
		{#if selected}<path d={pathData} stroke="transparent" stroke-width={shape.strokeWidth + 20} fill="none" data-shape-id={shape.id} style:cursor />{/if}
		{@render handles()}
		{@render memo()}
	</g>
{:else if arrow && shape.start && shape.end}
	<g id={`shape_id_${shape.id}`}>
		<line x1={shape.start.x} y1={shape.start.y} x2={shape.end.x} y2={shape.end.y} stroke="transparent" stroke-width={shape.strokeWidth + 20} data-shape-id={shape.id} style:cursor />
		<line
			x1={shape.start.x}
			y1={shape.start.y}
			x2={arrow.lineEnd.x}
			y2={arrow.lineEnd.y}
			stroke={shape.color}
			stroke-width={shape.strokeWidth}
			stroke-dasharray={dash}
			stroke-linecap="round"
			data-shape-id={shape.id}
			style:cursor
		/>
		<polygon points={`${shape.end.x},${shape.end.y} ${arrow.p1.x},${arrow.p1.y} ${arrow.p2.x},${arrow.p2.y}`} fill={shape.color} />
		{@render handles()}
		{@render memo()}
	</g>
{:else if shape.type === 'rect' && rect}
	<g id={`shape_id_${shape.id}`}>
		<rect x={rect.x} y={rect.y} width={rect.w} height={rect.h} fill="transparent" data-shape-id={shape.id} style:cursor />
		<rect x={rect.x} y={rect.y} width={rect.w} height={rect.h} stroke={shape.color} stroke-width={shape.strokeWidth} stroke-dasharray={dash} fill="none" pointer-events="none" />
		{@render handles()}
		{@render memo()}
	</g>
{:else if shape.type === 'circle' && rect}
	<g id={`shape_id_${shape.id}`}>
		<ellipse cx={rect.x + rect.w / 2} cy={rect.y + rect.h / 2} rx={rect.w / 2} ry={rect.h / 2} fill="transparent" data-shape-id={shape.id} style:cursor />
		<ellipse
			cx={rect.x + rect.w / 2}
			cy={rect.y + rect.h / 2}
			rx={rect.w / 2}
			ry={rect.h / 2}
			stroke={shape.color}
			stroke-width={shape.strokeWidth}
			stroke-dasharray={dash}
			fill="none"
			pointer-events="none"
		/>
		<ellipse
			cx={rect.x + rect.w / 2}
			cy={rect.y + rect.h / 2}
			rx={rect.w / 2}
			ry={rect.h / 2}
			stroke="transparent"
			stroke-width={shape.strokeWidth + 20}
			fill="none"
			data-shape-id={shape.id}
			style:cursor
		/>
		{@render handles()}
		{@render memo()}
	</g>
{:else if shape.type === 'text' && rect}
	<g id={`shape_id_${shape.id}`}>
		<rect x={rect.x} y={rect.y} width={Math.max(1, rect.w)} height={textH} fill="transparent" data-shape-id={shape.id} style:cursor />
		{#if !hideText}
			<text
				font-size={fontSize}
				fill={shape.color}
				font-weight={textWeight(shape.strokeWidth)}
				style="pointer-events:none;user-select:none;white-space:pre;paint-order:stroke;stroke:white;stroke-width:{Math.max(1, fontSize / 12) + boldExtra * 2};stroke-linejoin:round"
				>{#each lines as line, i (i)}<tspan x={rect.x + lineH * TEXT_PAD} y={rect.y + lineH * i + lineH * 0.8}>{line || ' '}</tspan>{/each}</text
			>
			{#if boldExtra}<text
					font-size={fontSize}
					fill={shape.color}
					font-weight={textWeight(shape.strokeWidth)}
					style="pointer-events:none;user-select:none;white-space:pre;stroke:{shape.color};stroke-width:{boldExtra};stroke-linejoin:round"
					>{#each lines as line, i (i)}<tspan x={rect.x + lineH * TEXT_PAD} y={rect.y + lineH * i + lineH * 0.8}>{line || ' '}</tspan>{/each}</text
				>{/if}
			{#if selected}<rect
					x={rect.x}
					y={rect.y}
					width={Math.max(1, rect.w)}
					height={textH}
					stroke={shape.color}
					stroke-width={unit}
					stroke-dasharray={`${5 * unit},${5 * unit}`}
					fill="none"
					pointer-events="none"
				/>{/if}
			{@render handles()}
		{/if}
		{@render memo()}
	</g>
{/if}
