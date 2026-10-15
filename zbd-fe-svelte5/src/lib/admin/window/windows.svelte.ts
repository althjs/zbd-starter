/**
 * 화면 위 창(React components/common/window-control + store/useWindowStore) — 관리자 화면 안에 떠 있는 여러 창.
 *
 * - 창 하나 = 경로 화면(`path` → 같은 출처 iframe `경로?popup=1`, 메뉴 없는 틀) 또는 Svelte 부품(`component` + `props`).
 * - 창은 관리자 껍데기의 `<WindowHost/>` 가 그린다 → 메뉴를 옮겨 다녀도 남는다.
 * - 같은 id 로 다시 열면 새로 만들지 않고 복원(최소화였으면) + 맨 앞. 경로 창 id = `win_` + base64(경로)(React 와 같은 규칙).
 * - 자리·크기는 localStorage `zbd:admin:window:{창 id}` 에 `{position, size}`(옛 React 키 `win_{창 id}` 는 migrateStorage 가 옮긴다).
 * - 순서(z-index) = order 배열의 뒤쪽이 맨 앞. 창 z-index 는 1250 + 순번(메뉴 바 1100·서랍 1200 위, 팝오버·Dialog 1300 아래 — 창 안에서 연 팝오버가 창 뒤로 숨지 않게).
 * - 창 안(iframe) 에서 경로 창을 열면 맨 위 문서의 창 시스템으로 넘긴다(창 안에 창을 만들지 않는다).
 *   부품 창은 넘기지 않는다 — 다른 문서의 Svelte 부품은 맨 위 문서에서 그릴 수 없다(런타임이 따로). 그 iframe 안에 뜬다.
 */
import type { Component } from 'svelte';
import { IN_WIN_FRAME } from './frame.ts';
import { SK } from '#lib/shared/storage-keys.ts';

export type ViewState = 'normal' | 'minimized' | 'maximized';
type Pt = { x: number; y: number };
type Sz = { width: number; height: number };

export type WinOptions = {
	/** 내용 폭(px) — 없으면 화면 폭의 90% */
	width?: number;
	/** 내용 높이(px, 제목줄 48 제외) — 없으면 화면 높이의 80% */
	height?: number;
	initialState?: 'normal' | 'maximized';
};

export type WinOpen = WinOptions & {
	id: string;
	title: string;
	/** 경로 화면 — iframe 으로 */
	path?: string;
	/** Svelte 부품 — 직접 그림(props 에 windowId 를 더해 넘긴다) */
	component?: Component<any>;
	props?: Record<string, any>;
};

export type Win = {
	id: string;
	title: string;
	path?: string;
	component?: Component<any>;
	props?: Record<string, any>;
	view: ViewState;
	/** 최소화 전 상태(복원 때) */
	prevView: ViewState;
	pos: Pt;
	size: Sz;
	/** 최대화·스냅 전 자리(되돌릴 때) */
	preMax: { pos: Pt; size: Sz } | null;
	preSnap: { pos: Pt; size: Sz } | null;
	snapped: boolean;
	initW: number;
	initH: number;
	/** 제목줄의 최소화 앞에 놓는 내용 쪽 버튼(이미지 편집 도구의 「다시 선택」 등) */
	actions?: WinAction[];
};
export type WinAction = { icon: string; title: string; onclick: () => void };

export const HEADER_HEIGHT = 48;
export const MIN_SIZE: Sz = { width: 600, height: 500 }; // React ResizableWindowContainer minConstraints 기본값
export const Z_BASE = 1250;
const LS_DOCK = SK.admin.showMinimizedDock;

const lsGet = (k: string) => {
	try {
		return localStorage.getItem(k);
	} catch {
		return null;
	}
};
const lsSet = (k: string, v: string | null) => {
	try {
		if (v == null) localStorage.removeItem(k);
		else localStorage.setItem(k, v);
	} catch {
		/* 저장소가 막혀 있어도 창은 연다 */
	}
};

export const wm = $state({
	list: [] as Win[],
	/** 뒤쪽이 맨 앞 */
	order: [] as string[],
	/** 최소화한 창을 아래 줄(독)에 띄울지 — React 「최소화바 표시」 */
	showDock: typeof window == 'undefined' ? true : lsGet(LS_DOCK) !== 'false'
});

/** 창을 닫기 전에 가로챌 것(「저장 안 한 내용」 확인 등) — React registerCloseHandler */
const closeHandlers = new Map<string, () => void>();
export const registerCloseHandler = (id: string, fn: () => void) => (closeHandlers.set(id, fn), () => closeHandlers.delete(id));

