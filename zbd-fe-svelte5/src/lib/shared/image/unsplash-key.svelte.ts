/**
 * Unsplash 공개용 Access Key — 사이트 설정 SITE.apiKeys.unsplashAccessKey(관리자 「환경 설정 › 관리자 페이지 › 기타정보」, 2026-10-04 — 그 전엔 코드에 박혀 있었다).
 * 관리자 화면은 admin/meta 가 메뉴 API 설정을 받을 때 채우고, 사용자 화면은 사이트 기본 데이터(siteMeta.apiKeys — Go sitepub 허용)에서 읽는다.
 * 비면 ImagePicker 가 Unsplash 탭을 숨긴다.
 */
import { site } from '../site.svelte.ts';

export const unsplashKeyState = $state({ admin: '' });

export function unsplashKey(): string {
	return unsplashKeyState.admin || site.base?.siteMeta?.apiKeys?.unsplashAccessKey || '';
}
