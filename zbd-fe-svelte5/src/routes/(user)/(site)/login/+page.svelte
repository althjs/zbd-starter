<script lang="ts">
	/**
	 * 로그인 — 이메일·비밀번호, 등록한 기기면 지문·얼굴(패스키). 소셜 로그인은 플러그인.
	 * 로그인 뒤 돌아갈 곳은 쿠키 `__user_login_path`(doLogin 이 남긴다 — 이 화면은 ?redirect= 를 읽지 않는다).
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onDestroy } from 'svelte';
	import { api, saveToken } from '#lib/shared/api.ts';
	import { HOSTING_ID, COOKIE_HOST } from '#lib/shared/config.ts';
	import { getCookie, removeCookie } from '#lib/shared/cookie.ts';
	import HtmlView from '#lib/shared/html/HtmlView.svelte';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Checkbox from '#lib/shared/mui/Checkbox.svelte';
	import Container from '#lib/shared/mui/Container.svelte';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import { siteMetaOf } from '#lib/user/site.svelte.ts';
	import { COOKIE_USER_LOGIN_PATH_KEY, session } from '#lib/user/session.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';
	import { cancelled, loginWithPasskey, passkeySupported } from '#lib/shared/passkey.ts';

	const LOGIN_KEY = SK.auth.userLogin;
	const saved = (() => {
		try {
			return JSON.parse(localStorage.getItem(LOGIN_KEY) || 'null') || { id: '', saveId: true };
		} catch {
			return { id: '', saveId: true };
		}
	})();
	let id = $state(saved.id || '');
	let password = $state('');
	let saveId = $state(saved.saveId !== false);

	const siteMeta = $derived(siteMetaOf());
	const isMobile = $derived(viewport.isMobile);
	const valid = $derived(siteMeta?.feature?.useCustomAccountIdExceptEmail ? id.length >= 4 : /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(id) && !!password);

	$effect(() => {
		if (page.url.searchParams.get('status') == '-1') {
			alert('탈퇴한 회원입니다.');
			goto('/login', { replace: true });
		}
	});
	onDestroy(() => removeCookie(COOKIE_USER_LOGIN_PATH_KEY, { domain: COOKIE_HOST }));

	/** 돌아갈 곳 — 같은 사이트 경로이거나 대표 도메인(서브도메인 포함) 주소만. 아니면 홈 */
	function backUrl() {
		const v = getCookie(COOKIE_USER_LOGIN_PATH_KEY) || '';
		if (v.startsWith('/') && !v.startsWith('//')) return v;
		try {
			const u = new URL(v);
			const root = COOKIE_HOST.replace(/^\./, '');
			if (u.protocol == 'https:' && root && (u.hostname == root || u.hostname.endsWith('.' + root))) return u.href;
		} catch {}
		return '/';
	}
	// 이미 로그인돼 있으면 들어오려던 곳으로(미니홈피 서브도메인에서 「친구 신청」 → 로그인 화면이 홈으로 보내던 것, 2026-10-03 대표 신고)
	$effect(() => {
		if (!session.profile?.userId) return;
		const back = backUrl();
		removeCookie(COOKIE_USER_LOGIN_PATH_KEY, { domain: COOKIE_HOST });
		if (back.startsWith('/')) goto(back, { replace: true });
		else location.replace(back);
	});

	async function login() {
		if (saveId) localStorage.setItem(LOGIN_KEY, JSON.stringify({ id, saveId, hostingId: HOSTING_ID }));
		else localStorage.removeItem(LOGIN_KEY);
		const o = await api.post('/service/login', { id, password, saveId, hostingId: HOSTING_ID, memberType: 'U' });
		if (!o.ok) {
			alert(o.error.message || '로그인 도중 오류가 발생하였습니다.');
			return;
		}
		finish(o.data);
	}
	function finish(data: { token?: string; isAdditionalInfoRequired?: boolean }) {
		saveToken('user', data?.token); // 운영은 서버가 세션 쿠키로 심는다
		if (data?.isAdditionalInfoRequired) {
			location.replace('/user/join');
			return;
		}
		const back = backUrl();
		removeCookie(COOKIE_USER_LOGIN_PATH_KEY, { domain: COOKIE_HOST });
		location.replace(back);
	}

	// 지문·얼굴(패스키) 로그인 — 내 정보에서 등록한 기기만. 즈비드 앱(WebView) 은 브라우저 패스키를 못 써 숨긴다
	const canPasskey = passkeySupported() && !(typeof navigator != 'undefined' && navigator.userAgent.includes('ZBD'));
	let passkeyBusy = $state(false);
	async function passkeyLogin() {
		passkeyBusy = true;
		try {
			finish(await loginWithPasskey('user'));
		} catch (e: any) {
			if (!cancelled(e)) alert(e?.message || '지문·얼굴 로그인에 실패했어요.');
		} finally {
			passkeyBusy = false;
		}
	}

	const link = 'MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover';
