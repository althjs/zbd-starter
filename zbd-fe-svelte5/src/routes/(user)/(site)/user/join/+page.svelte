<script lang="ts">
	/**
	 * 회원 가입 — 이메일 · 비밀번호 · 이름(· 닉네임) · 약관 동의 → `POST /service/user/member/join` → 같은 값으로 로그인.
	 * 소셜 가입(간편 회원가입)은 플러그인. 이미 로그인돼 있으면 홈으로.
	 */
	import { api, saveToken } from '#lib/shared/api.ts';
	import { HOSTING_ID } from '#lib/shared/config.ts';
	import Container from '#lib/shared/mui/Container.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { alert } from '#lib/shared/ui/dialog.svelte.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Agreement from '#lib/user/Agreement.svelte';
	import DemographicInput, { demographicOk } from '#lib/user/my/DemographicInput.svelte';
	import { site } from '#lib/shared/site.svelte.ts';
	import { session } from '#lib/user/session.svelte.ts';

	const JOIN_KEYS = ['이용약관', '개인정보 처리방침', '마케팅 수신동의'];

	let email = $state('');
	let password = $state('');
	let password2 = $state('');
	let name = $state('');
	let nickname = $state('');
	let agreed = $state(false);
	let agreeMap = $state<any>({});
	let busy = $state(false);
	let birthDate = $state('');
	let gender = $state('');
	// 사이트 설정 「생년월일/성별」 이면 필수
	const useDemo = $derived(!!site.base?.siteMeta?.feature?.useMemberDemographic);

	$effect(() => {
		if (session.profile?.userId) location.replace('/');
	});

	const help = $derived({
		email: email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) ? '이메일 주소가 올바르지 않습니다.' : '',
		password: password && password.length < 8 ? '비밀번호는 8자 이상으로 입력하세요.' : '',
		password2: password2 && password2 != password ? '비밀번호가 서로 다릅니다.' : '',
		name: name && name.trim().length < 2 ? '이름을 2자 이상 입력하세요.' : ''
	});
	const valid = $derived(
		agreed &&
			!!email &&
			!!password &&
			password == password2 &&
			name.trim().length >= 2 &&
			!help.email &&
			!help.password &&
			!help.password2 &&
			!help.name &&
			(!useDemo || demographicOk({ birthDate, gender }))
	);

	async function save() {
		if (!valid || busy) return;
		busy = true;
		try {
			const data = {
				email: email.trim(),
				password,
				name: name.trim(),
				nickname: nickname.trim(),
				json: JSON.stringify({ agree: true, agreeMap, ...(useDemo ? { birthDate, gender } : {}) }),
				hostingId: HOSTING_ID,
				memberType: 'U'
			};
			const r = await api.post('/service/user/member/join', data);
			if (!r.ok) return alert(r.error.code == '1062' || /Duplicate/.test(r.error.message || '') ? '이미 가입된 이메일입니다.' : r.error.message || '가입 도중 오류가 발생하였습니다.');
			// 가입한 값으로 바로 로그인
			const o = await api.post('/service/login', { id: data.email, password, hostingId: HOSTING_ID, memberType: 'U' });
			await alert('회원가입이 완료되었습니다.');
			if (o.ok) {
				saveToken('user', o.data?.token); // 운영은 서버가 세션 쿠키로 심는다
				location.replace('/');
			} else location.replace('/login');
		} finally {
			busy = false;
		}
	}
	const label = 'MuiTypography-root MuiTypography-subtitle2';
</script>

{#snippet field(title: string, required: boolean, body: import('svelte').Snippet, msg = '')}
	<Stack>
		<Stack direction="row">
			<div class={cx(label, toClass({ marginBottom: 4 }))}>{title}</div>
			{#if required}<span class="MuiTypography-root MuiTypography-body2" style="margin-left:1.6px;color:#fb3b3b">*</span>{/if}
		</Stack>
		{@render body()}
		{#if msg}<p class="MuiTypography-root MuiTypography-body2" style="color:#f04438">{msg}</p>{/if}
	</Stack>
{/snippet}
{#snippet emailF()}<TextField bind:value={email} size="small" placeholder="이메일 주소 — 로그인 아이디가 됩니다" />{/snippet}
{#snippet pwF()}<TextField bind:value={password} size="small" type="password" placeholder="8자 이상" />{/snippet}
{#snippet pw2F()}<TextField bind:value={password2} size="small" type="password" placeholder="한 번 더 입력" />{/snippet}
{#snippet nameF()}<TextField bind:value={name} size="small" placeholder="이름" />{/snippet}
{#snippet nickF()}<TextField bind:value={nickname} size="small" placeholder="화면에 보일 이름(없으면 이름)" />{/snippet}
{#snippet demoF()}<DemographicInput
		{birthDate}
		{gender}
		onChange={(v) => {
			birthDate = v.birthDate;
			gender = v.gender;
		}}
	/>{/snippet}
{#snippet agreeF()}
	<Agreement
		hiddenTitle
		groupSubKeys={JOIN_KEYS}
		isMobile={viewport.isMobile}
		onchange={(all, map) => {
			agreed = all;
			agreeMap = map;
		}}
	/>
{/snippet}

<Container sx={{ paddingTop: viewport.isMobile ? 40 : 80, paddingBottom: 80 }}>
	<Stack sx={{ width: '100%', maxWidth: 475, margin: '0 auto' }}>
		<Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'center' }}>
			<h2 class={cx('MuiTypography-root MuiTypography-h2', toClass({ fontWeight: 700 }))}>회원 가입</h2>
		</Stack>
		<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 24, marginTop: 24 }))}>
			<Stack class="max" sx={{ '& > :not(style) ~ :not(style)': { marginTop: 16 } }}>
				{@render field('이메일', true, emailF, help.email)}
				{@render field('비밀번호', true, pwF, help.password)}
				{@render field('비밀번호 확인', true, pw2F, help.password2)}
				{@render field('이름', true, nameF, help.name)}
				{@render field('닉네임', false, nickF)}
				{#if useDemo}{@render field('생년월일/성별', true, demoF)}{/if}
				{@render field('약관 동의', true, agreeF)}
			</Stack>
			<Stack sx={{ marginTop: 16 }}>
				<MuiButton variant="contained" sx={{ width: '100%' }} disabled={!valid || busy} onclick={save}>회원 가입</MuiButton>
			</Stack>
			<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ marginTop: 16, textAlign: 'center' }))}>
				이미 계정이 있나요? <a href="/login">로그인</a>
			</p>
		</div>
	</Stack>
</Container>
