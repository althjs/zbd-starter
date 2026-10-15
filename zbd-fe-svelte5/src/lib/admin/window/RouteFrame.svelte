<script lang="ts">
	/**
	 * 경로 화면 창의 내용 — 같은 출처 iframe(`경로?popup=1`, 메뉴 없는 틀).
	 * SvelteKit 은 다른 라우트 부품을 page.params 와 함께 끼워 그리기 어려워(화면들이 $app/state page 를 읽는다) iframe 으로 띄운다.
	 * 관리자 화면끼리라 같은 출처로 둔다(회원이 만든 앱의 격리 iframe 규칙과는 다른 경우).
	 * - 창 안에서 링크를 누르면 그 창 안에서 옮겨 간다(`data-zbd-win` 이 있으면 ?popup=1 이 빠져도 메뉴 없는 틀 — frame.ts).
	 * - 창 안을 누르면 맨 앞(iframe 안의 누름은 부모로 오지 않아 그 문서에 직접 붙인다).
	 * - 제목: 처음 경로면 연 쪽이 준 이름, 옮겨 가면 그 주소의 메뉴 이름(없으면 문서 제목).
	 */
	import { untrack } from 'svelte';
	import { admin } from '../profile.svelte.ts';
	import type { AdminMenu } from '#lib/shared/types/admin.ts';
	import { bringToFront, setTitle } from './windows.svelte.ts';

	let { id, path, title }: { id: string; path: string; title: string } = $props();

	/** 처음 열 때 받은 이름(제목은 옮겨 가면 바뀐다) */
	const firstTitle = untrack(() => title);
	const src = $derived(`${path}${path.includes('?') ? '&' : '?'}popup=1`);
	const basePath = $derived(path.split('?')[0]);
	let frame = $state<HTMLIFrameElement>();

	/** 주소 → 메뉴 이름(React syncCurrentNav 와 같은 순서 — 마지막에 맞은 것) */
	function menuTitle(p: string): string {
		let found = '';
		const walk = (ms: AdminMenu[] | null | undefined) => {
			for (const m of ms || []) {
				if (m.items?.length) {
					for (const s of m.items) if (s.path && (s.path == p || p.indexOf(s.path) === 0)) found = s.name;
				} else if (m.path && p.indexOf(m.path) === 0) found = m.name;
			}
		};
		walk(admin.menus);
		return found;
	}

	function attach() {
		const w = frame?.contentWindow;
		if (!w) return;
		try {
			w.addEventListener('pointerdown', () => bringToFront(id), true);
		} catch {
			/* 다른 출처로 나갔으면 못 붙인다 */
		}
	}

	// 창 안에서 옮겨 간 주소에 맞춰 제목(SvelteKit 이동은 load 이벤트가 없어 잠깐씩 본다)
	$effect(() => {
		const f = frame;
		if (!f) return;
		let last = '';
		const t = setInterval(() => {
			try {
				const p = f.contentWindow?.location.pathname || '';
				if (!p || p === last) return;
				last = p;
				setTitle(id, p === basePath ? firstTitle : menuTitle(p) || f.contentDocument?.title || p);
			} catch {
				/* 다른 출처 */
			}
		}, 700);
		return () => clearInterval(t);
	});
</script>

<iframe bind:this={frame} {src} title={firstTitle} data-zbd-win={id} onload={attach} style="display:block;width:100%;height:100%;border:0;background:transparent"></iframe>
