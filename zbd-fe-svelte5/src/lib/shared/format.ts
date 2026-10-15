/**
 * 날짜·숫자 표시 — 예전 utils/filters.tsx 의 date·number (한국 시간 기준).
 *   date(1790268300000)   → "2026.09.25(금) 01:45"
 *   date("2026-09-25")    → "2026.09.25(금)"
 *   number(12345)         → "12,345"
 */
const YOIL = ['일', '월', '화', '수', '목', '금', '토'];
const pad = (n: number) => (n < 10 ? '0' : '') + n;

/** 한국 시간의 연·월·일·요일·시·분 */
export function kst(v: any) {
	let t: number;
	if (v instanceof Date) t = v.getTime();
	else if (typeof v == 'number' || /^\d{11,}$/.test(String(v))) t = Number(v);
	else t = new Date(String(v).length == 10 ? `${v}T00:00:00+09:00` : String(v).replace(' ', 'T') + (/[zZ]|[+-]\d\d:?\d\d$/.test(String(v)) ? '' : '+09:00')).getTime();
	if (isNaN(t)) return null;
	const d = new Date(t + 9 * 3600_000);
	return { y: d.getUTCFullYear(), m: d.getUTCMonth() + 1, d: d.getUTCDate(), w: d.getUTCDay(), h: d.getUTCHours(), mi: d.getUTCMinutes() };
}

export function date(v: any, excludeHM = false) {
	if (!v) return v;
	if (String(v).length == 8 && /^\d{8}$/.test(String(v))) {
		v = `${String(v).slice(0, 4)}-${String(v).slice(4, 6)}-${String(v).slice(6, 8)}`;
		excludeHM = true;
	}
	if (String(v).length == 10 && !/^\d{10}$/.test(String(v))) excludeHM = true;
	const k = kst(v);
	if (!k) return '';
	let s = `${k.y}.${pad(k.m)}.${pad(k.d)}(${YOIL[k.w]})`;
	if (!excludeHM) s += ` ${pad(k.h)}:${pad(k.mi)}`;
	return s;
}

export function number(v: number | string | null | undefined) {
	if (v === 0 || v === '0') return '0';
	if (!v) return '';
	return String(v).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

/** 오늘이면 "13:05", 아니면 "09-25" — 예전 filters.formatDateByToday (브라우저 시간 기준, 숫자 문자열은 못 읽어 빈 값 — 예전과 같게) */
export function dateByToday(v: any) {
	if (!v) return '';
	const d = v instanceof Date ? v : typeof v == 'string' || typeof v == 'number' ? new Date(v) : null;
	if (!d || isNaN(d.getTime())) return '';
	const now = new Date();
	const today = d.getFullYear() == now.getFullYear() && d.getMonth() == now.getMonth() && d.getDate() == now.getDate();
	return today ? `${pad(d.getHours())}:${pad(d.getMinutes())}` : `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/**
 * 「3분 전」·「약 2시간 전」 — 예전 date-fns formatDistanceToNow(…, { addSuffix, locale: ko }) 와 같은 구간.
 */
export function timeAgo(v: any) {
	const t = typeof v == 'number' || /^\d{11,}$/.test(String(v)) ? Number(v) : new Date(v).getTime();
	if (isNaN(t)) return '';
	const diff = Date.now() - t;
	const future = diff < 0;
	const min = Math.round(Math.abs(diff) / 60000);
	let s: string;
	if (min < 1) s = '1분 미만';
	else if (min < 45) s = `${min}분`;
	else if (min < 90) s = '약 1시간';
	else if (min < 1440) s = `약 ${Math.round(min / 60)}시간`;
	else if (min < 2520) s = '1일';
	else if (min < 43200) s = `${Math.round(min / 1440)}일`;
	else if (min < 86400) s = `약 ${Math.round(min / 43200)}개월`;
	else {
		const months = Math.round(min / 43200);
		if (months < 12) s = `${months}개월`;
		else {
			const y = Math.floor(months / 12);
			const r = months % 12;
			s = r < 3 ? `약 ${y}년` : r < 9 ? `${y}년 이상` : `거의 ${y + 1}년`;
		}
	}
	return future ? `${s} 후` : `${s} 전`;
}

/** "09.25 13:05 (목)" — 예전 filters.shortDate (브라우저 시간) */
export function shortDate(v: any, excludeHM = false) {
	if (!v) return v;
	if (String(v).length == 10) excludeHM = true;
	const d = typeof v == 'string' || typeof v == 'number' ? new Date(/^\d{11,}$/.test(String(v)) ? Number(v) : v) : v;
	if (isNaN(d.getTime())) return '';
	let s = `${pad(d.getMonth() + 1)}.${pad(d.getDate())}`;
	if (!excludeHM) s += ` ${pad(d.getHours())}:${pad(d.getMinutes())}`;
	return s + ` (${YOIL[d.getDay()]})`;
}
