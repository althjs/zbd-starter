/**
 * 선택 목록 커서 — 펼친 목록(`ul[role=listbox]`)에 붙이면 「지금 가리키는 줄」 하나를 둔다(2026-10-16 대표 「select 가 커서로 선택되게」).
 *   <ul role="listbox" use:menuNav> <li class="MuiMenuItem-root …">…</li> </ul>
 * - 열리면 고른 줄(.Mui-selected, 없으면 첫 줄)에 커서 · ↑↓ Home End 로 옮김 · Enter·Space = 그 줄 누르기 · 글자 = 그 글자로 시작하는 줄로
 * - 마우스를 올린 줄로 커서가 따라온다(키보드와 마우스가 한 커서) · 커서 줄 = 클래스 `Mui-focusVisible`(모양은 shared/styles/mui.css)
 * - 막힌 줄(.Mui-disabled · aria-disabled)은 건너뛴다. 줄의 Enter 처리는 여기서 하니 줄에 따로 onkeydown 을 달지 않는다.
 */
const ON = 'Mui-focusVisible';

export function menuNav(node: HTMLElement) {
	const items = () => [...node.querySelectorAll<HTMLElement>('.MuiMenuItem-root, [role=option]')].filter((el) => !el.classList.contains('Mui-disabled') && el.getAttribute('aria-disabled') !== 'true');
	let cur: HTMLElement | null = null;

	function set(el: HTMLElement | null, focus = true) {
		if (cur && cur !== el) cur.classList.remove(ON);
		cur = el;
		if (!el) return;
		el.classList.add(ON);
		if (focus) {
			el.focus({ preventScroll: true });
			el.scrollIntoView({ block: 'nearest' });
		}
	}
	function move(step: number | 'first' | 'last') {
		const list = items();
		if (!list.length) return;
		const i = cur ? list.indexOf(cur) : -1;
		const n = step === 'first' ? 0 : step === 'last' ? list.length - 1 : i < 0 ? (step > 0 ? 0 : list.length - 1) : (i + step + list.length) % list.length;
		set(list[n]);
	}
	function keydown(e: KeyboardEvent) {
		if (e.isComposing) return;
		if (e.key == 'ArrowDown' || e.key == 'ArrowUp') move(e.key == 'ArrowDown' ? 1 : -1);
		else if (e.key == 'Home' || e.key == 'End') move(e.key == 'Home' ? 'first' : 'last');
		else if (e.key == 'Enter' || e.key == ' ') cur?.click();
		else if (e.key.length == 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
			// 글자로 찾기 — 지금 줄 다음부터 그 글자로 시작하는 줄
			const list = items();
			const k = e.key.toLowerCase();
			const start = cur ? list.indexOf(cur) + 1 : 0;
			const hit = [...list.slice(start), ...list.slice(0, start)].find((el) => (el.textContent || '').trim().toLowerCase().startsWith(k));
			if (hit) set(hit);
			else return;
		} else return;
		e.preventDefault();
		e.stopPropagation();
	}
	function over(e: MouseEvent) {
		const el = (e.target as HTMLElement).closest<HTMLElement>('.MuiMenuItem-root, [role=option]');
		if (el && node.contains(el) && el !== cur && items().includes(el)) set(el, false);
	}

	node.addEventListener('keydown', keydown);
	node.addEventListener('mousemove', over);
	// 목록 위치를 잡은 뒤(다음 그림) 고른 줄에 커서
	const raf = requestAnimationFrame(() => {
		const list = items();
		set(list.find((el) => el.classList.contains('Mui-selected')) || list[0] || null);
	});
	return {
		destroy() {
			cancelAnimationFrame(raf);
			node.removeEventListener('keydown', keydown);
			node.removeEventListener('mousemove', over);
		}
	};
}
