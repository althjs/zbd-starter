/** 날짜 문자열 도우미 — React utils(getYMD·getYesterdayYMD·getThisWeekYMD·getThisMonthYMD) 와 같은 계산 */
const p2 = (n: number) => (n < 10 ? '0' : '') + n;

export function getYMD(date?: any): string {
	if (typeof date == 'string') {
		const t = date.replaceAll('-', '').trim();
		if (t.length != 8) return date;
		date = new Date(`${t.substring(0, 4)}-${t.substring(4, 6)}-${t.substring(6, 8)}`);
	}
	let d = date || new Date();
	if (typeof d.getMonth != 'function') d = new Date(d);
	return `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`;
}

/** dayBefore 일 전(음수면 뒤) */
export function getYesterdayYMD(date: Date | null = null, dayBefore = 1) {
	const d = date ? new Date(date) : new Date();
	d.setDate(d.getDate() - dayBefore);
	return getYMD(d);
}

export function getThisWeekYMD() {
	const now = new Date();
	const day = now.getDay();
	const mon = new Date(now);
	mon.setDate(now.getDate() - ((day === 0 ? 7 : day) - 1));
	const sun = new Date(mon);
	sun.setDate(mon.getDate() + 6);
	return { from: getYMD(mon), to: getYMD(sun) };
}

export function getThisMonthYMD() {
	const now = new Date();
	return { from: getYMD(new Date(now.getFullYear(), now.getMonth(), 1)), to: getYMD(new Date(now.getFullYear(), now.getMonth() + 1, 0)) };
}

/** 목록 id 다음 번호 */
export function getNextId(list: { id?: number }[] | null | undefined) {
	return (list || []).reduce((m, o) => Math.max(m, Number(o.id) || 0), 0) + 1;
}

/** blob 주소를 파일로 받기 */
export function fileDownload(url: string, fileName: string) {
	const a = document.createElement('a');
	a.href = url;
	a.download = fileName;
	document.body.appendChild(a);
	a.click();
	a.remove();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export const cloneDeep = <T>(o: T): T => (o == null ? o : JSON.parse(JSON.stringify(o)));
