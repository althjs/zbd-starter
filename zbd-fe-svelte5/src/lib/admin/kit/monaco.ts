/**
 * Monaco 코드 편집기 불러오기 — React 판(@monaco-editor/react 4.7 → loader 기본값)과 같은 판·같은 곳(jsdelivr monaco-editor@0.55.1)에서.
 * 패키지를 더하지 않고(package.json 은 사용자 화면 세션 몫) 처음 쓸 때 AMD loader 를 붙인다.
 * 공용 후보: 메타 편집·웹하드·JSON 편집 서랍도 Monaco 를 쓴다.
 */
const VS = 'https://cdn.jsdelivr.net/npm/monaco-editor@0.55.1/min/vs';
let loading: Promise<any> | null = null;

export function loadMonaco(): Promise<any> {
	const w = window as any;
	if (w.monaco?.editor) return Promise.resolve(w.monaco);
	if (loading) return loading;
	loading = new Promise((resolve, reject) => {
		const boot = () => {
			w.require.config({ paths: { vs: VS } });
			w.require(['vs/editor/editor.main'], (m: any) => resolve(m || w.monaco), reject);
		};
		if (w.require?.config) return boot();
		const s = document.createElement('script');
		s.src = `${VS}/loader.js`;
		s.onload = boot;
		s.onerror = () => ((loading = null), reject(new Error('monaco 를 불러오지 못했습니다')));
		document.body.appendChild(s);
	});
	return loading;
}
