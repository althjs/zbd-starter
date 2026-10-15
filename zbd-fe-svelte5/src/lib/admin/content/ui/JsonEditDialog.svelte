<script lang="ts">
	/**
	 * 개발자용 JSON 바로 고치기(React components/common/JsonEditorDrawer 자리 — 바탕에 아직 없어 콘텐츠 묶음이 간단히 둔다).
	 * 글자 칸에 JSON 을 고치고 「적용」 하면 onchange(객체). 저장은 부르는 화면이 따로 한다.
	 */
	import JsonCode from '../../ui/JsonCode.svelte';
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import { alert } from '../../ui/modal.svelte.ts';

	let { open, value, onclose, onchange }: { open: boolean; value: any; onclose: () => void; onchange: (v: any) => void } = $props();
	let text = $state('');
	$effect(() => {
		if (open) text = JSON.stringify($state.snapshot(value), null, 2);
	});
	function apply() {
		try {
			onchange(JSON.parse(text));
			onclose();
		} catch (e: any) {
			alert(`JSON 형식이 올바르지 않습니다.\n${e?.message || ''}`);
		}
	}
</script>

<Dialog {open} {onclose} title="JSON 편집" maxWidth="lg" fullWidth zIndex={1500}>
	<JsonCode bind:value={text} height="65vh" onsave={apply} />
	{#snippet actions()}
		<Button color="inherit" onclick={onclose}>취소</Button>
		<Button variant="contained" onclick={apply}>적용</Button>
	{/snippet}
</Dialog>
