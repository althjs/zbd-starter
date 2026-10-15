/**
 * 사이트 기본 데이터(사이트 메타·게시판 목록 등) — site.svelte.ts(loadSite) 가 받는 base 를 읽는 도우미.
 * 스타터에는 페이지 디자인(GNB·위젯) 이 없다 — 머리·바닥·홈은 고정 화면(lib/user/layout).
 */
import { site, loadSite } from '#lib/shared/site.svelte.ts';

/** 사이트 메타(siteMeta) — base 를 아직 못 받았으면 null */
export function siteMetaOf() {
	return site.base?.siteMeta || null;
}

export function ensureBase() {
	return loadSite();
}

export function cloneDeep<T>(o: T): T {
	return o == null ? o : JSON.parse(JSON.stringify(o));
}
