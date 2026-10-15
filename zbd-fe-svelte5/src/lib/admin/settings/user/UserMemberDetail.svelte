<script lang="ts">
	/**
	 * 회원 상세 편집(React components/common/user/UserMemberDetail) — 입력 폼 + 바뀐 값 미리보기 + 저장(확인 창).
	 * 서랍 안(isDrawer)이면 DrawerFrame(머리 ✕ · 아래 고정 버튼 줄) 으로 감싼다. 개발자는 JSON 직접 고치기.
	 * 저장: PUT /service/admin/user/{id} (새 회원은 POST /service/admin/user) — 본문은 json 을 문자열로(React 와 같게).
	 */
	import { api } from '#lib/shared/api.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { clone } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import DrawerFrame from '../ui/DrawerFrame.svelte';
	import FormUserMember from './FormUserMember.svelte';

	let {
		user,
		onChange,
		onDelete = null,
		onSave = null,
		onAdd = null,
		isDrawer = false,
		onclose,
		ignoreEmail = false
	}: {
		user: any;
		onChange: (u: any) => void;
		onDelete?: ((id: number, un: boolean | number) => void) | null;
		onSave?: ((u: any) => void) | null;
		onAdd?: ((u: any) => void) | null;
		isDrawer?: boolean;
		/** 서랍 안이면 머리의 ✕ · 취소 */
		onclose?: () => void;
		ignoreEmail?: boolean;
	} = $props();

	const first = () => clone(user);
	let orig = $state(first());
	let jsonOpen = $state(false);
	const diffs = $derived(diffList(orig, user));

	function change(k: string, v: any, isJson = false) {
		const o = { ...user };
		if (isJson) o.Json = { ...(user.Json || { imageId: '', memo: '' }), [k]: v };
		else o[k] = v;
		onChange(o);
	}

	async function save() {
		if (!(await confirm('저장하시겠습니까?'))) return;
		const data = { ...user };
		if (ignoreEmail && !String(data.email || '').trim()) {
			data.email = ' ';
			if (!data.socialId) data.socialId = 'x_' + Date.now().toString(36);
		}
		data.json = JSON.stringify(data.Json);
		delete data.Json;
		const o = user.userId ? await api.put(`/service/admin/user/${user.userId}`, data) : await api.post('/service/admin/user', data);
		if (!o.ok) return alert('저장 도중 오류가 발생하였습니다.');
		const oo = o.data;
		oo.Json = JSON.parse(oo.json);
		delete oo.json;
		if (!user.userId) onAdd?.(oo);
		toast('정상적으로 저장되었습니다.');
		orig = clone(oo);
		onChange({ ...oo });
		onSave?.({ ...oo });
	}
</script>

{#snippet actions()}
	{#if admin.profile?.isDeveloper}
		<IconButton aria-label="JSON 편집" onclick={() => (jsonOpen = true)}><Icon name="data_object" /></IconButton>
	{/if}
	<span class={toClass({ flex: 1 })}></span>
	{#if isDrawer && onclose}<Button color="inherit" onclick={onclose}>취소</Button>{/if}
	<Button variant="contained" sx={{ minWidth: 120 }} disabled={!diffs.length} onclick={save}>저장</Button>
{/snippet}

{#snippet body()}
	<FormUserMember {user} onChange={change} {onDelete} {isDrawer} />
	{#if diffs.length && admin.profile?.isDeveloper}
		<div class={cx('MuiGrid-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>
	{/if}
{/snippet}

{#if isDrawer}
	<DrawerFrame title={`회원정보 ${user?.userId ? '수정' : '추가'}`} onclose={() => onclose?.()} footer={actions}>{@render body()}</DrawerFrame>
{:else}
	<div class={cx('MuiStack-root', toClass({ gap: 16 }))}>
		{@render body()}
		<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 8 }))}>{@render actions()}</div>
	</div>
{/if}

<JsonEditDialog bind:open={jsonOpen} value={user} onChange={(o) => onChange(o)} title="회원 JSON" />
