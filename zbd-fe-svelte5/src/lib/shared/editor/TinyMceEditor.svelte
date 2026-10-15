<script lang="ts" module>
	declare global {
		interface Window {
			tinymce?: any;
		}
	}
	/** TinyMCE 6.4 — 이 앱의 `/assets/tinymce` 에서 한 번만 싣는다(예전엔 oye 쪽 NCP CDN 에서 받았다) */
	const BASE = '/assets/tinymce';
	let loading: Promise<any> | null = null;
	export function loadTinymce(): Promise<any> {
		if (window.tinymce) return Promise.resolve(window.tinymce);
		loading ||= new Promise((res, rej) => {
			const s = document.createElement('script');
			s.id = 'tinymce-bundle';
			s.src = `${BASE}/tinymce.min.js`;
			s.onload = () => res(window.tinymce);
			s.onerror = rej;
			document.body.appendChild(s);
		});
		return loading;
	}
</script>

<script lang="ts">
	/**
	 * 글 편집기 — 예전 components/common/TinyMceEditor.tsx (관리자·사용자 공용).
	 * 도구 막대의 「이미지 첨부」(onImageUI) · 「링크 첨부」(OG 카드) 단추는 우리 것.
	 * 바깥에서 커서 자리에 넣기: bind:this={ed} → ed.insert(html)
	 */
	import { onDestroy } from 'svelte';
	import { toast } from '../ui/dialog.svelte.ts';
	import LinkAttachDialog from './LinkAttachDialog.svelte';
	import { uploadFiles } from '../image/files';
	import { HOSTING_ID, imagePrefix } from '../config';

	let {
		body = '',
		height = 600,
		onchange,
		onImageUI = null,
		readonly = false,
		isMobile = false,
		statusbar = true,
		customToolbar = '',
		imageId = '',
		userId = null,
		hostingId = HOSTING_ID,
		placeholder = ''
	}: {
		body?: string;
		height?: number;
		onchange: (html: string) => void;
		onImageUI?: (() => void) | null;
		readonly?: boolean;
		isMobile?: boolean;
		statusbar?: boolean;
		customToolbar?: string;
		/** 링크 카드 그림을 올릴 폴더(예: u_{userId}_post_{postId}) */
		imageId?: string;
		userId?: number | null;
		/** 붙여 넣은 그림을 올릴 호스팅 폴더(기본 ZBD) */
		hostingId?: string;
		placeholder?: string;
	} = $props();

	let host = $state<HTMLTextAreaElement>();
	let editor: any = null;
	let linkOpen = $state(false);
	let timer: any;

	export function insert(html: string) {
		if (!editor) return;
		editor.insertContent(html);
		editor.focus();
		onchange(editor.getContent());
	}

	// 높이를 먼저 읽는다 — 편집기 준비 전(editor 없음)엔 height 를 안 읽어 이 effect 가 높이 변화에 다시 돌지 않았다(창 크기 따라 바꾸는 화면이 처음 높이에 고정, 2026-10-04)
	$effect(() => {
		const h = height;
		if (editor) editor.getContainer().style.height = h + 'px';
	});

	$effect(() => {
		if (!host) return;
		let dead = false;
		loadTinymce().then((tinymce) => {
			if (dead || !host) return;
			// 플러그인 이름은 껍데기만 — 단추는 아래 setup 에서 편집기마다 만든다(화면마다 다른 동작)
			for (const n of ['ImagePickerPlugin', 'linkAttachPlugin']) if (!tinymce.PluginManager.get(n)) tinymce.PluginManager.add(n, () => ({ name: n }));
			let plugins = [
				'linkAttachPlugin',
				'advlist',
				'autolink',
				'lists',
				'image',
				'charmap',
				'preview',
				'anchor',
				'searchreplace',
				'visualblocks',
				'code',
				'codesample',
				'fullscreen',
				'insertdatetime',
				'media',
				'table',
				'wordcount',
				'emoticons',
				'pagebreak',
				'visualchars'
			];
			if (onImageUI) plugins = ['ImagePickerPlugin'].concat(plugins);
			let toolbar = `imagePickerButton linkAttachButton | undo redo | styles styleselect | fontselect fontsize | bold italic forecolor  | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent |  code  codesample | link emoticons  fullscreen pagebreak visualblocks visualchars  table`;
			if (isMobile) toolbar = `imagePickerButton linkAttachButton | undo redo |  bold italic forecolor emoticons `;
			if (customToolbar) toolbar = customToolbar;

			tinymce.init({
				target: host,
				base_url: BASE,
				suffix: '.min',
				license_key: 'gpl',
				menubar: true,
				statusbar,
				language: 'ko_KR',
				placeholder,
				height,
				width: '100%',
				resize: true,
				relative_urls: false,
				plugins,
				paste_block_drop: true,
				paste_data_images: false, // 그림은 paste 처리기에서 올려 주소로 넣는다
				valid_elements: '*[*]',
				extended_valid_elements: 'a[*],div[*],span[*],img[*],p[*],h1[*],h2[*],h3[*],h4[*],h5[*],h6[*]',
				toolbar,
				content_css: [
					'https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable.min.css',
					'https://fonts.googleapis.com/css2?family=Noto+Color+Emoji&display=swap'
				],
				content_style: `body { font-family:Pretendard Variable,"Noto Color Emoji",sans-serif; font-size:16px }
            .mce-content-body img {max-width: 100%;}
            .mce-content-body:not([dir=rtl])[data-mce-placeholder]:not(.mce-visualblocks)::before { white-space: pre-line; }`,
				codesample_languages: [
					{ text: 'Bash', value: 'bash' },
					{ text: 'Nginx', value: 'nginx' },
					{ text: 'HTML/XML', value: 'html' },
					{ text: 'JavaScript', value: 'javascript' },
					{ text: 'CSS', value: 'css' },
					{ text: 'PHP', value: 'php' },
					{ text: 'Python', value: 'python' },
					{ text: 'Java', value: 'java' }
				],
				setup(ed: any) {
					// 도구 막대 단추 — 편집기마다 등록(플러그인 이름 대신 setup 에서, 화면마다 다른 동작을 갖게)
					// 이미지 단추는 고르기 창을 넘겨준 화면에만(React 와 같게 — 없으면 단추가 눌러도 아무 일 없이 보였다)
					if (onImageUI) ed.ui.registry.addButton('imagePickerButton', { icon: 'image', tooltip: '이미지 첨부', onAction: () => onImageUI?.() });
					ed.ui.registry.addButton('linkAttachButton', { icon: 'link', tooltip: '링크 첨부', onAction: () => (linkOpen = true) });
					ed.on('init', () => {
						editor = ed;
						ed.setContent(body || '');
						if (readonly) ed.mode.set('readonly');
					});
					ed.on('input change undo redo SetContent', () => {
						clearTimeout(timer);
						timer = setTimeout(() => editor && onchange(editor.getContent()), 500);
					});
					ed.on('drop', () => toast('이미지를 첨부하시려면 본문에 붙여넣기 하거나, 이미지 첨부 아이콘을 클릭하세요'));
					// 붙여 넣은 그림은 바로 올리고 주소로 넣는다(React 와 같게) — 안 하면 data:/blob: 그림이 들어가 메일·저장 때 빈 그림이 된다
					ed.on('paste', async (e: ClipboardEvent) => {
						const items = e.clipboardData?.items;
						if (!items) return;
						const files: File[] = [];
						for (let i = 0; i < items.length; i++) {
							const it = items[i];
							if (it.kind === 'file' && it.type?.startsWith('image/')) {
								const f = it.getAsFile();
								if (f) files.push(f);
							}
						}
						if (!files.length) return;
						e.preventDefault();
						e.stopPropagation();
						if (!imageId) return toast('이미지 업로드 경로(imageId)가 없어 첨부할 수 없습니다.');
						toast(files.length > 1 ? `이미지 ${files.length}장 업로드 중...` : '이미지 업로드 중...');
						const ts = Date.now().toString(36);
						const up = await uploadFiles(
							files.map((f, i) => ({ blob: f, name: `paste_${ts}${i ? '_' + i : ''}.${((f.type || 'image/png').split('/')[1] || 'png').toLowerCase()}` })),
							imageId,
							{ userId, hostingId }
						);
						if (!up?.length) return toast('이미지 업로드에 실패하였습니다.');
						const html = up.map((u) => `<img src="${imagePrefix(hostingId)}/${imageId}/${encodeURIComponent(u.name)}" />`).join('');
						if (editor) {
							editor.insertContent(html);
							editor.focus();
							onchange(editor.getContent());
						}
						toast('이미지를 업로드했습니다.');
					});
				}
			});
		});
		return () => {
			dead = true;
		};
	});

	onDestroy(() => {
		clearTimeout(timer);
		editor?.remove();
		editor = null;
	});
</script>

<div class="MuiStack-root zbd-editor">
	<textarea bind:this={host} style="visibility:hidden;height:0;padding:0;border:0;margin:0"></textarea>
</div>

<LinkAttachDialog
	open={linkOpen}
	{imageId}
	{userId}
	{isMobile}
	onclose={() => (linkOpen = false)}
	onattach={(html) => {
		linkOpen = false;
		if (editor && html) {
			editor.insertContent(html);
			editor.focus();
		}
	}}
/>
