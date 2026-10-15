/**
 * 로그인한 관리자·권한 메뉴·알림 (React store/profileProvider 자리).
 *
 *   admin.profile / admin.menus / admin.notification   — 화면에서 바로 읽는다($state)
 *   loadAdmin(pathname)        레이아웃이 처음 한 번. 토큰이 없거나 거절이면 로그인 화면으로
 *   authCheckAndRedirect(ms, pathname)   권한 없는 주소면 첫 메뉴로(React doAuthCheckAndNaviateToFirstMenu)
 *   checkAuth(path)            그 주소 메뉴 권한이 있나
 *   logout()
 */
import { goto } from '$app/navigation';
import { api, endSession, getBrowserUid, isLoggedIn, saveToken } from '#lib/shared/api.ts';
import { ADMIN_LOGIN_PATH, LS } from './config.ts';
import { getMeta, globalMeta, setGlobalMeta, updateGlobalMeta } from './meta.svelte.ts';
import type { AdminMenu, AdminNotification, AdminProfile } from '#lib/shared/types/admin.ts';
import { toast, alert } from './ui/modal.svelte.ts';
import { SK } from '#lib/shared/storage-keys.ts';

export const admin = $state({
	profile: null as AdminProfile | null,
	menus: null as AdminMenu[] | null,
	notification: { items: [], checkedNotiIds: [] } as AdminNotification,
	metaVersion: 0,
	loaded: false
});

function rememberAndLogin() {
	try {
		localStorage.setItem(LS.loginPath, location.pathname + location.search || '');
	} catch {}
	location.replace(ADMIN_LOGIN_PATH);
}

/** 로컬 개발자 모드 — 본사(isOffice) 계정 + localStorage zbd:admin:isDeveloper=1 이면 화면 노출만 개발자로(권한은 BE 가 판정) */
function applyLocalDeveloperFlag(p: AdminProfile): AdminProfile {
	if (!p?.isOffice || p?.isDeveloper) return p;
	try {
		if (localStorage.getItem(SK.admin.isDeveloper) != '1') return p;
	} catch {
		return p;
	}
	return { ...p, isDeveloper: true };
}

/** 메뉴 트리 순서대로 첫 경로 */
export function getFirstAuthMenuPath(ms: AdminMenu[] | null) {
	let firstPath = '';
	const homeNav: { title: string; desc: string; sub: any } = { title: '', desc: '', sub: null };
	for (const m of ms || []) {
		if (!firstPath && m.path) firstPath = m.path;
		if (!homeNav.title) {
			homeNav.title = m.name;
			homeNav.desc = m.description || '';
		}
		for (const s of m.items || []) {
			if (!firstPath && s.path) {
				firstPath = s.path;
				homeNav.sub = { ...s };
			}
		}
	}
	return { homeNav, firstPath: firstPath || '/' };
}

/** 권한 없는 주소(또는 /admin) 면 첫 메뉴로 바꿔 간다. 마이(/admin/my)·업무관리 초대는 예외 */
export function authCheckAndRedirect(ms: AdminMenu[] | null, pathname: string) {
	if (pathname.indexOf('/admin/work-invite/') == 0) return;
	if (pathname == '/admin/my' || pathname.indexOf('/admin/my/') == 0) return;
	// 개발 서버의 부품 견본(/admin/dev-ui) — 메뉴에 없다
	if (import.meta.env.DEV && pathname.indexOf('/admin/dev-ui') == 0) return;
	if (!Array.isArray(ms) || ms.length == 0) return;
	let firstPath = '';
	let hasAuth = false;
	for (const m of ms) {
		if (!firstPath && m.path) firstPath = m.path;
		if (!hasAuth && m.path && pathname.indexOf(m.path) == 0) hasAuth = true;
		for (const s of m.items || []) {
			if (!firstPath && s.path) firstPath = s.path;
			if (!hasAuth && s.path && pathname.indexOf(s.path) == 0) hasAuth = true; // 빈 주소는 모든 주소에 맞으므로 뺀다
		}
	}
	if ((!hasAuth && pathname) || pathname == '/admin/' || pathname == '/admin' || pathname == '/') {
		goto(firstPath, { replaceState: true });
	}
}

export function checkAuth(path: string) {
	const ms = admin.menus;
	if (!ms) return true;
	return ms.some((m) => (m.path && path.indexOf(m.path) == 0) || (m.items || []).some((s) => s.path && path.indexOf(s.path) == 0));
}

