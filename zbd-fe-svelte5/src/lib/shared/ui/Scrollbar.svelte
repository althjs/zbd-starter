<script lang="ts">
	/**
	 * Scrollbar — React 판 `<Scrollbar>`(simplebar-react 3.3.2) 와 같은 것. 브라우저 기본 스크롤 막대 대신 simplebar 막대.
	 * DOM 은 simplebar-react 가 그리는 그대로(`[data-simplebar] > .simplebar-wrapper …`) — 운영 CSS(thin-scrollbar 등) 가 그 이름을 겨냥한다.
	 * 높이(height·maxHeight) 는 이 부품에 준다 — 넘치면 안쪽(.simplebar-content-wrapper) 이 스크롤된다.
	 *
	 *   scrollEl  실제로 스크롤되는 엘리먼트(scrollTop·scroll 이벤트가 필요할 때 bind)
	 *   fitWidth  폭을 내용에 맞춘다(대화 창처럼 부모가 내용만큼 줄어드는 곳). simplebar 의 스크롤 칸은 절대 위치라
	 *             그대로면 폭이 0 으로 접힌다 → 처음 한 번 내용을 흐름대로 놓아 폭을 재고 그 폭으로 고정한 뒤 막대를 붙인다.
	 *   thumbColor 막대 색(어두운 바탕에서 기본 검정 막대가 묻힐 때)
	 *
	 * React 판과 같게: 막대는 8px·각지게(터치 화면은 legacy.css 의 6px 그대로), 가로로 넘치면 **마우스로 끌어서 좌우 이동**
	 * (5px 넘게 움직여야 끌기 — 그 뒤의 클릭은 막아 줄·링크가 눌리지 않게, 가로로 안 넘치면 끌기 없음 = 글자 선택 그대로).
	 */
	import 'simplebar-core/dist/simplebar.css';
	import SimpleBarCore from 'simplebar-core';
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '../mui/sx';

	let {
		class: cls = '',
		style = '',
		sx = null,
		autoHide = true,
		fitWidth = false,
		thumbColor = '',
		scrollEl = $bindable(),
		children
	}: {
		class?: string;
		style?: string;
		sx?: Sx | null;
		autoHide?: boolean;
		fitWidth?: boolean;
		thumbColor?: string;
		scrollEl?: HTMLDivElement;
		children?: Snippet;
	} = $props();

	let root = $state<HTMLDivElement>();
	let content = $state<HTMLDivElement>();
	let measuring = $state(fitWidth);
	let retry = $state(0);

	// 마우스로 끌어 좌우 이동(React Scrollbar handleMouseDown/Move/Up)
	let dragging = $state(false);
	let moved = $state(false);
	let startX = 0;
	let startLeft = 0;
	function onDown(e: MouseEvent) {
		if (e.button !== 0 || !scrollEl) return;
		moved = false;
		// 막대 자체를 잡은 것은 simplebar 가 끈다 — 같이 움직이면 두 배로 간다
		if ((e.target as Element)?.closest?.('.simplebar-track')) return;
		if (scrollEl.scrollWidth <= scrollEl.clientWidth) return;
		dragging = true;
		startX = e.pageX;
		startLeft = scrollEl.scrollLeft;
	}
	function onMove(e: MouseEvent) {
		if (!dragging || !scrollEl) return;
		const dx = e.pageX - startX;
		if (Math.abs(dx) <= 5) return;
		if (!moved) window.getSelection()?.removeAllRanges();
		moved = true;
		e.preventDefault();
		scrollEl.scrollLeft = startLeft - dx * 1.5;
	}
	function onUp() {
		if (!dragging) return;
		dragging = false;
		if (moved) window.getSelection()?.removeAllRanges();
	}
	function onClickCapture(e: MouseEvent) {
		if (moved) {
			e.stopPropagation();
			e.preventDefault();
			moved = false;
		}
	}

	$effect(() => {
		if (!root || !scrollEl || !content) return;
		if (measuring) {
			void retry;
			const w = root.getBoundingClientRect().width;
			// 아직 화면에 안 붙었으면(숨긴 부모 안에서 그려진 뒤 portal 로 옮겨지는 창 등) 폭이 0 — 0 으로 고정하면 본문이 통째로 사라진다.
			// 다음 프레임에 다시 잰다(최대 30번).
			if (!w && retry < 30) {
				const id = requestAnimationFrame(() => retry++);
				return () => cancelAnimationFrame(id);
			}
			root.style.width = `${Math.ceil(w)}px`;
			measuring = false;
			return;
		}
		const bar = new SimpleBarCore(root, { autoHide, scrollableNode: scrollEl, contentNode: content });
		// 안쪽 내용 크기가 바뀌면(늦게 뜨는 사진 등) 다시 잰다 — simplebar 는 바깥 크기·DOM 변화만 본다
		const ro = new ResizeObserver(() => bar.recalculate());
		for (const el of content.children) ro.observe(el);
		return () => {
			ro.disconnect();
			bar.unMount();
		};
	});
