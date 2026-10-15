/**
 * 사이트 대표색 — siteMeta.design.theme 이름으로 팔레트를 골라 `:root` 변수로 깐다.
 * 변수 이름은 React 판(layouts/main/layout.tsx) 과 같다 — 운영자 CSS 가 `var(--theme-color-main)` 등을 쓴다.
 */
import { applyCSS } from '../css';
import { PALETTES, type Palette } from './palettes';

export function getPalette(name?: string): Palette {
	return PALETTES[name || ''] || PALETTES.blue;
}

/** WCAG 대비(React 판 getContrastRatio 와 같은 계산) */
function luminance(hex: string) {
	const h = hex.replace('#', '');
	const n =
		h.length == 3
			? h
					.split('')
					.map((c) => c + c)
					.join('')
			: h.slice(0, 6);
	const rgb = [0, 2, 4].map((i) => parseInt(n.slice(i, i + 2), 16) / 255).map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
	return 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2];
}
export function contrastRatio(a: string, b: string) {
	const [x, y] = [luminance(a), luminance(b)];
	return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05);
}

export function applyTheme(name?: string) {
	const p = getPalette(name);
	applyCSS(
		`:root {
  --theme-color-lightest: ${p.lightest};
  --theme-color-light: ${p.light};
  --theme-color-main: ${p.main};
  --theme-color-dark: ${p.dark};
  --theme-color-darkest: ${p.darkest};
  --theme-color-contrastText: ${contrastRatio(p.main, '#fff') > 3 ? '#fff' : '#111'};
  --theme-color-alpha4: ${p.alpha4};
  --theme-color-alpha8: ${p.alpha8};
  --theme-color-alpha12: ${p.alpha12};
  --theme-color-alpha50: ${p.alpha50};
  --theme-color-alpha30: ${p.alpha30};
}`,
		'octopus-theme-css-var'
	);
	return p;
}
