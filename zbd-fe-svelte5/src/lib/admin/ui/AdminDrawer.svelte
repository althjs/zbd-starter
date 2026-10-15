<script lang="ts">
	/**
	 * 오른쪽 편집 서랍 — 공용 shared/mui/Drawer 를 감싸 제목 줄(닫기 ✕) · 본문 · 아래 버튼 줄을 붙인다.
	 * 폭은 관리자 규칙(admin.css): 900 미만 꽉 차게, 이상은 최소 min(560px, 100vw) — width 로 더 넓힐 수 있다(예 '50%').
	 *   <AdminDrawer open={o} title="회원 정보" onclose={() => (o = false)} width="50%">
	 *     본문
	 *     {#snippet actions()}<Button variant="contained" onclick={save}>저장</Button>{/snippet}
	 *   </AdminDrawer>
	 * onclose 가 false 를 돌려주면(「저장 안 한 내용을 버릴까요?」 취소) 닫지 않는 것은 부르는 쪽이 open 을 그대로 두면 된다.
	 * 휴대폰 뒤로가기로도 닫힌다(installAutoBackToClose 가 Esc 를 보낸다).
	 */
	import type { Snippet } from 'svelte';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Icon from './Icon.svelte';
	import IconButton from './IconButton.svelte';

	let {
		open,
		onclose,
		title = '',
		width = '',
		paperSx = null,
		children,
		actions,
		head
	}: { open: boolean; onclose: () => void; title?: string; width?: string | number; paperSx?: Sx | null; children?: Snippet; actions?: Snippet; head?: Snippet } = $props();
</script>

<Drawer {open} anchor="right" {onclose} paperSx={{ ...(width ? { width } : {}), ...(paperSx || {}) }}>
	<div
		class={cx(
			'MuiStack-root',
			toClass({
				flexDirection: 'row',
				alignItems: 'center',
				justifyContent: 'space-between',
				padding: '12px 16px 12px 24px',
				borderBottom: '1px solid var(--adm-divider)',
				position: 'sticky',
				top: 0,
				zIndex: 2,
				backgroundColor: 'var(--adm-paper)'
			})
		)}
	>
		{#if head}{@render head()}{:else}<h6 class="MuiTypography-root MuiTypography-h6">{title}</h6>{/if}
		<IconButton onclick={onclose} aria-label="닫기"><Icon name="close" /></IconButton>
	</div>
	<div class={cx('MuiBox-root', toClass({ flex: '1 1 auto', padding: 24 }))}>{@render children?.()}</div>
	{#if actions}
		<div
			class={cx(
				'MuiStack-root',
				toClass({
					flexDirection: 'row',
					justifyContent: 'flex-end',
					gap: 8,
					padding: '12px 24px',
					borderTop: '1px solid var(--adm-divider)',
					position: 'sticky',
					bottom: 0,
					backgroundColor: 'var(--adm-paper)'
				})
			)}
		>
			{@render actions()}
		</div>
	{/if}
</Drawer>
