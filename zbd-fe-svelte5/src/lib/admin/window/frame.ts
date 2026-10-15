/**
 * 이 문서가 화면 위 창(WindowFrame 의 iframe, `data-zbd-win`) 안에서 열렸는지.
 * 창 안에서는 링크로 옮겨 다녀도(`?popup=1` 이 빠져도) 메뉴 없는 틀로 그리고, 창을 또 띄우면 맨 위 문서에 맡긴다.
 */
export const IN_WIN_FRAME: boolean = (() => {
	if (typeof window == 'undefined' || window.self === window.top) return false;
	try {
		return !!window.frameElement?.hasAttribute('data-zbd-win');
	} catch {
		return false;
	}
})();
