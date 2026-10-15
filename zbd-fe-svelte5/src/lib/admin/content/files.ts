/**
 * 보안 파일(첨부) 올리기·받기·지우기 — React SecureFileUpload·SecureFileDownload 와 같은 API.
 *   올리기  POST /service/user/file/secure-upload[?hostingId=] (FormData files·type·fileName·imageId)
 *   받기    POST /service/user/file/secure-download {imageId, fileName} → blob
 *   지우기  DELETE /service/user/file/secure-download?imageId=&fileName=
 * imageId 는 `{호스팅}/{폴더}` (호스팅이 있을 때).
 */
import { api, getBrowserUid, getToken } from '#lib/shared/api.ts';
import { fileDownload } from './dates.ts';

export async function secureUpload(file: File, imageId: string, hostingId = '', type = 'account') {
	const fd = new FormData();
	fd.append('files', file, file.name || 'unknown.jpg');
	fd.append('type', type);
	fd.append('fileName', file.name);
	fd.append('imageId', imageId);
	return api.post(`/service/user/file/secure-upload${hostingId ? `?hostingId=${hostingId}` : ''}`, fd, { headers: { imageId } });
}

/** blob 을 POST 로 받는다(api.blob 은 GET 만이라 여기서 직접) */
export async function secureDownload(imageId: string, fileName: string) {
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	const t = getToken();
	if (t) headers.Authorization = `Bearer ${t}`;
	const uid = getBrowserUid();
	if (uid) headers.zbduid = uid;
	const res = await fetch((import.meta.env.VITE_API_BASE || '') + '/service/user/file/secure-download', {
		method: 'POST',
		headers,
		credentials: 'include',
		body: JSON.stringify({ imageId, fileName })
	});
	if (!res.ok) return false;
	fileDownload(URL.createObjectURL(await res.blob()), fileName);
	return true;
}

export const secureDelete = (imageId: string, fileName: string) => api.delete(`/service/user/file/secure-download?imageId=${encodeURIComponent(imageId)}&fileName=${fileName}`);
