<script lang="ts">
	/**
	 * 목록 보기 옵션(톱니) — 열 보이기·고정·순서. 이 브라우저에만 저장, 「기본값」 은 호스팅 기본(TABLE_COLUMN_OPTIONS)으로.
	 * ⚠ React 판의 글자 크기·열 너비 끌기·호스팅 기본값 저장(공통코드)은 아직 없다.
	 */
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import Check from '../ui/Check.svelte';
	import { saveColumnOptions, type ColumnOption } from './columnOptions.ts';
	import { media } from '../../media.svelte.ts';

	let { optionKey, options, defaultOptions, onchange }: { optionKey: string; options: ColumnOption[]; defaultOptions: ColumnOption[]; onchange: (v: ColumnOption[]) => void } = $props();
	let open = $state(false);
	let list = $state<ColumnOption[]>([]);
	function show() {
		// 저장된 순서 먼저, 없는 열은 뒤에
		const saved = options || [];
		const out = saved.filter((s) => defaultOptions.find((d) => d.key == s.key)).map((s) => ({ ...defaultOptions.find((d) => d.key == s.key)!, ...s }));
		// 저장 뒤 새로 생긴 열은 기본 순서의 자리(앞 열 바로 뒤)에 — 표(ListTable rank)와 같은 순서
		defaultOptions.forEach((d, i) => {
			if (out.find((x) => x.key == d.key)) return;
			let at = 0;
			for (let j = i - 1; j >= 0; j--) {
				const p = out.findIndex((x) => x.key == defaultOptions[j].key);
				if (p !== -1) {
					at = p + 1;
					break;
				}
			}
			out.splice(at, 0, d);
		});
		list = out.map((x) => ({ ...x }));
		open = true;
	}
	function move(i: number, d: number) {
		const j = i + d;
		if (j < 0 || j >= list.length) return;
		const n = [...list];
		[n[i], n[j]] = [n[j], n[i]];
		list = n;
	}
	function apply() {
		saveColumnOptions(optionKey, list);
		onchange(list);
		open = false;
	}
</script>

<IconButton size="Small" sx={{ padding: 0, marginLeft: media.mdDown ? 0 : 4 }} onclick={show} aria-label="목록 보기 옵션"
	><Icon name="settings" size="Small" sx={{ fontSize: 15, color: 'var(--gt-700, #909090)' }} /></IconButton
>
<Dialog {open} onclose={() => (open = false)} title="목록 보기 옵션" maxWidth="xs" fullWidth mobileFullScreen>
	{#if media.mdDown}<p class="MuiTypography-root MuiTypography-body2" style="margin:0 0 8px;color:var(--adm-text2);font-size:13px">휴대폰에서는 「고정」 이 적용되지 않아요(PC 에서만).</p>{/if}
	{#each list as c, i (c.key)}
		<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;gap:4px;border-bottom:1px solid var(--adm-divider)">
			<span class="MuiTypography-root MuiTypography-body2" style="flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{c.name}</span>
			<Check label="보이기" labelSx={{ whiteSpace: 'nowrap' }} checked={c.use !== false} onchange={(v) => (list[i] = { ...c, use: v })} />
			<Check label="고정" labelSx={{ whiteSpace: 'nowrap' }} checked={!!c.sticky} onchange={(v) => (list[i] = { ...c, sticky: v })} />
			<IconButton size="Small" disabled={i == 0} onclick={() => move(i, -1)}><Icon name="arrow_upward" size="Small" /></IconButton>
			<IconButton size="Small" disabled={i == list.length - 1} onclick={() => move(i, 1)}><Icon name="arrow_downward" size="Small" /></IconButton>
		</div>
	{/each}
	{#snippet actions()}
		<Button
			color="inherit"
			onclick={() => {
				saveColumnOptions(optionKey, null);
				onchange(defaultOptions);
				open = false;
			}}>기본값</Button
		>
		<Button variant="contained" onclick={apply}>적용</Button>
	{/snippet}
</Dialog>
