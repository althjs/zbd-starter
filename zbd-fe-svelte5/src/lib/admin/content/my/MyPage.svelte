<script lang="ts">
	/** 마이페이지(React my/index) — 계정 · 푸시 알림(사이트 기능 켜졌거나 최고관리자·개발자) · 인증 기기 · 2단계 인증 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import { getCompanyName, getMeta, isSuperAdmin } from '../../meta.svelte.ts';
	import { setCurrentNav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Account from './Account.svelte';
	import LoginDevice from './LoginDevice.svelte';
	import Passkeys from './Passkeys.svelte';

	// 권한 메뉴에 없는 화면(공통 메뉴 -90000) — React 처럼 지금 메뉴를 직접 정한다(이동 경로·제목)
	$effect(() => {
		setCurrentNav({ title: '마이페이지', desc: '내 정보 관리', auth: 'RW', path: '/admin/my' }, isSuperAdmin(admin.profile) ? getCompanyName() : '');
	});
</script>

<PageContainer title="마이페이지" subtitle="내 정보를 관리합니다.">
	<div class={cx('MuiStack-root', toClass({ maxWidth: 600 }))}>
		<Account />
		<LoginDevice />
		<Passkeys />
	</div>
</PageContainer>
