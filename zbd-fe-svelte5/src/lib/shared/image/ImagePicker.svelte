<script lang="ts">
	/**
	 * 사진 고르기 — 예전 components/common/image/ImagePicker.tsx (관리자·사용자 공용, 같은 이름).
	 *
	 *   단추(before/after) → 창: [첨부(누르기·끌어 놓기·붙여넣기)] [이 폴더에 올린 사진 목록 → 선택·삭제] [Unsplash]
	 *   올릴 때 크기 맞추기: width×height 고정 = 자르기 창(autoCrop 이면 가운데 자동) · height auto = 폭에 맞춰 줄이기 · all = 원본
	 *   파일 이름 규칙 `{w}x{h}_{시각}.{확장자}` — 서버·다른 화면이 이 접두사로 크기를 읽는다.
	 *   회원(userId) 이면 회원 폴더 u_{userId}_{folder}(서버가 토큰으로 강제), 아니면 관리자 폴더.
	 *
	 * 밖에서 열기: bind:this={picker} → picker.open({ group }) (React 판 ref.open 과 같다)
	 * 관리자 전용 「공유 이미지(미디어)」 탭·여권 꼭짓점 자르기는 관리자 화면을 옮길 때 더한다.
	 */
	import type { Snippet } from 'svelte';
	import { imagePrefix } from '../config';
	import Dialog from '../mui/Dialog.svelte';
	import Icon from '../mui/Icon.svelte';
	import IconButton from '../mui/IconButton.svelte';
	import MuiButton from '../mui/MuiButton.svelte';
	import SvgIcon from '../mui/SvgIcon.svelte';
	import { alert, confirm, toast } from '../ui/dialog.svelte.ts';
	import { viewport } from '../media.svelte.ts';
	import CropDialog from './CropDialog.svelte';
	import { deleteFile, extOf, listFolder, resizeToWidth, uploadFiles, type ImageFile } from './files';
	import Unsplash from './Unsplash.svelte';
	import { unsplashKey } from './unsplash-key.svelte.ts';

	type Size = { width: number | 'all'; height: number | 'auto' };
	let {
		size,
		image = null,
		onSelect,
		folder,
		preFolder = '',
		subFolder = '',
		title = '',
		isMultiple = false,
		readonly = false,
		useUnsplash = false,
		accept = '.jpg, .jpeg, .png, .webp, .gif, .svg',
		multiple = true,
		disableCrop = false,
		newImageOnly = false,
		useDelete = false,
		useOriginalFileName = false,
		cropImageMimeType = 'image/webp',
		autoCrop = false,
		showToast = true,
		maxWidth = 0,
		userId = null,
		hostingId = '',
		before,
		after
	}: {
		size: Size;
		image?: { name?: string } | null | '';
		onSelect: (img: any, params?: any) => void;
		folder: string;
		preFolder?: string;
		subFolder?: string;
		title?: string;
		isMultiple?: boolean;
		readonly?: boolean;
		useUnsplash?: boolean;
		accept?: string;
		multiple?: boolean;
		disableCrop?: boolean;
		newImageOnly?: boolean;
		useDelete?: boolean;
		useOriginalFileName?: boolean;
		cropImageMimeType?: string;
		autoCrop?: boolean;
		showToast?: boolean;
		maxWidth?: number;
		userId?: number | null;
		hostingId?: string;
		before?: Snippet;
		after?: Snippet;
	} = $props();

	let open = $state(false);
	let tab = $state<'ATTACH' | 'UNSPLASH'>('ATTACH');
	// 사이트 설정에 키가 없으면 Unsplash 탭을 숨긴다
	const uKey = $derived(unsplashKey());
	let files = $state<ImageFile[] | null>(null);
	let selected = $state<string[]>([]);
	let uploading = $state(false);
	let params = $state<any>(null);
	let cropFile = $state<Blob | null>(null);
	let cropResolve: ((b: Blob | null) => void) | null = null;
	let input = $state<HTMLInputElement>();

	const opts = $derived({ userId, hostingId, preFolder, subFolder });
	const urlOf = (name: string) =>
		`${imagePrefix(hostingId || undefined)}${preFolder ? `/${preFolder}` : ''}/${userId ? `u_${userId}_` : ''}${folder}/${subFolder ? `${subFolder}/` : ''}${encodeURIComponent(name)}`;

	async function load(force = false) {
		if (newImageOnly) {
			files = [];
			return;
		}
		if (files && !force) return;
		files = await listFolder(folder, opts);
	}

	// 인코딩된 영상(…_480.mp4)이 있으면 그 표지 그림(….webp)은 숨긴다 — 같은 장면이 사진·영상으로 두 번 보여 표지를 사진으로 고르는 일이 있었다(2026-10-04)
	const videoPosters = $derived(new Set((files || []).filter((f) => /_480\.mp4$/.test(f.name)).map((f) => f.name.replace('_480.mp4', '.webp'))));
	/** 이 크기에 맞는 것만(+ 그룹) — 이름순 */
	const images = $derived(
		(files || [])
			.filter((f) => accept.includes(extOf(f.name)) || /\.mp4$/.test(f.name))
			.filter((f) => !videoPosters.has(f.name) && !/_share\.jpg$/.test(f.name)) // 동영상 공유용 정사각 그림도 숨김
			.filter((f) => size.width == 'all' || (f.size?.width == size.width && (size.height == 'auto' || f.size?.height == size.height)))
			.filter((f) => !params?.group || f.name.includes(`_group${params.group}`))
			.sort((a, b) => a.name.localeCompare(b.name))
	);

	export function open_(p: any = null) {
		params = p;
		selected = p?.selectedImageNames || [];
		open = true;
		load();
	}
	export { open_ as open };
	export function close() {
		open = false;
		selected = [];
	}

	const stamp = () => Date.now();
	const groupTag = () => (params?.group ? `_group${params.group}` : '');
	const sizeTag = () => `${size.width}x${size.height}`;

	function crop(file: Blob): Promise<Blob | null> {
		return new Promise((res) => {
			cropResolve = res;
			cropFile = file;
		});
	}

	async function prepare(f: File | Blob, origName = ''): Promise<{ blob: Blob; name: string } | null> {
		const name = origName || (f as File).name || 'image.png';
		const ext = extOf(f.type) || extOf(name) || 'jpg';
		const base = useOriginalFileName && name != 'image.png' ? name.replace(/\.\w+$/, '') : String(stamp());
		if (disableCrop) {
			// 묶음(group) 표시는 원래 이름으로 올릴 때도 붙인다 — 목록이 `_group…` 이 든 파일만 보여서, 빠지면 올린 그림이 안 보였다(2026-10-04 GNB 자유 배너)
			const nm = useOriginalFileName ? `${name.replace(/\.\w+$/, '')}${groupTag()}.${extOf(name) || ext}` : `${stamp()}${groupTag()}.${ext}`;
			return { blob: f, name: `${size.width == 'all' ? '' : sizeTag() + '_'}${nm}` };
		}
		if (f.size > 7000000 && !(autoCrop && maxWidth > 0)) {
			await alert('5MB 이하 파일만 업로드 가능합니다.');
			return null;
		}
		if (ext == 'svg') return { blob: f, name: `${sizeTag()}_${base}${groupTag()}.svg` };
		if ((size.width != 'all' && size.height == 'auto') || (autoCrop && maxWidth > 0)) {
			const b = await resizeToWidth(f, size.width == 'all' ? maxWidth : (size.width as number), cropImageMimeType);
			return { blob: b, name: `${sizeTag()}_${base}${groupTag()}.${extOf(b.type) || ext}` };
		}
		if (size.width != 'all') {
			const b = autoCrop ? await crop(f) : await crop(f);
			if (!b) return null;
			return { blob: b, name: `${sizeTag()}_${base}${groupTag()}.${extOf(b.type) || 'webp'}` };
		}
		return { blob: f, name: `${base}${groupTag()}.${ext}` };
	}

	async function addFiles(list: FileList | File[] | Blob[], names: string[] = []) {
		const out: { blob: Blob; name: string }[] = [];
		const arr = Array.from(list as any) as Blob[];
		for (let i = 0; i < arr.length; i++) {
			const f = arr[i];
			if (f.type == 'video/mp4') {
				out.push({ blob: f, name: `${sizeTag()}_${stamp()}.mp4` });
				continue;
			}
			const p = await prepare(f, names[i]);
			if (p) out.push(p);
		}
		if (!out.length) return;
		uploading = true;
		const saved = await uploadFiles(out, folder, opts);
		uploading = false;
		if (!saved) {
			alert('업로드 도중 오류가 발생하였습니다.');
			return;
		}
		// 동영상은 서버가 받아 두고 맥미니 인코딩(1분마다, 하나씩)이 재생용(…_480.mp4)을 만든 뒤에야 고를 수 있다 —
		// 창에 남아 있어 봐야 표지 그림만 보여 헷갈리므로 안내하고 닫는다(2026-10-04 대표). 다음에 열면 목록을 새로 받는다.
		const hasVideo = out.some((o) => /\.mp4$/i.test(o.name));
		if (hasVideo && !newImageOnly) {
			files = null;
			close();
			alert({
				title: '동영상을 올렸어요',
				body: '재생할 수 있게 바꾸는 데 몇 분 걸려요(짧은 영상은 1~3분, 긴 영상은 더).\n끝나면 이 사진 고르기에서 동영상을 고를 수 있어요.'
			});
			return;
		}
		if (showToast) toast('정상적으로 업로드 되었습니다');
		if (newImageOnly) {
			onSelect(multiple ? saved : saved[0], params);
			open = false;
			if (hasVideo) toast('동영상은 재생할 수 있게 바꾸는 데 몇 분 걸려요');
			return;
		}
		files = [...(files || []), ...out.map((o) => ({ name: o.name, size: size as any }))];
		tab = 'ATTACH';
	}

	async function remove(f: ImageFile) {
		if (!(await confirm('이미지를 삭제하시겠습니까?'))) return;
		if (!(await deleteFile(folder, f.name, opts))) {
			alert('삭제 도중 오류가 발생하였습니다.');
			return;
		}
		await load(true);
		if (showToast) toast('정상적으로 삭제되었습니다.');
	}

	function paste(e: ClipboardEvent) {
		const fs = [...(e.clipboardData?.items || [])]
			.filter((i) => i.kind == 'file')
			.map((i) => i.getAsFile()!)
			.filter(Boolean);
		if (fs.length) addFiles(fs);
	}
	const heading = $derived(title || `${size.width}x${size.height} 사이즈 이미지를 선택하세요`);
	const isVideo = (n?: string) => /\.mp4$/.test(n || '');
