<script lang="ts">
	/**
	 * 환경 설정 › 관리자 페이지(React setting/meta/site/SiteMetaManage) — SITE 메타: 제목·관리자 테마·본문 배경색·로고·파비콘·회사 정보.
	 * 저장 = PUT /service/admin/meta/SITE(?hostingId). 개발자는 JSON 직접 고치기.
	 * 최고 관리자(isSuperAdmin)면 「최고 관리자 설정」(super/ — 기능 켜기·끄기·커스텀 위젯·도움말 링크).
	 * (가맹점 계정·소속협회·관리자 I18N 은 즈비드에 없는 기능이라 옮기지 않았다)
	 */
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { adminImagePrefix, isSuperAdmin } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { ADMIN_THEMES } from '../../theme/palettes.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import ColorField from '../ui/ColorField.svelte';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import TagInput from '../../ui/TagInput.svelte';
	import OutlinedSelect from '../ui/OutlinedSelect.svelte';
	import SiteBasicInfoForm from './SiteBasicInfoForm.svelte';
	import SettingGroup from '../ui/SettingGroup.svelte';
	import SettingRow from '../ui/SettingRow.svelte';
	import SiteSuperAdminSettingForm from './super/SiteSuperAdminSettingForm.svelte';

	let { meta, hostingId, onChange }: { meta: any; hostingId: string; onChange: (m: any) => void } = $props();
	const siteDefault = { title: '', logo: null, favicon: null };
	let json = $state<any>({});
	let jsonOpen = $state(false);
	$effect(() => {
		void hostingId;
		const m = meta;
		untrack(() => (json = { ...siteDefault, ...clone(m.json || {}) }));
	});
	const diffs = $derived(diffList(meta.json, json).filter((d) => !(d.kind == 'add' && (d.after === '' || d.after === null) && ['title', 'logo', 'favicon'].includes(d.path))));
	const set = (k: string, v: any) => (json = { ...json, [k]: v });
	const setSeo = (k: string, v: any) => (json = { ...json, adminSeo: { ...(json.adminSeo || {}), [k]: v } });
	async function save() {
		const o = await api.put(`/service/admin/meta/${meta.metaKey}${hostingId ? `?hostingId=${hostingId}` : ''}`, json);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		onChange({ ...meta, json: clone(json) });
		toast('정상적으로 저장되었습니다.');
	}
	const PRESETS = [
		{ label: '흰색', value: '#ffffff' },
		{ label: '연회색', value: '#f6f6f6' },
		{ label: '회색', value: '#f2f2f2' }
	];
	const PRESETS_DARK = [
		{ label: '검정', value: '#0a0a0a' },
		{ label: '짙은 회색', value: '#121212' },
		{ label: '먹색', value: '#17191c' }
	];
	// 로고·파비콘 칸 = 이 설정 줄이 내 호스팅 것일 때. 줄의 호스팅(meta.hostingId)으로 본다 — json 안 hostingId 는 옛 데이터에만 있어 스타터 새 사이트에서 칸이 숨었다(2026-10-19)
	const showImages = $derived(admin.profile?.hostingId == '' || (!!admin.profile?.hostingId && admin.profile.hostingId == (hostingIdOf(meta) ?? hostingIdOf(json))));
	const dark = 'color-mix(in srgb, var(--theme-color-dark) 80%, transparent)';
	// 칸 설명(ⓘ 말풍선 — 공용 settings/ui/InfoTip)
	const TIPS: Record<string, { title?: string; lines: string[] }> = {
		'사무실 IP': {
			lines: [
				'사무실(회사)에서 인터넷에 나가는 공인 IP 예요. 여러 개면 콤마로.',
				'여기 적은 IP 는 접속 현황·웹로그에서 「사무실」 로 따로 묶이고, 외부 공격 차단 대상에서 빠지고, 공개 체험 횟수 제한에서 빠져요.',
				'서버 설정(.env OFFICE_IPS)에 적은 값도 함께 쓰여요. 바꾸면 1분 안에 반영돼요.'
			]
		},
		'페이지 타이틀': { lines: ['관리자 화면 브라우저 탭에 나오는 이름이에요.'] },
		'관리자 테마': { lines: ['관리자 화면의 기본 색 묶음(머리·버튼·강조색).', '사용자 사이트 색은 「홈페이지」 탭의 테마예요.'] },
		'본문 배경색': { lines: ['관리자 화면 내용 영역의 바탕색이에요.', '흰 카드가 잘 보이게 연회색을 권해요. 비우면 흰색.'] },
		'본문 배경색(다크)': { lines: ['다크 모드일 때 내용 영역의 바탕색이에요.', '카드(#1a1a1a)보다 어둡게 두면 카드가 잘 보여요. 비우면 다크 기본(#0a0a0a).'] },
		'관리자 로그인 유지 기간': {
			lines: [
				'관리자(backyard·메신저·메일 앱)가 이 기간 동안 접속하지 않으면 로그아웃돼요.',
				'접속할 때마다 다시 이 기간으로 늘어나요. 비우면 30일(최대 400일).',
				'회원 로그인 유지 기간은 「홈페이지」 탭에 있어요.'
			]
		},
		로고: { lines: ['관리자 화면 왼쪽 위에 나와요.', '사용자 사이트 로고는 「홈페이지」 탭에서 바꿔요.'] },
		파비콘: { lines: ['관리자 화면 브라우저 탭 아이콘이에요(정사각형).'] }
	};
