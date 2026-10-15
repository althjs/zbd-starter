<script lang="ts">
	/**
	 * 머티리얼 아이콘 이름 또는 첨부 이미지(파일명 `200x200_…`)를 고르는 선택기(React components/common/MenuIconSelector).
	 * 아이콘(또는 이미지)을 누르면 창 — 이름을 적거나 이미지를 첨부. 값은 아이콘 이름 또는 파일명.
	 *   <MenuIconSelector icon={o.icon} hostingId={c} onChange={(v) => …} />
	 *   compact = 비었을 때 큰 버튼 대신 작은 아이콘 자리(하위 메뉴 줄)
	 */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import { ADMIN_MENU_ICON_IMAGE_ID } from '../../config.ts';
	import { adminImagePrefix } from '../../meta.svelte.ts';
	import OutlinedField from './OutlinedField.svelte';

	let {
		icon,
		onChange,
		hostingId = '',
		btnText = '아이콘을 입력하세요',
		imageId = '',
		folder = '',
		userId = null,
		sx = null,
		disabled = false,
		compact = false
	}: {
		icon: string;
		onChange: (v: string) => void;
		hostingId?: string;
		btnText?: string;
		imageId?: string;
		folder?: string;
		userId?: number | null;
		sx?: Sx | null;
		disabled?: boolean;
		compact?: boolean;
	} = $props();
	let open = $state(false);
	let value = $state('');
	let picker = $state<any>();
	const isImg = $derived(/^\d+x\d+_/.test(icon || ''));
	function openModal() {
		if (disabled) return;
		value = icon || '';
		open = true;
	}
</script>

{#if icon}
	{#if isImg}
		<img
			class={cx('MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img', toClass({ width: 25, height: 25, cursor: 'pointer', ...(sx || {}) }))}
			src={`${adminImagePrefix(hostingId)}/${imageId || ADMIN_MENU_ICON_IMAGE_ID}/${icon}`}
			alt=""
			role="presentation"
			onclick={openModal}
		/>
	{:else}
		<Icon name={icon} sx={{ cursor: 'pointer', ...(sx || {}) }} onclick={openModal} />
	{/if}
{:else if compact}
	<!-- 하위 메뉴처럼 좁은 줄 — 빈 아이콘 자리만(누르면 같은 창) -->
	<button type="button" class="mis-empty" title={btnText} aria-label={btnText} onclick={openModal}><Icon name="add_photo_alternate" size="Small" /></button>
{:else}
	<div class={cx('MuiStack-root', toClass({ marginBottom: 8 }))}>
		<Button color="error" sx={{ cursor: 'pointer' }} onclick={openModal}><Icon name="search_check" /> {btnText}</Button>
	</div>
{/if}

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
					width: 400,
					maxWidth: 'calc(100vw - 32px)',
					backgroundColor: 'var(--adm-paper)',
					border: '2px solid #000',
					boxShadow: '0px 11px 15px -7px rgba(0,0,0,0.2),0px 24px 38px 3px rgba(0,0,0,0.14),0px 9px 46px 8px rgba(0,0,0,0.12)',
					padding: 32,
					boxSizing: 'border-box'
				})
			)}
		>
			<div class={cx('MuiStack-root', toClass({ fontSize: 14 }))}>
				<h6 class="MuiTypography-root MuiTypography-h6">Material 아이콘 텍스트를 입력하세요</h6>
				<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineAlways" href="https://fonts.google.com/icons" target="_blank" rel="noopener">https://fonts.google.com/icons</a>
				에서 확인하실 수 있습니다.
				<OutlinedField size="small" bind:value />
				<div class={cx('MuiTypography-root MuiTypography-body2', toClass({ marginTop: 8 }))}>
					또는 아이콘 이미지를
					<span
						class={cx('MuiTypography-root MuiLink-root MuiLink-underlineAlways', toClass({ marginLeft: 4, marginRight: 4, cursor: 'pointer' }))}
						role="button"
						tabindex="0"
						onclick={() => picker?.open?.()}
						onkeydown={() => {}}>첨부</span
					>
					하세요
					<span style="display:inline-block">
						<ImagePicker
							bind:this={picker}
							size={{ width: 200, height: 200 }}
							image={isImg ? { name: icon } : null}
							{hostingId}
							{userId}
							folder={folder || imageId || ADMIN_MENU_ICON_IMAGE_ID}
							onSelect={(img) => (value = img?.name || '')}
						>
							{#snippet before()}{/snippet}
							{#snippet after()}{/snippet}
						</ImagePicker>
					</span>
				</div>
			</div>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', marginTop: 16, gap: 8 }))}>
				<Button variant="outlined" fullWidth onclick={() => (open = false)}>취소</Button>
				<Button variant="contained" fullWidth onclick={() => (onChange(value), (open = false))}>확인</Button>
			</div>
		</div>
	</div>
{/if}

<style>
	.mis-empty {
		display: inline-flex;
		padding: 2px;
		border: 1px dashed var(--adm-divider, #d0d5dd);
		border-radius: 6px;
		background: none;
		color: var(--adm-text2, #98a2b3);
		cursor: pointer;
	}
</style>
