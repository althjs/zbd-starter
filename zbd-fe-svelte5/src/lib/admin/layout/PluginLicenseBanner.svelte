<script lang="ts">
	/**
	 * 즈비드 플러그인 라이선스 확인 배너 — 서버가 즈비드에 확인한 결과(GET /service/admin/plugin-license)에 문제가 있을 때만 맨 위에 보인다.
	 * 사이트는 멈추지 않는다. 붙인 플러그인이 없거나 아직 확인 전이면 아무것도 안 그린다.
	 */
	import { onMount } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { dark } from '../dark.svelte.ts';

	const STORE = 'https://zbd.kr/app/zbd-starter'; // 플러그인 소개 — ?plugin=키 면 그 플러그인 자세히 보기 창이 열린다

	let problems = $state<{ code: string; plugin?: string; message: string }[]>([]);

	onMount(async () => {
		const o = await api.get('/service/admin/plugin-license');
		if (o.ok) problems = ((o.data as any)?.problems || []) as typeof problems;
	});
</script>

{#if problems.length}
	<div class="plb" class:plb-dark={dark.resolved == 'dark'} role="alert">
		<strong>플러그인 라이선스를 확인하지 못했어요</strong>
		<ul>
			{#each problems as p, i (i)}
				<li>
					{p.message}
					{#if p.plugin}<a class="plb-more" href={`${STORE}?plugin=${encodeURIComponent(p.plugin)}`} target="_blank" rel="noopener">자세히 보기 ›</a>{/if}
				</li>
			{/each}
		</ul>
		<a href={STORE} target="_blank" rel="noopener">즈비드 플러그인 ›</a>
	</div>
{/if}

<style>
	.plb {
		margin: 0 0 16px;
		padding: 12px 16px;
		border: 1px solid #f0c36d;
		border-radius: 12px;
		background: #fff8e6;
		color: #6b4a00;
		font-size: 14px;
		line-height: 1.5;
	}
	.plb ul {
		margin: 4px 0 6px;
		padding-left: 18px;
	}
	.plb a {
		color: inherit;
		font-weight: 600;
	}
	.plb .plb-more {
		margin-left: 8px;
		font-size: 13px;
		white-space: nowrap;
	}
	.plb-dark {
		background: #3a2e12;
		border-color: #7a5c1d;
		color: #f5dfa8;
	}
</style>
