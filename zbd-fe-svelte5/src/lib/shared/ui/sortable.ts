/**
 * 끌어서 순서 바꾸기 — 관리자·사용자 공용 액션 하나(2026-10-04 대표 — 「드래그할 때 움직이는 게 없어 어색하다」, 같은 일을 하던 세 벌을 합쳤다).
 *   <div use:sortable={{ handle: '.handle', onmove: (from, to) => (list = moveItem(list, from, to)) }}>
 *     {#each list as o (o.id)}<div>…<span class="handle">⋮⋮</span>…</div>{/each}
 *   </div>
 * 끄는 줄이 손가락·마우스를 따라 움직이고, 지나가는 줄들이 자리를 비켜 준다(SortableJS 처럼). 놓으면 onmove(옛 자리, 새 자리) — 배열은 부르는 쪽이 바꾼다.
 * 대상 = 액션을 붙인 상자의 **바로 아래 자식들**(세로 목록). 마우스·터치 모두(포인터 이벤트), 다른 라이브러리 없이.
 * 모양은 인라인으로 잠깐 주고 놓으면 지운다(그림자·위로 띄움) — 따로 CSS 를 두지 않아도 된다.
 */
import { tick } from 'svelte';

export type SortableOpts = { handle: string; onmove: (from: number, to: number) => void; disabled?: boolean };

const DUR = '160ms';

export function sortable(node: HTMLElement, opts: SortableOpts) {
	let o = opts;
	let item: HTMLElement | null = null;
	let kids: HTMLElement[] = [];
	let centers: number[] = [];
	let from = -1;
	let to = -1;
	let startY = 0;
	let step = 0;
	let pid = -1;
	let saved = { position: '', zIndex: '', boxShadow: '', transition: '', cursor: '' };

	function down(e: PointerEvent) {
		if (o.disabled || item || e.button > 0) return;
		const h = (e.target as HTMLElement).closest(o.handle);
		if (!h || !node.contains(h)) return;
		kids = [...node.children] as HTMLElement[];
		const it = kids.find((k) => k.contains(h));
		if (!it) return;
		e.preventDefault();
		item = it;
		from = to = kids.indexOf(it);
		startY = e.clientY;
		pid = e.pointerId;
		const rects = kids.map((k) => k.getBoundingClientRect());
		centers = rects.map((r) => r.top + r.height / 2);
		// 비켜 줄 거리 = 끄는 줄 높이 + 줄 사이 간격(gap·margin)
		const r = rects[from];
		const gap = rects[from + 1] ? rects[from + 1].top - r.bottom : rects[from - 1] ? r.top - rects[from - 1].bottom : 0;
		step = r.height + Math.max(0, gap);
		for (const k of kids) if (k !== it) k.style.transition = `transform ${DUR} ease`;
		saved = { position: it.style.position, zIndex: it.style.zIndex, boxShadow: it.style.boxShadow, transition: it.style.transition, cursor: document.body.style.cursor };
		if (getComputedStyle(it).position == 'static') it.style.position = 'relative';
		it.style.zIndex = '20';
		it.style.transition = 'box-shadow 120ms ease';
		it.style.boxShadow = '0 10px 28px rgba(0,0,0,0.18)';
		document.body.style.cursor = 'grabbing';
		document.body.style.userSelect = 'none';
		window.addEventListener('pointermove', move);
		window.addEventListener('pointerup', up);
		window.addEventListener('pointercancel', up);
	}

	function move(e: PointerEvent) {
		if (!item || e.pointerId != pid) return;
		const dy = e.clientY - startY;
		item.style.transform = `translateY(${dy}px)`;
		// 끄는 줄의 가운데가 다른 줄의 가운데를 넘으면 그 자리로
		const c = centers[from] + dy;
		let t = from;
		for (let i = from + 1; i < kids.length; i++) if (c > centers[i]) t = i;
		for (let i = from - 1; i >= 0; i--) if (c < centers[i]) t = i;
		if (t == to) return;
		to = t;
		kids.forEach((k, i) => {
			if (k === item) return;
			const s = from < to && i > from && i <= to ? -step : from > to && i >= to && i < from ? step : 0;
			k.style.transform = s ? `translateY(${s}px)` : '';
		});
	}

	async function up(e: PointerEvent) {
		if (!item || e.pointerId != pid) return;
		window.removeEventListener('pointermove', move);
		window.removeEventListener('pointerup', up);
		window.removeEventListener('pointercancel', up);
		const it = item;
		const ks = kids;
		const f = from;
		const t = to;
		item = null;
		for (const k of ks) k.style.transition = 'none';
		// 배열을 바꾸고 화면이 새 순서로 그려진 뒤(같은 틱 안이라 깜빡임 없음) 비켜 두었던 자리를 지운다
		if (f >= 0 && t >= 0 && f != t) {
			o.onmove(f, t);
			await tick();
		}
		for (const k of ks) {
			k.style.transform = '';
			k.style.transition = '';
		}
		it.style.position = saved.position;
		it.style.zIndex = saved.zIndex;
		it.style.boxShadow = saved.boxShadow;
		it.style.transition = saved.transition;
		document.body.style.cursor = saved.cursor;
		document.body.style.userSelect = '';
		from = to = -1;
	}

	node.addEventListener('pointerdown', down);
	return {
		update(n: SortableOpts) {
			o = n;
		},
		destroy() {
			node.removeEventListener('pointerdown', down);
			window.removeEventListener('pointermove', move);
			window.removeEventListener('pointerup', up);
			window.removeEventListener('pointercancel', up);
		}
	};
}

/** 배열에서 하나를 옮긴 새 배열(원본은 그대로) */
export function moveItem<T>(list: T[], from: number, to: number): T[] {
	const next = [...list];
	const [x] = next.splice(from, 1);
	next.splice(to, 0, x);
	return next;
}
