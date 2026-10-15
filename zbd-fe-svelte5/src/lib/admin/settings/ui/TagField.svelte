<script lang="ts">
	/**
	 * 태그 입력(React components/common/TagInput 모양 그대로) — 칩 줄 위, 밑줄 입력칸 아래(+ 글자 수 n/1000).
	 * 값은 "a,b,c". 엔터·콤마·칸 벗어나기로 더하고, 칩 ✕ 로 뺀다. 이미 있으면 토스트.
	 *   <TagField tags={v} onChange={(t) => (v = t)} disableThreshold placeholder="…" inputFilter={(s) => s.toLowerCase()} />
	 * (바탕 ui/TagInput 은 칩과 입력이 한 줄이라 React 와 모양이 달라 이 묶음은 이걸 쓴다 — 공용 후보)
	 */
	import '../settings.css';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { toast } from '../../ui/modal.svelte.ts';

	let {
		tags = '',
		onChange,
		placeholder = '콤마 "," 로 태그가 구분됩니다.',
		thresholdLength = 1000,
		disableThreshold = false,
		readonly = false,
		inputFilter,
		sx = null
	}: {
		tags?: string;
		onChange: (v: string) => void;
		placeholder?: string;
		thresholdLength?: number;
		disableThreshold?: boolean;
		readonly?: boolean;
		inputFilter?: (t: string) => string;
		sx?: Sx | null;
	} = $props();
	const list = $derived(
		String(tags || '')
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean)
	);
	const length = $derived(list.join(',').length);
	let text = $state('');
	function add(v: string) {
		v = (v || '').trim();
		if (!v) return;
		if (list.includes(v)) return toast(`"${v}" 는 이미 등록된 태그입니다.`);
		text = '';
		onChange([...list, v].join(','));
	}
	const remove = (t: string) => onChange(list.filter((x) => x != t).join(','));
	const over = $derived(length >= thresholdLength);
</script>

<div class={cx('MuiStack-root', toClass({ overflow: 'hidden', ...(sx || {}) }))}>
	<ul class={cx('MuiStack-root', toClass({ flexDirection: 'row', display: 'flex', flexWrap: 'wrap', listStyle: 'none', padding: '4px 0', margin: 0, marginLeft: -4.8 }))}>
		{#each list as t (t)}
			<li class={toClass({ margin: 4 })}>
				<div class={cx('MuiChip-root MuiChip-filled MuiChip-sizeSmall MuiChip-colorDefault MuiChip-filledDefault', !readonly && 'MuiChip-deletable', toClass({ borderRadius: 4 }))}>
					<span class="MuiChip-label MuiChip-labelSmall">{t}</span>
					{#if !readonly}
						<svg
							class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiChip-deleteIcon MuiChip-deleteIconSmall adm-chip-del"
							role="button"
							tabindex="-1"
							aria-label="지우기"
							onclick={() => remove(t)}
							onkeydown={() => {}}
							focusable="false"
							viewBox="0 0 24 24"
							><path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z"
							></path></svg
						>
					{/if}
				</div>
			</li>
		{/each}
	</ul>
	<div class={cx('MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root')}>
		<div
			class={cx(
				'MuiInputBase-root MuiInput-root MuiInput-underline MuiInputBase-colorPrimary MuiInputBase-fullWidth adm-std-input',
				(over || readonly) && 'Mui-disabled',
				!disableThreshold && 'MuiInputBase-adornedEnd'
			)}
		>
			<input
				class="MuiInputBase-input MuiInput-input"
				placeholder={!disableThreshold && over ? '최대 입력 글자 수가 초과되었습니다.' : placeholder}
				disabled={over || readonly}
				value={text}
				oninput={(e) => {
					let v = (e.currentTarget as HTMLInputElement).value;
					if (v.includes(',')) {
						add(v.split(',')[0]);
						(e.currentTarget as HTMLInputElement).value = text;
						return;
					}
					if (inputFilter) v = inputFilter(v);
					text = v;
					(e.currentTarget as HTMLInputElement).value = v;
				}}
				onkeyup={(e) => e.key == 'Enter' && !e.isComposing && (e.preventDefault(), add(text.split(',')[0]))}
				onblur={() => text && add(text.split(',')[0])}
			/>
			{#if !disableThreshold}<span class={cx('MuiTypography-root MuiTypography-overline', toClass({ alignSelf: 'end', marginBottom: 0, color: '#ddd' }))}>{length}/{thresholdLength}</span>{/if}
		</div>
	</div>
</div>
