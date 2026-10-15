/**
 * 화면 폭 구간 — React useMediaQuery(theme.breakpoints) 자리. 관리자 테마 구간: sm 600 · md 900 · lg 1280 · xl 1440.
 *   media.mdDown  (휴대폰·태블릿, < 900)   media.lgUp (≥ 1280)   media.smDown (< 600)
 */
export const media = $state({ smDown: false, mdDown: false, lgUp: true, xlUp: true });

if (typeof window != 'undefined') {
	const q = (s: string, k: keyof typeof media) => {
		const m = window.matchMedia(s);
		media[k] = m.matches;
		m.addEventListener('change', (e) => (media[k] = e.matches));
	};
	q('(max-width:599.95px)', 'smDown');
	q('(max-width:899.95px)', 'mdDown');
	q('(min-width:1280px)', 'lgUp');
	q('(min-width:1440px)', 'xlUp');
}
