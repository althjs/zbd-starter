/**
 * 날짜 도우미(브라우저 시간) — React utils/index.ts 의 getYMD·getYesterdayYMD·getThisWeekYMD·getThisMonthYMD 와 같은 계산.
 */
const n2 = (n: number) => (n < 10 ? '0' : '') + n;

export function getYMD(date?: any) {
	if (typeof date == 'string') {
		const t = date.replaceAll('-', '').trim();
		if (t.length != 8) return date;
		date = new Date(`${t.slice(0, 4)}-${t.slice(4, 6)}-${t.slice(6, 8)}`);
	}
	const d = date ? (typeof date.getMonth == 'function' ? date : new Date(date)) : new Date();
	return `${d.getFullYear()}-${n2(d.getMonth() + 1)}-${n2(d.getDate())}`;
}

/** n 일 전 */
export function getYesterdayYMD(date: Date | null = null, dayBefore = 1) {
	const d = date ? new Date(date) : new Date();
	d.setDate(d.getDate() - dayBefore);
	return getYMD(d);
}

/** 이번 주 월~일 */
export function getThisWeekYMD() {
	const now = new Date();
	const day = now.getDay();
	const mon = new Date(now);
	mon.setDate(now.getDate() - ((day === 0 ? 7 : day) - 1));
	const sun = new Date(mon);
	sun.setDate(mon.getDate() + 6);
	return { from: getYMD(mon), to: getYMD(sun) };
}

/** 이번 달 1일~말일 */
export function getThisMonthYMD() {
	const now = new Date();
	return { from: getYMD(new Date(now.getFullYear(), now.getMonth(), 1)), to: getYMD(new Date(now.getFullYear(), now.getMonth() + 1, 0)) };
}
