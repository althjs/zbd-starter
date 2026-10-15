<script lang="ts">
	/**
	 * 순서 변경 버튼 + 창 (React components/common/ContentsSortDialog) — 목록을 끌어 순서를 바꾸고 「적용」.
	 *   <SortDialog items={[{ id: 1, name: '공지' }, …]} onChange={(sorted) => save(sorted)} />
	 *   btnText · okText · objectKey(id 칸 이름, 기본 'id') · variant · disabled
	 * 버튼 없이 열려면 bind:open.
	 */
	import type { Sx } from '#lib/shared/mui/sx.ts';
	import Button from './Button.svelte';
	import Dialog from './Dialog.svelte';
	import Icon from './Icon.svelte';
	import SortList from './SortList.svelte';

	let {
		items,
		onChange,
		btnText = '순서 변경',
		okText = '적용',
		title = '순서 변경',
		objectKey = 'id',
		variant = 'outlined',
		size = 'medium',
		disabled = false,
		showButton = true,
		sx = null,
		open = $bindable(false)
	}: {
		items: any[];
		onChange: (sorted: any[]) => void;
		btnText?: string;
		okText?: string;
		title?: string;
		objectKey?: string;
		variant?: 'outlined' | 'contained' | 'text';
		size?: 'small' | 'medium' | 'large';
		disabled?: boolean;
		showButton?: boolean;
		sx?: Sx | null;
		open?: boolean;
	} = $props();

	let rows = $state<any[]>([]);
	$effect(() => {
		if (open) rows = [...items];
	});
	const changed = $derived(rows.map((o) => o?.[objectKey]).join('|') != items.map((o) => o?.[objectKey]).join('|'));
</script>

{#if showButton}
	<Button {variant} {size} {disabled} {sx} onclick={() => (open = true)}>
		{#snippet startIcon()}<Icon name="drag_handle" size="Small" />{/snippet}
		{btnText}
	</Button>
{/if}
<Dialog {open} onclose={() => (open = false)} {title} maxWidth="sm" fullWidth mobileFullScreen>
	<SortList items={rows} label={(o) => o?.name ?? o?.title ?? String(o?.[objectKey])} onchange={(n) => (rows = n)} />
	{#snippet actions()}
		<Button onclick={() => (open = false)} color="inherit">취소</Button>
		<Button
			variant="contained"
			disabled={!changed}
			onclick={() => {
				onChange(rows);
				open = false;
			}}>{okText}</Button
		>
	{/snippet}
</Dialog>
