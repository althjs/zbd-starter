/**
 * 관리자 대표색 팔레트 — React 관리자(fe-react) theme/colors.ts 의 값 그대로(사이트 메타 SITE.theme 이름으로 고른다).
 * 스타터는 6가지(blue·skyblue·indigo·purple·green·orange)만 둔다. 사용자 화면 팔레트(shared/theme/palettes.ts)와 값이 일부 다르다(blue 등).
 */
import type { Palette } from '#lib/shared/theme/palettes.ts';
export type { Palette };

export const ADMIN_THEMES: Record<string, Palette> = {
	blue: {
		lightest: '#D1E5FF',
		light: '#559EFF',
		main: '#2985FF',
		dark: '#0060E0',
		darkest: '#0048A8',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(41, 133, 255, 0.04)',
		alpha8: 'rgba(41, 133, 255, 0.08)',
		alpha12: 'rgba(41, 133, 255, 0.12)',
		alpha30: 'rgba(41, 133, 255, 0.3)',
		alpha50: 'rgba(41, 133, 255, 0.5)'
	},
	skyblue: {
		lightest: '#D1EEFF',
		light: '#00A2FF',
		main: '#008FE0',
		dark: '#006EAD',
		darkest: '#005180',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(0, 143, 224, 0.04)',
		alpha8: 'rgba(0, 143, 224, 0.08)',
		alpha12: 'rgba(0, 143, 224, 0.12)',
		alpha30: 'rgba(0, 143, 224, 0.3)',
		alpha50: 'rgba(0, 143, 224, 0.5)'
	},
	indigo: {
		lightest: '#E0E5FF',
		light: '#8094FF',
		main: '#667EFF',
		dark: '#3353FF',
		darkest: '#0025EB',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(102, 126, 255, 0.04)',
		alpha8: 'rgba(102, 126, 255, 0.08)',
		alpha12: 'rgba(102, 126, 255, 0.12)',
		alpha30: 'rgba(102, 126, 255, 0.3)',
		alpha50: 'rgba(102, 126, 255, 0.5)'
	},
	purple: {
		lightest: '#F2EBFF',
		light: '#B088FC',
		main: '#9C6AFB',
		dark: '#7B38FA',
		darkest: '#5106E0',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(156, 106, 251, 0.04)',
		alpha8: 'rgba(156, 106, 251, 0.08)',
		alpha12: 'rgba(156, 106, 251, 0.12)',
		alpha30: 'rgba(156, 106, 251, 0.3)',
		alpha50: 'rgba(156, 106, 251, 0.5)'
	},
	green: {
		lightest: '#C2ECF0',
		light: '#2DABB7',
		main: '#2899A4',
		dark: '#1E737B',
		darkest: '#16545A',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(40, 153, 164, 0.04)',
		alpha8: 'rgba(40, 153, 164, 0.08)',
		alpha12: 'rgba(40, 153, 164, 0.12)',
		alpha30: 'rgba(40, 153, 164, 0.3)',
		alpha50: 'rgba(40, 153, 164, 0.5)'
	},
	orange: {
		lightest: '#FFE4E0',
		light: '#FF705A',
		main: '#F94D37',
		dark: '#C22F19',
		darkest: '#912313',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(249, 77, 55, 0.04)',
		alpha8: 'rgba(249, 77, 55, 0.08)',
		alpha12: 'rgba(249, 77, 55, 0.12)',
		alpha30: 'rgba(249, 77, 55, 0.3)',
		alpha50: 'rgba(249, 77, 55, 0.5)'
	}
};

/** 테마별 보조색(secondary) — 없으면 MUI 기본 보조색 */
export const ADMIN_THEME_SECONDARY: Record<string, Palette> = {
	blue: {
		lightest: '#A3F9FF',
		light: '#00ABB8',
		main: '#0098A3',
		dark: '#00727A',
		darkest: '#00565C',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(0, 152, 163, 0.04)',
		alpha8: 'rgba(0, 152, 163, 0.08)',
		alpha12: 'rgba(0, 152, 163, 0.12)',
		alpha30: 'rgba(0, 152, 163, 0.3)',
		alpha50: 'rgba(0, 152, 163, 0.5)'
	},
	skyblue: {
		lightest: '#7AFFED',
		light: '#00B29B',
		main: '#009985',
		dark: '#007566',
		darkest: '#00574B',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(0, 153, 133, 0.04)',
		alpha8: 'rgba(0, 153, 133, 0.08)',
		alpha12: 'rgba(0, 153, 133, 0.12)',
		alpha30: 'rgba(0, 153, 133, 0.3)',
		alpha50: 'rgba(0, 153, 133, 0.5)'
	},
	indigo: {
		lightest: '#CCEDFF',
		light: '#00A6FF',
		main: '#008FDB',
		dark: '#006DA8',
		darkest: '#00507A',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(0, 143, 219, 0.04)',
		alpha8: 'rgba(0, 143, 219, 0.08)',
		alpha12: 'rgba(0, 143, 219, 0.12)',
		alpha30: 'rgba(0, 143, 219, 0.3)',
		alpha50: 'rgba(0, 143, 219, 0.5)'
	},
	purple: {
		lightest: '#E1E6FE',
		light: '#8899FC',
		main: '#6A80FB',
		dark: '#3351FA',
		darkest: '#0627E4',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(106, 128, 251, 0.04)',
		alpha8: 'rgba(106, 128, 251, 0.08)',
		alpha12: 'rgba(106, 128, 251, 0.12)',
		alpha30: 'rgba(106, 128, 251, 0.3)',
		alpha50: 'rgba(106, 128, 251, 0.5)'
	},
	green: {
		lightest: '#CEF3E3',
		light: '#2CB47C',
		main: '#269C6B',
		dark: '#1D7751',
		darkest: '#165A3E',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(38, 156, 107, 0.04)',
		alpha8: 'rgba(38, 156, 107, 0.08)',
		alpha12: 'rgba(38, 156, 107, 0.12)',
		alpha30: 'rgba(38, 156, 107, 0.3)',
		alpha50: 'rgba(38, 156, 107, 0.5)'
	},
	orange: {
		lightest: '#FDDFB0',
		light: '#FBC060',
		main: '#F89C07',
		dark: '#EE9607',
		darkest: '#955E04',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(248, 156, 7, 0.04)',
		alpha8: 'rgba(248, 156, 7, 0.08)',
		alpha12: 'rgba(248, 156, 7, 0.12)',
		alpha30: 'rgba(248, 156, 7, 0.3)',
		alpha50: 'rgba(248, 156, 7, 0.5)'
	}
};

