<script lang="ts">
	/** 웹하드 엑셀·CSV·파워포인트·워드 미리보기 — 웹오피스 보기 화면(인쇄 모양 읽기, 격리 iframe)으로 바로 */
	// 웹오피스는 플러그인 — 있을 때만 붙는다(스타터엔 없음, 파일이 없으면 glob 이 비어 이 화면은 쓰이지 않는다)
	const FMT = import.meta.glob<any>('../office/core/formats.ts');
	const PRT = import.meta.glob<any>('../office/core/print.ts');
	const SHT = import.meta.glob<any>('../office/sheet/sheet-doc.ts');
	const SLD = import.meta.glob<any>('../office/slide/slide-doc.ts');
	const DOC = import.meta.glob<any>('../office/doc/doc-doc.ts');
	const toDataUrl = (b: Blob) =>
		new Promise<string>((res, rej) => {
			const r = new FileReader();
			r.onload = () => res(String(r.result));
			r.onerror = () => rej(r.error);
			r.readAsDataURL(b);
		});

	let { blob, name }: { blob: Blob; name: string } = $props();
	const MAX = 10 * 1024 * 1024;
	let html = $state('');
	let err = $state('');

	$effect(() => {
		const b = blob;
		const n = name;
		html = '';
		err = '';
		if (b.size > MAX) return void (err = '파일이 커서(10MB 초과) 미리보기를 건너뜁니다. 다운로드하거나 웹오피스에서 열어 주세요.');
		let live = true;
		(async () => {
			try {
				const [f, p, sh, sl, dc] = await Promise.all([Object.values(FMT)[0]?.(), Object.values(PRT)[0]?.(), Object.values(SHT)[0]?.(), Object.values(SLD)[0]?.(), Object.values(DOC)[0]?.()]);
				if (!f || !p || !sh || !sl || !dc) throw new Error('웹오피스 없음');
				const im = f.importerFor(n);
				if (!im) throw new Error('읽을 수 없는 형식');
				// 그림은 서버에 올리지 않고 문서 안에 그대로(data 주소) — 미리보기는 저장하지 않는다
				const data = await im.run(new File([b], n), { postId: '', upload: toDataUrl });
				const parsed = im.kind == 'sheet' ? sh.parseSheet(data) : im.kind == 'slide' ? sl.parseSlides(data) : dc.parseDocFile(data);
				const v = p.printView(im.kind, parsed);
				const doc = `<!doctype html><html><head><meta charset="utf-8"><base target="_blank"><style>${v.css}
body{margin:0;padding:12px;font:13px/1.4 -apple-system,BlinkMacSystemFont,'Pretendard',sans-serif;background:#fff;color:#111}table{background:#fff}</style></head><body>${v.html}</body></html>`;
				if (live) html = doc;
			} catch (e) {
				if (live) err = `문서를 읽지 못했어요: ${(e as Error)?.message || e}`;
			}
		})();
		return () => (live = false);
	});
</script>

{#if err}
	<p class="zov-msg">{err}</p>
{:else if html}
	<iframe class="zov" title={name} sandbox="" srcdoc={html}></iframe>
{:else}
	<p class="zov-msg">불러오는 중…</p>
{/if}

<style>
	.zov {
		width: 100%;
		flex: 1;
		min-height: 420px;
		border: 0;
		background: #fff;
	}
	.zov-msg {
		padding: 24px;
		text-align: center;
		color: var(--adm-text2);
		font-size: 0.875rem;
	}
</style>
