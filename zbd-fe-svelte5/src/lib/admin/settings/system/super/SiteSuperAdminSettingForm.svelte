<script lang="ts">
	/**
	 * 최고 관리자 사이트 설정(React setting/meta/site/SiteSuperAdminSettingForm) — 기능 켜기·끄기(SITE 메타 json.feature) · 관리자 도움말 링크(helpUrl).
	 * 환경 설정 › 관리자 페이지에서 isSuperAdmin 일 때만. 값은 늘 `{ ...item.feature }` 에 덮어 써서 이 화면이 모르는 칸은 그대로 남는다.
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Card from '../../../ui/Card.svelte';
	import TagInput from '../../../ui/TagInput.svelte';
	import Tooltip from '../../../ui/Tooltip.svelte';
	import OutlinedField from '../../ui/OutlinedField.svelte';
	import { SITE_FEATURE_GROUP_NAMES, siteFeatures, type SiteFeatureGroup } from './constants.ts';

	let { item, onChange }: { item: any; onChange: (k: string, v: any) => void } = $props();

	const GROUPS: SiteFeatureGroup[] = ['GENERAL', 'MEMBER'];
	const rowSx = toClass({ flexDirection: 'row', flexWrap: 'wrap', gap: 8, paddingLeft: 8 });
	const groupTitle = cx('MuiTypography-root MuiTypography-subtitle2', toClass({ paddingTop: 16 }));

	function setFeature(key: string, on: boolean) {
		const o = { ...(item.feature || {}) };
		if (!on) delete o[key];
		else o[key] = on;
		onChange('feature', o);
	}
	const tip = (f: { desc: string; descWarn?: string }) => (f.descWarn ? `${f.desc}\n${f.descWarn}` : f.desc);
</script>

{#snippet sw(checked: boolean, label: string, title: string, onchange: (v: boolean) => void)}
	<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
		<span class="MuiSwitch-root MuiSwitch-sizeSmall">
			<span class={cx('MuiButtonBase-root MuiSwitch-switchBase MuiSwitch-colorPrimary PrivateSwitchBase-root', checked && 'Mui-checked')}>
				<input class="PrivateSwitchBase-input MuiSwitch-input" type="checkbox" {checked} onchange={(e) => onchange((e.currentTarget as HTMLInputElement).checked)} />
				<span class="MuiSwitch-thumb"></span>
			</span>
			<span class="MuiSwitch-track"></span>
		</span>
		{#if title}
			<Tooltip {title} placement="top"><span class="MuiTypography-root MuiTypography-caption">{label}</span></Tooltip>
		{:else}
			<span class="MuiTypography-root MuiTypography-caption">{label}</span>
		{/if}
	</label>
{/snippet}

<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
	<h6 class="MuiTypography-root MuiTypography-subtitle2">사이트 기능설정</h6>

	{#each GROUPS as groupKey (groupKey)}
		{@const features = siteFeatures.filter((f) => f.group === groupKey && !f.ignoreUI && (!f.condition || f.condition(item.feature)))}
		{#if siteFeatures.some((f) => f.group === groupKey)}
			<h6 class={groupTitle}>{SITE_FEATURE_GROUP_NAMES[groupKey]} 설정</h6>
			<div class={cx('MuiStack-root', rowSx)}>
				{#each features as f (f.key)}
					{#if f.type == 'string[]'}
						<Card sx={{ padding: 8, width: '100%' }}>
							<span class="MuiTypography-root MuiTypography-caption">{f.label}</span>
							<TagInput
								tags={(item?.feature?.[f.key] || []).join(',')}
								placeholder={f.desc || ''}
								sx={{ width: '100%' }}
								onChange={(ts) => onChange('feature', { ...(item.feature || {}), [f.key]: !ts ? [] : ts.split(',') })}
							/>
						</Card>
					{:else}
						{@render sw(!!item?.feature?.[f.key], f.label, f.desc ? tip(f) : '', (v) => setFeature(f.key, v))}
					{/if}
				{/each}
			</div>
		{/if}
	{/each}

	<h6 class={groupTitle}>관리자 도움말 링크</h6>
	<OutlinedField size="small" value={item?.helpUrl || ''} placeholder="관리자 화면 맨 위 줄 「도움말」 이 여는 주소" oninput={(v) => onChange('helpUrl', v)} />
</div>
