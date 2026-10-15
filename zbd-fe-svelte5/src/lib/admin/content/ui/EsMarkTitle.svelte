<script lang="ts">
	/** 검색 엔진 결과의 <mark> 를 노란 강조로(React board/components/EsMarkTitle). 그 밖의 글자는 그대로 이스케이프 */
	let { title = '' }: { title?: string } = $props();
	const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	const html = $derived(
		esc(
			String(title || '')
				.replace(/&lt;/g, '<')
				.replace(/&gt;/g, '>')
				.replace(/&amp;/g, '&')
		)
			.replace(/&lt;mark&gt;/g, '<span style="background:#ffe066; color:#222; font-weight:bold; padding:0 2px; border-radius:2px;">')
			.replace(/&lt;\/mark&gt;/g, '</span>')
	);
</script>

<span>{@html html}</span>
