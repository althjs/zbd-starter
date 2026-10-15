/**
 * 휴대폰 뒤로가기 = 덮은 창만 닫기 (React hooks/useBackToClose.ts 를 그대로 옮김, 2026-09-30 규칙).
 *
 * - `installAutoBackToClose()` — 관리자 레이아웃이 한 번 부른다. 휴대폰(≤899px) 에서 화면을 덮는 창
 *   (body 바로 아래 `.MuiModal-root` 중 전체 화면 Dialog(`.MuiDialog-paperFullScreen`) · 서랍(`.MuiDrawer-modal`))이 열리면
 *   히스토리를 하나 쌓고, 뒤로가기면 그 창에 Esc 를 보내 닫는다. 여러 겹이면 맨 위만. 화면마다 붙일 것 없음.
 * - `useBackToClose(() => 열림 && 휴대폰, 닫기)` — 창이 아닌 「덮는 화면」(휴대폰 메일 읽기 등) 에 직접. 컴포넌트 초기화 때 부른다.
 *   닫기가 false 를 돌려주면 「안 닫힘」 으로 보고 다시 쌓는다(「버릴까요?」 취소).
 */

type Layer = { id: number; popped: boolean; alive: boolean; t: number; close: () => unknown; auto?: boolean; claimed?: boolean };

const stack: Layer[] = [];
let seq = 0;
let skipPops = 0;
let listening = false;

function unwind(layer: Layer) {
	if (window.history.state?.overlay !== layer.id) return;
	skipPops++;
	window.history.back();
}

let rawPush: History['pushState'] = (...a) => window.history.pushState(...a);
let patched = false;
/** 창이 떠 있는 채로 페이지를 옮기면 우리 항목 자리에 새 주소를 바꿔 끼운다(뒤로가기 두 번 방지) */
function patchPushState() {
	if (patched) return;
	patched = true;
	const orig = window.history.pushState.bind(window.history);
	rawPush = orig;
	window.history.pushState = function (data: any, unused: string, url?: string | URL | null) {
		const cur = window.history.state?.overlay;
		const top = cur != null ? stack.find((l) => l.id === cur) : undefined;
		if (top) {
			top.popped = true;
			const i = stack.indexOf(top);
			if (i >= 0) stack.splice(i, 1);
			return window.history.replaceState(data, unused, url);
		}
		return orig(data, unused, url);
	};
}

function onPop() {
	if (skipPops > 0) {
		skipPops--;
		return;
	}
	const top = stack.pop();
	if (!top) return;
	top.popped = true;
	Promise.resolve(top.close()).then((r) => {
		if (r === false && top.alive) push(top);
	});
}

function push(layer: Layer) {
	layer.popped = false;
	stack.push(layer);
	patchPushState();
	// SvelteKit 이 쓰는 history.state 칸을 지우지 않게 지금 값에 덧붙인다
	rawPush({ ...(window.history.state || {}), overlay: layer.id }, '');
	if (!listening) {
		listening = true;
		window.addEventListener('popstate', onPop);
	}
}

/** 직접 쓰기 — 쌓고, 정리 함수를 돌려준다(버튼으로 닫힐 때 그 정리 함수를 부른다) */
export function pushBackLayer(onClose: () => unknown): () => void {
	const layer: Layer = { id: ++seq, popped: false, alive: true, t: Date.now(), close: onClose };
	push(layer);
	return () => {
		layer.alive = false;
		const i = stack.indexOf(layer);
		if (i >= 0) stack.splice(i, 1);
		if (!layer.popped) unwind(layer);
	};
}

/** 컴포넌트 안에서 — `useBackToClose(() => open && isMobile, () => (open = false))` */
export function useBackToClose(active: () => boolean, onClose: () => unknown) {
	$effect(() => {
		if (!active()) return;
		return pushBackLayer(() => onClose());
	});
}

let autoInstalled = false;
function isCoveringOverlay(el: Element) {
	if (!el.classList.contains('MuiModal-root') || el.classList.contains('MuiModal-hidden')) return false;
	return !!el.querySelector('.MuiDialog-paperFullScreen') || el.classList.contains('MuiDrawer-modal');
}
const isOpen = (el: Element) => el.isConnected && !el.classList.contains('MuiModal-hidden');

export function installAutoBackToClose(maxWidth = 899.95) {
	if (autoInstalled || typeof window == 'undefined') return;
	autoInstalled = true;
	const tracked = new Map<Element, Layer | 'hook'>();
	let scheduled = false;
	const scan = () => {
		scheduled = false;
		const mobile = window.matchMedia(`(max-width:${maxWidth}px)`).matches;
		const now = new Set<Element>();
		if (mobile) for (const el of Array.from(document.body.children)) if (isCoveringOverlay(el)) now.add(el);
		for (const [el, layer] of tracked) {
			if (now.has(el)) continue;
			tracked.delete(el);
			if (layer == 'hook') continue;
			layer.alive = false;
			const i = stack.indexOf(layer);
			if (i >= 0) stack.splice(i, 1);
			if (!layer.popped) unwind(layer);
		}
		for (const el of now) {
			if (tracked.has(el)) continue;
			tracked.set(el, 'hook');
			const seen = Date.now();
			setTimeout(() => {
				if (!isOpen(el) || tracked.get(el) != 'hook') return;
				const hookLayer = stack.find((l) => !l.auto && !l.claimed && l.t >= seen - 300);
				if (hookLayer) {
					hookLayer.claimed = true;
					return;
				}
				const layer: Layer = {
					id: ++seq,
					popped: false,
					alive: true,
					t: Date.now(),
					auto: true,
					close: () =>
						new Promise((resolve) => {
							el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
							setTimeout(() => resolve(!(isOpen(el) && tracked.get(el) === layer)), 600);
						})
				};
				tracked.set(el, layer);
				push(layer);
			}, 80);
		}
	};
	const schedule = () => {
		if (scheduled) return;
		scheduled = true;
		requestAnimationFrame(scan);
	};
	new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['class'] });
	window.addEventListener('resize', schedule);
	schedule();
}
