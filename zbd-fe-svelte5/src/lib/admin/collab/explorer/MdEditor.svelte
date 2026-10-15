<script lang="ts">
	/**
	 * 마크다운 편집 + 미리보기(React components/common/markdown/MarkdownChartEditorForm).
	 * 좌우·위아래 나란히, 경계를 끌어 비율 조절(PC — 비율은 이 브라우저에 resizeKey 로 기억, 두 번 누르면 반반). Ctrl+S = doSave.
	 * React 는 편집기가 Monaco(+@멘션 자동 완성·AI 도우미 버튼)였다 — 여기선 CodeArea(줄 번호 있는 글 칸).
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { media } from '../../media.svelte.ts';
	import CodeArea from '../ui/CodeArea.svelte';
	import MarkdownView from '../ui/MarkdownView.svelte';
	import { MD_BOX_SX } from '../ui/md.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	let {
		markdown,
		onChange,
		doSave,
		height = 'calc(100vh - 260px)',
		minHeight = '600px',
		language = 'markdown',
		hideEditor = false,
		hidePreview = false,
		splitDirection = 'row',
		resizeKey = '',
		previewEl = $bindable()
	}: {
		markdown: string;
		onChange?: (md: string) => void;
		doSave?: (text: string) => void;
		height?: string;
		minHeight?: string;
		language?: string;
		hideEditor?: boolean;
		hidePreview?: boolean;
		splitDirection?: 'row' | 'column';
		resizeKey?: string;
		previewEl?: HTMLElement;
	} = $props();

	const both = $derived(!hideEditor && !hidePreview && language == 'markdown');
	const dir = $derived<'row' | 'column'>(splitDirection === 'column' || media.mdDown ? 'column' : 'row');
	const resizable = $derived(both && !media.mdDown && !!resizeKey);
	const ratioKey = $derived(SK.explorer.mdSplitRatio(resizeKey, dir));
	let ratio = $state(0.5);
	$effect(() => {
		let v = 0;
		try {
			v = Number(localStorage.getItem(ratioKey));
		} catch {}
		ratio = v > 0.1 && v < 0.9 ? v : 0.5;
	});
	let splitEl = $state<HTMLElement>();
	let dragging = $state(false);
	function move(e: PointerEvent) {
		if (!dragging || !splitEl) return;
		const r = splitEl.getBoundingClientRect();
		const v = dir === 'row' ? (e.clientX - r.left) / r.width : (e.clientY - r.top) / r.height;
		ratio = Math.min(0.85, Math.max(0.15, v));
	}
	function up() {
		if (!dragging) return;
		dragging = false;
		try {
			localStorage.setItem(ratioKey, String(ratio));
		} catch {}
	}
	let text = $state('');
	$effect(() => {
		text = markdown;
	});
	const stripIAL = (md: string) => (md ? md.replace(/(\]\([^)]+\))\{:[^}]*\}/g, '$1') : md);
	const gt200 = 'var(--gt-200, #f2f2f2)';
	const boxSx = MD_BOX_SX;
	const stackSx = $derived(hideEditor ? { marginBottom: 40 } : { minHeight, height, ...(dragging ? { userSelect: 'none', cursor: dir === 'row' ? 'col-resize' : 'row-resize' } : {}) });
	const gap = $derived(resizable ? {} : dir == 'row' ? { '& > :not(:first-child)': { marginLeft: 16 } } : { '& > :not(:first-child)': { marginTop: 16 } });
</script>

<div bind:this={splitEl} class={cx('MuiStack-root', toClass({ flexDirection: dir, ...stackSx, ...gap }))}>
	{#if !hideEditor}
		<div
			class={cx(
				'MuiPaper-root MuiPaper-elevation MuiPaper-elevation4',
				toClass({
					flex: resizable ? `0 0 calc(${(ratio * 100).toFixed(2)}% - 6px)` : 1,
					minWidth: 0,
					minHeight: 0,
					...(dragging ? { pointerEvents: 'none' } : {}),
					display: 'flex',
					flexDirection: 'column',
					overflow: 'hidden',
					border: `1px solid ${gt200}`,
					backgroundColor: '#1e1e1e',
					borderRadius: 0,
					boxShadow: '0px 2px 4px -1px rgba(0,0,0,0.2),0px 4px 5px 0px rgba(0,0,0,0.14),0px 1px 10px 0px rgba(0,0,0,0.12)'
				})
			)}
		>
			<CodeArea bind:value={text} oninput={(v) => onChange?.(v)} onsave={(v) => doSave?.(v)} />
		</div>
	{/if}
	{#if resizable}
		<div
			class={cx(
				'MuiBox-root',
				toClass({
					flex: '0 0 12px',
					cursor: dir === 'row' ? 'col-resize' : 'row-resize',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					touchAction: 'none',
					'&::after': {
						content: '""',
						display: 'block',
						borderRadius: 2,
						backgroundColor: dragging ? 'var(--theme-color-main)' : 'var(--gt-300, #e0e0e0)',
						...(dir === 'row' ? { width: 4, height: 40 } : { width: 40, height: 4 })
					},
					'&:hover::after': { backgroundColor: 'var(--theme-color-main)' }
				})
			)}
			title="끌어서 크기 조절 · 두 번 누르면 반반"
			role="separator"
			onpointerdown={(e) => {
				e.preventDefault();
				(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
				dragging = true;
			}}
			onpointermove={move}
			onpointerup={up}
			onpointercancel={up}
			ondblclick={() => {
				ratio = 0.5;
				try {
					localStorage.setItem(ratioKey, '0.5');
				} catch {}
			}}
		></div>
	{/if}
	{#if language == 'markdown' && !hidePreview}
		<div
			class={cx(
				'MuiPaper-root MuiPaper-elevation MuiPaper-elevation0',
				toClass({
					flex: 1,
					minWidth: 0,
					minHeight: 0,
					...(dragging ? { pointerEvents: 'none' } : {}),
					display: 'flex',
					flexDirection: 'column',
					overflow: 'hidden',
					border: `1px solid ${gt200}`,
					background: 'linear-gradient(135deg, #ffffff 0%, #f9fafb 100%)',
					boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.07)',
					borderRadius: 0
				})
			)}
		>
			<div bind:this={previewEl} class={cx('MuiBox-root md-preview-wrap', toClass({ flex: 1, paddingLeft: 16, paddingRight: 16, overflowY: 'auto', ...boxSx }))}>
				<MarkdownView text={stripIAL(text)} class="md-preview-body" />
			</div>
		</div>
	{/if}
</div>
