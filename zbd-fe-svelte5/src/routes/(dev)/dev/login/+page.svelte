<script lang="ts">
	/**
	 * 개발 로그인 — localhost 에서만(+page.ts).
	 *
	 * 운영 로그인은 HttpOnly 쿠키라 화면이 토큰을 읽을 수 없고, 소셜 로그인은 운영 주소로만 돌아온다.
	 *  - 관리자: localhost 의 /admin → 관리자 로그인 화면에서 그대로 로그인(개발 서버는 토큰을 본문으로 받는다).
	 *    또는 backyard 「관리자 회원 관리」 의 「로그인(DEV)」 이 복사해 주는 주소 `…/dev/login#admin=토큰` 을 열면 그 관리자로 들어간다
	 *    (토큰은 # 뒤라 서버로 가지 않는다. 브라우저에 묶이지 않는 토큰 — 발급은 admin_token_logs 에 남는다).
	 *  - 회원: dev.zbd.kr 에서 zbd.kr 로그인을 넘겨받는다(talk-handoff). 아래 「묶음 붙여 넣기」 는 예전 방식(ZBDDEV:…) 그대로.
	 * 운영 토큰이 이 맥의 localhost 저장소에 남는다 — 끝나면 「지우기」.
	 */
	import { page } from '$app/state';
	import { TOKEN_KEYS, BROWSER_UID_COOKIE, type Realm } from '#lib/shared/api.ts';
	import { SK } from '#lib/shared/storage-keys.ts';
	import { getCookie, setCookie, removeCookie } from '#lib/shared/cookie.ts';

	const PREFIX = 'ZBDDEV:';
	// 「로그인(DEV)」 주소 — #admin=토큰 을 받아 관리자 토큰으로 넣고 /admin 으로(주소에서 바로 지운다)
	if (typeof location != 'undefined' && location.hash.startsWith('#admin=')) {
		const tok = decodeURIComponent(location.hash.slice(7));
		history.replaceState(null, '', location.pathname);
		if (tok) {
			localStorage.setItem(TOKEN_KEYS.admin, tok);
			location.replace('/admin');
		}
	}

	let text = $state('');
	let msg = $state('');
	const back = page.url.searchParams.get('back') || '/';
	let now = $state(snapshot());

	function snapshot() {
		if (typeof localStorage == 'undefined') return { user: false, admin: false, uid: '' };
		return {
			user: !!localStorage.getItem(TOKEN_KEYS.user),
			admin: !!localStorage.getItem(TOKEN_KEYS.admin),
			uid: getCookie(BROWSER_UID_COOKIE)
		};
	}

	function apply() {
		msg = '';
		const s = text.trim();
		if (!s.startsWith(PREFIX)) {
			msg = `${PREFIX} 로 시작하는 묶음을 붙여 넣으세요`;
			return;
		}
		let o: { r: Realm; t: string; u: string };
		try {
			o = JSON.parse(decodeURIComponent(escape(atob(s.slice(PREFIX.length)))));
		} catch {
			msg = '묶음을 읽지 못했습니다 — 다시 복사하세요';
			return;
		}
		if (!o.t || (o.r != 'user' && o.r != 'admin')) {
			msg = '토큰이 비어 있습니다 — 운영 화면에서 먼저 로그인하세요';
			return;
		}
		localStorage.setItem(TOKEN_KEYS[o.r], o.t);
		// 회원 토큰은 발급받은 브라우저 id 와 헤더 zbduid 가 같아야 받는다
		if (o.u) setCookie(BROWSER_UID_COOKIE, o.u, { days: 30 });
		location.href = back.startsWith('/') && !back.startsWith('//') ? back : '/';
	}

	function clear() {
		localStorage.removeItem(TOKEN_KEYS.user);
		localStorage.removeItem(TOKEN_KEYS.admin);
		removeCookie(BROWSER_UID_COOKIE);
		now = snapshot();
		msg = '지웠습니다';
	}
</script>

<svelte:head><title>개발 로그인 · localhost</title></svelte:head>

<main>
	<h1>개발 로그인 <small>localhost 전용</small></h1>
	<p class="now">
		지금: 회원 토큰 <b>{now.user ? '있음' : '없음'}</b> · 관리자 토큰 <b>{now.admin ? '있음' : '없음'}</b> · 브라우저 id <b>{now.uid ? '있음' : '없음'}</b>
	</p>

	<h2>관리자</h2>
	<p>
		<a href="/admin">/admin</a> 에서 이메일·비밀번호로 그대로 로그인하거나, backyard 「관리자 회원 관리」 의 <b>로그인(DEV)</b> 복사 주소를 이 브라우저 주소창에 붙여 넣는다.
	</p>
	<h2>회원</h2>
	<p>
		소셜 로그인은 localhost 로 돌아오지 않는다 — <a href="https://dev.zbd.kr" target="_blank" rel="noreferrer">dev.zbd.kr</a> 에서 zbd.kr 로그인을 넘겨받는다. 예전 묶음(ZBDDEV:…)이 있으면 여기 붙여 넣는다.
	</p>
	<textarea bind:value={text} rows="3" placeholder="ZBDDEV:…"></textarea>
	<button type="button" class="primary" onclick={apply}>적용</button>
	<button type="button" onclick={clear}>지우기</button>
	{#if msg}<p class="msg">{msg}</p>{/if}
	<p class="note">
		API 는 개발 프록시(ZBD_API, 기본 운영)로 가므로 <b>운영 데이터에 그대로 쓰인다</b>. 쓰기 시험은 조심하거나
		<code>ZBD_API=http://127.0.0.1:4529</code>(맥미니 시험 DB) 로 띄운다 — 그때는 시험 DB 용 토큰(~/zbd-port/tokens)을 같은 모양으로 넣는다.
	</p>
</main>

<style>
	main {
		max-width: 760px;
		margin: 40px auto;
		padding: 0 16px;
		line-height: 1.7;
		font-family: var(--zbd-font-ui, sans-serif);
	}
	small {
		font-size: 14px;
		color: #888;
		font-weight: 400;
	}
	h2 {
		font-size: 17px;
		margin: 24px 0 4px;
	}
	textarea {
		display: block;
		width: 100%;
		box-sizing: border-box;
		margin: 6px 0;
		font-family: monospace;
	}
	button {
		margin-right: 6px;
		padding: 4px 12px;
		cursor: pointer;
	}
	.primary {
		background: #1d433d;
		color: #fff;
		border: 0;
		border-radius: 4px;
	}
	.msg {
		color: #c0392b;
	}
	.note {
		font-size: 13px;
		color: #666;
	}
</style>
