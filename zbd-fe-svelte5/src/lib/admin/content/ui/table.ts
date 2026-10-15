/**
 * 표 몸통 칸 바로잡기 — React 관리자 테마 값과 바탕 admin.css 가 다른 두 곳(게시판 목록 실측, 바탕에 알림):
 *  1. 몸통 칸 줄 높이: React 1.5(21px) · 바탕 1.43 → 여러 줄 칸이 1px 씩 어긋난다.
 *  2. 휴대폰(<600) 몸통 td 여백: React 는 `4px 16px` 그대로(theme `& td` 가 이김) · 바탕은 8px 로 줄어 표 폭이 달라진다.
 *     바탕 규칙(html.zbd-admin main .MuiTableCell-root, 0,2,2)을 이기려고 클래스를 세 번 겹친다.
 */
export const bodyCellFix = {
	'& .MuiTableCell-body': { lineHeight: 1.5 },
	'@media (max-width:599.95px)': { '&&& td.MuiTableCell-root': { padding: '4px 16px' } }
};
// ⚠ 화면이 몸통 칸 여백을 따로 주는 표(동영상·알림톡 등)는 위 휴대폰 규칙(&&&)을 이기도록 '&&&& td.MuiTableCell-root' 로 쓴다.
