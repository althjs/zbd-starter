/**
 * 관리자 화면 설정 — 값은 빌드 환경변수(VITE_*) 로 바꾼다. 비밀값은 두지 않는다.
 */
import { SK } from '#lib/shared/storage-keys.ts';

/** 관리자 화면을 열어도 되는 주소(CLAUDE.md 1) — 사용자 화면 검사와 같은 값이라 공용 설정에 둔다 */
import { ADMIN_HOSTS } from '#lib/shared/config.ts';
export { ADMIN_HOSTS };

/** 관리자 화면이 쓰는 업로드 주소 앞부분 — 같은 출처의 /upload(운영 backyard·개발 vite 프록시) */
export const ADMIN_IMAGE_PREFIX: string = import.meta.env.VITE_ADMIN_IMAGE_URL_PREFIX || '/upload';

/** 로그인 화면 */
export const ADMIN_LOGIN_PATH = '/login/admin-login';

/** localStorage 키 — 원본은 #lib/shared/storage-keys.ts(SK) 한 곳. 관리자 화면이 자주 쓰는 것만 짧은 이름으로 */
export const LS = {
	loginPath: SK.auth.adminLoginPath,
	login: SK.auth.adminLogin,
	meta: SK.admin.meta,
	hostingId: SK.admin.hostingId,
	groupId: SK.auth.adminGroupId,
	themeOpts: SK.admin.themeOpts,
	dark: SK.admin.darkMode,
	sideNavMode: SK.admin.sideNavMode,
	pageSize: SK.admin.pageSize,
	pageBorder: SK.admin.pageBorder
} as const;

/** 왼쪽 메뉴 폭 */
export const SIDE_NAV_WIDTH = 280;
export const SIDE_NAV_MINI_WIDTH = 88;
export const TOP_NAV_HEIGHT = 64;

/** 메뉴 아이콘 이미지 폴더 */
export const ADMIN_MENU_ICON_IMAGE_ID = 'admin_menu_icons';

/** 단독 앱(메뉴 없는 설치 앱) — 관리자 주소(「도메인」 설정 — 서버가 문서에 넣는다) 중 chat.{도메인} = 메신저만(PWA 채팅 앱 플러그인 — 배포가 넣는다) · mail.{도메인} = 웹메일만 · office.{도메인} = 웹오피스만(PWA 오피스 앱 플러그인) · photo.{도메인} = 포토샵만(PWA 포토샵 앱 플러그인).
 *  그 주소가 없으면 host·url 이 빈 값 — 「채팅 앱 설치」 버튼이 숨는다 */
export type SingleAppKey = 'messenger' | 'mail' | 'office' | 'photo';
export type SingleApp = { key: SingleAppKey; host: string; url: string; home: string; paths: string[]; name: string };
const appHost = (prefix: string) => ADMIN_HOSTS.find((h) => h.indexOf(prefix) == 0) || '';
/** 단독 앱 이름 — 관리자가 「환경 설정 › 설치 앱 이름」 에서 정한 이름(전역 메타 SITE.PWA_MANIFEST_* — 지난 접속 때 받아 둔 값), 없으면 기본 이름 */
function appName(key: string, base: string): string {
	try {
		const m = JSON.parse((typeof localStorage != 'undefined' && localStorage.getItem(LS.meta)) || '{}')?.SITE?.[key];
		return String(m?.name || m?.short_name || '').trim() || base;
	} catch {
		return base;
	}
}
/** 관리자 앱 이름(대시보드 「웹앱 설치」 등) — 「설치 앱 이름」 의 관리자 앱 값, 없으면 기본 */
export const ADMIN_APP_NAME = appName('PWA_MANIFEST_ADMIN', '관리자 앱');
const CHAT_HOST = appHost('chat.');
const MAIL_HOST = appHost('mail.');
const OFFICE_HOST = appHost('office.');
const PHOTO_HOST = appHost('photo.');
export const SINGLE_APPS: Record<SingleAppKey, SingleApp> = {
	messenger: {
		key: 'messenger',
		host: CHAT_HOST,
		url: CHAT_HOST ? `https://${CHAT_HOST}/admin/messenger` : '',
		home: '/admin/messenger',
		paths: ['/admin/messenger', '/admin/claude', '/login'],
		name: appName('PWA_MANIFEST_MESSENGER', '메신저')
	},
	mail: { key: 'mail', host: MAIL_HOST, url: MAIL_HOST ? `https://${MAIL_HOST}/admin/mail` : '', home: '/admin/mail', paths: ['/admin/mail', '/login'], name: appName('PWA_MANIFEST_MAIL', '메일') },
	office: {
		key: 'office',
		host: OFFICE_HOST,
		url: OFFICE_HOST ? `https://${OFFICE_HOST}/admin/office` : '',
		home: '/admin/office',
		paths: ['/admin/office', '/login'],
		name: appName('PWA_MANIFEST_OFFICE', '오피스')
	},
	photo: {
		key: 'photo',
		host: PHOTO_HOST,
		url: PHOTO_HOST ? `https://${PHOTO_HOST}/admin/photoshop` : '',
		home: '/admin/photoshop',
		paths: ['/admin/photoshop', '/login'],
		name: appName('PWA_MANIFEST_PHOTO', '포토샵')
	}
};
export const SINGLE_APP: SingleApp | null = typeof location == 'undefined' ? null : Object.values(SINGLE_APPS).find((a) => a.host && a.host == location.hostname) || null;

/** 호스트로 호스팅이 고정된 로그인(React pages/login 의 호스트별 고정 표 중 즈비드 몫) */
export const HOST_HOSTING: Record<string, string> = { 'chat.zbd.kr': 'ZBD', 'mail.zbd.kr': 'ZBD', 'backyard.zbd.kr': 'ZBD' };
