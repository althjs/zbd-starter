<script lang="ts">
	/**
	 * 선택 약관 동의 항목(React components/common/user/OptionalTermsFilterForm) — 약관 게시판(TERMS) 하위 중
	 * 선택 동의(agreementOptional) 이거나 세부 항목(agreementItems)이 있는 것. 세부 항목은 체크, 아니면 「동의」 스위치.
	 */
	import Checkbox from '#lib/shared/mui/Checkbox.svelte';
	import Switch from '#lib/shared/mui/Switch.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { getTermPostGroup } from '../common.ts';

	type Map = Record<string, boolean | Record<string, boolean>>;
	let { agreeMap, onChange, readonly = false, target = '' }: { agreeMap: Map; onChange?: (m: Map) => void; readonly?: boolean; target?: '' | 'EMAIL' | 'SMS' } = $props();
	const subs = $derived((getTermPostGroup()?.items || []).filter((o: any) => o.agreementOptional || o.agreementItems?.length > 0));
	const disabled = $derived(readonly || !onChange);
</script>

{#each subs as o, i (i)}
	<div class="MuiStack-root">
		<span class="MuiTypography-root MuiTypography-caption">{o.groupSubKey}</span>
		{#if o.agreementItems?.length > 0}
			<div class={cx('MuiFormGroup-root MuiFormGroup-row', toClass({ display: 'flex', flexDirection: 'row', flexWrap: 'wrap', '& > label': { marginTop: -8 } }))}>
				{#each o.agreementItems.filter((it: string) => !target || target == it) as item (item)}
					<Checkbox
						checked={!!(agreeMap?.[o.groupSubKey] as any)?.[item]}
						label={item}
						onchange={(tf) => {
							if (disabled) return;
							const mm: Map = { ...(agreeMap || {}) };
							mm[o.groupSubKey] = { ...((mm[o.groupSubKey] as Record<string, boolean>) || {}), [item]: tf };
							onChange?.(mm);
						}}
					/>
				{/each}
			</div>
		{:else}
			<Switch
				checked={!!agreeMap?.[o.groupSubKey]}
				label="동의"
				onchange={(tf) => {
					if (disabled) return;
					onChange?.({ ...(agreeMap || {}), [o.groupSubKey]: tf });
				}}
			/>
		{/if}
	</div>
{/each}
