<script lang="ts">
	/** 환경 설정 한 칸 — 제목(+ ⓘ 설명) · 내용. SettingGroup 안에서 쓴다(2026-10-04)
	 *  wrap 묶음에서 칸 폭: basis(기본 폭 px)·grow(남는 폭 나눠 갖는 비율) — 안 주면 묶음 기본(260px, 똑같이 나눔). 좁은 칸은 grow 0, 넓게 쓸 칸은 grow 2 처럼 */
	import type { Snippet } from 'svelte';
	import InfoTip from './InfoTip.svelte';
	let { label, tip = null, basis = 0, grow = -1, children }: { label: string; tip?: { title?: string; lines: string[] } | null; basis?: number; grow?: number; children: Snippet } = $props();
	const flex = $derived(basis || grow >= 0 ? `flex: ${grow >= 0 ? grow : 1} 1 ${basis || 260}px` : '');
</script>

<div class="srow" style={flex}>
	<InfoTip {label} {tip} />
	{@render children()}
</div>

<style>
	.srow {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
	}
</style>
