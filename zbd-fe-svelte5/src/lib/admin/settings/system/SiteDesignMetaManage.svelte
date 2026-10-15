<script lang="ts">
	/**
	 * 환경 설정 › 국내/해외 홈페이지(React setting/system/components/SiteDesignMetaManage) — SITE 메타의 design(또는 design_outbound).
	 * 제목·SEO·robots·사이트맵·테마·홈페이지·로고·파비콘·OG·앱바 색·전역 CSS/HTML·로그인 안내 HTML. 저장 = PUT /service/admin/meta/SITE (json 통째).
	 * (I18N 사전은 isI18n 호스팅만 — 즈비드엔 없어 옮기지 않았다)
	 */
	import { hostingIdOf } from '#lib/shared/config.ts';
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { adminImagePrefix, siteHomepage } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { ADMIN_THEMES } from '../../theme/palettes.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import InfoTip from '../ui/InfoTip.svelte';
	import SettingGroup from '../ui/SettingGroup.svelte';
	import SettingRow from '../ui/SettingRow.svelte';
	import ColorField from '../ui/ColorField.svelte';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import OutlinedField from '../ui/OutlinedField.svelte';
	import OutlinedSelect from '../ui/OutlinedSelect.svelte';
	import Switch from '#lib/shared/mui/Switch.svelte';

	let { meta, subKey, onChange }: { meta: any; subKey: string; onChange: (m: any) => void } = $props();
	let json = $state<any>({});
	$effect(() => {
		void subKey;
		const m = meta;
		untrack(() => (json = clone(m.json || {})));
	});
	const diffs = $derived(diffList(meta.json, json));
	const d = $derived(json?.[subKey] || {});
	// 문자 발신번호가 등록됐는가(서버가 관리자 설정에 실어 준다 — 번호는 안 온다). 없으면 SMS 를 고를 수 없다
	const set = (k: string, v: any, isDesign = true) => (json = isDesign ? { ...json, [subKey]: { ...(json[subKey] || {}), [k]: v } } : { ...json, [k]: v });
	async function save() {
		const o = await api.put(`/service/admin/meta/${meta.metaKey}${meta.hostingId ? `?hostingId=${meta.hostingId}` : ''}`, json);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		onChange({ ...meta, json: clone(json) });
	}

	// 칸·브랜드 카드 설명(ⓘ 말풍선 — 공용 settings/ui/InfoTip)
	const TIPS: Record<string, { title: string; lines: string[] }> = {
		logo: {
			title: '로고',
			lines: ['사이트 머리(헤더) 왼쪽에 나와요.', '가로로 긴 PNG·SVG, 투명 바탕을 권해요.', '높이에 맞춰 줄어드니 여백 없이 잘라 올려 주세요.', '「로고 링크주소」 = 로고를 눌렀을 때 갈 곳(보통 /).']
		},
		favicon: { title: '파비콘', lines: ['브라우저 탭·즐겨찾기·홈 화면 아이콘이에요.', '정사각형 300×300 이상(PNG·SVG·ICO).', '작게 보이니 글자보다 모양 위주가 좋아요.'] },
		og: {
			title: 'OG 이미지',
			lines: [
				'카톡·페이스북·검색에 링크를 붙였을 때 나오는 미리보기 그림이에요.',
				'1200×630(가로 1.91:1) 권장, 1MB 이하.',
				'가장자리는 잘릴 수 있어 중요한 글자는 가운데에.',
				'페이지마다 따로 정하지 않으면 이 그림을 써요.'
			]
		},
		appbar: { title: '앱바 바탕색', lines: ['설치 앱(PWA)과 휴대폰 브라우저의 맨 위 막대 색이에요.', '사이트 머리 색과 맞추면 자연스러워요.', '비워 두면 브라우저 기본색.'] },
		'페이지 타이틀': { title: '페이지 타이틀', lines: ['브라우저 탭과 검색 결과 제목에 나오는 사이트 기본 이름이에요.', '페이지마다 SEO 를 따로 정하면 그 값이 먼저예요.'] },
		'SEO 키워드': { title: 'SEO 키워드', lines: ['쉼표로 구분한 대표 낱말(예: 일기, 라이프로그).', '요즘 검색엔진은 거의 보지 않아 짧게만 두면 돼요.'] },
		'SEO 사이트 설명': { title: 'SEO 사이트 설명', lines: ['검색 결과 제목 아래 나오는 한두 문장이에요.', '80~150자 안에서 무엇을 하는 곳인지 분명하게.'] },
		'robots.txt 설정': {
			title: 'robots.txt',
			lines: ['검색엔진에게 읽어도 되는 곳·안 되는 곳을 알려 주는 글이에요.', '주소: 도메인/robots.txt', '잘못 쓰면 사이트가 검색에서 빠질 수 있어요 — 모르면 그대로 두세요.']
		},
		'사이트맵 XML': { title: '사이트맵', lines: ['검색엔진에 알려 줄 공개 페이지 목록이에요(자동으로 만들어져요).', '구글 서치 콘솔·네이버 서치어드바이저에 이 주소를 등록하세요.'] },
		테마: { title: '테마', lines: ['사이트 기본 색 묶음이에요(버튼·강조색 등).', '바꾸면 저장 뒤 사용자 화면 전체 색이 바뀌어요.'] },
		홈페이지: { title: '홈페이지 주소', lines: ['사이트의 대표 주소 — 「도메인」 탭의 첫 사용자 주소로 정해져요.', '메일·알림·공유 링크를 만들 때 이 주소를 써요.'] },
		'사용자 로그인 유지 기간': { title: '회원 로그인 유지 기간', lines: ['회원이 이 기간 동안 접속하지 않으면 로그아웃돼요.', '접속할 때마다 다시 이 기간으로 늘어나요. 비우면 30일.'] }
	};
	// 사이트 공개(즈비드 호스팅 사이트에서만 보인다 — 서버 /admin/site-open 이 hosted 로 알려 준다, 2026-10-08)
	let siteOpen = $state<{ hosted: boolean; open: boolean } | null>(null);
	$effect(() => {
		if (subKey != 'design') return;
		api.get('/service/admin/site-open').then((o: any) => o.ok && (siteOpen = o.data));
	});
	async function setSiteOpen(open: boolean) {
		const o: any = await api.put('/service/admin/site-open', { open });
		if (!o.ok) {
			siteOpen = siteOpen && { ...siteOpen }; // 스위치를 원래대로
			return alert(o.error?.message || '저장 도중 오류가 발생하였습니다.');
		}
		siteOpen = o.data;
		// 이 화면이 들고 있는 SITE 값도 맞춘다 — 안 그러면 다른 칸을 저장할 때 옛 값으로 되돌린다
		const fix = (j: any) => {
			const n = { ...j };
			if (open) delete n.siteClosed;
			else n.siteClosed = true;
			return n;
		};
		json = fix(json);
		onChange({ ...meta, json: fix(meta.json || {}) });
		toast(open ? '사이트를 공개했습니다.' : '사이트를 비공개로 바꿨습니다.');
	}
	const homepage = siteHomepage();
	const hostingId = admin.profile?.hostingId || '';
	async function copy(t: string) {
		try {
			await navigator.clipboard.writeText(t);
		} catch {}
		toast('주소가 복사되었습니다.');
	}
	const dark = 'color-mix(in srgb, var(--theme-color-dark) 80%, transparent)';
	const img = (folder: string, name?: string, cid?: string) => `${adminImagePrefix(cid || hostingIdOf(json))}/${folder}/${name}`;
	const field = { size: 'small' as const, fullWidth: true };
