<script lang="ts">
	/**
	 * 코드 하나 편집 서랍(React codes/CodeCategoryDrawer) — 「추가」 단추 + 오른쪽 서랍(분류명·설명·전시·상태 라벨·사용 + 하위 분류 표) + 아래 「저장」.
	 * 밖에서: bind:this → openDrawer(코드, component, parentIds) · updateModel(저장된 코드)
	 * 저장 안 한 내용은 서랍 밖(추가 단추 아래)에 바뀐 값으로 보인다(React 와 같다).
	 */
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import Icon from '../../ui/Icon.svelte';
	import Popover from '../../ui/Popover.svelte';
	import { confirm } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import SwitchField from '../ui/SwitchField.svelte';
	import CodeCategoryListTable from './CodeCategoryListTable.svelte';
	import StatusLabelEditor from './StatusLabelEditor.svelte';

	let {
		title,
		nextId,
		depth,
		maxDepth = null,
		readonly = false,
		onChange,
		useCreateButton = true,
		createComponentName = '',
		createParentIds = [],
		meta = null
	}: {
		title: string;
		nextId: any;
		depth: number;
		maxDepth?: number | null;
		readonly?: boolean;
		onChange: (o: any) => Promise<any> | void;
		useCreateButton?: boolean;
		createComponentName?: string;
		createParentIds?: any[];
		meta?: any;
	} = $props();

	let show = $state(false);
	let item = $state<any>(null);
	let orig = $state<any>(null);
	let xMaxDepth = $state<number | null>(null);
	let componentName = $state('');
	let parentIds = $state<any[]>([]);
	let labelAnchor = $state<HTMLElement | null>(null);

	const diffs = $derived(show && item ? diffList(orig, item) : []);
	const isStatusLabelCode = $derived(componentName === 'STATUS_LABEL' || item?.componentName === 'STATUS_LABEL');
	const isWide = $derived(!!(componentName || item?.componentName));
	const mx = $derived(xMaxDepth || maxDepth);

	export function updateModel(m: any) {
		orig = clone(m);
		item = clone(m);
	}
	export function openDrawer(code: any, component = '', pids: any[] = []) {
		orig = clone(code);
		item = clone(code);
		parentIds = pids || [];
		componentName = component || '';
		xMaxDepth = code.maxDepth || null;
		show = true;
	}
	const set = (k: string, v: any) => (item = { ...item, [k]: v });

	async function close() {
		if (diffs.length) {
			if (!(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
			item = clone(orig);
		}
		show = false;
	}
	function save() {
		const cur = clone(item);
		const df = onChange(cur);
		if (df) df.then(() => updateModel(cur));
	}
	/** 하위 분류 표에서 한 줄을 고치거나 더함 → 이 코드째 저장 */
	function saveChild(o: any) {
		let isNew = true;
		const os = (item.items || []).map((oo: any) => (oo.id == o.id ? ((isNew = false), o) : oo));
		if (isNew) os.push(o);
		const mm = { ...item, items: os };
		const df = onChange(mm);
		if (df) df.then(() => updateModel(mm));
	}

	const h6 = 'MuiTypography-root MuiTypography-h6';
	const paperW = $derived(isWide ? '70%' : '50%');
</script>

{#snippet row(name: string, body: import('svelte').Snippet)}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'start', width: '100%' }))}>
		<div class={cx('MuiStack-root', toClass({ flexGrow: 1, marginBottom: 32 }))}>
			<h6 class={cx(h6, toClass({ width: 150, marginBottom: 8 }))}>{name}</h6>
			{@render body()}
		</div>
	</div>
{/snippet}

{#snippet titleRow()}<div class={cx(h6, toClass({ display: 'flex', flexDirection: 'row', marginBottom: 8 }))}>{title} {item?.name ? ` > ${item.name}` : ''}</div>{/snippet}

{#snippet nameRow()}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<OutlinedField fullWidth size="small" placeholder="분류명을 입력하세요" disabled={readonly} value={item?.name || ''} oninput={(v: string) => set('name', v)} />
	</div>
{/snippet}

{#snippet descRow()}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<OutlinedField fullWidth size="small" placeholder="분류설명을 입력하세요" disabled={readonly} value={item?.desc || ''} oninput={(v: string) => set('desc', v)} />
	</div>
{/snippet}

{#snippet dispRow()}<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<SwitchField checked={!!item?.display} label="전시함" disabled={readonly} onchange={(v) => set('display', v)} />
	</div>{/snippet}

{#snippet labelRow()}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<Button variant="outlined" size="small" disabled={readonly} onclick={(e: MouseEvent) => (labelAnchor = e.currentTarget as HTMLElement)}>
			{#snippet startIcon()}<Icon name="label" sx={{ fontSize: '18px !important' }} />{/snippet}
			상태 {(item?.items || []).length}개
		</Button>
	</div>
{/snippet}

{#snippet useRow()}<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
		<SwitchField checked={!!item?.use} label="사용함" disabled={readonly} onchange={(v) => set('use', v)} />
	</div>{/snippet}

<div class="MuiStack-root">
	{#if useCreateButton}
		<Button
			variant="contained"
			sx={{ width: 'fit-content' }}
			disabled={readonly}
			onclick={() => {
				const code: any = { id: nextId, name: '', sortNo: '', use: true };
				if (depth <= (mx ?? 0)) code.items = [];
				openDrawer(code, createComponentName, createParentIds);
			}}>추가</Button
		>
	{/if}

	<Drawer open={show} anchor="right" onclose={close} paperSx={{ width: paperW }}>
		{#if item}
			<div class={cx('MuiStack-root', toClass({ padding: 24 }))}>
				<Card sx={{ padding: 24 }}>
					{@render row('', titleRow)}
					{@render row('분류명', nameRow)}
					{#if componentName == 'MEDICAL'}
						{@render row('분류설명', descRow)}
					{/if}
					{#if admin.profile?.isDeveloper}
						{@render row('전시(홈페이지)', dispRow)}
					{/if}
					{#if isStatusLabelCode}
						{@render row('상태 라벨', labelRow)}
					{/if}
					{@render row('사용여부', useRow)}
				</Card>
				<hr class={cx('MuiDivider-root', toClass({ height: 10, border: 0, margin: 0 }))} />

				{#if item.items && !isStatusLabelCode && (!maxDepth || maxDepth >= depth)}
					<Card sx={{ padding: 24 }}>
						<h6 class={cx(h6, toClass({ width: 150, marginBottom: 8 }))}>하위 분류 {JSON.stringify(parentIds)}</h6>
						<CodeCategoryListTable
							inDrawer
							componentName={componentName || item.componentName || ''}
							parentIds={[...parentIds, item.id]}
							doSortUpdate={(os) => {
								const m = { ...item, items: os };
								const df = onChange(m);
								if (df) df.then(() => updateModel(m));
							}}
							{meta}
							onChangeBySubComponent={(o) =>
								set(
									'items',
									item.items.map((oo: any) => (oo.id == o.id ? o : oo))
								)}
							onChange={(o) => saveChild(o)}
							title={`${title} > ${item.name}`}
							{depth}
							maxDepth={mx}
							codeItems={item.items}
							readonly={!String(item.name || '').trim()}
						/>
					</Card>
				{/if}
				<hr class={cx('MuiDivider-root', toClass({ height: 100, border: 0, margin: 0 }))} />
				<div
					class={cx(
						'MuiStack-root',
						toClass({ position: 'fixed', zIndex: 1, background: 'var(--zc-fff, #fff)', bottom: 0, right: 0, width: '100%', '@media (min-width:900px)': { width: `max(min(560px, 100vw), ${paperW})` } })
					)}
				>
					<AuthButton variant="contained" fullWidth sx={{ height: 48, borderRadius: 0 }} disabled={readonly || !diffs.length} onclick={save}>저장</AuthButton>
				</div>
			</div>
		{/if}
	</Drawer>

	{#if diffs.length && admin.profile?.isDeveloper}
		<div class={toClass({ margin: 16, maxWidth: 1000, overflow: 'hidden' })}><DiffPreview items={diffs} /></div>
	{/if}
</div>

<Popover anchor={labelAnchor} onclose={() => (labelAnchor = null)} paperSx={{ padding: 16, borderRadius: 16, boxShadow: '0px 4px 20px rgba(0,0,0,0.15)', minWidth: 360, maxWidth: 460 }}>
	<StatusLabelEditor value={item?.items || []} disabled={readonly} onChange={(list) => set('items', list)} />
	<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ display: 'block', marginTop: 8, color: 'var(--adm-text2)' }))}>변경 내용은 아래 [저장] 을 눌러야 반영됩니다.</span>
</Popover>
