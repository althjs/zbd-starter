<script lang="ts">
	/**
	 * JSON 직접 고치기(개발자) — React JsonEditorDrawer 자리. 글로 고치고 「적용」 하면 객체를 onChange 로(저장은 부르는 쪽).
	 *   <JsonEditDialog bind:open value={obj} onChange={(o) => …} title="회원 JSON" />
	 */
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import { alert } from '../../ui/modal.svelte.ts';
	import JsonCode from '../../ui/JsonCode.svelte';

	let { open = $bindable(false), value, onChange, onclose, title = 'JSON 편집' }: { open?: boolean; value: any; onChange: (o: any) => void; onclose?: () => void; title?: string } = $props();
	const close = () => ((open = false), onclose?.());
	let text = $state('');
	$effect(() => {
		if (open) text = JSON.stringify(value, null, 2);
	});
	function apply() {
		try {
			onChange(JSON.parse(text));
			close();
		} catch (e: any) {
			alert(`JSON 형식 오류: ${e?.message || e}`);
		}
	}
</script>

<Dialog {open} onclose={close} {title} maxWidth="md" fullWidth mobileFullScreen>
	<JsonCode bind:value={text} height="65vh" onsave={apply} />
	{#snippet actions()}
		<Button color="inherit" onclick={close}>취소</Button>
		<Button variant="contained" onclick={apply}>적용</Button>
	{/snippet}
</Dialog>
