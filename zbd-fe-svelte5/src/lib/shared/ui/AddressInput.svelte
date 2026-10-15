<script lang="ts">
	/**
	 * 주소 입력(사용자 화면) — 예전 fe-service components/common/address/AddressInput + KakaoAddressDialog.
	 * 주소(검색으로만 채움)·우편번호 · 「주소 검색」(카카오 우편번호 — 스크립트는 처음 열 때 읽는다) · 상세주소.
	 * 값 모양 {address1, address2, zipcode} = 회원 json.address(서버 checkRequiredUserInfo 가 address1 을 본다).
	 * 관리자 화면은 admin/design/ui/AddressInput(관리자 부품) — 같은 동작.
	 */
	import Dialog from '../mui/Dialog.svelte';
	import MuiButton from '../mui/MuiButton.svelte';
	import TextField from '../mui/TextField.svelte';
	import { cx, sx as toClass } from '../mui/sx.ts';

	let { detail, onChange, fullScreen = false }: { detail: any; onChange: (v: any) => void; fullScreen?: boolean } = $props();
	let open = $state(false);
	let ready = $state(typeof window != 'undefined' && !!(window as any).kakao?.Postcode);
	let box = $state<HTMLDivElement>();
	let wrap = $state<HTMLDivElement>();

	function load() {
		if ((window as any).kakao?.Postcode) return void (ready = true);
		if (document.getElementById('kakao-address')) return;
		const s = document.createElement('script');
		s.id = 'kakao-address';
		s.src = 'https://t1.kakaocdn.net/mapjsapi/bundle/postcode/prod/postcode.v2.js';
		s.onload = () => (ready = true);
		document.body.appendChild(s);
	}
	$effect(() => {
		if (!open || !ready) return;
		const t = setTimeout(() => {
			if (!box) return;
			new (window as any).kakao.Postcode({
				oncomplete(data: any) {
					onChange({ ...(detail || {}), address1: data.roadAddress || data.address, zipcode: data.zonecode });
					open = false;
					// 상세주소 칸으로
					setTimeout(() => wrap?.querySelector<HTMLInputElement>('.address2 input')?.focus(), 50);
				},
				width: '100%',
				height: '100%',
				maxSuggestItems: 5
			}).embed(box);
		}, 100);
		return () => clearTimeout(t);
	});
</script>

<div bind:this={wrap} class={cx('MuiStack-root', toClass({ gap: 8 }))}>
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
		<div class={toClass({ flex: 1, minWidth: 0 })}>
			<TextField value={detail?.address1 || ''} size="small" placeholder="주소 검색을 눌러 주세요" disabled sx={{ width: '100%' }} />
		</div>
		<MuiButton variant="outlined" sx={{ height: 40, minWidth: 96, flexShrink: 0 }} onclick={() => (load(), (open = true))}>주소 검색</MuiButton>
	</div>
	<div class="address2">
		<TextField value={detail?.address2 || ''} size="small" placeholder="상세주소" disabled={!detail?.address1} sx={{ width: '100%' }} oninput={(v) => onChange({ ...(detail || {}), address2: v })} />
	</div>
	{#if detail?.zipcode}<p class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: '#98a2b3', margin: 0 }))}>우편번호 {detail.zipcode}</p>{/if}
</div>

<!-- 폭을 정해 둔다 — 안 정하면 창이 내용 폭으로 쪼그라들고 카카오 검색 화면(폭 100%)이 기댈 폭이 없다(2026-10-04) -->
<Dialog {open} onclose={() => (open = false)} {fullScreen} maxWidth={520} paperSx={fullScreen ? null : { width: 'calc(100% - 32px)' }}>
	{#snippet title()}주소 검색{/snippet}
	<div bind:this={box} class={toClass({ width: '100%', height: fullScreen ? 'calc(100vh - 120px)' : 470 })}>
		{#if !ready}<p class="MuiTypography-root MuiTypography-body2" style="padding:16px">불러오는 중…</p>{/if}
	</div>
</Dialog>
