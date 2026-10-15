<script lang="ts" module>
	export type AutocompleteOption = string | { label: string; value: string };
</script>

<script lang="ts">
	/**
	 * Autocomplete — 입력칸 + 후보 목록(관리자·사용자 공용, 2026-10-04 — 브라우저 datalist 는 모양·검색을 못 바꿔서).
	 *   <Autocomplete options={names} placeholder="지역 검색" onselect={(v) => (area = v)} />
	 *   <Autocomplete options={[{ label: '홍길동 (기획팀)', value: '12' }]} onselect={(v, o) => …} />   ← 보이는 글자와 값이 다를 때
	 * 찾기: 띄어쓰기·대소문자 무시 포함 검색, 앞에서 맞는 것 먼저. 키보드 ↑↓·Enter·Esc, 목록은 입력칸 바로 아래(Scrollbar).
	 * freeSolo(기본 켬) — 후보에 없는 글자도 Enter 로 보낸다(서버가 찾는 경우). 고르면 칸을 비운다(clearOnSelect).
	 */
	import Scrollbar from './Scrollbar.svelte';

	let {
		options = [],
		value = $bindable(''),
		placeholder = '',
		max = 50,
		freeSolo = true,
		clearOnSelect = true,
		class: cls = '',
		style = '',
		onselect
	}: {
		options?: AutocompleteOption[];
		value?: string;
		placeholder?: string;
		max?: number;
		freeSolo?: boolean;
		clearOnSelect?: boolean;
		class?: string;
		style?: string;
		onselect?: (v: string, option: AutocompleteOption | null) => void;
	} = $props();

	const uid = $props.id();
	let open = $state(false);
	let hi = $state(-1);
	let listEl = $state<HTMLDivElement>();
	const norm = (s: string) => s.toLowerCase().replace(/\s+/g, '');
	const labelOf = (o: AutocompleteOption) => (typeof o == 'string' ? o : o.label);
	const valueOf = (o: AutocompleteOption) => (typeof o == 'string' ? o : o.value);

	const shown = $derived.by(() => {
		const q = norm(value);
		if (!q) return options.slice(0, max);
		const head: AutocompleteOption[] = [];
		const rest: AutocompleteOption[] = [];
		for (const o of options) {
			const n = norm(labelOf(o));
			const i = n.indexOf(q);
			if (i == 0) head.push(o);
			else if (i > 0) rest.push(o);
			if (head.length >= max) break;
		}
		return head.concat(rest).slice(0, max);
	});

	function pick(o: AutocompleteOption | null, typed = '') {
		open = false;
		hi = -1;
		value = clearOnSelect ? '' : o ? labelOf(o) : typed;
		onselect?.(o ? valueOf(o) : typed, o);
	}
	function key(e: KeyboardEvent) {
		if (e.isComposing) return; // 한글 조합 중 Enter 는 글자 확정
		if (e.key == 'ArrowDown' || e.key == 'ArrowUp') {
			e.preventDefault();
			open = true;
			const n = shown.length;
			if (!n) return;
			hi = e.key == 'ArrowDown' ? (hi + 1) % n : (hi - 1 + n) % n;
			listEl?.querySelector(`[data-i="${hi}"]`)?.scrollIntoView({ block: 'nearest' });
		} else if (e.key == 'Enter') {
			e.preventDefault();
			if (hi >= 0 && shown[hi]) pick(shown[hi]);
			else if (shown.length && norm(labelOf(shown[0])) == norm(value)) pick(shown[0]);
			else if (freeSolo && value.trim()) pick(null, value.trim());
		} else if (e.key == 'Escape') {
			if (open) e.stopPropagation();
			open = false;
		}
	}
</script>

<div class="zbd-ac {cls}" {style}>
	<input
		class="zbd-ac-in"
		{placeholder}
		bind:value
		autocomplete="off"
		role="combobox"
		aria-controls="{uid}-list"
		aria-expanded={open}
		aria-autocomplete="list"
		onfocus={() => (open = true)}
		oninput={() => {
			open = true;
			hi = -1;
		}}
		onblur={() => setTimeout(() => (open = false), 120)}
		onkeydown={key}
	/>
	<span class="zbd-ac-ar" aria-hidden="true">▾</span>
	{#if open && shown.length}
		<div class="zbd-ac-pop" id="{uid}-list" role="listbox" bind:this={listEl}>
			<Scrollbar class="thin-scrollbar" style="max-height:240px">
				{#each shown as o, i (valueOf(o) + '\u0000' + i)}
					<div
						class="zbd-ac-op"
						class:hi={i == hi}
						role="option"
						tabindex="-1"
						aria-selected={i == hi}
						data-i={i}
						onmousedown={(e) => {
							e.preventDefault();
							pick(o);
						}}
						onmouseenter={() => (hi = i)}
					>
						{labelOf(o)}
					</div>
				{/each}
			</Scrollbar>
		</div>
	{/if}
</div>

<style>
	.zbd-ac {
		position: relative;
		display: inline-block;
	}
	.zbd-ac-in {
		width: 100%;
		height: 38px;
		box-sizing: border-box;
		padding: 8px 28px 8px 12px;
		border: 1px solid var(--zbd-input-border, #c4c4c4);
		border-radius: 8px;
		font: inherit;
		font-weight: 400;
		font-size: 14px;
		background: #fff;
		color: inherit;
	}
	.zbd-ac-in:focus {
		outline: none;
		border-color: var(--theme-color-primary, #1976d2);
		box-shadow: 0 0 0 1px var(--theme-color-primary, #1976d2);
	}
	.zbd-ac-ar {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		font-size: 12px;
		color: #666;
		pointer-events: none;
	}
	.zbd-ac-pop {
		position: absolute;
		left: 0;
		right: 0;
		top: calc(100% + 4px);
		z-index: 10;
		background: #fff;
		border-radius: 8px;
		box-shadow:
			0 5px 5px -3px rgba(0, 0, 0, 0.2),
			0 8px 10px 1px rgba(0, 0, 0, 0.14),
			0 3px 14px 2px rgba(0, 0, 0, 0.12);
		padding: 6px 0;
		font-weight: 400;
		font-size: 14px;
	}
	.zbd-ac-op {
		padding: 8px 14px;
		cursor: pointer;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.zbd-ac-op.hi {
		background: rgba(0, 0, 0, 0.06);
	}
</style>
