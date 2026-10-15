/**
 * 브라우저 id 쿠키 `zbd_uid`(5년) — 로그인 토큰이 묶이는 브라우저 표시. 없으면 여기서 만든다(OS·브라우저 글자 + 시각).
 * 방문 기록·웹로그는 스타터에 없다(방문 분석 플러그인).
 */
import { getBrowserUid } from '#lib/shared/api.ts';
import { COOKIE_HOST } from '#lib/shared/config.ts';
import { setCookie } from '#lib/shared/cookie.ts';

function makeUID() {
	const ua = navigator.userAgent;
	const os = ua.includes('Macintosh') ? 'M' : ua.includes('Windows') ? 'W' : ua.includes('iPhone') || ua.includes('iPad') ? 'I' : ua.includes('Android') ? 'A' : 'E';
	const br = ua.includes('Edg') ? 'E' : ua.includes('Firefox') ? 'F' : ua.includes('Whale') ? 'W' : ua.includes('Chrome') ? 'C' : ua.includes('Safari') ? 'S' : 'E';
	return os + br + Date.now().toString(36);
}

/** 브라우저 id — 없으면 만들어 쿠키에(5년) */
export function ensureUid(): [string, boolean] {
	let id = getBrowserUid();
	const isNew = !id;
	if (!id) id = makeUID();
	setCookie('zbd_uid', id, { domain: COOKIE_HOST, days: 1825 });
	return [id, isNew];
}
