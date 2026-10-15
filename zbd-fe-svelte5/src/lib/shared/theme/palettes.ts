/**
 * 대표색 팔레트 — React 판 fe-service theme/colors.ts 의 값 그대로(사이트 메타 design.theme 이름으로 고른다).
 * 자동 생성(2026-10-03) — 손으로 고치지 말고 원본 값이 바뀌면 다시 뽑는다.
 */
export type Palette = {
	lightest: string;
	light: string;
	main: string;
	dark: string;
	darkest: string;
	contrastText: string;
	alpha4: string;
	alpha8: string;
	alpha12: string;
	alpha30: string;
	alpha50: string;
};
export const PALETTES: Record<string, Palette> = {
	blue: {
		lightest: '#bbdefb',
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
export const SUCCESS = {
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
};
export const INFO = {
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
};
export const ERROR = {
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
};
export const WARNING = {
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
};
