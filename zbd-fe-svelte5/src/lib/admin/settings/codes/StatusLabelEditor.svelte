<script lang="ts">
	/**
	 * 상태 라벨 편집기(React codes/StatusLabelEditor) — 색 + 이름 목록, 끌어서 순서. 저장은 부르는 쪽(value/onChange).
	 *   fixedRows 면 줄 추가·삭제 잠금(이름·색만)
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Popover from '../../ui/Popover.svelte';
	import { moveItem, sortable } from '#lib/shared/ui/sortable.ts';
	import { newStatusLabel, STATUS_LABEL_THEMES } from './codes.ts';

	let { value = [], onChange, disabled = false, fixedRows = false }: { value: any[]; onChange: (list: any[]) => void; disabled?: boolean; fixedRows?: boolean } = $props();
	const list = $derived(value || []);
	let theme = $state<{ el: HTMLElement; idx: number } | null>(null);
	const patch = (idx: number, k: string, v: string) => onChange(list.map((x, i) => (i == idx ? { ...x, [k]: v } : x)));
	const hex = (v: string, d: string) => (/^#[0-9a-f]{6}/i.test(v || '') ? v.slice(0, 7) : d);
	const swatch = toClass({
		width: 24,
		height: 24,
		padding: 0,
		border: 0,
		borderRadius: 4,
		cursor: 'pointer',
		position: 'relative',
		boxShadow: 'inset 0 0 0 1px rgba(0,0,0,0.1)',
		display: 'inline-flex',
		alignItems: 'center',
		justifyContent: 'center',
		fontWeight: 'bold',
		fontSize: 14,
		textShadow: '0 0 2px #fff'
	});
</script>

<div class="MuiBox-root">
	<div class={toClass({ display: 'flex', flexDirection: 'column', gap: 8 })} use:sortable={{ handle: '.drag-handle', disabled, onmove: (f, t) => onChange(moveItem(list, f, t)) }}>
		{#each list as item, idx (item.id || `temp_${idx}`)}
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 8 }))}>
				<div
					class={cx(
						'drag-handle MuiBox-root',
						toClass({ cursor: 'grab', color: 'var(--adm-text2)', display: 'inline-flex', alignItems: 'center', touchAction: 'none', '& > *': { pointerEvents: 'none' } })
					)}
				>
					<Icon name="drag_indicator" sx={{ fontSize: 20 }} />
				</div>
				<label class={swatch} style="background-color:{item.bgColor || '#f5f5f5'}">
					<input
						type="color"
						{disabled}
						value={hex(item.bgColor, '#f5f5f5')}
						class={toClass({ position: 'absolute', inset: 0, opacity: 0, cursor: 'pointer' })}
						oninput={(e) => patch(idx, 'bgColor', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class={swatch} style="background-color:{item.color || '#000'};color:{item.color || '#000'}">
					<span class={toClass({ color: 'inherit', filter: 'invert(0)' })}>A</span>
					<input
						type="color"
						{disabled}
						value={hex(item.color, '#000000')}
						class={toClass({ position: 'absolute', inset: 0, opacity: 0, cursor: 'pointer' })}
						oninput={(e) => patch(idx, 'color', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<div class={cx('MuiFormControl-root MuiTextField-root', toClass({ flexGrow: 1 }))}>
					<!-- 바탕색은 바깥 상자에 — 안쪽 입력칸에 칠하면 모서리 둥글기가 테두리와 달라 모서리에 바탕이 삐져나왔다(2026-10-04) -->
					<div
						class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-formControl MuiInputBase-sizeSmall"
						style="background-color:{item.bgColor || '#fff'};color:{item.color || '#000'}"
					>
						<input
							class={cx(
								'MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall',
								toClass({ '&&': { padding: '4px 8px', fontSize: 12, fontWeight: 'bold', background: 'transparent', color: 'inherit' } })
							)}
							placeholder="라벨 이름"
							{disabled}
							value={item.name || ''}
							oninput={(e) => patch(idx, 'name', (e.currentTarget as HTMLInputElement).value)}
						/>
						<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate">​</span></legend></fieldset>
					</div>
				</div>
				<IconButton size="Small" {disabled} sx={{ padding: 4 }} title="색 테마 선택" onclick={(e: MouseEvent) => (theme = { el: e.currentTarget as HTMLElement, idx })}
					><Icon name="palette" sx={{ fontSize: 18, color: 'var(--adm-text2)' }} /></IconButton
				>
				{#if !fixedRows}
					<IconButton size="Small" {disabled} sx={{ padding: 4 }} aria-label="삭제" onclick={() => onChange(list.filter((_, i) => i !== idx))}
						><Icon name="delete_outline" sx={{ fontSize: 18, color: 'var(--adm-error-main)' }} /></IconButton
					>
				{/if}
			</div>
		{/each}
	</div>
	{#if !fixedRows}
		<Button
			size="small"
			{disabled}
			sx={{ width: '100%', marginTop: 8, fontSize: 12, border: '1px dashed #ddd', paddingTop: 8, paddingBottom: 8 }}
			onclick={() => onChange([...list, newStatusLabel(list)])}>+ 새로운 라벨</Button
		>
	{/if}
</div>

<Popover anchor={theme?.el || null} align="right" onclose={() => (theme = null)} paperSx={{ padding: 12, borderRadius: 16, maxWidth: 320 }}>
	<div class={toClass({ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 6 })}>
		{#each STATUS_LABEL_THEMES as t (t.bg + t.fg)}
			<button
				type="button"
				class={toClass({
					padding: '4px 8px',
					borderRadius: 4,
					fontSize: 12,
					fontWeight: 'bold',
					textAlign: 'center',
					cursor: 'pointer',
					whiteSpace: 'nowrap',
					overflow: 'hidden',
					textOverflow: 'ellipsis',
					border: '1px solid rgba(0,0,0,0.08)',
					'&:hover': { outline: '2px solid rgba(0,0,0,0.35)' }
				})}
				style="background-color:{t.bg};color:{t.fg}"
				onclick={() => {
					if (!theme) return;
					const i = theme.idx;
					onChange(list.map((x, j) => (j == i ? { ...x, bgColor: t.bg, color: t.fg } : x)));
					theme = null;
				}}>{(theme && list[theme.idx]?.name) || '라벨'}</button
			>
		{/each}
	</div>
</Popover>
