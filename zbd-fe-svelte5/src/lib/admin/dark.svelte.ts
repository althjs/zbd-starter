/**
 * 다크모드 — 라이트 / 다크 / 자동(시스템 따라) (React hooks/useDarkMode 와 같은 저장 키·규칙).
 *   dark.scheme   고른 값  ·  dark.resolved   실제 적용('light' | 'dark')
 *   setDarkMode('dark')
 * 모듈 하나가 상태를 들고 있어 머리줄에서 바꾸면 테마를 만드는 레이아웃이 바로 따라간다.
 */
import { LS } from './config.ts';

export type ColorScheme = 'light' | 'dark' | 'auto';

const sys = () => (typeof window != 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
function initial(): ColorScheme {
	try {
		const v = localStorage.getItem(LS.dark) as ColorScheme;
		if (v && ['light', 'dark', 'auto'].includes(v)) return v;
	} catch {}
	return 'auto';
}
const first = typeof window != 'undefined' ? initial() : 'auto';

export const dark = $state({ scheme: first as ColorScheme, resolved: (first == 'auto' ? sys() : first) as 'light' | 'dark' });

export function setDarkMode(next: ColorScheme) {
	dark.scheme = next;
	dark.resolved = next == 'auto' ? sys() : next;
	try {
		localStorage.setItem(LS.dark, next);
	} catch {}
}

if (typeof window != 'undefined' && window.matchMedia) {
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
		if (dark.scheme != 'auto') return;
		dark.resolved = e.matches ? 'dark' : 'light';
	});
}