export const ADMIN_STATUS = {
	success: {
		lightest: '#F0FDF9',
		light: '#3FC79A',
		main: '#10B981',
		dark: '#0B815A',
		darkest: '#134E48',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(16, 185, 129, 0.04)',
		alpha8: 'rgba(16, 185, 129, 0.08)',
		alpha12: 'rgba(16, 185, 129, 0.12)',
		alpha30: 'rgba(16, 185, 129, 0.3)',
		alpha50: 'rgba(16, 185, 129, 0.5)'
	},
	info: {
		lightest: '#ECFDFF',
		light: '#03A9F4',
		main: '#0288D1',
		dark: '#01579B',
		darkest: '#164C63',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(2, 136, 209, 0.04)',
		alpha8: 'rgba(2, 136, 209, 0.08)',
		alpha12: 'rgba(2, 136, 209, 0.12)',
		alpha30: 'rgba(2, 136, 209, 0.3)',
		alpha50: 'rgba(2, 136, 209, 0.5)'
	},
	error: {
		lightest: '#FEF3F2',
		light: '#EF5350',
		main: '#D32F2F',
		dark: '#C62828',
		darkest: '#7A271A',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(211, 47, 47, 0.04)',
		alpha8: 'rgba(211, 47, 47, 0.08)',
		alpha12: 'rgba(211, 47, 47, 0.12)',
		alpha30: 'rgba(211, 47, 47, 0.3)',
		alpha50: 'rgba(211, 47, 47, 0.5)'
	},
	warning: {
		lightest: '#FFFAEB',
		light: '#FF9800',
		main: '#EF6C00',
		dark: '#EF6C00',
		darkest: '#7A2E0E',
		contrastText: '#FFFFFF',
		alpha4: 'rgba(239, 108, 0, 0.04)',
		alpha8: 'rgba(239, 108, 0, 0.08)',
		alpha12: 'rgba(239, 108, 0, 0.12)',
		alpha30: 'rgba(239, 108, 0, 0.3)',
		alpha50: 'rgba(239, 108, 0, 0.5)'
	}
};

export const NEUTRAL = {
	'50': '#F8F9FA',
	'100': '#F3F4F6',
	'200': '#E5E7EB',
	'300': '#D2D6DB',
	'400': '#9DA4AE',
	'500': '#6C737F',
	'600': '#4D5761',
	'700': '#2F3746',
	'800': '#1C2536',
	'900': '#111927'
};

/** 공용 회색(라이트) — React 의 grayToneRaw. 화면에서는 var(--gt-200) 처럼 변수로 쓴다 */
export const GRAY_TONE = {
	'100': '#fafafa',
	'150': '#f6f6f6',
	'200': '#f2f2f2',
	'300': '#ebebeb',
	'400': '#e1e1e1',
	'500': '#cfcfcf',
	'600': '#bdbdbd',
	'700': '#909090',
	D700: '#7e7e7e',
	D600: '#606060',
	D500: '#575757',
	D400: '#494949',
	D300: '#404040',
	D200: '#2d2d2d',
	D100: '#1b1b1b'
};

/** 공용 회색(다크) — 다크모드일 때 --gt-* 로 깐다 */
export const GRAY_TONE_DARK = {
	'100': '#202020',
	'150': '#232323',
	'200': '#262626',
	'300': '#2e2e2e',
	'400': '#383838',
	'500': '#6a6a6a',
	'600': '#7d7d7d',
	'700': '#9a9a9a',
	D700: '#9a9a9a',
	D600: '#adadad',
	D500: '#b5b5b5',
	D400: '#c2c2c2',
	D300: '#cccccc',
	D200: '#dddddd',
	D100: '#ebebeb'
};
