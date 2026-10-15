<script lang="ts">
	/**
	 * 웹하드 문서 편집(React explorer/components/MarkdownFile) — `/admin/explorer/{루트}/{경로…}`.
	 * 보기 모드(md): 없음 = 좌우 · 3 = 위아래 · 2 = 에디터만 · 1 = 미리보기만(?viewMode=, 마지막 모드는 이 브라우저에 기억).
	 * 실시간: 다른 곳에서 저장하면 소켓 WEB_HDD 로 다시 읽는다 — 고치던 게 없으면 바꾸고, 있으면 줄 단위 3-way 병합(같은 곳이면 내 것을 두고 경고).
	 * 함께 보기: 15초마다 presence → 상단 칩 「보는 중 / 저장 안 한 편집」. 저장은 baseHash 와 함께 — 그사이 바뀌었으면 합쳐 다시 저장.
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { media } from '../../media.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { onSocketDataUpdate, socketState } from '../../socket.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import Tooltip from '../../ui/Tooltip.svelte';
	import Chip from '../ui/Chip.svelte';
	import '../ui/collab.css';
	import { saveLastViewMode } from './constants.ts';
	import { getTextFile, updateFile } from './files.ts';
	import MdEditor from './MdEditor.svelte';
	import { merge3, textHash } from './merge3.ts';

	let { rootFolder, targetFile }: { rootFolder: string; targetFile: string } = $props();
	function langOf(f: string) {
		const ext = f.split('.').pop()?.toLowerCase() || '';
		if (ext == 'md') return 'markdown';
		if (['js', 'jsx'].includes(ext)) return 'javascript';
		if (['ts', 'tsx'].includes(ext)) return 'typescript';
		if (['yml', 'yaml'].includes(ext)) return 'yaml';
		if (['json', 'html', 'css', 'xml'].includes(ext)) return ext;
		return 'plaintext';
	}
	const language = $derived(langOf(targetFile));
	const canToggle = $derived(language === 'markdown');
	const vm = $derived(page.url.searchParams.get('viewMode'));
	const mode = $derived(!canToggle ? 'split' : vm === '1' ? 'view' : vm === '2' ? 'edit' : vm === '3' ? 'splitv' : 'split');
	$effect(() => {
		if (canToggle) saveLastViewMode(vm && /^[123]$/.test(vm) ? vm : '');
	});
	function changeMode(m: string) {
		if (m === mode) return;
		const u = new URL(location.href);
		if (m === 'view') u.searchParams.set('viewMode', '1');
		else if (m === 'edit') u.searchParams.set('viewMode', '2');
		else if (m === 'splitv') u.searchParams.set('viewMode', '3');
		else u.searchParams.delete('viewMode');
		goto(u.pathname + u.search, { replace: true, reset: false });
	}

	let text = $state<string | null>(null);
	let textOrig = $state<string | null>(null);
	$effect(() => {
		const f = targetFile;
		text = null;
		textOrig = null;
		getTextFile(rootFolder, f).then(({ text: t }) => {
			if (f !== targetFile) return;
			text = t;
			textOrig = t;
		});
	});

	// 바뀐 미리보기 문단 칠하기
	let previewEl = $state<HTMLElement>();
	let highlight: Set<string> | null = null;
	const blocks = () => Array.from(previewEl?.querySelectorAll('.md-preview-body > *') || []) as HTMLElement[];
	const keyOf = (el: HTMLElement) => el.tagName + ':' + (el.textContent || '').trim();
	const markBeforeChange = () => (highlight = new Set(blocks().map(keyOf)));
	$effect(() => {
		void text;
		const prev = highlight;
		if (!prev) return;
		highlight = null;
		setTimeout(() => {
			const changed = blocks().filter((el) => !prev.has(keyOf(el)));
			changed.forEach((el) => el.classList.add('md-changed'));
			if (changed[0]) changed[0].scrollIntoView({ block: 'center', behavior: 'smooth' });
			setTimeout(() => changed.forEach((el) => el.classList.remove('md-changed')), 3500);
		}, 50);
	});
	function applyRemote(fresh: string, who: string) {
		if (fresh === textOrig) return;
		if (text === textOrig) {
			markBeforeChange();
			text = fresh;
			textOrig = fresh;
			return toast({ msg: `${who} 수정한 내용을 불러왔습니다.`, type: 'info' });
		}
		const m = merge3(textOrig || '', text || '', fresh);
		if (m.conflict) {
			textOrig = fresh;
			return toast({ msg: `${who} 같은 곳을 고쳤습니다. 내 편집은 그대로 두었습니다 — 저장할 때 확인합니다.`, type: 'warning' });
		}
		markBeforeChange();
		text = m.text;
		textOrig = fresh;
		toast({ msg: `${who} 수정한 내용을 내 편집과 합쳤습니다.`, type: 'info' });
	}

	// 함께 보기(presence)
	let peers = $state<Record<number, { name: string; editing: boolean; at: number }>>({});
	const dirty = $derived(text !== null && text !== textOrig);
	const myId = $derived(Number(admin.profile?.adminId || 0));
	let lastSent = 0;
	function sendPresence(on = true) {
		if (!rootFolder || !targetFile || language !== 'markdown') return;
		lastSent = Date.now();
		api.post(`/service/admin/file-explorer/${rootFolder}/presence`, { filePath: targetFile, editing: on && dirty, on });
	}
	$effect(() => {
		if (!socketState.connected) return;
		void targetFile;
		sendPresence(true);
		const iv = setInterval(() => sendPresence(true), 15000);
		const sweep = setInterval(() => {
			const n = Date.now();
			const next = Object.fromEntries(Object.entries(peers).filter(([, v]) => n - v.at < 40000));
			if (Object.keys(next).length !== Object.keys(peers).length) peers = next;
		}, 5000);
		return () => {
			clearInterval(iv);
			clearInterval(sweep);
			sendPresence(false);
			peers = {};
		};
	});
	$effect(() => {
		void dirty;
		if (socketState.connected) sendPresence(true);
	});
	$effect(() => {
		let timer: any = null;
		const off = onSocketDataUpdate(
			'WEB_HDD',
			(data) => {
				if (data.key !== rootFolder || !(data.paths || []).includes(targetFile)) return;
				if (data.action === 'presence') {
					const id = Number(data.by) || 0;
					if (!id || id === myId) return;
					if (data.on === false) {
						const { [id]: _, ...rest } = peers;
						peers = rest;
						return;
					}
					if (!peers[id] && Date.now() - lastSent > 3000) setTimeout(() => sendPresence(true), 300);
					peers = { ...peers, [id]: { name: data.byName || `관리자 ${id}`, editing: !!data.editing, at: Date.now() } };
					return;
				}
				if (!['update', 'create', 'upload'].includes(data.action)) return;
				const who = Number(data.by) === myId ? '다른 창에서' : data.byName ? `${data.byName}님이` : '다른 곳에서';
				clearTimeout(timer);
				timer = setTimeout(async () => {
					const { text: fresh } = await getTextFile(rootFolder, targetFile);
					if (typeof fresh === 'string') applyRemote(fresh, who);
				}, 300);
			},
			{ self: true }
		);
		return () => {
			clearTimeout(timer);
			off();
		};
	});

	async function doSave(txt?: string, force = false): Promise<void> {
		const body = txt ?? text;
		if (!body) return;
		const base = textOrig || '';
		const o = await updateFile(rootFolder, targetFile, body, force ? undefined : textHash(base));
		if (o.conflict && typeof o.data === 'string') {
			const m = merge3(base, body, o.data);
			if (!m.conflict) {
				markBeforeChange();
				text = m.text;
				textOrig = o.data;
				const r = await updateFile(rootFolder, targetFile, m.text, textHash(o.data));
				if (!r.ok) return void alert(r.message || '저장하지 못했습니다');
				textOrig = m.text;
				return toast('다른 곳의 수정과 합쳐 저장했습니다.');
			}
			const ok = await confirm({
				title: '같은 곳을 다른 곳에서도 고쳤습니다',
				body: '내 편집으로 덮어쓰면 그쪽에서 고친 내용이 사라집니다. 취소하면 저장하지 않고 내 편집을 그대로 둡니다.',
				okTitle: '내 편집으로 덮어쓰기',
				cancelTitle: '취소'
			});
			if (ok) return doSave(body, true);
			return;
		}
		if (!o.ok) return void alert(o.message || '저장하지 못했습니다');
		text = body;
		textOrig = body;
		toast('정상적으로 수정되었습니다.');
	}
	const MODES = [
		{ v: 'split', icon: 'vertical_split', t: '좌우로 나란히' },
		{ v: 'splitv', icon: 'horizontal_split', t: '위아래로 나란히' },
		{ v: 'edit', icon: 'edit_note', t: '에디터만' },
		{ v: 'view', icon: 'article', t: '미리보기만' }
	];
</script>

<PageContainer sx={{ paddingBottom: 0, position: 'relative' }} subtitle={targetFile}>
	{#if textOrig !== null}
		{#if canToggle}
			<div
				class={cx(
					'MuiStack-root',
					toClass(
						media.mdDown
							? { flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'flex-end', alignItems: 'center', gap: 8, marginBottom: 8 }
							: { flexDirection: 'row', justifyContent: 'flex-end', alignItems: 'center', position: 'absolute', right: 0, top: 30, '& > :not(:first-child)': { marginLeft: 8 } }
					)
				)}
			>
				{#each Object.entries(peers) as [id, p] (id)}
					<Tooltip title={p.editing ? `${p.name} — 저장하지 않은 편집이 있습니다` : `${p.name} — 보는 중`}>
						<Chip
							size="small"
							color={p.editing ? 'warning' : 'default'}
							variant={p.editing ? 'filled' : 'outlined'}
							icon={p.editing ? 'edit' : 'visibility'}
							iconSx={{ fontSize: 16 }}
							label={p.name}
						/>
					</Tooltip>
				{/each}
				<div class="MuiToggleButtonGroup-root adm-tbg" role="group">
					{#each MODES as o (o.v)}
						<Tooltip title={o.t}>
							<button
								type="button"
								class={cx('MuiButtonBase-root MuiToggleButton-root MuiToggleButton-sizeSmall MuiToggleButtonGroup-grouped', mode === o.v && 'Mui-selected')}
								aria-label={o.t}
								aria-pressed={mode === o.v}
								onclick={() => changeMode(o.v)}><Icon name={o.icon} size="Small" /></button
							>
						</Tooltip>
					{/each}
				</div>
			</div>
		{/if}
		<div
			class={cx(
				'MuiBox-root',
				toClass({ '& .md-preview-body > *': { transition: 'background-color 1.2s ease' }, '& .md-preview-body > .md-changed': { backgroundColor: 'rgba(255, 213, 79, 0.45)', transition: 'none' } })
			)}
		>
			<MdEditor
				markdown={text || ''}
				onChange={(v) => (text = v)}
				doSave={(v) => doSave(v)}
				{language}
				height="calc(100vh - 230px)"
				hideEditor={mode === 'view'}
				hidePreview={mode === 'edit'}
				splitDirection={mode === 'splitv' ? 'column' : 'row'}
				resizeKey="explorer"
				bind:previewEl
			/>
		</div>
		{#if mode !== 'view'}
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', marginTop: 8 }))}><Button disabled={text == textOrig} onclick={() => doSave()}>저장 (Ctrl+S)</Button></div>
		{/if}
	{/if}
</PageContainer>
