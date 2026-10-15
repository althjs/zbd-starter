/**
 * 두 값의 차이(React utils/jsondiffpatch 의 doDiff 자리) — 라이브러리 없이 바뀐 경로만 뽑는다.
 *   diffList(원본, 지금) → [{ path: 'Json.agreeMap.마케팅', before, after, kind: 'add'|'remove'|'change' }]
 * 화면은 「저장 안 한 변경이 있나」(length > 0) 와 미리보기(DiffPreview) 에 쓴다.
 */
export type DiffItem = { path: string; before: any; after: any; kind: 'add' | 'remove' | 'change' };

const isObj = (v: any) => v && typeof v == 'object' && !(v instanceof Date);
const norm = (v: any) => (v instanceof Date ? v.toISOString() : v);

export function diffList(a: any, b: any, path = '', out: DiffItem[] = []): DiffItem[] {
	a = norm(a);
	b = norm(b);
	if (isObj(a) && isObj(b) && Array.isArray(a) == Array.isArray(b)) {
		const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
		for (const k of keys) diffList(a[k], b[k], path ? `${path}.${k}` : k, out);
		return out;
	}
	if (JSON.stringify(a) === JSON.stringify(b)) return out;
	if (a === undefined) out.push({ path, before: a, after: b, kind: 'add' });
	else if (b === undefined) out.push({ path, before: a, after: b, kind: 'remove' });
	else out.push({ path, before: a, after: b, kind: 'change' });
	return out;
}
