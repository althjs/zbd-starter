import { error } from '@sveltejs/kit';

// 개발 서버(localhost)에서만 — 배포 빌드에선 없는 주소
export function load() {
	if (!import.meta.env.DEV) error(404, 'Not Found');
}
