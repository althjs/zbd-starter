/**
 * 관리자 알림·확인·입력·내용 창과 토스트 — React components/common/modal/ModalDialog·toast/Toast 자리(같은 모양).
 *
 *   await alert('저장했습니다')                       → true
 *   if (await confirm('삭제할까요?')) …               → true/false
 *   const v = await prompt({ title: '이름', promptDefaultValue: 'a' })   → 문자열 | false
 *   modal({ title: '지역 설정', child: snippet })      → 닫으려면 closeModal()
 *   toast('저장했습니다') · toast({ msg: '실패', type: 'error' })
 *
 * 그리는 것은 관리자 레이아웃의 <AdminDialogs/> 하나(공용 shared/ui/Dialogs 는 사용자 화면 모양이라 관리자는 이걸 쓴다).
 * React 판처럼 **한 번에 하나** — 새 창이 오면 기다리던 창은 false 로 끝난다.
 */
import type { Snippet } from 'svelte';

export type ModalType = 'alert' | 'confirm' | 'prompt' | 'modal';
export type ModalData = {
	type?: 'info' | 'success' | 'warning' | 'error' | 'none';
	title?: string;
	body?: string;
	child?: Snippet;
	persistent?: boolean;
	width?: number | string;
	cancelTitle?: string;
	okTitle?: string;
	closeUse?: boolean;
	hiddenIcon?: boolean;
	promptDefaultValue?: string;
	promptAllowEmpty?: boolean;
	promptNumeric?: boolean;
	placeholder?: string;
};

type Current = { kind: ModalType; data: ModalData; resolvers: ((v: any) => void)[]; seq: number };

export const modalState = $state({ cur: null as Current | null });
let seq = 0;

function open(kind: ModalType, p: ModalData | string): Promise<any> {
	const data: ModalData = typeof p == 'string' ? { body: p, persistent: kind != 'modal' } : { ...p };
	data.type ||= kind == 'modal' ? 'none' : 'info';
	if (kind != 'modal') data.persistent = data.persistent === undefined || data.persistent;
	return new Promise((resolve) => {
		const prev = modalState.cur;
		if (prev && (prev.data.title != data.title || prev.data.body != data.body || prev.kind != kind)) {
			prev.resolvers.forEach((r) => r(false));
		}
		modalState.cur = { kind, data, resolvers: kind == 'modal' ? [] : [resolve], seq: ++seq };
		if (kind == 'modal') resolve(true);
	});
}

/** 창 닫기 — value 가 그대로 기다리던 쪽에 간다 */
export function closeModal(value: any = false) {
	const c = modalState.cur;
	if (!c) return;
	c.resolvers.forEach((r) => r(value));
	modalState.cur = null;
}

export const alert = (p: ModalData | string): Promise<true | false> => open('alert', p);
export const confirm = (p: ModalData | string): Promise<boolean> => open('confirm', p);
export const prompt = (p: ModalData | string): Promise<string | false> => open('prompt', p);
export const modal = (p: ModalData | string): void => void open('modal', p);

export type ToastType = 'success' | 'info' | 'warning' | 'error';
export type ToastData = { msg: string; type?: ToastType; autoHideDuration?: number };
export const toastState = $state({ cur: null as (ToastData & { show: boolean; id: number }) | null });
let toastTimer: any = null;
let toastSeq = 0;

/** 문자열이면 성공(초록), 객체면 type 없을 때 오류(빨강) — React useToast 와 같은 기본값 */
export function toast(p: ToastData | string) {
	clearTimeout(toastTimer);
	const d: ToastData = typeof p == 'string' ? { msg: p, type: 'success' } : { ...p, type: p.type || 'error' };
	const id = ++toastSeq;
	toastState.cur = { ...d, show: false, id };
	setTimeout(() => toastState.cur?.id == id && (toastState.cur.show = true), 150);
	toastTimer = setTimeout(() => toastState.cur?.id == id && (toastState.cur.show = false), d.autoHideDuration || 3000);
}
