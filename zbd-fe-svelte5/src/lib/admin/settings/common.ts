/** 묶음 settings 공용 값 — React components/common/user/constants · utils/tag 에서 옮김 */

/** 가입 방식(회원 groupId) */
export const socialUserGroupIdMap: Record<number, string> = { 0: '이메일', 1: '카카오', 2: '네이버', 3: '구글' };

const osMap: Record<string, string> = { A: 'Android', W: 'Windows', M: 'Macintosh', I: 'iOS', Z: 'Etc' };
const browserMap: Record<string, string> = { C: 'Chrome', S: 'Safari', E: 'Edge', F: 'Firefox', W: 'Whale', Z: 'Etc' };

/** 브라우저 id(zbd_uid) 앞 두 글자 = 기기·브라우저, 나머지 = 만든 시각(36진수) */
export function parseBrowserUID(uid?: string | null) {
	if (!uid) return { os: 'UNKNOWN', browser: 'UNKNOWN', createdAt: 'UNKNOWN' as number | string };
	return { os: osMap[uid.charAt(0)] || 'UNKNOWN', browser: browserMap[uid.charAt(1)] || 'UNKNOWN', createdAt: parseInt(uid.substring(2), 36) || 'UNKNOWN' };
}

import { getHostingItems, getMeta } from '../meta.svelte.ts';

/** 회원 상태 */
export const USER_STATUS_MAP: Record<string, string> = { 0: '정상회원', 99: '탈퇴요청', '-1': '탈퇴완료' };

/** 회원 기본 권한 그룹(COMMON_CODES AUTH_GROUP 중 사용·수동 그룹 아님) */
export function getMemberBasicAuths(): any[] {
	return ((getMeta('COMMON_CODES')?.items || []).find((x: any) => x.group == 'AUTH_GROUP')?.items || []).filter((x: any) => x.group != 'POST_GROUP_MANUAL' && x.use);
}

/** 약관 게시판(POST_GROUPS adminPostGroups TERMS) */
export function getTermPostGroup(): any {
	return (getMeta('POST_GROUPS')?.adminPostGroups || []).find((x: any) => x.groupKey == 'TERMS');
}

/** 숫자만 남기기(React utils getNumber 그대로) — 빈 값은 그대로, 0 이면 0, 숫자 */
export function getNumber(str: any, len: number | null = null, allowMinus = false): any {
	if (!str) return str;
	if (allowMinus && str === '-') return str;
	const isMinus = allowMinus && String(str).substring(0, 1) == '-';
	let num = String(str).replace(/[^0-9]/g, '');
	if (len && num.length > len) num = num.substring(0, len);
	if (parseInt(num, 10) === 0) return 0;
	if (num.length == 0) return '';
	return isMinus ? parseInt(num, 10) * -1 : parseInt(num, 10);
}

/** 전화번호 꼴(React utils getPhoneNumberFormat 그대로) */
export function getPhoneNumberFormat(num: any, mobile = false) {
	if (!num) return num;
	if (num.indexOf('02') == 0 && num.length > 12) num = num.substring(0, 12);
	else if (num.length > 13) num = num.substring(0, 13);
	if (num.indexOf('01') == 0) mobile = true;
	num = String(num).replace(/[^0-9]/g, '');
	const isMobile = num.length == 11 || mobile;
	if (!isMobile && num.indexOf('02') != 0 && num.length == 8) return num.substring(0, 4) + '-' + num.substring(4, 8);
	let no = '';
	num.split('').forEach((n: string, idx: number) => {
		if (isMobile) {
			if (idx == 3 || idx == 7) no += '-';
		} else {
			const first = num.indexOf('02') == 0 ? 2 : 3;
			if (idx == first || idx == (num.length == 9 ? 5 : 6)) no += '-';
		}
		no += n;
	});
	return no.length > 13 ? no.substring(0, 13) : no;
}

/** 다음 번호(React utils getNextId) — 가장 큰 id + 1 */
export function getNextId(items: any[] | null | undefined, idKey = 'id') {
	if (!items) return null;
	let max = 0;
	for (const o of items) {
		const n = parseInt(o?.[idKey], 10) || 0;
		if (max < n) max = n;
	}
	return max + 1;
}

/** 깊은 복사(JSON 값) */
export const clone = <T>(o: T): T => (o === undefined ? o : JSON.parse(JSON.stringify(o)));

/** 날짜 + 초(React filters.date(v, false, true)) — 「2026.10.03(토) 07:20:11」 */
export function dateSec(v: any) {
	if (!v) return v;
	const d = new Date(typeof v == 'string' && /^\d+$/.test(v) ? Number(v) : v);
	if (isNaN(d.getTime())) return '';
	const p = (n: number) => (n < 10 ? '0' : '') + n;
	return `${d.getFullYear()}.${p(d.getMonth() + 1)}.${p(d.getDate())}(${'일월화수목금토'[d.getDay()]}) ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

/** 호스팅 이름(설정 HOSTINGS) — 없으면 id 그대로 */
export function hostingName(hostingId: string) {
	return getHostingItems().find((c: any) => c.hostingId == hostingId)?.name || hostingId;
}
