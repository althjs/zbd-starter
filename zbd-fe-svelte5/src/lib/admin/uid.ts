/**
 * 관리자 브라우저 id(쿠키 zbd_uid) — 없으면 만든다(React utils/tag getUID). 로그인 기기 확인이 이 값에 묶인다.
 * 읽기만은 공용 getBrowserUid(api.ts).
 */
import { getBrowserUid } from '#lib/shared/api.ts';
import { setCookie } from '#lib/shared/cookie.ts';
import { COOKIE_HOST } from '#lib/shared/config.ts';

function makeUID() {
	const ua = navigator.userAgent;
	const os = ua.includes('Macintosh') ? 'M' : ua.includes('Windows') ? 'W' : ua.includes('iPhone') || ua.includes('iPad') ? 'I' : ua.includes('Android') ? 'A' : 'E';
	const br = ua.includes('Edg') ? 'E' : ua.includes('Firefox') ? 'F' : ua.includes('Whale') ? 'W' : ua.includes('Chrome') ? 'C' : ua.includes('Safari') ? 'S' : 'E';
	return os + br + Date.now().toString(36);
}

export function ensureBrowserUid(): string {
	const id = getBrowserUid() || makeUID();
	// 대표 도메인(COOKIE_HOST) — 관리자 주소(backyard·chat·mail) 가 같은 브라우저 id 를 쓴다(세션 쿠키를 나눠 쓰므로). 개발은 지금 호스트
	setCookie('zbd_uid', id, { domain: COOKIE_HOST || location.hostname, days: 1825 });
	return id;
}
