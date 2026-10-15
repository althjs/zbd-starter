<script lang="ts">
	/** 회사 정보 입력(React setting/meta/site/SiteBasicInfoForm) — 기본·은행·기타(분석·지도·사진 키) */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import TextField from '../../ui/TextField.svelte';
	import AddressFields from '../ui/AddressFields.svelte';
	import type { Address } from '#lib/shared/types/member.ts';

	let { item, onChange, useEtc = true }: { item: any; onChange: (o: any) => void; useEtc?: boolean } = $props();
	const set = (k: string, v: any) => onChange({ ...item, [k]: v });
	// 예전 값은 한 줄 글(companyAddress)뿐 — 찾기 전까지는 그 글을 주소 칸에 보인다
	const address = $derived<Address>(item.companyAddressInfo || (item.companyAddress ? { address1: item.companyAddress } : {}));
	const setAddress = (a: Address) => onChange({ ...item, companyAddressInfo: a, companyAddress: [a.address1, a.address2].filter(Boolean).join(' ') });
	// 칸들은 줄바꿈 — 휴대폰에서 다섯 칸이 한 줄에 눌려 라벨이 「대…」 로 잘리고, 최소 폭(350 등) 칸은 카드 밖으로 삐져나갔다(2026-10-04 대표). 칸 폭은 아래 style 의 flex-basis
	const row = cx('MuiStack-root sbi-row', toClass({ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }));
</script>

<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
	<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
		<h6 class="MuiTypography-root MuiTypography-subtitle2">기본정보</h6>
		<div class={row}>
			<TextField label="대표자 성명" value={item.ceoName} oninput={(v) => set('ceoName', v)} />
			<TextField label="회사명" value={item.companyName} oninput={(v) => set('companyName', v)} />
			<TextField label="전화번호" value={item.companyTel} oninput={(v) => set('companyTel', v)} />
			<TextField label="팩스번호" value={item.companyFax} oninput={(v) => set('companyFax', v)} />
			<TextField label="이메일" value={item.companyEmail} oninput={(v) => set('companyEmail', v)} />
		</div>
	</div>
	<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
		<div class={row}>
			<TextField label="사업자등록번호" value={item.companyNo} oninput={(v) => set('companyNo', v)} />
			<TextField label="통신판매업신고번호" value={item['통신판매업신고번호']} oninput={(v) => set('통신판매업신고번호', v)} />
			<TextField label="개인정보관리책임자" value={item['개인정보관리책임자']} oninput={(v) => set('개인정보관리책임자', v)} />
			<TextField label="업종" value={item.companyType} oninput={(v) => set('companyType', v)} />
		</div>
		<div class={row}>
			<!-- 회사 주소 = 카카오 주소 찾기(Address 모양 companyAddressInfo) — 푸터 매크로 {companyAddress} 는 한 줄 글이라 같이 맞춘다(2026-10-04 대표). 지역 칸은 뺐다 -->
			<AddressFields label="회사 주소" value={address} onChange={setAddress} />
			{#if item.companyAddressOldType}<TextField label="회사 주소 (구 주소형식)" sx={{ flexBasis: 300 }} value={item.companyAddressOldType} oninput={(v) => set('companyAddressOldType', v)} />{/if}
		</div>
	</div>
	<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
		<h6 class="MuiTypography-root MuiTypography-subtitle2">은행정보</h6>
		<div class={row}>
			<TextField label="은행명" value={item.bank?.bankName || ''} oninput={(v) => set('bank', { ...(item.bank || {}), bankName: v })} />
			<TextField label="계좌번호" value={item.bank?.bankNo || ''} oninput={(v) => set('bank', { ...(item.bank || {}), bankNo: v })} />
			<TextField label="예금주" value={item.bank?.name || ''} oninput={(v) => set('bank', { ...(item.bank || {}), name: v })} />
		</div>
	</div>
	{#if useEtc}
		<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
			<h6 class="MuiTypography-root MuiTypography-subtitle2">기타정보</h6>
			<div class={row}>
				<TextField label="구글애널리틱스 ID" sx={{ flexBasis: 150 }} value={item.ga4Id || ''} oninput={(v) => set('ga4Id', v)} />
				<TextField label="구글맵 키" sx={{ flexBasis: 350 }} value={item.apiKeys?.googleMapKey || ''} oninput={(v) => set('apiKeys', { ...(item.apiKeys || {}), googleMapKey: v })} />
				<TextField
					label="Unsplash Access Key"
					sx={{ flexBasis: 360 }}
					value={item.apiKeys?.unsplashAccessKey || ''}
					placeholder="없으면 사진 고르기에서 Unsplash 탭이 숨겨져요"
					oninput={(v) => set('apiKeys', { ...(item.apiKeys || {}), unsplashAccessKey: v.trim() })}
				/>
				<TextField label="네이버 웹마스터도구 도메인 인증키" sx={{ flexBasis: 360 }} value={item.webMasterToolNaverKey || ''} oninput={(v) => set('webMasterToolNaverKey', v)} />
				<TextField label="네이버 MAP ClientID" sx={{ flexBasis: 360 }} value={item.naverMapClientId || ''} oninput={(v) => set('naverMapClientId', v)} />
			</div>
		</div>
	{/if}
</div>

<style>
	/* 기본 한 칸 140px 이상, 남는 폭은 나눠 갖고 · 폭을 정한 칸(flexBasis)도 화면보다 넓어지지 않게 */
	/* :where — 구체성 0 이라 칸마다 준 sx flexBasis 가 이긴다 */
	:global(:where(.sbi-row > *)) {
		flex: 1 1 140px;
		min-width: 0;
		max-width: 100%;
	}
</style>
