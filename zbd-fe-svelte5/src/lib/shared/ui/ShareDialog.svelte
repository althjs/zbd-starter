<script lang="ts">
	/**
	 * 공유하기 창 — 페이스북·카카오톡·링크 복사(예전 ShareDialogue).
	 * 카카오는 사이트 설정(siteMeta.auth.KAKAO.JAVASCRIPT_KEY) 이 있을 때만 SDK 를 붙인다.
	 */
	import type { Snippet } from 'svelte';
	import { loadSite, siteMeta } from '#lib/shared/site.svelte.ts';
	import { toast } from '#lib/shared/ui/dialog.svelte.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import { sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';

	let {
		url,
		title = '',
		desc = '',
		imageUrl = '',
		isMobile = false,
		asLink = false,
		linkSx = null,
		children
	}: {
		url: string;
		title?: string;
		desc?: string;
		imageUrl?: string;
		isMobile?: boolean;
		/** 게시글 머리의 공유 아이콘처럼 MUI Link 모양(.MuiLink-root.share-icon) 으로 — 운영자 CSS 가 그 이름을 겨냥한다 */
		asLink?: boolean;
		linkSx?: Sx | null;
		children?: Snippet;
	} = $props();

	let open = $state(false);
	let kakaoReady = $state(typeof window != 'undefined' && !!window.Kakao?.isInitialized?.());

	const KAKAO_SDK = 'https://t1.kakaocdn.net/kakao_js_sdk/2.4.0/kakao.min.js';
	const KAKAO_SRI = 'sha384-mXVrIX2T/Kszp6Z0aEWaA8Nm7J6/ZeWXbL8UpGRjKwWe56Srd/iyNmWMBhcItAjH';

	$effect(() => {
		if (!open || kakaoReady) return;
		loadSite().then(() => {
			const key = siteMeta()?.auth?.KAKAO?.JAVASCRIPT_KEY;
			if (!key) return;
			if (window.Kakao) {
				if (!window.Kakao.isInitialized()) window.Kakao.init(key);
				kakaoReady = true;
				return;
			}
			if (document.getElementById('kakao-js')) return;
			const js = document.createElement('script');
			js.id = 'kakao-js';
			js.crossOrigin = 'anonymous';
			js.integrity = KAKAO_SRI;
			js.src = KAKAO_SDK;
			js.onload = () => {
				window.Kakao.init(key);
				kakaoReady = true;
			};
			document.body.appendChild(js);
		});
	});

	function meta(p: string) {
		return document.querySelector(`meta[property="${p}"]`)?.getAttribute('content') || '';
	}

	function share(id: string) {
		if (id == 'facebook') {
			window.open('http://www.facebook.com/sharer/sharer.php?u=' + url);
		} else if (id == 'kakaotalk') {
			window.Kakao?.Share.sendDefault({
				objectType: 'feed',
				content: {
					title: title || meta('og:title'),
					description: desc || meta('og:description'),
					imageUrl: imageUrl || meta('og:image'),
					link: { mobileWebUrl: url, webUrl: url }
				},
				buttons: [{ title: '바로가기', link: { mobileWebUrl: url, webUrl: url } }]
			});
		}
	}

	async function copy() {
		try {
			await navigator.clipboard.writeText(url);
		} catch (e) {
			const t = document.createElement('textarea');
			t.value = url;
			document.body.appendChild(t);
			t.select();
			document.execCommand('copy');
			t.remove();
		}
		toast('공유주소가 복사되었습니다.');
	}
</script>

{#if asLink}
	<a
		class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone share-icon {toClass({
			alignSelf: 'center',
			cursor: 'pointer',
			display: 'flex',
			alignItems: 'center',
			...(linkSx || {})
		})}"
		role="button"
		tabindex="0"
		onclick={() => (open = true)}
		onkeydown={(e) => e.key == 'Enter' && (open = true)}
	>
		{#if children}{@render children()}{:else}<Icon name="share" sx={{ fontSize: '22px', color: '#909090' }} />{/if}
	</a>
{:else}
	<span class="share-icon" role="button" tabindex="0" onclick={() => (open = true)} onkeydown={(e) => e.key == 'Enter' && (open = true)}>
		{@render children?.()}
	</span>
{/if}

{#if open}
	<div class="layer" use:portal>
		<div class="backdrop" role="presentation" onclick={() => (open = false)}></div>
		<div class="share" role="dialog" aria-modal="true">
			<h2 class:mobile={isMobile}>공유하기</h2>
			<div class="list">
				<img src="/images/social/facebook.svg" alt="페이스북" onclick={() => share('facebook')} />
				{#if kakaoReady}
					<img src="/images/social/kakaotalk.svg" alt="카카오톡" onclick={() => share('kakaotalk')} />
				{/if}
				<img src="/images/social/link.svg" alt="링크 복사" onclick={copy} />
			</div>
			<button class="cancel" onclick={() => (open = false)}>취소</button>
		</div>
	</div>
{/if}

<style>
	.layer {
		display: contents;
	}
	span.share-icon {
		align-self: center;
		cursor: pointer;
		display: flex;
		align-items: center;
	}
	.backdrop {
		position: fixed;
		inset: 0;
		background: var(--zbd-backdrop);
		z-index: var(--zbd-z-modal);
	}
	.share {
		position: fixed;
		z-index: 1301;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		box-sizing: content-box;
		width: 100vw;
		max-width: 280px;
		background: var(--zbd-paper);
		border: 2px solid #000;
		box-shadow: var(--zbd-shadow-dialog);
		padding: 32px;
	}
	/* 예전 MUI Typography h4(PC) / h5(휴대폰) */
	h2 {
		text-align: center;
		font-family: var(--zbd-font-ui);
		font-size: 34px;
		font-weight: 400;
		line-height: 1.235;
		letter-spacing: 0.00735em;
	}
	h2.mobile {
		font-size: 24px;
		line-height: 1.334;
		letter-spacing: 0;
	}
	.list {
		display: flex;
		justify-content: space-around;
		margin: 30px auto 20px;
		width: 80%;
	}
	/* 예전 Stack spacing — 간격을 gap 이 아니라 왼쪽 여백으로 줘야 space-around 배치가 같다 */
	.list img + img {
		margin-left: 8px;
	}
	.list img {
		width: 50px;
		height: 50px;
		cursor: pointer;
	}
	/* 예전 MUI Button text·large·fullWidth */
	.cancel {
		display: block;
		box-sizing: border-box;
		width: 100%;
		border: 0;
		background: none;
		font-family: var(--zbd-font-ui);
		color: var(--zbd-text-muted);
		font-weight: 700;
		font-size: 15px;
		line-height: 1.75;
		letter-spacing: var(--zbd-ls-button);
		padding: 8px 11px;
		border-radius: var(--zbd-radius-sm);
		cursor: pointer;
	}
	.cancel:hover {
		background: var(--zbd-hover);
	}
</style>
