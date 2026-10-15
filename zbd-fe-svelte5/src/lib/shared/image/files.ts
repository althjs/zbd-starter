/**
 * 업로드 파일 다루기 — 예전 ImagePicker.tsx 의 uploadFileDirect · store/zustandShopImage.ts(폴더 목록) · 삭제.
 *   회원(userId 있음) = /service/user/file/… (서버가 토큰의 회원 폴더 u_{userId}_{folder} 로만 — 2026-09-26 보안)
 *   관리자            = /service/admin/file/…
 */
import { api } from '../api';

export type FolderOpts = { userId?: number | null; hostingId?: string; preFolder?: string; subFolder?: string };
/** pendingVideo = 올렸지만 아직 인코딩 중인 동영상의 표지(서버 목록의 pendingVideos — 2026-10-04) — 고르지 못하게 「변환 중」 으로 */
export type ImageFile = { name: string; size?: { width: any; height: any }; url?: string; imageId?: string; geo?: any; pendingVideo?: boolean };

const q = (o: FolderOpts, extra: Record<string, string> = {}) => {
	const p = new URLSearchParams(extra);
	if (o.hostingId) p.set('hostingId', o.hostingId);
	if (o.userId) p.set('userId', String(o.userId));
	if (o.subFolder) p.set('subFolder', o.subFolder);
	if (o.preFolder) p.set('preFolder', o.preFolder);
	const s = p.toString();
	return s ? `?${s}` : '';
};

/** 폴더의 파일 목록(이름에서 크기 접두사 `300x300_` 를 읽어 size 로) */
export async function listFolder(folder: string, o: FolderOpts = {}): Promise<ImageFile[]> {
	const r = await api.get(`/service/${o.userId ? 'user' : 'admin'}/file/${folder}${q(o)}`);
	if (!r.ok) return [];
	// 관리자 API 는 {files:[이름…]}, 회원 API 는 배열로 준다
	const list: any[] = Array.isArray(r.data) ? r.data : r.data?.files || [];
	const pending = new Set<string>(Array.isArray(r.data?.pendingVideos) ? r.data.pendingVideos : []);
	return list.map((x: any) => {
		const f: ImageFile = typeof x == 'string' ? { name: x } : { ...x };
		if (pending.has(f.name)) f.pendingVideo = true;
		if (!f.size) {
			const m = /^(\d+|all)x(\d+|auto)_/.exec(f.name);
			f.size = m ? { width: isNaN(+m[1]) ? m[1] : +m[1], height: isNaN(+m[2]) ? m[2] : +m[2] } : { width: 'all', height: 'auto' };
		}
		return f;
	});
}

/** 올리기 — files: [{ blob, name }] → 저장된 이름 목록 */
export async function uploadFiles(files: { blob: Blob; name: string }[], folder: string, o: FolderOpts = {}): Promise<ImageFile[] | null> {
	const fd = new FormData();
	for (const f of files) fd.append('files[]', f.blob, f.name || 'unknown.jpg');
	const r = await api.post(`/service/${o.userId ? 'user' : 'admin'}/file/upload${q(o, { folder })}`, fd, o.userId ? { headers: { userId: String(o.userId) } } : {});
	if (!r.ok) return null;
	return (r.data || []).map((img: any) => ({ name: img.newFilename, ...(img.geo ? { geo: img.geo } : {}) }));
}

export async function deleteFile(folder: string, name: string, o: FolderOpts = {}) {
	const r = await api.delete(`/service/${o.userId ? 'user' : 'admin'}/file/${folder}${q(o, { name })}`);
	return r.ok;
}

export function blobToDataUrl(b: Blob): Promise<string> {
	return new Promise((res) => {
		const r = new FileReader();
		r.onload = () => res(String(r.result));
		r.readAsDataURL(b);
	});
}

export function loadImage(src: string): Promise<HTMLImageElement> {
	return new Promise((res, rej) => {
		const img = new Image();
		img.crossOrigin = 'anonymous';
		img.onload = () => res(img);
		img.onerror = rej;
		img.src = src;
	});
}

/** 캔버스 → Blob */
export function canvasBlob(c: HTMLCanvasElement, mime = 'image/webp', quality = 1): Promise<Blob> {
	return new Promise((res) => c.toBlob((b) => res(b!), mime, quality));
}

/** 폭에 맞춰 줄이기(height auto) — 원본이 더 작고 webp/png 면 그대로 */
export async function resizeToWidth(file: Blob, width: number, mime = 'image/webp'): Promise<Blob> {
	const img = await loadImage(await blobToDataUrl(file));
	const ratio = width / img.width;
	if (ratio > 1 && /webp|png/.test(file.type)) return file;
	const c = document.createElement('canvas');
	c.width = width;
	c.height = Math.round(img.height * ratio);
	c.getContext('2d')!.drawImage(img, 0, 0, c.width, c.height);
	return canvasBlob(c, mime);
}

export const extOf = (mimeOrName: string) => {
	const m = /image\/(\w+)/.exec(mimeOrName);
	if (m) return m[1] == 'jpeg' ? 'jpg' : m[1].replace('svg+xml', 'svg');
	const e = /\.(\w+)(\?|$)/.exec(mimeOrName);
	return e ? e[1].toLowerCase() : '';
};
