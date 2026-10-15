<script lang="ts">
	/** 메뉴 아이콘 — `200x200_…` 이면 업로드 이미지(admin_menu_icons), 아니면 머티리얼 심볼 이름 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { ADMIN_MENU_ICON_IMAGE_ID } from '../config.ts';
	import { adminImagePrefix } from '../meta.svelte.ts';
	import Icon from '../ui/Icon.svelte';
	let { icon = '', sx = null }: { icon?: string; sx?: Sx | null } = $props();
</script>

{#if icon.indexOf('200x200') == 0}
	<img class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img', toClass({ width: 25, height: 25 }))} src={`${adminImagePrefix()}/${ADMIN_MENU_ICON_IMAGE_ID}/${icon}`} alt="" />
{:else if icon}
	<Icon name={icon} {sx} />
{/if}
