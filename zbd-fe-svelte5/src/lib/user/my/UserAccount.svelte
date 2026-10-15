<script lang="ts">
	/**
	 * 회원 정보 — 사진·이름·닉네임·한 줄 소개·이메일·(이메일 가입이면 비밀번호)·휴대폰. 예전 my/components/UserAccount.tsx
	 * 즈비드 설정에서 꺼진 것(휴대폰 인증·주소·아이디 따로·선택 약관) 은 옮기지 않았다 — 켜면 여기에 더한다.
	 */
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID, imagePrefix } from '#lib/shared/config.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import Avatar from '#lib/shared/mui/Avatar.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { alert, toast } from '#lib/shared/ui/dialog.svelte.ts';
	import AddressInput from '#lib/shared/ui/AddressInput.svelte';
	import DemographicInput, { demographicOk } from './DemographicInput.svelte';
	import { site } from '#lib/shared/site.svelte.ts';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import { session } from '../session.svelte.ts';

	let { onChangeMember }: { onChangeMember?: (m: any) => void } = $props();

	let member = $state<any>(null);
	let orig = '';
	let mobileHelp = $state('');

	function apply(oo: any) {
		oo.Json = typeof oo.json == 'string' ? JSON.parse(oo.json || '{}') : oo.json || {};
		delete oo.json;
		oo.password = '';
		oo.passwordOld = '';
		orig = JSON.stringify(oo);
		member = oo;
		if (session.profile) Object.assign(session.profile, oo, { image: { imageId: `u_${oo.userId}_my`, name: oo.image } });
	}
	$effect(() => {
		api.get('/service/user/my/account').then((r) => {
			if (r.ok) apply(r.data);
			else alert('회원정보 로딩 도중 오류가 발생하였습니다.');
		});
	});

	const passwordHelp = $derived(
		member?.password && !/^(?=.*[A-Za-z])(?=.*\d)(?=.*[^A-Za-z\d]).{8,20}$/.test(member.password) ? '8자 이상 20자 이하 영문, 숫자, 특수문자를 사용하여 입력해주세요.' : ''
	);
	// 휴대폰은 선택 — 비어 있어도 저장된다(가입 화면이 휴대폰을 받지 않아 새 회원은 닉네임·비밀번호도 못 바꿨다, 2026-10-06 E2E J3). 넣었으면 형식만 본다
	const mobileOk = $derived(!member?.mobile || /^01\d-\d{3,4}-\d{4}$/.test(member.mobile));
	const valid = $derived(!!member && mobileOk && !passwordHelp && !mobileHelp && (!site.base?.siteMeta?.feature?.useMemberDemographic || demographicOk(member?.Json)));
	const edited = $derived(!!member && JSON.stringify(member) != orig);

	const phone = (v: string) => v.replace(/[^0-9]/g, '').replace(/^(\d{3})(\d{3,4})(\d{4})$/, '$1-$2-$3');
	async function mobileInput(v: string) {
		v = phone(v);
		member.mobile = v;
		if (!/^01\d-\d{3,4}-\d{4}$/.test(v)) return;
		const o = JSON.parse(orig);
		if (o.mobile == v) {
			mobileHelp = '';
			return;
		}
		const r = await api.post(`/service/user/${HOSTING_ID}/check-duplicated`, { mobileNo: v });
		if (!r.ok) mobileHelp = r.error.message || '이미 사용중인 휴대폰 번호입니다.';
		else {
			mobileHelp = '';
			toast('사용 가능한 휴대폰 번호입니다.');
		}
	}

	async function save() {
		const data = { ...member, json: JSON.stringify(member.Json) };
		delete data.Json;
		const r = await api.put('/service/user/my/account', data);
		if (!r.ok) return alert(r.error.code == '0001' ? '기존 패스워드가 일치하지 않습니다.' : '처리 도중 오류가 발생하였습니다.');
		apply(r.data);
		onChangeMember?.(member);
		toast('정상적으로 수정되었습니다.');
	}
	const label = cx('MuiTypography-root MuiTypography-subtitle2', toClass({ minWidth: 130 }));
</script>

