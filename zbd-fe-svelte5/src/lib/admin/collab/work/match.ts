/** 보드 내용 검색 — 제목·본문·메모 제목·메모 본문에서 글자를 찾아 앞뒤를 잘라 보인다(React workKeywordMatch) */
import type { WorkItem } from '#lib/shared/types/work.ts';

export type MatchField = 'title' | 'desc' | 'shapeTitle' | 'shapeBody';
export const FIELD_LABEL: Record<MatchField, string> = { title: '제목', desc: '본문', shapeTitle: '메모 제목', shapeBody: '메모 본문' };
export type Match = { workId: number; workTitle: string; field: MatchField; shapeNo?: number; shapeId?: string | number; snippet: string; at: number; len: number };

export function stripHtml(html: string) {
	return String(html || '')
		.replace(/<(br|\/p|\/div|\/li|\/h[1-6])\s*\/?>/gi, ' ')
		.replace(/<[^>]*>/g, '')
		.replace(/&nbsp;/gi, ' ')
		.replace(/&amp;/gi, '&')
		.replace(/&lt;/gi, '<')
		.replace(/&gt;/gi, '>')
		.replace(/\s+/g, ' ')
		.trim();
}
export function findMatches(items: WorkItem[], keyword: string): Match[] {
	const kw = String(keyword || '')
		.trim()
		.toLowerCase();
	if (!kw) return [];
	const out: Match[] = [];
	const push = (o: WorkItem, field: MatchField, text: any, shapeNo?: number, shapeId?: string | number) => {
		const plain = String(text || '');
		const at = plain.toLowerCase().indexOf(kw);
		if (!plain || at < 0) return;
		const from = Math.max(0, at - 24);
		const to = Math.min(plain.length, at + kw.length + 24);
		const head = from > 0 ? '…' : '';
		out.push({
			workId: Number(o.workId) || 0,
			workTitle: String(o.Json?.title || '(제목 없음)'),
			field,
			shapeNo,
			shapeId,
			snippet: `${head}${plain.slice(from, to)}${to < plain.length ? '…' : ''}`,
			at: head.length + (at - from),
			len: kw.length
		});
	};
	for (const o of items || []) {
		push(o, 'title', o.Json?.title);
		push(o, 'desc', stripHtml(o.Json?.desc || ''));
		(Array.isArray(o.Json?.shapes) ? o.Json.shapes : []).forEach((s, i) => {
			if (!s || s.isDeleted) return;
			push(o, 'shapeTitle', s.title, i + 1, s.id);
			push(o, 'shapeBody', stripHtml(s.html || ''), i + 1, s.id);
		});
	}
	return out;
}
