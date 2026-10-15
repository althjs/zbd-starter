<script lang="ts">
	/**
	 * 관리자 알림·확인·입력·내용 창 + 토스트를 그린다 — 관리자 레이아웃에 하나만 둔다. 부르는 법은 modal.svelte.ts.
	 * 모양은 React ModalDialog(가운데 종이, 큰 버튼 두 개) · Toast(왼쪽 아래 filled Alert) 그대로.
	 */
	import { fly } from 'svelte/transition';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Icon from './Icon.svelte';
	import SvgIcon from './SvgIcon.svelte';
	import { closeModal, modalState, toastState } from './modal.svelte.ts';

	let input = $state('');
	let error = $state(false);
	let inputEl = $state<HTMLInputElement>();

	$effect(() => {
		const c = modalState.cur;
		if (!c) return;
		input = c.kind == 'prompt' ? (c.data.promptNumeric ? fmt(String(c.data.promptDefaultValue ?? '')) : c.data.promptDefaultValue || '') : '';
		error = false;
		if (c.kind == 'prompt') setTimeout(() => inputEl?.focus(), 10);
	});

	const raw = (v: string) => String(v ?? '').replace(/(?!^-)[^0-9]/g, '');
	function fmt(v: string) {
		const r = raw(v);
		if (r === '' || r === '-') return r;
		const n = Number(r);
		return Number.isFinite(n) ? n.toLocaleString() : r;
	}

	function ok() {
		const c = modalState.cur;
		if (!c) return;
		if (c.kind == 'prompt') {
			const v = c.data.promptNumeric ? raw(input) : input;
			if (!c.data.promptAllowEmpty && v === '') {
				error = true;
				return;
			}
			closeModal(v);
		} else closeModal(true);
	}
	function onKey(e: KeyboardEvent) {
		const c = modalState.cur;
		if (!c) return;
		if (e.key == 'Enter' && c.kind != 'modal') {
			e.preventDefault();
			ok();
		} else if (e.key == 'Escape') {
			e.preventDefault();
			closeModal(false);
		}
	}
	const ICON: Record<string, string> = { success: 'SuccessOutlined', error: 'ErrorOutline', warning: 'ReportProblemOutlined', info: 'InfoOutlined' };
	const cap = (s: string) => s[0].toUpperCase() + s.slice(1);
</script>

<svelte:window onkeydown={onKey} />

{#if modalState.cur}
	{@const c = modalState.cur}
	<div use:portal role="presentation" class="MuiModal-root adm-modal">
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true" onclick={() => !c.data.persistent && closeModal(false)}></div>
		<div class="MuiBox-root adm-modal-box modal-box" style:min-width={typeof c.data.width == 'number' ? `${c.data.width}px` : c.data.width || '300px'}>
			<div class="MuiContainer-root adm-modal-container" class:MuiContainer-maxWidthSm={!c.data.hiddenIcon}>
				<div class="MuiPaper-root MuiPaper-elevation MuiPaper-elevation12 adm-modal-paper" style:min-width={typeof c.data.width == 'number' ? `${c.data.width}px` : c.data.width || '300px'}>
					{#if c.data.closeUse}
						<Icon name="close" class="adm-modal-x" onclick={() => closeModal(false)} />
					{/if}
					<div class="MuiBox-root">
						{#if c.data.title}<h6 class="MuiTypography-root MuiTypography-h6 adm-modal-title">{c.data.title}</h6>{/if}
						{#if c.data.body}<div class="MuiTypography-root MuiTypography-body2 MuiTypography-alignCenter adm-modal-body">{c.data.body}</div>{/if}
						{#if c.data.child}{@render c.data.child()}{/if}
						{#if c.kind == 'prompt'}
							<div class="MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root adm-modal-input">
								<div
									class="MuiInputBase-root MuiFilledInput-root MuiInputBase-colorPrimary MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-sizeSmall MuiInputBase-hiddenLabel Mui-focused"
									class:Mui-error={error}
								>
									<input
										bind:this={inputEl}
										class="MuiInputBase-input MuiFilledInput-input MuiInputBase-inputSizeSmall MuiInputBase-inputHiddenLabel"
										placeholder={c.data.placeholder || ''}
										inputmode={c.data.promptNumeric ? 'numeric' : undefined}
										style:text-align={c.data.promptNumeric ? 'right' : undefined}
										value={input}
										oninput={(e) => {
											const v = (e.currentTarget as HTMLInputElement).value;
											input = c.data.promptNumeric ? fmt(v) : v;
											if (c.data.promptNumeric) (e.currentTarget as HTMLInputElement).value = input;
										}}
									/>
								</div>
							</div>
						{/if}
					</div>
					{#if c.kind != 'modal'}
						<div class="MuiStack-root max adm-modal-actions">
							{#if c.kind == 'prompt' || c.kind == 'confirm'}
								<button
									type="button"
									class="MuiButtonBase-root MuiButton-root MuiButton-contained MuiButton-containedInherit MuiButton-sizeMedium MuiButton-containedSizeMedium MuiButton-colorInherit MuiButton-fullWidth adm-modal-btn"
									onclick={() => closeModal(false)}
								>
									<p class="MuiTypography-root MuiTypography-body1">{c.data.cancelTitle || '취소'}</p>
								</button>
							{/if}
							<button
								type="button"
								class="MuiButtonBase-root MuiButton-root MuiButton-contained MuiButton-containedPrimary MuiButton-sizeMedium MuiButton-containedSizeMedium MuiButton-colorPrimary MuiButton-fullWidth adm-modal-btn"
								onclick={ok}>{c.data.okTitle || '확인'}</button
							>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

{#if toastState.cur?.show}
	{@const t = toastState.cur}
	<div use:portal class="MuiSnackbar-root MuiSnackbar-anchorOriginBottomLeft adm-snackbar" transition:fly={{ x: -400, duration: 225 }}>
		<div
			class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation0 MuiAlert-root MuiAlert-colorSuccess MuiAlert-filled MuiAlert-filled{cap(t.type || 'success')} adm-toast"
			role="alert"
		>
			<div class="MuiAlert-icon"><SvgIcon name={ICON[t.type || 'success']} size="Inherit" /></div>
			<div class="MuiAlert-message">{t.msg}</div>
		</div>
	</div>
{/if}
