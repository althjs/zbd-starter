/**
 * 앱 전체 설정 — 관리자·사용자 화면이 같이 쓴다.
 * 값은 빌드 환경변수(VITE_*) 로 바꾼다. 비밀값은 여기에 두지 않는다(브라우저로 나간다).
 */

/** 호스팅(옛 이름 회사·companyId) — 즈비드 전용 앱이라 기본 ZBD */
export const HOSTING_ID: string = import.meta.env.VITE_HOSTING_ID || 'ZBD';

/**
 * DB json 칸 안에 저장된 이미지·첨부·설정 객체의 호스팅 id — 다시 저장되기 전까지는 옛 이름(companyId)일 수 있어 둘 다 본다.
 * 옛 이름 companyId — 2026-10 옮기는 기간
 */
export function hostingIdOf(o: any): string | undefined {
	return o?.hostingId ?? o?.companyId;
}

/**
 * 게시판 설정(POST_GROUPS 항목) 의 글 저장 호스팅(customHostingId).
 * 옛 이름 customCompanyId — 2026-10 옮기는 기간
 */
export function customHostingIdOf(pg: any): string | undefined {
	return pg?.customHostingId ?? pg?.customCompanyId;
}

/** 업로드 이미지 주소 앞부분(호스팅 폴더 앞까지) */
export const IMAGE_URL_PREFIX: string = import.meta.env.VITE_IMAGE_URL_PREFIX || '/upload';

/**
 * 로그인 쿠키 도메인(예 `.zbd.kr`) — 브라우저 id(zbd_uid)·돌아갈 주소 쿠키를 여기에 심는다(서버 세션 쿠키와 같은 범위).
 * 서버가 문서에 넣는 SITE 설정 cookieHost(`<meta name="zbd:cookie-host">` — 「도메인」 의 대표 사용자 주소로 계산). 개발은 빈 값(지금 호스트).
 * 지금 주소가 그 범위 밖(사용자 주소 아래가 아닌 관리자 주소 — 예 admin-example.com)이면 빈 값 = 이 주소 전용.
 * 브라우저는 범위 밖 domain 쿠키를 조용히 버린다 — 그래서 처음 온 브라우저는 브라우저 id 가 없어 관리자 로그인이 풀렸다(2026-10-06 E2E 로 발견).
 */
function cookieHostHere(): string {
	if (typeof document == 'undefined') return '';
	const v = document.querySelector('meta[name="zbd:cookie-host"]')?.getAttribute('content') || '';
	const root = v.replace(/^\./, '');
	const host = location.hostname;
	if (!v && TUNNEL_BASE) return '.' + TUNNEL_BASE;
	return root && (host == root || host.endsWith('.' + root)) ? v : '';
}

/**
 * 외부 접속 주소(zbd tunnel)의 사이트 주소(예 starter-76.zbd.co.kr) — 지금 주소가 그 주소이거나 그 아래 앱 주소(admin.·chat.·mail.…)일 때만.
 * 개발 서버가 문서에 넣는다(hooks.server.ts tunnelHead). 앱 주소끼리 로그인을 나눠 쓰려고 이 주소에서는 서버 세션 쿠키(운영과 같은 HttpOnly)·브라우저 id 를 이 주소에 심는다
 * (2026-10-16 대표 — 메신저 주소로 가면 다시 로그인했다). 범위가 이 고객 주소 아래뿐이라 다른 고객(starter-11…)으로는 가지 않는다.
 */
function tunnelBaseHere(): string {
	if (typeof document == 'undefined') return '';
	const m = (document.querySelector('meta[name="zbd:site-base"]')?.getAttribute('content') || '').match(/^https:\/\/(starter-[0-9]{1,10}\.zbd\.co\.kr)\/$/);
	const host = location.hostname;
	return m && (host == m[1] || host.endsWith('.' + m[1])) ? m[1] : '';
}
export const TUNNEL_BASE: string = tunnelBaseHere();
export const COOKIE_HOST: string = cookieHostHere();

/** 로그인 화면 */
export const LOGIN_URL: string = import.meta.env.VITE_LOGIN_URL || '/login';

/** `{prefix}/{호스팅}` — 예전 getImageUrlPrefix() */
export function imagePrefix(hostingId = HOSTING_ID) {
	return hostingId ? `${IMAGE_URL_PREFIX}/${hostingId}` : IMAGE_URL_PREFIX;
}

/** 업로드 이미지 주소 — 폴더·파일명이 없으면 빈 문자열 */
export function imageUrl(folder: string, name?: string | null) {
	return folder && name ? `${imagePrefix()}/${folder}/${name}` : '';
}

const metaContent = (name: string) => (typeof document != 'undefined' && document.querySelector(`meta[name="${name}"]`)?.getAttribute('content')) || '';

/**
 * 이 주소가 관리자(admin)·사용자(user) 어느 쪽인가(CLAUDE.md 1) — 관리자 화면은 admin 에서만, 사용자 화면은 user 에서만 연다.
 * 값은 서버가 문서에 넣는다(`<meta name="zbd:host-realm">` — SITE 설정 domains, 관리자 「환경 설정 › 도메인」 한 곳). 2026-10-06 대표 「하나로만」 — 빌드 값(VITE_ADMIN_HOSTS)은 없앴다.
 * 개발 한 주소(localhost)는 서버가 싣지 않는다 = 빈 값 = 검사 끔.
 */
export const HOST_REALM: '' | 'admin' | 'user' = ((v) => (v == 'admin' || v == 'user' ? v : ''))(metaContent('zbd:host-realm'));

/** 관리자 주소 목록(관리자 주소의 문서에만 온다 — `<meta name="zbd:admin-hosts">`). 첫 번째가 대표 관리자 주소(브라우저 id 를 합칠 때 이 값이 이긴다) */
export const ADMIN_HOSTS: string[] = metaContent('zbd:admin-hosts')
	.split(',')
	.map((s) => s.trim())
	.filter(Boolean);

/** 관리자 로고 주소(「관리자 페이지」 설정 SITE.logo — 관리자 주소의 문서에만 온다 `<meta name="zbd:admin-logo">`). 로그인 전 화면용 — 없으면 '' */
export const ADMIN_LOGO: string = metaContent('zbd:admin-logo');

/** 한 주소 개발 — run_dev(localhost·내부망 IP·외부 접속 주소 터널 *.zbd.co.kr). 서버가 주소 구분(zbd:host-realm)을 싣지 않아 사용자·관리자 화면이 같은 출처에 있다 */
export const SINGLE_ORIGIN = HOST_REALM == '';
/** 사용자 사이트 주소(끝 /) — 넘긴 설정 주소(SITE homepage·도메인 설정). 한 주소 개발이면 늘 지금 출처
 *  (설정은 운영 도메인이라 터널·localhost 의 관리자에서 누르면 운영 사이트로 나갔다 — 2026-10-08). 관리자 → 사용자 화면 링크·미리보기 주소는 모두 이것으로 */
export function siteBase(configured?: string | null): string {
	const here = typeof location != 'undefined' ? location.origin + '/' : '/';
	// 외부 접속 주소(터널)의 앱 주소(admin.starter-N…) — 개발 서버가 사이트 주소를 넣어 준다(hooks.server.ts tunnelHead)
	const tunnel = metaContent('zbd:site-base');
	if (tunnel) return tunnel;
	const v = SINGLE_ORIGIN ? '' : String(configured || '').trim();
	return v ? (v.endsWith('/') ? v : v + '/') : here;
}
