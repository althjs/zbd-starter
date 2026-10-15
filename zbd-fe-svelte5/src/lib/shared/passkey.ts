/**
 * 패스키(지문·얼굴) — 브라우저 WebAuthn 과 Go `/service/login/passkey`·`/service/{admin|user}/my/passkey` 를 잇는다(2026-10-04, 관리자 → 같은 날 회원).
 * 관리자·회원은 서버가 주소로 가른다(관리자 주소 = 관리자 열쇠). 개발(localhost)은 한 출처라 회원 쪽 요청에 `?realm=user` 를 붙인다.
 * 등록 정보는 response.getPublicKey()(SPKI) 로 보내 서버가 CBOR 를 풀지 않는다. 본인 확인(userVerification) 은 필수.
 */
import { api } from '#lib/shared/api.ts';

export type PasskeyRealm = 'admin' | 'user';
const myUrl = (r: PasskeyRealm) => `/service/${r}/my/passkey${r == 'user' ? '?realm=user' : ''}`;
const loginUrl = (r: PasskeyRealm) => `/service/login/passkey${r == 'user' ? '?realm=user' : ''}`;

/** 내 등록 기기 목록 */
export async function listPasskeys(r: PasskeyRealm) {
	const o = await api.get(myUrl(r));
	return o.ok ? ((o.data || []) as any[]) : [];
}

/** 등록 기기 지우기 */
export async function removePasskey(r: PasskeyRealm, id: number) {
	const o = await api.delete(`/service/${r}/my/passkey?id=${id}${r == 'user' ? '&realm=user' : ''}`);
	if (!o.ok) throw new Error(o.error.message || '지우지 못했습니다');
}

const b64 = (buf: ArrayBuffer | null | undefined) => {
	if (!buf) return '';
	let s = '';
	for (const c of new Uint8Array(buf)) s += String.fromCharCode(c);
	return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
};
const unb64 = (s: string) => {
	const t = atob(s.replace(/-/g, '+').replace(/_/g, '/') + '==='.slice((s.length + 3) % 4));
	return Uint8Array.from(t, (c) => c.charCodeAt(0));
};

/** 이 브라우저에서 패스키를 쓸 수 있나(https·localhost + WebAuthn) */
export const passkeySupported = () => typeof window != 'undefined' && !!window.PublicKeyCredential && !!navigator.credentials && window.isSecureContext;

/** 지문 인증기(기기 안 인증) 가 있나 — 없으면 휴대폰 QR 로 넘어간다 */
export async function platformAuthenticator() {
	try {
		return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
	} catch {
		return false;
	}
}

/** 사용자가 창을 닫았거나 시간이 지난 경우(조용히 넘긴다) */
export const cancelled = (e: any) => e?.name == 'NotAllowedError' || e?.name == 'AbortError';

/** 기기 이름(목록에서 알아보게) */
export function deviceName() {
	const ua = navigator.userAgent;
	const os = /iPhone/.test(ua) ? 'iPhone' : /iPad/.test(ua) ? 'iPad' : /Android/.test(ua) ? 'Android' : /Mac OS X/.test(ua) ? 'Mac' : /Windows/.test(ua) ? 'Windows' : '기기';
	const br = /SamsungBrowser/.test(ua) ? '삼성 인터넷' : /Edg\//.test(ua) ? 'Edge' : /CriOS|Chrome\//.test(ua) ? 'Chrome' : /Safari\//.test(ua) ? 'Safari' : /Firefox/.test(ua) ? 'Firefox' : '';
	const app = matchMedia('(display-mode: standalone)').matches ? ' 앱' : '';
	return `${os}${br ? ' · ' + br : ''}${app}`;
}

/** 등록 — 실패 시 throw(취소는 cancelled(e)) */
export async function registerPasskey(realm: PasskeyRealm) {
	const o = await api.post(myUrl(realm), { a: 'options' });
	if (!o.ok) throw new Error(o.error.message);
	const op = o.data;
	const cred = (await navigator.credentials.create({
		publicKey: {
			challenge: unb64(op.challenge),
			rp: { id: op.rpId, name: realm == 'admin' ? '즈비드 관리자' : '즈비드' },
			user: { id: unb64(op.userId), name: op.userName || op.displayName, displayName: op.displayName || op.userName },
			pubKeyCredParams: [
				{ type: 'public-key', alg: -7 },
				{ type: 'public-key', alg: -8 },
				{ type: 'public-key', alg: -257 }
			],
			authenticatorSelection: { residentKey: 'required', requireResidentKey: true, userVerification: 'required' },
			excludeCredentials: (op.exclude || []).map((x: any) => ({ type: 'public-key', id: unb64(x.id), transports: x.transports?.length ? x.transports : undefined })),
			attestation: 'none',
			timeout: op.timeout
		}
	})) as PublicKeyCredential | null;
	if (!cred) throw new Error('패스키를 만들지 못했습니다');
	const r = cred.response as AuthenticatorAttestationResponse;
	if (typeof r.getPublicKey != 'function' || typeof r.getAuthenticatorData != 'function') throw new Error('이 브라우저는 패스키 등록을 지원하지 않습니다 — 브라우저를 최신으로 바꿔 주세요');
	const s = await api.post(myUrl(realm), {
		a: 'register',
		id: b64(cred.rawId),
		clientDataJSON: b64(r.clientDataJSON),
		authenticatorData: b64(r.getAuthenticatorData()),
		publicKey: b64(r.getPublicKey()),
		publicKeyAlgorithm: r.getPublicKeyAlgorithm(),
		transports: r.getTransports?.() || [],
		name: deviceName()
	});
	if (!s.ok) throw new Error(s.error.message);
	return s.data as any[];
}

/** 로그인 — 성공하면 서버 응답 data({token?, isAdditionalInfoRequired?}), 실패 시 throw(취소는 cancelled(e)) */
export async function loginWithPasskey(realm: PasskeyRealm) {
	const o = await api.post(loginUrl(realm), { a: 'options' });
	if (!o.ok) throw new Error(o.error.message);
	const cred = (await navigator.credentials.get({
		publicKey: { challenge: unb64(o.data.challenge), rpId: o.data.rpId, userVerification: 'required', timeout: o.data.timeout }
	})) as PublicKeyCredential | null;
	if (!cred) throw new Error('패스키를 고르지 않았습니다');
	const r = cred.response as AuthenticatorAssertionResponse;
	const s = await api.post(loginUrl(realm), {
		a: 'verify',
		id: b64(cred.rawId),
		clientDataJSON: b64(r.clientDataJSON),
		authenticatorData: b64(r.authenticatorData),
		signature: b64(r.signature)
	});
	if (!s.ok) throw new Error(s.error.message);
	return s.data as { token?: string; isAdditionalInfoRequired?: boolean };
}
