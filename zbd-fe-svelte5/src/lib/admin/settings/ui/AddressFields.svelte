<script lang="ts">
	/**
	 * 주소 칸(관리자 설정 폼 모양 — 라벨이 상자 안 위인 TextField) — 우편번호·주소는 「주소 검색」(카카오 우편번호)으로만, 상세주소는 직접.
	 * 값 모양 = Address {zipcode, address1, address2}. 카카오 스크립트는 처음 열 때 읽는다(design/ui/AddressInput 과 같은 방식).
	 */
	import type { Address } from '#lib/shared/types/member.ts';
	import Button from '../../ui/Button.svelte';
	import Dialog from '../../ui/Dialog.svelte';
	import TextField from '../../ui/TextField.svelte';
	import { media } from '../../media.svelte.ts';

	let { value, label = '주소', onChange }: { value: Address | null | undefined; label?: string; onChange: (v: Address) => void } = $props();
	let open = $state(false);
	let ready = $state(!!(globalThis as any).kakao?.Postcode);
	let box = $state<HTMLDivElement>();
	let wrap = $state<HTMLDivElement>();

	function load() {
		if ((window as any).kakao?.Postcode) return void (ready = true);
		if (!document.getElementById('kakao-address')) {
			const s = document.createElement('script');
			s.id = 'kakao-address';
			s.src = 'https://t1.kakaocdn.net/mapjsapi/bundle/postcode/prod/postcode.v2.js';
			s.onload = () => (ready = true);
			document.body.appendChild(s);
		}
	}
	function search() {
		load();
		open = true;
	}
	$effect(() => {
		if (!open || !ready) return;
		const t = setTimeout(() => {
			if (!box) return;
			new (window as any).kakao.Postcode({
				oncomplete(data: any) {
					onChange({ ...(value || {}), zipcode: data.zonecode, address1: data.roadAddress || data.address });
					open = false;
					setTimeout(() => wrap?.querySelector<HTMLInputElement>('.addr2 input')?.focus(), 50);
				},
				width: '100%',
				height: '100%',
				maxSuggestItems: 5
			}).embed(box);
		}, 200);
		return () => clearTimeout(t);
	});
</script>

<div class="addr" bind:this={wrap}>
	<div class="zip"><TextField label="우편번호" readonly value={value?.zipcode || ''} placeholder="검색" fullWidth onclick={search} /></div>
	<div class="a1"><TextField {label} readonly value={value?.address1 || ''} placeholder="주소 검색을 눌러 주세요" fullWidth onclick={search} /></div>
	<Button variant="outlined" sx={{ minWidth: 96, alignSelf: 'stretch' }} onclick={search}>주소 검색</Button>
	<div class="a2 addr2"><TextField label="상세주소" value={value?.address2 || ''} disabled={!value?.address1} fullWidth oninput={(v) => onChange({ ...(value || {}), address2: v })} /></div>
</div>

<Dialog open={open && ready} onclose={() => (open = false)} title="주소 검색" maxWidth={false} mobileFullScreen paperSx={{ margin: 0, width: media.smDown ? '100%' : 'auto' }}>
	<div bind:this={box} style="width:{media.smDown ? '100%' : '500px'};height:{media.smDown ? 'calc(100vh - 120px)' : '470px'}"></div>
</Dialog>

<style>
	/* 우편번호 | 주소 | [주소 검색] | 상세주소 한 줄 — 좁으면 아래로 넘긴다 */
	.addr {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		width: 100%;
	}
	.zip {
		flex: 0 0 110px;
	}
	.a1 {
		flex: 2 1 260px;
		min-width: 0;
	}
	.a2 {
		flex: 1 1 200px;
		min-width: 0;
	}
	.zip :global(input),
	.a1 :global(input) {
		cursor: pointer;
	}
</style>
