<script lang="ts">
	/**
	 * 운영자 HTML 보기 — 예전 components/common/html/HtmlView.tsx.
	 * 운영자(관리자) 가 쓴 HTML 이라 그대로 넣는다(React 판도 같다). **회원이 쓴 글은 여기로 넣지 말 것** — DOMPurify 를 거친다.
	 *
	 * 바꿔 넣는 말: `{사이트메타키}`·`{메타.하위키}`(회사명·전화 등) · `{회원이름}`(닉네임 → 이름)
	 * 모양 규칙은 app.css 의 `.html-view`(이미지 블록·폭 100%) — 운영자 CSS 와 같은 구체성으로.
	 */
	import { cx, sx as toClass, type Sx } from '../mui/sx';
	import { site } from '../site.svelte.ts';

	let {
		html,
		siteMeta = null,
		memberName = '',
		class: cls = '',
		sx = null,
		el = $bindable()
	}: { html: string; siteMeta?: any; memberName?: string; class?: string; sx?: Sx | null; el?: HTMLElement } = $props();

	function fill(tag: string, key: string, o: any) {
		for (const sub of Object.keys(o || {})) tag = tag.replaceAll(`{${key}.${sub}}`, o?.[sub] || '');
		return tag;
	}

	// siteMeta 를 주지 않으면 사이트 설정에서 읽는다(React 판은 저장소에서 늘 읽었다 — 안 주면 빈 화면이 됐다)
	const meta = $derived(siteMeta || site.base?.siteMeta || null);
	const real = $derived.by(() => {
		const siteMeta = meta;
		if (!siteMeta || !html) return '';
		let tag = html;
		for (const key of Object.keys(siteMeta)) {
			const v = siteMeta[key] || '미등록';
			if (typeof v != 'string') {
				if (typeof v == 'object') tag = fill(tag, key, v);
				else continue;
			}
			tag = tag.replaceAll(`{${key}}`, v);
		}
		return tag.replaceAll('{회원이름}', memberName || '');
	});
</script>

<div bind:this={el} class={cx('html-view', cls, 'MuiBox-root', toClass(sx))}>{@html real}</div>
