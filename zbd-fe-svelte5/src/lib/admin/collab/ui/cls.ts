/** 협업 화면에서 자주 쓰는 MUI 이름표 묶음 */
export const CARD = 'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root';
export const TYPO = 'MuiTypography-root MuiTypography-body1';
export const NOWRAP = 'MuiTypography-root MuiTypography-body1 MuiTypography-noWrap';
export const MENUITEM = 'MuiButtonBase-root MuiMenuItem-root MuiMenuItem-gutters';
/** 글자 색 — text.secondary · text.disabled */
export const SEC = 'var(--adm-text2)';
export const DIS = 'var(--adm-text-disabled)';
export const DIV = '1px solid var(--adm-divider)';
/** React `<TextField hiddenLabel multiline size="small">` 의 칸 여백 — 바탕 ui/TextField 는 라벨 없는 여러 줄을 라벨 있는 모양(21/4)으로 그린다(바탕 요청 후보) */
export const HIDDEN_MULTI_SX = { '& .MuiInputBase-root.MuiInputBase-multiline.MuiInputBase-sizeSmall': { padding: '8px 12px 9px' } };
