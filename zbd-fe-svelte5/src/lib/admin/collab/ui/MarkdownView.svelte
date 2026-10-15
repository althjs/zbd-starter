<script lang="ts">
	/**
	 * 마크다운 보기(React ReactMarkdown + MarkdownComponents) — 링크 규칙도 같게:
	 *   chrome:// 등 보호 스킴 = 누르면 주소 복사 + 안내 · 상대경로 = 같은 화면에서 이동(지금 쿼리 이어 줌) · 그 밖 = 새 탭.
	 * 바깥 여백(p·h)은 브라우저 기본(.zbd-raw) — 부품 기본값(여백 0)을 되돌린다.
	 */
	import { goto } from '$app/navigation';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import { alert } from '../../ui/modal.svelte.ts';
	import { markdownToHtml } from './md.ts';

	let { text, sx = null, class: cls = '' }: { text: string; sx?: Sx | null; class?: string } = $props();
	const html = $derived(markdownToHtml(text));
	const RESTRICTED = /^(chrome|chrome-extension|edge|brave|opera|about|file):/i;
	const isRelative = (u: string) => !!u && !/^([a-z][a-z0-9+.-]*:|\/|#)/i.test(u);

	async function click(e: MouseEvent) {
		const a = (e.target as HTMLElement).closest('a');
		if (!a) return;
		const url = a.getAttribute('href') || '';
		if (RESTRICTED.test(url)) {
			e.preventDefault();
			try {
				await navigator.clipboard?.writeText(url);
			} catch {}
			await alert({ title: '주소가 클립보드에 복사되었습니다', body: `브라우저 보안 정책상 클릭 이동이 차단되는 주소입니다.\n주소창에 붙여넣어 (Cmd/Ctrl+V) 이동해주세요.\n\n${url}` });
			return;
		}
		if (isRelative(url)) {
			if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
			e.preventDefault();
			const abs = new URL(url, location.href);
			if (!abs.search) abs.search = location.search;
			goto(abs.pathname + abs.search + abs.hash);
			return;
		}
	}
	// 바깥 링크는 새 탭
	function links(node: HTMLElement) {
		const fix = () =>
			node.querySelectorAll('a[href]').forEach((a) => {
				const u = a.getAttribute('href') || '';
				if (!RESTRICTED.test(u) && !isRelative(u) && !u.startsWith('#')) {
					a.setAttribute('target', '_blank');
					a.setAttribute('rel', 'noopener noreferrer');
				}
			});
		fix();
		const mo = new MutationObserver(fix);
		mo.observe(node, { childList: true, subtree: true });
		return { destroy: () => mo.disconnect() };
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class={cx('zbd-raw adm-md', cls, toClass(sx))} use:links onclick={click}>{@html html}</div>
