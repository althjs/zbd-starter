<script lang="ts">
	/**
	 * 홈(/) — 고정 화면. 스타터에는 페이지 디자인이 없으니 이 파일을 고쳐 내 홈을 만든다(Claude 에게 「홈에 ○○ 넣어 줘」).
	 */
	import { session } from '#lib/user/session.svelte.ts';
	import { siteMetaOf } from '#lib/user/site.svelte.ts';

	const siteName = $derived(siteMetaOf()?.title || '내 사이트');

	// 개발 모드에서만 — .dev/account.txt 의 계정(개발 서버의 /__dev/account, 이 컴퓨터에서만 열린다)
	const DEV = import.meta.env.DEV;
	let shown = $state<Record<string, { email: string; password: string } | string>>({});
	async function reveal(who: 'admin' | 'user') {
		if (shown[who]) return (shown = { ...shown, [who]: '' });
		const r = await fetch(`/__dev/account?who=${who}`);
		shown = { ...shown, [who]: r.ok ? await r.json() : await r.text() };
	}
</script>

<section class="zs-hero">
	<p class="zs-badge">zbd-starter</p>
	<h1>{siteName}가 준비됐어요</h1>
	<p class="zs-lead">이 화면은 <code>src/routes/(user)/(site)/+page.svelte</code> 예요. 이 파일을 고쳐 내 홈을 만들면 돼요.</p>
	<div class="zs-cta">
		<a class="zs-btn" href="/admin">관리자 화면</a>
		{#if !session.profile}<a class="zs-btn ghost" href="/user/join">회원가입</a>{/if}
	</div>
</section>

<section class="zs-steps">
	<h2>처음 할 일</h2>
	<ol>
		<li><b>관리자로 들어가기</b> — <a href="/admin">/admin</a>, 첫 계정은 <code>.dev/account.txt</code></li>
		<li><b>비밀번호 바꾸기 · 지문/얼굴 로그인 등록</b> — 관리자 마이페이지</li>
		<li><b>사이트 이름 · 회사 정보</b> — 설정 › 환경 설정</li>
		<li><b>이 홈 바꾸기</b> — <code>src/routes/(user)/(site)/+page.svelte</code></li>
		<li><b>머리 메뉴 · 바닥글</b> — <code>src/lib/user/layout/SiteHeader.svelte</code> · <code>SiteFooter.svelte</code></li>
	</ol>
</section>

{#if DEV}
	<section class="zs-steps zs-dev">
		<h2>개발용 계정</h2>
		<p class="zs-lead">개발 모드에서만 보여요(배포한 사이트엔 없어요). 값은 <code>.dev/account.txt</code>.</p>
		<div class="zs-cta">
			<button class="zs-btn ghost" onclick={() => reveal('admin')}>{shown.admin ? '관리자 암호 숨기기' : '관리자 암호 보기'}</button>
			<button class="zs-btn ghost" onclick={() => reveal('user')}>{shown.user ? '사용자 암호 숨기기' : '사용자 암호 보기'}</button>
		</div>
		{#each [['admin', '관리자', '/admin'], ['user', '사용자', '/login']] as [k, label, href] (k)}
			{@const v = shown[k]}
			{#if v}
				<p class="zs-acc">
					<b>{label}</b>
					{#if typeof v == 'string'}{v}{:else}<a {href}>{href}</a> · <code>{v.email}</code> · <code>{v.password}</code>{/if}
				</p>
			{/if}
		{/each}
	</section>
{/if}

<style>
	.zs-hero,
	.zs-steps {
		max-width: 760px;
		margin: 0 auto;
		padding: 0 20px;
		color: #1f2937;
		line-height: 1.7;
	}
	.zs-hero {
		padding-top: 88px;
		padding-bottom: 40px;
	}
	.zs-badge {
		display: inline-block;
		padding: 2px 10px;
		border-radius: 999px;
		background: var(--theme-color-alpha12);
		color: var(--theme-color-main);
		font-size: 13px;
		font-weight: 700;
	}
	h1 {
		font-size: 36px;
		font-weight: 800;
		margin: 10px 0 12px;
		line-height: 1.3;
	}
	h2 {
		font-size: 20px;
		font-weight: 800;
		margin-bottom: 8px;
	}
	.zs-lead {
		color: #4b5563;
	}
	.zs-cta {
		display: flex;
		gap: 10px;
		margin-top: 24px;
	}
	.zs-btn {
		padding: 10px 18px;
		border-radius: 10px;
		background: var(--theme-color-main);
		color: var(--theme-color-contrastText);
		text-decoration: none;
		font-weight: 700;
	}
	button.zs-btn {
		cursor: pointer;
		font: inherit;
		font-weight: 700;
	}
	.zs-dev {
		margin-top: 40px;
		padding-top: 24px;
		border-top: 1px dashed #d0d5dd;
	}
	.zs-acc {
		margin: 12px 0 0;
		font-size: 14px;
	}
	.zs-btn.ghost {
		background: #fff;
		color: var(--theme-color-main);
		border: 1px solid var(--theme-color-alpha30);
	}
	ol {
		padding-left: 20px;
	}
	li {
		margin: 8px 0;
	}
	code {
		background: #f3f4f6;
		padding: 1px 6px;
		border-radius: 4px;
		font-size: 14px;
	}
	a {
		color: var(--theme-color-main);
	}
	@media (max-width: 599px) {
		.zs-hero {
			padding-top: 48px;
		}
		h1 {
			font-size: 28px;
		}
	}
</style>