</script>

{#snippet row(name: string, body: import('svelte').Snippet, basis = 0, grow = -1)}
	<SettingRow label={name} tip={TIPS[name]} {basis} {grow}>{@render body()}</SettingRow>
{/snippet}

<div class="MuiStack-root">
	{#snippet title()}<OutlinedField size="small" value={json.title || ''} oninput={(v) => set('title', v)} />{/snippet}
	{#snippet theme()}
		<div class={toClass({ maxWidth: 220 })}>
			<OutlinedSelect
				size="small"
				value={json.theme || ''}
				options={[{ value: '', label: '선택 안함' }, ...Object.keys(ADMIN_THEMES).map((k) => ({ value: k, label: k.toUpperCase() }))]}
				onchange={(v) => set('theme', v)}
			>
				{#snippet item(o)}<span class={toClass({ display: 'inline-flex', alignItems: 'center' })}
						>{#if o.value}<span class={toClass({ display: 'inline-block', background: ADMIN_THEMES[o.value]?.main, width: 18, height: 18, borderRadius: 4, marginRight: 8 })}
							></span>{/if}{o.label}</span
					>{/snippet}
			</OutlinedSelect>
		</div>
	{/snippet}
	{#snippet bgPick(key: string, presets: { label: string; value: string }[], hint: string)}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
			<ColorField value={json[key] || ''} sx={{ background: 'var(--zc-fff, #fff)', width: 150 }} onChange={(c) => set(key, c)} />
			{#each presets as p (p.value)}
				<Button size="small" variant={(json[key] || '').toLowerCase() === p.value ? 'contained' : 'outlined'} sx={{ minWidth: 0, paddingLeft: 8, paddingRight: 8 }} onclick={() => set(key, p.value)}>
					<span class={toClass({ display: 'inline-block', width: 14, height: 14, marginRight: 4, border: '1px solid var(--gt-400, #e1e1e1)', background: p.value })}></span>{p.label}
				</Button>
			{/each}
			<Button size="small" onclick={() => set(key, '')}>초기화</Button>
			<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-700, #909090)' }))}>{hint}</span>
		</div>
	{/snippet}
	{#snippet bg()}{@render bgPick('pageBgColor', PRESETS, '비워두면 흰색(기본).')}{/snippet}
	{#snippet bgDark()}{@render bgPick('pageBgColorDark', PRESETS_DARK, '비워두면 다크 기본.')}{/snippet}
	{#snippet officeIps()}
		<div class={cx('MuiStack-root', toClass({ gap: 6 }))}>
			<TagInput
				tags={(Array.isArray(json.officeIps) ? json.officeIps : []).join(',')}
				onChange={(v) => {
					const ips = [
						...new Set(
							String(v || '')
								.split(/[\s,]+/)
								.map((x) => x.trim())
								.filter((x) => /^[0-9a-fA-F:.\/]+$/.test(x))
						)
					];
					set('officeIps', ips.length ? ips : undefined);
				}}
				placeholder="예: 123.45.67.89 — 입력 후 엔터"
			/>
			<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-700, #909090)' }))}
				>사무실 공인 IP. 접속 현황·웹로그의 사무실 구분, 공격 차단 제외, 체험 횟수 제외에 쓰여요. 비워 두면 서버 설정(.env OFFICE_IPS)만 씁니다.</span
			>
		</div>
	{/snippet}
	{#snippet adminSession()}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
			<OutlinedField
				size="small"
				type="number"
				sx={{ width: 120 }}
				value={json.adminSessionDays ?? ''}
				placeholder="30"
				oninput={(v) => set('adminSessionDays', String(v).trim() === '' ? undefined : Math.max(1, Math.min(400, Math.floor(Number(v) || 0))) || undefined)}
			/>
			<span class="MuiTypography-root MuiTypography-body2">일</span>
			<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-700, #909090)' }))}
				>관리자(backyard·메신저·메일 앱)가 이 기간 동안 접속하지 않으면 로그아웃됩니다. 접속할 때마다 다시 이 기간으로 늘어납니다. 비워 두면 30일(최대 400일).</span
			>
		</div>
	{/snippet}
	{#snippet logo()}
		<ImagePicker
			size={{ width: 720, height: 156 }}
			image={json?.logo}
			{hostingId}
			folder="site"
			useUnsplash
			useOriginalFileName
			disableCrop
			useDelete
			onSelect={(i) => set('logo', { hostingId: admin.profile?.hostingId, name: i.name })}
		>
			{#snippet after()}<div
					class={cx('MuiStack-root', toClass({ alignItems: 'center', justifyContent: 'center', backgroundColor: dark, width: 'fit-content', height: 64, paddingLeft: 8, paddingRight: 8 }))}
				>
					<img
						class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img"
						style="width:200px;cursor:pointer"
						src={`${adminImagePrefix(json?.logo?.hostingId || hostingId)}/site/${json.logo?.name}`}
						alt=""
					/>
				</div>{/snippet}
		</ImagePicker>
	{/snippet}
	{#snippet fav()}
		<ImagePicker
			size={{ width: 300, height: 300 }}
			image={json?.favicon}
			{hostingId}
			folder="site"
			useUnsplash
			useOriginalFileName
			disableCrop
			useDelete
			onSelect={(i) => set('favicon', { name: i.name })}
		>
			{#snippet after()}<div class={cx('MuiStack-root', toClass({ alignItems: 'center', justifyContent: 'center', backgroundColor: dark, width: 'fit-content', height: 64 }))}>
					<img class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img" style="width:60px;height:60px;cursor:pointer" src={`${adminImagePrefix(hostingId)}/site/${json.favicon?.name}`} alt="" />
				</div>{/snippet}
		</ImagePicker>
	{/snippet}
	{#snippet adminOg()}
		<div class={cx('MuiStack-root', toClass({ gap: 12, width: '100%', maxWidth: 640 }))}>
			<OutlinedField
				size="small"
				fullWidth
				label="제목"
				placeholder={json.design?.title || json.title || '비우면 사이트 제목'}
				value={json.adminSeo?.title || ''}
				oninput={(v) => setSeo('title', v)}
			/>
			<OutlinedField
				size="small"
				fullWidth
				label="설명"
				placeholder={json.design?.description || '비우면 사이트 설명'}
				value={json.adminSeo?.description || ''}
				oninput={(v) => setSeo('description', v)}
			/>
			<p class="MuiTypography-root MuiTypography-caption">그림 — 1200×630 권장, 비우면 사이트 OG 이미지</p>
			<ImagePicker
				size={{ width: 800, height: 400 }}
				image={json.adminSeo?.image}
				{hostingId}
				folder="site"
				accept=".jpg, .jpeg, .png, .webp, .gif"
				useUnsplash
				useOriginalFileName
				disableCrop
				useDelete
				onSelect={(i) => setSeo('image', i?.name ? { name: i.name } : null)}
			>
				{#snippet after()}{#if json.adminSeo?.image?.name}<img
							class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img"
							style="width:240px;aspect-ratio:1200/630;object-fit:cover;border-radius:6px"
							src={`${adminImagePrefix(hostingId)}/site/${json.adminSeo.image.name}`}
							alt=""
						/>{/if}{/snippet}
			</ImagePicker>
			{#if json.adminSeo?.image?.name}<Button size="small" sx={{ width: 'fit-content' }} onclick={() => setSeo('image', null)}>그림 빼기</Button>{/if}
		</div>
	{/snippet}
	{#snippet info()}<SiteBasicInfoForm item={json} onChange={(o) => (json = o)} />{/snippet}
	{#snippet sup()}<SiteSuperAdminSettingForm item={json} onChange={set} />{/snippet}
	<div class="sgrps">
		<SettingGroup title="관리자 화면" desc="관리자(backyard) 화면의 이름과 색" wrap>
			{@render row('페이지 타이틀', title, 160, 0)}
			{@render row('관리자 테마', theme, 170, 0)}
			{@render row('본문 배경색', bg, 360, 1)}
			{@render row('본문 배경색(다크)', bgDark, 360, 1)}
		</SettingGroup>
		<SettingGroup title="관리자 로그인" desc="로그인 유지 기간">
			{@render row('관리자 로그인 유지 기간', adminSession)}
			{@render row('사무실 IP', officeIps)}
		</SettingGroup>
		{#if showImages}
			<SettingGroup title="관리자 로고 · 파비콘" desc="관리자 화면 왼쪽 위 로고와 탭 아이콘" wrap>
				{@render row('로고', logo)}
				{@render row('파비콘', fav)}
			</SettingGroup>
		{/if}
		<SettingGroup title="관리자 주소 링크 미리보기" desc="관리자 주소를 메신저·카톡에 붙였을 때 나오는 제목·설명·그림 — 비운 칸은 「홈페이지」 탭의 사이트 값">
			{@render adminOg()}
		</SettingGroup>
		<SettingGroup title="회사 정보" desc="사이트 아래(푸터)·메일·약관에 쓰는 사업자 정보">
			{@render info()}
		</SettingGroup>
		{#if isSuperAdmin(admin.profile)}
			<SettingGroup title="최고 관리자 설정" desc="사이트 기능 켜기·끄기 — 최고 관리자만 보여요">
				{@render sup()}
			</SettingGroup>
		{/if}
	</div>

	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', gap: 8 }))}>
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}>
			<AuthButton variant="contained" sx={{ width: 'fit-content' }} disabled={!diffs.length} onclick={save}>저장</AuthButton>
			{#if admin.profile?.isDeveloper}<IconButton aria-label="JSON 편집" onclick={() => (jsonOpen = true)}><Icon name="data_object" /></IconButton>{/if}
		</div>
	</div>
	{#if diffs.length && admin.profile?.isDeveloper}<div class={cx('MuiStack-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>{/if}
	<JsonEditDialog bind:open={jsonOpen} value={json} onChange={(o) => (json = o)} title="SITE JSON" />
</div>

<style>
	.sgrps {
		display: flex;
		flex-direction: column;
		gap: 20px;
		margin-bottom: 24px;
	}
</style>
