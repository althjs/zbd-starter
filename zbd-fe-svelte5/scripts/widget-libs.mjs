// 격리(iframe) 실행 앱이 읽어 갈 라이브러리를 static/assets/widget-libs 에 만든다 — 예전 fe-service vite.config widgetLibs()
//   chart.umd.js         Chart(쓰는 앱에만 넣는다)
//   material-icons.css   글꼴을 data: 로 넣은 아이콘 CSS — opaque 출처 iframe 은 교차출처 글꼴이 CORS 에 막힌다
// 실행: node scripts/widget-libs.mjs [chart.js dist 경로]  (chart.js 판을 올릴 때만 다시)
import fs from 'node:fs';
const out = 'static/assets/widget-libs';
fs.mkdirSync(out, { recursive: true });
const chart = process.argv[2] || '../octopus-fe-service/node_modules/chart.js/dist/chart.umd.js';
fs.copyFileSync(chart, `${out}/chart.umd.js`);
const css = fs.readFileSync('node_modules/material-icons/iconfont/filled.css', 'utf8');
const b64 = fs.readFileSync('node_modules/material-icons/iconfont/material-icons.woff2').toString('base64');
fs.writeFileSync(`${out}/material-icons.css`, css.replace(/src:[^;]+;/, `src: url("data:font/woff2;base64,${b64}") format("woff2");`));
console.log('ok', fs.readdirSync(out));
