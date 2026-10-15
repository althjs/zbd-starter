<script lang="ts">
	/**
	 * 태그 입력(React components/common/TagInput) — 값은 콤마로 이은 문자열("a,b,c"). 엔터·콤마로 더하고 ✕ 로 뺀다.
	 *   <TagInput tags={item.tags} onChange={(v) => (item.tags = v)} placeholder="콤마로 구분" />
	 * inputFilter 로 넣기 전에 글자를 고칠 수 있다(예: 소문자). readonly 면 칩만.
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Icon from './Icon.svelte';

	let {
		tags = '',
		onChange,
		placeholder = '콤마 "," 로 태그가 구분됩니다.',
		inputFilter,
		readonly = false,
		thresholdLength = 1000,
		sx = null
	}: { tags?: string; onChange: (v: string) => void; placeholder?: string; inputFilter?: (t: string) => string; readonly?: boolean; thresholdLength?: number; sx?: Sx | null } = $props();
	const list = $derived(
		String(tags || '')
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean)
	);
	let text = $state('');
	function add(raw: string) {
		const add = raw
			.split(',')
			.map((s) => (inputFilter ? inputFilter(s.trim()) : s.trim()))
			.filter(Boolean);
		const next = [...list];
		for (const a of add) if (!next.includes(a)) next.push(a);
		const v = next.join(',');
		if (v.length > thresholdLength) return;
		onChange(v);
		text = '';
	}
	const remove = (t: string) => onChange(list.filter((x) => x != t).join(','));
</script>

<div class={cx('MuiStack-root adm-tags', toClass({ flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 4, ...(sx || {}) }))}>
	{#each list as t (t)}
		<div class="MuiChip-root MuiChip-filled MuiChip-sizeSmall MuiChip-colorDefault MuiChip-filledDefault">
			<span class="MuiChip-label MuiChip-labelSmall">{t}</span>
			{#if !readonly}<Icon name="cancel" class="MuiChip-deleteIcon" sx={{ fontSize: 16, marginRight: 4, cursor: 'pointer', color: 'var(--adm-action)' }} onclick={() => remove(t)} />{/if}
		</div>
	{/each}
	{#if !readonly}
		<input
			class="adm-tags-input"
			{placeholder}
			bind:value={text}
			onkeydown={(e) => {
				if ((e.key == 'Enter' || e.key == ',') && !e.isComposing) {
					e.preventDefault();
					if (text.trim()) add(text);
				} else if (e.key == 'Backspace' && !text && list.length) remove(list[list.length - 1]);
			}}
			onblur={() => text.trim() && add(text)}
		/>
	{/if}
</div>
