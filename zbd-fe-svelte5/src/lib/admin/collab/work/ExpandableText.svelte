<script lang="ts">
	/** 설명 열 — 3줄까지, 넘치면 「더보기/숨기기」. 현금영수증 문의는 증빙용도·번호를 위에(React ExpandableText) */
	import { tick } from 'svelte';
	import { toDescHtml } from './codes.ts';
	let { text = '', maxLines = 3, cashPurpose = '', cashNumber = '' }: { text?: string; maxLines?: number; cashPurpose?: string; cashNumber?: string } = $props();
	let expanded = $state(false);
	let over = $state(false);
	let el = $state<HTMLElement>();
	const html = $derived(toDescHtml(text, '<br/>'));
	$effect(() => {
		void html;
		tick().then(() => el && (over = el.scrollHeight > el.clientHeight));
	});
</script>

{#if html || cashPurpose || cashNumber}
	<div class="MuiBox-root">
		{#if cashPurpose || cashNumber}
			<div class="MuiStack-root" style="display:flex;flex-direction:row;gap:6px;align-items:center;line-height:1.2;margin-top:8px">
				<span class="MuiTypography-root MuiTypography-caption" style="font-weight:700;color:var(--gt-700, #909090);line-height:1.2"
					>{cashPurpose === 'BUSINESS' ? '사업자지출증빙' : '개인소득공제'}</span
				>
				{#if cashNumber}<span class="MuiTypography-root MuiTypography-caption" style="font-weight:600;line-height:1.2">{cashNumber}</span>{/if}
			</div>
		{/if}
		{#if html}
			<div
				bind:this={el}
				class="MuiTypography-root MuiTypography-body2 adm-cv-desc"
				style="white-space:pre-wrap;display:-webkit-box;-webkit-line-clamp:{expanded ? 'unset' : maxLines};-webkit-box-orient:vertical;overflow:hidden;text-overflow:ellipsis;font-size:12px"
			>
				<div>{@html html}</div>
			</div>
		{/if}
		{#if over}
			<button
				type="button"
				class="MuiButtonBase-root MuiButton-root MuiButton-text MuiButton-textPrimary MuiButton-sizeSmall MuiButton-textSizeSmall"
				style="padding:0;margin-top:4px;min-width:auto;font-size:11px"
				onclick={() => (expanded = !expanded)}>{expanded ? '숨기기' : '더보기'}</button
			>
		{/if}
	</div>
{/if}
