<script lang="ts">
	/**
	 * 관리자 첫 화면 — 회원 현황 · 업무 상태 · 게시판별 글 수 · 최근 글·회원·업무.
	 * `GET /service/admin/dashboard` 한 번. 업무 상태 이름·색은 업무 보드(TASK) 설정(공통 코드)에서.
	 */
	import { api } from '#lib/shared/api.ts';
	import { date, number } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Icon from '../ui/Icon.svelte';
	import AppInstallCard from './AppInstallCard.svelte';
	import { getWorkStatusCodes } from '../collab/work/codes.ts';
	import type { DashboardData } from '#lib/shared/types/admin.ts';

	let data = $state<DashboardData | null>(null);
	let error = $state('');
	$effect(() => {
		api.get('/service/admin/dashboard').then((o) => {
			if (o.ok) data = o.data;
			else error = o.error?.message || '대시보드를 불러오지 못했어요.';
		});
	});

	const statusCodes = $derived(getWorkStatusCodes('TASK').codes);
	const statusOf = (v: number) => statusCodes.find((c) => Number(c.id) == Number(v));
	const workTotal = $derived((data?.work.byStatus || []).reduce((a, b) => a + b.count, 0));
	const ago = (ms: number) => {
		if (!ms) return '';
		const m = Math.floor((Date.now() - ms) / 60000);
		if (m < 1) return '방금';
		if (m < 60) return `${m}분 전`;
		if (m < 1440) return `${Math.floor(m / 60)}시간 전`;
		return date(ms);
	};
</script>

