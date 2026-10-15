/** 쿠키 읽기·쓰기 — js-cookie 대신 */

export function getCookie(name: string) {
	if (typeof document == 'undefined') {
		return '';
	}
	const m = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/[.$?*|{}()[\]\\/+^]/g, '\\$&') + '=([^;]*)'));
	return m ? decodeURIComponent(m[1]) : '';
}

export function setCookie(name: string, value: string, opts: { domain?: string; days?: number } = {}) {
	let s = `${name}=${encodeURIComponent(value)}; path=/`;
	if (opts.domain && opts.domain != 'localhost') {
		s += `; domain=${opts.domain}`;
	}
	if (opts.days) {
		s += `; expires=${new Date(Date.now() + opts.days * 864e5).toUTCString()}`;
	}
	document.cookie = s;
}

export function removeCookie(name: string, opts: { domain?: string } = {}) {
	let s = `${name}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
	if (opts.domain && opts.domain != 'localhost') s += `; domain=${opts.domain}`;
	document.cookie = s;
}
