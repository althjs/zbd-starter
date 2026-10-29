/**
 * 왼쪽 메뉴 상태 + 지금 화면의 메뉴 위치 (React store/zustandNavInfo 자리).
 *
 *   nav.open / nav.mini        펼침·축약·숨김 — ☰ 와 Ctrl/Cmd+B 는 toggleSideNavByUser(PC 여부)
 *   nav.current                지금 메뉴(페이지 제목·이동 경로) — 레이아웃이 주소가 바뀔 때 syncCurrentNav 로 맞춘다
 *   activeNav()                권한 판정용(하위 메뉴가 켜져 있으면 그것) — AuthButton 이 본다
 *   setSubDepthTitle(title)    PageContainer 가 제목을 따로 주면 이동 경로 끝에 붙는다
 */
import { LS } from './config.ts';
import type { CurrentNav, NavItem } from './types.ts';

type Mode = 'full' | 'mini' | 'hidden';
function loadMode(): Mode {
	try {
		const v = localStorage.getItem(LS.sideNavMode);
		if (v === 'full' || v === 'mini' || v === 'hidden') return v;
		return 'full'; // 아주 옛 SIDE_NAV_OPEN('0') 은 migrateStorage 가 모드로 옮겨 둔다
	} catch {
		return 'full';
	}
}
const initial: Mode = typeof window != 'undefined' ? loadMode() : 'full';

export const nav = $state({
	open: initial != 'hidden',
	mini: initial == 'mini',
	current: null as CurrentNav | null,
	/** setCurrentNav 로 받은 원래 값(subDepthTitle 없는) */
	base: null as CurrentNav | null
});

/**
 * 넓은 화면(PC)이 되면 저장해 둔 펼침·축약·숨김으로 되돌린다 — 좁은 화면에서 서랍을 닫은 것이 PC 메뉴까지 닫아 두지 않게.
 * 주소창에 치면 크롬이 화면을 미리 띄워 두는(prerender) 동안 폭이 좁게 잡혀 메뉴가 닫힌 채로 열렸다(2026-10-30, 새로고침하면 보임)
 */
export function restoreSideNav() {
	const m = loadMode();
	nav.open = m != 'hidden';
	nav.mini = m == 'mini';
}

export function setOpenSideNav(tf: boolean) {
	nav.open = tf;
}

/** PC: 펼침 → 축약 → 숨김 순환(기억). 휴대폰: 열고 닫기만 */
export function toggleSideNavByUser(desktop: boolean) {
	if (!desktop) {
		nav.open = !nav.open;
		return;
	}
	const cur: Mode = !nav.open ? 'hidden' : nav.mini ? 'mini' : 'full';
	const next: Mode = cur == 'full' ? 'mini' : cur == 'mini' ? 'hidden' : 'full';
	try {
		localStorage.setItem(LS.sideNavMode, next);
	} catch {}
	nav.mini = next == 'mini' || (next == 'hidden' && nav.mini);
	nav.open = next != 'hidden';
	// 크기에 기대는 화면(표·차트)이 다시 재게
	setTimeout(() => window.dispatchEvent(new Event('resize')));
}

export function setCurrentNav(item: CurrentNav, titlePrefix = '') {
	document.title = `${titlePrefix ? `${titlePrefix} | ` : ''}${item?.sub?.title ? `${item.sub.title} < ` : ''}${item?.title}`;
	nav.base = item;
	nav.current = { ...item, subDepthTitle: null };
}

export function setSubDepthTitle(title: string | null | undefined) {
	if (!nav.base) return;
	nav.current = { ...nav.base, subDepthTitle: title || null };
}

export function activeNav(): (CurrentNav | NavItem) | null {
	const c = nav.current;
	return c?.sub?.active ? c.sub : c;
}

/**
 * 주소가 메뉴 주소 「안」 인지 — 같거나, 메뉴 주소 뒤가 경로 경계(/ ? #)일 때만.
 * 그냥 앞부분 비교면 /admin/board/FREECELL 이 /admin/board/FREE(자유게시판)에도 걸리고, 빈 주소는 모든 주소에 걸린다.
 */
export function pathUnder(pathname: string, p: string | null | undefined): boolean {
	if (!p) return false;
	if (pathname === p) return true;
	if (pathname.indexOf(p) !== 0) return false;
	return p.endsWith('/') || '/?#'.includes(pathname.charAt(p.length));
}

/** 하위 메뉴 중 지금 주소에 맞는 것 하나(주소가 가장 긴 것) — 프리셀(FREECELL)·자유게시판(FREE) 이 같이 켜지지 않게 */
export function bestSubPath(items: { path?: string | null }[], pathname: string): string | null {
	let best: string | null = null;
	for (const s of items || []) {
		const p = String(s.path || '').split('?')[0];
		if (pathUnder(pathname, p) && (!best || p.length > best.length)) best = p;
	}
	return best;
}

/**
 * 주소 → 지금 메뉴. React 는 메뉴 줄마다 효과(effect)로 맞췄다 — 같은 순서로 훑어 마지막에 맞은 것이 이긴다.
 * 위(권한 메뉴) → 아래 보조 메뉴(고객센터) 순. 웹메일(special mail) 은 그 주소 안이면 메뉴 자체.
 */
export function syncCurrentNav(items: NavItem[], subMenus: NavItem[], pathname: string, titlePrefix: string) {
	// 여럿이 맞으면 주소가 가장 긴(가장 구체적인) 메뉴. 같은 길이면 뒤의 것(React 와 같은 순서)
	let found: CurrentNav | null = null;
	let best = -1;
	const take = (p: string | null | undefined, v: CurrentNav) => {
		const n = String(p || '').length;
		if (n >= best) ((best = n), (found = v));
	};
	for (const it of [...items, ...subMenus]) {
		if (it.special == 'mail' && !(it.items || []).length) {
			if (pathUnder(pathname, it.path || '/admin/mail')) take(it.path || '/admin/mail', { ...it });
			continue;
		}
		if (it.items && it.items.length) {
			for (const s of it.items) s.active = false;
			const hits = it.items.filter((s) => pathUnder(pathname, s.path));
			const top = hits.sort((a, b) => String(b.path).length - String(a.path).length)[0];
			if (top) {
				top.active = true;
				take(top.path, { title: it.title, sub: top });
			}
		} else if (pathUnder(pathname, it.path)) {
			take(it.path, { ...it });
		}
	}
	if (found) setCurrentNav(found, titlePrefix);
}
