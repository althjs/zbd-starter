<script lang="ts">
	/** 1:1 문의 상세 — 예전 pages/user/my/inquiry/[postId].tsx */
	import { page } from '$app/state';
	import MyInquiryView from '#lib/user/inquiry/MyInquiryView.svelte';
</script>

<MyInquiryView postId={page.params.postId || ''} />
