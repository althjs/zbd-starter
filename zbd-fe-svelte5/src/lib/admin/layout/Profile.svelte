<script lang="ts">
	/** 내 사진(→ /admin/my) + (왼쪽 메뉴에서는) 인사·로그아웃·앱 설치 (React layouts/dashboard/components/Profile) */
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { adminImageUrl } from '../meta.svelte.ts';
	import { admin, logout } from '../profile.svelte.ts';
	import { confirm } from '../ui/modal.svelte.ts';

	let { useLogout = false, sx = null }: { useLogout?: boolean; sx?: Sx | null } = $props();
	const p = $derived(admin.profile);
	const image = $derived(p?.image?.name ? adminImageUrl(p.image.imageId || `admin_profile_${p.adminId}`, p.image.name) : '');
	let failedSrc = $state(''); // 못 읽은 주소 — 사진을 바꾸면 다시 시도
	const failed = $derived(!!image && failedSrc === image);

	// 설치 신호(app.html 이 window.__zbdInstallPrompt 에 붙잡아 둔다) — 있을 때만 「앱으로 설치」
	let installable = $state(false);
	const standalone = () => window.matchMedia?.('(display-mode: standalone)').matches || (navigator as any).standalone === true;
	$effect(() => {
		const up = () => (installable = !!(window as any).__zbdInstallPrompt && !standalone());
		up();
		addEventListener('zbd:installable', up);
		addEventListener('zbd:installed', up);
		return () => {
			removeEventListener('zbd:installable', up);
			removeEventListener('zbd:installed', up);
		};
	});
	async function install() {
		const e = (window as any).__zbdInstallPrompt;
		if (!e) return;
		e.prompt();
		await e.userChoice.catch(() => null);
		(window as any).__zbdInstallPrompt = null;
		installable = false;
	}
</script>

<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', ...(sx || {}) }))}>
	<a class="MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineNone" href="/admin/my">
		<div class={cx('MuiAvatar-root MuiAvatar-circular', (!image || failed) && 'MuiAvatar-colorDefault', toClass({ cursor: 'pointer', height: 40, width: 40, marginRight: 8 }))}>
			{#if image && !failed}<img src={image} alt="" class="MuiAvatar-img" onerror={() => (failedSrc = image)} />{:else}{(p?.name || '').substring(0, 1)}{/if}
		</div>
	</a>
	{#if useLogout}
		<div class="MuiStack-root">
			<p class="MuiTypography-root MuiTypography-body2">
				<strong>{p?.name}</strong>님{#if (p?.name || '').length > 5},<br />{/if} 좋은하루 되세요
			</p>
			<!-- svelte-ignore a11y_missing_attribute -->
			<a
				class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ cursor: 'pointer' }))}
				role="button"
				tabindex="0"
				onclick={async () => (await confirm({ body: '로그아웃 하시겠습니까?' })) && logout()}
				onkeydown={() => {}}>Logout</a
			>
			{#if installable}
				<!-- svelte-ignore a11y_missing_attribute -->
				<a
					class={cx('MuiTypography-root MuiTypography-inherit MuiLink-root MuiLink-underlineHover', toClass({ cursor: 'pointer', fontSize: 13 }))}
					role="button"
					tabindex="0"
					onclick={install}
					onkeydown={() => {}}>앱으로 설치</a
				>
			{/if}
		</div>
	{/if}
</div>
