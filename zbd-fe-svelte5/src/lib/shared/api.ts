/**
 * API 호출 — 관리자·사용자 화면이 같이 쓴다. **응답은 새 형식 하나로만** 화면에 넘긴다(CLAUDE.md 3).
 *
 *   성공  { ok: true,  data, meta }
 *   실패  { ok: false, error: { code, message }, meta }
 *
 * 성공·실패 판단은 여기 한 곳에서. 화면은 `code == "0000"` 을 보지 않는다 — `res.ok`·`res.error.code` 만.
 * 옛 `/service/…` 응답(`{code:"0000", data, …}`, 실패도 HTTP 200) 은 여기서 새 모양으로 바꾼다.
 * Go 백엔드의 `/api/v1/…` 는 이미 새 형식이라 HTTP 상태로 판단한다.
 *
 * 인증 = **서버가 심는 HttpOnly 세션 쿠키**(2026-10-03 대표 — 대표 도메인 쿠키라 zbd.kr·talk·{주소}.zbd.kr / backyard·chat·mail 이 로그인을 나눠 쓴다).
 *   화면 스크립트는 토큰을 읽지도 저장하지도 않는다(localStorage 에 토큰 없음). 로그인 여부는 표시 쿠키 `zbd_user_in`·`zbd_admin_in`(값 1).
 *   회원은 헤더 `zbduid`(쿠키 `zbd_uid`) 를 함께 — 토큰이 이 브라우저 id 에 묶여 있다.
 *   개발(vite, localhost·dev.zbd.kr) 만 예전처럼 localStorage 토큰 + 헤더(쿠키를 쓸 수 없는 주소).
 *   예전에 localStorage 에 남은 토큰은 첫 요청 때 서버에 보내 쿠키로 바꾸고 지운다(migrateSession — 옮기는 기간).
 * 어느 쪽인지(사용자/관리자)는 화면 묶음(route group) 이 `setRealm` 으로 정한다. 서버는 주소로 쿠키를 고른다.
 */
import { ADMIN_HOSTS, COOKIE_HOST, TUNNEL_BASE } from './config';
import { getCookie, removeCookie, setCookie } from './cookie';
import { SK } from './storage-keys';

export type ApiMeta = { requestId?: string; total?: number; [k: string]: any };
export type ApiError = { code: string; message: string };
export type ApiRes<T = any> = { ok: true; data: T; meta: ApiMeta; error?: undefined } | { ok: false; data?: undefined; meta: ApiMeta; error: ApiError };
export type Realm = 'user' | 'admin';

export const TOKEN_KEYS: Record<Realm, string> = {
	user: SK.auth.user,
	admin: SK.auth.admin
};
export const BROWSER_UID_COOKIE = 'zbd_uid';
// 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
const OLD_BROWSER_UID_COOKIE = 'ox_uid';

/**
 * 브라우저 id(쿠키 zbd_uid) — 회원 토큰이 이 값을 품고 있다. 여기서는 만들지 않는다(기존 화면이 만든다).
 * 옛 쿠키만 있으면 **같은 값**을 새 이름으로 옮긴다(로그인 유지). 쿠키 범위는 기존 화면과 같게 —
 * 사용자·관리자 모두 대표 도메인(COOKIE_HOST), 5년.
 */
export function getBrowserUid(): string {
	shareUid();
	const id = getCookie(BROWSER_UID_COOKIE);
	if (id) return id;
	// 옮기는 기간(2026-10): 옛 이름 oxuid/ox_uid 도 받는다 — 지울 때 이 표시를 찾는다
	const old = getCookie(OLD_BROWSER_UID_COOKIE);
	if (old && typeof location != 'undefined') {
		const domain = COOKIE_HOST || location.hostname;
		setCookie(BROWSER_UID_COOKIE, old, { domain, days: 1825 });
	}
	return old;
}

/**
 * 브라우저 id 를 대표 도메인(COOKIE_HOST) 쿠키 하나로(2026-10-03 세션 쿠키) — 예전엔 주소마다(zbd.kr·{주소}.zbd.kr / backyard·chat·mail) 따로 심겨
 * 세션 쿠키를 나눠 써도 토큰(브라우저 id 에 묶임)이 다른 주소에서 맞지 않았다. 한 페이지에 한 번:
 *   이 주소에만 걸린 쿠키를 지운다 → 대표 주소(사용자 = 대표 도메인 / 관리자 = ADMIN_HOSTS 첫 번째) 는 자기 값을 대표 도메인으로 올리고(로그인이 그 값에 묶여 있다),
 *   다른 주소는 대표 도메인에 이미 있는 값을 쓴다(없으면 새로 만들 때 대표 도메인에 심긴다).
 * 옮기는 기간(2026-10) — 지울 때 이 표시를 찾는다
 */
