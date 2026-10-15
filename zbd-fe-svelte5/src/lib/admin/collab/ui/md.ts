/**
 * 마크다운 → HTML — React 판 react-markdown + remark-gfm + remark-breaks 와 같은 결과를 목표로 한 작은 변환기
 * (이 앱엔 마크다운 라이브러리가 없고 package.json 은 다른 세션 소유라 더할 수 없다 — 공용 후보).
 *
 *   제목 · 문단(줄바꿈 = <br>, remark-breaks) · 굵게/기울임/취소선 · 인라인 코드 · 코드 블록(```언어) · 인용 ·
 *   목록(중첩·숫자·할 일 [ ] [x]) · 표(GFM, 정렬) · 가로줄 · 링크 · 그림 · 주소 자동 링크(GFM)
 * 원문 HTML 은 글자로 보인다(react-markdown 기본과 같게 — 그리지 않는다). 결과는 DOMPurify 를 한 번 더 거친다.
 * 코드 블록은 `<pre><code class="language-x">` — 색칠(구문 강조)·mermaid 그림은 하지 않는다.
 */
import DOMPurify from 'dompurify';

const esc = (s: string) => s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c] as string);

/** 인라인 — 코드 조각을 먼저 빼 두고(안쪽은 그대로), 나머지에 강조·링크를 입힌다 */
function inline(src: string): string {
	const codes: string[] = [];
	let s = src.replace(/(`+)([\s\S]*?[^`])\1(?!`)/g, (_, _t, c) => {
		codes.push(`<code>${esc(c.replace(/^ (.*) $/, '$1'))}</code>`);
		return `\u0000${codes.length - 1}\u0000`;
	});
	// 백슬래시 이스케이프
	const escapes: string[] = [];
	s = s.replace(/\\([\\`*_{}\[\]()#+\-.!|~>])/g, (_, c) => {
		escapes.push(esc(c));
		return `\u0001${escapes.length - 1}\u0001`;
	});
	s = esc(s);
	const links: string[] = [];
	const keep = (html: string) => {
		links.push(html);
		return `\u0002${links.length - 1}\u0002`;
	};
	// 그림 · 링크
	s = s.replace(/!\[([^\]]*)\]\(\s*([^)\s]+)(?:\s+&quot;([^&]*)&quot;)?\s*\)/g, (_, alt, url, title) => keep(`<img src="${url}" alt="${alt}"${title ? ` title="${title}"` : ''}>`));
	s = s.replace(/\[([^\]]+)\]\(\s*([^)\s]+)(?:\s+&quot;([^&]*)&quot;)?\s*\)/g, (_, txt, url, title) => keep(`<a href="${url}"${title ? ` title="${title}"` : ''}>${txt}</a>`));
	s = s.replace(/&lt;(https?:\/\/[^\s&]+)&gt;/g, (_, url) => keep(`<a href="${url}">${url}</a>`));
	// GFM 자동 링크(주소·www.)
	s = s.replace(/(^|[\s(（「])((?:https?:\/\/|www\.)[^\s<]*[^\s<.,:;"')\]}」）!?])/g, (_, pre, url) => pre + keep(`<a href="${url.startsWith('www.') ? 'http://' + url : url}">${url}</a>`));
	// 강조
	s = s.replace(/(\*\*|__)(?=\S)([\s\S]*?\S)\1/g, '<strong>$2</strong>');
	s = s.replace(/(^|[^\w*])\*(?=\S)([\s\S]*?\S)\*(?!\*)/g, '$1<em>$2</em>');
	s = s.replace(/(^|[^\w_])_(?=\S)([\s\S]*?\S)_(?![\w_])/g, '$1<em>$2</em>');
	s = s.replace(/~~(?=\S)([\s\S]*?\S)~~/g, '<del>$1</del>');
	s = s.replace(/(^|[^~])~(?=\S)([^~]*?\S)~(?!~)/g, '$1<del>$2</del>');
	// 줄바꿈(remark-breaks) — 줄 끝 공백 둘·백슬래시도
	s = s.replace(/ {2,}\n|\\\n|\n/g, '<br>\n');
	s = s.replace(/\u0002(\d+)\u0002/g, (_, i) => links[+i]);
	s = s.replace(/\u0001(\d+)\u0001/g, (_, i) => escapes[+i]);
	return s.replace(/\u0000(\d+)\u0000/g, (_, i) => codes[+i]);
}

const isHr = (l: string) => /^ {0,3}([-*_])(?:\s*\1){2,}\s*$/.test(l);
const listRe = /^( {0,})([-*+]|\d{1,9}[.)])\s+(.*)$/;
const fenceRe = /^ {0,3}(`{3,}|~{3,})\s*([\w+#.-]*)?.*$/;
const isTableSep = (l: string) => /^\s*\|?\s*:?-{1,}:?\s*(\|\s*:?-{1,}:?\s*)*\|?\s*$/.test(l) && l.includes('-');
const cells = (l: string) =>
	l
		.trim()
		.replace(/^\|/, '')
		.replace(/(?<!\\)\|$/, '')
		.split(/(?<!\\)\|/)
		.map((c) => c.trim().replace(/\\\|/g, '|'));

function blocks(lines: string[]): string {
	let out = '';
	let i = 0;
	while (i < lines.length) {
		const line = lines[i];
		if (!line.trim()) {
			i++;
			continue;
		}
		// 코드 블록
		const f = line.match(fenceRe);
		if (f) {
			const fence = f[1];
			const lang = (f[2] || '').trim();
			const body: string[] = [];
			i++;
			while (i < lines.length && !lines[i].trim().startsWith(fence)) body.push(lines[i++]);
			i++;
			out += `<pre><code${lang ? ` class="language-${esc(lang)}"` : ''}>${esc(body.join('\n'))}\n</code></pre>\n`;
			continue;
		}
		// 제목
		const h = line.match(/^ {0,3}(#{1,6})\s+(.*?)\s*#*\s*$/);
		if (h) {
			out += `<h${h[1].length}>${inline(h[2])}</h${h[1].length}>\n`;
			i++;
			continue;
		}
		if (isHr(line)) {
			out += '<hr>\n';
			i++;
			continue;
		}
		// 인용
		if (/^ {0,3}>/.test(line)) {
			const body: string[] = [];
			while (i < lines.length && lines[i].trim() && /^ {0,3}>/.test(lines[i])) body.push(lines[i++].replace(/^ {0,3}> ?/, ''));
			out += `<blockquote>\n${blocks(body)}</blockquote>\n`;
			continue;
		}
		// 표
		if (line.includes('|') && i + 1 < lines.length && isTableSep(lines[i + 1])) {
			const head = cells(line);
			const aligns = cells(lines[i + 1]).map((c) => (c.startsWith(':') && c.endsWith(':') ? 'center' : c.endsWith(':') ? 'right' : c.startsWith(':') ? 'left' : ''));
			i += 2;
			const rows: string[][] = [];
			while (i < lines.length && lines[i].trim() && lines[i].includes('|')) rows.push(cells(lines[i++]));
			const al = (k: number) => (aligns[k] ? ` style="text-align:${aligns[k]}"` : '');
			out += `<table>\n<thead>\n<tr>\n${head.map((c, k) => `<th${al(k)}>${inline(c)}</th>`).join('\n')}\n</tr>\n</thead>\n`;
			if (rows.length) out += `<tbody>\n${rows.map((r) => `<tr>\n${head.map((_, k) => `<td${al(k)}>${inline(r[k] || '')}</td>`).join('\n')}\n</tr>`).join('\n')}\n</tbody>\n`;
			out += '</table>\n';
			continue;
		}
		// 목록
		const lm = line.match(listRe);
		if (lm && lm[1].length < 4) {
			const indent = lm[1].length;
			const ordered = /\d/.test(lm[2]);
			const start = ordered ? parseInt(lm[2]) : 1;
			const items: string[][] = [];
			let loose = false;
			while (i < lines.length) {
				const m = lines[i].match(listRe);
				if (m && m[1].length == indent && /\d/.test(m[2]) == ordered) {
					items.push([m[3]]);
					i++;
					continue;
				}
				if (!lines[i].trim()) {
					// 빈 줄 뒤에 같은 목록이 이어지면 느슨한 목록
					const next = lines[i + 1];
					const nm = next?.match(listRe);
					const sameList = nm && (nm[1].length > indent || /\d/.test(nm[2]) == ordered);
					if (next && ((nm && sameList) || (!nm && /^\s{2,}\S/.test(next)))) {
						if (nm && nm[1].length == indent) loose = true;
						items[items.length - 1].push('');
						i++;
						continue;
					}
					break;
				}
				if (/^\s+\S/.test(lines[i]) && lines[i].search(/\S/) > indent) {
					items[items.length - 1].push(lines[i].slice(Math.min(lines[i].search(/\S/), indent + (ordered ? 3 : 2))));
					i++;
					continue;
				}
				if (items.length && !m && !isHr(lines[i]) && !fenceRe.test(lines[i]) && !/^ {0,3}(#|>)/.test(lines[i])) {
					items[items.length - 1].push(lines[i]); // 게으른 이어 쓰기
					i++;
					continue;
				}
				break;
			}
			const tag = ordered ? 'ol' : 'ul';
			out += `<${tag}${ordered && start != 1 ? ` start="${start}"` : ''}>\n`;
			for (const it of items) {
				let first = it[0];
				let task = '';
				const t = first.match(/^\[( |x|X)\]\s+(.*)$/);
				if (t) {
					task = `<input type="checkbox" disabled${t[1] != ' ' ? ' checked' : ''}> `;
					first = t[2];
				}
				const rest = [first, ...it.slice(1)];
				// 하위 블록(중첩 목록·코드)이 없고 느슨하지 않으면 문단 없이
				let inner = blocks(rest);
				// 촘촘한 목록(tight)은 첫 문단을 <p> 없이 — remark 와 같게
				if (!loose) inner = inner.replace(/^<p>([\s\S]*?)<\/p>\n/, '$1\n');
				const body = !loose && !/\n[\s\S]/.test(inner.replace(/\n$/, '')) ? inner.replace(/\n$/, '') : `${loose ? '\n' : ''}${inner}`;
				out += `<li${task ? ' class="task-list-item"' : ''}>${task}${body}</li>\n`;
			}
			out += `</${tag}>\n`;
			continue;
		}
		// 문단 — 다음 블록이 시작되기 전까지
		const para: string[] = [];
		while (
			i < lines.length &&
			lines[i].trim() &&
			!fenceRe.test(lines[i]) &&
			!/^ {0,3}(#{1,6}\s|>)/.test(lines[i]) &&
			!isHr(lines[i]) &&
			!(para.length && listRe.test(lines[i]) && lines[i].match(listRe)![1].length < 4) &&
			!(lines[i].includes('|') && isTableSep(lines[i + 1] || ''))
		) {
			para.push(lines[i++].replace(/^ {0,3}/, ''));
		}
		if (!para.length) para.push(lines[i++]);
		out += `<p>${inline(para.join('\n').replace(/\s+$/, ''))}</p>\n`;
	}
	return out;
}

export function markdownToHtml(md: string): string {
	const html = blocks(`${md || ''}`.replace(/\r\n?/g, '\n').replace(/\t/g, '    ').split('\n'));
	return DOMPurify.sanitize(html, { ADD_ATTR: ['target', 'start', 'align'] });
}

/** React markdownBoxSx — 마크다운 보기 상자(제목 밑줄·표 격자·줄무늬) */
const GT200 = 'var(--gt-200, #f2f2f2)';
export const MD_BOX_SX = {
	'& h1, & h2': { borderBottom: `1px solid ${GT200}`, paddingBottom: 8, marginBottom: 8, fontSize: '1.2rem' },
	'& :not(pre) > code': { paddingLeft: 4, paddingRight: 4, borderRadius: '4px' },
	'& table': { width: '100%', borderCollapse: 'collapse', marginBottom: 16, border: `1px solid ${GT200}` },
	'& th, & td': { border: `1px solid ${GT200}`, padding: 8, textAlign: 'left' },
	'& th': { backgroundColor: 'var(--gt-100, #fafafa)', fontWeight: 600 },
	'& thead tr': { backgroundColor: GT200 },
	'& tr:nth-of-type(even)': { backgroundColor: 'var(--gt-150, #f5f5f5)' }
};
/** React HTML_CONTENT_TABLE_SX — 편집기로 쓴 본문의 표 */
export const HTML_TABLE_SX = {
	'& table': { borderCollapse: 'collapse', width: '100%', fontSize: 14 },
	'& table th, & table td': { border: '1px solid #d0d0d0', padding: '6px 10px', verticalAlign: 'middle', lineHeight: 1.5 },
	'& table th': { background: '#f2f2f2', fontWeight: 700, textAlign: 'center' },
	'& table p': { margin: 0 }
};
