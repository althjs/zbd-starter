/**
 * `window.zbdSite` — 운영자 HTML·GNB 링크(SCRIPT 형) 가 부르는 함수들(예전 window.octopus — 2026-10-03 운영 데이터를 바꾸고 옛 이름은 지웠다).
 * 예전 UserPostManageDrawer 가 푸터에서 등록하던 것과 같은 이름·동작(로그인·로그아웃·약관·토스트…).
 * 사용자 화면 묶음 레이아웃이 한 번 등록한다 — 푸터가 없는 화면(라이프·미니홈피)에서도 쓸 수 있다.
 */
import DOMPurify from 'dompurify';
import { navigate } from '#lib/shared/navigate.ts';
import { api, endSession, getBrowserUid, getToken } from '#lib/shared/api.ts';
import { HOSTING_ID } from '#lib/shared/config.ts';
import { alert, confirm, modal, toast } from '#lib/shared/ui/dialog.svelte.ts';
import { openFooterInquiry } from './post/editor.svelte.ts';
import { doLogin } from './session.svelte.ts';

declare global {
	interface Window {
		/** 운영자 HTML·GNB 링크(SCRIPT) 가 부르는 사이트 함수 — 앱 격리 안의 `zbd.*` 브리지와 다른 것 */
		zbdSite?: any;
	}
}

/** 약관 창 — 약관 게시판(TERMS) 의 글을 창으로 */
export async function openTerms(postId: string) {
	const o = await api.get(`/service/user/terms?hostingId=${HOSTING_ID}`);
	if (!o.ok) return;
	const post = (o.data || []).find((x: any) => x.postId == postId);
	if (!post) {
		alert('등록된 약관이 없습니다.');
		return;
	}
	modal({ title: post.title || post.groupSubKey, html: DOMPurify.sanitize(post.body || ''), wide: true });
}

export async function doLogout() {
	if (!(await confirm('로그아웃 하시겠습니까?'))) return;
	await endSession('user');
	location.replace('/');
}

export function registerZbdSite() {
	if (typeof window == 'undefined' || window.zbdSite) return;
	window.zbdSite = {
		doLogin: () => doLogin(),
		doLogout,
		toast,
		alert,
		confirm,
		modal,
		closeModal: () => {},
		navigate: (path: string, opts: { replace?: boolean } = {}) => navigate(path, opts),
		openTerms,
		/** 푸터 「문의하기」 등 — 글쓰기 창을 연락처 칸과 함께(비로그인 접수) */
		openDrawer: (groupKey: string, groupSubKey: string, headTitle = '') => openFooterInquiry(groupKey, groupSubKey, headTitle),
		toggleGnbDropdown() {},
		scrollToElement(elementId: string, offsetY = 0) {
			const el = document.getElementById(elementId);
			if (el) window.scrollTo({ top: window.scrollY + el.getBoundingClientRect().top + offsetY, behavior: 'smooth' });
		},
		async downloadFile(path: string, fileName: string) {
			// 운영은 세션 쿠키(같은 출처라 함께 간다) — 토큰은 개발 주소에서만
			const token = getToken('user');
			const uid = getBrowserUid();
			const r = await fetch(`/service/user/file/download?name=${encodeURIComponent(path)}`, {
				credentials: 'include',
				headers: { ...(token ? { Authorization: `Bearer ${token}` } : {}), ...(uid ? { zbduid: uid } : {}) }
			});
			if (r.status != 200) {
				alert('다운로드 도중 오류가 발생하였습니다.');
				return;
			}
			const url = URL.createObjectURL(await r.blob());
			const a = document.createElement('a');
			a.href = url;
			a.download = fileName;
			a.click();
			setTimeout(() => URL.revokeObjectURL(url), 1000);
		}
	};
}
