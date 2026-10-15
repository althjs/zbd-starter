<script lang="ts">
	/**
	 * 내 계정(React my/components/Account) — 프로필 사진·닉네임·비밀번호 바꾸기·휴대폰·두레이 알림 주소.
	 * 읽기 `GET /service/admin/my/account` · 저장 `PUT /service/admin/my/account`(json 은 문자열). 기존 비밀번호가 틀리면 서버 code 0001.
	 */
	import { api } from '#lib/shared/api.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { adminImagePrefix, feature, getMeta } from '../../meta.svelte.ts';
	import { admin, updateProfile } from '../../profile.svelte.ts';
	import Button from '../../ui/Button.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { cloneDeep } from '../dates.ts';
	import { getPhoneNumberFormat } from '#lib/shared/phone.ts';
	import StandardInput from './StandardInput.svelte';

	let member = $state<any>(null);
	let orig = $state('');
	// 못 읽은 그림 주소 — 플래그 하나로 두면 한 번 깨진 뒤 새 사진을 골라도 계속 빈 얼굴로 남는다
	let brokenSrc = $state('');
	const avatarSrc = $derived(member?.json?.image?.name ? `${adminImagePrefix()}/admin_profile_${member.adminId}/${member.json.image.name}` : '');
	const broken = $derived(!!avatarSrc && brokenSrc === avatarSrc);

	function isValidPassword(v: string) {
		v = v || '';
		if (v.length < 8 || v.length > 20) return false;
		return /[A-Za-z]/.test(v) && /[0-9]/.test(v) && /[^A-Za-z0-9]/.test(v);
	}
	function apply(oo: any) {
		oo.password = '';
		oo.passwordOld = '';
		oo.json = typeof oo.json == 'string' ? JSON.parse(oo.json) : oo.json || {};
		member = oo;
		orig = JSON.stringify(oo);
		updateProfile({ ...oo, image: oo.json?.image || admin.profile?.image });
	}
	$effect(() => {
		api.get(`/service/admin/my/account`).then((o) => {
			if (o.ok) apply(o.data);
			else alert('회원정보 로딩 도중 오류가 발생하였습니다.');
		});
	});
	const edited = $derived(!!member && JSON.stringify(member) !== orig);
	const pwHelp = $derived(member?.password && !isValidPassword(member.password) ? '8자 이상 20자 이하 영문, 숫자, 특수문자를 사용하여 입력해주세요.' : '');
	const valid = $derived(!!member?.mobile && !(member?.passwordOld && !member?.password) && !pwHelp);

	async function save() {
		const data = { ...cloneDeep($state.snapshot(member)), json: JSON.stringify(member.json) };
		const o = await api.put(`/service/admin/my/account`, data);
		if (!o.ok && o.error.code == '0001') return alert('기존 패스워드가 일치하지 않습니다.');
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		apply(o.data);
		toast('정상적으로 수정되었습니다.');
	}
	const setJ = (k: string, v: any) => (member.json = { ...(member.json || {}), [k]: v });
	const showDooray = $derived(admin.profile?.hostingId == '' || !!getMeta('SITE', 'feature', 'useDoolay'));
	const rows = $derived([
		{ key: 'name', name: '이름' },
		{ key: 'nickname', name: '닉네임' },
		{ key: 'email', name: member?.email && member.email.indexOf('@') !== -1 ? '이메일' : '아이디' },
		{ key: 'passwordOld', name: '기존 비밀번호' },
		{ key: 'password', name: '신규 비밀번호' },
		{ key: 'mobile', name: '휴대폰번호' },
		...(showDooray ? [{ key: 'dooray', name: '두레이' }] : [])
	]);
</script>