/**
 * 특수 메뉴 「관리자 포스트」(adminPosts)·「사용자 포스트」(userPosts) — 하위 메뉴를 게시판 설정(POST_GROUPS)의 게시판 목록으로 채운다.
 * 메뉴를 받는 이 자리에서 채워야 권한 검사(authCheckAndRedirect·checkAuth)도 그 게시판 주소(/admin/board/{groupKey})를 안다. 권한은 그 메뉴의 것.
 * 사용 안 함(use:false)·메뉴에 숨김(hideInMenu — 메뉴 설정 화면의 스위치) 게시판은 뺀다. DB 에 둔 하위 메뉴는 쓰지 않는다(게시판 설정이 바뀌면 메뉴도 따라간다 — 다음 로그인·새로고침 때).
 */
export const POST_MENU_SPECIALS: Record<string, 'adminPostGroups' | 'userPostGroups'> = { adminPosts: 'adminPostGroups', userPosts: 'userPostGroups' };
function expandPostMenu(m: AdminMenu): AdminMenu {
	const key = POST_MENU_SPECIALS[m.special || ''];
	if (!key) return m;
	const groups: any[] = (getMeta('POST_GROUPS')?.[key] || []).filter((g: any) => g && g.groupKey && g.use !== false && !g.hideInMenu);
	return {
		...m,
		path: '',
		items: groups.map((g, i) => ({ menuId: m.menuId * 1000 + i + 1, name: g.name || g.groupKey, description: g.description || '', path: `/admin/board/${g.groupKey}`, auth: m.auth }))
	};
}

let loading: Promise<void> | null = null;

/** 그 주소(정확히 같은) 메뉴에 받은 권한 — 화면의 실행·고치기 버튼을 RW 일 때만(2026-10-09 대표). 서버도 같은 기준으로 막는다 */
export function menuAuthOf(path: string): '' | 'R' | 'RW' {
	if (admin.profile?.isDeveloper || admin.profile?.isSuperAdmin) return 'RW';
	let best: '' | 'R' | 'RW' = '';
	for (const m of admin.menus || [])
		for (const x of [m, ...(m.items || [])]) {
			if (x.path != path || !x.auth) continue;
			if (x.auth == 'RW') return 'RW';
			best = 'R';
		}
	return best;
}

export function loadAdmin(pathname: string) {
	if (loading) return loading;
	loading = (async () => {
		// 로그인 여부는 서버가 판단(세션 쿠키 — 옛 localStorage 토큰은 첫 요청이 쿠키로 옮긴다)
		const o = await api.get('/service/admin/common/menus', { headers: { adminId: String(admin.profile?.adminId || 0) } });
		const oo = o.ok ? o.data : null;
		if (!oo?.profile) return rememberAndLogin();
		if (oo.profile.token) {
			saveToken('admin', oo.profile.token); // 개발 주소에서만 온다(운영은 서버가 쿠키로)
			delete oo.profile.token;
		}
		if (oo.config && !oo.config?.SITE?.hostingId) oo.config.SITE = { ...(oo.config.SITE || {}), hostingId: oo.hostingId };
		setGlobalMeta(oo.config || {});
		localStorage.setItem(LS.hostingId, oo.hostingId || '');
		localStorage.setItem(LS.groupId, String(oo.profile.groupId));
		admin.profile = applyLocalDeveloperFlag({ ...oo.profile, browserUid: getBrowserUid() });
		// 읽기 권한뿐인 빈 묶음(경로·하위 없음)은 뺀다
		admin.menus = ((oo.admin?.menus || []) as AdminMenu[]).map(expandPostMenu).filter((m) => !(m.auth == 'R' && !m.path && !(m.items || []).length));
		admin.notification = oo.notification || { items: [], checkedNotiIds: [] };
		admin.loaded = true;
		authCheckAndRedirect(admin.menus, pathname);
		startRefresh();
	})();
	return loading;
}

let refreshTimer: any = null;
/** 10분마다 로그인 확인 + 메타 갱신. 거절이면 로그인 화면으로 */
export async function refreshToken() {
	if (!isLoggedIn('admin')) {
		clearInterval(refreshTimer);
		return;
	}
	const o = await api.get('/service/admin/common/menus?check=1', { headers: { adminId: String(admin.profile?.adminId || 0) } });
	if (!o.ok) {
		if (o.error.code == 'NETWORK' || o.error.code == 'ABORTED') return;
		localStorage.setItem(LS.loginPath, location.pathname);
		location.replace(ADMIN_LOGIN_PATH);
		return;
	}
	updateGlobalMeta(o.data?.config);
	admin.metaVersion++;
}
function startRefresh() {
	clearInterval(refreshTimer);
	refreshTimer = setInterval(refreshToken, 10 * 60 * 1000);
}

/** 프로필 일부 바꾸기(화면이 저장한 뒤) */
export function updateProfile(p: Partial<AdminProfile>) {
	if (admin.profile) admin.profile = { ...admin.profile, ...p };
}

export async function logout() {
	await endSession('admin');
	location.replace(ADMIN_LOGIN_PATH);
}

void globalMeta;
