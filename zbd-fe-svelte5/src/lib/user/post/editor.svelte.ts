/**
 * 글쓰기·글 수정 창 상태 — 예전 UserPostManageDrawer 의 ref.openDrawer 자리.
 * 창(<PostEditorDrawer/>) 은 사용자 화면 묶음 레이아웃에 하나만 둔다. 어디서나 openPostEditor(…) 로 연다.
 *
 *   새 글       openPostEditor({ groupKey, groupSubKey, onSave })
 *   수정        openPostEditor({ post, onSave })                    — 서버에서 최신본을 다시 읽는다
 *   답글        openPostEditor({ post: {...새 글, isNew:true} })
 *   푸터 문의   window.zbdSite.openDrawer(groupKey, groupSubKey, 제목) — 비로그인(guest) 저장, 연락처 칸
 * 휴대폰 뒤로가기 = 창 닫기(예전처럼 history 에 한 칸 쌓는다).
 */
import { api } from '#lib/shared/api.ts';
import { HOSTING_ID } from '#lib/shared/config.ts';
import { alert } from '#lib/shared/ui/dialog.svelte.ts';
import { findPostGroup } from '../board/board.svelte.ts';

export const jsonDefault = { body: '' };
export const postDefault = { postId: '', groupKey: '', groupSubKey: '', title: '', json: jsonDefault, memo: '', status: 0 };
export const userPostStatusMap: Record<string, string> = { 0: '공개', 1: '비공개', 2: '...' };

export type EditorOpen = {
	post?: any;
	groupKey?: string;
	groupSubKey?: string;
	headTitle?: string;
	isGnbPostWriteWidget?: boolean;
	placeholder?: string;
	onSave?: (p: any) => void;
};

export const postEditor = $state({
	open: false,
	post: null as any,
	/** 처음 연 모양(바뀐 것이 있는지 비교) */
	orig: '' as string,
	headTitle: '',
	groupKey: '',
	placeholder: '',
	isGnbPostWriteWidget: false,
	onSave: null as ((p: any) => void) | null,
	/** 편집기를 새로 그리게(다른 글을 열 때) */
	seq: 0
});

const parse = (v: any) => (v ? (typeof v == 'string' ? JSON.parse(v) : v) : null);
const newId = () => Date.now().toString(36);

/** 받은 글을 창 모양으로(jsonUser·json 풀기) */
export function normalize(o: any) {
	o.jsonUser = parse(o.jsonUser) || { ...jsonDefault };
	o.json = parse(o.json) || {};
	o.memo = o.jsonUser.memo || '';
	return o;
}

function show(post: any, o: EditorOpen) {
	postEditor.post = post;
	postEditor.orig = JSON.stringify(post);
	postEditor.headTitle = o.headTitle || '';
	postEditor.groupKey = o.groupKey || post?.groupKey || '';
	postEditor.placeholder = o.placeholder || '';
	postEditor.isGnbPostWriteWidget = !!o.isGnbPostWriteWidget;
	postEditor.onSave = o.onSave || null;
	postEditor.seq++;
	if (!postEditor.open) {
		history.pushState({ ...(history.state || {}), post_open: true }, '');
		addEventListener('popstate', onPop);
	}
	postEditor.open = true;
}

function onPop() {
	removeEventListener('popstate', onPop);
	postEditor.open = false;
	postEditor.post = null;
}

/** 창 닫기 — 쌓아 둔 한 칸을 걷어 낸다(그 popstate 가 실제로 닫는다) */
export function closePostEditor() {
	if (history.state?.post_open) history.back();
	else onPop();
}

export async function openPostEditor(o: EditorOpen = {}) {
	const p = o.post;
	if (!p) {
		const m: any = { ...structuredClone(postDefault), postId: newId(), groupKey: o.groupKey || '', ...(o.groupSubKey ? { groupSubKey: o.groupSubKey } : {}) };
		const pg = findPostGroup(o.groupKey);
		if (typeof pg?.component?.defaultStatus == 'number') m.status = pg.component.defaultStatus;
		if (pg?.component?.defaultJson && typeof pg.component.defaultJson == 'object') {
			const col = pg.dataColumn || 'json';
			m[col] = { ...(m[col] || {}), ...pg.component.defaultJson };
		}
		m.jsonUser = { ...jsonDefault };
		show(m, o);
		return;
	}
	if (p.isNew || p.isFooterInquiry) {
		show({ jsonUser: { ...jsonDefault }, ...structuredClone($state.snapshot(p)) }, o);
		return;
	}
	const r = await api.get(`/service/user/post/${p.postId}`);
	if (!r.ok) {
		alert('데이터 조회 도중 오류가 발생하였습니다.');
		return;
	}
	show(normalize(r.data), o);
}

/** 푸터 「문의하기」 등 — 운영자 HTML 이 window.zbdSite.openDrawer 로 부른다 */
export function openFooterInquiry(groupKey: string, groupSubKey: string, headTitle = '') {
	return openPostEditor({
		headTitle,
		post: { ...structuredClone(postDefault), postId: newId(), groupKey, groupSubKey, hostingId: HOSTING_ID, isFooterInquiry: true, status: 0, jsonUser: { ...jsonDefault } }
	});
}
