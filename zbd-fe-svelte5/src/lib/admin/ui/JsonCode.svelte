<script lang="ts">
	/**
	 * JSON 편집 칸 — Monaco(json, vs-dark, 문법 색칠·접기·괄호 짝·오류 표시). 관리자 「{}」(data_object) 편집 창들이 같이 쓴다(2026-10-03 — 그 전엔 글자 칸).
	 *   <JsonCode bind:value={text} height="60vh" onsave={(v) => apply()} />   Ctrl/⌘+S = onsave
	 * 처음 값으로 편집기를 만들고, 밖에서 value 가 바뀌면 그 값으로 바꾼다(창을 다시 열 때).
	 */
	import CodeEditor from '#lib/admin/kit/CodeEditor.svelte';

	let { value = $bindable(''), height = '60vh', readonly = false, onsave }: { value?: string; height?: number | string; readonly?: boolean; onsave?: (v: string) => void } = $props();
</script>

<div class="json-code" style="height:{typeof height == 'number' ? `${height}px` : height}">
	<CodeEditor
		{value}
		language="json"
		theme="vs-dark"
		options={{ tabSize: 2, readOnly: readonly, minimap: { enabled: false }, wordWrap: 'on', scrollBeyondLastLine: false, fontSize: 13 }}
		onChange={(t) => (value = t)}
		onMount={(ed, monaco) => {
			ed.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => onsave?.(ed.getValue()));
		}}
	/>
</div>

<style>
	.json-code {
		width: 100%;
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid var(--adm-divider, #e4e7ec);
	}
</style>
