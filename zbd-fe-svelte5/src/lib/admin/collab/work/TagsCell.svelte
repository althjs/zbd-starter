<script lang="ts">
	/** 태그·이메일 열(React TagInput isModal) — 이름표들(없으면 안내 글), 누르면 고치는 창 */
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import TagInput from '../../ui/TagInput.svelte';
	import Chip from '../ui/Chip.svelte';

	let {
		tags = '',
		onchange,
		readonly = false,
		emptyText = '태그 없음 (클릭하여 추가)',
		modalTitle = '태그 수정',
		placeholder = '콤마 "," 로 태그가 구분됩니다.'
	}: { tags?: string; onchange: (v: string) => void; readonly?: boolean; emptyText?: string; modalTitle?: string; placeholder?: string } = $props();
	let open = $state(false);
	const list = $derived(
		String(tags || '')
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean)
	);
</script>

<div class="MuiBox-root" style="min-width:120px">
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
	<ul class="MuiStack-root" style="display:flex;flex-direction:row;flex-wrap:wrap;list-style:none;padding:4px 0;margin:0 0 0 -4.8px;cursor:pointer" onclick={() => !readonly && (open = true)}>
		{#each list as t, i (i)}
			<li style="margin-left:4.8px;margin-bottom:4px"><Chip size="small" label={t} sx={{ borderRadius: 4 }} /></li>
		{:else}
			<span class="MuiTypography-root MuiTypography-caption" style="margin-left:8px;padding:4px 0;color:var(--adm-text2)">{emptyText}</span>
		{/each}
	</ul>
</div>
<Dialog {open} onclose={() => (open = false)} title={modalTitle} maxWidth="xs" fullWidth>
	<TagInput {tags} onChange={(v) => onchange(v)} {placeholder} />
	{#snippet actions()}<Button variant="contained" size="small" onclick={() => (open = false)}>닫기</Button>{/snippet}
</Dialog>
