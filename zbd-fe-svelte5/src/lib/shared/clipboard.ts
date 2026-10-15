/**
 * 클립보드 — http 주소에서도 복사가 되게(2026-10-06 대표).
 *
 * `navigator.clipboard` 는 https·localhost(안전한 주소)에서만 있다. 사내망 http://172.30.1.50:5000 같은 개발 주소로 열면
 * 없어서 화면의 복사 버튼이 전부 조용히 실패했다. 앱 시작 때(hooks.client) 없으면 옛 방식(글 상자 + execCommand('copy'))으로
 * 채워 둔다 — 복사 버튼마다 고치지 않아도 되고, 플러그인이 붙인 화면도 같이 된다. 안전한 주소에서는 아무것도 안 한다.
 */

/** copyFallback — 숨긴 글 상자에 넣고 골라 복사(누른 직후에 불러야 한다). 성공하면 true */
export function copyFallback(text: string): boolean {
	const ta = document.createElement('textarea');
	ta.value = text;
	ta.setAttribute('readonly', '');
	ta.style.cssText = 'position:fixed;top:0;left:0;width:1px;height:1px;opacity:0;pointer-events:none';
	const prev = document.activeElement as HTMLElement | null;
	document.body.appendChild(ta);
	ta.focus();
	ta.select();
	ta.setSelectionRange(0, text.length); // iOS
	let ok = false;
	try {
		ok = document.execCommand('copy');
	} catch {
		ok = false;
	}
	ta.remove();
	prev?.focus?.();
	return ok;
}

/** copyText — 복사(안전한 주소면 클립보드 API, 아니면 옛 방식). 성공하면 true */
export async function copyText(text: string): Promise<boolean> {
	if (window.isSecureContext && navigator.clipboard?.writeText) {
		try {
			await navigator.clipboard.writeText(text);
			return true;
		} catch {
			/* 권한 거절 등 — 옛 방식으로 */
		}
	}
	return copyFallback(text);
}

/** installClipboardFallback — 클립보드 API 가 없는 주소(http)에서 navigator.clipboard.writeText 를 옛 방식으로 채운다 */
export function installClipboardFallback() {
	if (typeof navigator === 'undefined' || (navigator as any).clipboard?.writeText) return;
	const shim = {
		writeText: (text: string) => (copyFallback(String(text ?? '')) ? Promise.resolve() : Promise.reject(new Error('복사하지 못했습니다'))),
		readText: () => Promise.reject(new Error('이 주소(http)에서는 붙여넣기를 읽을 수 없습니다')),
		write: () => Promise.reject(new Error('이 주소(http)에서는 그림 복사를 할 수 없습니다')),
		read: () => Promise.reject(new Error('이 주소(http)에서는 붙여넣기를 읽을 수 없습니다'))
	};
	try {
		Object.defineProperty(navigator, 'clipboard', { value: shim, configurable: true });
	} catch {
		/* 막힌 브라우저 — 그대로 */
	}
}
