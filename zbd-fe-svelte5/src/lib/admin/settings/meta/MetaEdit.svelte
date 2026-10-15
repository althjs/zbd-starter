<script lang="ts">
	/**
	 * 메타 RAW 수정(React setting/meta/components/MetaEdit) — 이름(50자)·메모(100자) · JSON 편집기(Monaco) · 정렬하기·한줄로 만들기 · 저장(RW, 바뀐 것이 있을 때) ·
	 * JSON 오류는 빨간 상자, 바뀐 내용 미리보기. 주소 `/admin/setting/meta/setting/{호스팅|edit}/{metaKey}` — API `/service/admin/meta?metaKey=&hostingId=`.
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import PageContainer from '../../layout/PageContainer.svelte';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Card from '../../ui/Card.svelte';
	import TextField from '../../ui/TextField.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { getHostingItems } from '../../meta.svelte.ts';
	import CodeEditor from '#lib/admin/kit/CodeEditor.svelte';
	import DiffPreview from '#lib/admin/kit/DiffPreview.svelte';
	import { diffHtml } from '#lib/admin/kit/diff.ts';

	let { hostingId, metaKey }: { hostingId: string; metaKey: string } = $props();
	const xx = $derived(hostingId == 'edit' ? '' : hostingId);
	/** React COMPANIES_MAP[호스팅]?.name — 호스팅 목록 설정(HOSTINGS.items, 옛 COMPANIES) 에서 */
	const companyName = $derived(getHostingItems().find((o) => o.hostingId == (xx || ''))?.name || '');
	let meta = $state<any>(null);
	let orig: any = null;
	let diff = $state('');
	let err = $state('');
	let editor: any = null;
	const h = typeof window != 'undefined' ? (window.innerHeight - 600 > 1000 ? window.innerHeight - 600 : 1000) : 1000;

	$effect(() => {
		const k = metaKey;
		untrack(async () => {
			const o = await api.get(`/service/admin/meta?metaKey=${k}${xx ? `&hostingId=${xx}` : ''}`);
			if (!o.ok) return alert('데이터 로딩 도중 오류가 발생하였습니다.');
			orig = JSON.parse(JSON.stringify(o.data ?? {}));
			meta = o.data || {};
		});
	});
	$effect(() => {
		if (!meta) return;
		const m = $state.snapshot(meta);
		diffHtml(orig, m).then((d) => (diff = d));
	});
	function onJson(v: string) {
		try {
			meta = { ...meta, json: JSON.parse(v) };
			err = '';
		} catch (e: any) {
			err = String(e?.stack || e?.message || e)
				.split('\n')
				.slice(0, 4)
				.join('\n');
		}
	}
	async function save() {
		const o = await api.put(`/service/admin/meta?metaKey=${metaKey}${admin.profile?.isMaster ? `&hostingId=${xx || ''}` : ''}`, { ...meta, metaKey });
		if (!o.ok) return alert('데이터 로딩 도중 오류가 발생하였습니다.');
		editor?.setValue(JSON.stringify(o.data.json, null, 2));
		orig = JSON.parse(JSON.stringify(o.data));
		meta = o.data;
		toast('정상적으로 저장되었습니다.');
	}
</script>

<PageContainer hintPostId="lug2tbcq" title={`${metaKey} 메타정보 수정`} subtitle="메타 RAW 데이터를 관리합니다.">
	{#if meta}
		<Card sx={{ padding: 24 }}>
			<div class={cx('MuiStack-root', toClass({ '& > :not(style) ~ :not(style)': { marginTop: 8 } }))}>
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', '& > :not(style) ~ :not(style)': { marginLeft: 8 } }))}>
					<TextField label={`${companyName} 메타 이름`} value={meta.name || ''} oninput={(v: string) => (meta = { ...meta, name: v.length > 50 ? v.substring(0, 50) : v })} />
					<TextField label="메모" value={meta.memo || ''} fullWidth multiline minRows={2} oninput={(v: string) => (meta = { ...meta, memo: v.length > 100 ? v.substring(0, 100) : v })} />
				</div>
				<CodeEditor height={h} language="json" defaultValue={JSON.stringify(meta.json, null, 2)} theme="vs-dark" options={{ tabSize: 2 }} onChange={onJson} onMount={(e) => (editor = e)} />
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', '& > :not(style) ~ :not(style)': { marginLeft: 8 } }))}>
					<Button variant="outlined" disabled={!!err} onclick={() => editor?.setValue(JSON.stringify(meta.json, null, 2))}>정렬하기</Button>
					<Button variant="outlined" disabled={!!err} onclick={() => editor?.setValue(JSON.stringify(meta.json))}>한줄로 만들기</Button>
					<AuthButton variant="contained" sx={{ width: 'fit-content' }} disabled={!diff} onclick={save}>저장</AuthButton>
				</div>
			</div>
		</Card>
	{/if}
	{#if err}
		<Card sx={{ marginTop: 16, padding: 24, backgroundColor: '#e53935', color: 'yellow' }}><pre>{err}</pre></Card>
	{/if}
	{#if !err && diff && admin.profile?.isDeveloper}
		<div class={cx('MuiGrid-root', toClass({ margin: 16, maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview diffHTML={diff} /></div>
	{/if}
</PageContainer>
