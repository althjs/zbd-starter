<script lang="ts">
	/**
	 * 링크 첨부 — 주소를 붙여 넣으면 서버가 OG 정보를 읽어 카드 HTML 을 만든다(박스형·가로형). 예전 modal/ModalLinkAttach.tsx
	 * 그림은 서버가 base64 로 주고, 글 폴더(imageId) 에 올려 그 주소를 쓴다(남의 서버 그림을 직접 걸지 않게).
	 */
	import { api } from '../api';
	import { imagePrefix } from '../config';
	import { uploadFiles } from '../image/files';
	import MuiButton from '../mui/MuiButton.svelte';
	import TextField from '../mui/TextField.svelte';
	import { portal } from '../ui/portal';
	import Scrollbar from '../ui/Scrollbar.svelte';
	import { alert, toast } from '../ui/dialog.svelte.ts';

	let {
		open,
		imageId = '',
		userId = null,
		isMobile = false,
		onclose,
		onattach
	}: { open: boolean; imageId?: string; userId?: number | null; isMobile?: boolean; onclose: () => void; onattach: (html: string) => void } = $props();

	let url = $state('');
	let link = $state({ title: '', description: '', base64Image: '', imageUrl: '' });
	let design = $state<'BOX' | 'HORIZONTAL'>('BOX');
	let loading = $state(false);

	$effect(() => {
		if (!open) {
			url = '';
			link = { title: '', description: '', base64Image: '', imageUrl: '' };
		}
	});

	const esc = (s: string) => String(s || '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]!);
	const host = (u: string) => {
		try {
			return new URL(u).hostname;
		} catch {
			return '';
		}
	};

	async function upload(b64: string) {
		try {
			const blob = await (await fetch(b64)).blob();
			const name = `link-image-${Date.now().toString(36)}.png`;
			const folder = userId ? imageId.replace(`u_${userId}_`, '') : imageId;
			const r = await uploadFiles([{ blob, name }], folder, { userId });
			const saved = r?.[0]?.name || name;
			link = { ...link, base64Image: '', imageUrl: `${imagePrefix()}/${imageId}/${saved}` };
		} catch {
			toast('이미지 업로드에 실패했습니다.');
		}
	}

	async function parse(u: string) {
		if (!/^https?:\/\//.test(u)) return;
		loading = true;
		const r = await api.get(`/service/seo/parse-open-graph?url=${encodeURIComponent(u)}`);
		loading = false;
		if (!r.ok) {
			alert('링크정보를 가져오는 도중 오류가 발생하였습니다.');
			return;
		}
		link = { title: r.data?.title || '', description: r.data?.description || '', base64Image: r.data?.base64Image || '', imageUrl: '' };
		if (link.base64Image && imageId) upload(link.base64Image);
	}

	function html() {
		const img = link.imageUrl || link.base64Image;
		const t = esc(link.title);
		const d = esc(link.description);
		const u = esc(url);
		const h = esc(host(url));
		if (design == 'HORIZONTAL') {
			return `<div class="og-link-container horizental" style="border: 1px solid #e1e5e9; border-radius: 8px; max-width: 700px; overflow: hidden; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; height: 140px; cursor: pointer;">${img ? `<img src="${esc(img)}" style="width: 200px; height: 140px; object-fit: cover; flex-shrink: 0;" alt="Link preview" />` : ''}<div class="og-link-content" style="flex: 1; padding: 16px; display: flex; flex-direction: column; justify-content: space-between;"><div><h3 style="margin: 0 0 8px 0; font-size: 16px; font-weight: 600; color: #1d2129; line-height: 1.3; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">${t}</h3>${d ? `<p class="og-link-desc" style="margin: 0 0 8px 0; font-size: 14px; color: #606770; line-height: 1.4; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">${d}</p>` : ''}</div><span class="og-link-host" style="font-size: 12px; color: #8a8d91; text-transform: uppercase; margin-top: auto;"><a href="${u}" target="_blank" rel="noopener noreferrer" style="color: inherit; text-decoration: none;">${h}</a></span></div></div>`;
		}
		return `<div class="og-link-container box" style="border: 1px solid #e1e5e9; border-radius: 8px; max-width: 500px; overflow: hidden; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; cursor: pointer;">${img ? `<img src="${esc(img)}" style="width: 100%; height: 200px; object-fit: cover;" alt="Link preview" />` : ''}<div class="og-link-box" style="padding: 16px;"><h3 style="margin: 0 0 8px 0; font-size: 16px; font-weight: 600; color: #1d2129; line-height: 1.3;">${t}</h3>${d ? `<p class="og-link-desc" style="margin: 0 0 8px 0; font-size: 14px; color: #606770; line-height: 1.4;">${d}</p>` : ''}<span class="og-link-host" style="font-size: 12px; color: #8a8d91; text-transform: uppercase;"><a href="${u}" target="_blank" rel="noopener noreferrer" style="color: inherit; text-decoration: none;">${h}</a></span></div></div>`;
	}
</script>

{#if open}
	<div use:portal role="presentation" class="MuiModal-root" style="z-index:1400">
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true" onclick={onclose}></div>
		<div class="la-box" class:mobile={isMobile}>
			<Scrollbar style="max-height:90vh">
				<div style="padding:24px 24px 0">
					<h6 class="MuiTypography-root MuiTypography-h6" style="margin-bottom:16px">첨부할 링크를 추가하세요</h6>
					<TextField
						bind:value={url}
						placeholder="https://example.com"
						disabled={loading}
						sx={{ marginBottom: 16 }}
						onpaste={(e) => {
							const t = e.clipboardData?.getData('text') || '';
							if (/^https?:\/\//.test(t)) setTimeout(() => parse(t), 100);
						}}
						onblur={() => isMobile && parse(url)}
					/>
					{#if link.title}
						<div style="margin-bottom:16px">
							<p class="MuiTypography-root MuiTypography-subtitle2" style="margin-bottom:8px">미리보기</p>
							<div style="display:flex;gap:16px;margin-bottom:8px">
								<label><input type="radio" bind:group={design} value="BOX" /> 박스형</label>
								<label><input type="radio" bind:group={design} value="HORIZONTAL" /> 가로형</label>
							</div>
							<div class="zbd-raw preview">{@html html()}</div>
						</div>
					{/if}
				</div>
				<div style="display:flex;gap:8px;justify-content:flex-end;padding:16px 24px 24px">
					<MuiButton variant="outlined" onclick={onclose}>취소</MuiButton>
					<MuiButton variant="contained" disabled={!url || !link.title || loading} onclick={() => onattach(html())}>첨부</MuiButton>
				</div>
			</Scrollbar>
		</div>
	</div>
{/if}

<style>
	.la-box {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		width: 500px;
		max-height: 90%;
		overflow: hidden;
		background: #fff;
		border: 2px solid #000;
		box-shadow:
			0px 11px 15px -7px rgba(0, 0, 0, 0.2),
			0px 24px 38px 3px rgba(0, 0, 0, 0.14),
			0px 9px 46px 8px rgba(0, 0, 0, 0.12);
	}
	.la-box.mobile {
		width: 90%;
		border: none;
	}
	.preview :global(.og-link-container) {
		pointer-events: none;
	}
</style>
