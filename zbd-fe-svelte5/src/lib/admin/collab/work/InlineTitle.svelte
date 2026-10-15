<script lang="ts">
	/** 누르면 고치는 한 줄 글(React TextInputInline hint) — 다른 곳을 누르거나 Enter 면 onsave, Esc 면 되돌림 */
	import { tick } from 'svelte';
	let { value = '', placeholder = '', onsave }: { value?: string; placeholder?: string; onsave: (v: string) => void } = $props();
	let edit = $state(false);
	let text = $state('');
	let input = $state<HTMLInputElement>();
	async function start() {
		text = value || '';
		edit = true;
		await tick();
		setTimeout(() => (input?.focus(), input?.select()), 100);
	}
	const esc = (s: string) => s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c] as string);
</script>

{#if !edit}
	<div class="MuiStack-root" style="display:flex;flex-direction:column;position:relative">
		<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
		<div class="MuiBox-root" style="width:100%;border-bottom:1px solid #eee;{value ? '' : 'color:#bbb'}" onclick={start}>{@html esc(String(value || placeholder)).replaceAll('\n', '<br/>')}</div>
	</div>
{:else}
	<div class="MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root">
		<div class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-sizeSmall Mui-focused" style="border-radius:0">
			<input
				bind:this={input}
				class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall"
				{placeholder}
				bind:value={text}
				onblur={() => ((edit = false), onsave(text))}
				onkeyup={(e) => {
					if (e.key == 'Enter') input?.blur();
					else if (e.key == 'Escape') ((text = value), (edit = false));
				}}
			/>
			<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline" style="border-radius:0"><legend><span class="notranslate" aria-hidden="true">​</span></legend></fieldset>
		</div>
	</div>
{/if}
