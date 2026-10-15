/**
 * 관리자 전역 메타(GLOBAL_META) — 메뉴 API(`/service/admin/common/menus`) 의 config(SITE·COMMON·COMMON_CODES·POST_GROUPS …).
 * React 판 common-constants 의 getMeta·setGlobalMeta 자리. localStorage `zbd:admin:meta` 에 캐시해 첫 화면부터 쓴다.
 *
 *   getMeta('SITE', 'theme')            → 'skyblue'
 *   getMeta('SITE', 'feature', 'useESPost')
 * 값이 바뀌면($state) 읽는 화면도 따라 그린다.
 */
import { dark } from './dark.svelte.ts';
import { LS, ADMIN_IMAGE_PREFIX } from './config.ts';
import { siteBase } from '#lib/shared/config.ts';
import type { AdminProfile } from '#lib/shared/types/admin.ts';
import { unsplashKeyState } from '#lib/shared/image/unsplash-key.svelte.ts';

function load(): Record<string, any> {
	try {
		return JSON.parse((typeof localStorage != 'undefined' && localStorage.getItem(LS.meta)) || '{}') || {};
	} catch {
		return {};
	}
}

export const globalMeta = $state<{ v: Record<string, any> }>({ v: load() });

// 공용 부품(ImagePicker)이 쓰는 키 — 관리자 설정이 바뀔 때마다 넘겨 둔다
function syncShared() {
	unsplashKeyState.admin = globalMeta.v?.SITE?.apiKeys?.unsplashAccessKey || '';
}
syncShared();

/** 메뉴 API 로 새로 받은 config 를 통째로 넣는다(저장까지) */
export function setGlobalMeta(config: Record<string, any>) {
	globalMeta.v = { ...globalMeta.v, ...config };
	syncShared();
	try {
		localStorage.setItem(LS.meta, JSON.stringify(globalMeta.v));
	} catch {}
}

/** 10분 확인(check=1) 때 — 있는 키만 바꾼다(React updateGlobalMeta) */
export function updateGlobalMeta(config: Record<string, any> | null | undefined) {
	if (!config) return;
	const next = { ...globalMeta.v };
	for (const k of Object.keys(next)) next[k] = config[k] || next[k];
	globalMeta.v = next;
	syncShared();
}

export function getMeta(groupKey: string, key1 = '', key2 = ''): any {
	const g = globalMeta.v?.[groupKey];
	if (key2) return g?.[key1]?.[key2] ?? null;
	if (key1) return g?.[key1] || null;
	return g ?? null;
}

export const getHostingId = (): string => getMeta('SITE', 'hostingId') || '';
export const getCompanyName = (): string => getMeta('SITE', 'companyName') || '';

// 옛 이름 COMPANIES·항목의 companyId — 2026-10 옮기는 기간
/** 호스팅 목록 설정(HOSTINGS) — 없으면 옛 COMPANIES */
export const getHostingsMeta = (): Record<string, any> | null => getMeta('HOSTINGS') || getMeta('COMPANIES');
/** 호스팅 목록(HOSTINGS.items) — 옛 항목(companyId) 도 hostingId 로 맞춰 준다 */
export function getHostingItems(): any[] {
	return ((getHostingsMeta()?.items || []) as any[]).map((o) => (o && o.hostingId === undefined && o.companyId !== undefined ? { ...o, hostingId: o.companyId } : o));
}
/** 본문 배경색(사이트 메타) — 다크 모드면 pageBgColorDark. 없으면 null(테마 기본 --adm-bg) */
/** 사용자 사이트 주소(끝 / 포함) — SITE.homepage → 홈페이지 탭의 design.homepage → 지금 출처. 「홈페이지 오픈」·사이트맵 주소 등
 *  (비면 `board/…` 같은 상대 주소가 돼 지금 관리자 주소 뒤에 붙었다 — 스타터 기본 데이터는 design.homepage 에만 있다) */
export function siteHomepage(): string {
	return siteBase(getMeta('SITE', 'homepage') || getMeta('SITE', 'design', 'homepage'));
}

export const getPageBgColor = (): string | null => String(getMeta('SITE', dark.resolved == 'dark' ? 'pageBgColorDark' : 'pageBgColor') || '').trim() || null;
/** 기능 켜짐 — getMeta('SITE','feature',name) */
export const feature = (name: string): boolean => !!getMeta('SITE', 'feature', name);

/** 최고 관리자 — 그룹 1 · isSuperAdmin · isDeveloper · COMMON.masterGroupIds */
export function isSuperAdmin(p: Partial<AdminProfile> | null | undefined) {
	if (!p) return false;
	if (p.groupId == 1 || p.isSuperAdmin || p.isDeveloper) return true;
	return (getMeta('COMMON', 'masterGroupIds') || []).indexOf(p.groupId) != -1;
}

/** 업로드 이미지 주소 — `/upload/{호스팅}/{폴더}/{파일}` (React getImageUrlPrefix + 파일) */
export function adminImagePrefix(hostingId = '') {
	let cid = hostingId || (typeof localStorage != 'undefined' ? localStorage.getItem(LS.hostingId) : '') || '';
	if (cid == 'MASTER') cid = '';
	return cid ? `${ADMIN_IMAGE_PREFIX}/${cid}` : ADMIN_IMAGE_PREFIX;
}
export function adminImageUrl(folder: string, name?: string | null, hostingId = '') {
	return folder && name ? `${adminImagePrefix(hostingId)}/${folder}/${name}` : '';
}

/**
 * 테두리 없는 본문(React isIgnoreBorder) — `zbd:admin:pageBorder` 가 "0" 이 아니면 참.
 * 사이트 메타 disablePageBorder 가 있으면 그 값으로 저장해 둔다(React dashboard/layout 과 같은 순서).
 */
export function isIgnoreBorder() {
	try {
		const d = getMeta('SITE', 'disablePageBorder');
		if (d !== null) {
			if (d) localStorage.setItem(LS.pageBorder, '1');
			else localStorage.removeItem(LS.pageBorder);
		}
		return localStorage.getItem(LS.pageBorder) != '0';
	} catch {
		return true;
	}
}