let uidShared = false;
function readAll(name: string) {
	return document.cookie
		.split('; ')
		.filter((x) => x.startsWith(name + '='))
		.map((x) => decodeURIComponent(x.slice(name.length + 1)));
}
function shareUid() {
	if (uidShared || !COOKIE_HOST || typeof document == 'undefined') return;
	uidShared = true;
	const root = COOKIE_HOST.replace(/^\./, '');
	const host = location.hostname;
	if (host != root && !host.endsWith('.' + root)) return;
	const before = readAll(BROWSER_UID_COOKIE);
	if (!before.length) return;
	removeCookie(BROWSER_UID_COOKIE, { domain: host }); // 옛 관리자 화면이 심던 domain=지금 주소
	document.cookie = `${BROWSER_UID_COOKIE}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`; // 주소 전용(domain 없음)
	const shared = readAll(BROWSER_UID_COOKIE)[0] || '';
	const own = before.find((v) => v != shared) || '';
	const primary = host == root || host == ADMIN_HOSTS[0];
	const id = (primary ? own || shared : shared || own) || '';
	if (id) setCookie(BROWSER_UID_COOKIE, id, { domain: COOKIE_HOST, days: 1825 });
}

let realm: Realm = 'user';
/** 401(로그인 만료) 일 때 부를 곳 — 화면 묶음이 정한다(사용자 = 로그인 화면으로) */
let onUnauthorized: (() => void) | null = null;

export function setRealm(r: Realm, unauthorized?: () => void) {
	realm = r;
	if (unauthorized) onUnauthorized = unauthorized;
}

export function getRealm() {
	return realm;
}

/** 세션 쿠키로 로그인한다(운영 빌드). 개발 서버는 localStorage 토큰 — 외부 접속 주소(터널)로 연 개발 화면은 운영처럼 쿠키
 *  (서버가 방문자가 연 주소로 쿠키를 심는다 — Go web/session_tunnel.go, 2026-10-16 대표 「운영처럼 서버 쿠키로」) */
export const COOKIE_SESSION = !import.meta.env.DEV || !!TUNNEL_BASE;
const MARK_COOKIE: Record<Realm, string> = { user: 'zbd_user_in', admin: 'zbd_admin_in' };

/** 헤더에 실을 토큰 — 운영(세션 쿠키) 에서는 늘 ""(화면은 토큰을 갖지 않는다) */
export function getToken(r: Realm = realm) {
	if (COOKIE_SESSION || typeof localStorage == 'undefined') return '';
	return localStorage.getItem(TOKEN_KEYS[r]) || '';
}

/** 로그인돼 있는가(서버에 묻지 않고) — 운영은 표시 쿠키, 개발은 토큰 */
export function isLoggedIn(r: Realm = realm) {
	return COOKIE_SESSION ? getCookie(MARK_COOKIE[r]) == '1' : !!getToken(r);
}

/** 서버가 본문으로 준 토큰(개발 주소에서만 온다) 을 저장 */
export function saveToken(r: Realm, token?: string) {
	if (!COOKIE_SESSION && token && typeof localStorage != 'undefined') localStorage.setItem(TOKEN_KEYS[r], token);
}

/** 로그아웃 — 서버가 세션 쿠키를 지운다(HttpOnly 라 화면이 못 지운다) */
export async function endSession(r: Realm = realm) {
	await request('GET', '/service/login/logout').catch(() => null);
	if (typeof localStorage != 'undefined') localStorage.removeItem(TOKEN_KEYS[r]);
}

