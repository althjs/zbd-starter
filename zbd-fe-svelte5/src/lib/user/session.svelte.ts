/**
 * 사용자(zbd.kr) 로그인 상태 — 예전 userProfileProvider 자리.
 * 로그인 표시 쿠키(zbd_user_in) 가 있을 때만 `/service/user/profile` 을 부른다(없으면 손님 — 토큰은 HttpOnly 세션 쿠키라 화면이 못 본다).
 */
import { api, hasLegacyToken, isLoggedIn } from '#lib/shared/api.ts';
import { COOKIE_HOST, LOGIN_URL } from '#lib/shared/config.ts';
import { setCookie } from '#lib/shared/cookie.ts';
import { navigate } from '#lib/shared/navigate.ts';
import type { UserProfile } from '#lib/shared/types/member.ts';

export const session = $state({ profile: null as UserProfile | null, checked: false });

let loading: Promise<UserProfile | null> | null = null;

export function loadProfile() {
	if (session.checked) {
		return Promise.resolve(session.profile);
	}
	loading ||= (async () => {
		if (!isLoggedIn('user') && !hasLegacyToken('user')) {
			session.checked = true;
			return null;
		}
		const o = await api.get('/service/user/profile');
		session.profile = o.ok && o.data?.userId ? o.data : null;
		session.checked = true;
		return session.profile;
	})();
	return loading;
}

/** 돌아올 주소 쿠키 — 로그인 화면은 `?redirect=` 를 읽지 않는다(CLAUDE.md 「반복한 실수」) */
export const COOKIE_USER_LOGIN_PATH_KEY = '__user_login_path';

export function doLogin(replace = false, path = '') {
	const back = path ? new URL(path, location.href).href : location.href;
	setCookie(COOKIE_USER_LOGIN_PATH_KEY, back, { domain: COOKIE_HOST });
	// 서브도메인(미니홈피 {주소}.zbd.kr·talk) 은 대표 도메인 로그인 화면으로 — 세션 쿠키를 같이 쓰므로 로그인하고 돌아오면 로그인돼 있다
	const root = COOKIE_HOST.replace(/^\./, '');
	if (root && location.hostname != root && location.hostname.endsWith('.' + root)) {
		const to = `https://${root}${LOGIN_URL}`;
		return replace ? location.replace(to) : (location.href = to);
	}
	// 화면 안 이동(SPA) — 페이지째 이동하면 뒤로가기가 bfcache 로 옛 화면을 되살려 흰 화면이 됐다(2026-10-03 대표 신고)
	navigate(LOGIN_URL, { replace });
}
