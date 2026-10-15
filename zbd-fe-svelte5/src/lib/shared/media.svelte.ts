/**
 * 화면 크기·스크롤 — 예전 useMediaQuery(theme.breakpoints.down("md")) · store/window-size 자리.
 * isMobile = 폭 900 미만(MUI md 아래). 관리자·사용자 공용.
 */
export const viewport = $state({ width: 1280, height: 900, scrollY: 0, isMobile: false });

if (typeof window != 'undefined') {
	const read = () => {
		viewport.width = window.innerWidth;
		viewport.height = window.innerHeight;
		viewport.isMobile = window.innerWidth < 900;
	};
	read();
	viewport.scrollY = window.scrollY;
	window.addEventListener('resize', read);
	window.addEventListener('scroll', () => (viewport.scrollY = window.scrollY), { passive: true });
}