{#snippet avatar()}
	<div
		class={cx(
			'MuiAvatar-root MuiAvatar-circular',
			toClass({
				position: 'relative',
				display: 'flex',
				alignItems: 'center',
				justifyContent: 'center',
				flexShrink: 0,
				width: 40,
				height: 40,
				borderRadius: '50%',
				overflow: 'hidden',
				cursor: 'pointer',
				color: '#fff',
				backgroundColor: broken || !member?.json?.image?.name ? 'rgb(189, 189, 189)' : 'transparent'
			})
		)}
	>
		{#if member?.json?.image?.name && !broken}
			<img class={cx('MuiAvatar-img', toClass({ width: '100%', height: '100%', objectFit: 'cover' }))} src={avatarSrc} alt="" onerror={() => (brokenSrc = avatarSrc)} />
		{:else}
			<svg class="MuiSvgIcon-root MuiAvatar-fallback" style="width:75%;height:75%;fill:currentColor" viewBox="0 0 24 24"
				><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"></path></svg
			>
		{/if}
	</div>
{/snippet}

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 24 }))}>
	{#if member}
		{#each rows as o, idx (o.key)}
			<div class={cx('MuiStack-root', toClass({ marginBottom: idx == 6 ? 0 : 16 }))}>
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }))}>
					<div class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ width: 100 }))}>{o.name}</div>
					<div class={cx('MuiStack-root', toClass({ flexGrow: 1 }))}>
						{#if o.key == 'name'}
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', '& > :not(style) ~ :not(style)': { marginLeft: 4 } }))}>
								<ImagePicker
									size={{ width: 300, height: 300 }}
									image={member.json?.image}
									folder={`admin_profile_${member.adminId}`}
									useUnsplash
									onSelect={(img: any) => setJ('image', { imageId: img.imageId, name: img.name })}
								>
									{#snippet after()}{@render avatar()}{/snippet}
									{#snippet before()}<Button size="small" sx={{ width: 100 }}>프로필 이미지</Button>{/snippet}
								</ImagePicker>
								<p class="MuiTypography-root MuiTypography-body1" style="margin-left:8px">{member.name}</p>
							</div>
						{:else if o.key == 'nickname'}
							<StandardInput mt={false} value={member.json?.nickname || ''} placeholder="닉네임을 입력하세요" oninput={(v) => setJ('nickname', v.substring(0, 20))} />
						{:else if o.key == 'email'}
							<div class="MuiStack-root">{member.email}</div>
						{:else if o.key == 'passwordOld'}
							<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between' }))}>
								{#if member.isPasswordResetRequired}
									<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ color: 'var(--adm-error-main)' }))}>임시 비밀번호로 로그인하셨습니다. 비밀번호를 변경해주세요.</h6>
								{:else}
									<StandardInput type="password" value={member.passwordOld || ''} placeholder="암호를 입력하세요" oninput={(v) => (member.passwordOld = v)} />
								{/if}
							</div>
						{:else if o.key == 'password'}
							<div class={cx('MuiStack-root', toClass({ justifyContent: 'space-between' }))}>
								<StandardInput type="password" value={member.password || ''} placeholder="암호를 입력하세요" oninput={(v) => (member.password = v)} />
								{#if pwHelp}<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ color: 'var(--adm-error-main)' }))}>{pwHelp}</h6>{/if}
							</div>
						{:else if o.key == 'mobile'}
							<StandardInput value={member.mobile || ''} placeholder="전화번호를 입력해주세요" oninput={(v) => (member.mobile = getPhoneNumberFormat(v))} />
						{:else if o.key == 'dooray'}
							<StandardInput
								value={member.json?.doorayNotiUrl || ''}
								placeholder="두레이 알림 URL을 입력해주세요"
								oninput={(v) => (member.json = { ...(member.json || {}), doorayNotiUrl: v, doolayNotiHost: admin.profile?.hostingId == '' ? '' : `https://${location.hostname}` })}
							/>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	{/if}
	<Button fullWidth variant="contained" sx={{ marginTop: 16 }} disabled={!edited || !valid} onclick={save}>저장</Button>
</div>