/** 경로 창 id — React WindowOpenerForRoutePath 와 같은 계산(한글 등 비 Latin1 은 UTF-8 로 바꿔서) */
export function routeWindowId(path: string) {
	let b = '';
	try {
		b = btoa(path);
	} catch {
		b = btoa(unescape(encodeURIComponent(path)));
	}
	return `win_${b.replace(/[^a-zA-Z0-9]/g, '')}`;
}

export const getWin = (id: string) => wm.list.find((w) => w.id === id);
export const zOf = (id: string) => {
	const i = wm.order.indexOf(id);
	return Math.min(1299, Z_BASE + (i < 0 ? 0 : i));
};
export const topId = () => {
	for (let i = wm.order.length - 1; i >= 0; i--) {
		const w = getWin(wm.order[i]);
		if (w && w.view !== 'minimized') return w.id;
	}
	return null;
};

function loadSaved(id: string): { position: Pt; size: Sz } | null {
	const raw = lsGet(SK.admin.window(id));
	if (!raw) return null;
	try {
		const { position, size } = JSON.parse(raw);
		// 화면 밖이면 버린다(50px 이상 보여야 — React 와 같은 검사)
		if (position.x > -size.width + 50 && position.y > -size.height + 50 && position.x < window.innerWidth - 50 && position.y < window.innerHeight - 50) return { position, size };
	} catch {
		/* 깨진 값 무시 */
	}
	return null;
}
export function saveGeom(w: Win) {
	if (w.view !== 'normal' || w.snapped) return; // 최대화·최소화·스냅 상태는 저장하지 않는다(React 와 같게)
	lsSet(SK.admin.window(w.id), JSON.stringify({ position: $state.snapshot(w.pos), size: $state.snapshot(w.size) }));
}
function defaultPos(id: string, width: number, height: number): Pt {
	let x = Math.max(0, (window.innerWidth - width) / 2);
	let y = Math.max(0, (window.innerHeight - height - HEADER_HEIGHT) / 2);
	const idx = wm.list.findIndex((w) => w.id === id);
	if (idx > 0) {
		x += idx * 30; // 겹치지 않게 30px 씩
		y += idx * 30;
	}
	return { x, y };
}

export function bringToFront(id: string) {
	if (wm.order[wm.order.length - 1] === id) return;
	wm.order = [...wm.order.filter((x) => x !== id), id];
}

/** 열기 — 이미 있으면 복원(최소화였으면) + 맨 앞. 경로·부품·props 는 바꾸지 않는다(React 와 같게 — 바꾸려면 닫고 다시) */
export function openWindow(o: WinOpen) {
	if (IN_WIN_FRAME && !o.component) {
		const top = (window.top as any)?.__zbdWindows;
		if (top?.openWindow) return top.openWindow(o);
	}
	const ex = getWin(o.id);
	if (ex) {
		if (ex.view === 'minimized') ex.view = ex.prevView === 'minimized' ? 'normal' : ex.prevView;
		if (o.title && !ex.path) ex.title = o.title;
		bringToFront(o.id);
		return;
	}
	const initW = Math.round(o.width || window.innerWidth * 0.9);
	const initH = Math.round(o.height || window.innerHeight * 0.8);
	// 목록에 먼저 넣어야 겹침 오프셋(순번) 이 잡힌다
	const w: Win = {
		id: o.id,
		title: o.title || o.path || '창',
		path: o.path,
		component: o.component,
		props: o.props,
		view: 'normal',
		prevView: 'normal',
		pos: { x: 0, y: 0 },
		size: { width: initW, height: initH },
		preMax: null,
		preSnap: null,
		snapped: false,
		initW,
		initH
	};
	wm.list.push(w);
	const live = getWin(o.id)!;
	const saved = loadSaved(o.id);
	if (o.initialState === 'maximized') {
		live.preMax = { pos: saved?.position || defaultPos(o.id, initW, initH), size: saved?.size || { width: initW, height: initH } };
		live.view = 'maximized';
	} else if (saved) {
		live.pos = saved.position;
		live.size = saved.size;
	} else live.pos = defaultPos(o.id, initW, initH);
	bringToFront(o.id);
	window.getSelection()?.removeAllRanges(); // Shift+누르기로 생긴 글자 선택 해제
}

