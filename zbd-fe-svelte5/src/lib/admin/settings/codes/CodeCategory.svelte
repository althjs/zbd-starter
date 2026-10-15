<script lang="ts">
	/**
	 * 공통 코드 분류 관리(React components/common/codes/CodeCategory) — 1단계 표. 그룹 화면(groupKey)은 그 그룹의 하위 단계 표를 옆에 늘어놓는다.
	 *   스타터는 「펼침」(표에서 하위 단계를 골라 여는 것) 없음.
	 *   groupKey 를 주면 그 그룹(예 SEO) 의 하위 코드만(/admin/common-codes/[groupKey]).
	 * 저장은 그 1단계 코드째(doSaveMeta) — 깊은 단계는 1단계 코드를 복사해 그 자리를 바꿔 넣는다.
	 */
	import { onMount, tick, untrack } from 'svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { media } from '../../media.svelte.ts';
	import Icon from '../../ui/Icon.svelte';
	import { toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import CodeCategoryListTable from './CodeCategoryListTable.svelte';
	import { doSaveMeta, doSortMeta, maxDepthOf } from './codes.ts';

	let { meta, hostingId = '', title, onChange, groupKey = '' }: { meta: any; hostingId?: string; title: string; onChange: (m: any) => void; groupKey?: string } = $props();

	let spreadIds = $state<any[]>([]);
	let componentName = $state('');
	const codeItems = $derived<any[]>(meta?.json?.items || []);
	const isDev = $derived(!!admin.profile?.isDeveloper);

	$effect(() => {
		if (!groupKey) return;
		const ids: any[] = [];
		for (const o of untrack(() => meta).json.items) {
			if (o.group === groupKey) {
				ids.push(o.id);
				componentName = o.component || '';
			}
		}
		spreadIds = ids;
	});

	async function doSave(item: any) {
		item.updatedAt = new Date();
		const r = await doSaveMeta(hostingId, meta, item);
		if (!r) return item;
		onChange(r.m);
		toast('정상적으로 저장되었습니다.');
		return r.o;
	}
	async function doSortUpdate(os: any[]) {
		const m = await doSortMeta(hostingId, meta, os);
		if (!m) return;
		onChange(m);
		toast('정상적으로 저장되었습니다.');
	}

	function getSpreadItem(i: number) {
		const base = codeItems.find((x) => x.id == spreadIds[0]);
		if (!base) return { parent: { items: [] }, titles: [] as string[], parentIds: [] as any[] };
		const titles = [base.name];
		const parentIds = [base.id];
		let cur = base;
		for (let d = 1; d <= i; d++) {
			const next = (cur.items || []).find((x: any) => x.id == spreadIds[d]) || { items: [] };
			titles.push(next.name);
			parentIds.push(next.id);
			cur = next;
		}
		return { parent: cur, titles, parentIds };
	}

	/** depth 단계 표의 한 줄 저장 → 1단계 코드 사본의 그 자리에 넣어 저장 */
	function saveDeep(depth: number, parentIds: any[], o: any) {
		const base = clone(getSpreadItem(0).parent);
		const walk = (p: any, now: number) => {
			if (now > base.maxDepth) return;
			if (depth == now && !p.items?.find((x: any) => x.id == o.id)) {
				(p.items ||= []).push({ ...o });
				return;
			}
			for (const oo of p.items) {
				if (depth == now && o.id == oo.id) {
					p.items = p.items.map((x: any) => (x.id == o.id ? { ...o } : x));
					return;
				}
				if (base.maxDepth > now && !oo.items) oo.items = [];
				if (oo.id != parentIds[now - 1] || !oo.items) continue;
				walk(oo, now + 1);
			}
		};
		walk(base, 2);
		return doSave(clone(base)).then(() => o);
	}
	function sortDeep(depth: number, parentIds: any[], os: any[]) {
		const base = clone(getSpreadItem(0).parent);
		const walk = (p: any, now: number) => {
			if (now > base.maxDepth) return;
			if (depth == now) {
				p.items = os;
				return;
			}
			for (const oo of p.items) if (oo.id == parentIds[now - 1] && oo.items) walk(oo, now + 1);
		};
		walk(base, 2);
		return doSave(base);
	}

	// 옆으로 넘기기(React Swiper navigation) — 넘칠 때만 화살표
	let scroller = $state<HTMLElement>();
	let canPrev = $state(false);
	let canNext = $state(false);
	const syncNav = () => {
		if (!scroller) return;
		canPrev = scroller.scrollLeft > 1;
		canNext = scroller.scrollLeft + scroller.clientWidth < scroller.scrollWidth - 1;
	};
	$effect(() => {
		spreadIds;
		setTimeout(syncNav);
	});
	onMount(() => {
		const ro = new ResizeObserver(syncNav);
		if (scroller) ro.observe(scroller);
		return () => ro.disconnect();
	});
	const pad = $derived(media.mdDown ? 0 : 32);
	const slideBy = (dir: number) => {
		if (!scroller) return;
		const slides = [...scroller.children] as HTMLElement[];
		const x = scroller.scrollLeft;
		const t = dir > 0 ? slides.find((s) => s.offsetLeft - pad > x + 1) : [...slides].reverse().find((s) => s.offsetLeft - pad < x - 1);
		scroller.scrollTo({ left: t ? t.offsetLeft - pad : dir > 0 ? scroller.scrollWidth : 0, behavior: 'smooth' });
	};
	const navBtn = toClass({
		position: 'absolute',
		top: '50%',
		zIndex: 10,
		width: 27,
		height: 44,
		marginTop: -22,
		display: 'flex',
		alignItems: 'center',
		justifyContent: 'center',
		cursor: 'pointer',
		color: 'var(--theme-color-main)',
		background: 'none',
		border: 0,
		padding: 0
	});
	// 휴대폰 — 단계마다 화면 폭 하나(옆으로 밀어 넘김), PC — 여러 단계를 나란히
	const slideW = $derived(media.mdDown ? { width: '100%', minWidth: 0 } : spreadIds.length > 0 ? { minWidth: isDev ? 900 : 750, width: '30%' } : { width: '100%' });
	const subW = $derived(
		media.mdDown ? { width: '100%', minWidth: 0 } : groupKey && spreadIds.length == 1 ? { width: '100%' } : { minWidth: isDev ? 700 : 600, width: spreadIds.length == 2 ? '50%' : '30%' }
	);
</script>

<div class={cx('MuiStack-root', toClass({ display: 'block', userSelect: 'none' }))}>
	<div class={cx('swiper', toClass({ position: 'relative', margin: media.mdDown ? '0 -16px' : '-16px -32px' }))}>
		<div
			bind:this={scroller}
			onscroll={syncNav}
			class={cx(
				'swiper-wrapper',
				toClass({
					display: 'flex',
					overflowX: 'auto',
					padding: media.mdDown ? 0 : '16px 32px',
					...(media.mdDown ? { scrollSnapType: 'x mandatory' } : {}),
					scrollbarWidth: 'none',
					'&::-webkit-scrollbar': { display: 'none' },
					'& > .swiper-slide': { flexShrink: 0, height: 'auto', boxSizing: 'border-box', display: 'flex', scrollSnapAlign: 'start' },
					'& > .swiper-slide ~ .swiper-slide': { marginLeft: media.mdDown ? 0 : 20 }
				})
			)}
		>
			{#if !groupKey}
				<div class={cx('swiper-slide', toClass(slideW))}>
					<CodeCategoryListTable
						{title}
						onChange={(o) => doSave(o)}
						{doSortUpdate}
						depth={1}
						maxDepth={3}
						{codeItems}
						spreadParentId={spreadIds[0] || 0}
						onDelete={(item) => ((spreadIds = []), doSave(item))}
						{meta}
					/>
				</div>
			{/if}
			{#each spreadIds as _sid, i (i)}
				{@const depth = i + 2}
				{@const sp = getSpreadItem(i)}
				{@const base = codeItems.find((x) => x.id == spreadIds[0])}
				{@const mx = maxDepthOf(meta?.metaKey, base)}
				{#if mx >= depth}
					<div class={cx('swiper-slide', toClass(subW))}>
						<CodeCategoryListTable
							title={sp.titles.join(' > ')}
							caption={`${groupKey ? i + 1 : depth}단계`}
							componentName={groupKey ? componentName : ''}
							spreadParentId={spreadIds[i + 1]}
							parentIds={sp.parentIds}
							onChange={(o) => saveDeep(depth, sp.parentIds, o)}
							doSortUpdate={(os) => sortDeep(depth, sp.parentIds, os)}
							{depth}
							maxDepth={mx}
							codeItems={sp.parent.items || []}
							isSpread
							onDelete={async (b, code) => {
								await doSave(b);
								if (spreadIds[depth - 1] == code.id) spreadIds = spreadIds.slice(0, depth - 1);
							}}
							{meta}
						/>
					</div>
				{/if}
			{/each}
		</div>
		{#if canPrev && !media.mdDown}<button type="button" aria-label="이전" class={cx('swiper-button-prev', navBtn, toClass({ left: 10 }))} onclick={() => slideBy(-1)}
				><Icon name="arrow_back_ios" /></button
			>{/if}
		{#if canNext && !media.mdDown}<button type="button" aria-label="다음" class={cx('swiper-button-next', navBtn, toClass({ right: 10 }))} onclick={() => slideBy(1)}
				><Icon name="arrow_forward_ios" /></button
			>{/if}
	</div>
</div>