</script>

<!-- svelte-ignore a11y_no_static_element_interactions — 끌어서 넘기기는 마우스 보조 동작(키보드는 안쪽 영역의 화살표 스크롤) -->
<div
	bind:this={root}
	data-simplebar="init"
	class={cx('zbd-sb', cls, measuring && 'zbd-sb-measure', fitWidth && 'zbd-sb-fit', dragging && moved && 'is-dragging', toClass(sx))}
	style="{thumbColor ? `--zbd-sb-thumb:${thumbColor};` : ''}{style}"
	onmousedown={onDown}
	onmousemove={onMove}
	onmouseup={onUp}
	onmouseleave={onUp}
	onclickcapture={onClickCapture}
>
	<div class="simplebar-wrapper">
		<div class="simplebar-height-auto-observer-wrapper"><div class="simplebar-height-auto-observer"></div></div>
		<div class="simplebar-mask">
			<div class="simplebar-offset">
				<!-- svelte-ignore a11y_no_noninteractive_tabindex — simplebar-react 와 같게(키보드로 스크롤) -->
				<div bind:this={scrollEl} class="simplebar-content-wrapper" tabindex="0" role="region" aria-label="scrollable content">
					<div bind:this={content} class="simplebar-content" ondragstart={(e) => e.preventDefault()}>{@render children?.()}</div>
				</div>
			</div>
		</div>
		<div class="simplebar-placeholder"></div>
	</div>
	<div class="simplebar-track simplebar-horizontal"><div class="simplebar-scrollbar"></div></div>
	<div class="simplebar-track simplebar-vertical"><div class="simplebar-scrollbar"></div></div>
</div>

<style>
	/* 폭 재기 — 스크롤 칸을 잠깐 흐름 안에 둔다(fitWidth) */
	.zbd-sb-measure :global(.simplebar-mask),
	.zbd-sb-measure :global(.simplebar-offset) {
		position: static;
	}
	.zbd-sb-fit {
		max-width: 100%;
	}
	/* 막대 — React 판과 같게 기본(칸 11px·둥근 막대)보다 얇고 각지게. 보이는 막대는 ::before(.simplebar-scrollbar 는 잡는 영역).
	   터치 화면은 legacy.css 의 6px 규칙, .thin-scrollbar 는 그쪽(구체성이 더 높다)이 정한다. */
	@media not (pointer: coarse) {
		:global(.zbd-sb > .simplebar-track.simplebar-horizontal) {
			height: 8px;
		}
		:global(.zbd-sb > .simplebar-track.simplebar-vertical) {
			width: 8px;
		}
	}
	:global(.zbd-sb > .simplebar-track .simplebar-scrollbar::before) {
		border-radius: 0;
	}
	/* 막대 색(thumbColor) — 어두운 바탕에서는 평소에도 보이게 */
	:global(.zbd-sb[style*='--zbd-sb-thumb'] > .simplebar-track .simplebar-scrollbar::before) {
		background: var(--zbd-sb-thumb);
		opacity: 0.45;
	}
	:global(.zbd-sb[style*='--zbd-sb-thumb'] > .simplebar-track .simplebar-scrollbar.simplebar-visible::before) {
		opacity: 0.8;
	}
	/* 끄는 중 — 커서 고정·글자 선택 막기 */
	:global(.zbd-sb.is-dragging),
	:global(.zbd-sb.is-dragging *) {
		cursor: grabbing !important;
		user-select: none !important;
	}
</style>
