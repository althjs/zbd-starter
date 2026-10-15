<script lang="ts" module>
	export type DateRangeValue = { from: string; to?: string };
</script>

<script lang="ts">
	/**
	 * 기간 고르기 창 — React components/common/DateRangeModal(항공 시절부터 쓰던 여러 달 달력) 과 같은 동작·모양 + 편의 기능.
	 *   <DateRangeModal schedule={val} onchange={(s) => save(s)} useReset fromLabel="시작일" toLabel="종료일">
	 *     {#snippet template()}<span>일정 선택</span>{/snippet}        ← 누르면 창(React template)
	 *   </DateRangeModal>
	 *   template 없이 bind:open 으로 열어도 된다(DateRange 의 날짜 칸).
	 *
	 * 동작(원본 그대로): 첫 클릭 = 시작, 둘째 클릭 = 끝(앞 날짜를 누르면 서로 바꿈), 셋째 클릭 = 새 시작 · 요일 색 · 공휴일(COMMON.holydays) ·
	 *   오늘 칸 회색 · 지난 날 흐리게 · minYMD 앞은 못 고름 · 「날짜 초기화」(useReset, 확인 창) · 취소/적용 · 휴대폰 전체 화면.
	 * 더한 것: 위에 고른 기간(「10월 3일(금) ~ 10월 9일(목) · 7일」) · 빠른 선택 칩 · 사이 칸 칠하기 · 끝 날짜 미리 칠하기(마우스·키보드) ·
	 *   같은 날 두 번 = 하루 · 「당일」(시작만 고르고 바로 적용) · 키보드(←→↑↓ 하루·한 주, PageUp/Down 한 달, Enter 고르기, ⌘/Ctrl+Enter 적용, Esc 닫기) ·
	 *   직접 입력(YYYY-MM-DD · 20261003 · 2026.10.3) 과 달력이 서로 따라감 · ◀ ▶ 달 넘기기 · 지난 달도(pastMonths) — 열면 고른 달(없으면 이번 달)부터 보인다.
	 * 값: onchange({ from, to }) — 'YYYY-MM-DD'. 초기화는 onchange(null). isSingle 이면 누르는 즉시 onchange({ from }).
	 */
	import type { Snippet } from 'svelte';
	import { tick, untrack } from 'svelte';
	import { portal } from '#lib/shared/ui/portal.ts';
	import { getMeta } from '../meta.svelte.ts';
	import { media } from '../media.svelte.ts';
	import Button from './Button.svelte';
	import Icon from './Icon.svelte';
	import IconButton from './IconButton.svelte';
	import { confirm } from './modal.svelte.ts';

	let {
		schedule = null,
		onchange,
		template,
		open = $bindable(false),
		readonly = false,
		fromLabel = '가는날',
		toLabel = '오는날',
		minYMD = '',
		isSingle = false,
		useReset = false,
		resetLabel = '날짜 초기화',
		resetConfirm = true,
		quick = true,
		pastMonths = 12,
		months: monthCount = 13
	}: {
		schedule?: DateRangeValue | null;
		onchange: (s: DateRangeValue | null) => void;
		template?: Snippet;
		open?: boolean;
		readonly?: boolean;
		fromLabel?: string;
		toLabel?: string;
		minYMD?: string;
		isSingle?: boolean;
		useReset?: boolean;
		resetLabel?: string;
		/** 초기화 전에 묻기(React 와 같게 기본 켬) */
		resetConfirm?: boolean;
		/** 빠른 선택 칩 */
		quick?: boolean;
		/** 이번 달 앞으로 더 둘 달 수(위로 올리면 보인다) */
		pastMonths?: number;
		/** 이번 달부터 그릴 달 수(React 13) */
		months?: number;
	} = $props();

	// ── 날짜 도우미 ('YYYY-MM-DD' 문자열로만 다룬다) ──
	const pad = (n: number) => String(n).padStart(2, '0');
	const fmt = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
	const parse = (s: string) => new Date(Number(s.slice(0, 4)), Number(s.slice(5, 7)) - 1, Number(s.slice(8, 10)));
	const addDays = (s: string, n: number) => {
		const d = parse(s);
		d.setDate(d.getDate() + n);
		return fmt(d);
	};
	const addMonths = (ym: string, n: number) => {
		const d = new Date(Number(ym.slice(0, 4)), Number(ym.slice(5, 7)) - 1 + n, 1);
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}`;
	};
	const monthDiff = (a: string, b: string) => (Number(b.slice(0, 4)) - Number(a.slice(0, 4))) * 12 + Number(b.slice(5, 7)) - Number(a.slice(5, 7));
	const dayCount = (a: string, b: string) => Math.round((parse(b).getTime() - parse(a).getTime()) / 86400000) + 1;
	const WEEK = ['일', '월', '화', '수', '목', '금', '토'];
	const YOIL_COLOR: Record<number, string> = { 0: '#E24B1E', 6: '#005dff' };
	const valid = (s: string) => /^\d{4}-\d{2}-\d{2}$/.test(s) && fmt(parse(s)) == s;

	/** 직접 입력 → 'YYYY-MM-DD' (20261003 · 2026.10.3 · 2026-10-3 · 26/10/3) */
	function normalize(raw: string): string {
		const t = raw.trim();
		if (!t) return '';
		let y = '',
			m = '',
			d = '';
		if (/^\d{8}$/.test(t)) [y, m, d] = [t.slice(0, 4), t.slice(4, 6), t.slice(6, 8)];
		else {
			const p = t.split(/[^0-9]+/).filter(Boolean);
			if (p.length != 3) return '';
			[y, m, d] = p;
		}
		if (y.length == 2) y = `20${y}`;
		const s = `${y}-${pad(Number(m))}-${pad(Number(d))}`;
		return valid(s) ? s : '';
	}

	let today = $state(fmt(new Date()));
	const holydays = $derived(((getMeta('COMMON', 'holydays') || []) as { ymd: string; name: string }[]).reduce((m, o) => ((m[o.ymd] = o.name), m), {} as Record<string, string>));

	// ── 창 상태 ──
	let from = $state('');
	let to = $state('');
	let hover = $state(''); // 끝 날짜 미리 보기(마우스) — 키보드는 cursor
	let cursor = $state('');
	let kbd = $state(false); // 키보드로 움직였을 때만 커서 테두리
	let fromText = $state('');
	let toText = $state('');
	let fromBad = $state(false);
	let toBad = $state(false);
	let startYm = $state('');
	/** 열 때 맨 앞에 보이는 달 — React 는 이 달부터 그려 그 달 첫 주의 앞 달 날짜가 비어 있었다(같게) */
	let anchorYm = $state('');
	let count = $state(13);
	let box = $state<HTMLElement>();
	let scroller = $state<HTMLElement>();
	let innerH = $state(typeof window != 'undefined' ? window.innerHeight : 900);

	/** 한 줄에 몇 달(원본 폭 규칙 xs·sm 100% · md 48% · lg 49% · xl 32%) */
	const cols = $derived(media.mdDown ? 1 : media.xlUp ? 3 : 2);
	const monthList = $derived(Array.from({ length: count }, (_, i) => addMonths(startYm || today.slice(0, 7), i)));
	const monthSet = $derived(new Set(monthList));

	function weeksOf(ym: string): string[][] {
		const first = `${ym}-01`;
		const f = parse(first);
		let cur = addDays(first, -f.getDay());
		const last = fmt(new Date(f.getFullYear(), f.getMonth() + 1, 0));
		const end = addDays(last, 6 - parse(last).getDay());
		const rows: string[][] = [];
		while (cur <= end) {
			const w: string[] = [];
			for (let i = 0; i < 7; i++) w.push(addDays(cur, i));
			rows.push(w);
			cur = addDays(cur, 7);
		}
		return rows;
	}

	/** 열 때: 고른 값을 옮겨 오고, 고른 달(없으면 이번 달)이 줄 맨 앞에 오게 달 목록을 잡는다 */
	function init() {
		today = fmt(new Date());
		from = schedule?.from || '';
		to = isSingle ? '' : schedule?.to || '';
		hover = '';
		kbd = false;
		cursor = from || today;
		syncText();
		layoutMonths(from || today);
	}
	function layoutMonths(target: string, keep = false) {
		const tYm = target.slice(0, 7);
		if (!keep) anchorYm = tYm;
		const curYm = today.slice(0, 7);
		const c = cols;
		const past = Math.ceil(Math.max(pastMonths, monthDiff(tYm, curYm), 0) / c) * c;
		// 고른 달이 이번 달보다 앞이면 그 달이 줄 맨 앞에 오게(이번 달 기준으로 c 의 배수 만큼 앞)
		let start = addMonths(curYm, -past);
		if (monthDiff(start, tYm) % c) start = addMonths(start, monthDiff(start, tYm) % c);
		let n = Math.max(monthDiff(start, curYm) + monthCount, monthDiff(start, tYm) + monthCount);
		const toYm = (to || '').slice(0, 7);
		if (toYm && monthDiff(start, toYm) >= n) n = monthDiff(start, toYm) + 1;
		if (keep && startYm) {
			// 이미 그린 목록을 넓히기만(보던 자리 유지)
			const s = monthDiff(startYm, start) < 0 ? start : startYm;
			const e = addMonths(start, n - 1);
			const oldEnd = addMonths(startYm, count - 1);
			startYm = s;
			count = Math.max(monthDiff(s, e), monthDiff(s, oldEnd)) + 1;
			return;
		}
		startYm = start;
		count = n;
	}
	function syncText() {
		fromText = from;
		toText = to;
		fromBad = toBad = false;
	}

	async function show() {
		if (readonly) return;
		init();
		open = true;
	}
	function close() {
		open = false;
	}
	let wasOpen = false;
	$effect(() => {
		if (open && !wasOpen) {
			wasOpen = true;
			untrack(() => {
				if (!startYm || !template) init();
			});
			tick().then(() => {
				scrollToMonth((from || today).slice(0, 7), false);
				box?.focus({ preventScroll: true });
			});
		} else if (!open) wasOpen = false;
	});
	$effect(() => {
		if (!open) return;
		const prev = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		const onResize = () => (innerH = window.innerHeight);
		onResize();
		window.addEventListener('resize', onResize);
		return () => {
			document.body.style.overflow = prev;
			window.removeEventListener('resize', onResize);
		};
	});

	// ── 스크롤 ──
	function scrollToMonth(ym: string, smooth = true) {
		const el = scroller?.querySelector<HTMLElement>(`[data-ym="${ym}"]`);
		if (!el || !scroller) return;
		const top = el.offsetTop - 8;
		scroller.scrollTo({ top, behavior: smooth ? 'smooth' : 'auto' });
	}
	/** ◀ ▶ — 한 줄(PC 2~3달 · 휴대폰 한 달)씩 */
	function step(dir: 1 | -1) {
		if (!scroller) return;
		const cards = [...scroller.querySelectorAll<HTMLElement>('[data-ym]')];
		const tops = [...new Set(cards.map((c) => c.offsetTop - 8))].sort((a, b) => a - b);
		const now = scroller.scrollTop;
		const next = dir > 0 ? tops.find((t) => t > now + 4) : [...tops].reverse().find((t) => t < now - 4);
		if (next != null) scroller.scrollTo({ top: next, behavior: 'smooth' });
		else if (dir < 0) {
			// 맨 위 — 앞 달을 더 붙인다
			const h = scroller.scrollHeight;
			startYm = addMonths(startYm, -cols);
			count += cols;
			tick().then(() => scroller && (scroller.scrollTop += scroller.scrollHeight - h));
		}
	}
	function revealDay(ymd: string, smooth = true) {
		const el = scroller?.querySelector<HTMLElement>(`[data-own="${ymd}"]`);
		if (!el) return;
		el.scrollIntoView({ block: 'nearest', behavior: smooth ? 'smooth' : 'auto' });
	}
	/** 달력 밖 날짜로 갈 때 달 목록을 넓힌다 */
	async function ensureVisible(ymd: string, smooth = true) {
		const ym = ymd.slice(0, 7);
		if (!monthSet.has(ym)) {
			const h = scroller?.scrollHeight || 0;
			const before = monthDiff(ym, startYm) > 0;
			if (before) {
				const add = Math.ceil(monthDiff(ym, startYm) / cols) * cols;
				startYm = addMonths(startYm, -add);
				count += add;
			} else count = monthDiff(startYm, ym) + 1;
			await tick();
			if (before && scroller) scroller.scrollTop += scroller.scrollHeight - h;
		}
		revealDay(ymd, smooth);
	}

	// ── 고르기 ──
	const disabledDay = (ymd: string) => !!minYMD && ymd < minYMD;
	function pick(ymd: string) {
		if (disabledDay(ymd)) return;
		cursor = ymd;
		if (isSingle) {
			onchange({ from: ymd });
			close();
			return;
		}
		let f = from,
			t = to;
		if (f && t) [f, t] = [ymd, ''];
		else if (f)
			t = ymd; // 같은 날을 한 번 더 = 하루
		else f = ymd;
		if (f && t && f > t) [f, t] = [t, f];
		from = f;
		to = t;
		hover = '';
		syncText();
	}
	function setRange(f: string, t: string) {
		if (minYMD && f < minYMD) f = minYMD;
		if (t && t < f) t = f;
		from = f;
		to = isSingle ? '' : t;
		cursor = t || f;
		hover = '';
		syncText();
		ensureVisible(f);
	}
	function apply() {
		if (!from || (!isSingle && !to)) return;
		onchange(isSingle ? { from } : { from, to });
		close();
	}
	function sameDay() {
		if (!from) return;
		onchange(isSingle ? { from } : { from, to: from });
		close();
	}
	async function reset() {
		if (resetConfirm && !(await confirm('날짜를 초기화 하시겠습니까?'))) return;
		onchange(null);
		close();
	}

	// ── 빠른 선택 ──
	const weekStart = (s: string) => addDays(s, -parse(s).getDay());
	const monthEnd = (ym: string) => fmt(new Date(Number(ym.slice(0, 4)), Number(ym.slice(5, 7)), 0));
	const chips = $derived.by(() => {
		const t = today;
		const base = from || t;
		const list: { key: string; label: string; f: string; to: string }[] = [
			{ key: 'today', label: '오늘', f: t, to: t },
			{ key: 'tomorrow', label: '내일', f: addDays(t, 1), to: addDays(t, 1) },
			{ key: 'week', label: '이번 주', f: weekStart(t), to: addDays(weekStart(t), 6) },
			{ key: 'nextweek', label: '다음 주', f: addDays(weekStart(t), 7), to: addDays(weekStart(t), 13) },
			{ key: 'month', label: '이번 달', f: `${t.slice(0, 7)}-01`, to: monthEnd(t.slice(0, 7)) },
			{ key: 'nextmonth', label: '다음 달', f: `${addMonths(t.slice(0, 7), 1)}-01`, to: monthEnd(addMonths(t.slice(0, 7), 1)) },
			{ key: '7', label: '+7일', f: base, to: addDays(base, 6) },
			{ key: '30', label: '+30일', f: base, to: addDays(base, 29) }
		];
		return isSingle ? list.filter((c) => c.f == c.to) : list;
	});

	// ── 위에 보이는 고른 기간 ──
	function label(s: string) {
		const d = parse(s);
		const y = d.getFullYear() != parse(today).getFullYear() ? `${d.getFullYear()}년 ` : '';
		return `${y}${d.getMonth() + 1}월 ${d.getDate()}일(${WEEK[d.getDay()]})`;
	}
	const preview = $derived(!isSingle && from && !to ? (kbd ? cursor : hover) : '');
	const summary = $derived.by(() => {
		if (!from) return { text: isSingle ? '날짜를 고르세요' : `${fromLabel} 날짜를 고르세요`, hint: true };
		if (isSingle) return { text: label(from), hint: false };
		if (!to) {
			const p = preview && preview != from ? preview : '';
			if (p) {
				const [a, b] = p < from ? [p, from] : [from, p];
				return { text: `${label(a)} ~ ${label(b)} · ${dayCount(a, b)}일`, hint: true };
			}
			return { text: `${label(from)} ~ ${toLabel} 날짜를 고르세요`, hint: true };
		}
		return { text: from == to ? `${label(from)} · 하루` : `${label(from)} ~ ${label(to)} · ${dayCount(from, to)}일`, hint: false };
	});

	// 칸 모양 — 사이 칸·미리 칠하기
	const range = $derived.by(() => {
		if (from && to) return { a: from, b: to, pre: false };
		if (from && preview) return preview < from ? { a: preview, b: from, pre: true } : { a: from, b: preview, pre: true };
		return null;
	});

	// ── 키보드 ──
	async function onKey(e: KeyboardEvent) {
		if (e.key == 'Escape') {
			e.preventDefault();
			e.stopPropagation();
			close();
			return;
		}
		const tgt = e.target as HTMLElement;
		if (tgt?.tagName == 'INPUT' || tgt?.tagName == 'BUTTON' || tgt?.closest('button')) return;
		const moves: Record<string, number> = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7 };
		if (e.key in moves || e.key == 'PageUp' || e.key == 'PageDown' || e.key == 'Home') {
			e.preventDefault();
			kbd = true;
			let c = cursor || from || today;
			if (e.key in moves) c = addDays(c, moves[e.key]);
			else if (e.key == 'Home') c = today;
			else {
				const d = parse(c);
				const nd = new Date(d.getFullYear(), d.getMonth() + (e.key == 'PageDown' ? 1 : -1), 1);
				const last = new Date(nd.getFullYear(), nd.getMonth() + 1, 0).getDate();
				c = fmt(new Date(nd.getFullYear(), nd.getMonth(), Math.min(d.getDate(), last)));
			}
			cursor = c;
			await ensureVisible(c);
			return;
		}
		if (e.key == 'Enter') {
			e.preventDefault();
			if (e.metaKey || e.ctrlKey) return apply();
			kbd = true;
			pick(cursor || today);
		}
	}

	// ── 직접 입력 ──
	function commitText(which: 'from' | 'to', final: boolean) {
		const raw = which == 'from' ? fromText : toText;
		const v = normalize(raw);
		const bad = !!raw.trim() && !v;
		if (!v) {
			if (final) {
				if (which == 'from') fromBad = bad;
				else toBad = bad;
			}
			return;
		}
		if (minYMD && v < minYMD) {
			if (final) which == 'from' ? (fromBad = true) : (toBad = true);
			return;
		}
		let f = which == 'from' ? v : from;
		let t = which == 'to' ? v : to;
		if (f && t && f > t) [f, t] = [t, f];
		from = f;
		to = t;
		cursor = v;
		if (final) syncText();
		else which == 'from' ? (fromBad = false) : (toBad = false);
		ensureVisible(v);
	}

	const scrollH = $derived(media.mdDown ? 0 : Math.max(innerH - 300, 240));
</script>

{#if template}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div
		class="MuiStack-root adm-drm-trigger"
		style:cursor={readonly ? 'default' : 'pointer'}
		onclick={(e) => {
			if (readonly) return;
			e.stopPropagation();
			show();
		}}
	>
		{@render template()}
	</div>
{/if}

{#if open}
	<div use:portal role="presentation" class="MuiModal-root adm-drm-root" onkeydown={onKey}>
		<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
		<div class="MuiBackdrop-root MuiModal-backdrop adm-drm-backdrop" aria-hidden="true" onclick={close}></div>
		<div bind:this={box} class={['MuiBox-root adm-drm-box', media.mdDown && 'MuiDialog-paperFullScreen adm-drm-full']} role="dialog" aria-modal="true" aria-label="기간 고르기" tabindex="-1">
			<IconButton size="Small" class="adm-drm-close" aria-label="닫기" onclick={close}><Icon name="close" sx={{ fontSize: 30 }} /></IconButton>

			<div class="adm-drm-head">
				<div class={['adm-drm-summary', summary.hint && 'hint']} aria-live="polite">{summary.text}</div>
				<div class="adm-drm-nav">
					<IconButton size="Small" aria-label="앞 달" onclick={() => step(-1)}><Icon name="chevron_left" /></IconButton>
					<IconButton size="Small" aria-label="다음 달" onclick={() => step(1)}><Icon name="chevron_right" /></IconButton>
				</div>
			</div>
			{#if quick}
				<div class="adm-drm-chips">
					{#each chips as c (c.key)}
						{@const on = !!from && from == c.f && (isSingle || (to || from) == c.to)}
						<button type="button" class={['adm-drm-chip', on && 'on']} disabled={!!minYMD && c.to < minYMD} onclick={() => setRange(c.f, c.to)}>{c.label}</button>
					{/each}
				</div>
			{/if}

			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="adm-drm-scroll" bind:this={scroller} style:height={scrollH ? `${scrollH}px` : null} onmouseleave={() => (hover = '')}>
				<div class="adm-drm-grid">
					{#each monthList as ym (ym)}
						<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root adm-drm-card" data-ym={ym}>
							<div class="adm-drm-cardin">
								<h4 class="MuiTypography-root MuiTypography-h4 adm-drm-month">{ym.replace('-', '.')}월</h4>
								<table class="adm-drm-table" role="grid">
									<thead>
										<tr
											>{#each WEEK as w (w)}<th>{w}</th>{/each}</tr
										>
									</thead>
									<tbody>
										{#each weeksOf(ym) as row, ri (ri)}
											<tr>
												{#each row as ymd, wd (ymd)}
													{@const shown = monthSet.has(ymd.slice(0, 7)) && !(ym == anchorYm && ymd.slice(0, 7) < ym)}
													{@const own = ymd.slice(0, 7) == ym}
													{@const dis = disabledDay(ymd)}
													{@const holy = holydays[ymd]}
													{@const isFrom = shown && ymd == from}
													{@const isTo = shown && !isSingle && ymd == to}
													{@const isPre = shown && !!range?.pre && ymd == preview && !isFrom}
													{@const inRange = shown && !!range && ymd >= range.a && ymd <= range.b}
													<td
														class={[
															'adm-drm-td',
															ymd < today && 'past',
															ymd == today && 'today',
															inRange && (range?.pre ? 'pre' : 'in'),
															inRange && ymd == range?.a && 'ra',
															inRange && ymd == range?.b && 'rb'
														]}
														style:cursor={dis || !shown ? 'default' : 'pointer'}
														data-own={own ? ymd : undefined}
														aria-selected={isFrom || isTo}
														onclick={() => shown && ((kbd = false), pick(ymd))}
														onmouseenter={() => shown && !dis && ((hover = ymd), (kbd = false))}
													>
														{#if shown}
															<div class="adm-drm-cell" style:color={holy ? YOIL_COLOR[0] : YOIL_COLOR[wd] || null} style:opacity={dis ? 0.5 : null}>
																<span class={['adm-drm-day', (isFrom || isTo) && 'sel', isPre && 'pre', kbd && own && ymd == cursor && 'cur']}>{Number(ymd.slice(8))}</span>
																{#if isFrom && isTo}
																	<span class="adm-drm-tag">당일</span>
																{:else if isFrom}
																	<span class="adm-drm-tag">{fromLabel}</span>
																{:else if isTo}
																	<span class="adm-drm-tag">{toLabel}</span>
																{/if}
																{#if holy}<span class="adm-drm-holy">{holy}</span>{/if}
															</div>
														{/if}
													</td>
												{/each}
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div class="adm-drm-foot">
				<div class="adm-drm-left">
					{#if useReset}
						<Button variant="outlined" color="error" class="adm-drm-btn40" onclick={reset}>{resetLabel}</Button>
					{/if}
					{#if !media.mdDown}
						<label class={['adm-drm-input', fromBad && 'bad']}>
							<input
								placeholder="YYYY-MM-DD"
								aria-label={`${fromLabel} 직접 입력`}
								bind:value={fromText}
								oninput={() => commitText('from', false)}
								onblur={() => commitText('from', true)}
								onkeydown={(e) => e.key == 'Enter' && (e.preventDefault(), commitText('from', true))}
							/>
							<Icon name="calendar_month" sx={{ fontSize: 22, color: 'var(--adm-action)' }} />
						</label>
						{#if !isSingle}
							<span>~</span>
							<label class={['adm-drm-input', toBad && 'bad']}>
								<input
									placeholder="YYYY-MM-DD"
									aria-label={`${toLabel} 직접 입력`}
									bind:value={toText}
									oninput={() => commitText('to', false)}
									onblur={() => commitText('to', true)}
									onkeydown={(e) => e.key == 'Enter' && (e.preventDefault(), commitText('to', true))}
								/>
								<Icon name="calendar_month" sx={{ fontSize: 22, color: 'var(--adm-action)' }} />
							</label>
						{/if}
					{/if}
				</div>
				<div class="adm-drm-right">
					<Button variant="outlined" onclick={close}>취소</Button>
					{#if !isSingle}
						<Button variant="outlined" disabled={!from || (!!to && to != from)} onclick={sameDay}>당일</Button>
					{/if}
					<Button variant="contained" disabled={!from || (!isSingle && !to)} onclick={apply}>적용</Button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.adm-drm-trigger {
		display: flex;
		flex-direction: column;
	}
	.adm-drm-root {
		position: fixed;
		inset: 0;
		z-index: 1300;
	}
	.adm-drm-backdrop {
		position: fixed;
		inset: 0;
		z-index: -1;
		background-color: rgba(0, 0, 0, 0.5);
	}
	/* 원본: 가운데 70% · 바탕 종이 · 여백 24 24 16 · 그림자(테마 elevation 24) */
	.adm-drm-box {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		width: 70%;
		background-color: var(--adm-paper, #fff);
		box-shadow: 0px 9px 46px rgba(0, 0, 0, 0.08);
		z-index: 1;
		padding: 24px 24px 16px;
		outline: none;
		display: flex;
		flex-direction: column;
		box-sizing: border-box;
		max-height: 100%;
	}
	.adm-drm-full {
		inset: 0;
		width: 100%;
		height: 100%;
		max-width: 100vw;
		transform: none;
		padding: 0;
		overflow-x: hidden;
	}
	.adm-drm-box :global(.adm-drm-close) {
		position: absolute;
		border-radius: 0;
		right: 8px;
		top: 8px;
		z-index: 10;
	}

	/* 더한 것: 고른 기간 줄 · 빠른 선택 */
	.adm-drm-head {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 48px 8px 8px;
		min-height: 40px;
	}
	.adm-drm-full .adm-drm-head {
		padding: 8px 52px 4px 16px;
	}
	.adm-drm-summary {
		flex: 1;
		min-width: 0;
		font-family: var(--mui-font);
		font-size: 16px;
		font-weight: 700;
		line-height: 1.5;
		color: var(--adm-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.adm-drm-summary.hint {
		color: var(--adm-text2);
		font-weight: 600;
	}
	.adm-drm-nav {
		display: flex;
		flex: 0 0 auto;
	}
	.adm-drm-chips {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		padding: 0 8px 4px;
	}
	.adm-drm-full .adm-drm-chips {
		flex-wrap: nowrap;
		overflow-x: auto;
		padding: 0 16px 4px;
		scrollbar-width: none;
	}
	.adm-drm-chip {
		flex: 0 0 auto;
		height: 30px;
		padding: 0 12px;
		border-radius: 15px;
		border: 1px solid var(--theme-color-alpha30);
		background: transparent;
		color: var(--theme-color-main);
		font-family: var(--mui-font);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		white-space: nowrap;
	}
	.adm-drm-chip:hover {
		background: var(--theme-color-alpha8);
	}
	.adm-drm-chip.on {
		background: var(--theme-color-main);
		border-color: var(--theme-color-main);
		color: var(--theme-color-contrast, #fff);
	}
	.adm-drm-chip:disabled {
		opacity: 0.4;
		cursor: default;
	}

	.adm-drm-scroll {
		position: relative;
		overflow-y: auto;
		overflow-x: hidden;
	}
	.adm-drm-full .adm-drm-scroll {
		flex: 1;
		min-height: 0;
	}
	.adm-drm-grid {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 16px;
		padding: 8px;
	}
	.adm-drm-grid :global(*) {
		box-sizing: border-box;
	}
	.adm-drm-card {
		padding: 16px 8px;
		width: 100%;
	}
	@media (min-width: 900px) {
		.adm-drm-card {
			width: 48%;
		}
	}
	@media (min-width: 1280px) {
		.adm-drm-card {
			width: 49%;
		}
	}
	@media (min-width: 1440px) {
		.adm-drm-card {
			width: 32%;
		}
	}
	.adm-drm-full .adm-drm-card {
		padding: 8px 4px;
	}
	.adm-drm-cardin {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.adm-drm-month {
		margin: 0;
		text-align: center;
	}
	.adm-drm-table {
		width: 100%;
		table-layout: fixed;
		border-collapse: collapse;
		border-spacing: 0;
	}
	.adm-drm-table th {
		width: 14.28%;
		padding: 12px 0;
		font-family: var(--mui-font);
		font-size: 12px;
		font-weight: normal;
		line-height: 12px;
		letter-spacing: 0.5px;
		color: var(--adm-n700, rgb(47, 55, 70));
		text-align: center;
		background: transparent;
	}
	.adm-drm-full .adm-drm-table th {
		padding: 8px 0; /* 글자는 12 그대로(React 도 테마 머리 칸 값이 이긴다) */
	}
	.adm-drm-td {
		width: 14.28%;
		padding: 0;
		border: 0;
		vertical-align: top;
		font-family: var(--mui-font);
		font-size: 12px;
		line-height: 1.5;
		color: var(--gt-d700, #7e7e7e);
		user-select: none;
	}
	.adm-drm-full .adm-drm-td {
		font-size: 11px;
	}
	.adm-drm-td.past {
		color: var(--gt-500, #cfcfcf);
	}
	.adm-drm-td.today {
		background-color: var(--gt-300, #ebebeb);
	}
	/* 사이 칸 — 고른 범위는 진하게, 미리 보기는 옅게 */
	.adm-drm-td.in {
		background-color: var(--theme-color-alpha12);
	}
	.adm-drm-td.pre {
		background-color: var(--theme-color-alpha8);
	}
	.adm-drm-td.ra {
		border-top-left-radius: 8px;
		border-bottom-left-radius: 8px;
	}
	.adm-drm-td.rb {
		border-top-right-radius: 8px;
		border-bottom-right-radius: 8px;
	}
	.adm-drm-cell {
		display: flex;
		flex-direction: column;
		flex-grow: 1;
		justify-content: center;
		align-items: center;
		min-height: 50px;
		position: relative;
	}
	.adm-drm-day {
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 30px;
		min-height: 30px;
		border-radius: 8px;
		font-family: var(--mui-font);
		font-size: 15px;
		font-weight: 700;
		line-height: 1.5;
	}
	.adm-drm-full .adm-drm-day {
		min-width: 24px;
		min-height: 24px;
	}
	.adm-drm-td:not(.past):hover .adm-drm-day {
		box-shadow: inset 0 0 0 1px var(--theme-color-alpha50);
	}
	.adm-drm-day.sel {
		color: var(--theme-color-main);
		background: var(--theme-color-alpha30);
	}
	.adm-drm-day.pre {
		color: var(--theme-color-main);
		box-shadow: inset 0 0 0 1px var(--theme-color-main);
	}
	.adm-drm-day.cur {
		outline: 2px solid var(--theme-color-main);
		outline-offset: 1px;
	}
	.adm-drm-tag {
		font-family: var(--mui-font);
		font-size: 12px;
		font-weight: 700;
		line-height: 1.5;
		color: inherit; /* React 와 같게 — 칸 글자색(요일 색) */
		white-space: nowrap;
	}
	.adm-drm-holy {
		font-family: var(--mui-font);
		font-size: 11px;
		font-weight: 500;
		line-height: 1.5;
		letter-spacing: -1.5px;
		color: var(--theme-color-alpha50);
	}

	.adm-drm-foot {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
		padding-top: 16px;
	}
	.adm-drm-full .adm-drm-foot {
		padding: 12px 16px 16px;
		border-top: 1px solid var(--adm-divider);
	}
	.adm-drm-left,
	.adm-drm-right {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.adm-drm-right {
		padding-top: 4px;
		justify-content: flex-end;
	}
	.adm-drm-left :global(.adm-drm-btn40) {
		height: 40px;
		white-space: nowrap;
		flex-shrink: 0;
	}
	.adm-drm-input {
		display: flex;
		align-items: center;
		width: 145px;
		height: 40px;
		flex-shrink: 0;
		padding: 0 8px 0 14px;
		border: 1px solid var(--adm-n200, #d2d6db);
		border-radius: 8px;
		box-sizing: border-box;
		cursor: text;
	}
	.adm-drm-input:hover {
		border-color: var(--adm-text);
	}
	.adm-drm-input:focus-within {
		border: 2px solid var(--theme-color-main);
		padding-left: 13px;
		padding-right: 7px;
	}
	.adm-drm-input.bad {
		border-color: var(--adm-error-main, #f04438);
	}
	.adm-drm-input input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: none;
		background: transparent;
		padding: 0;
		font-family: var(--mui-font);
		font-size: 14px;
		color: var(--adm-text);
	}
</style>
