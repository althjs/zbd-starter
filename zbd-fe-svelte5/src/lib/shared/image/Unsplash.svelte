<script lang="ts">
	/**
	 * Unsplash 무료 사진 찾기 — 고르면 사진 파일(Blob) 을 돌려준다(자르기·올리기는 ImagePicker 가).
	 * 공개용 Access Key 는 사이트 설정에서(unsplash-key.svelte.ts — 코드에 두지 않는다) · 한 번에 30장(상한).
	 */
	import { onMount } from 'svelte';
	import MuiButton from '../mui/MuiButton.svelte';
	import TextField from '../mui/TextField.svelte';

	let { onpick, accessKey }: { onpick: (b: Blob, name: string) => void | Promise<void>; accessKey: string } = $props();

	// 처음 열면 자주 쓸 만한 키워드 하나로 바로 찾아 둔다(빈 화면 대신, 2026-10-03 대표). 아래 칩으로 바꿔 볼 수 있다
	const SUGGEST = ['sky', 'sunset', 'ocean', 'forest', 'mountain', 'flower', 'city night', 'pastel', 'minimal', 'texture', 'coffee', 'cat', 'travel', 'rain'];

	let keyword = $state(SUGGEST[Math.floor(Math.random() * SUGGEST.length)]);
	let list = $state<any[]>([]);
	let page = $state(1);
	let busy = $state(false);

	async function search(more = false) {
		if (!keyword.trim()) return;
		busy = true;
		page = more ? page + 1 : 1;
		const r = await fetch(`https://api.unsplash.com/search/photos?query=${encodeURIComponent(keyword)}&per_page=30&page=${page}&client_id=${encodeURIComponent(accessKey)}`)
			.then((x) => x.json())
			.catch(() => null);
		list = more ? [...list, ...(r?.results || [])] : r?.results || [];
		busy = false;
	}
	// 한 번만(effect 로 두면 keyword 를 따라가 글자를 칠 때마다 찾는다)
	onMount(() => search());
	function pickKeyword(k: string) {
		keyword = k;
		search();
	}
	// 올리기가 끝날 때까지 잠근다(받기만 기다리면 줄이는 사이 다시 눌려 같은 사진이 여러 번 올라갔다, 2026-10-05)
	async function pick(p: any) {
		if (busy) return;
		busy = true;
		try {
			// Unsplash 규칙 — 내려받기를 알린다
			fetch(`${p.links.download_location}&client_id=${encodeURIComponent(accessKey)}`).catch(() => {});
			const b = await fetch(p.urls.regular).then((x) => x.blob());
			await onpick(b, `unsplash_${p.id}.jpg`);
		} finally {
			busy = false;
		}
	}
</script>

<div class="bar" role="presentation" onkeydown={(e) => e.key == 'Enter' && search()}>
	<TextField bind:value={keyword} size="small" placeholder="영어로 검색하면 더 많이 나와요 (예: coffee)" />
	<MuiButton variant="contained" onclick={() => search()}>찾기</MuiButton>
</div>
<div class="chips">
	{#each SUGGEST as k (k)}
		<button type="button" class="chip" class:on={keyword == k} onclick={() => pickKeyword(k)}>{k}</button>
	{/each}
</div>
<div class="grid">
	{#each list as p (p.id)}
		<button class="item" onclick={() => pick(p)} disabled={busy}><img src={p.urls.thumb} alt={p.alt_description || ''} loading="lazy" /></button>
	{/each}
</div>
{#if list.length >= 30 * page}
	<MuiButton onclick={() => search(true)} disabled={busy}>더 보기</MuiButton>
{/if}

<style>
	.bar {
		display: flex;
		gap: 8px;
		align-items: center;
		margin: 8px 0 12px;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin: -4px 0 12px;
	}
	.chip {
		border: 1px solid #e4e7ec;
		background: #fff;
		color: #475467;
		border-radius: 999px;
		padding: 4px 10px;
		font-size: 12px;
		cursor: pointer;
	}
	.chip:hover {
		border-color: #98a2b3;
	}
	.chip.on {
		border-color: #1d433d;
		background: #1d433d;
		color: #fff;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
		gap: 6px;
	}
	.item {
		aspect-ratio: 1;
		overflow: hidden;
		border-radius: 6px;
		cursor: pointer;
	}
	.item img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
</style>
