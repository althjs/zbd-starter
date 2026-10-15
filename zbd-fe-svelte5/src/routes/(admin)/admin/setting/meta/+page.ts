import { redirect } from '@sveltejs/kit';

// React pages/admin/setting/meta/index — 메타 첫 화면은 「home」 탭
export function load() {
	redirect(307, '/admin/setting/meta/home');
}
