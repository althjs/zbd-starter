<script lang="ts">
	/**
	 * Monaco 코드 편집기(React @monaco-editor/react `<Editor>` 와 같은 틀: section(flex) 안에 Loading... → 편집기).
	 *   <CodeEditor height={600} language="html" value={html} theme="vs-dark" onChange={(t) => …} onMount={(ed) => …} />
	 * value 가 밖에서 바뀌면 편집기 내용을 바꾼다(커서는 끝으로 — React 판 executeEdits 와 같은 방식).
	 * defaultValue 만 주면 처음 값만 쓴다. 공용 후보.
	 */
	import { onDestroy } from 'svelte';
	import { loadMonaco } from './monaco.ts';

	let {
		value = undefined,
		defaultValue = '',
		language = 'html',
		theme = 'vs-dark',
		width = '100%',
		height = '100%',
		options = {},
		onChange,
		onMount
	}: {
		value?: string;
		defaultValue?: string;
		language?: string;
		theme?: string;
		width?: number | string;
		height?: number | string;
		options?: Record<string, any>;
		onChange?: (v: string) => void;
		onMount?: (editor: any, monaco: any) => void;
	} = $props();

	let box = $state<HTMLDivElement>();
	let ready = $state(false);
	let editor: any = null;
	let silent = false;
	const px = (v: number | string) => (typeof v == 'number' ? `${v}px` : v);

	$effect(() => {
		if (!box || editor) return;
		let dead = false;
		loadMonaco().then((monaco) => {
			if (dead || !box) return;
			const model = monaco.editor.createModel(value ?? defaultValue ?? '', language);
			editor = monaco.editor.create(box, { model, automaticLayout: true, ...options });
			monaco.editor.setTheme(theme);
			editor.onDidChangeModelContent(() => !silent && onChange?.(editor.getValue()));
			ready = true;
			onMount?.(editor, monaco);
		});
		return () => (dead = true);
	});
	$effect(() => {
		const v = value;
		if (!ready || !editor || v === undefined) return;
		if (v === editor.getValue()) return;
		silent = true;
		editor.executeEdits('', [{ range: editor.getModel().getFullModelRange(), text: v, forceMoveMarkers: true }]);
		editor.pushUndoStop();
		silent = false;
	});
	$effect(() => {
		if (ready) editor?.updateOptions(options);
	});
	onDestroy(() => {
		editor?.getModel()?.dispose();
		editor?.dispose();
		editor = null;
	});
</script>

<section style="display:flex;position:relative;text-align:initial;width:{px(width)};height:{px(height)}">
	{#if !ready}<div style="display:flex;height:100%;width:100%;justify-content:center;align-items:center">Loading...</div>{/if}
	<div bind:this={box} style="width:100%;{ready ? '' : 'display:none'}"></div>
</section>

<style>
	/* React App.css `.monaco-editor * { font-weight: 600 }` — 관리자 전체 규칙이라 바탕(admin.css)으로 옮길 후보 */
	:global(html.zbd-admin .monaco-editor *) {
		font-weight: 600;
	}
</style>
