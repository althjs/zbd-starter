<script lang="ts">
	/**
	 * 호스팅 공통 메타(COMMON · HOSTINGS(옛 COMPANIES)) 탭 — React 는 전용 화면(공통 메타 관리 543줄 · 호스팅 항목 폼 1,779줄 — 여러 호스팅·커머스·i18n·휴일·워터마크)이다.
	 * 즈비드 한 호스팅 기준: 이 메타는 호스팅 공통(hostingId "")이라 관리자 메타 API 가 주지 않는다(GET /service/admin/meta 에 없음) → 없으면 안내만.
	 * 있으면(다른 계정·나중에 API 가 줄 때) JSON 을 바로 고쳐 PUT /service/admin/meta/{metaKey} (React 저장과 같은 주소). 여러 호스팅 탭·커머스 칸은 옮기지 않았다.
	 */
	import { admin } from '../../profile.svelte.ts';
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import AuthButton from '../../ui/AuthButton.svelte';
	import { alert, toast } from '../../ui/modal.svelte.ts';
	import CodeEditor from '#lib/admin/kit/CodeEditor.svelte';
	import DiffPreview from '#lib/admin/kit/DiffPreview.svelte';
	import { diffHtml } from '#lib/admin/kit/diff.ts';

	let { meta, metaKey, hostingId = '', onChange }: { meta: any; metaKey: string; hostingId?: string; onChange: (m: any) => void } = $props();
	let json = $state<any>(null);
	let err = $state('');
	let diff = $state('');
	$effect(() => {
		json = meta ? JSON.parse(JSON.stringify(meta.json)) : null;
	});
	$effect(() => {
		if (!meta || !json) return;
		diffHtml(meta.json, $state.snapshot(json)).then((d) => (diff = d));
	});
	async function save() {
		const o = await api.put(`/service/admin/meta/${metaKey}${hostingId ? `?hostingId=${hostingId}` : ''}`, json);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		diff = '';
		onChange({ ...meta, json: { ...json } });
		toast('정상적으로 저장되었습니다.');
	}
</script>

{#if !meta}
	<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ textAlign: 'center', margin: '40px 0', color: 'var(--adm-text2)' }))}>
		{metaKey} 는 호스팅 공통 메타라 이 계정에서 읽을 수 없습니다.
	</p>
{:else}
	<div class={cx('MuiStack-root', toClass({ gap: 16, marginTop: 16 }))}>
		<CodeEditor
			height={600}
			language="json"
			defaultValue={JSON.stringify(meta.json, null, 2)}
			options={{ tabSize: 2 }}
			onChange={(v) => {
				try {
					json = JSON.parse(v);
					err = '';
				} catch (e: any) {
					err = String(e?.message || e);
				}
			}}
		/>
		{#if err}<pre class={toClass({ background: '#e53935', color: 'yellow', padding: 16, borderRadius: 8 })}>{err}</pre>{/if}
		<AuthButton variant="contained" sx={{ width: 'fit-content' }} disabled={!diff || !!err} onclick={save}>저장</AuthButton>
		{#if diff && admin.profile?.isDeveloper}<div class={cx('MuiStack-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview diffHTML={diff} /></div>{/if}
	</div>
{/if}
