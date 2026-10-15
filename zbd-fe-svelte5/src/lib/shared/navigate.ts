/**
 * 주소 이동 — 이 앱에 그 화면이 있으면 goto, 아직 옮기지 않은 화면(/life/desktop 등)이면 페이지째 이동.
 * goto 는 없는 경로에 `navigation_route_missing` 을 던진다 — 그때는 같은 주소로 페이지째 이동(서버가 답한다).
 */
import { goto } from '$app/navigation';

export async function navigate(path: string, opts: { replace?: boolean } = {}) {
	try {
		await goto(path, { replaceState: !!opts.replace });
	} catch (e: any) {
		if (!String(e?.message || e).includes('navigation_route_missing') && !String(e?.message || e).includes('does not resolve to a route')) throw e;
		if (opts.replace) location.replace(path);
		else location.assign(path);
	}
}
