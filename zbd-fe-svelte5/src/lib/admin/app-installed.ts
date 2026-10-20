/**
 * 설치 앱(PWA) 「설치됨」 표시 — 대시보드 「웹앱 설치」 바로가기가 이미 설치한 앱을 숨기는 데 쓴다(2026-10-16 대표).
 * 브라우저는 다른 주소(chat.·mail.…)에 설치된 앱을 알려 주지 않는다 → 앱이 설치 창으로 열릴 때(standalone)·설치될 때(appinstalled)
 * 그 앱 키를 쿠키 `zbd_apps` 에 남긴다. 쿠키 범위 = 관리자·단독 앱 주소의 공통 상위 주소(admin.X·chat.X → X) — 대시보드(관리자 주소)가 읽는다.
 * 앱을 지워도 브라우저는 알려 주지 않아 표시는 남는다(180일, 앱을 열 때마다 늘어남) — 지운 뒤 다시 보이게 하려면 「다시 보기」.
 */
import { ADMIN_HOSTS, adminCookieParent } from '#lib/shared/config.ts';
import { getCookie, setCookie } from '#lib/shared/cookie.ts';

const NAME = 'zbd_apps';

/** 쿠키 범위 — 관리자 주소들(첫 칸을 뗀 값)이 모두 같은 상위 주소면 그것. 아니면 ""(이 주소 전용). 순수 함수 */
export function appCookieDomain(hosts: string[], here: string): string {
	const parent = (h: string) => {
		const i = h.indexOf('.');
		return i > 0 && h.indexOf('.', i + 1) > 0 ? h.slice(i + 1) : '';
	};
	const p = parent(here);
	if (!p || !hosts.length) return '';
	return hosts.every((h) => parent(h.split(':')[0]) == p) && here.endsWith('.' + p) ? p : '';
}

function domain() {
	return typeof location == 'undefined' ? '' : adminCookieParent(location.hostname, ADMIN_HOSTS) || appCookieDomain(ADMIN_HOSTS, location.hostname);
}

export function installedApps(): string[] {
	return (getCookie(NAME) || '').split(',').filter(Boolean);
}

export function markInstalled(key: string, on = true) {
	const now = installedApps().filter((k) => k != key);
	if (on) now.push(key);
	const d = domain();
	setCookie(NAME, now.join(','), { ...(d ? { domain: d } : {}), days: 180 });
}

export function isStandalone() {
	if (typeof window == 'undefined') return false;
	return window.matchMedia?.('(display-mode: standalone)').matches || (navigator as any).standalone === true;
}
