import { installClipboardFallback } from '#lib/shared/clipboard.ts';
/**
 * 앱 시작 훅(브라우저).
 *
 * 설치 앱(PWA)으로 열렸으면 앱을 시작하기 전에 한 번 그릴 틈을 준다 — 앱 스크립트가 첫 화면을 만드는 동안(휴대폰에서 0.4~0.5초)
 * 브라우저가 한 번도 그리지 못해 안드로이드가 창을 투명하게 둬 바탕화면이 비쳤다(2026-10-03 대표). 틈을 주면 app.html 이 칠한
 * 시작 화면 색·덮개(#zbd-boot)가 먼저 그려진다. 웹(설치 앱 아님)은 기다리지 않는다.
 */
export async function init() {
	installClipboardFallback(); // http 주소에서도 복사 버튼이 되게(lib/shared/clipboard.ts)
	if (!(window as any).__zbdBoot) return;
	await new Promise<void>((r) => requestAnimationFrame(() => setTimeout(r, 0)));
}

/**
 * 화면 파일(JS)을 못 받았을 때 — 오류 화면 대신 **한 번만** 새로 읽는다(2026-10-06 대표, E2E C6).
 * 망이 잠깐 끊기거나 배포 뒤 옛 파일 이름을 찾을 때 「화면을 그리다 문제가 생겼어요」 가 떴다. 새로 읽으면 대개 그대로 열린다.
 * 무한 새로고침 방지: 새로 읽은 시각을 세션 저장소에 두고, 1분 안에 또 실패하면 새로 읽지 않고 오류 화면을 그대로 보인다.
 */
const CHUNK_RELOAD_KEY = 'zbd:app:chunkReloadAt';
const CHUNK_ERROR = /Failed to fetch dynamically imported module|Importing a module script failed|error loading dynamically imported module|Unable to preload CSS/i;

/** reloadOnce — 1분 안에 새로 읽은 적이 없을 때만 새로 읽는다. 새로 읽으면 true */
function reloadOnce(): boolean {
	try {
		const last = Number(sessionStorage.getItem(CHUNK_RELOAD_KEY) || 0);
		if (Date.now() - last < 60_000) return false;
		sessionStorage.setItem(CHUNK_RELOAD_KEY, String(Date.now()));
	} catch {
		return false; // 저장소를 못 쓰면(사생활 모드 등) 횟수를 셀 수 없으니 새로 읽지 않는다
	}
	location.reload();
	return true;
}

if (typeof window !== 'undefined') {
	// 미리 받기(modulepreload)가 실패했을 때 vite 가 보내는 신호
	window.addEventListener('vite:preloadError', (e) => {
		if (reloadOnce()) e.preventDefault();
	});
}

export function handleError({ error, message }: { error: unknown; message: string }) {
	const text = error instanceof Error ? error.message : String(error ?? '');
	if (CHUNK_ERROR.test(text)) reloadOnce();
	return { message };
}
