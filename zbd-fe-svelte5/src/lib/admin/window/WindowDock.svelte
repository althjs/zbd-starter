<script lang="ts">
	/**
	 * 창 목록 손잡이(React window-control/WindowDock) — 오른쪽 가운데, 열린 창 수. 올리면(휴대폰은 누르면) 목록:
	 * 창 누르기 = 복원 + 맨 앞, ✕ = 닫기, 아래 줄 = 최소화 줄 표시 · 모두 최소화 · 모두 닫기.
	 */
	import IconButton from '../ui/IconButton.svelte';
	import SvgIcon from '../ui/SvgIcon.svelte';
	import Tooltip from '../ui/Tooltip.svelte';
	import { confirm } from '../ui/modal.svelte.ts';
	import { closeAllWindows, focusWindow, minimizeAll, requestCloseWindow, toggleShowDock, topId, wm } from './windows.svelte.ts';

	let hover = $state(false);
	const allMin = $derived(wm.list.length > 0 && wm.list.every((w) => w.view === 'minimized'));
	const top = $derived((void wm.order, void wm.list.map((w) => w.view), topId()));

	async function closeAll() {
		if (await confirm('현재 열려있는 모든 윈도우를 닫으시겠습니까?')) closeAllWindows();
	}
</script>

{#if wm.list.length}
	<div class="zwd" role="region" aria-label="열린 창" onmouseenter={() => (hover = true)} onmouseleave={() => (hover = false)}>
		<div class="zwd-handle" role="button" tabindex="0" aria-label="열린 창 목록" onclick={() => (hover = !hover)} onkeydown={(e) => e.key == 'Enter' && (hover = !hover)}>
			<SvgIcon name="Apps" color="Primary" />
			<span class="zwd-count">{wm.list.length}</span>
		</div>
		{#if hover}
			<div class="zwd-panel MuiPaper-root MuiPaper-elevation8">
				<div class="zwd-list">
					{#each wm.list as w (w.id)}
						{@const isMin = w.view === 'minimized'}
						{@const isTop = w.id === top && !isMin}
						<div class="zwd-item" class:top={isTop} class:min={isMin} role="button" tabindex="0" onclick={() => focusWindow(w.id)} onkeydown={(e) => e.key == 'Enter' && focusWindow(w.id)}>
							<p class="MuiTypography-root MuiTypography-body2 zwd-name">{w.title || '제목 없음'}</p>
							<Tooltip title="닫기" placement="top" zIndex={2200}
								><IconButton size="Small" sx={{ marginLeft: 8, padding: 4 }} aria-label="닫기" onclick={(e: MouseEvent) => (e.stopPropagation(), requestCloseWindow(w.id))}
									><SvgIcon name="Close" size="Inherit" /></IconButton
								></Tooltip
							>
						</div>
					{/each}
				</div>
				<hr class="MuiDivider-root zwd-hr" />
				<div class="zwd-tools">
					<Tooltip title="최소화바 표시" placement="top" zIndex={2200}
						><span class="MuiSwitch-root MuiSwitch-sizeSmall"
							><span class="MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root" class:Mui-checked={wm.showDock}
								><input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" aria-label="최소화바 표시" checked={wm.showDock} onchange={toggleShowDock} /><span class="MuiSwitch-thumb"
								></span></span
							><span class="MuiSwitch-track"></span></span
						></Tooltip
					>
					<Tooltip title={allMin ? '모든 창이 최소화됨' : '모든 창 최소화'} placement="top" zIndex={2200}
						><span><IconButton size="Small" color="Primary" aria-label="모든 창 최소화" disabled={allMin} onclick={minimizeAll}><SvgIcon name="Remove" /></IconButton></span></Tooltip
					>
					<Tooltip title="모든 창 닫기" placement="top" zIndex={2200}
						><IconButton size="Small" aria-label="모든 창 닫기" sx={{ color: 'var(--adm-error-main)' }} onclick={closeAll}><SvgIcon name="DeleteSweep" /></IconButton></Tooltip
					>
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.zwd {
		position: fixed;
		right: 0;
		top: 50%;
		transform: translateY(-50%);
		z-index: 2100;
		display: flex;
		flex-direction: row-reverse;
		align-items: center;
	}
	.zwd-handle {
		box-sizing: border-box;
		width: 40px;
		height: 80px;
		background-color: color-mix(in srgb, var(--adm-paper, #fff) 80%, transparent);
		border-top-left-radius: 12px;
		border-bottom-left-radius: 12px;
		box-shadow:
			0px 2px 4px -1px rgba(0, 0, 0, 0.2),
			0px 4px 5px 0px rgba(0, 0, 0, 0.14),
			0px 1px 10px 0px rgba(0, 0, 0, 0.12);
		display: flex;
		flex-direction: column;
		gap: 4px;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		backdrop-filter: blur(8px);
		border: 1px solid var(--adm-divider);
		border-right: 0;
		transition: all 0.2s ease;
	}
	.zwd-handle:hover {
		width: 48px;
		background-color: color-mix(in srgb, var(--adm-paper, #fff) 95%, transparent);
	}
	.zwd-count {
		font-size: 10px;
		font-weight: bold;
		line-height: 1.66;
		font-family: var(--mui-font, Roboto, Helvetica, Arial, sans-serif);
	}
	.zwd-panel {
		margin-right: 8px;
		padding: 8px;
		max-height: 80vh;
		overflow-y: auto;
		background-color: color-mix(in srgb, var(--adm-paper, #fff) 90%, transparent);
		backdrop-filter: blur(10px);
		min-width: 200px;
		max-width: 300px;
		border-radius: 4px;
		box-shadow:
			0px 5px 5px -3px rgba(0, 0, 0, 0.2),
			0px 8px 10px 1px rgba(0, 0, 0, 0.14),
			0px 3px 14px 2px rgba(0, 0, 0, 0.12);
		color: var(--adm-text);
	}
	.zwd-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.zwd-item {
		padding: 8px;
		border-radius: 4px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		border: 1px solid transparent;
	}
	.zwd-item:hover {
		background-color: color-mix(in srgb, var(--adm-hover, rgba(0, 0, 0, 0.04)) 10%, transparent);
	}
	.zwd-item.top {
		background-color: var(--theme-color-alpha10, color-mix(in srgb, var(--theme-color-main) 10%, transparent));
		border-color: var(--theme-color-light);
	}
	.zwd-item.min {
		opacity: 0.6;
	}
	.zwd-name {
		margin: 0;
		flex-grow: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--adm-text);
	}
	.zwd-item.top .zwd-name {
		font-weight: 600;
		color: var(--theme-color-main);
	}
	.zwd-item.min .zwd-name {
		color: var(--adm-text2);
	}
	.zwd-hr {
		margin: 8px 0;
		border: 0;
		border-bottom: 1px solid var(--adm-divider);
	}
	.zwd-tools {
		display: flex;
		justify-content: space-evenly;
		align-items: center;
	}
</style>
