/** 전화번호 모양(010-1234-5678 · 02-123-4567) */
export function getPhoneNumberFormat(num: string, mobile = false) {
	if (!num) return num;
	if (num.indexOf('02') == 0 && num.length > 12) num = num.substring(0, 12);
	else if (num.length > 13) num = num.substring(0, 13);
	if (num.indexOf('01') == 0) mobile = true;
	num = String(num).replace(/[^0-9]/g, '');
	const isMobile = num.length == 11 || mobile;
	if (!isMobile && num.indexOf('02') != 0 && num.length == 8) return num.substring(0, 4) + '-' + num.substring(4, 8);
	let no = '';
	num.split('').forEach((n, idx) => {
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