/** 경로 화면을 창으로(왼쪽 메뉴 창 아이콘 · Shift+링크) */
export function openRouteWindow(path: string, title?: string, opts: WinOptions = {}) {
	if (IN_WIN_FRAME) {
		const top = (window.top as any)?.__zbdWindows;
		if (top?.openRouteWindow) return top.openRouteWindow(path, title, opts);
	}
	openWindow({ id: routeWindowId(path), title: title || path, path, ...opts });
}

export function closeWindow(id: string) {
	wm.list = wm.list.filter((w) => w.id !== id);
	wm.order = wm.order.filter((x) => x !== id);
	closeHandlers.delete(id);
}
export function requestCloseWindow(id: string) {
	const h = closeHandlers.get(id);
	if (h) h();
	else closeWindow(id);
}
export function closeAllWindows() {
	// React 와 같게 — 모두 닫으면 기억한 자리도 지운다
	for (const w of wm.list) lsSet(SK.admin.window(w.id), null);
	wm.list = [];
	wm.order = [];
	closeHandlers.clear();
}

export function minimize(id: string) {
	const w = getWin(id);
	if (!w || w.view === 'minimized') return;
	w.prevView = w.view;
	w.view = 'minimized';
}
/** 독·목록에서 누름 — 복원 + 맨 앞 (React focusWindow) */
export function focusWindow(id: string) {
	const w = getWin(id);
	if (!w) return;
	if (w.view === 'minimized') w.view = w.prevView === 'minimized' ? 'normal' : w.prevView;
	bringToFront(id);
}
export function minimizeAll() {
	for (const w of wm.list) if (w.view !== 'minimized') ((w.prevView = w.view), (w.view = 'minimized'));
}
export function toggleMaximize(id: string) {
	const w = getWin(id);
	if (!w) return;
	if (w.view === 'maximized') {
		const p = w.preMax || { pos: defaultPos(id, w.initW, w.initH), size: { width: w.initW, height: w.initH } };
		w.pos = p.pos;
		w.size = p.size;
		w.view = 'normal';
		w.snapped = false;
	} else {
		w.preMax = w.snapped && w.preSnap ? w.preSnap : { pos: { ...w.pos }, size: { ...w.size } };
		w.view = 'maximized';
		w.snapped = false;
	}
}
/** Ctrl+Shift+화살표 — 화면 반쪽에 붙이기(React 스냅) */
export function snap(id: string, dir: 'ArrowLeft' | 'ArrowRight' | 'ArrowUp' | 'ArrowDown') {
	const w = getWin(id);
	if (!w) return;
	if (w.view === 'maximized') return void (w.view = 'normal');
	if (!w.snapped) ((w.preSnap = { pos: { ...w.pos }, size: { ...w.size } }), (w.snapped = true));
	const W = window.innerWidth;
	const H = window.innerHeight;
	if (dir === 'ArrowLeft') ((w.pos = { x: 0, y: 0 }), (w.size = { width: W / 2, height: H - HEADER_HEIGHT }));
	else if (dir === 'ArrowRight') ((w.pos = { x: W / 2, y: 0 }), (w.size = { width: W / 2, height: H - HEADER_HEIGHT }));
	else if (dir === 'ArrowUp') ((w.pos = { x: 0, y: 0 }), (w.size = { width: W, height: H / 2 - HEADER_HEIGHT }));
	else ((w.pos = { x: 0, y: H / 2 }), (w.size = { width: W, height: H / 2 - HEADER_HEIGHT }));
}
/** 스냅 상태에서 끌거나 크기를 바꾸려 하면 원래 자리로(React 와 같게 — 그 동작은 취소) */
export function unsnap(w: Win) {
	if (!w.snapped || !w.preSnap) return false;
	w.pos = w.preSnap.pos;
	w.size = w.preSnap.size;
	w.snapped = false;
	return true;
}
export function setTitle(id: string, title: string) {
	const w = getWin(id);
	if (w && title && w.title !== title) w.title = title;
}
export function setWindowActions(id: string, actions: WinAction[]) {
	const w = getWin(id);
	if (w) w.actions = actions;
}
export function toggleShowDock() {
	wm.showDock = !wm.showDock;
	lsSet(LS_DOCK, String(wm.showDock));
}

// 창 안 문서(iframe)가 맨 위 창 시스템을 찾는 자리
if (typeof window != 'undefined' && !IN_WIN_FRAME) (window as any).__zbdWindows = { openWindow, openRouteWindow };
