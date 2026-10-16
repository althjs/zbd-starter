/**
 * 개발 서버 전용 — 운영처럼 **BE SSR 이 만든 문서 머리말**을 그대로 받아 이 앱을 붙인다.
 *
 * 운영 zbd.kr 은 정적 index.html 이 없고 BE(`pages/api/seo/index.ts`) 가 요청마다 문서를 만든다:
 * 제목·SEO 메타·구조화 데이터, 사이트 기본 데이터(`window.___zbd_data_model`), 전역 HTML(useWaitingRoom 등).
 * 여기서도 같은 문서를 받아 React 묶음(`/assets/main-*.js`·`index-*.css`) 만 빼고 끼워 넣는다 —
 * 그래야 운영과 같은 조건(같은 데이터·같은 전역 CSS)에서 화면을 비교할 수 있다.
 *
 * 배포(정적 빌드)에서는 이 파일이 돌지 않는다. 운영 교체 때는 BE SSR 이 이 앱의 진입 스크립트를 쓰게 바꾼다.
 */
import { existsSync } from 'node:fs';
import type { Handle } from '@sveltejs/kit/hooks';

const UPSTREAM = process.env.ZBD_SSR || process.env.ZBD_API || 'http://127.0.0.1:5001';
/** 이 컴퓨터의 BE 는 nginx 가 없으니 운영 nginx 가 하던 일(경로 → /service/seo?p=)을 여기서 한다. 호스팅은 주소로 못 찾으니 HOSTING_ID 로 */
const LOCAL = /^http:\/\/(127\.0\.0\.1|localhost)(:|\/|$)/.test(UPSTREAM);
const HOSTING = process.env.VITE_HOSTING_ID || 'ZBD';
const docURL = (path: string) => (LOCAL ? `${UPSTREAM}/service/seo?p=${encodeURIComponent(path)}&hostingId=${HOSTING}` : UPSTREAM + path);

