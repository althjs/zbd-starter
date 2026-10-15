/** 마이페이지 숫자(작성한 글·댓글 수 …) — 예전 pages/user/zustandUserSummaryStore.ts */
import { api } from '#lib/shared/api.ts';
import { HOSTING_ID } from '#lib/shared/config.ts';

export const summary = $state({ postCount: 0, commentCount: 0, loaded: false });

export async function loadSummary() {
	if (summary.loaded) return;
	const r = await api.get(`/service/user/${HOSTING_ID}/summary`);
	if (r.ok) Object.assign(summary, r.data || {}, { loaded: true });
}
