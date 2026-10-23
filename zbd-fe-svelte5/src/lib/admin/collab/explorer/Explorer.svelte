<script lang="ts">
	/**
	 * 웹하드(React explorer/[rootFolder]) — 칸 넘기기(macOS Finder 처럼) 목록 + 오른쪽 미리보기·정보.
	 * 오른쪽 클릭 메뉴: 새 파일·새 폴더·이름 변경·삭제·잘라내기/이동·편집. 파일·폴더를 끌어다 놓거나 「파일/폴더 올리기」 로 그 폴더에 올린다(폴더는 하위 폴더째).
	 * 특별한 파일(웹오피스 바로가기 등 — ext.ts)은 누르면 그 기능으로 열고, 폴더 메뉴에서 새로 만들 수 있다.
	 * 다른 곳에서 바꾸면 소켓 WEB_HDD → 목록·보던 미리보기를 다시 읽는다. 마지막에 고른 경로는 이 브라우저에 기억.
	 */
	import { tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Menu from '#lib/shared/mui/Menu.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { admin } from '../../profile.svelte.ts';
	import { onSocketDataUpdate } from '../../socket.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import { alert, confirm, prompt, toast } from '../../ui/modal.svelte.ts';
	import MarkdownView from '../ui/MarkdownView.svelte';
	import Spinner from '../ui/Spinner.svelte';
	import '../ui/collab.css';
	import { MENUITEM, SEC, TYPO } from '../ui/cls.ts';
	import { withLastViewMode } from './constants.ts';
	import { getFile, getFileTree, isImageFile, isTextFile, mkdir, mkfile, mv, rm, uploadFiles } from './files.ts';
	import type { FILE_NODE } from '#lib/shared/types/file.ts';
	import { SK } from '#lib/shared/storage-keys.ts';
	import { fileActions, fileExt, newItems } from './ext.ts';

	let { rootFolder }: { rootFolder: string } = $props();
	let root = $state<FILE_NODE | null>(null);
	let selectedPaths = $state<string[]>([]);
	let selectedFile = $state<FILE_NODE | null>(null);
	let preview = $state<any>('');
	let previewUrl = $state('');
	let loading = $state(false);
	let fromPaths = $state<string[]>([]);
	// 여러 개 고르기(⌘/Ctrl = 하나씩 더하기·빼기, Shift = 범위) — 한 칸(같은 폴더) 안에서만
	let multi = $state<string[]>([]);
	let anchor = '';
	let dragTarget = $state<string | null>(null);
	let ctx = $state<{ x: number; y: number; type: 'directory' | 'file' | 'blank'; node?: FILE_NODE; parentPath?: string } | null>(null);
	let ctxAnchor = $state<HTMLElement | null>(null);

	const formatBytes = (bytes: number, decimals = 2) => {
		if (!+bytes) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${parseFloat((bytes / Math.pow(k, i)).toFixed(decimals))} ${sizes[i]}`;
	};
	const formatDate = (date: any) => {
		if (!date) return '-';
		const d = new Date(date);
		if (isNaN(d.getTime())) return '-';
		return new Intl.DateTimeFormat('ko-KR', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long', hour: 'numeric', minute: 'numeric', hour12: false }).format(d);
	};
	const typeKo = (name: string) => {
		const fx = fileExt(name);
		if (fx) return fx.label;
		const ext = (name.match(/\.([a-z0-9]+)$/i)?.[1] || '').toUpperCase();
		if (/\.(jpeg|jpg|png|gif|webp|svg|bmp)$/i.test(name)) return `${ext} 이미지`;
		if (/\.(md|txt)$/i.test(name)) return '텍스트 문서';
		if (/\.(json|js|ts|jsx|tsx|html|css)$/i.test(name)) return '소스 코드';
		return ext ? `${ext} 파일` : '알 수 없는 파일';
	};
	const fileIcon = (name: string) => {
		const fx = fileExt(name);
		if (fx) return fx.icon;
		const ext = (name.match(/\.([a-z0-9]+)$/i)?.[1] || '').toLowerCase();
		if (['jpeg', 'jpg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return 'image';
		if (['md', 'txt'].includes(ext)) return 'description';
		if (['js', 'ts', 'jsx', 'tsx'].includes(ext)) return 'code';
		return (
			(
				{
					json: 'data_object',
					html: 'html',
					css: 'css',
					csv: 'table_chart',
					xls: 'table_chart',
					xlsx: 'table_chart',
					pdf: 'picture_as_pdf',
					zip: 'folder_zip',
					tar: 'folder_zip',
					gz: 'folder_zip'
				} as Record<string, string>
			)[ext] || 'insert_drive_file'
		);
	};

	async function loadTree() {
		root = await getFileTree(rootFolder);
	}
	const columns = $derived.by(() => {
		const cols: FILE_NODE[][] = [];
		if (!root?.items) return cols;
		cols.push(root.items);
		let cur = root.items;
		for (const p of selectedPaths) {
			const n = cur.find((x) => x.path === p);
			if (n && n.type === 'directory' && n.items) {
				cols.push(n.items);
				cur = n.items;
			} else break;
		}
		return cols;
	});
	async function showPreview(node: FILE_NODE | null) {
		selectedFile = node;
		preview = '';
		if (!node || !(isImageFile(node.name) || isTextFile(node.name))) return;
		loading = true;
		try {
			preview = await getFile(rootFolder, node.path);
		} catch {
			preview = '파일을 불러오는 중 오류가 발생했습니다.';
		} finally {
			loading = false;
		}
	}
	$effect(() => {
		if (preview instanceof Blob) {
			const u = URL.createObjectURL(preview);
			previewUrl = u;
			return () => URL.revokeObjectURL(u);
		}
		previewUrl = typeof preview == 'string' ? preview : '';
	});
	function clickItem(node: FILE_NODE, col: number, e?: MouseEvent, items: FILE_NODE[] = []) {
		const sameCol = (q: string) => items.some((x) => x.path === q);
		if (e && (e.metaKey || e.ctrlKey)) {
			const base = multi.filter(sameCol);
			const cur = selectedPaths[col];
			if (!base.length && cur && cur !== node.path && sameCol(cur)) base.push(cur);
			multi = base.includes(node.path) ? base.filter((q) => q !== node.path) : [...base, node.path];
			anchor = node.path;
			return;
		}
		if (e?.shiftKey) {
			const from = anchor && sameCol(anchor) ? anchor : selectedPaths[col] && sameCol(selectedPaths[col]) ? selectedPaths[col] : node.path;
			const a = items.findIndex((x) => x.path === from);
			const b = items.findIndex((x) => x.path === node.path);
			multi = items.slice(Math.min(a, b), Math.max(a, b) + 1).map((x) => x.path);
			return;
		}
		multi = [];
		anchor = node.path;
		selectedPaths = [...selectedPaths.slice(0, col), node.path];
		showPreview(node.type === 'file' ? node : null);
		if (e && node.type === 'file') openSpecial(node);
	}
	// 특별한 파일(ext.ts) — 그 기능으로 연다
	function openSpecial(node: FILE_NODE) {
		const fx = fileExt(node.name);
		if (fx) fx.open(rootFolder, node);
	}
	async function onNewSpecial(create: (root: string, folder: string) => Promise<string | null>) {
		const c = ctx;
		close();
		const folder = c?.type === 'directory' ? c.node?.path || '' : c?.type === 'blank' ? c.parentPath || '' : '';
		const path = await create(rootFolder, folder);
		if (!path) return;
		await loadTree();
		selectByPath('/' + path);
	}
	const storageKey = $derived(SK.explorer.lastPath(rootFolder));
	$effect(() => {
		if (selectedPaths.length) {
			try {
				localStorage.setItem(storageKey, selectedPaths[selectedPaths.length - 1]);
			} catch {}
		}
	});
	async function selectByPath(target: string) {
		let p = target;
		try {
			p = decodeURIComponent(target);
		} catch {}
		if (p.startsWith('/')) p = p.slice(1);
		else if (selectedFile) {
			const parts = selectedFile.path.split('/').slice(0, -1);
			let rel = false;
			for (const x of p.split('/')) {
				if (x === '.') rel = true;
				else if (x === '..') ((rel = true), parts.pop());
				else parts.push(x);
			}
			if (rel) p = parts.join('/');
		}
		const sel: string[] = [];
		let items = root?.items || [];
		let node: FILE_NODE | null = null;
		let q = '';
		for (const part of p.split('/')) {
			q = q ? `${q}/${part}` : part;
			const nq = q.normalize('NFC');
			let n = items.find((x) => x.path.normalize('NFC') === nq) || items.find((x) => x.path.normalize('NFC') === `${nq}.md`);
			if (!n) {
				node = null;
				break;
			}
			node = n;
			sel.push(n.path);
			items = n.items || [];
		}
		if (sel.length) selectedPaths = sel;
		if (node) showPreview(node.type === 'file' ? node : null);
	}
	loadTree().then(() => {
		try {
			const saved = localStorage.getItem(storageKey);
			if (saved) selectByPath('/' + saved);
		} catch {}
	});
	$effect(() => {
		let timer: any = null;
		const off = onSocketDataUpdate(
			'WEB_HDD',
			(data) => {
				if (data.key !== rootFolder) return;
				clearTimeout(timer);
				timer = setTimeout(async () => {
					await loadTree();
					const sel = selectedFile;
					if (sel && data.action === 'update' && (data.paths || []).includes(sel.path) && (isImageFile(sel.name) || isTextFile(sel.name))) {
						preview = await getFile(rootFolder, sel.path);
						toast({ msg: `「${sel.name}」 수정된 내용을 불러왔습니다.`, type: 'info' });
					} else {
						const label = ({ create: '새로 만들어졌습니다', upload: '올라왔습니다', delete: '삭제되었습니다', move: '옮겨졌습니다' } as Record<string, string>)[data.action || ''];
						const name = (data.paths?.[data.paths.length - 1] || '').split('/').pop();
						if (label && name && Number(data.by) !== Number(admin.profile?.adminId || 0)) toast({ msg: `「${name}」 ${label}.`, type: 'info' });
					}
				}, 300);
			},
			{ self: true }
		);
		return () => {
			clearTimeout(timer);
			off();
		};
	});
	// 해시(#제목) 로 자동 스크롤
	$effect(() => {
		const h = page.url.hash;
		if (!h || loading || !preview) return;
		setTimeout(() => document.getElementById(decodeURIComponent(h.slice(1)))?.scrollIntoView({ behavior: 'smooth' }), 100);
	});

	// 오른쪽 클릭 메뉴 — 누른 자리에 보이지 않는 기준점을 두고 Menu 를 띄운다
	let pin = $state<HTMLElement>();
	async function openCtx(e: MouseEvent, type: 'directory' | 'file' | 'blank', node?: FILE_NODE, parentPath?: string) {
		e.preventDefault();
		if (ctx) return close();
		ctx = { x: e.clientX + 2, y: e.clientY - 6, type, node, parentPath };
		await tick();
		ctxAnchor = pin || null;
	}
	const close = () => ((ctx = null), (ctxAnchor = null));
	const badName = (n: string) => /[/\\?%*:|"><]/.test(n);
	// 오른쪽 클릭한 것이 여러 개 고른 것 안이면 그 묶음 전체, 아니면 그 하나
	const targetsOf = (n?: FILE_NODE) => (n && multi.length > 1 && multi.includes(n.path) ? [...multi] : n ? [n.path] : []);
	async function onDelete() {
		const n = ctx?.node;
		close();
		if (!n) return;
		const ts = targetsOf(n);
		if (ts.length > 1) {
			if (!(await confirm(`${ts.length}개 항목을 삭제하시겠습니까?`))) return;
			let fail = 0;
			for (const t of ts) if (!(await rm(rootFolder, t)).ok) fail++;
			multi = [];
			await loadTree();
			if (fail) alert(`${fail}개는 삭제하지 못했습니다(내용이 있는 폴더 등).`);
			return;
		}
		if (await confirm(`'${n.name}'을(를) 삭제하시겠습니까?`)) {
			const o = await rm(rootFolder, n.path);
			if (!o.ok) return alert('삭제 실패');
			loadTree();
		}
	}
	async function onRename() {
		const n = ctx?.node;
		close();
		if (!n) return;
		const name = await prompt({ body: '새 이름을 입력하세요.', promptDefaultValue: n.name });
		if (!name || typeof name !== 'string' || name === n.name) return;
		if (badName(name)) return alert('이름에 사용할 수 없는 문자(/ \\ ? % * : | " < >)가 포함되어 있습니다.');
		const parts = n.path.split('/');
		parts.pop();
		parts.push(name);
		const o = await mv(rootFolder, n.path, parts.join('/'));
		if (!o.ok) return alert('이름 변경에 실패했습니다.');
		loadTree();
	}
	function targetOf(c: typeof ctx, fn: string) {
		if (c?.type === 'blank' && c.parentPath !== undefined) return c.parentPath ? `${c.parentPath}/${fn}` : fn;
		if (c?.node) return `${c.node.path}/${fn}`;
		return fn;
	}
	async function onMk(folder: boolean) {
		const c = ctx;
		close();
		const fn = await prompt(folder ? '새 폴더명을 입력하세요.' : '생성할 파일명을 입력하세요. (단, 폴더명은 .md를 제외하고 입력 가능하며 컨벤션을 따름)');
		if (!fn || typeof fn !== 'string') return;
		if (badName(fn)) return alert(`${folder ? '폴더명' : '파일명'}에 사용할 수 없는 문자(/ \\ ? % * : | " < >)가 포함되어 있습니다.`);
		const o = folder ? await mkdir(rootFolder, targetOf(c, fn)) : await mkfile(rootFolder, targetOf(c, fn));
		if (!o.ok) return alert(folder ? '폴더 생성에 실패했습니다.' : '파일 생성에 실패했습니다.');
		loadTree();
	}
	async function onMove() {
		const c = ctx;
		close();
		const to = c?.type === 'directory' ? c.node?.path || '' : c?.type === 'blank' ? c.parentPath || '' : '';
		let fail = 0;
		for (const p of fromPaths) {
			// 자기 자신·자기 아래로는 옮기지 않는다
			if (to === p || to.startsWith(p + '/')) continue;
			if (!(await mv(rootFolder, p, to)).ok) fail++;
		}
		fromPaths = [];
		multi = [];
		await loadTree();
		if (fail) alert(`${fail}개는 옮기지 못했습니다.`);
	}
	const openEdit = (path: string) => goto(withLastViewMode(`/admin/explorer/${rootFolder}/${path}`));
	// 올릴 것 = {rel: 놓은 곳 기준 상대 경로, file} — 폴더를 끌어 놓거나 「폴더 올리기」 로 고르면 하위 폴더째(빈 폴더는 dirs)
	type Up = { rel: string; file: File };
	async function readEntry(en: any, base: string, out: Up[], dirs: string[]) {
		if (en.isFile) {
			const file: File = await new Promise((ok, no) => en.file(ok, no));
			if (file.name !== '.DS_Store') out.push({ rel: base + file.name, file });
		} else if (en.isDirectory) {
			const dir = base + en.name;
			dirs.push(dir);
			const reader = en.createReader();
			// readEntries 는 한 번에 100개까지 — 빈 배열이 올 때까지 읽는다
			for (;;) {
				const list: any[] = await new Promise((ok, no) => reader.readEntries(ok, no));
				if (!list.length) break;
				for (const c of list) await readEntry(c, dir + '/', out, dirs);
			}
		}
	}
	async function uploadTree(target: string, list: Up[], dirs: string[] = []) {
		if (!list.length && !dirs.length) return;
		const join = (d: string) => [target, d].filter(Boolean).join('/');
		const groups = new Map<string, File[]>();
		for (const u of list) {
			const i = u.rel.lastIndexOf('/');
			const d = i < 0 ? '' : u.rel.slice(0, i);
			groups.set(d, [...(groups.get(d) || []), u.file]);
		}
		let fail = 0;
		for (const [d, files] of groups) {
			// 한 번에 너무 많이 보내지 않게 50개씩
			for (let i = 0; i < files.length; i += 50) {
				const o = await uploadFiles(rootFolder, join(d), files.slice(i, i + 50));
				if (!o.ok) fail++;
			}
		}
		for (const d of dirs) if (![...groups.keys()].some((g) => g === d || g.startsWith(d + '/'))) await mkdir(rootFolder, join(d));
		await loadTree();
		if (fail) return void window.alert('업로드 중 오류가 발생했습니다.');
		toast({ msg: `${list.length}개 파일이 업로드 되었습니다.`, type: 'success' });
	}
	async function drop(e: DragEvent, target: string) {
		e.preventDefault();
		e.stopPropagation();
		dragTarget = null;
		const dt = e.dataTransfer;
		if (!dt) return;
		// getAsEntry 는 이벤트가 끝나면 못 읽는다 — 먼저 모두 꺼낸다
		const entries = Array.from(dt.items || [])
			.filter((it) => it.kind === 'file')
			.map((it) => (it as any).webkitGetAsEntry?.())
			.filter(Boolean);
		const list: Up[] = [];
		const dirs: string[] = [];
		if (entries.length) for (const en of entries) await readEntry(en, '', list, dirs);
		else Array.from(dt.files || []).forEach((file) => list.push({ rel: file.name, file }));
		await uploadTree(target, list, dirs);
	}
	let fileInput = $state<HTMLInputElement>();
	let dirInput = $state<HTMLInputElement>();
	let pickTarget = '';
	function pick(folder: boolean) {
		const c = ctx;
		pickTarget = c?.type === 'directory' ? c.node?.path || '' : c?.type === 'blank' ? c.parentPath || '' : '';
		close();
		(folder ? dirInput : fileInput)?.click();
	}
	async function picked(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const list = Array.from(el.files || [])
			.filter((f) => f.name !== '.DS_Store')
			.map((file) => ({ rel: (file as any).webkitRelativePath || file.name, file }));
		el.value = '';
		await uploadTree(pickTarget, list);
	}
	const over = (e: DragEvent, id: string) => {
		e.preventDefault();
		e.stopPropagation();
		if (dragTarget !== id) dragTarget = id;
	};
	async function download() {
		if (!selectedFile) return;
		const c = await getFile(rootFolder, selectedFile.path);
		const url = URL.createObjectURL(c instanceof Blob ? c : new Blob([c]));
		const a = document.createElement('a');
		a.href = url;
		a.download = selectedFile.name;
		document.body.appendChild(a);
		a.click();
		a.remove();
		URL.revokeObjectURL(url);
	}
	// 미리보기 마크다운 — 제목에 id(목차 링크), 상대 링크는 탐색기 안에서 이동
	function mdLinks(node: HTMLElement) {
		const fix = () =>
			node.querySelectorAll('h1,h2,h3,h4,h5,h6').forEach((h) => {
				if (!h.id)
					h.id = (h.textContent || '')
						.toLowerCase()
						.replace(/\s+/g, '-')
						.replace(/[^\wㄱ-ㅎㅏ-ㅣ가-힣\-]+/g, '')
						.replace(/--+/g, '-')
						.replace(/^-+|-+$/g, '');
			});
		fix();
		const mo = new MutationObserver(fix);
		mo.observe(node, { childList: true, subtree: true });
		const click = (e: MouseEvent) => {
			const a = (e.target as HTMLElement).closest('a');
			const href = a?.getAttribute('href') || '';
			if (!a || !href) return;
			if (href.startsWith('#')) {
				e.preventDefault();
				e.stopImmediatePropagation();
				document.getElementById(decodeURIComponent(href.slice(1)))?.scrollIntoView({ behavior: 'smooth' });
			} else if (!href.startsWith('http') && !href.startsWith('mailto:')) {
				e.preventDefault();
				e.stopImmediatePropagation();
				selectByPath(href);
			}
		};
		node.addEventListener('click', click, true);
		return { destroy: () => (mo.disconnect(), node.removeEventListener('click', click, true)) };
	}
	const mdSx = {
		'& h1, & h2, & h3, & h4, & h5, & h6': { scrollMarginTop: '24px' },
		'& h1': { fontSize: '1.75rem' },
		'& h2': { fontSize: '1.5rem' },
		'& h1, & h2': { borderBottom: '1px solid var(--adm-divider)', paddingBottom: 8, marginBottom: 16, color: 'var(--theme-color-main)' },
		'& :not(pre) > code': { backgroundColor: 'var(--adm-hover)', paddingLeft: 4, paddingRight: 4, borderRadius: '4px' },
		'& table': { width: '100%', borderCollapse: 'collapse', marginBottom: 16, border: '1px solid var(--adm-divider)' },
		'& th, & td': { border: '1px solid var(--adm-divider)', padding: 8, textAlign: 'left' },
		'& th': { backgroundColor: 'var(--adm-hover)', fontWeight: 600 }
	};
	const body2 = 'MuiTypography-root MuiTypography-body2';
	type CtxItem = { icon: string; label: string; fn: () => void; disabled?: boolean };
	const extItems: CtxItem[] = newItems.map((k) => ({ icon: k.icon, label: k.label, fn: () => onNewSpecial(k.create) }));
	const ctxItems = $derived.by((): CtxItem[] => {
		const c = ctx;
		if (!c) return [];
		const n = targetsOf(c.node).length;
		const cut: CtxItem = { icon: 'content_cut', label: n > 1 ? `잘라내기 (${n}개)` : '잘라내기', fn: () => ((fromPaths = targetsOf(c.node)), close()) };
		const move: CtxItem[] = fromPaths.length ? [{ icon: 'content_paste', label: fromPaths.length > 1 ? `붙여넣기 (${fromPaths.length}개)` : '붙여넣기', fn: onMove }] : [];
		if (c.type === 'directory')
			return [
				{ icon: 'delete', label: n > 1 ? `삭제 (${n}개)` : '삭제', fn: onDelete, disabled: n < 2 && !!c.node?.items?.length },
				{ icon: 'note_add', label: '새 파일', fn: () => onMk(false) },
				{ icon: 'create_new_folder', label: '새 폴더', fn: () => onMk(true) },
				...extItems,
				{ icon: 'upload_file', label: '파일 올리기', fn: () => pick(false) },
				{ icon: 'drive_folder_upload', label: '폴더 올리기', fn: () => pick(true) },
				{ icon: 'drive_file_rename_outline', label: '이름 변경', fn: onRename },
				cut,
				...move
			];
		if (c.type === 'file')
			return [
				{ icon: 'delete', label: n > 1 ? `삭제 (${n}개)` : '삭제', fn: onDelete },
				{ icon: 'drive_file_rename_outline', label: '이름 변경', fn: onRename, disabled: n > 1 },
				...(fileExt(c.node?.name || '') ? [{ icon: 'open_in_new', label: fileExt(c.node?.name || '')!.openLabel, fn: () => (close(), openSpecial(c.node!)) }] : []),
				...fileActions(c.node?.name || '').map((a) => ({ icon: a.icon, label: a.label, fn: () => (close(), a.run(rootFolder, c.node!)) })),
				{ icon: 'edit', label: '편집', fn: () => (close(), openEdit(c.node!.path)), disabled: !isTextFile(c.node?.name || '') },
				cut
			];
		return [
			{ icon: 'note_add', label: '새 파일', fn: () => onMk(false) },
			{ icon: 'create_new_folder', label: '새 폴더', fn: () => onMk(true) },
			...extItems,
			{ icon: 'upload_file', label: '파일 올리기', fn: () => pick(false) },
			{ icon: 'drive_folder_upload', label: '폴더 올리기', fn: () => pick(true) },
			...move
		];
	});
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && ((multi = []), (fromPaths = []))} />

<PageContainer sx={{ paddingBottom: 0 }}>
	<input bind:this={fileInput} type="file" multiple hidden onchange={picked} />
	<input bind:this={dirInput} type="file" webkitdirectory multiple hidden onchange={picked} />
	<div
		class={cx(
			'MuiBox-root',
			toClass({
				display: 'flex',
				flexDirection: 'row',
				height: 'calc(100vh - 200px)',
				minHeight: 400,
				overflowX: 'auto',
				overflowY: 'hidden',
				border: '1px solid var(--adm-divider)',
				backgroundColor: 'var(--adm-paper)',
				borderRadius: 8
			})
		)}
		role="presentation"
		ondragleave={() => (dragTarget = null)}
	>
		{#each columns as items, ci (ci)}
			{@const parentPath = ci === 0 ? '' : selectedPaths[ci - 1]}
			{@const multiCol = multi.length > 0 && items.some((x) => multi.includes(x.path))}
			<div
				class={cx(
					'MuiBox-root',
					toClass({
						minWidth: 260,
						width: 260,
						borderRight: '1px solid var(--adm-divider)',
						overflowY: 'auto',
						display: 'flex',
						flexDirection: 'column',
						padding: '4px 0',
						backgroundColor: dragTarget === `col-${ci}` ? 'rgba(0, 0, 0, 0.04)' : 'transparent',
						transition: 'background-color 0.2s'
					})
				)}
				role="list"
				oncontextmenu={(e) => openCtx(e, 'blank', undefined, parentPath)}
				ondragover={(e) => over(e, `col-${ci}`)}
				ondrop={(e) => drop(e, parentPath)}
			>
				{#each items as node (node.path)}
					{@const sel = multiCol ? multi.includes(node.path) : selectedPaths.includes(node.path)}
					{@const cut = fromPaths.includes(node.path)}
					<div
						class={cx(
							'MuiBox-root',
							toClass({
								display: 'flex',
								alignItems: 'center',
								padding: '5px 11px',
								margin: '2px 8px',
								borderRadius: 6,
								cursor: 'pointer',
								userSelect: 'none',
								backgroundColor: dragTarget === node.path ? 'rgba(0, 0, 0, 0.08)' : sel ? 'var(--theme-color-main)' : cut ? 'var(--adm-selected)' : 'transparent',
								color: sel ? 'var(--theme-color-contrastText)' : 'var(--adm-text)',
								border: cut ? '1px dashed var(--adm-text2)' : '1px solid transparent',
								opacity: cut ? 0.6 : 1,
								'&:hover': { backgroundColor: sel ? 'var(--theme-color-main)' : 'var(--adm-hover)' }
							})
						)}
						role="listitem"
						onclick={(e) => clickItem(node, ci, e, items)}
						onmousedown={(e) => (e.shiftKey || e.metaKey || e.ctrlKey) && e.preventDefault()}
						onkeydown={() => {}}
						oncontextmenu={(e) => (e.stopPropagation(), openCtx(e, node.type, node))}
						ondragover={node.type === 'directory' ? (e) => over(e, node.path) : undefined}
						ondrop={node.type === 'directory' ? (e) => drop(e, node.path) : undefined}
					>
						<Icon
							name={node.type === 'directory' ? 'folder' : fileIcon(node.name)}
							sx={{ color: sel ? 'inherit' : node.type === 'directory' ? '#4fc3f7' : fileExt(node.name)?.color || SEC, marginRight: 8, fontSize: 20 }}
						/>
						<p class={cx(body2, toClass({ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }))}>{node.name}</p>
						{#if node.type === 'directory'}<Icon name="chevron_right" sx={{ fontSize: 20, opacity: 0.7 }} />{/if}
						{#if node.type === 'file' && isTextFile(node.path)}
							<Button
								size="small"
								variant="contained"
								color="inherit"
								sx={{
									minWidth: 0,
									padding: '2px 8px',
									fontSize: '0.7rem',
									color: 'var(--adm-text)',
									marginLeft: 8,
									boxShadow: 'none',
									'&:hover': { boxShadow: 'none', backgroundColor: 'var(--adm-hover)' }
								}}
								onclick={(e) => (e.stopPropagation(), openEdit(node.path))}>편집</Button
							>
						{/if}
					</div>
				{/each}
			</div>
		{/each}

		{#if selectedFile}
			<div
				class={cx('MuiBox-root', toClass({ display: 'flex', flex: 1, minWidth: 400, borderLeft: '1px solid var(--adm-divider)', backgroundColor: 'var(--zc-fafafa, #fafafa)', overflow: 'hidden' }))}
			>
				<div class={cx('MuiBox-root', toClass({ flex: 1, position: 'relative', overflowY: 'auto', display: 'flex', flexDirection: 'column' }))}>
					{#if loading}
						<div class={cx('MuiBox-root', toClass({ display: 'flex', justifyContent: 'center', paddingTop: 80, flex: 1 }))}><Spinner /></div>
					{:else}
						<div class={cx('MuiBox-root', toClass({ flex: 1, ...mdSx }))}>
							{#if isImageFile(selectedFile.name)}
								<div class={cx('MuiBox-root', toClass({ textAlign: 'center' }))}>
									<img src={previewUrl} alt={selectedFile.name} style="max-width:100%;max-height:calc(100vh - 350px);object-fit:contain;display:inline" />
									{#each fileActions(selectedFile.name) as a}
										{@const f = selectedFile}
										<div class={cx('MuiBox-root', toClass({ marginTop: 12 }))}>
											<Button variant="outlined" onclick={() => a.run(rootFolder, f)}><Icon name={a.icon} sx={{ fontSize: 18, marginRight: 6 }} />{a.label}</Button>
										</div>
									{/each}
								</div>
							{:else if isTextFile(selectedFile.name)}
								{#if selectedFile.name.endsWith('.md')}
									<div class={cx('MuiBox-root', toClass({ paddingLeft: 24, paddingRight: 24 }))} use:mdLinks><MarkdownView text={typeof preview === 'string' ? preview : ''} /></div>
								{:else}
									{#if fileActions(selectedFile.name).length}
										{@const f = selectedFile}
										<div class={cx('MuiBox-root', toClass({ display: 'flex', gap: 8, padding: '12px 16px 0' }))}>
											{#each fileActions(f.name) as a}
												<Button variant="outlined" size="small" onclick={() => a.run(rootFolder, f)}><Icon name={a.icon} sx={{ fontSize: 18, marginRight: 6 }} />{a.label}</Button>
											{/each}
										</div>
									{/if}
									<pre
										class={toClass({
											margin: 0,
											padding: '16px 16px 4px',
											fontFamily: 'Menlo, Monaco, monospace',
											fontSize: 14,
											lineHeight: '19px',
											whiteSpace: 'pre-wrap',
											wordBreak: 'break-word',
											backgroundColor: '#fafafa'
										})}>{typeof preview === 'string' ? preview : ''}</pre>
								{/if}
							{:else if fileExt(selectedFile.name)}
								{@const of = selectedFile}
								{@const fx = fileExt(of.name)!}
								<div class={cx('MuiBox-root', toClass({ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 24 }))}>
									<Icon name={fx.icon} sx={{ fontSize: 64, color: fx.color, marginBottom: 16 }} />
									<p class={cx(TYPO, toClass({ marginBottom: 24 }))}>{fx.hint}</p>
									<Button variant="contained" onclick={() => openSpecial(of)}>{fx.openLabel}</Button>
								</div>
							{:else}
								<div class={cx('MuiBox-root', toClass({ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 24 }))}>
									<Icon name={fileIcon(selectedFile.name)} sx={{ fontSize: 64, color: SEC, marginBottom: 16 }} />
									<p class={cx(TYPO, toClass({ marginBottom: 24 }))}>미리보기를 지원하지 않는 파일 형식입니다.</p>
									<div class={cx('MuiBox-root', toClass({ display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center' }))}>
										<Button variant="contained" onclick={download}>다운로드</Button>
										{#each fileActions(selectedFile.name) as a}
											{@const f = selectedFile}
											<Button variant="outlined" onclick={() => a.run(rootFolder, f)}><Icon name={a.icon} sx={{ fontSize: 18, marginRight: 6 }} />{a.label}</Button>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/if}
					<div class={cx('MuiStack-root', toClass({ alignItems: 'center', background: 'var(--zc-fff, #fff)', padding: 16 }))}>
						<div class={cx('MuiBox-root', toClass({ minWidth: 500 }))}>
							<h6 class={cx('MuiTypography-root MuiTypography-h6', toClass({ fontWeight: 600, marginBottom: 4, wordBreak: 'break-all', fontSize: '1.1rem' }))}>{selectedFile.name}</h6>
							<p class={cx(body2, toClass({ color: SEC, marginBottom: 32 }))}>{typeKo(selectedFile.name)} - {formatBytes(selectedFile.size)}</p>
							<div
								class={cx(
									'MuiBox-root',
									toClass({ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #eee', paddingBottom: 4, marginBottom: 8, maxWidth: 500 })
								)}
							>
								<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ fontWeight: 600 }))}>정보</h6>
							</div>
							{#each [['생성일', 12], ['수정일', 12], ['최근 사용일', 32]] as [label, mb] (label)}
								<div class={cx('MuiBox-root', toClass({ display: 'flex', justifyContent: 'space-between', marginBottom: mb as number, maxWidth: 500 }))}>
									<p class={cx(body2, toClass({ color: SEC }))}>{label}</p>
									<p class={body2}>{formatDate(selectedFile.updatedAt)}</p>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
	{#if ctx}<div bind:this={pin} style="position:fixed;left:{ctx.x}px;top:{ctx.y}px;width:0;height:0"></div>{/if}
	<Menu anchor={ctxAnchor} open={!!ctxAnchor} onclose={close}>
		{#each ctxItems as it (it.label)}
			<li class={cx(MENUITEM, it.disabled && 'Mui-disabled')} role="menuitem" tabindex="0" onclick={() => !it.disabled && it.fn()} onkeydown={() => {}}>
				<div class={cx('MuiListItemIcon-root', toClass({ minWidth: 36 }))}><Icon name={it.icon} size="Small" /></div>
				<div class="MuiListItemText-root"><span class={TYPO}>{it.label}</span></div>
			</li>
		{/each}
	</Menu>
</PageContainer>
