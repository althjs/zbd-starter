<script lang="ts">
	/** 사이트 바닥(고정) — 사업자 정보(환경 설정의 siteMeta) · 약관 · 저작권. 고치려면 이 파일을 */
	import { siteMetaOf } from '#lib/user/site.svelte.ts';

	const meta = $derived(siteMetaOf() || {});
	const year = new Date().getFullYear();
</script>

<footer class="zs-foot">
	<div class="zs-wrap">
		<nav>
			<a href="/board/notice">공지사항</a>
			<a href="/board/free">자유게시판</a>
			<a href="/cs">고객센터</a>
		</nav>
		{#if meta.companyName}
			<p>
				{meta.companyName}{#if meta.ceoName}&nbsp;· 대표 {meta.ceoName}{/if}{#if meta.companyNo}&nbsp;· 사업자등록번호 {meta.companyNo}{/if}{#if meta.companyTel}&nbsp;· {meta.companyTel}{/if}
			</p>
		{/if}
		<p class="zs-copy">© {year} {meta.companyName || meta.title || ''}</p>
	</div>
</footer>

<style>
	.zs-foot {
		margin-top: 80px;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
		color: #6b7280;
		font-size: 13px;
	}
	.zs-wrap {
		max-width: 1200px;
		margin: 0 auto;
		padding: 28px 20px 40px;
	}
	nav {
		display: flex;
		gap: 16px;
		margin-bottom: 10px;
	}
	a {
		color: #374151;
		text-decoration: none;
	}
	p {
		margin: 4px 0;
	}
</style>
