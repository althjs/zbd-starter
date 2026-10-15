<script lang="ts">
	/**
	 * 약관 동의 — 「전체 동의」 + 약관마다 체크 · 이름을 누르면 약관 창(보고 나면 동의로 체크). 예전 components/common/Agreement.tsx
	 * 약관은 TERMS 게시판 글 중 groupSubKeys 에 든 것만(순서 그대로). onchange(모두 동의?, 약관별 동의 표)
	 */
	import DOMPurify from 'dompurify';
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import SvgIcon from '#lib/shared/mui/SvgIcon.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import { modal } from '#lib/shared/ui/dialog.svelte.ts';

	let {
		groupSubKeys,
		onchange,
		hiddenTitle = false,
		isMobile = false,
		help = ''
	}: { groupSubKeys: string[]; onchange: (all: boolean, map: Record<string, boolean>) => void; hiddenTitle?: boolean; isMobile?: boolean; help?: string } = $props();

	let terms = $state<any[]>([]);
	let agree = $state<boolean[]>([]);
	$effect(() => {
		api.get(`/service/user/terms?hostingId=${HOSTING_ID}`).then((r) => {
			if (!r.ok) return;
			terms = groupSubKeys.map((k) => (r.data || []).find((x: any) => String(x.groupSubKey).trim() == k)).filter(Boolean);
			agree = terms.map(() => false);
		});
	});
	// 약관 글이 하나도 없으면(새 사이트) 「전체 동의」 한 칸으로 — 약관은 관리자 게시판(TERMS)에 글을 쓰면 여기에 나온다
	let allBox = $state(false);
	const all = $derived(terms.length ? agree.every(Boolean) : allBox);
	function emit() {
		onchange(all, Object.fromEntries(terms.map((t, i) => [t.groupSubKey, agree[i]])));
	}
	function setAll(v: boolean) {
		allBox = v;
		agree = agree.map(() => v);
		emit();
	}
	function view(i: number) {
		const t = terms[i];
		modal({ title: t.title || t.groupSubKey, html: DOMPurify.sanitize(t.body || ''), wide: true });
		agree[i] = true;
		emit();
	}
	const box = (on: boolean) => (on ? 'CheckBox' : 'CheckBoxOutlineBlank');
</script>

<Stack class="agreement" sx={{ gap: 8 }}>
	{#if !hiddenTitle}
		<Stack direction="row" sx={{ alignItems: 'flex-end', gap: 8 }}>
			<svelte:element this={isMobile ? 'h5' : 'h3'} class="MuiTypography-root MuiTypography-{isMobile ? 'h5' : 'h3'}">약관동의</svelte:element>
			{#if help}<span class="MuiTypography-root MuiTypography-caption" style="color:#f04438">{help}</span>{/if}
		</Stack>
	{/if}
	<div
		class={cx(
			'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
			toClass({ paddingLeft: 16, paddingRight: 16, paddingTop: 8, paddingBottom: 8, [up('md')]: { paddingLeft: 24, paddingRight: 24 } })
		)}
	>
		<label class="MuiFormControlLabel-root MuiFormControlLabel-labelPlacementEnd">
			<span class="MuiButtonBase-root MuiCheckbox-root MuiCheckbox-sizeSmall{all ? ' Mui-checked' : ''}" style="position:relative">
				<input class="PrivateSwitchBase-input" type="checkbox" checked={all} onchange={(e) => setAll(e.currentTarget.checked)} />
				<SvgIcon name={box(all)} size="Small" />
			</span>
			<h6 class="MuiTypography-root MuiTypography-subtitle1" style="font-weight:500;font-size:15px;margin-left:-2px">전체 동의</h6>
		</label>
		<hr class="MuiDivider-root MuiDivider-fullWidth" />
		<Stack sx={{ marginTop: 8 }}>
			{#each terms as t, i (t.postId)}
				<Stack direction="row" sx={{ alignItems: 'center' }}>
					<span class="MuiButtonBase-root MuiCheckbox-root{agree[i] ? ' Mui-checked' : ''}" style="position:relative">
						<input class="PrivateSwitchBase-input" type="checkbox" checked={agree[i]} onchange={(e) => ((agree[i] = e.currentTarget.checked), emit())} />
						<SvgIcon name={box(agree[i])} />
					</span>
					<a
						class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineAlways"
						style="display:flex;align-items:center;cursor:pointer;justify-content:space-between;flex-grow:1"
						role="button"
						tabindex="0"
						onclick={() => view(i)}
						onkeydown={() => {}}
					>
						<p class="MuiTypography-root MuiTypography-body2" style="color:#606060;font-weight:500">{t.groupSubKey}</p>
						<Icon name="keyboard_arrow_right" sx={{ color: '#aaa' }} />
					</a>
				</Stack>
			{/each}
		</Stack>
	</div>
</Stack>
