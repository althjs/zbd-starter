<script lang="ts">
	/** 로그아웃 주소 — 서버가 세션 쿠키를 지우고 홈으로(예전 pages/logout.tsx) */
	import { endSession } from '#lib/shared/api.ts';
	endSession('user').then(() => location.replace('/'));
</script>
