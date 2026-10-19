<script lang="ts">
	/**
	 * 대시보드 「웹앱 설치」 바로가기(2026-10-16 대표) — 이 사이트에 있는 설치 앱 중 아직 안 깐 것만.
	 *  · 관리자 앱: 브라우저가 지금 주소에서 설치 신호(beforeinstallprompt — 설치 앱 플러그인이 app.html 에서 먼저 받아 둔다)를 줄 때 → 바로 설치 창
	 *  · 단독 앱(메신저 chat. · 웹메일 mail. · 웹오피스 office. · 포토샵 photo.): 그 주소가 관리자 주소 목록에 있고 내 메뉴에 그 화면이 있을 때
	 *    → 새 창 `?install=1`(설치 창은 그 주소에서만 뜬다 — 그 앱이 설치 안내를 연다)
	 * 이미 설치한 앱은 숨긴다 — 앱이 설치 창으로 열릴 때 남기는 표시(app-installed.ts). 다 깔았거나 설치할 것이 없으면 카드 자체가 없다.
	 */
	import { ADMIN_APP_NAME, SINGLE_APP, SINGLE_APPS } from '../config.ts';
	import { checkAuth } from '../profile.svelte.ts';
	import { installedApps, isStandalone, markInstalled } from '../app-installed.ts';
	import Icon from '../ui/Icon.svelte';

	type Item = { key: string; name: string; desc: string; icon: string; color: string; run: () => void };
	const APPS: Record<string, { desc: string; icon: string; color: string }> = {
		admin: { desc: '관리자 화면을 앱으로', icon: 'admin_panel_settings', color: '#173b37' },
		messenger: { desc: '메뉴 없이 메신저만', icon: 'forum', color: '#31726c' },
		mail: { desc: '메뉴 없이 웹메일만', icon: 'mail', color: '#2f5d8a' },
		office: { desc: '문서·시트·슬라이드', icon: 'description', color: '#5b4bc4' },
		photo: { desc: '사진 편집', icon: 'photo_filter', color: '#1b3a6b' }
	};

	let prompt = $state<any>(typeof window != 'undefined' ? (window as any).__zbdInstallPrompt || null : null);
	let installed = $state(installedApps());
	$effect(() => {
		const ready = () => (prompt = (window as any).__zbdInstallPrompt || null);
		const done = () => {
			prompt = null;
			installed = installedApps();
		};
		addEventListener('zbd:installable', ready);
		addEventListener('zbd:installed', done);
		addEventListener('focus', done);
		return () => {
			removeEventListener('zbd:installable', ready);
			removeEventListener('zbd:installed', done);
			removeEventListener('focus', done);
		};
	});

	const items = $derived.by(() => {
		const out: Item[] = [];
		if (!SINGLE_APP && !isStandalone() && prompt && !installed.includes('admin'))
			out.push({
				key: 'admin',
				name: ADMIN_APP_NAME,
				...APPS.admin,
				run: async () => {
					const e = prompt;
					prompt = null;
					e.prompt();
					const r = await e.userChoice.catch(() => null);
					if (r?.outcome == 'accepted') {
						markInstalled('admin');
						installed = installedApps();
					}
				}
			});
		for (const a of Object.values(SINGLE_APPS)) {
			if (!a.host || !a.url || installed.includes(a.key) || SINGLE_APP?.key == a.key || !checkAuth(a.home)) continue;
			out.push({ key: a.key, name: a.name, ...APPS[a.key], run: () => window.open(a.url + '?install=1', '_blank') });
		}
		return out;
	});
</script>

{#if items.length}
	<section class="dash-apps">
		<header>
			<h6>웹앱 설치</h6>
			<span>홈 화면·작업 표시줄에서 바로 열어요</span>
		</header>
		<div class="dash-apps-list">
			{#each items as it (it.key)}
				<button type="button" class="dash-app" onclick={it.run}>
					<span class="dash-app-ic" style:background={it.color}><Icon name={it.icon} sx={{ fontSize: 22, color: '#fff' }} /></span>
					<span class="dash-app-tx"><b>{it.name}</b><small>{it.desc}</small></span>
					<Icon name="download" sx={{ fontSize: 20, color: 'var(--zc-888, #888)' }} />
				</button>
			{/each}
		</div>
	</section>
{/if}

<style>
	.dash-apps {
		background: var(--zc-fff, #fff);
		border-radius: 14px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
		padding: 16px 20px;
	}
	.dash-apps header {
		display: flex;
		align-items: baseline;
		gap: 10px;
		margin-bottom: 10px;
	}
	.dash-apps h6 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}
	.dash-apps header span {
		font-size: 12px;
		color: var(--zc-888, #888);
	}
	.dash-apps-list {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 10px;
	}
	.dash-app {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border: 1px solid var(--zc-eee, #eee);
		border-radius: 12px;
		background: none;
		color: inherit;
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.dash-app:hover {
		background: var(--zc-f8f9fa, #f8f9fa);
	}
	.dash-app-ic {
		width: 38px;
		height: 38px;
		border-radius: 10px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex: none;
	}
	.dash-app-tx {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}
	.dash-app-tx small {
		font-size: 12px;
		color: var(--zc-888, #888);
	}
</style>