</script>

<!-- 단추 -->
<div class="MuiStack-root" style="width:fit-content">
	{#if image && (image as any).name}
		{#if after}
			<div class="MuiBox-root" role="presentation" onclick={() => !readonly && open_()}>{@render after()}</div>
		{:else}
			<div class="MuiBox-root" style="position:relative">
				{#if isVideo((image as any).name)}
					<video class="MuiCardMedia-root MuiCardMedia-media" style="height:50px" src={urlOf((image as any).name)} onclick={() => !readonly && open_()}></video>
				{:else}
					<img
						class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img"
						style="aspect-ratio:1;width:50px;height:50px"
						src={urlOf((image as any).name)}
						alt=""
						onclick={() => !readonly && open_()}
					/>
				{/if}
			</div>
		{/if}
	{:else if before}
		<span style="cursor:pointer" role="button" tabindex="0" onclick={() => !readonly && open_()} onkeydown={() => {}}>{@render before()}</span>
	{:else}
		<MuiButton onclick={() => !readonly && open_()}>선택</MuiButton>
	{/if}
</div>

<!-- 휴대폰 = 전체 화면 + 사진 바둑판(꽉 차게, 2026-10-04 대표) · PC = 표 -->
<Dialog {open} onclose={close} fullScreen={viewport.isMobile} class={viewport.isMobile ? 'zip-m' : ''} maxWidth={1000} paperSx={{ width: '100%' }} fitWidth={false} onpaste={paste}>
	<!-- 창 제목 snippet 이름이 prop title 을 가리므로(안에서 title = 이 snippet 자신 → invalid_snippet_arguments, 500) 값은 heading 으로 -->
	{#snippet title()}{heading}{/snippet}
	{#if useUnsplash && uKey}
		<div class="tabs">
			<button class:on={tab == 'ATTACH'} onclick={() => (tab = 'ATTACH')}>내 사진</button>
			<button class:on={tab == 'UNSPLASH'} onclick={() => (tab = 'UNSPLASH')}>Unsplash 무료 사진</button>
		</div>
	{/if}
	{#if tab == 'UNSPLASH' && uKey}
		<Unsplash accessKey={uKey} onpick={(b, name) => addFiles([b], [name])} />
	{:else}
		<div
			class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root drop"
			role="presentation"
			ondragover={(e) => e.preventDefault()}
			ondrop={(e) => {
				e.preventDefault();
				if (e.dataTransfer?.files) addFiles(e.dataTransfer.files);
			}}
		>
			<label style="cursor:pointer;display:block">
				<p class="MuiTypography-root MuiTypography-body2" style="color:#2f3746">
					<Icon name="attach_file" /><br />
					이미지 첨부는 클릭하거나, Drag & Drop 하세요<br />
					{#if size.width != 'all' && size.height != 'auto'}{size.width}x{size.height} 사이즈 이미지{/if}
					{#if size.width != 'all' && size.height == 'auto'}가로 {size.width}px 이미지{/if}
					{#if useOriginalFileName}원본 파일명으로 업로드 됩니다{/if}
				</p>
				<input bind:this={input} type="file" style="display:none" {multiple} {accept} onchange={(e) => (e.currentTarget.files && addFiles(e.currentTarget.files), (e.currentTarget.value = ''))} />
			</label>
			{#if uploading}<div class="busy"><span class="MuiCircularProgress-root spin"></span></div>{/if}
		</div>

		{#if !newImageOnly && viewport.isMobile}
			<div class="grid">
				{#each images as f (f.name)}
					{@const on = selected.includes(f.name)}
					<div class="tile" class:on class:pending={f.pendingVideo}>
						<button
							type="button"
							class="pick"
							disabled={f.pendingVideo}
							aria-label={isMultiple ? '고르기' : '선택'}
							onclick={() => {
								if (!isMultiple) return void (onSelect({ ...f, imageId: folder }, params), close());
								selected = on ? selected.filter((n) => n != f.name) : [...selected, f.name];
							}}
						>
							<img src={urlOf(isVideo(f.name) ? f.name.replace('_480.mp4', '.webp') : f.name)} alt="" loading="lazy" />
							{#if isVideo(f.name)}<span class="vid-badge" aria-label="동영상"><svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M8 5v14l11-7z" /></svg></span>{/if}
							{#if f.pendingVideo}<span class="tile-note">동영상 변환 중</span>{/if}
							{#if isMultiple}<span class="check" aria-hidden="true">{on ? selected.indexOf(f.name) + 1 : ''}</span>{/if}
						</button>
						{#if useDelete}
							<button type="button" class="del" aria-label="삭제" onclick={() => remove(f)}><SvgIcon name="Delete" size="Small" /></button>
						{/if}
					</div>
				{:else}
					<p class="empty">{files ? '올린 사진이 없습니다' : '불러오는 중…'}</p>
				{/each}
			</div>
		{:else if !newImageOnly}
			<table class="MuiTable-root list">
				<tbody class="MuiTableBody-root">
					{#each images as f (f.name)}
						<tr class="MuiTableRow-root">
							{#if isMultiple}
								<td class="MuiTableCell-root MuiTableCell-body" style="width:40px">
									<input
										type="checkbox"
										disabled={f.pendingVideo}
										checked={selected.includes(f.name)}
										onchange={(e) => (selected = e.currentTarget.checked ? [...selected, f.name] : selected.filter((n) => n != f.name))}
									/>
								</td>
							{/if}
							<td class="MuiTableCell-root MuiTableCell-body" style="width:90px">
								{#if isVideo(f.name)}<span class="vid"
										><video src={urlOf(f.name)} poster={urlOf(f.name.replace('_480.mp4', '.webp'))} preload="metadata" muted playsinline style="width:70px"></video><span
											class="vid-badge"
											aria-label="동영상"><svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M8 5v14l11-7z" /></svg></span
										></span
									>{:else}<img
										src={urlOf(f.name)}
										alt=""
										style="width:{size.width != 'all' && size.height != 'auto' && size.width > size.height ? 70 : 50}px;max-width:none;{f.pendingVideo ? 'opacity:.45' : ''}"
									/>{/if}
								{#if f.pendingVideo}<div class="pending">동영상 변환 중</div>{/if}
							</td>
							{#if !viewport.isMobile}<td class="MuiTableCell-root MuiTableCell-body name">{params?.group ? f.name.replace(`_group${params.group}_`, '') : f.name}</td>{/if}
							{#if !isMultiple}
								<td class="MuiTableCell-root MuiTableCell-body" style="width:100px;text-align:center">
									<MuiButton disabled={f.pendingVideo} onclick={() => (onSelect({ ...f, imageId: folder }, params), close())}>{f.pendingVideo ? '변환 중' : '선택'}</MuiButton>
								</td>
							{/if}
							{#if useDelete}
								<td class="MuiTableCell-root MuiTableCell-body" style="width:60px;text-align:center">
									<IconButton size="Small" onclick={() => remove(f)}><SvgIcon name="Delete" size="Small" /></IconButton>
								</td>
							{/if}
						</tr>
					{:else}
						<tr><td class="MuiTableCell-root MuiTableCell-body" style="text-align:center;color:#98a2b3">{files ? '올린 사진이 없습니다' : '불러오는 중…'}</td></tr>
					{/each}
				</tbody>
			</table>
		{/if}
	{/if}
	{#snippet actions()}
		<MuiButton onclick={close}>닫기</MuiButton>
		{#if isMultiple}
			<MuiButton variant="contained" disabled={!selected.length} onclick={() => (onSelect(selected.map((n) => images.find((x) => x.name == n)).filter(Boolean), params), close())}
				>{selected.length}개 선택</MuiButton
			>
		{/if}
	{/snippet}
</Dialog>

<CropDialog
	file={cropFile}
	width={size.width as number}
	height={size.height as number}
	mime={cropImageMimeType}
	{autoCrop}
	ondone={(b) => {
		cropFile = null;
		cropResolve?.(b);
		cropResolve = null;
	}}
/>

<style>
	.tabs {
		display: flex;
		gap: 4px;
		margin-bottom: 12px;
		border-bottom: 1px solid #eaecf0;
	}
	.tabs button {
		padding: 8px 14px;
		font-size: 14px;
		font-weight: 600;
		color: #667085;
		border-bottom: 2px solid transparent;
		cursor: pointer;
	}
	.tabs button.on {
		color: var(--theme-color-main);
		border-bottom-color: var(--theme-color-main);
	}
	.drop {
		position: relative;
		padding: 24px;
		text-align: center;
	}
	.busy {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.7);
	}
	.spin {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 3px solid var(--theme-color-alpha12);
		border-top-color: var(--theme-color-main);
	}
	.list {
		margin-top: 12px;
	}
	/* 휴대폰 — 전체 화면 창 · 여백을 줄이고 사진은 3칸 바둑판 */
	:global(.zip-m .MuiDialogContent-root) {
		padding: 0 8px 8px;
	}
	:global(.zip-m) .tabs {
		margin-bottom: 8px;
	}
	:global(.zip-m) .drop {
		padding: 12px;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 4px;
		margin-top: 8px;
	}
	.tile {
		position: relative;
		aspect-ratio: 1;
		background: #f2f4f7;
		border-radius: 4px;
		overflow: hidden;
	}
	.pick {
		all: unset;
		position: absolute;
		inset: 0;
		cursor: pointer;
	}
	.pick img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.tile.on .pick::after {
		content: '';
		position: absolute;
		inset: 0;
		border: 3px solid var(--theme-color-main, #1d433d);
		border-radius: 4px;
		background: rgba(255, 255, 255, 0.15);
	}
	.tile.pending img {
		opacity: 0.45;
	}
	.check {
		position: absolute;
		top: 6px;
		left: 6px;
		width: 22px;
		height: 22px;
		border-radius: 50%;
		border: 2px solid #fff;
		background: rgba(0, 0, 0, 0.25);
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		box-sizing: border-box;
	}
	.tile.on .check {
		background: var(--theme-color-main, #1d433d);
	}
	.tile-note {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 4px;
		text-align: center;
		font-size: 11px;
		color: var(--theme-color-main);
	}
	.del {
		all: unset;
		position: absolute;
		top: 4px;
		right: 4px;
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.45);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.del :global(svg) {
		fill: currentColor;
		width: 18px;
		height: 18px;
	}
	.empty {
		grid-column: 1 / -1;
		text-align: center;
		color: #98a2b3;
		padding: 24px 0;
	}
	.name {
		word-break: break-all;
	}
	/* 영상 칸 — 표지 위 재생 표시(사진과 구분) */
	.vid {
		position: relative;
		display: inline-block;
		line-height: 0;
	}
	.vid-badge {
		position: absolute;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.55);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}
	.pending {
		margin-top: 4px;
		font-size: 11px;
		color: var(--theme-color-main, #1976d2);
		white-space: nowrap;
	}
</style>
