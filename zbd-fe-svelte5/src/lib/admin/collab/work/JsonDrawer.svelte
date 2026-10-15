<script lang="ts">
	/** 개발자 — 문서·도형 JSON 그대로 고치기(React JsonEditorDrawer). 저장 전에 JSON 인지 확인한다. */
	import AdminDrawer from '../../ui/AdminDrawer.svelte';
	import Button from '../../ui/Button.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import JsonCode from '../../ui/JsonCode.svelte';

	let { value, onclose, onsave, btnText = '저장' }: { value: any; onclose: () => void; onsave: (v: any) => void; btnText?: string } = $props();
	let text = $state('');
	$effect(() => {
		text = value ? JSON.stringify(value, null, 2) : '';
	});
	function save() {
		let v: any;
		try {
			v = JSON.parse(text);
		} catch {
			return void alert('JSON 형식이 아닙니다.');
		}
		onsave(v);
		onclose();
	}
</script>

<AdminDrawer open={!!value} {onclose} title="JSON 편집">
	<JsonCode bind:value={text} height="calc(100vh - 160px)" onsave={save} />
	<div style="margin-top:8px"><Button variant="contained" onclick={save}>{btnText}</Button></div>
</AdminDrawer>
