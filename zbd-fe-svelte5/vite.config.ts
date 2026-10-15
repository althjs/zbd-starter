import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { readFileSync, realpathSync } from 'node:fs';
import { resolve } from 'node:path';
import type { Plugin } from 'vite';
import { defineConfig, loadEnv } from 'vite';

/**
 * 개발 서버 전용 — 홈의 「관리자·사용자 암호 보기」 가 부르는 `GET /__dev/account?who=admin|user`.
 * run_dev.sh 가 만든 ../.dev/account.txt 를 읽어 준다. 이 컴퓨터(localhost·127.0.0.1)에서 온 요청만 — 빌드(배포)에는 들어가지 않는다.
 */
function devAccount(): Plugin {
	return {
		name: 'zbd-dev-account',
		apply: 'serve',
		configureServer(srv) {
			srv.middlewares.use('/__dev/account', (req, res) => {
				const host = String(req.headers.host || '').split(':')[0];
				if (host != 'localhost' && host != '127.0.0.1') {
					res.statusCode = 403;
					return res.end('localhost 에서만');
				}
				const who = new URL(req.url || '', 'http://x').searchParams.get('who') == 'user' ? '[사용자]' : '[관리자]';
				let txt = '';
				try {
					txt = readFileSync(resolve(process.cwd(), '../.dev/account.txt'), 'utf8');
				} catch {
					res.statusCode = 404;
					return res.end('.dev/account.txt 가 없습니다 — ./run_dev.sh 로 만듭니다');
				}
				const block = txt.split(/\n(?=\[)/).find((b) => b.trim().startsWith(who)) || '';
				const get = (k: string) => (block.match(new RegExp('^' + k + '=(.*)$', 'm')) || [])[1] || '';
				res.setHeader('Content-Type', 'application/json; charset=utf-8');
				res.setHeader('Cache-Control', 'no-store');
				res.end(JSON.stringify({ email: get('email'), password: get('password') }));
			});
		}
	};
}

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	// 개발 때 API 를 받을 곳 — 기본은 이 컴퓨터의 BE(run_dev.sh 가 띄운다). 운영 서버를 보려면 ZBD_API=https://도메인
	const target = env.ZBD_API || 'http://127.0.0.1:5001';

	return {
		plugins: [
			devAccount(),
			sveltekit({
				compilerOptions: {
					// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
					runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
				},
				// SPA — 사용자 화면의 HTML 은 지금처럼 BE SSR(seo) 이 만들어 이 앱을 붙인다.
				adapter: adapter({ fallback: 'index.html' }),
				// 빌드 파일은 /assets/ 아래로만(nginx 는 /assets/ 에만 파일 경로가 있다 — CLAUDE.md 2)
				appDir: 'assets/_app'
			})
		],
		server: {
			port: Number(env.FE_PORT || 5000),
			// 기본 = 같은 네트워크의 다른 기기(휴대폰·다른 PC)에서도 http://{이 컴퓨터 IP}:5000 으로 열린다(2026-10-06). 이 컴퓨터에서만 열려면 FE_HOST=localhost
			host: env.FE_HOST || true,
			strictPort: true,
			// 시험 워크트리(.wt-sv5-stable 등)는 node_modules 를 본 폴더에서 심링크로 빌린다 — 그 실제 경로도 열어야 아이콘 글꼴 등이 403 이 안 난다
			fs: { allow: ['.', realpathSync('node_modules')] },
			proxy: {
				// 원래 주소(localhost:5000)를 서버에 알린다 — 간편 로그인의 돌아오는 주소처럼 서버가 요청 주소로 만드는 값이 개발에서도 맞게
				'/service': {
					target,
					changeOrigin: true,
					secure: target.startsWith('https'),
					configure: (proxy) => proxy.on('proxyReq', (req, inc) => inc.headers.host && req.setHeader('x-forwarded-host', inc.headers.host))
				},
				// 새 목록 API(Go v1) — 운영 nginx 에 `/api/v1/` 가 있어야 한다
				'/api/v1': { target, changeOrigin: true, secure: target.startsWith('https') },
				// 웹소켓(소켓 데몬 127.0.0.1:4006, run_dev 의 SOCK_PORT 면 ZBD_WS 로 바뀐다) — 웹소켓 플러그인을 붙였을 때만 쓰인다(운영은 웹서버의 location = /ws)
				'/ws': { target: env.ZBD_WS || 'ws://127.0.0.1:4006', ws: true, changeOrigin: true },
				'/upload': { target, changeOrigin: true, secure: target.startsWith('https') },
				// 운영 BE 가 주는 정적 파일·설치 앱 정보 — 개발에서도 같은 것을 받는다
				'/swiper': { target, changeOrigin: true, secure: target.startsWith('https') },
				// 운영 문서 루트(/srv/zbd) 의 공개 파일 — 회원이 만든 앱이 `/images/freecell/cards/…` 처럼 바로 부른다(배포 때 이 파일들을 그대로 둘 것)
				'/images': { target, changeOrigin: true, secure: target.startsWith('https') },
				'/uber-fonts': { target, changeOrigin: true, secure: target.startsWith('https') },
				'/1x1.gif': { target, changeOrigin: true, secure: target.startsWith('https') },
				'/service-worker.js': { target, changeOrigin: true, secure: target.startsWith('https') },
				'/manifest.json': { target, changeOrigin: true, secure: target.startsWith('https') }
				// 실시간 /ws 프록시는 웹소켓 연동 플러그인이 더한다
			}
		}
	};
});
