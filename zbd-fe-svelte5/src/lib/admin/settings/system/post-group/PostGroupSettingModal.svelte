<script lang="ts">
	/**
	 * 게시판 기본 설정 창 — 환경 설정 › 게시판 관리의 톱니·「추가」. 스타터는 groupKey · 구분 · 게시판 명만 고친다.
	 * 기능(feature)·하위분류(items) 등 나머지 칸은 화면에 없지만 지우지 않는다 — 읽은 게시판 객체를 복사해 **바꾼 칸만** 덮어쓴다.
	 *
	 * 「적용」 은 저장이 아니다 — 고친 게시판 객체를 onChange 로 넘기면 게시판 관리 화면이 그 탭 목록에 끼우고, 화면의 「저장」 이
	 * POST_GROUPS 메타를 다시 읽어 그 탭만 바꿔 PUT 한다.
	 */
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { portal } from '#lib/shared/ui/portal.ts';
	import Button from '../../../ui/Button.svelte';
	import { alert, confirm } from '../../../ui/modal.svelte.ts';
	import { clone } from '../../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import FloatWindow from '#lib/admin/kit/FloatWindow.svelte';
	import OutlinedField from '../../ui/OutlinedField.svelte';

	let { open = $bindable(false), postGroup, isSuperAdmin = false, onChange }: { open?: boolean; postGroup: any; isSuperAdmin?: boolean; onChange: (g: any) => void } = $props();

	let item = $state<any>(null);
	let orig = $state<any>(null);
	let innerWidth = $state(1280);
	const diffs = $derived(item && orig ? diffList(orig, item) : []);

	$effect(() => {
		if (!open) return;
		orig = clone(postGroup || {});
		item = clone(postGroup || {});
	});

	async function close() {
		if (diffs.length && !(await confirm('변경된 내용이 있습니다. 취소하시겠습니까?'))) return;
		open = false;
	}
	function apply() {
		if (!String(item.groupKey ?? '').trim()) return alert('게시판 groupKey를 입력하세요');
		if (!String(item.name ?? '').trim()) return alert('게시판 명을 입력하세요');
		onChange(clone(item));
		open = false;
	}
	const set = (patch: any) => (item = { ...item, ...patch });
	const cap = 'MuiTypography-root MuiTypography-caption';
</script>

<svelte:window bind:innerWidth onkeydown={(e) => open && e.key == 'Escape' && close()} />

{#if open && item}
	<div use:portal class={toClass({ position: 'fixed', inset: 0, zIndex: 1300, pointerEvents: 'none', '& > *': { pointerEvents: 'auto' } })}>
		<FloatWindow
			windowId="post-group-setting"
			windowTitle={`${item.name ? `"${item.name}" ` : ''}게시판 설정`}
			x={Math.max(16, Math.round((innerWidth - 620) / 2))}
			y={80}
			width={620}
			height={300}
			minWidth={420}
			minHeight={260}
			windowSx={{ backgroundColor: 'var(--adm-paper)', borderRadius: 12, overflow: 'hidden', boxShadow: '0 20px 60px rgba(16,24,40,0.25)', border: '1px solid var(--adm-divider, #e4e7ec)' }}
			onclose={close}
		>
			<div class={toClass({ display: 'flex', flexDirection: 'column', height: '100%' })}>
				<div class={toClass({ overflowY: 'auto', flex: 1, padding: '16px 24px' })}>
					<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, justifyContent: 'space-between' }))}>
							<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
								<span class={cap}>groupKey</span>
								<OutlinedField value={item.groupKey} disabled={!isSuperAdmin} oninput={(v) => set({ groupKey: v.toUpperCase() })} />
							</div>
							<div class={cx('MuiStack-root', toClass({ gap: 4, flexGrow: 1 }))}>
								<span class={cap}>구분</span>
								<OutlinedField fullWidth value={item.gubun || ''} oninput={(v) => set({ gubun: v.toUpperCase() })} />
							</div>
						</div>
						<div class={cx('MuiStack-root', toClass({ gap: 4 }))}>
							<span class={cap}>게시판 명</span>
							<OutlinedField value={item.name} disabled={!isSuperAdmin} oninput={(v) => set({ name: v })} />
						</div>
					</div>
				</div>
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'flex-end', gap: 8, padding: '12px 24px', borderTop: '1px solid var(--adm-divider, #e4e7ec)', flex: 'none' }))}>
					<Button size="small" variant="outlined" onclick={close}>취소</Button>
					<Button variant="contained" disabled={!diffs.length} onclick={apply}>적용</Button>
				</div>
			</div>
		</FloatWindow>
	</div>
{/if}