{#snippet name_()}
	<Stack direction="row" sx={{ alignItems: 'center', marginTop: 4, '& > :not(style) ~ :not(style)': { marginLeft: 4 } }}>
		<ImagePicker size={{ width: 300, height: 300 }} image={member.image ? { name: member.image } : null} folder="my" useUnsplash userId={member.userId} onSelect={(img) => (member.image = img.name)}>
			{#snippet after()}<Avatar src={`${imagePrefix()}/u_${member.userId}_my/${member.image}`} size={50} sx={{ cursor: 'pointer' }} />{/snippet}
			{#snippet before()}<MuiButton size="small" variant="outlined" sx={{ width: '50px', height: '50px', padding: '0', minWidth: 50, borderRadius: 40 }}
					><span class="MuiTypography-root MuiTypography-caption" style="line-height:1.2">Profile <br />Image</span></MuiButton
				>{/snippet}
		</ImagePicker>
		{#if JSON.parse(orig).name}<p class="MuiTypography-root MuiTypography-body1">{member.name}</p>{:else}<TextField
				bind:value={member.name}
				size="small"
				placeholder="성명을 입력하세요"
				sx={{ marginTop: 8 }}
			/>{/if}
	</Stack>
{/snippet}
{#snippet nick()}<Stack direction="row" sx={{ justifyContent: 'space-between' }}><TextField bind:value={member.nickname} size="small" placeholder="닉네임을 입력하세요" sx={{ marginTop: 8 }} /></Stack
	>{/snippet}
{#snippet desc()}<Stack direction="row" sx={{ justifyContent: 'space-between' }}
		><TextField bind:value={member.description} size="small" multiline minRows={2} placeholder="한줄소개를 입력하세요" sx={{ marginTop: 8 }} /></Stack
	>{/snippet}
{#snippet email()}<Stack sx={{ '& > :not(style) ~ :not(style)': { marginTop: 8 } }}><p class="MuiTypography-root MuiTypography-body2" style="width:100%;word-break:break-all">{member.email}</p></Stack
	>{/snippet}
{#snippet pwOld()}<Stack direction="row" sx={{ justifyContent: 'space-between' }}
		><TextField bind:value={member.passwordOld} type="password" size="small" placeholder="암호를 입력하세요" sx={{ marginTop: 8 }} /></Stack
	>{/snippet}
{#snippet pwNew()}
	<Stack sx={{ justifyContent: 'space-between' }}>
		<TextField bind:value={member.password} type="password" size="small" placeholder="암호를 입력하세요" sx={{ marginTop: 8 }} />
		{#if passwordHelp}<h6 class="MuiTypography-root MuiTypography-subtitle2" style="color:#f04438">{passwordHelp}</h6>{/if}
	</Stack>
{/snippet}
{#snippet addressF()}
	<!-- 사이트 설정 「주소정보 받기」 일 때만(2026-10-04 — Svelte 이전 때 칸이 비어 있었다) -->
	<AddressInput detail={member.Json?.address || null} fullScreen={viewport.isMobile} onChange={(a) => (member.Json = { ...(member.Json || {}), address: a })} />
{/snippet}
{#snippet demoF()}
	<DemographicInput birthDate={member.Json?.birthDate || ''} gender={member.Json?.gender || ''} onChange={(v) => (member.Json = { ...(member.Json || {}), ...v })} />
{/snippet}
{#snippet mobile()}
	<Stack>
		<Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
			<div role="presentation" style="display:contents" oninput={(e) => mobileInput((e.target as HTMLInputElement).value)}>
				<TextField value={member.mobile || ''} size="small" placeholder="휴대폰번호를 입력해주세요" />
			</div>
		</Stack>
		{#if mobileHelp}<h6 class="MuiTypography-root MuiTypography-subtitle2" style="color:#f04438">{mobileHelp}</h6>{/if}
	</Stack>
{/snippet}

<Stack sx={{ gap: 16 }}>
	<div
		class={cx(
			'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
			toClass({ paddingLeft: 16, paddingRight: 16, paddingTop: 16, paddingBottom: 16, [up('md')]: { paddingLeft: 24, paddingRight: 24, paddingTop: 24, paddingBottom: 24 } })
		)}
	>
		{#if member}
			{@const forms = [
				{ name: '이름', key: 'name', body: name_ },
				{ name: '닉네임', key: 'nickname', body: nick },
				{ name: '한줄 소개', key: 'description', body: desc },
				{ name: '이메일', key: 'email', body: email },
				{ name: '기존 비밀번호', key: 'password', body: member.groupId == 0 ? pwOld : null },
				{ name: '신규 비밀번호', key: 'password_new', body: member.groupId == 0 ? pwNew : null },
				{ name: '휴대폰번호', key: 'mobile', body: mobile },
				{ name: '생년월일/성별', key: 'demography', body: site.base?.siteMeta?.feature?.useMemberDemographic ? demoF : null },
				{ name: '주소', key: 'address', body: site.base?.siteMeta?.feature?.useMemberAddress ? addressF : null }
			]}
			<!-- 조건이 꺼진 칸은 React 판처럼 빈 <span> 으로 자리만 -->
			{#each forms as o, idx (o.key)}
				{#if !o.body}
					<span></span>
				{:else}
					<Stack direction="column" sx={{ marginBottom: forms.length - 1 == idx ? 0 : 32, [up('md')]: { marginBottom: forms.length - 1 == idx ? 0 : 16 } }}>
						<Stack sx={{ justifyContent: 'space-between', flexDirection: 'column', alignItems: 'flex-start', [up('md')]: { flexDirection: 'row', alignItems: 'center' } }}>
							<div class={label}>{o.name}</div>
							<Stack sx={{ flexGrow: 1, width: '100%', [up('md')]: { width: 'auto' } }}>{@render o.body()}</Stack>
						</Stack>
					</Stack>
				{/if}
			{/each}
		{/if}
	</div>
	<MuiButton variant="contained" sx={{ width: '100%' }} disabled={!edited || !valid} onclick={save}>저장</MuiButton>
</Stack>
