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

// 빌드 경고 「Some chunks are larger than 500 kB」 대신(2026-10-20 대표 — 오류처럼 보여서) · 그 경고에 걸리던 것은 누를 때만 받는 덩어리
// (mermaid·elk·exceljs·관리자 다국어 단어표)라 처음 화면과 무관했다. 처음 화면에 실리는 것(시작 파일 + 정적으로 딸린 것)만 재서 넘으면 알린다.
const FIRST_LOAD_WARN_KB = 500;
function firstLoadSize(): Plugin {
	return {
		name: 'zbd-first-load-size',
		apply: 'build',
		applyToEnvironment: (e) => e.name === 'client',
		generateBundle(_o, bundle) {
			const chunks = Object.values(bundle).filter((c) => c.type === 'chunk') as any[];
			const byName = new Map(chunks.map((c) => [c.fileName, c]));
			const seen = new Set<string>();
			const walk = (f: string) => {
				if (seen.has(f)) return;
				seen.add(f);
				for (const i of byName.get(f)?.imports || []) walk(i);
			};
			for (const c of chunks) if (c.isEntry) walk(c.fileName);
			const big = [...seen].map((f) => byName.get(f)).filter((c) => c && c.code.length > FIRST_LOAD_WARN_KB * 1024);
			for (const c of big)
				this.warn(
					`처음 화면에 실리는 파일이 ${Math.round(c.code.length / 1024)}kB — ${c.fileName} (${Object.keys(c.modules)
						.slice(-3)
						.map((m) => m.replace(process.cwd() + '/', ''))
						.join(', ')}) · import() 로 나눌 것`
				);
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
			firstLoadSize(),
			sveltekit({
				compilerOptions: {
					// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
					runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true),
					// 빌드 로그에서 뺄 경고(오류처럼 보여서) · 접근성(a11y_*)과 「처음 값만 읽음」(props 로 첫 값을 잡는 일부러 쓴 모양).
					// 나머지(non_reactive_update·css_unused_selector 등 진짜 실수일 수 있는 것)는 그대로 보인다. svelte-check 에는 다 나온다.
					warningFilter: (w) => !(w.code.startsWith('a11y_') || w.code === 'state_referenced_locally')
				},
				// SPA — 사용자 화면의 HTML 은 지금처럼 BE SSR(seo) 이 만들어 이 앱을 붙인다.
				adapter: adapter({ fallback: 'index.html' }),
				// 빌드 파일은 /assets/ 아래로만(nginx 는 /assets/ 에만 파일 경로가 있다 — CLAUDE.md 2)
				appDir: 'assets/_app'
			})
		],
		build: {
			// 크기 경고는 firstLoadSize 가 처음 화면 것만 본다 · 플러그인 시간 표(PLUGIN_TIMINGS)는 정보일 뿐이라 끔
			chunkSizeWarningLimit: 100000,
			rolldownOptions: { checks: { pluginTimings: false } }
		},
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
