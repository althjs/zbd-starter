/**
 * 바뀐 내용 미리보기(React utils/jsondiffpatch) — jsondiffpatch 0.4.1 을 React 판과 같은 파일로(index.html 이 싣던 것을 jsdelivr 에서 필요할 때).
 *   const html = await diffHtml(원본, 지금)   // 같으면 ''
 * 공용 후보: 메타·게시판·디자인 등 「변경사항 저장」 화면 여럿이 쓴다.
 */
const JS = 'https://cdn.jsdelivr.net/npm/jsondiffpatch@0.4.1/dist/jsondiffpatch.umd.min.js';
const CSS = 'https://cdn.jsdelivr.net/npm/jsondiffpatch@0.4.1/dist/formatters-styles/html.css';
let loading: Promise<any> | null = null;

export function loadDiffLib(): Promise<any> {
	const w = window as any;
	if (w.jsondiffpatch) return Promise.resolve(w.jsondiffpatch);
	if (loading) return loading;
	if (!document.querySelector(`link[href="${CSS}"]`)) {
		const l = document.createElement('link');
		l.rel = 'stylesheet';
		l.href = CSS;
		document.head.appendChild(l);
	}
	loading = new Promise((resolve) => {
		const s = document.createElement('script');
		s.src = JS;
		s.onload = () => resolve(w.jsondiffpatch);
		s.onerror = () => ((loading = null), resolve(null));
		document.head.appendChild(s);
	});
	return loading;
}

const safe = (v: any) => {
	try {
		return JSON.parse(JSON.stringify(v ?? {}));
	} catch {
		return {};
	}
};

/** 바뀌었으면 html(없으면 '다름'), 같으면 '' */
export async function diffHtml(left: any, right: any): Promise<string> {
	const L = safe(left),
		R = safe(right);
	const same = JSON.stringify(L) === JSON.stringify(R);
	if (same) return '';
	const lib = await loadDiffLib();
	if (!lib) return '다름';
	try {
		const delta = lib
			.create({
				textDiff: { minLength: Number.MAX_SAFE_INTEGER },
				objectHash: (o: any, i: number) => (!o ? undefined : (o._id ?? o.id ?? o.name ?? '$$index:' + i))
			})
			.diff(L, R);
		if (!delta) return '';
		lib.formatters.html.hideUnchanged();
		return lib.formatters.html.format(delta, L) || '다름';
	} catch {
		return '다름';
	}
}

/** 동기 판정(버튼 켜기용) — 같으면 false */
export const isChanged = (left: any, right: any) => JSON.stringify(safe(left)) !== JSON.stringify(safe(right));
