/**
 * `use:portal` — 엘리먼트를 body 끝으로 옮겨 그린다(예전 MUI Modal 의 portal).
 *
 * 팝업·전체 화면 보기는 반드시 이걸 쓴다. 화면 안에 그대로 두면 부모의 글꼴(미니홈피 webfonts 의 `*` 규칙)·
 * overflow·z-index 를 물려받아 React 판과 모양이 달라진다(2026-10-03 공유 창 「취소」 글꼴이 바뀌었던 원인).
 */
export function portal(node: HTMLElement, target: HTMLElement | string = 'body') {
	const host = typeof target == 'string' ? document.querySelector(target) : target;
	host?.appendChild(node);
	return {
		destroy() {
			node.remove();
		}
	};
}
