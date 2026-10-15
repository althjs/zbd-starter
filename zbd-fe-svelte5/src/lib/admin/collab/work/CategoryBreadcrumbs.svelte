<script lang="ts">
	/**
	 * 분류 이동 경로(React CategoryBreadcrumbs) — 「대분류 > 중분류 > 소분류」, 누르면 그 분류 화면(또는 onselect).
	 * linkType="LIST" 면 링크는 목록(/admin/work/{보드})이고 누르면 onselect 로 그 분류를 고른다.
	 */
	import { page } from '$app/state';
	import type { Category, CodeItem } from '#lib/shared/types/work.ts';

	let {
		category,
		codes,
		groupKey,
		workId = 0,
		onselect = null,
		linkType = ''
	}: { category: Category | undefined; codes: CodeItem[]; groupKey: string; workId?: number; onselect?: ((c: Category) => void) | null; linkType?: '' | 'LIST' } = $props();
	const c = $derived(category || { c1: 0, c2: 0, c3: 0 });
	const code1 = $derived(c.c1 ? codes.find((x) => x.id === c.c1) : null);
	const code2 = $derived(code1 && c.c2 ? code1.items?.find((x) => x.id === c.c2) : null);
	const code3 = $derived(code2 && c.c3 ? code2.items?.find((x) => x.id === c.c3) : null);
	const q = $derived(workId ? `?workId=${workId}` : '');
	const items = $derived(
		[
			code1 && { label: code1.name, cat: { c1: c.c1, c2: 0, c3: 0 }, href: `/admin/work/${groupKey}/category/${c.c1}${q}` },
			code2 && { label: code2.name, cat: { c1: c.c1, c2: c.c2, c3: 0 }, href: `/admin/work/${groupKey}/category/${c.c1}/${c.c2}${q}` },
			code3 && { label: code3.name, cat: { ...c }, href: `/admin/work/${groupKey}/category/${c.c1}/${c.c2}/${c.c3}${q}` }
		].filter(Boolean) as { label: string; cat: Category; href: string }[]
	);
	const inDetail = $derived(!!(page.params as Record<string, string>).workId || !!(page.params as Record<string, string>).c1);
</script>

{#if !c.c1}
	<p class="MuiTypography-root MuiTypography-body2">분류 미선택</p>
{:else if !code1}
	<p class="MuiTypography-root MuiTypography-body2">미선택</p>
{:else}
	<div class="MuiStack-root" style="display:flex;flex-direction:row;align-items:center;flex-wrap:wrap">
		{#each items as it, i (i)}
			{#if i > 0}<span class="MuiTypography-root MuiTypography-caption" style="color:var(--adm-text2);margin-left:4px">&gt;</span>{/if}
			<a
				class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover"
				style={i > 0 ? 'margin-left:4px;white-space:nowrap' : 'white-space:nowrap'}
				href={linkType === 'LIST' ? `/admin/work/${groupKey}` : it.href}
				onclick={(e) => {
					if (onselect && (!inDetail || linkType === 'LIST')) {
						e.preventDefault();
						e.stopPropagation();
						onselect(it.cat);
					}
				}}>{it.label}</a
			>
		{/each}
	</div>
{/if}
