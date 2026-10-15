<script lang="ts">
	/**
	 * MUI Pagination(outlined · rounded) 호환 — 쪽 번호 계산은 MUI usePagination 과 같다(boundaryCount 1).
	 */
	import { cx } from './sx';

	let {
		page,
		count,
		siblingCount = 1,
		variant = 'outlined',
		shape = 'rounded',
		onchange
	}: { page: number; count: number; siblingCount?: number; variant?: 'outlined' | 'text'; shape?: 'rounded' | 'circular'; onchange: (p: number) => void } = $props();

	const range = (a: number, b: number) => (b < a ? [] : Array.from({ length: b - a + 1 }, (_, i) => a + i));
	const items = $derived.by(() => {
		const B = 1;
		const startPages = range(1, Math.min(B, count));
		const endPages = range(Math.max(count - B + 1, B + 1), count);
		const sStart = Math.max(Math.min(page - siblingCount, count - B - siblingCount * 2 - 1), B + 2);
		const sEnd = Math.min(Math.max(page + siblingCount, B + siblingCount * 2 + 2), endPages.length > 0 ? endPages[0] - 2 : count - 1);
		return [
			'previous',
			...startPages,
			...(sStart > B + 2 ? ['start-ellipsis'] : B + 1 < count - B ? [B + 1] : []),
			...range(sStart, sEnd),
			...(sEnd < count - B - 1 ? ['end-ellipsis'] : count - B > B ? [count - B] : []),
			...endPages,
			'next'
		] as (number | string)[];
	});
	const base = $derived(`MuiPaginationItem-root MuiPaginationItem-sizeMedium MuiPaginationItem-${variant} MuiPaginationItem-${shape}`);
</script>

<nav aria-label="pagination navigation" class="MuiPagination-root MuiPagination-{variant}">
	<ul class="MuiPagination-ul">
		{#each items as it, i (i)}
			<li>
				{#if it == 'start-ellipsis' || it == 'end-ellipsis'}
					<div class="MuiPaginationItem-root MuiPaginationItem-sizeMedium MuiPaginationItem-ellipsis">…</div>
				{:else if it == 'previous' || it == 'next'}
					{@const disabled = it == 'previous' ? page <= 1 : page >= count}
					<button
						type="button"
						class={cx('MuiButtonBase-root', base, 'MuiPaginationItem-previousNext', disabled && 'Mui-disabled')}
						aria-label={it == 'previous' ? 'Go to previous page' : 'Go to next page'}
						tabindex={disabled ? -1 : 0}
						{disabled}
						onclick={() => onchange(it == 'previous' ? page - 1 : page + 1)}
					>
						<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiPaginationItem-icon" focusable="false" aria-hidden="true" viewBox="0 0 24 24"
							>{#if it == 'previous'}<path d="M15.41 7.41 14 6l-6 6 6 6 1.41-1.41L10.83 12z"></path>{:else}<path d="M10 6 8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"></path>{/if}</svg
						>
					</button>
				{:else}
					<button
						type="button"
						class={cx('MuiButtonBase-root', base, 'MuiPaginationItem-page', it == page && 'Mui-selected')}
						aria-current={it == page ? 'page' : undefined}
						aria-label={it == page ? `page ${it}` : `Go to page ${it}`}
						onclick={() => it != page && onchange(it as number)}>{it}</button
					>
				{/if}
			</li>
		{/each}
	</ul>
</nav>
