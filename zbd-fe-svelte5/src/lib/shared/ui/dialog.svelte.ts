/**
 * 알림·확인 창과 토스트 — 예전 useModal()/useToast() 자리.
 * 화면 어디서나 `await confirm("…")` 처럼 부르고, 그리는 것은 루트 레이아웃의 <Dialogs/> 하나다.
 */
import type { Snippet } from 'svelte';

export type DialogItem = {
	id: number;
	kind: 'alert' | 'confirm' | 'modal';
	title?: string;
	body?: string;
	/** 본문을 직접 그릴 때 (약관처럼 HTML 이 긴 창) */
	html?: string;
	wide?: boolean;
	resolve: (v: boolean) => void;
};

export type ToastItem = { id: number; msg: string };

let seq = 0;

export const dialogs = $state({ list: [] as DialogItem[], toasts: [] as ToastItem[] });

function open(kind: DialogItem['kind'], p: string | Partial<DialogItem>): Promise<boolean> {
	return new Promise((resolve) => {
		const base = typeof p == 'string' ? { body: p } : p;
		dialogs.list.push({ id: ++seq, kind, ...base, resolve });
	});
}

export function close(id: number, value: boolean) {
	const i = dialogs.list.findIndex((d) => d.id == id);
	if (i == -1) {
		return;
	}
	const [d] = dialogs.list.splice(i, 1);
	d.resolve(value);
}

export const alert = (p: string | Partial<DialogItem>) => open('alert', p);
export const confirm = (p: string | Partial<DialogItem>) => open('confirm', p);
export const modal = (p: Partial<DialogItem>) => open('modal', p);

export function toast(msg: string, ms = 2500) {
	const id = ++seq;
	dialogs.toasts.push({ id, msg });
	setTimeout(() => {
		const i = dialogs.toasts.findIndex((t) => t.id == id);
		if (i != -1) dialogs.toasts.splice(i, 1);
	}, ms);
}

export type { Snippet };
