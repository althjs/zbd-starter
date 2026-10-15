/**
 * 줄 단위 3-way 병합(2026-10-02, 웹하드 문서 동시 수정용) — 라이브러리 없이.
 *   base = 내가 불러온 원본, mine = 내가 고친 것, theirs = 그사이 다른 곳에서 저장된 것.
 *   서로 다른 곳을 고쳤으면 합치고, 같은 곳을 다르게 고쳤으면 conflict = true.
 *   import 없는 순수 파일.
 */

/** 문서 내용 지문 — BE `server-libs/file-explorer.ts textHash` 와 같은 계산(FNV-1a 32bit, UTF-16 단위) */
export function textHash(s: string): string {
	let h = 0x811c9dc5;
	for (let i = 0; i < s.length; i++) {
		h ^= s.charCodeAt(i);
		h = Math.imul(h, 0x01000193) >>> 0;
	}
	return h.toString(16) + ':' + s.length;
}

/** a·b 의 최장 공통 줄 — [aIndex, bIndex] 오름차순. 너무 크면 null */
function lcsPairs(a: string[], b: string[]): [number, number][] | null {
	// 앞뒤 같은 줄은 잘라 두고 가운데만 비교(대부분의 수정은 가운데 몇 줄이다)
	let pre = 0;
	while (pre < a.length && pre < b.length && a[pre] === b[pre]) pre++;
	let suf = 0;
	while (suf < a.length - pre && suf < b.length - pre && a[a.length - 1 - suf] === b[b.length - 1 - suf]) suf++;
	const A = a.slice(pre, a.length - suf);
	const B = b.slice(pre, b.length - suf);
	const n = A.length;
	const m = B.length;
	if (n * m > 4_000_000) return null;
	const dp: Uint32Array[] = [];
	for (let i = 0; i <= n; i++) dp.push(new Uint32Array(m + 1));
	for (let i = n - 1; i >= 0; i--) {
		for (let j = m - 1; j >= 0; j--) {
			dp[i][j] = A[i] === B[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
		}
	}
	const out: [number, number][] = [];
	for (let k = 0; k < pre; k++) out.push([k, k]);
	let i = 0;
	let j = 0;
	while (i < n && j < m) {
		if (A[i] === B[j]) {
			out.push([i + pre, j + pre]);
			i++;
			j++;
		} else if (dp[i + 1][j] >= dp[i][j + 1]) i++;
		else j++;
	}
	for (let k = 0; k < suf; k++) out.push([a.length - suf + k, b.length - suf + k]);
	return out;
}

const same = (x: string[], y: string[]) => x.length === y.length && x.every((v, i) => v === y[i]);

export function merge3(base: string, mine: string, theirs: string): { text: string; conflict: boolean } {
	if (mine === theirs || theirs === base) return { text: mine, conflict: false };
	if (mine === base) return { text: theirs, conflict: false };

	const B = base.split('\n');
	const M = mine.split('\n');
	const T = theirs.split('\n');
	const pm = lcsPairs(B, M);
	const pt = lcsPairs(B, T);
	if (!pm || !pt) return { text: mine, conflict: true };

	const bm = new Map<number, number>(pm);
	const bt = new Map<number, number>(pt);
	// 세 쪽 모두에 그대로 남은 base 줄 = 기준점
	const anchors: [number, number, number][] = [];
	for (let i = 0; i < B.length; i++) {
		if (bm.has(i) && bt.has(i)) anchors.push([i, bm.get(i)!, bt.get(i)!]);
	}
	anchors.push([B.length, M.length, T.length]); // 끝

	const out: string[] = [];
	let conflict = false;
	let bi = 0;
	let mi = 0;
	let ti = 0;
	for (const [ab, am, at] of anchors) {
		const cb = B.slice(bi, ab);
		const cm = M.slice(mi, am);
		const ct = T.slice(ti, at);
		if (same(cm, cb)) out.push(...ct);
		else if (same(ct, cb)) out.push(...cm);
		else if (same(cm, ct)) out.push(...cm);
		else {
			conflict = true;
			out.push(...cm);
		}
		if (ab < B.length) out.push(B[ab]);
		bi = ab + 1;
		mi = am + 1;
		ti = at + 1;
	}
	return { text: out.join('\n'), conflict };
}
