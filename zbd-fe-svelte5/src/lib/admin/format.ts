/**
 * 관리자 화면 날짜 표시 — 공용 format.ts(date·number·timeAgo) 에 없는 것만(React utils/filters).
 *   shortDate('2026-10-03T01:02:00Z') → "10.03 10:02 (토)"  (브라우저 시간)
 */
const YOIL = ['일', '월', '화', '수', '목', '금', '토'];
const p2 = (n: number) => (n < 10 ? '0' : '') + n;

export function shortDate(v: any, excludeHM = false) {
	if (!v) return v;
	if (String(v).length == 10) excludeHM = true;
	const d = v instanceof Date ? v : new Date(v);
	if (isNaN(d.getTime())) return '-';
	let s = `${p2(d.getMonth() + 1)}.${p2(d.getDate())}`;
	if (!excludeHM) s += ` ${p2(d.getHours())}:${p2(d.getMinutes())}`;
	return `${s} (${YOIL[d.getDay()]})`;
}
