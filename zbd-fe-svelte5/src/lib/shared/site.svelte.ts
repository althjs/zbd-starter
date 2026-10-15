/**
 * 사이트 기본 데이터(siteMeta·commonCodes·GNB …) — 사용자 화면(관리자는 admin/meta 를 쓴다).
 *
 * `/service/design/{호스팅}/base` 로 받는다 — 서버가 사용자 화면이 쓰는 칸만 골라 준다(Go sitepub 허용 목록).
 * 운영 HTML 은 데이터를 싣지 않고 `<link rel="preload" id="zbd-base" href="…/base?v=지문">` 만 건다(2026-10-03 — 소스보기에
 * 사이트 설정이 통째로 보였다). 그 주소를 그대로 받으면 미리 받아 둔 응답을 쓰고, 지문 주소라 다음 방문엔 캐시에서 나온다.
 */
import { api } from './api';
import { HOSTING_ID } from './config';

declare global {
	interface Window {
		Kakao?: any;
	}
}

export const site = $state({ base: null as any, ready: false });

let loading: Promise<any> | null = null;

export function loadSite() {
	if (site.ready) {
		return Promise.resolve(site.base);
	}
	loading ||= (async () => {
		const pre = takePreload('zbd-base');
		let data: any = null;
		if (pre) data = await pre;
		if (!data) {
			const o = await api.get(`/service/design/${HOSTING_ID}/base`);
			data = o.ok ? o.data : {};
		}
		site.base = data;
		site.ready = true;
		return site.base;
	})();
	return loading;
}

/**
 * 서버가 HTML 에 건 미리 받기(`<link rel="preload" as="fetch" id=…>`) 를 한 번 쓴다 — 같은 주소·같은 방식(fetch 기본값)으로 받아야
 * 브라우저가 미리 받은 응답을 내준다. 응답은 옛 /service 모양({code, data}) — data 만 돌려준다. 실패하면 null(부르는 쪽이 다시 받는다).
 */
export function takePreload(id: string): Promise<any> | null {
	if (typeof document == 'undefined') return null;
	const el = document.getElementById(id) as HTMLLinkElement | null;
	if (!el?.href) return null;
	el.removeAttribute('id'); // 한 번만
	return fetch(el.href)
		.then((r) => (r.ok ? r.json() : null))
		.then((j) => (j && j.code == '0000' ? j.data : null))
		.catch(() => null);
}

/** 앱(격리 실행)에 넘기는 사이트 데이터 — 예전 window.___zbd_data_model 자리(앱은 base.label 등을 읽는다) */
export function appDataModel() {
	return site.base ? { base: site.base } : null;
}

export function siteMeta() {
	return site.base?.siteMeta || {};
}
