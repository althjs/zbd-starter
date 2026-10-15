<script lang="ts">
	/**
	 * 제목 + ⓘ 설명 말풍선(환경 설정 칸·카드 공용, 2026-10-04). PC 는 올리면, 휴대폰은 누르면 뜨고 바깥을 누르면 닫힌다.
	 *   <InfoTip label="파비콘" tip={{ title: '파비콘', lines: ['…', '…'] }} />   tip 이 없으면 제목만.
	 */
	let { label, tip = null }: { label: string; tip?: { title?: string; lines: string[] } | null } = $props();
	let open = $state(false);
	let wrap = $state<HTMLElement>();
	$effect(() => {
		if (!open) return;
		const off = (e: PointerEvent) => wrap && !wrap.contains(e.target as Node) && (open = false);
		window.addEventListener('pointerdown', off);
		return () => window.removeEventListener('pointerdown', off);
	});
</script>

<div class="bth">
	<p class="MuiTypography-root MuiTypography-subtitle2 bt">{label}</p>
	{#if tip}
		<span
			bind:this={wrap}
			class="tipw"
			role="button"
			tabindex="0"
			aria-label="{label} 설명"
			onpointerenter={(e) => e.pointerType != 'touch' && (open = true)}
			onpointerleave={(e) => e.pointerType != 'touch' && (open = false)}
			onclick={() => (open = !open)}
			onkeydown={(e) => (e.key == 'Enter' || e.key == ' ') && (open = !open)}
		>
			<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"
				><path fill="currentColor" d="M11 7h2v2h-2zm0 4h2v6h-2zm1-9C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2m0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8"></path></svg
			>
			{#if open}
				<span class="tip" role="tooltip">
					<b>{tip.title || label}</b>
					{#each tip.lines as l, i (i)}<span class="tl">{l}</span>{/each}
				</span>
			{/if}
		</span>
	{/if}
</div>

<style>
	.bth {
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.bt {
		margin: 0;
	}
	.tipw {
		position: relative;
		display: inline-flex;
		color: var(--gt-600, #a0a0a0);
		cursor: help;
		outline: none;
	}
	.tipw:hover,
	.tipw:focus-visible {
		color: var(--theme-color-main, #3b82f6);
	}
	.tip {
		position: absolute;
		top: calc(100% + 8px);
		left: -12px;
		z-index: 20;
		width: max-content;
		max-width: min(280px, 72vw);
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 14px;
		border-radius: 10px;
		background: #1f2937;
		color: #f3f4f6;
		font-size: 12.5px;
		line-height: 1.55;
		font-weight: 400;
		word-break: keep-all;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
		cursor: default;
		animation: tipin 120ms ease-out;
	}
	.tip::before {
		content: '';
		position: absolute;
		top: -5px;
		left: 15px;
		width: 10px;
		height: 10px;
		background: #1f2937;
		transform: rotate(45deg);
	}
	.tip b {
		font-size: 13px;
		color: #fff;
		margin-bottom: 2px;
	}
	.tl {
		position: relative;
		padding-left: 10px;
	}
	.tl::before {
		content: '·';
		position: absolute;
		left: 0;
		color: #9ca3af;
	}
	@keyframes tipin {
		from {
			opacity: 0;
			transform: translateY(-3px);
		}
	}
</style>
