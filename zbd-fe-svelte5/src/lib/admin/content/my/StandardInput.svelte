<script lang="ts">
	/** 밑줄 입력칸(MUI TextField variant="standard" small) — 내 계정 화면. mt=false 면 위 여백 없이(Stack spacing 첫 칸) */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';

	let { value = '', placeholder = '', type = 'text', mt = true, oninput }: { value?: string; placeholder?: string; type?: string; mt?: boolean; oninput: (v: string) => void } = $props();
	let focused = $state(false);
	const root = toClass({
		position: 'relative',
		display: 'flex',
		alignItems: 'center',
		fontSize: 16,
		lineHeight: '23px',
		color: 'var(--adm-text)',
		'&::before': { content: '""', position: 'absolute', left: 0, right: 0, bottom: 0, borderBottom: '1px solid rgba(0, 0, 0, 0.42)', pointerEvents: 'none', transition: 'border-bottom-color 200ms' },
		'&:hover::before': { borderBottom: '2px solid var(--adm-text)' },
		'&::after': {
			content: '""',
			position: 'absolute',
			left: 0,
			right: 0,
			bottom: 0,
			borderBottom: '2px solid var(--theme-color-main)',
			transform: 'scaleX(0)',
			transition: 'transform 200ms cubic-bezier(0, 0, 0.2, 1)',
			pointerEvents: 'none'
		},
		'&.Mui-focused::after': { transform: 'scaleX(1)' }
	});
	const input = toClass({
		font: 'inherit',
		fontSize: 14,
		fontWeight: 500,
		letterSpacing: 'inherit',
		color: 'currentColor',
		padding: '1px 0 5px',
		border: 0,
		boxSizing: 'content-box',
		background: 'none',
		height: '1.4375em',
		margin: 0,
		display: 'block',
		minWidth: 0,
		width: '100%',
		outline: 0
	});
</script>

<div class={cx('MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root', toClass({ display: 'flex', flexDirection: 'column', position: 'relative', marginTop: mt ? 8 : 0 }))}>
	<div
		class={cx('MuiInputBase-root MuiInput-root MuiInput-underline MuiInputBase-colorPrimary MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-sizeSmall', root, focused && 'Mui-focused')}
	>
		<input
			class={cx('MuiInputBase-input MuiInput-input MuiInputBase-inputSizeSmall', input)}
			{type}
			{placeholder}
			autocomplete="one-time-code"
			{value}
			oninput={(e) => oninput((e.currentTarget as HTMLInputElement).value)}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
		/>
	</div>
</div>
