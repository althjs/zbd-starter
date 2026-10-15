<script lang="ts">
	import DOMPurify from 'dompurify';
	import { close, dialogs } from './dialog.svelte.ts';
	import Scrollbar from './Scrollbar.svelte';

	function onKey(e: KeyboardEvent) {
		const top = dialogs.list[dialogs.list.length - 1];
		if (top && e.key == 'Escape' && top.kind != 'alert') {
			close(top.id, false);
		}
	}
</script>

<svelte:window onkeydown={onKey} />

{#each dialogs.list as d (d.id)}
	<div class="backdrop" role="presentation" onclick={() => d.kind == 'modal' && close(d.id, false)}></div>
	<div class="dialog" class:wide={d.wide} role="dialog" aria-modal="true">
		{#if d.kind == 'modal'}
			<button class="x" aria-label="닫기" onclick={() => close(d.id, false)}>
				<span class="material-icons">close</span>
			</button>
		{/if}
		{#if d.title}<h2>{d.title}</h2>{/if}
		{#if d.html}
			<!-- 약관 등 긴 HTML — React 판처럼 Scrollbar(최대 화면 높이 - 150) -->
			<Scrollbar style="max-height:calc(100vh - 150px)"><div class="html">{@html DOMPurify.sanitize(d.html)}</div></Scrollbar>
		{:else if d.body}
			<p class="body">{d.body}</p>
		{/if}
		{#if d.kind != 'modal'}
			<div class="actions">
				{#if d.kind == 'confirm'}
					<button class="btn text" onclick={() => close(d.id, false)}>취소</button>
				{/if}
				<button class="btn primary" onclick={() => close(d.id, true)}>확인</button>
			</div>
		{/if}
	</div>
{/each}

<div class="toasts">
	{#each dialogs.toasts as t (t.id)}
		<div class="toast">{t.msg}</div>
	{/each}
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		background: var(--zbd-backdrop);
		z-index: var(--zbd-z-modal);
	}
	.dialog {
		font-family: var(--zbd-font-ui);
		position: fixed;
		z-index: 1301;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		width: calc(100% - 64px);
		max-width: 444px;
		max-height: calc(100% - 64px);
		overflow: auto;
		background: var(--zbd-paper);
		border-radius: 4px;
		box-shadow: var(--zbd-shadow-dialog);
		padding: 20px 24px 8px;
	}
	.dialog.wide {
		max-width: 800px;
		width: calc(100% - 10%);
	}
	h2 {
		font-size: 24px;
		font-weight: 800;
		padding-right: 28px;
		margin-bottom: 12px;
	}
	.body {
		font-size: 15px;
		line-height: 1.6;
		white-space: pre-wrap;
		color: #333;
		padding: 8px 0 12px;
	}
	.html {
		font-size: 14px;
		padding-bottom: 16px;
	}
	.x {
		position: absolute;
		right: 12px;
		top: 12px;
		border: 0;
		background: none;
		cursor: pointer;
		padding: 0;
		color: var(--zbd-text);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding: 8px 0;
	}
	.btn {
		border: 0;
		border-radius: 4px;
		padding: 6px 16px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		background: none;
	}
	.btn.primary {
		color: var(--zbd-primary);
	}
	.btn.text {
		color: var(--zbd-text-muted);
	}
	.toasts {
		position: fixed;
		left: 50%;
		bottom: 24px;
		transform: translateX(-50%);
		z-index: var(--zbd-z-toast);
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: center;
		pointer-events: none;
	}
	.toast {
		background: #323232;
		color: #fff;
		font-size: 14px;
		padding: 10px 16px;
		border-radius: 4px;
		box-shadow: 0 3px 5px -1px rgba(0, 0, 0, 0.2);
		max-width: calc(100vw - 32px);
	}
</style>
