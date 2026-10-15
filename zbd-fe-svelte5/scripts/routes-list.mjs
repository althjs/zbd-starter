// 화면에 있는 주소 목록 — build/routes.json(빌드 뒤 늘 만든다, `npm run build`).
// 서버(Go site/seo)가 이 목록에 없는 주소에는 404 + noindex 로 문서를 낸다(화면은 그대로 「없는 페이지」).
// 소스 라우트 폴더를 그대로 옮긴다 — 묶음 (이름) 은 빼고, [x] 한 칸 · [[x]] 있어도 없어도 · [...x] 나머지 전부. 플러그인이 더한 화면도 같이 들어간다.
import fs from 'node:fs';
import path from 'node:path';
const root = 'src/routes';
const out = new Set();
(function walk(dir) {
	for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
		if (e.isDirectory()) walk(path.join(dir, e.name));
		else if (/^\+page\.(svelte|ts|js)$/.test(e.name)) {
			const rel = path
				.relative(root, dir)
				.split(path.sep)
				.filter((s) => s && !/^\(.*\)$/.test(s));
			out.add('/' + rel.join('/'));
		}
	}
})(root);
fs.mkdirSync('build', { recursive: true });
fs.writeFileSync('build/routes.json', JSON.stringify([...out].sort()) + '\n');
console.log(`routes.json — 주소 ${out.size}개`);
