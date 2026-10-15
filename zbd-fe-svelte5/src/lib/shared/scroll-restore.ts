/**
 * 뒤로·앞으로 가기 때 스크롤 자리 되살리기 — 관리자·사용자 화면 공용(루트 layout 에서 한 번).
 *
 * SvelteKit 도 popstate 때 window 스크롤을 되돌리지만, 우리 화면은 대부분 그린 뒤에 데이터를 받아 오므로
 * 그 순간엔 문서가 짧아 맨 위로 잘린다(SPA 고질병). → 주소마다 마지막 자리를 기억해 두고,
 * 뒤로 가면 문서가 그만큼 길어질 때까지(최대 4초) 기다렸다 맞춘다. 사용자가 그사이 스크롤·터치·키를 쓰면 멈춘다.
 */
import { afterNavigate, beforeNavigate } from '$app/navigation';
import { SK } from '#lib/shared/storage-keys.ts';

const KEY = SK.nav.scroll;
const WAIT_MS = 4000;

function read(): Record<string, number> {
	try {
		return JSON.parse(sessionStorage.getItem(KEY) || '{}');
	} catch {
		return {};
	}
}
function write(m: Record<string, number>) {
	try {
		const keys = Object.keys(m);
		// 너무 많이 쌓이지 않게 최근 200개만
		if (keys.length > 200) for (const k of keys.slice(0, keys.length - 200)) delete m[k];
		sessionStorage.setItem(KEY, JSON.stringify(m));
	} catch {}
}
const urlKey = (u: URL | Location) => u.pathname + u.search;

let stop: (() => void) | null = null;

function restoreTo(y: number) {
	stop?.();
	if (y <= 0) return;
	const until = Date.now() + WAIT_MS;
	let raf = 0;
	let done = false;
	const cancel = () => {
		done = true;
		cancelAnimationFrame(raf);
		for (const ev of ['wheel', 'touchstart', 'keydown', 'mousedown'] as const) removeEventListener(ev, cancel, true);
		stop = null;
	};
	for (const ev of ['wheel', 'touchstart', 'keydown', 'mousedown'] as const) addEventListener(ev, cancel, true);
	stop = cancel;
	const tick = () => {
		if (done) return;
		const max = document.documentElement.scrollHeight - innerHeight;
		if (max >= y - 2) {
			scrollTo(0, y);
			// 그 뒤로도 그림이 들어와 밀릴 수 있어 잠깐 더 붙잡는다
			if (Math.abs(scrollY - y) < 2 && Date.now() > until - WAIT_MS + 300) return cancel();
		} else if (max > 0) scrollTo(0, max);
		if (Date.now() < until) raf = requestAnimationFrame(tick);
		else cancel();
	};
	raf = requestAnimationFrame(tick);
}

/** 루트 layout 의 <script> 에서 한 번 부른다 */
export function setupScrollRestore() {
	if (typeof history != 'undefined') history.scrollRestoration = 'manual';
	beforeNavigate(({ from }) => {
		if (!from?.url) return;
		const m = read();
		m[urlKey(from.url)] = Math.round(scrollY);
		write(m);
	});
	afterNavigate(({ type, to }) => {
		if (type !== 'popstate' || !to?.url) return stop?.();
		restoreTo(read()[urlKey(to.url)] || 0);
	});
}