// 옮기는 기간(2026-10): localStorage 에 남은 옛 토큰 → 세션 쿠키. 지울 때 이 표시를 찾는다
const OLD_TOKEN_KEYS: Record<Realm, string[]> = { user: [SK.auth.user, '__user_token'], admin: [SK.auth.admin, '__admin_token'] };
/** 쿠키로 옮기지 않은 옛 토큰이 남아 있나(그러면 첫 요청이 옮긴다) */
export function hasLegacyToken(r: Realm = realm) {
	try {
		return COOKIE_SESSION && OLD_TOKEN_KEYS[r].some((k) => !!localStorage.getItem(k));
	} catch {
		return false;
	}
}
let migrating: Promise<void> | null = null;
function migrateSession(): Promise<void> {
	if (!COOKIE_SESSION || typeof localStorage == 'undefined') return Promise.resolve();
	return (migrating ||= (async () => {
		const keys = [...OLD_TOKEN_KEYS.user, ...OLD_TOKEN_KEYS.admin];
		const token = OLD_TOKEN_KEYS[realm].map((k) => localStorage.getItem(k)).find(Boolean);
		if (token && !isLoggedIn(realm)) {
			const uid = getBrowserUid();
			await fetch(API_BASE + '/service/login/session', {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}`, ...(uid ? { zbduid: uid } : {}) },
				body: '{}'
			}).catch(() => null);
		}
		for (const k of keys) localStorage.removeItem(k);
	})());
}

/** API 주소 앞부분 — 개발은 vite 프록시(/service), 배포는 같은 출처 */
const API_BASE: string = import.meta.env.VITE_API_BASE || '';

type Opts = { signal?: AbortSignal; blob?: boolean; headers?: Record<string, string> };

/** 옛 `/service` 응답 → 새 모양. 목록 API 의 totalCount 등 data 밖의 칸은 meta 로 */
function fromLegacy(json: any): ApiRes {
	if (!json || typeof json != 'object' || !('code' in json)) {
		return { ok: true, data: json, meta: {} };
	}
	const { code, message, data, meta: inner, status, ...rest } = json;
	// 옛 응답은 meta 를 안쪽에 따로 두기도 한다(목록: { meta: { totalCount, summary, fixedPosts } })
	const meta: ApiMeta = { ...rest, ...(inner && typeof inner == 'object' ? inner : {}) };
	if (typeof meta.totalCount == 'number') meta.total = meta.totalCount;
	if (code == '0000') return { ok: true, data, meta };
	return { ok: false, error: { code: String(code), message: message || '요청을 처리하지 못했습니다' }, meta };
}

/** 주소 고치기 자리 — 화면 한 곳이 자기 API 주소에 값을 붙일 때(웹메일 = 지금 고른 메일 계정 `acc`). 주소를 그대로 돌려주면 손대지 않는다 */
const urlHooks: ((url: string) => string)[] = [];
export function addUrlHook(f: (url: string) => string) {
	if (!urlHooks.includes(f)) urlHooks.push(f);
}

async function request<T>(method: string, url: string, body?: unknown, opts: Opts = {}): Promise<ApiRes<T>> {
	await migrateSession();
	for (const f of urlHooks) url = f(url);
	const headers: Record<string, string> = { ...(opts.headers || {}) };
	const token = getToken();
	if (token) headers['Authorization'] = `Bearer ${token}`;
	const uid = getBrowserUid();
	if (uid) headers['zbduid'] = uid;
	// 개발 서버는 쿠키를 못 받는다(localhost) — 로그인 응답의 토큰을 본문으로 달라고 한다(서버 session.go IssueSession)
	if (!COOKIE_SESSION) headers['X-Token-Delivery'] = 'body';
	const isForm = typeof FormData != 'undefined' && body instanceof FormData;
	if (body !== undefined && !isForm) headers['Content-Type'] = 'application/json';
	let res: Response;
	try {
		res = await fetch(API_BASE + url, {
			method,
			headers,
			credentials: 'include',
			signal: opts.signal,
			body: body === undefined ? undefined : isForm ? (body as FormData) : JSON.stringify(body)
		});
	} catch (e: any) {
		if (e?.name == 'AbortError') return { ok: false, error: { code: 'ABORTED', message: '취소됨' }, meta: {} };
		return { ok: false, error: { code: 'NETWORK', message: '네트워크 오류' }, meta: {} };
	}
	if (res.status == 401) onUnauthorized?.();
	if (opts.blob) {
		if (!res.ok) return { ok: false, error: { code: `HTTP_${res.status}`, message: `HTTP ${res.status}` }, meta: {} };
		return { ok: true, data: (await res.blob()) as T, meta: {} };
	}
	const json = await res.json().catch(() => null);
	// 새 형식(/api/v1) — HTTP 상태로 판단
	if (json && (('data' in json && 'meta' in json) || 'error' in json) && !('code' in json)) {
		return res.ok ? { ok: true, data: json.data, meta: json.meta || {} } : { ok: false, error: json.error || { code: `HTTP_${res.status}`, message: `HTTP ${res.status}` }, meta: json.meta || {} };
	}
	if (!res.ok && !json) return { ok: false, error: { code: `HTTP_${res.status}`, message: `HTTP ${res.status}` }, meta: {} };
	return fromLegacy(json) as ApiRes<T>;
}

export const api = {
	get: <T = any>(url: string, opts?: Opts) => request<T>('GET', url, undefined, opts),
	post: <T = any>(url: string, body: unknown = {}, opts?: Opts) => request<T>('POST', url, body, opts),
	put: <T = any>(url: string, body: unknown = {}, opts?: Opts) => request<T>('PUT', url, body, opts),
	delete: <T = any>(url: string, opts?: Opts) => request<T>('DELETE', url, undefined, opts),
	/** 파일 받기 — data 가 Blob */
	blob: (url: string, opts?: Opts) => request<Blob>('GET', url, undefined, { ...opts, blob: true })
};
