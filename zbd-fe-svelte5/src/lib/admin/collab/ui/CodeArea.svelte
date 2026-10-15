<script lang="ts">
	/**
	 * 글 편집 칸 — 어두운 바탕 + 줄 번호(React 는 Monaco Editor vs-dark, 줄 바꿈 켬). Monaco 는 이 앱에 없다 — 문법 색칠·자동 완성은 없다(공용 후보).
	 *   <CodeArea bind:value={text} onsave={(v) => save(v)} />   Ctrl/⌘+S = onsave · Tab = 공백 2칸
	 * 줄이 접히면(줄 바꿈) 줄 번호가 그 줄 첫머리에 오도록 보이지 않는 거울 칸으로 줄 높이를 잰다.
	 */
	import { tick } from 'svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	let {
		value = $bindable(''),
		readonly = false,
		onsave,
		oninput,
		sx = null
	}: { value?: string; readonly?: boolean; onsave?: (v: string) => void; oninput?: (v: string) => void; sx?: Sx | null } = $props();
	let ta = $state<HTMLTextAreaElement>();
	let gutter = $state<HTMLElement>();
	let mirror = $state<HTMLElement>();
	let heights = $state<number[]>([]);
	const lines = $derived(`${value || ''}`.split('\n'));
	async function measure() {
		await tick();
		if (!mirror || !ta) return;
		mirror.style.width = `${ta.clientWidth}px`;
		heights = Array.from(mirror.children).map((c) => (c as HTMLElement).offsetHeight);
	}
	$effect(() => {
		void lines.length;
		void value;
		measure();
	});
	$effect(() => {
		if (!ta) return;
		const ro = new ResizeObserver(() => measure());
		ro.observe(ta);
		return () => ro.disconnect();
	});
	function key(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() == 's') {
			e.preventDefault();
			onsave?.(value);
			return;
		}
		if (e.key == 'Tab' && !readonly) {
			e.preventDefault();
			const el = ta!;
			const s = el.selectionStart,
				t = el.selectionEnd;
			value = value.slice(0, s) + '  ' + value.slice(t);
			oninput?.(value);
			requestAnimationFrame(() => (el.selectionStart = el.selectionEnd = s + 2));
		}
	}
</script>

<div class={cx('adm-code', toClass(sx))}>
	<div class="adm-code-gutter" bind:this={gutter} aria-hidden="true">
		{#each lines as _, i (i)}<div style:height={heights[i] ? `${heights[i]}px` : undefined}>{i + 1}</div>{/each}
	</div>
	<div class="adm-code-body">
		<textarea bind:this={ta} spellcheck="false" {readonly} bind:value oninput={() => oninput?.(value)} onkeydown={key} onscroll={() => gutter && ta && (gutter.scrollTop = ta.scrollTop)}></textarea>
		<div class="adm-code-mirror" bind:this={mirror} aria-hidden="true">
			{#each lines as l, i (i)}<div>{l || ' '}</div>{/each}
		</div>
	</div>
</div>