/** 개발에서 빼는 것 — 운영 화면 묶음(React 옛 main/index · Svelte _app), 방문 통계(gtag), 정의가 겹치는 charset·viewport. 서비스 워커는 둔다(설치 앱 신호가 운영과 같게) */
function cleanHead(head: string) {
	return (
		head
			.replace(/<script[^>]*type="module"[^>]*src="\/assets\/[^"]*"[^>]*><\/script>/g, '')
			.replace(/<link[^>]*href="\/assets\/index-[^"]*\.css"[^>]*>/g, '')
			// 운영 Svelte 빌드의 미리 받기·CSS 링크(2026-10-03 전환 뒤) — 개발 서버엔 없는 파일이라 404 가 줄줄이 났다
			.replace(/<link[^>]*href="\/assets\/_app\/[^"]*"[^>]*>/g, '')
			.replace(/<script async src="https:\/\/www\.googletagmanager\.com[^>]*><\/script>/g, '')
			.replace(/<script>\s*window\.dataLayer[\s\S]*?<\/script>/g, '')
			.replace(/<meta charset="[^"]*"\s*\/?>/i, '')
			.replace(/<meta name="viewport"[^>]*>/i, '')
	);
}

/** 본문에서는 사이트 데이터 스크립트(→ zbd:data) 와 검색엔진용 noscript(→ zbd:body) 만 가져온다 */
function pickBody(body: string) {
	const data = body.match(/<script>\s*window\.___zbd_data_model[\s\S]*?<\/script>/);
	const noscript = body.match(/<noscript>[\s\S]*?<\/noscript>/);
	return { data: data ? data[0] : '', noscript: noscript ? noscript[0] : '' };
}

type Doc = { at: number; head: string; data: string; noscript: string; redirect?: string };
const cache = new Map<string, Doc>();

async function upstreamDoc(path: string, ua: string): Promise<Doc> {
	const hit = cache.get(path);
	if (hit && Date.now() - hit.at < 20_000) return hit;
	try {
		const r = await fetch(docURL(path), { headers: { 'user-agent': ua || 'zbd-fe-svelte5-dev', accept: 'text/html' }, redirect: 'manual' });
		if (r.status >= 300 && r.status < 400) return { at: Date.now(), head: '', data: '', noscript: '', redirect: r.headers.get('location') || '' };
		const html = await r.text();
		const head = cleanHead(html.slice(html.indexOf('<head>') + 6, html.indexOf('</head>')));
		const doc = { at: Date.now(), head, ...pickBody(html.slice(html.indexOf('<body'))) };
		cache.set(path, doc);
		return doc;
	} catch {
		return { at: Date.now(), head: '', data: '', noscript: '' };
	}
}

/**
 * 외부 접속 주소(zbd tunnel)의 앱 주소 — 2026-10-16 대표 「모두 웹앱 설치 가능하게」.
 * 터널은 admin.·chat.·mail.·office.·photo.{사이트 주소} 도 이 개발 서버로 넘기고, 방문자가 연 주소(x-zbd-public-host)와
 * 사이트 주소(x-zbd-tunnel-base — 앱 주소 인증서가 준비됐을 때만)를 붙인다. 개발 서버는 한 주소(localhost) 모드라 서버가 관리자 주소 목록을
 * 문서에 넣지 않으니 여기서 넣는다: 관리자 주소 목록(zbd:admin-hosts → 단독 앱 알아보기) · 사이트 주소(zbd:site-base → 「사이트 보기」 링크) ·
 * 그 주소의 설치 정보 링크(같은 출처에 설치 앱 하나). 헤더를 위조해도 바뀌는 것은 이 개발 화면의 주소 표시뿐이라 모양만 확인한다.
 */
const TUNNEL_BASE = /^starter-[0-9]{1,10}\.zbd\.co\.kr$/;
const TUNNEL_APP: Record<string, string> = { admin: '', chat: 'messenger', mail: 'mail', office: 'office', photo: 'photo' };
function tunnelHead(headers: Headers): string {
	const base = (headers.get('x-zbd-tunnel-base') || '').toLowerCase();
	const pub = (headers.get('x-zbd-public-host') || '').toLowerCase();
	if (!TUNNEL_BASE.test(base)) return '';
	// 단독 앱은 그 플러그인(pwa-chat 등)을 붙였을 때만 — 붙이면 앱 아이콘이 저장소 db/assets/site 에 들어온다. 안 붙인 앱까지 넣으면 「설치」 버튼이 아이콘 없이 뜬다
	const hosts = Object.keys(TUNNEL_APP)
		.filter((a) => !TUNNEL_APP[a] || existsSync(`../db/assets/site/192x192_${TUNNEL_APP[a]}_icon.png`))
		.map((a) => a + '.' + base);
	let out = `<meta name="zbd:admin-hosts" content="${hosts.join(',')}"><meta name="zbd:site-base" content="https://${base}/">`;
	const app = pub.endsWith('.' + base) ? pub.slice(0, -base.length - 1) : '';
	if (app in TUNNEL_APP) out += `<link rel="manifest" href="/service/pwa/manifest?admin=1&${TUNNEL_APP[app] ? 'app=' + TUNNEL_APP[app] + '&' : ''}hostingId=${HOSTING}">`;
	return out;
}

export const handle: Handle = async ({ event, resolve }) => {
	// /sitemap.xml · /robots.txt — 운영 nginx 처럼 BE 로(개발 서버가 화면으로 그리지 않게)
	if (LOCAL && (event.url.pathname == '/sitemap.xml' || event.url.pathname == '/robots.txt')) {
		const u = event.url.pathname == '/robots.txt' ? `${UPSTREAM}/service/seo/robots?hostingId=${HOSTING}` : docURL('/sitemap.xml');
		const r = await fetch(u);
		return new Response(await r.text(), { status: r.status, headers: { 'content-type': r.headers.get('content-type') || 'text/plain' } });
	}
	const accept = event.request.headers.get('accept') || '';
	if (event.request.method != 'GET' || accept.indexOf('text/html') == -1) return resolve(event);
	// 관리자 화면(/admin·/login)은 사용자 사이트 머리말이 필요 없다 — 운영 zbd.kr 은 /admin 을 홈으로 넘기므로(2026-10-03) 받으러 가면 개발에서도 홈으로 튕긴다
	const tunnel = tunnelHead(event.request.headers);
	if (/^\/(admin|login)(\/|$)/.test(event.url.pathname)) return resolve(event, { transformPageChunk: ({ html }) => html.replace('<!--zbd:head-->', '<title>관리자</title>' + tunnel) });

	const doc = await upstreamDoc(event.url.pathname + event.url.search, event.request.headers.get('user-agent') || '');
	if (doc.redirect) {
		// 서브도메인·옛 주소 넘기기 등 BE 가 정한 이동은 그대로 따른다
		const to = doc.redirect.replace(/^https:\/\/zbd\.kr/, '');
		return new Response(null, { status: 302, headers: { location: to } });
	}
	return resolve(event, {
		transformPageChunk: ({ html }) =>
			html
				.replace('<!--zbd:head-->', (doc.head || '<title>내 사이트</title>') + tunnel)
				.replace('<!--zbd:data-->', doc.data)
				.replace('<!--zbd:body-->', doc.noscript)
	});
};
