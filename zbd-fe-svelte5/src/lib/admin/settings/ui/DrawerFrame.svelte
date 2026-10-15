<script lang="ts">
	/**
	 * 오른쪽 편집 서랍의 틀 — 위 머리(제목·닫기, 붙어 있음) · 가운데 내용 · 아래 버튼 줄(붙어 있음). 관리자·회원 상세 서랍이 같이 쓴다.
	 *   <Drawer …><DrawerFrame title="관리자 수정" onclose={close}>{#snippet footer()}…{/snippet} … </DrawerFrame></Drawer>
	 * 안의 구역은 클래스로: .dd-card(묶음 카드) · .dd-card-title · .dd-grid(2칸, 좁으면 1칸) · .dd-help · .dd-danger(위험 구역). 칸은 관리자 입력칸(ui/TextField·SelectField — 라벨이 칸 안)
	 */
	import type { Snippet } from 'svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';

	let { title, onclose, footer, children }: { title: string; onclose: () => void; footer?: Snippet; children?: Snippet } = $props();
</script>

<div class="dd-frame">
	<div class="dd-head">
		<h2 class="dd-title">{title}</h2>
		<IconButton aria-label="닫기" onclick={onclose}><Icon name="close" /></IconButton>
	</div>
	<div class="dd-body">{@render children?.()}</div>
	{#if footer}<div class="dd-foot">{@render footer()}</div>{/if}
</div>

<style>
	.dd-frame {
		display: flex;
		flex-direction: column;
		/* 오른쪽 서랍 = 화면 높이 — 내용이 짧아도 버튼 줄이 맨 아래(Scrollbar 안이라 100% 는 안 먹는다) */
		min-height: 100vh;
		min-height: 100dvh;
		background: var(--zc-f7f8fa, #f7f8fa);
	}
	.dd-head {
		position: sticky;
		top: 0;
		z-index: 2;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 12px 12px 12px 24px;
		background: var(--zc-fff, #fff);
		border-bottom: 1px solid var(--gt-200, #f2f2f2);
	}
	.dd-title {
		margin: 0;
		font-size: 18px;
		font-weight: 700;
		line-height: 28px;
	}
	.dd-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 20px 24px 24px;
	}
	.dd-foot {
		position: sticky;
		bottom: 0;
		z-index: 2;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 24px;
		background: var(--zc-fff, #fff);
		border-top: 1px solid var(--gt-200, #f2f2f2);
	}
	@media (max-width: 599.95px) {
		.dd-head {
			padding-left: 16px;
		}
		.dd-body {
			padding: 16px;
		}
		.dd-foot {
			padding: 12px 16px;
		}
	}
	.dd-body :global(.dd-card) {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 20px;
		background: var(--zc-fff, #fff);
		border: 1px solid var(--adm-divider, #eaeaea);
		border-radius: 14px;
	}
	.dd-body :global(.dd-card.dd-danger) {
		border-color: rgba(229, 57, 53, 0.35);
	}
	.dd-body :global(.dd-card-title) {
		margin: 0;
		font-size: 15px;
		font-weight: 700;
		line-height: 22px;
	}
	.dd-body :global(.dd-danger .dd-card-title) {
		color: #e53935;
	}
	.dd-body :global(.dd-grid) {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 16px;
	}
	.dd-body :global(.dd-help) {
		font-size: 12px;
		line-height: 17px;
		color: var(--gt-500, #999);
	}
</style>
