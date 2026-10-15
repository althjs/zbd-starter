<script lang="ts" module>
	/**
	 * 목록 컬럼 보기 옵션(React components/common/modal/ModalTableColumnOption) — 톱니 버튼 + 창.
	 * 값은 localStorage `zbd:admin:tableColumn:{optionKey}` · 호스팅 기본값은 COMMON_CODES 「TABLE_COLUMN_OPTIONS」.
	 * (React 의 끌어서 순서 바꾸기·묶음 컬럼·기본 글자 크기는 이 묶음 화면이 안 써서 옮기지 않았다)
	 */
	import { getMeta } from '../../meta.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';
	export type ColumnOption = { key: string; name: string; use?: boolean; width?: number; hideForAgency?: boolean; sticky?: boolean; [k: string]: any };

	function allItems(): any[] {
		const codes = (getMeta('COMMON_CODES')?.items || []).filter((x: any) => x.group == 'TABLE_COLUMN_OPTIONS');
		const all: any[] = [];
		const seen = new Set<string>();
		for (const c of codes) for (const it of c.items || []) if (it?.key && !seen.has(it.key)) (seen.add(it.key), all.push(it));
		return all;
	}
	export function getTableColumnOptionsDefault(optionKey: string): ColumnOption[] | null {
		const v = allItems().find((x) => x.key == optionKey)?.value;
		return v ? JSON.parse(JSON.stringify(v)) : null;
	}
	export function getTableColumnOptions(optionKey: string): ColumnOption[] {
		try {
			const s = localStorage.getItem(SK.admin.tableColumn(optionKey));
			if (s) return JSON.parse(s);
		} catch {}
		return allItems().find((x) => x.key == optionKey)?.value || [];
	}
</script>

<script lang="ts">
	import { api } from '#lib/shared/api.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Switch from '#lib/shared/mui/Switch.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { isSuperAdmin } from '../../meta.svelte.ts';
	import { getNextId } from '../common.ts';

	let {
		optionKey,
		options,
		defaultOptions,
		onChange,
		useCommonCodes = false,
		sx = null
	}: { optionKey: string; options: ColumnOption[]; defaultOptions: ColumnOption[]; onChange: (o: ColumnOption[]) => void; useCommonCodes?: boolean; sx?: Sx | null } = $props();
	let open = $state(false);
	const sup = $derived(isSuperAdmin(admin.profile));

	const ordered = $derived(
		defaultOptions.map((d) => {
			const saved = (options || []).find((o) => o.key === d.key);
			return saved ? { ...d, ...saved, name: d.name } : { ...d };
		})
	);
	$effect(() => {
		if (!options || options.length === 0) onChange(JSON.parse(JSON.stringify(defaultOptions)));
	});

	function change(key: string, value: any, field: 'use' | 'width' | 'hideForAgency' = 'use') {
		const next = ordered.map((o) => {
			if (o.key !== key) return { ...o };
			if (field == 'hideForAgency') {
				const { hideForAgency: _d, ...rest } = o;
				return value ? { ...o, hideForAgency: true } : rest;
			}
			if (field == 'width') {
				const w = Number(value);
				const { width: _d, ...rest } = o;
				return !Number.isFinite(w) || w <= 0 ? rest : { ...o, width: w };
			}
			return { ...o, use: value };
		});
		try {
			localStorage.setItem(SK.admin.tableColumn(optionKey), JSON.stringify(next));
		} catch {}
		onChange(next);
	}

	async function saveDefault() {
		if (!(await confirm('옵션을 기본값으로 저장 하시겠습니까?'))) return;
		const res = await api.get('/service/admin/meta/COMMON_CODES');
		if (!res.ok) return toast('오류가 발생했습니다.');
		const codes = res.data;
		const group = codes.items?.find((x: any) => x.group === 'TABLE_COLUMN_OPTIONS');
		if (!group) return toast('TABLE_COLUMN_OPTIONS 그룹을 찾을 수 없습니다.');
		const target = group.items?.find((x: any) => x.key === optionKey);
		if (target) {
			target.value = [...options];
			target.updatedAt = new Date();
		} else {
			group.items ||= [];
			group.items.push({ id: getNextId(group.items), name: optionKey, key: optionKey, value: [...options], updatedAt: new Date() });
		}
		const o = await api.put('/service/admin/meta/COMMON_CODES', codes);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		toast('정상적으로 저장되었습니다.');
	}
	async function reset() {
		if (!(await confirm('옵션을 초기화 하시겠습니까?'))) return;
		try {
			localStorage.removeItem(SK.admin.tableColumn(optionKey));
		} catch {}
		onChange(JSON.parse(JSON.stringify(getTableColumnOptionsDefault(optionKey) || defaultOptions)));
		open = false;
	}
</script>

<IconButton size="Small" sx={{ padding: 0, ...(sx || {}) }} onclick={() => (open = true)} aria-label="목록 보기 옵션">
	<Icon name="settings" size="Small" sx={{ fontSize: 15, color: 'var(--gt-700, #909090)' }} />
</IconButton>
{#if open}
	<div use:portal role="presentation" class="MuiModal-root" style="position:fixed;inset:0;z-index:1300">
		<div class="MuiBackdrop-root MuiModal-backdrop" aria-hidden="true" onclick={() => (open = false)}></div>
		<div
			class={cx(
				'MuiBox-root',
				toClass({
					position: 'absolute',
					top: '50%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					width: 'calc(100vw - 16px)',
					maxWidth: '100vw',
					overflowX: 'auto',
					backgroundColor: 'var(--adm-paper)',
					border: '2px solid #000',
					boxShadow: '0px 11px 15px -7px rgba(0,0,0,0.2),0px 24px 38px 3px rgba(0,0,0,0.14),0px 9px 46px 8px rgba(0,0,0,0.12)',
					padding: 16,
					maxHeight: '90vh',
					display: 'flex',
					flexDirection: 'column',
					'@media (min-width:600px)': { width: 600, padding: 32 }
				})
			)}
		>
			<h2 class="MuiTypography-root MuiTypography-h6">목록 보기 옵션</h2>
			<div class={cx('MuiStack-root', toClass({ marginTop: 16, flex: 1, minHeight: 0, overflow: 'auto', gap: 8, paddingRight: 4 }))}>
				{#each ordered as opt (opt.key)}
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }))}>
						<Switch checked={opt.use !== false} label={opt.name} onchange={(v) => change(opt.key, v)} />
						{#if sup}
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 8 }))}>
								<Switch checked={!!opt.hideForAgency} label="거래처 제외" onchange={(v) => change(opt.key, v, 'hideForAgency')} />
								<input
									type="number"
									min="0"
									placeholder="auto"
									value={opt.width ?? ''}
									disabled={opt.use === false}
									oninput={(e) => change(opt.key, (e.currentTarget as HTMLInputElement).value, 'width')}
									class={toClass({ width: 56, height: 24, padding: '0 4px', fontSize: 11, textAlign: 'right', border: '1px solid var(--gt-300, #ebebeb)', borderRadius: 4, outline: 'none' })}
								/>
								<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-600, #b0b0b0)' }))}>px</span>
							</div>
						{/if}
					</div>
				{/each}
			</div>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', marginTop: 8, flexShrink: 0 }))}>
				{#if useCommonCodes && sup}<Button size="small" variant="contained" onclick={saveDefault}>기본옵션으로 저장</Button>{/if}
				<Button size="small" onclick={reset}>옵션 초기화</Button>
			</div>
		</div>
	</div>
{/if}
