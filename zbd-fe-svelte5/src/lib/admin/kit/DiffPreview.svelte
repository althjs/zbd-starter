<script lang="ts">
	/** 바뀐 내용(jsondiffpatch html) + 「같은 데이터 보기/숨기기」 — React DiffPreview. 개발자가 아니면 운영에선 안 보인다(React 와 같게) */
	import Button from '#lib/admin/ui/Button.svelte';
	import SvgIcon from './Svg.svelte';
	import { admin } from '#lib/admin/profile.svelte.ts';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { loadDiffLib } from './diff.ts';

	let { diffHTML = '', sx = null }: { diffHTML?: string; sx?: Sx | null } = $props();
	let hide = $state(true);
	const visible = $derived(!!diffHTML && (!import.meta.env.PROD || !!admin.profile?.isDeveloper));
	async function toggle() {
		const lib = await loadDiffLib();
		hide = !hide;
		if (lib) hide ? lib.formatters.html.hideUnchanged() : lib.formatters.html.showUnchanged();
	}
</script>

{#if visible}
	<div class={cx('MuiGrid-root', toClass({ marginLeft: -10, ...(sx || {}) }))}>{@html diffHTML}</div>
	<Button variant="outlined" size="small" sx={{ width: 'fit-content' }} onclick={toggle}>
		<SvgIcon name={hide ? 'Visibility' : 'VisibilityOff'} size="Small" sx={{ marginRight: 8 }} />같은 데이터 {hide ? '보기' : '숨기기'}
	</Button>
{/if}
