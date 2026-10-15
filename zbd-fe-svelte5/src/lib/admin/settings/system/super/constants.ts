/**
 * 사이트 기능 설정 목록(React setting/meta/site/constants.ts siteFeatures) — SITE 메타 `json.feature` 의 켜고 끄는 칸.
 * 최고 관리자 설정(SiteSuperAdminSettingForm) 이 이 목록으로 스위치를 그린다. React 판과 같은 목록·순서로 유지.
 * 여행·커머스·쿠폰·포인트·항공·호텔·가상계좌 칸은 뺐다(값은 메타에 남으면 그대로 보존).
 */
export type SiteFeature = Record<string, any> | undefined | null;

export type SiteFeatureGroup = 'GENERAL' | 'MEMBER';

export const SITE_FEATURE_GROUP_NAMES: Record<SiteFeatureGroup, string> = {
	GENERAL: '일반',
	MEMBER: '회원'
};

export interface SiteFeatureItem {
	key: string;
	label: string;
	desc: string;
	descWarn?: string;
	type: string;
	group: SiteFeatureGroup;
	ignoreUI?: boolean;
	condition?: (feature: SiteFeature) => boolean;
}

export const siteFeatures: SiteFeatureItem[] = [
	{
		key: 'useMemberAddress',
		label: '주소정보 받기',
		desc: '회원가입폼에 주소정보도 받습니다.',
		type: 'boolean',
		group: 'MEMBER'
	},
	{
		key: 'useMemberDemographic',
		label: '생년월일/성별',
		desc: '생년월일/성별 정보를 필수로 받습니다.',
		type: 'boolean',
		group: 'MEMBER'
	}
];
