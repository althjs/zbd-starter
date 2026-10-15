<script lang="ts">
	/**
	 * 환경 설정 묶음 — 제목 · 한 줄 설명 · 칸들(SettingRow). wrap 이면 PC 에서 칸을 가로로 놓고 좁으면 줄바꿈(휴대폰은 한 칸씩) — 2026-10-04
	 * 묶음 하나 = 관리자 카드 하나(2026-10-04 대표 — 사용자 회원 관리처럼 묶음마다 맨 위 카드. 다른 카드 안에 넣으면 mui.css 규칙이 테두리 카드로 맞춘다).
	 * 머리를 눌러 접고 편다(기본 — fold={false} 로 끔). 접힘은 브라우저 저장소(SK.admin.settingGroupFolded)에 남는다 — 키는 foldKey, 없으면 「주소|제목」.
	 * defaultFolded = 처음(저장값 없을 때) 접힌 채로 — 긴 목록용.
	 */
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import { SK } from '#lib/shared/storage-keys.ts';
	let {
		title,
		desc = '',
		wrap = false,
		fold = true,
		foldKey = '',
		defaultFolded = false,
		children
	}: { title: string; desc?: string; wrap?: boolean; fold?: boolean; foldKey?: string; defaultFolded?: boolean; children: Snippet } = $props();
	const key = $derived(fold ? foldKey || `${page.url.pathname}|${title}` : '');

	function readFolded(): Record<string, boolean> {
		try {
			return JSON.parse(localStorage.getItem(SK.admin.settingGroupFolded) || '{}') || {};
		} catch {
			return {};
		}
	}
	let folded = $state(false);
	// 처음 그리기 전에 저장된 접힘을 읽는다(foldKey 가 바뀌면 그 키의 값으로)
	$effect.pre(() => {
		folded = key ? (readFolded()[key] ?? defaultFolded) : false;
	});
	function toggle() {
		folded = !folded;
		try {
			const m = readFolded();
			if (folded === defaultFolded) delete m[key];
			else m[key] = folded;
			localStorage.setItem(SK.admin.settingGroupFolded, JSON.stringify(m));
		} catch {}
	}
</script>

<section class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root sgrp" class:folded>
	{#if key}
		<button type="button" class="sgh fold" aria-expanded={!folded} onclick={toggle}>
			<span class="tx">
				<h6 class="MuiTypography-root MuiTypography-subtitle1">{title}</h6>
				{#if desc}<p class="MuiTypography-root MuiTypography-caption">{desc}</p>{/if}
			</span>
			<svg class="MuiSvgIcon-root chev" class:up={!folded} viewBox="0 0 24 24" aria-hidden="true"><path d="M7 10l5 5 5-5z"></path></svg>
		</button>
	{:else}
		<div class="sgh">
			<h6 class="MuiTypography-root MuiTypography-subtitle1">{title}</h6>
			{#if desc}<p class="MuiTypography-root MuiTypography-caption">{desc}</p>{/if}
		</div>
	{/if}
	{#if !folded}<div class="sgb" class:wrap>{@render children()}</div>{/if}
</section>

<style>
	/* 겉모양(바탕·모서리·그림자, 카드 안이면 테두리)은 MuiCard-root 가 — 여기서 주지 않는다 */
	.sgrp {
		flex-shrink: 0;
	}
	.sgh {
		padding: 16px 20px 12px;
		border-bottom: 1px solid var(--adm-divider, #eaecf0);
	}
	.folded .sgh {
		border-bottom: 0;
	}
	.sgh.fold {
		all: unset;
		box-sizing: border-box;
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 20px 12px;
		border-bottom: 1px solid var(--adm-divider, #eaecf0);
		cursor: pointer;
	}
	.folded .sgh.fold {
		border-bottom: 0;
	}
	.sgh.fold:focus-visible {
		outline: 2px solid var(--theme-color-main);
		outline-offset: -2px;
	}
	.tx {
		flex: 1;
		min-width: 0;
	}
	.chev {
		width: 24px;
		height: 24px;
		fill: currentColor;
		color: var(--gt-700, #909090);
		transition: transform 150ms;
	}
	.chev.up {
		transform: rotate(180deg);
	}
	.sgh h6 {
		margin: 0;
		font-weight: 700;
	}
	.sgh p {
		margin: 2px 0 0;
		color: var(--gt-700, #909090);
	}
	.sgb {
		display: flex;
		flex-direction: column;
		gap: 22px;
		padding: 18px 20px 20px;
	}
	.sgb.wrap {
		flex-direction: row;
		flex-wrap: wrap;
		gap: 16px 20px;
	}
	.sgb.wrap > :global(*) {
		flex: 1 1 260px;
	}
	@media (max-width: 599px) {
		.sgh,
		.sgh.fold {
			padding: 14px 14px 10px;
		}
		.sgb {
			padding: 14px;
		}
	}
</style>