{#snippet stat(icon: string, color: string, label: string, value: number, href: string, sub = '')}
	<a class="dash-stat" {href}>
		<span class="dash-ic" style:background={color}><Icon name={icon} sx={{ fontSize: 26, color: '#fff' }} /></span>
		<span class="dash-sv">
			<span class="dash-sl">{label}</span>
			<b>{number(value)}</b>
			{#if sub}<span class="dash-ss">{sub}</span>{/if}
		</span>
	</a>
{/snippet}

{#snippet card(title: string, href: string, body: import('svelte').Snippet)}
	<section class="dash-card">
		<header>
			<h6>{title}</h6>
			{#if href}<a {href}>전체 보기 ›</a>{/if}
		</header>
		{@render body()}
	</section>
{/snippet}

{#if error}
	<p class="dash-err">{error}</p>
{:else if data}
	{@const m = data.memberSummary}
	<div class="dash">
		<div class="dash-stats">
			{@render stat('person_add', '#4338ca', '오늘 가입', m.todayCount, '/admin/user/member')}
			{@render stat('group', '#0f9d58', '전체 회원', m.totalCount - m.leaveCount, '/admin/user/member', m.leaveRequestCount ? `탈퇴 신청 ${number(m.leaveRequestCount)}명` : '')}
			{@render stat('view_kanban', '#1c7ed6', '업무 카드', workTotal, '/admin/work/TASK')}
		</div>

		<AppInstallCard />

		<div class="dash-grid">
			{#snippet workBody()}
				{#if data!.work.byStatus.length}
					<div class="dash-bar">
						{#each data!.work.byStatus as s (s.status)}
							{@const c = statusOf(s.status)}
							<span title={`${c?.name || s.status} ${s.count}`} style:flex={s.count} style:background={c?.bgColor || '#ced4da'}></span>
						{/each}
					</div>
					<div class="dash-legend">
						{#each data!.work.byStatus as s (s.status)}
							{@const c = statusOf(s.status)}
							<span><i style:background={c?.bgColor || '#ced4da'}></i>{c?.name || '미설정'} {s.count}</span>
						{/each}
					</div>
				{/if}
				<ul class="dash-list">
					{#each data!.work.recent as w (w.workId)}
						{@const c = statusOf(w.status)}
						<li>
							<a href={`/admin/work/TASK/${w.workId}`}>
								<span class="dash-chip" style:background={c?.bgColor || '#ced4da'} style:color={c?.color || '#fff'}>{c?.name || '미설정'}</span>
								<span class="dash-t">{w.title || `#${w.workId}`}</span>
								<span class="dash-m">{ago(w.updatedAt)}</span>
							</a>
						</li>
					{:else}
						<li class="dash-empty">아직 업무 카드가 없어요 — 업무관리에서 「추가」</li>
					{/each}
				</ul>
			{/snippet}
			{@render card('업무', '/admin/work/TASK', workBody)}

			{#snippet postsBody()}
				<ul class="dash-list">
					{#each data!.recentPosts as p (p.postId)}
						<li>
							<a href={`/admin/board/${p.groupKey}/${p.postId}`}>
								<span class="dash-t">{p.title}</span>
								<span class="dash-m">{p.jsonUser?.name || ''} · {date(p.createdAt)}</span>
							</a>
						</li>
					{:else}
						<li class="dash-empty">회원이 쓴 글이 아직 없어요</li>
					{/each}
				</ul>
			{/snippet}
			{@render card('최근 게시글', '', postsBody)}

			{#snippet boardsBody()}
				<ul class="dash-list">
					{#each data!.boards as b (b.groupKey)}
						<li>
							<a href={`/admin/board/${b.groupKey}`}>
								<span class="dash-chip ghost">{b.kind == 'admin' ? '관리자' : '사용자'}</span>
								<span class="dash-t">{b.name || b.groupKey}</span>
								<span class="dash-m"
									>{number(b.count)}개{#if b.week}&nbsp;· 이번 주 +{number(b.week)}{/if}</span
								>
							</a>
						</li>
					{:else}
						<li class="dash-empty">게시판이 없어요 — 설정 › 환경 설정 › 게시판 관리</li>
					{/each}
				</ul>
			{/snippet}
			{@render card('게시판', '/admin/setting/system/POST_GROUPS', boardsBody)}

			{#snippet membersBody()}
				<ul class="dash-list">
					{#each data!.recentMembers as u (u.userId)}
						<li>
							<a href={`/admin/user/member/${u.userId}`}>
								<span class="dash-t">{u.nickname || `회원 ${u.userId}`}</span>
								<span class="dash-m">{u.status == 99 ? '탈퇴 신청 · ' : ''}{date(u.createdAt)} 가입</span>
							</a>
						</li>
					{:else}
						<li class="dash-empty">아직 가입한 회원이 없어요</li>
					{/each}
				</ul>
			{/snippet}
			{@render card('최근 가입 회원', '/admin/user/member', membersBody)}
		</div>
	</div>
{/if}

<style>
	.dash {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.dash-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 16px;
	}
	.dash-stat,
	.dash-card {
		background: var(--zc-fff, #fff);
		border-radius: 14px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
		color: inherit;
		text-decoration: none;
	}
	.dash-stat {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 18px;
	}
	.dash-ic {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		border-radius: 12px;
		flex-shrink: 0;
	}
	.dash-sv {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.dash-sl {
		font-size: 13px;
		color: var(--gt-700, #868e96);
	}
	.dash-sv b {
		font-size: 26px;
		line-height: 1.2;
	}
	.dash-ss {
		font-size: 12px;
		color: #e03131;
	}
	.dash-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
	}
	.dash-card {
		padding: 18px 20px;
		min-width: 0;
	}
	.dash-card header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 10px;
	}
	.dash-card h6 {
		font-size: 16px;
		font-weight: 700;
		margin: 0;
	}
	.dash-card header a {
		font-size: 13px;
		color: var(--theme-color-main);
		text-decoration: none;
	}
	.dash-list {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.dash-list li a {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 0;
		border-bottom: 1px solid var(--gt-200, #f1f3f5);
		color: inherit;
		text-decoration: none;
		font-size: 14px;
	}
	.dash-list li:last-child a {
		border-bottom: 0;
	}
	.dash-t {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.dash-m {
		flex-shrink: 0;
		font-size: 12px;
		color: var(--gt-700, #868e96);
	}
	.dash-chip {
		flex-shrink: 0;
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 999px;
	}
	.dash-chip.ghost {
		background: var(--gt-200, #f1f3f5);
		color: var(--gt-700, #495057);
	}
	.dash-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #e8590c;
		flex-shrink: 0;
	}
	.dash-empty {
		padding: 16px 0;
		font-size: 14px;
		color: var(--gt-700, #868e96);
	}
	.dash-bar {
		display: flex;
		height: 10px;
		border-radius: 999px;
		overflow: hidden;
		margin: 4px 0 8px;
		gap: 2px;
	}
	.dash-legend {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 12px;
		font-size: 12px;
		color: var(--gt-700, #495057);
		margin-bottom: 6px;
	}
	.dash-legend i {
		display: inline-block;
		width: 8px;
		height: 8px;
		border-radius: 2px;
		margin-right: 4px;
	}
	.dash-err {
		color: #e03131;
	}
	@media (max-width: 1199px) {
		.dash-stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
	@media (max-width: 899px) {
		.dash-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
