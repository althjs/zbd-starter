<script lang="ts">
	/**
	 * 색 입력(React mui-color-input 의 MuiColorInput — 왼쪽 색 칸 + hex 글자, 칸을 누르면 고르기). 값은 #rrggbb(aa).
	 *   <ColorField value={c} onChange={(v) => …} sx={{ height: 40 }} />
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let { value = '', label = '', onChange, sx = null }: { value?: string; label?: string; onChange: (v: string) => void; sx?: Sx | null } = $props();
	let picker = $state<HTMLInputElement>();
	const id = `adm-cf-${Math.random().toString(36).slice(2, 8)}`;
	const hex6 = $derived(/^#[0-9a-f]{6}/i.test(value) ? value.slice(0, 7) : '#000000');
	const alpha = $derived(/^#[0-9a-f]{8}$/i.test(value) ? value.slice(7) : value.length == 9 ? '' : 'ff');
</script>

<div class={cx('MuiFormControl-root MuiTextField-root MuiColorInput-TextField', toClass(sx))}>
	{#if label}<label for={id} class="MuiFormLabel-root MuiInputLabel-root MuiInputLabel-formControl MuiInputLabel-animated MuiInputLabel-shrink MuiInputLabel-outlined">{label}</label>{/if}
	<div class={cx('MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-formControl MuiInputBase-adornedStart', !label && 'MuiInputBase-sizeSmall')}>
		<span class={toClass({ display: 'flex', alignItems: 'center', marginRight: 8 })}>
			<button
				type="button"
				aria-label="색 고르기"
				class={toClass({
					width: 24,
					height: 24,
					padding: 0,
					border: 0,
					borderRadius: 4,
					cursor: 'pointer',
					...(value ? { backgroundColor: value } : { backgroundImage: 'repeating-conic-gradient(#ccc 0% 25%, #fff 0% 50%)', backgroundSize: '8px 8px' }),
					boxShadow: 'inset 0 0 0 1px rgba(0,0,0,0.1)'
				})}
				onclick={() => picker?.click()}
			></button>
			<input
				bind:this={picker}
				type="color"
				value={hex6}
				class={toClass({ width: 0, height: 0, opacity: 0, border: 0, padding: 0, position: 'absolute' })}
				oninput={(e) => onChange((e.currentTarget as HTMLInputElement).value + (alpha || ''))}
			/>
		</span>
		<input
			{id}
			class={cx('MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputAdornedStart', !label && 'MuiInputBase-inputSizeSmall')}
			{value}
			oninput={(e) => onChange((e.currentTarget as HTMLInputElement).value)}
		/>
		<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline">
			<legend class={cx(label && 'adm-of-notch')}
				>{#if label}<span>{label}</span>{:else}<span class="notranslate">​</span>{/if}</legend
			>
		</fieldset>
	</div>
</div>
