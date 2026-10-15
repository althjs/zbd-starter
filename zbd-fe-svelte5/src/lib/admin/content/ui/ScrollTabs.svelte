<script lang="ts">
	/**
	 * 넘치면 양 끝에 ‹ › 단추가 붙는 탭(MUI Tabs variant="scrollable" scrollButtons) — 바탕 ui/Tabs 는 단추가 없어 콘텐츠 묶음이 따로 둔다.
	 *   <ScrollTabs items={[{ value: '', label: '전체' }]} value={tab} onchange={(v) => …}>
	 *     {#snippet label(t)}<h6 …>{t.label}</h6>{/snippet}      탭 글자를 직접 그릴 때
	 *   </ScrollTabs>
	 * 단추는 휴대폰(<600) 에서 숨는다(scrollButtonsHideMobile). 넘치지 않으면 단추가 없다(auto).
	 */
	import type { Snippet } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	type T = { value: any; label: string; [k: string]: any };
	let { items, value, onchange, sx = null, label }: { items: T[]; value: any; onchange: (v: any) => void; sx?: Sx | null; label?: Snippet<[T]> } = $props();

	let scroller = $state<HTMLElement>();
	let bar = $state({ left: 0, width: 0 });
	let over = $state(false);
	let atStart = $state(true);
	let atEnd = $state(false);

	function measure() {
		const el = scroller;
		if (!el) return;
		over = el.scrollWidth > el.clientWidth + 1;
		atStart = el.scrollLeft <= 1;
		atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1;
		const sel = el.querySelector('.Mui-selected') as HTMLElement | null;
		bar = sel ? { left: sel.offsetLeft, width: sel.offsetWidth } : { left: 0, width: 0 };
	}
	/** MUI 처럼 고른 탭이 가려져 있으면 보이도록 밀어 둔다(왼쪽이면 왼쪽 끝, 오른쪽이면 오른쪽 끝에 맞춤) */
	let aligned: '' | 'left' | 'right' = '';
	function intoView() {
		const el = scroller;
		const sel = el?.querySelector('.Mui-selected') as HTMLElement | null;
		if (!el || !sel) return;
		const l = sel.offsetLeft;
		const r = l + sel.offsetWidth;
		// 글꼴이 늦게 들어와 폭이 바뀌면 처음 맞춘 쪽 끝에 다시 맞춘다(React 는 다 그린 뒤 한 번 맞춘 결과와 같게)
		if (aligned == 'left' || l < el.scrollLeft) ((el.scrollLeft = l), (aligned = 'left'));
		else if (aligned == 'right' || r > el.scrollLeft + el.clientWidth) ((el.scrollLeft = r - el.clientWidth), (aligned = 'right'));
	}
	$effect(() => {
		void value;
		aligned = '';
		void items.map((t) => t.label).join('|');
		requestAnimationFrame(() => (intoView(), measure()));
		// 탭 글자·아이콘 글꼴이 늦게 들어와 폭이 바뀌면 다시 맞춘다(사용자가 민 뒤에는 건드리지 않게 처음 2초만)
		const t0 = Date.now();
		const ro = new ResizeObserver(() => (Date.now() - t0 < 2000 && intoView(), measure()));
		if (scroller) {
			ro.observe(scroller);
			scroller.querySelectorAll('.MuiTab-root').forEach((b) => ro.observe(b));
		}
		return () => ro.disconnect();
	});
	function move(dir: number) {
		if (!scroller) return;
		scroller.scrollBy({ left: dir * scroller.clientWidth, behavior: 'smooth' });
	}
	const btn = toClass({ width: 40, flexShrink: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', '@media (max-width:599.95px)': { display: 'none' } });
	const hideSm = toClass({ '& .MuiTab-root': { flex: '0 0 auto' } });
</script>

{#snippet arrow(dir: number, disabled: boolean)}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		role="button"
		tabindex="-1"
		class={cx(
			'MuiButtonBase-root MuiTabScrollButton-root MuiTabScrollButton-horizontal MuiTabs-scrollButtons MuiTabs-scrollButtonsHideMobile',
			disabled && 'Mui-disabled',
			btn,
			toClass({ opacity: disabled ? 0.3 : 1 })
		)}
		onclick={() => !disabled && move(dir)}
	>
		<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeSmall" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
			><path d={dir < 0 ? 'M15.41 16.09l-4.58-4.59 4.58-4.59L14 5.5l-6 6 6 6z' : 'M8.59 16.34l4.58-4.59-4.58-4.59L10 5.75l6 6-6 6z'}></path></svg
		>
	</div>
{/snippet}

<div class={cx('MuiTabs-root', toClass(sx))}>
	{#if over}{@render arrow(-1, atStart)}{/if}
	<div class={cx('MuiTabs-scroller MuiTabs-hideScrollbar MuiTabs-scrollableX', hideSm)} bind:this={scroller} onscroll={measure}>
		<div role="tablist" class="MuiTabs-list MuiTabs-flexContainer">
			{#each items as t (t.value)}
				<button
					type="button"
					role="tab"
					aria-selected={value === t.value}
					class={cx('MuiButtonBase-root MuiTab-root MuiTab-textColorPrimary', value === t.value && 'Mui-selected')}
					onclick={() => onchange(t.value)}
					>{#if label}{@render label(t)}{:else}{t.label}{/if}</button
				>
			{/each}
		</div>
		<span class="MuiTabs-indicator" style="left:{bar.left}px;width:{bar.width}px"></span>
	</div>
	{#if over}{@render arrow(1, atEnd)}{/if}
</div>