</script>

{#snippet row(name: string, body: import('svelte').Snippet)}
	<SettingRow label={name} tip={TIPS[name]}>{@render body()}</SettingRow>
{/snippet}
<!-- 묶음 — 제목·한 줄 설명·칸들(2026-10-04 정리) -->
{#snippet group(title: string, desc: string, body: import('svelte').Snippet, wrap = false)}
	<SettingGroup {title} {desc} {wrap}>{@render body()}</SettingGroup>
{/snippet}

<div class="MuiStack-root">
	{#snippet title()}<OutlinedField {...field} value={d.title || ''} oninput={(v) => set('title', v)} />{/snippet}
	{#snippet keyword()}<OutlinedField {...field} value={d.keyword || ''} oninput={(v) => set('keyword', v)} />{/snippet}
	{#snippet desc()}<OutlinedField {...field} value={d.description || ''} oninput={(v) => set('description', v)} />{/snippet}
	{#snippet robots()}<OutlinedField {...field} multiline minRows={3} value={d.robots || ''} oninput={(v) => set('robots', v)} />{/snippet}
	{#snippet sitemap()}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
			<p class="MuiTypography-root MuiTypography-body2" style="word-break:break-all">{homepage}sitemap.xml</p>
			<Button size="small" onclick={() => copy(`${homepage}sitemap.xml`)}>주소복사</Button>
		</div>
	{/snippet}
	{#snippet theme()}
		<OutlinedSelect
			fullWidth
			size="small"
			value={d.theme || ''}
			placeholder="테마를 선택하세요"
			options={[{ value: '', label: '테마를 선택하세요', disabled: true }, ...Object.keys(ADMIN_THEMES).map((k) => ({ value: k, label: k.toUpperCase() }))]}
			onchange={(v) => set('theme', v)}
		>
			{#snippet item(o)}{#if o.value}<div class={toClass({ background: ADMIN_THEMES[o.value]?.main, width: 25, height: 25, marginRight: 8 })}></div>{/if}{o.label}{/snippet}
		</OutlinedSelect>
	{/snippet}
	<!-- 홈 주소는 「도메인」 탭의 대표 사용자 주소로 정해진다(2026-10-06 — 주소는 한 곳에서) -->
	{#snippet hp()}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
			<p class="MuiTypography-root MuiTypography-body2" style="word-break:break-all">{json?.homepage || '-'}</p>
			<Button size="small" href="/admin/setting/system/DOMAIN">도메인에서 바꾸기</Button>
		</div>
	{/snippet}
	{#snippet session()}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
			<OutlinedField
				size="small"
				type="number"
				sx={{ width: 120 }}
				value={json.userSessionDays ?? ''}
				placeholder="30"
				oninput={(v) => set('userSessionDays', String(v).trim() === '' ? undefined : Math.max(1, Math.min(400, Math.floor(Number(v) || 0))) || undefined, false)}
			/>
			<span class="MuiTypography-root MuiTypography-body2">일</span>
			<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-700, #909090)' }))}
				>로그인한 회원이 이 기간 동안 접속하지 않으면 로그아웃됩니다. 접속할 때마다 다시 이 기간으로 늘어납니다. 비워 두면 30일(최대 400일).</span
			>
		</div>
	{/snippet}
	<!-- 브랜드 · 공유 이미지(2026-10-04 정리) — 카드 4장, 넓으면 한 줄·좁으면(휴대폰) 한 칸씩. 미리보기를 누르면 이미지 고르기 -->
	{#snippet brandPreview(src: string, h: number, w: string)}
		<div class="bp" style="height:{h}px;background:{dark}">
			{#if src}<img {src} alt="" style="max-width:{w};max-height:{h - 16}px" />{:else}<span class="bp-empty">이미지 선택</span>{/if}
		</div>
	{/snippet}
	{#snippet brand()}
		<div class="brand">
			<div class="bc">
				<InfoTip label="로고" tip={TIPS['logo']} />
				<p class="MuiTypography-root MuiTypography-caption bd">헤더 왼쪽 · 가로형 PNG/SVG(투명 바탕 권장)</p>
				<ImagePicker
					size={{ width: 720, height: 156 }}
					image={d.logo}
					folder="site"
					useUnsplash
					useOriginalFileName
					disableCrop
					useDelete
					onSelect={(i) => set('logo', { name: i.name, hostingId: json?.hostingId })}
				>
					{#snippet after()}{@render brandPreview(d.logo?.name ? img('site', d.logo?.name, d.logo?.hostingId) : '', 96, '90%')}{/snippet}
				</ImagePicker>
				<OutlinedField size="small" fullWidth label="로고 링크주소" placeholder="ex) /" value={d.logoHref || ''} oninput={(v) => set('logoHref', v)} />
			</div>
			<div class="bc">
				<InfoTip label="파비콘" tip={TIPS['favicon']} />
				<p class="MuiTypography-root MuiTypography-caption bd">브라우저 탭·홈 화면 아이콘 · 정사각형(300×300 이상)</p>
				<ImagePicker
					size={{ width: 300, height: 300 }}
					image={d.favicon}
					folder="site"
					accept=".jpg, .jpeg, .png, .webp, .gif, .svg, .ico"
					useUnsplash
					useOriginalFileName
					disableCrop
					useDelete
					onSelect={(i) => set('favicon', { name: i.name })}
				>
					{#snippet after()}{@render brandPreview(d.favicon?.name ? img('site', d.favicon?.name) : '', 96, '80px')}{/snippet}
				</ImagePicker>
			</div>
			<div class="bc">
				<InfoTip label="OG 이미지" tip={TIPS['og']} />
				<p class="MuiTypography-root MuiTypography-caption bd">카톡·SNS 링크 미리보기 · 1200×630 권장</p>
				<ImagePicker
					size={{ width: 800, height: 400 }}
					image={d.image}
					folder="site"
					accept=".jpg, .jpeg, .png, .webp, .gif, .svg, .ico"
					useUnsplash
					useOriginalFileName
					disableCrop
					useDelete
					onSelect={(i) => set('image', { name: i.name })}
				>
					{#snippet after()}{@render brandPreview(d.image?.name ? img('site', d.image?.name) : '', 120, '100%')}{/snippet}
				</ImagePicker>
			</div>
			<div class="bc">
				<InfoTip label="앱바 바탕색" tip={TIPS['appbar']} />
				<p class="MuiTypography-root MuiTypography-caption bd">설치 앱(PWA)·휴대폰 브라우저 위쪽 막대 색</p>
				<div class="bp" style="height:96px;background:{d.appBarBgColor || 'var(--gt-200, #f2f2f2)'}">
					{#if !d.appBarBgColor}<span class="bp-empty" style="color:#909090">기본색</span>{/if}
				</div>
				<ColorField value={d.appBarBgColor || ''} sx={{ background: 'var(--zc-fff, #fff)', height: 40 }} onChange={(v) => set('appBarBgColor', v || '')} />
			</div>
		</div>
	{/snippet}
	{#snippet gSeo()}
		{@render row('페이지 타이틀', title)}
		{@render row('SEO 키워드', keyword)}
		{@render row('SEO 사이트 설명', desc)}
		{#if subKey == 'design'}{@render row('robots.txt 설정', robots)}{/if}
		{#if subKey == 'design'}{@render row('사이트맵 XML', sitemap)}{/if}
	{/snippet}
	{#snippet gSite()}
		{@render row('테마', theme)}
		{#if subKey == 'design'}{@render row('홈페이지', hp)}{/if}
	{/snippet}
	{#snippet gMember()}
		{@render row('사용자 로그인 유지 기간', session)}
	{/snippet}
	{#snippet gBrand()}
		{@render brand()}
	{/snippet}
	{#snippet gOpen()}
		<SettingRow label="사이트 공개" tip={{ title: '사이트 공개', lines: ['끄면 방문자에게는 「준비 중」 화면만 보여요.', '검색엔진에도 노출되지 않아요.', '관리자 화면은 그대로 쓸 수 있어요.'] }}>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }))}>
				<Switch checked={!!siteOpen?.open} label={siteOpen?.open ? '공개 중' : '비공개(준비 중)'} onchange={(v: boolean) => setSiteOpen(v)} />
				<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ color: 'var(--gt-700, #909090)' }))}>바로 반영돼요(저장 버튼 없이).</span>
			</div>
		</SettingRow>
	{/snippet}
	<div class="sgrps">
		{#if siteOpen?.hosted}{@render group('사이트 공개', '방문자에게 사이트를 보일지 — 즈비드 호스팅', gOpen)}{/if}
		{@render group('검색엔진 · SEO', '검색 결과에 어떻게 보일지, 검색엔진이 무엇을 읽을지', gSeo)}
		{@render group('사이트 주소 · 테마', '대표 주소와 기본 색', gSite, true)}
		{#if subKey == 'design'}{@render group('회원 로그인', '로그인 유지 기간', gMember)}{/if}
		{@render group('브랜드 · 공유 이미지', '로고·아이콘·링크 미리보기 그림', gBrand)}
	</div>

	<AuthButton variant="contained" sx={{ width: 'fit-content' }} disabled={!diffs.length} onclick={save}>저장</AuthButton>
	{#if diffs.length && admin.profile?.isDeveloper}<div class={cx('MuiStack-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>{/if}
</div>

<style>
	.sgrps {
		display: flex;
		flex-direction: column;
		gap: 20px;
		margin-bottom: 24px;
	}

	.brand {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 12px;
	}
	.bc {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
		border: 1px solid var(--adm-divider, #eaecf0);
		border-radius: 12px;
		background: var(--adm-paper, #fff);
		min-width: 0;
	}
	.bd {
		margin: -4px 0 0;
		color: var(--gt-700, #909090);
		word-break: keep-all;
	}
	.bp {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		border-radius: 8px;
		overflow: hidden;
		cursor: pointer;
	}
	.bp img {
		display: block;
		object-fit: contain;
	}
	.bp-empty {
		font-size: 13px;
		color: #fff;
		opacity: 0.85;
	}
</style>
