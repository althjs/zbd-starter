<script lang="ts">
	/**
	 * 담당자 고르기(React AdminSelectPopover) — 사진 겹친 버튼 · 누르면 이름 검색 + 목록(누르면 넣고 빼기).
	 * targetAdminIds 가 있으면 그 사람들만 고를 수 있다.
	 */
	import Avatar from '../ui/Av.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import Popover from '../../ui/Popover.svelte';
	import Tooltip from '../../ui/Tooltip.svelte';
	import Chip from '../ui/Chip.svelte';
	import type { NotiMember } from '#lib/shared/types/work.ts';

	let {
		adminIds = [],
		targetAdminIds = [],
		admins = [],
		readonly = false,
		onchange
	}: { adminIds?: number[]; targetAdminIds?: number[]; admins?: NotiMember[]; readonly?: boolean; onchange: (ids: number[]) => void } = $props();
	let anchor = $state<HTMLElement | null>(null);
	let q = $state('');
	const ids = $derived(Array.isArray(adminIds) ? adminIds : []);
	const selected = $derived(admins.filter((m) => ids.includes(m.adminId)));
	const list = $derived(admins.filter((m) => (!q || m.name.toLowerCase().includes(q.toLowerCase())) && (!targetAdminIds?.length || targetAdminIds.includes(m.adminId))));
	const toggle = (id: number) => onchange(ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id]);
</script>

{#snippet btn()}
	<Button
		variant="outlined"
		size="small"
		onclick={(e) => !readonly && (anchor = e.currentTarget as HTMLElement)}
		sx={{
			minWidth: 100,
			minHeight: 32,
			padding: 4,
			borderColor: '#e0e0e0',
			color: 'var(--adm-text)',
			display: 'flex',
			justifyContent: 'flex-start',
			gap: 4,
			'&:hover': { borderColor: '#bdbdbd', backgroundColor: 'transparent' }
		}}
	>
		{#if selected.length}
			<div class="MuiStack-root" style="display:flex;flex-direction:row;margin-left:4px">
				{#each selected.slice(0, 3) as m, i (m.adminId)}<Avatar src={m.profileImage} size={24} sx={{ boxShadow: '0 0 2px rgba(0,0,0,0.1)', ...(i ? { marginLeft: -6.4 } : {}) }} />{/each}
				{#if selected.length > 3}<Avatar
						text={`+${selected.length - 3}`}
						size={24}
						sx={{ marginLeft: -6.4, fontSize: 11, backgroundColor: 'var(--zc-f5f5f5, #f5f5f5)', color: 'var(--adm-text2)', boxShadow: '0 0 2px rgba(0,0,0,0.1)' }}
					/>{/if}
			</div>
		{:else}
			<Avatar size={24} sx={{ marginLeft: 4, backgroundColor: 'var(--zc-eee, #eee)' }} />
		{/if}
	</Button>
{/snippet}

<div class="MuiBox-root">
	{#if selected.length}<Tooltip title={selected.map((m) => m.name).join(', ')} placement="top">{@render btn()}</Tooltip>{:else}{@render btn()}{/if}
</div>
<Popover {anchor} onclose={() => ((anchor = null), (q = ''))} paperSx={{ marginTop: 8, width: 320, borderRadius: 16, boxShadow: '0px 4px 20px rgba(0,0,0,0.1)', padding: 12 }}>
	<div class="MuiStack-root" style="display:flex;flex-direction:column;gap:12px">
		{#if selected.length}
			<div class="MuiBox-root" style="display:flex;flex-wrap:wrap;gap:4px">
				{#each selected as m (m.adminId)}<Chip size="small" label={m.name} ondelete={() => toggle(m.adminId)} sx={{ backgroundColor: '#e3f2fd' }} />{/each}
			</div>
		{/if}
		<div class="MuiFormControl-root MuiFormControl-fullWidth MuiTextField-root">
			<div
				class="MuiInputBase-root MuiOutlinedInput-root MuiInputBase-colorPrimary MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-sizeSmall MuiInputBase-adornedStart"
				style="border-radius:16px"
			>
				<Icon name="search" sx={{ color: 'var(--adm-text2)', fontSize: 20, marginRight: 8 }} />
				<!-- svelte-ignore a11y_autofocus -->
				<input class="MuiInputBase-input MuiOutlinedInput-input MuiInputBase-inputSizeSmall MuiInputBase-inputAdornedStart" placeholder="이름, 직위 또는 팀을 검색하세요." bind:value={q} autofocus />
				<fieldset aria-hidden="true" class="MuiOutlinedInput-notchedOutline"><legend><span class="notranslate">&ZeroWidthSpace;</span></legend></fieldset>
			</div>
		</div>
		<div class="MuiBox-root">
			<span class="MuiTypography-root MuiTypography-caption" style="display:block;color:var(--adm-text2);margin:0 0 4px 8px">추천인</span>
			<div style="max-height:240px;overflow-y:auto">
				{#each list as m (m.adminId)}
					{@const on = ids.includes(m.adminId)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<div
						role="option"
						aria-selected={on}
						tabindex="-1"
						onclick={() => toggle(m.adminId)}
						class="adm-cv-pick"
						style="display:flex;align-items:center;padding:8px 16px;border-radius:8px;cursor:pointer;margin-bottom:1.6px;background-color:{on ? '#f0f7ff' : 'transparent'}"
					>
						<div style="min-width:40px"><Avatar src={m.profileImage} size={32} /></div>
						<span class="MuiTypography-root MuiTypography-body1" style="font-size:13px;font-weight:{on ? 'bold' : 'normal'}">{m.name}</span>
					</div>
				{/each}
			</div>
		</div>
		<div class="MuiBox-root" style="background-color:#e3f2fd;padding:8px;border-radius:16px;display:flex;align-items:center;gap:8px">
			<Icon name="notifications" sx={{ color: '#1976d2', fontSize: 18 }} />
			<span class="MuiTypography-root MuiTypography-caption" style="color:#1976d2">담당자에게 알림이 전송됩니다.</span>
		</div>
	</div>
</Popover>
