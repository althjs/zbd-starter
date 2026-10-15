/**
 * 관리자 테마 — 사이트 메타 SITE.theme(대표색) + 라이트/다크를 CSS 변수로 깐다(React createTheme 자리).
 *
 * 변수는 `html.zbd-admin` 아래에만 둔다 — 관리자 화면을 떠나면(클래스가 빠지면) 사용자 화면 값으로 돌아간다.
 *   --theme-color-*      대표색(사용자 화면과 같은 이름 — 공용 부품·운영자 CSS 가 쓴다)
 *   --adm-secondary-* · --adm-success · --adm-error · --adm-warning · --adm-info
 *   --adm-n50 … --adm-n900   neutral(다크면 밝은/어두운 단계가 뒤집힌다)
 *   --adm-text · --adm-text2 · --adm-bg · --adm-paper · --adm-card · --adm-divider · --adm-action · --adm-hover · --adm-cell-line
 *   --gt-* · --zc-*      다크일 때만(라이트는 var(--gt-200, #f2f2f2) 의 기본값 그대로) — React colors.ts·create-components.ts 와 같은 값
 */
import { applyCSS } from '#lib/shared/css.ts';
import { contrastRatio } from '#lib/shared/theme/theme.ts';
import { ADMIN_STATUS, ADMIN_THEMES, ADMIN_THEME_SECONDARY, GRAY_TONE_DARK, NEUTRAL, type Palette } from './palettes.ts';

/** MUI 기본 보조색(테마에 보조색이 없을 때) */
const MUI_SECONDARY = {
	lightest: '#f3e5f5',
	light: '#ba68c8',
	main: '#9c27b0',
	dark: '#7b1fa2',
	darkest: '#4a148c',
	contrastText: '#fff',
	alpha4: 'rgba(156,39,176,0.04)',
	alpha8: 'rgba(156,39,176,0.08)',
	alpha12: 'rgba(156,39,176,0.12)',
	alpha30: 'rgba(156,39,176,0.3)',
	alpha50: 'rgba(156,39,176,0.5)'
};

const DARK_CODE_COLORS: Record<string, string> = {
	'--zc-fff': '#1a1a1a',
	'--zc-ffffff': '#1a1a1a',
	'--zc-white': '#1a1a1a',
	'--zc-fafafa': '#202020',
	'--zc-f9fafb': '#202020',
	'--zc-f8f9fa': '#202020',
	'--zc-f8fafc': '#202020',
	'--zc-f7f7f7': '#222222',
	'--zc-f6f6f6': '#232323',
	'--zc-f5f5f5': '#252525',
	'--zc-f3f4f6': '#262626',
	'--zc-f2f2f2': '#262626',
	'--zc-f0f0f0': '#2a2a2a',
	'--zc-eee': '#2e2e2e',
	'--zc-eeeeee': '#2e2e2e',
	'--zc-f9f9f9': '#212121',
	'--zc-f6f8f7': '#222524',
	'--zc-f4f4f4': '#262626',
	'--zc-fcfcfc': '#1d1d1d',
	'--zc-fbfbfb': '#1e1e1e',
	'--zc-000': '#f0f0f0',
	'--zc-000000': '#f0f0f0',
	'--zc-111': '#ececec',
	'--zc-111111': '#ececec',
	'--zc-1b1b1b': '#e8e8e8',
	'--zc-222': '#e0e0e0',
	'--zc-222222': '#e0e0e0',
	'--zc-333': '#d6d6d6',
	'--zc-333333': '#d6d6d6',
	'--zc-101828': '#e6e6e6',
	'--zc-344054': '#c8ccd2',
	'--zc-444': '#c4c4c4',
	'--zc-555': '#b0b0b0',
	'--zc-666': '#a0a0a0',
	'--zc-1d433d': '#8fd3c4',
	'--zc-5b6b66': '#a9b6b2'
};

const NEUTRAL_DARK = { ...NEUTRAL, 50: '#222326', 100: '#26282c', 200: '#30333a', 300: '#3c4048', 600: '#c1c7cf', 700: '#d2d6db', 800: '#e5e7eb', 900: '#f3f4f6' } as Record<string, string>;

export function getAdminPalette(name?: string): Palette {
	return ADMIN_THEMES[name || ''] || ADMIN_THEMES.blue;
}

/** hex → rgba(…, a) */
export function alpha(hex: string, a: number) {
	const h = hex.replace('#', '');
	const n =
		h.length == 3
			? h
					.split('')
					.map((c) => c + c)
					.join('')
			: h.slice(0, 6);
	const [r, g, b] = [0, 2, 4].map((i) => parseInt(n.slice(i, i + 2), 16));
	return `rgba(${r}, ${g}, ${b}, ${a})`;
}

function vars(prefix: string, p: Palette) {
	return Object.entries(p)
		.map(([k, v]) => `${prefix}-${k}: ${v};`)
		.join('\n');
}

export function applyAdminTheme(name: string | undefined, mode: 'light' | 'dark') {
	const p = getAdminPalette(name);
	const s = ADMIN_THEME_SECONDARY[name || ''] || MUI_SECONDARY;
	const dark = mode == 'dark';
	const n = dark ? NEUTRAL_DARK : (NEUTRAL as Record<string, string>);
	const lines = [
		vars('--theme-color', p),
		`--theme-color-contrastText: ${contrastRatio(p.main, '#fff') > 3 ? '#fff' : '#111'};`,
		vars('--adm-secondary', s),
		...(['success', 'error', 'warning', 'info'] as const).map((k) => vars(`--adm-${k}`, ADMIN_STATUS[k] as Palette)),
		...Object.entries(n).map(([k, v]) => `--adm-n${k}: ${v};`),
		`--adm-text: ${dark ? NEUTRAL[100] : NEUTRAL[900]};`,
		`--adm-text2: ${dark ? NEUTRAL[400] : NEUTRAL[500]};`,
		`--adm-text-disabled: ${dark ? alpha(NEUTRAL[100], 0.38) : alpha(NEUTRAL[900], 0.38)};`,
		`--adm-action: ${dark ? NEUTRAL[400] : NEUTRAL[500]};`,
		`--adm-hover: ${dark ? alpha(NEUTRAL[100], 0.04) : alpha(NEUTRAL[900], 0.04)};`,
		`--adm-selected: ${dark ? alpha(NEUTRAL[100], 0.12) : alpha(NEUTRAL[900], 0.12)};`,
		`--adm-disabled-bg: ${dark ? alpha(NEUTRAL[100], 0.12) : alpha(NEUTRAL[900], 0.12)};`,
		`--adm-bg: ${dark ? '#0a0a0a' : '#fff'};`,
		`--adm-paper: ${dark ? '#1a1a1a' : '#fff'};`,
		`--adm-card: ${dark ? '#1e1e1e' : '#ffffff'};`,
		`--adm-divider: ${dark ? '#333' : '#ddd'};`,
		`--adm-cell-line: ${dark ? '#333' : '#eee'};`,
		`--adm-header-bg: ${alpha(p.dark, 0.8)};`,
		`--adm-scroll: ${dark ? '#6b6b6b #2b2b2b' : '#c1c1c1 #f1f1f1'};`,
		`color-scheme: ${dark ? 'dark' : 'light'};`,
		...(dark ? Object.entries(GRAY_TONE_DARK).map(([k, v]) => `--gt-${k}: ${v};`) : []),
		...(dark ? Object.entries(DARK_CODE_COLORS).map(([k, v]) => `${k}: ${v};`) : [])
	];
	applyCSS(`html.zbd-admin { ${lines.join(' ')} }`, 'zbd-admin-theme');
	return p;
}