</script>

<Container class="login-container" sx={{ paddingTop: isMobile ? 40 : 80 }}>
	<Stack sx={{ width: '100%', maxWidth: 475, margin: '0 auto' }}>
		<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center', marginBottom: 24, '& > :not(style) ~ :not(style)': { marginLeft: 4 } }}>
			<svelte:element this={isMobile ? 'h4' : 'h2'} class={cx(`MuiTypography-root MuiTypography-${isMobile ? 'h4' : 'h2'}`, toClass({ fontWeight: 700 }))}>로그인</svelte:element>
		</Stack>

		<div
			class={cx(
				'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
				toClass({ paddingTop: 24, paddingBottom: 24, marginTop: 16, paddingLeft: 12, paddingRight: 12, [up('sm')]: { paddingLeft: 24, paddingRight: 24 } })
			)}
		>
			<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ marginBottom: 4 }))}>아이디</h6>
			<TextField bind:value={id} size="small" maxLength={100} placeholder={siteMeta?.feature?.useCustomAccountIdExceptEmail ? 'ID를 입력하세요' : 'email 주소를 입력하세요'} />
			<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ marginTop: 16, marginBottom: 4 }))}>비밀번호</h6>
			<div role="presentation" onkeydown={(e) => e.key == 'Enter' && valid && login()}>
				<TextField bind:value={password} size="small" maxLength={100} placeholder="비밀번호를 입력하세요" type="password" />
			</div>
			<Checkbox bind:checked={saveId} label="아이디 저장" />
			<MuiButton variant="contained" disabled={!valid} onclick={login} sx={{ width: '100%', marginTop: 12, marginBottom: 12 }}>로그인</MuiButton>

			{#if canPasskey}
				<MuiButton variant="outlined" disabled={passkeyBusy} onclick={passkeyLogin} sx={{ width: '100%', marginBottom: 16, gap: 6 }}
					><Icon name="fingerprint" sx={{ fontSize: '20px' }} />지문·얼굴로 로그인</MuiButton
				>
			{/if}
			<Stack sx={{ gap: 40, marginTop: 8 }}>
				<Stack direction="row" sx={{ justifyContent: 'space-between' }}>
					<a class={cx(link, toClass({ display: 'flex', gap: 4, alignItems: 'center' }))} href="/user/join">
						<Icon name="person_add_alt" sx={{ fontSize: '18px' }} />
						<p class="MuiTypography-root MuiTypography-body2">회원 가입</p>
					</a>
					<!-- 사이트 주인이 첫 관리자 계정으로 여기서 로그인하다 막혔다(2026-10-14) — 사용자 주소의 /admin 은 관리자 주소로 넘어간다(nginx) -->
					<a class={cx(link, toClass({ display: 'flex', gap: 4, alignItems: 'center' }))} href="/admin">
						<Icon name="admin_panel_settings" sx={{ fontSize: '18px' }} />
						<p class="MuiTypography-root MuiTypography-body2">관리자 로그인</p>
					</a>
				</Stack>
			</Stack>
		</div>
	</Stack>
</Container>
