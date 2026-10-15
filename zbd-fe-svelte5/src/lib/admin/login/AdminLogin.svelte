<script lang="ts">
	/**
	 * 관리자 로그인(React pages/login + AdminLogin) — 호스팅 · 아이디(이메일) · 암호 · 아이디 저장 → `POST /service/login`
	 * 2단계 인증이 켜진 계정은 6자리 코드를 한 번 더. 성공하면 서버가 세션 쿠키를 심고 들어오려던 주소(zbd:auth:adminLoginPath) 또는 /admin.
	 * 호스트로 호스팅이 정해진 주소(chat/mail/backyard.zbd.kr) 는 호스팅 칸을 숨긴다.
	 * 2026-10-07 운영과 같은 디자인(청록 바탕 + 흰 카드) — 로고만 스타터 것.
	 */
	import { goto } from '$app/navigation';
	import { api, saveToken } from '#lib/shared/api.ts';
	import { HOST_HOSTING, LS } from '../config.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import { getHostingItems } from '../meta.svelte.ts';
	import Logo from '../layout/Logo.svelte';
	import SvgIcon from '../ui/SvgIcon.svelte';
	import { alert } from '../ui/modal.svelte.ts';
	import { ensureBrowserUid } from '../uid.ts';
	import { cancelled, loginWithPasskey, passkeySupported } from '#lib/shared/passkey.ts';

	const hostHosting = typeof location != 'undefined' ? HOST_HOSTING[location.host] : '';
	function saved() {
		try {
			return JSON.parse(localStorage.getItem(LS.login) || 'null');
		} catch {
			return null;
		}
	}
	const s = saved() || { id: '', password: '', saveId: true };
	// 옛 이름 companyId — 2026-10 옮기는 기간(저장해 둔 로그인 값)
	// 스타터: 사이트(호스팅)가 하나면 그 사이트로 — 「공통」 이 기본으로 잡혀 첫 관리자가 로그인하지 못했다
	const realHostings = getHostingItems().filter((h: any) => h.hostingId);
	const onlyHosting = realHostings.length <= 1 ? realHostings[0]?.hostingId || HOSTING_ID : '';
	let account = $state({ id: s.id || '', password: '', saveId: s.saveId ?? true, hostingId: onlyHosting || (s.hostingId ?? s.companyId ?? hostHosting ?? '') });
	let showPw = $state(false);

	const hostingItems = getHostingItems();
	const hostings: { hostingId: string; name: string }[] = hostingItems.length
		? hostingItems
		: [
				// 설정(HOSTINGS) 을 못 읽었을 때만 — 호스팅 "" (공통) 과 즈비드
				{ hostingId: '', name: '공통' },
				{ hostingId: 'ZBD', name: '즈비드' }
			];
	const showHosting = !onlyHosting && !hostHosting && hostings.length > 1;
	const emailOk = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v);
	const valid = $derived(emailOk(account.id || '') && !!account.password);

	$effect(() => void ensureBrowserUid());

	async function doLogin() {
		if (account.saveId) localStorage.setItem(LS.login, JSON.stringify({ id: account.id, saveId: account.saveId, hostingId: account.hostingId }));
		else localStorage.removeItem(LS.login);
		const body: any = { ...account, memberType: 'A' };
		const o = await api.post('/service/login', body);
		if (!o.ok) return alert(o.error.message || '로그인 도중 오류가 발생하였습니다.');
		done(o.data?.token);
	}
	function done(token?: string) {
		saveToken('admin', token); // 운영은 서버가 세션 쿠키로 심는다 — 본문에 토큰이 없다
		const pre = localStorage.getItem(LS.loginPath);
		goto(pre || '/admin');
	}
	// 지문·얼굴(패스키) 로그인 — 마이페이지에서 등록한 기기만(서버가 계정을 정한다, 2단계 인증 대신)
	const canPasskey = passkeySupported();
	let passkeyBusy = $state(false);
	async function doPasskey() {
		passkeyBusy = true;
		try {
			done((await loginWithPasskey('admin')).token);
		} catch (e: any) {
			if (!cancelled(e)) alert(e?.message || '지문·얼굴 로그인에 실패했습니다.');
		} finally {
			passkeyBusy = false;
		}
	}
