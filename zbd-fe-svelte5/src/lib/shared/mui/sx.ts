/**
 * sx — 예전 MUI `sx` 처럼 **클래스 하나짜리 규칙**을 만들어 준다.
 *
 * 왜 인라인 style 을 쓰지 않나: 인라인은 구체성이 가장 높아 **운영자 CSS 가 덮을 수 없다.**
 * React 판의 sx 는 클래스(.css-xxx, 구체성 0,1,0) 였고 운영자 CSS(GNB·전역) 가 그 값을 덮어 모양을 잡아 두었다
 * (예: 띠배너 높이 36 → 휴대폰 30). 같은 조건이 되도록 여기서도 클래스로 만든다.
 *
 * 규칙은 문서 앞쪽에 고정된 <style id="zbd-sx"> 하나에 쌓인다 → 나중에 붙는 운영자 CSS 가 같은 구체성에서 이긴다.
 *
 *   sx({ height: 36, background: '#1d433d', '&:hover': { color: 'red' }, '@media (max-width:899.95px)': { height: 30 } })
 *   → 'zx-1a2b3c'
 */
type SxValue = string | number | null | undefined | false;
export type Sx = { [k: string]: SxValue | Sx };

/** 숫자에 px 를 붙이지 않는 속성 */
const UNITLESS = new Set(['flexGrow', 'flexShrink', 'flex', 'opacity', 'zIndex', 'fontWeight', 'lineHeight', 'order', 'zoom', 'gridRow', 'gridColumn', 'aspectRatio']);

const kebab = (k: string) => (k.startsWith('--') ? k : k.replace(/[A-Z]/g, (m) => '-' + m.toLowerCase()));

function decls(o: Sx) {
	let out = '';
	for (const [k, v] of Object.entries(o)) {
		if (v === null || v === undefined || v === false || v === '' || typeof v == 'object') continue;
		const val = typeof v == 'number' && !UNITLESS.has(k) ? `${v}px` : String(v);
		out += `${kebab(k)}:${val};`;
	}
	return out;
}

function rules(sel: string, o: Sx): string[] {
	const list: string[] = [];
	const own = decls(o);
	if (own) list.push(`${sel}{${own}}`);
	for (const [k, v] of Object.entries(o)) {
		if (!v || typeof v != 'object') continue;
		if (k.startsWith('@')) {
			const inner = rules(sel, v as Sx);
			if (inner.length) list.push(`${k}{${inner.join('')}}`);
		} else {
			// '&:hover' · '& img' · '&.on'
			list.push(...rules(k.replace(/&/g, sel), v as Sx));
		}
	}
	return list;
}

function hash(s: string) {
	let h = 5381;
	for (let i = 0; i < s.length; i++) h = ((h << 5) + h + s.charCodeAt(i)) | 0;
	return (h >>> 0).toString(36);
}

let sheet: CSSStyleSheet | null = null;
const made = new Set<string>();

function getSheet() {
	if (sheet) return sheet;
	let el = document.getElementById('zbd-sx') as HTMLStyleElement | null;
	if (!el) {
		el = document.createElement('style');
		el.id = 'zbd-sx';
		// 문서 앞쪽(첫 운영자 CSS 보다 앞) 에 고정 — 운영자 CSS 가 같은 구체성에서 이기도록
		const firstOperator = document.head.querySelector('style[id$="_css"]');
		document.head.insertBefore(el, firstOperator);
	}
	sheet = el.sheet!;
	return sheet;
}

/** sx 객체 → 클래스 이름. 같은 내용이면 같은 클래스(한 번만 만든다) */
export function sx(o: Sx | null | undefined | false): string {
	if (!o || typeof document == 'undefined') return '';
	const body = JSON.stringify(o);
	if (body == '{}') return '';
	const cls = 'zx-' + hash(body);
	if (!made.has(cls)) {
		made.add(cls);
		const s = getSheet();
		for (const r of rules('.' + cls, o)) {
			try {
				s.insertRule(r, s.cssRules.length);
			} catch (e) {
				console.warn('[sx] 규칙 오류', r);
			}
		}
	}
	return cls;
}

/** 클래스 이름 합치기 — 빈 값은 건너뛴다 */
export function cx(...list: (string | false | null | undefined)[]) {
	return list.filter(Boolean).join(' ');
}

/** 화면 폭 구간 (MUI 테마 breakpoints — fe-service theme/index.ts) */
export const BP = { sm: 600, md: 900, lg: 1280, xl: 1440 };
export const up = (k: keyof typeof BP) => `@media (min-width:${BP[k]}px)`;
export const down = (k: keyof typeof BP) => `@media (max-width:${BP[k] - 0.05}px)`;
