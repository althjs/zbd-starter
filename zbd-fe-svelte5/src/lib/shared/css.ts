/**
 * 운영자 CSS 를 문서에 넣고 빼기 — React 판 utils/index.ts applyCSS/clearCSS 와 같은 규칙.
 * id 도 같게 쓴다(`global-style`, `gnb_13`, `page_12`, `html_1` …) — `<style id="{id}_css">`.
 * 같은 id 를 다시 넣으면 앞 것을 지우고 맨 뒤에 붙인다(그래서 나중에 넣은 것이 이긴다).
 */
export function applyCSS(cssText: string, id: string) {
	if (typeof document == 'undefined') return;
	document.getElementById(`${id}_css`)?.remove();
	const el = document.createElement('style');
	el.id = `${id}_css`;
	el.setAttribute('type', 'text/css');
	el.innerText = (cssText || '').replaceAll('\r', ' ').replaceAll('\n', ' ');
	document.head.appendChild(el);
}

export function clearCSS(id: string) {
	if (typeof document == 'undefined') return;
	const el = document.getElementById(`${id}_css`);
	if (el && el.tagName == 'STYLE') el.remove();
}

/** 운영자 스크립트 실행(위젯·GNB 의 javascript 칸) — React 판처럼 전역에서 eval */
export function runScript(code: string, where = '') {
	if (!code) return;
	try {
		(0, eval)(code);
	} catch (e) {
		console.warn('[운영자 스크립트 오류]', where, e);
	}
}
