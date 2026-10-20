<script lang="ts">
	/**
	 * 환경 설정 › 도메인(2026-10-06 대표 「하나로만」) — 사용자 주소 · 관리자 주소를 SITE 설정 domains 한 곳에.
	 * 홈 주소 · 쿠키 범위 · 사이트맵 주소는 서버가 이 두 목록으로 계산해 같이 저장한다(GET/PUT /service/admin/domain — Go core/domain.go).
	 * nginx·인증서는 서버 권한이 필요해 저장만으로는 안 바뀐다 — 저장 뒤 `zbd deploy` 가 DB 주소대로 맞춘다.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { adminCookieParent } from '#lib/shared/config.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import TagInput from '../../ui/TagInput.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import SettingGroup from '../ui/SettingGroup.svelte';
	import SettingRow from '../ui/SettingRow.svelte';

	let { onSaved }: { onSaved?: () => void } = $props();
	type View = { user: string[]; admin: string[]; homepage: string; cookieHost: string; current: string };
	let saved = $state<View | null>(null);
	let user = $state('');
	let adminHosts = $state('');
	let busy = $state(false);

	function apply(v: View) {
		saved = v;
		user = (v.user || []).join(',');
		adminHosts = (v.admin || []).join(',');
	}
	onMount(async () => {
		const o = await api.get('/service/admin/domain');
		if (!o.ok) return alert(o.error?.message || '도메인 설정을 읽지 못했습니다.');
		apply(o.data);
	});
	const split = (s: string) =>
		s
			.split(',')
			.map((x) => x.trim())
			.filter(Boolean);
	const changed = $derived(!!saved && (user != saved.user.join(',') || adminHosts != saved.admin.join(',')));
	// 주소 칸은 호스트만 — https:// · 끝 / · 대문자는 넣을 때 정리
	const clean = (t: string) =>
		t
			.toLowerCase()
			.replace(/^https?:\/\//, '')
			.replace(/\/+$/, '');

	async function save() {
		if (!(await confirm('도메인을 저장할까요?\n\n새 주소는 DNS 연결 뒤 `zbd deploy` 를 한 번 해야 열립니다(nginx·인증서).'))) return;
		busy = true;
		const o = await api.put('/service/admin/domain', { user: split(user), admin: split(adminHosts) });
		busy = false;
		if (!o.ok) return alert(o.error?.message || '저장하지 못했습니다.');
		apply(o.data);
		toast('저장했습니다.');
		onSaved?.();
	}
	// 관리자 쿠키 범위 — 서버(web/session.go cookieScope)와 같은 규칙: 사용자 쿠키 범위 안의 주소면 그 범위를 같이 쓰고,
	// 밖이면 그 주소 전용(Domain 없음), 개발 주소(localhost 등)는 쿠키 없이 헤더 토큰
	const isLocal = (h: string) => /^(localhost|127\.|\[::1\])/.test(h) || h.split(':')[0].endsWith('.localhost');
	function adminCookieScope(h: string, cookieHost: string, admins: string[]) {
		const host = h.split(':')[0];
		const d = cookieHost.replace(/^\./, '');
		if (d && (host == d || host.endsWith('.' + d))) return `${h} → .${d}(사용자와 같은 범위)`;
		if (isLocal(host)) return `${h} → 없음(개발 주소 — 쿠키 없이 로그인)`;
		const p = adminCookieParent(host, admins);
		if (p && p != host) return `${h} → .${p}(관리자 ${p} 와 같은 범위)`;
		if (p && admins.some((x) => x.split(':')[0] != p && x.split(':')[0].endsWith('.' + p))) return `${h} → .${p}(아래 앱 주소와 같은 범위)`;
		return `${h} → 이 주소 전용`;
	}
	const TIPS = {
		user: { title: '사용자 주소', lines: ['방문자가 들어오는 홈페이지 주소예요.', '첫 번째가 대표 주소 — 나머지 주소로 오면 대표 주소로 넘겨요.', '예: example.com, www.example.com'] },
		admin: {
			title: '관리자 주소',
			lines: ['관리자 화면은 이 주소에서만 열려요.', '첫 번째가 대표 관리자 주소예요.', '지금 접속한 주소는 뺄 수 없어요(잠김 방지) — 새 주소로 접속해 확인한 뒤 빼 주세요.']
		},
		calc: {
			title: '자동으로 정해지는 값',
			lines: [
				'두 주소 목록에서 계산해 같이 저장해요 — 따로 고치지 않아요.',
				'쿠키 범위 = 로그인 쿠키가 따라가는 주소. 사용자 쿠키는 대표 사용자 주소와 그 아래 주소에, 관리자 쿠키는 그 범위 안이면 같이, 밖이면 그 관리자 주소에만 심어요.'
			]
		}
	};
</script>

<div class="dm">
	<SettingGroup title="도메인" desc="사용자 주소 · 관리자 주소 — 사이트 주소는 여기 한 곳에서 정해요" fold={false}>
		{#if saved}
			<SettingRow label="사용자 주소" tip={TIPS.user}>
				<TagInput tags={user} onChange={(v) => (user = v)} inputFilter={clean} placeholder="example.com — 엔터로 추가" />
			</SettingRow>
			<SettingRow label="관리자 주소" tip={TIPS.admin}>
				<TagInput tags={adminHosts} onChange={(v) => (adminHosts = v)} inputFilter={clean} placeholder="admin.example.com — 엔터로 추가" />
			</SettingRow>
			<SettingRow label="자동으로 정해지는 값" tip={TIPS.calc}>
				<dl class="calc">
					<dt>홈 주소</dt>
					<dd>{saved.homepage || '-'}</dd>
					<dt>사용자 쿠키 범위</dt>
					<dd>{saved.cookieHost || '없음(개발 주소 — 쿠키 없이 로그인)'}</dd>
					<dt>관리자 쿠키 범위</dt>
					<dd>
						{#each saved.admin || [] as h (h)}<div>{adminCookieScope(h, saved.cookieHost || '', saved.admin || [])}</div>{:else}-{/each}
					</dd>
					<dt>사이트맵</dt>
					<dd>{saved.homepage ? `${saved.homepage}sitemap.xml` : '-'}</dd>
					<dt>지금 접속한 주소</dt>
					<dd>{typeof location != 'undefined' ? location.host : saved.current}</dd>
				</dl>
			</SettingRow>
			<p class="MuiTypography-root MuiTypography-caption note">
				사용자·관리자에 같은 주소를 넣으면 「한 주소」 로 열려요(개발용 localhost 만). 새 주소는 DNS 를 서버로 연결한 뒤 <code>zbd deploy</code> 를 한 번 하면 nginx·인증서가 맞춰져요.
			</p>
		{/if}
	</SettingGroup>
	<AuthButton variant="contained" sx={{ width: 'fit-content' }} disabled={!changed || busy} onclick={save}>저장</AuthButton>
</div>

<style>
	.dm {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.calc {
		display: grid;
		grid-template-columns: max-content 1fr;
		gap: 6px 16px;
		margin: 0;
		font-size: 14px;
	}
	.calc dt {
		color: var(--gt-700, #909090);
	}
	.calc dd {
		margin: 0;
		word-break: break-all;
	}
	.note {
		color: var(--gt-700, #909090);
	}
</style>
