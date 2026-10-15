<script lang="ts">
	/**
	 * 최소화한 창 줄(React GlobalWindowRenderer 의 #minimized-dock + ResizableWindowContainer 의 최소화 막대) — 화면 아래 오른쪽.
	 * 막대를 누르면 복원 + 맨 앞. 창 목록의 「최소화바 표시」 를 끄면 숨는다(창 목록에서는 여전히 고를 수 있다).
	 */
	import Icon from '../ui/Icon.svelte';
	import IconButton from '../ui/IconButton.svelte';
	import Tooltip from '../ui/Tooltip.svelte';
	import { focusWindow, wm } from './windows.svelte.ts';

	const mins = $derived(wm.list.filter((w) => w.view === 'minimized'));
</script>

<div id="minimized-dock" class="zmd">
	{#if wm.showDock}
		{#each mins as w (w.id)}
			<div class="zmd-bar" role="button" tabindex="0" aria-label="{w.title} 복원" onclick={() => focusWindow(w.id)} onkeydown={(e) => e.key == 'Enter' && focusWindow(w.id)}>
				<span class="zmd-name">{w.title}</span>
				<Tooltip placement="top" title="복원" zIndex={2200}
					><IconButton size="Small" aria-label="복원" onclick={(e: MouseEvent) => (e.stopPropagation(), focusWindow(w.id))}><Icon name="expand_less" size="Small" /></IconButton></Tooltip
				>
			</div>
		{/each}
	{/if}
</div>

<style>
	.zmd {
		position: fixed;
		bottom: 0;
		right: 0;
		left: 0;
		display: flex;
		justify-content: flex-end;
		align-items: flex-end;
		flex-wrap: wrap;
		padding: 8px;
		gap: 8px;
		z-index: 2000;
		pointer-events: none;
	}
	.zmd-bar {
		box-sizing: border-box;
		width: 250px;
		height: 48px;
		background-color: var(--zc-white, white);
		border-top-left-radius: 8px;
		border-top-right-radius: 8px;
		box-shadow: 0px -2px 10px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 16px;
		cursor: pointer;
		border: 1px solid var(--zc-ddd, #ddd);
		pointer-events: auto;
		color: var(--adm-text);
	}
	.zmd-name {
		font-weight: 600;
		font-size: 14px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		flex: 1;
		margin-right: 8px;
	}
</style>
