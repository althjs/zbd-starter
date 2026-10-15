<script lang="ts">
	/** MUI Divider(vertical·middle·flexItem) 호환 — 메뉴 사이 세로줄 */
	import { cx, sx as toClass, type Sx } from './sx';

	let { vertical = true, sx = null }: { vertical?: boolean; sx?: Sx | null } = $props();
</script>

<div class={cx('MuiDivider-root MuiDivider-middle', vertical && 'MuiDivider-vertical MuiDivider-flexItem', toClass(sx))} role="separator" aria-orientation={vertical ? 'vertical' : 'horizontal'}></div>