</script>

<div class="zl">
	<div class="zl-brand">
		<Logo height={30} color="#fff" heart="#ff8f87" />
		<span class="zl-chip">관리자</span>
	</div>

	<div class="zl-card">
		<h1 class="zl-title">로그인</h1>
		<p class="zl-sub">관리자 계정으로 계속하세요</p>

		{#if showHosting}
			<label class="zl-label" for="admin_hosting">호스팅</label>
			<div class="zl-field zl-select">
				<select id="admin_hosting" bind:value={account.hostingId}>
					{#each hostings as c (c.hostingId)}
						<option value={c.hostingId || ''}>{c.name}</option>
					{/each}
				</select>
				<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 10l5 5 5-5z" /></svg>
			</div>
		{/if}

		<label class="zl-label" for="admin_id">아이디</label>
		<div class="zl-field">
			<input
				id="admin_id"
				type="email"
				autocomplete="username"
				placeholder="이메일 주소"
				value={account.id}
				oninput={(e) => (account.id = (e.currentTarget as HTMLInputElement).value.substring(0, 45))}
				onblur={() => (account.id = account.id.trim())}
			/>
		</div>

		<label class="zl-label" for="admin_passworld">암호</label>
		<div class="zl-field">
			<input
				id="admin_passworld"
				type={showPw ? 'text' : 'password'}
				autocomplete="current-password"
				placeholder="암호"
				value={account.password}
				oninput={(e) => (account.password = (e.currentTarget as HTMLInputElement).value.substring(0, 30))}
				onkeydown={(e) => e.key == 'Enter' && valid && doLogin()}
			/>
			<button type="button" class="zl-eye" aria-label={showPw ? '암호 숨기기' : '암호 보기'} onclick={() => (showPw = !showPw)}>
				{#if showPw}
					<svg viewBox="0 0 24 24"
						><path
							d="M12 7c2.76 0 5 2.24 5 5 0 .65-.13 1.26-.36 1.83l2.92 2.92c1.51-1.26 2.7-2.89 3.43-4.75-1.73-4.39-6-7.5-11-7.5-1.4 0-2.74.25-3.98.7l2.16 2.16C10.74 7.13 11.35 7 12 7M2 4.27l2.28 2.28.46.46C3.08 8.3 1.78 10.02 1 12c1.73 4.39 6 7.5 11 7.5 1.55 0 3.03-.3 4.38-.84l.42.42L19.73 22 21 20.73 3.27 3zM7.53 9.8l1.55 1.55c-.05.21-.08.43-.08.65 0 1.66 1.34 3 3 3 .22 0 .44-.03.65-.08l1.55 1.55c-.67.33-1.41.53-2.2.53-2.76 0-5-2.24-5-5 0-.79.2-1.53.53-2.2m4.31-.78 3.15 3.15.02-.16c0-1.66-1.34-3-3-3z"
						/></svg
					>
				{:else}
					<svg viewBox="0 0 24 24"
						><path
							d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5M12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5m0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3"
						/></svg
					>
				{/if}
			</button>
		</div>

		<label class="zl-check">
			<input type="checkbox" bind:checked={account.saveId} />
			<span>아이디 저장</span>
		</label>

		<button type="button" class="zl-btn" disabled={!valid} onclick={() => doLogin()}>로그인</button>

		{#if canPasskey}
			<div class="zl-or"><span>또는</span></div>
			<button type="button" class="zl-btn zl-btn-line" disabled={passkeyBusy} onclick={doPasskey}><SvgIcon name="Fingerprint" />지문·얼굴로 로그인</button>
		{/if}
	</div>

	<p class="zl-foot">Powered by ZBD Starter</p>
</div>

<style>
	.zl {
		width: 100%;
		max-width: 400px;
		display: flex;
		flex-direction: column;
		align-items: center;
		color: #10221f;
	}
	.zl-brand {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 28px;
	}
	.zl-chip {
		padding: 4px 10px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.14);
		color: #e6f2ef;
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.02em;
	}
	.zl-card {
		width: 100%;
		box-sizing: border-box;
		background: #fff;
		border-radius: 20px;
		padding: 32px 28px 28px;
		box-shadow: 0 24px 60px rgba(4, 24, 21, 0.28);
		display: flex;
		flex-direction: column;
	}
	.zl-title {
		margin: 0;
		font-size: 24px;
		font-weight: 800;
		letter-spacing: -0.02em;
	}
	.zl-sub {
		margin: 6px 0 24px;
		font-size: 14px;
		color: #6b7c79;
	}
	.zl-label {
		font-size: 13px;
		font-weight: 600;
		color: #3d4f4c;
		margin: 0 0 6px;
	}
	.zl-field {
		position: relative;
		margin-bottom: 16px;
	}
	.zl-field input,
	.zl-field select {
		width: 100%;
		box-sizing: border-box;
		height: 50px;
		padding: 0 16px;
		border: 1.5px solid #dfe7e5;
		border-radius: 12px;
		background: #f7faf9;
		font: inherit;
		font-size: 16px;
		color: #10221f;
		outline: none;
		transition:
			border-color 0.15s,
			box-shadow 0.15s,
			background 0.15s;
		appearance: none;
	}
	.zl-field input::placeholder {
		color: #9aa8a5;
	}
	.zl-field input:focus,
	.zl-field select:focus {
		border-color: #31726c;
		background: #fff;
		box-shadow: 0 0 0 4px rgba(49, 114, 108, 0.14);
	}
	.zl-field input[type='password'],
	.zl-field input#admin_passworld {
		padding-right: 50px;
	}
	.zl-select svg {
		position: absolute;
		right: 14px;
		top: 50%;
		width: 22px;
		height: 22px;
		transform: translateY(-50%);
		fill: #6b7c79;
		pointer-events: none;
	}
	.zl-eye {
		position: absolute;
		right: 6px;
		top: 50%;
		transform: translateY(-50%);
		width: 40px;
		height: 40px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		cursor: pointer;
		display: grid;
		place-items: center;
	}
	.zl-eye svg {
		width: 20px;
		height: 20px;
		fill: #6b7c79;
	}
	.zl-eye:hover {
		background: #eef3f2;
	}
	.zl-check {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		color: #3d4f4c;
		cursor: pointer;
		margin: 2px 0 22px;
		align-self: flex-start;
	}
	.zl-check input {
		width: 18px;
		height: 18px;
		margin: 0;
		accent-color: #31726c;
	}
	.zl-btn {
		width: 100%;
		box-sizing: border-box;
		height: 52px;
		border: 0;
		border-radius: 12px;
		background: linear-gradient(135deg, #31726c, #1d433d);
		color: #fff;
		font: inherit;
		font-size: 16px;
		font-weight: 700;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		transition:
			opacity 0.15s,
			transform 0.1s;
	}
	.zl-btn:active:not(:disabled) {
		transform: scale(0.99);
	}
	.zl-btn:disabled {
		opacity: 0.4;
		cursor: default;
	}
	.zl-btn-line {
		background: #fff;
		color: #1d433d;
		border: 1.5px solid #dfe7e5;
	}
	.zl-btn-line:hover:not(:disabled) {
		background: #f3f8f7;
	}
	.zl-or {
		display: flex;
		align-items: center;
		gap: 12px;
		margin: 18px 0;
		font-size: 12px;
		color: #9aa8a5;
	}
	.zl-or::before,
	.zl-or::after {
		content: '';
		flex: 1;
		height: 1px;
		background: #e6ecea;
	}
	.zl-foot {
		margin: 24px 0 0;
		font-size: 12px;
		color: rgba(230, 242, 239, 0.6);
	}
	@media (max-width: 480px) {
		.zl-card {
			padding: 28px 20px 24px;
			border-radius: 18px;
		}
		.zl-brand {
			margin-bottom: 24px;
		}
	}
</style>
