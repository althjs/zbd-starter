<script lang="ts">
	/**
	 * 탭(MUI Tabs 모양 — 글자 14·500, 고른 탭 아래 대표색 막대). 넘치면 가로로 밀린다.
	 *   <Tabs items={[{ value: 'ALL', label: '전체' }, { value: 'NEW', label: '신규', count: 3 }]} bind:value={tab} onchange={(v) => load(v)} />
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		items,
		value = $bindable(),
		onchange,
		sx = null
	}: { items: { value: any; label: string; count?: number; disabled?: boolean }[]; value?: any; onchange?: (v: any) => void; sx?: Sx | null } = $props();
	let box = $state<HTMLElement>();
	let bar = $state({ left: 0, width: 0 });
	$effect(() => {
		void value;
		void items.length;
		requestAnimationFrame(() => {
			const el = box?.querySelector('.Mui-selected') as HTMLElement | null;
			bar = el ? { left: el.offsetLeft, width: el.offsetWidth } : { left: 0, width: 0 };
		});
	});
</script>

<div class={cx('MuiTabs-root', toClass(sx))}>
	<div class="MuiTabs-scroller MuiTabs-scrollableX" bind:this={box}>
		<div class="MuiTabs-flexContainer" role="tablist">
			{#each items as t (t.value)}
				<button
					type="button"
					role="tab"
					class={cx('MuiButtonBase-root MuiTab-root MuiTab-textColorPrimary', value === t.value && 'Mui-selected', t.disabled && 'Mui-disabled')}
					aria-selected={value === t.value}
					disabled={t.disabled}
					onclick={() => {
						value = t.value;
						onchange?.(t.value);
					}}
					>{#if t.count != null}<span class="adm-tab-label">{t.label}<span class="adm-tab-count">{t.count}</span></span>{:else}{t.label}{/if}</button
				>
			{/each}
		</div>
		<span class="MuiTabs-indicator" style="left:{bar.left}px;width:{bar.width}px"></span>
	</div>
</div>
