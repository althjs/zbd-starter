<script lang="ts">
	/**
	 * 관리자 추가·수정 서랍(React setting/member/components/MemberManageDrawer).
	 * 이름(+ 프로필 사진) · 닉네임 · 직위·직책(json.org) · 이메일 · 관리자 그룹(+ 개별 권한) · 비밀번호 · 휴대폰 · AI 계정(개발자) · 메모 · 삭제.
	 * 저장: 수정 PUT /service/admin/member/{adminId} · 추가 POST /service/admin/member — json 은 문자열로(React 와 같은 본문).
	 * 열기: member 를 넘기면(새로 = { adminId: 0 }) 열리고, 기존이면 상세를 다시 읽는다.
	 */
	import { untrack } from 'svelte';
	import { api } from '#lib/shared/api.ts';
	import { number } from '#lib/shared/format.ts';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import Switch from '#lib/shared/mui/Switch.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import AuthButton from '../../ui/AuthButton.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import TextField from '../../ui/TextField.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { adminImagePrefix, feature, getMeta, isSuperAdmin } from '../../meta.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import { adminGroups } from '../adminGroups.svelte.ts';
	import { clone, getPhoneNumberFormat } from '../common.ts';
	import { diffList } from '#lib/shared/diff.ts';
	import DiffPreview from '#lib/shared/ui/DiffPreview.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import PersonAvatar from '../ui/PersonAvatar.svelte';
	import SelectField from '../ui/SelectField.svelte';
	import DrawerFrame from '../ui/DrawerFrame.svelte';
	import AuthSettingDrawer from './AuthSettingDrawer.svelte';
	import MemoHistory from './MemoHistory.svelte';

	let {
		member: input,
		hostingId,
		menus,
		onclose,
		onUpdated,
		onAdded
	}: { member: any; hostingId: string; menus: any[] | null; onclose: () => void; onUpdated?: (m: any) => void; onAdded?: (m: any) => void } = $props();

	const memberDefault = { adminId: 0, name: '', groupId: 0, email: '', password: '', mobile: '', status: 0, json: {} as any, hostingId: '' };
	let m = $state<any>(null);
	let orig = $state<any>(null);
	let memo = $state('');
	let authGroup = $state<any>(null);
	let jsonOpen = $state(false);
	const masterGroupIds: number[] = getMeta('COMMON', 'masterGroupIds') || [];
	const agencyAdminGroupIds: number[] = getMeta('SITE', 'feature', 'agencyAdminGroupIds') || [];
	const useESPost = true; // 닉네임 칸 — 스타터는 늘 보인다
	const sup = $derived(isSuperAdmin(admin.profile));

	$effect(() => {
		const x = input;
		untrack(() => {
			memo = '';
			if (!x) return (m = null);
			const base = x.adminId ? clone(x) : { ...clone(memberDefault), hostingId };
			m = base;
			orig = clone(base);
			if (x.adminId) load(x.adminId);
		});
	});
	async function load(adminId: number) {
		const o = await api.get(`/service/admin/member/${adminId}`, { headers: { cache: 'false' } });
		if (!o.ok) return alert('상세정보 로딩 도중 오류가 발생하였습니다.');
		const oo = o.data;
		oo.json = typeof oo.json == 'string' ? JSON.parse(oo.json) : oo.json || {};
		orig = clone(oo);
		m = oo;
	}
	const diffs = $derived(m && orig ? diffList(orig, m) : []);
	const edited = $derived(diffs.length > 0 || !!memo);
	const group = $derived((adminGroups.list || []).find((g) => g.groupId == m?.groupId) || null);
	const isAgency = $derived(agencyAdminGroupIds.includes(m?.groupId));
	const setJson = (patch: any) => (m = { ...m, json: { ...(m.json || {}), ...patch } });

	function memberSaveErrorMessage(msg: any, email?: string) {
		if (/ER_DUP_ENTRY|Duplicate entry/i.test(String(msg || ''))) return `이미 등록된 이메일(ID)입니다${email ? `: ${email}` : ''}. 다른 이메일을 사용하거나 기존 관리자를 수정하세요.`;
		return `처리 도중 오류가 발생하였습니다.${msg ? ` (${msg})` : ''}`;
	}
	async function save() {
		const data: any = clone(m);
		if (memo) (data.json.memos ||= []).push({ dt: new Date(), memo });
		data.json = JSON.stringify(data.json);
		if (!m.groupId) return alert('관리자 그룹을 선택하세요');
		if (!String(m.name || '').trim()) return alert('이름을 입력하세요');
		// 이메일 = 로그인 아이디 — 비우면 서버가 「잘못된 요청」 으로 거절한다
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(m.email || '').trim())) return alert('로그인에 쓸 이메일을 입력하세요');
		data.email = String(m.email).trim();
		if (!m.adminId && !m.password) return alert('비밀번호를 입력하세요');
		if (m.adminId) {
			const o = await api.put(`/service/admin/member/${m.adminId}`, data);
			if (!o.ok) return alert(memberSaveErrorMessage(o.error.message, data.email));
			toast('정상적으로 수정되었습니다.');
			const u = o.data;
			if (typeof u.json == 'string') u.json = JSON.parse(u.json);
			onUpdated?.(u);
			onclose();
			return;
		}
		data.hostingId = hostingId;
		const o = await api.post('/service/admin/member', data);
		if (!o.ok) return alert(memberSaveErrorMessage(o.error.message, data.email));
		toast('정상적으로 추가되었습니다.');
		const nm = o.data;
		onAdded?.(nm);
		const mm = { ...nm, json: typeof nm.json == 'string' ? JSON.parse(nm.json) : nm.json };
		m = mm;
		orig = clone(mm);
		onclose();
	}
	async function remove() {
		if (!(await confirm(`${m.name} 관리자를 삭제하시겠습니까?`))) return;
		const o = await api.delete(`/service/admin/member?adminId=${m.adminId}`);
		if (!o.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		toast('삭제되었습니다.');
		onUpdated?.({ ...m, status: -1 });
		onclose();
	}
	async function close() {
		if (edited && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		onclose();
	}
	const groupOptions = $derived([
		{ value: 0, label: '그룹 미지정', disabled: true },
		...(adminGroups.list || [])
			.filter((o) => !(masterGroupIds.indexOf(o.groupId) !== -1 && !sup))
			.filter((o) => !(hostingId === '' && !!o.hostingId))
			.filter((o) => !m?.hostingId || o.hostingId == m.hostingId)
			.map((o) => ({ value: o.groupId, label: o.name, disabled: o.groupId == 95 && m?.groupId == 95 }))
	]);
	// Claude 통합 플러그인이 붙었을 때만 「Claude 권한 관리」 — 플러그인 화면 파일이 있는지로 본다(빌드 때 정해짐)
	const hasClaude = Object.keys(import.meta.glob('../../collab/claude/ClaudeChat.svelte')).length > 0;
</script>

<Drawer open={!!input && !!m} anchor="right" onclose={close} paperSx={{ borderRadius: 0, width: '50%' }}>
	{#if m}
		<DrawerFrame title={m.adminId ? '관리자 수정' : '관리자 추가'} onclose={close}>
			<div class="dd-card">
				<div class="mm-profile">
					{#if m.adminId > 0}
						<ImagePicker
							size={{ width: 300, height: 300 }}
							image={m.json?.image}
							folder={`admin_profile_${m.adminId}`}
							{hostingId}
							useUnsplash
							onSelect={(img) => setJson({ image: { imageId: img.imageId, name: img.name } })}
						>
							{#snippet after()}<PersonAvatar size={64} sx={{ cursor: 'pointer' }} src={`${adminImagePrefix(hostingId)}/admin_profile_${m.adminId}/${m.json?.image?.name}`} />{/snippet}
							{#snippet before()}<PersonAvatar size={64} sx={{ cursor: 'pointer' }} />{/snippet}
						</ImagePicker>
					{:else}
						<PersonAvatar size={64} />
					{/if}
					<div class="mm-who">
						<strong
							>{m.name || '새 관리자'}{#if m.json?.nickname && m.json.nickname != m.name}<span> · {m.json.nickname}</span>{/if}</strong
						>
						<span>{[group?.name, m.email].filter(Boolean).join(' · ') || '이름·이메일·그룹을 입력하세요'}</span>
						{#if m.adminId > 0}<span class="mm-hint">사진을 눌러 프로필 이미지를 바꿔요</span>{/if}
					</div>
				</div>
			</div>

			<div class="dd-card">
				<h3 class="dd-card-title">기본 정보</h3>
				<div class="dd-grid">
					<TextField label="이름" value={m.name} size="small" fullWidth placeholder="관리자 성명" oninput={(v) => (m = { ...m, name: v })} id="member_name" />
					{#if useESPost}
						<TextField
							label="닉네임"
							value={m?.json?.nickname || ''}
							size="small"
							fullWidth
							placeholder="닉네임"
							maxLength={20}
							autocomplete="one-time-code"
							oninput={(v) => setJson({ nickname: v })}
						/>
					{/if}
					{#each ['position', 'duty'] as k (k)}
						<TextField
							label={k == 'position' ? '직위(직급)' : '직책'}
							value={m?.json?.org?.[k] || ''}
							size="small"
							fullWidth
							maxLength={30}
							autocomplete="off"
							placeholder={k == 'position' ? '예) 과장' : '예) 팀장'}
							oninput={(v) => {
								const org = { ...(m?.json?.org || {}), [k]: v };
								if (!v) delete org[k];
								setJson({ org });
							}}
						/>
					{/each}
					<TextField label="이메일" value={m.email} size="small" fullWidth placeholder="로그인·연락 이메일" oninput={(v) => (m = { ...m, email: v })} id="member_email" />
					<TextField label="휴대폰번호" value={m.mobile || ''} size="small" fullWidth placeholder="010-0000-0000" oninput={(v) => (m = { ...m, mobile: getPhoneNumberFormat(v) })} id="member_mobile" />
					{#if isAgency}
						<TextField label="대리점 추적 키" value={m.json?.agencyLinkKey || ''} size="small" fullWidth oninput={(v) => setJson({ agencyLinkKey: v })} />
					{/if}
				</div>
			</div>

			<div class="dd-card">
				<h3 class="dd-card-title">계정 · 권한</h3>
				<div class="dd-grid">
					<div class="mm-group">
						{#if !group && m.hostingId == 'SITE'}<span class="dd-help">사이트 마스터</span>{/if}
						<SelectField label="관리자 그룹" value={m.groupId} options={groupOptions} disabled={orig?.groupId == 95 && !admin.profile?.isDeveloper} onchange={(v) => (m = { ...m, groupId: v || 0 })} />
						{#if group}
							<Button size="small" onclick={() => (authGroup = group)}>
								<span class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'center', gap: 4 }))}><Icon name="admin_panel_settings" />개별 권한관리</span>
							</Button>
						{/if}
					</div>
					<TextField
						label="비밀번호"
						value={m.password || ''}
						type="password"
						size="small"
						autocomplete="one-time-code"
						fullWidth
						placeholder={m.adminId ? '바꿀 때만 입력' : '비밀번호'}
						oninput={(v) => (m = { ...m, password: v })}
						id="member_password"
					/>
				</div>
			</div>

			{#if sup && hasClaude}
				<div class="dd-card">
					<h3 class="dd-card-title">Claude 권한 관리</h3>
					<Switch
						checked={!!m.json?.claudeManager}
						label="업무 채팅 「관리자별 허용 권한」 을 설정할 수 있어요 — 개발자가 아니어도 다른 관리자에게 Claude 사용 단계를 주고 거둡니다."
						onchange={(v) => {
							const j = { ...(m.json || {}) };
							if (v) j.claudeManager = true;
							else delete j.claudeManager;
							m = { ...m, json: j };
						}}
					/>
				</div>
			{/if}

			{#if admin.profile?.isDeveloper}
				<div class="dd-card">
					<h3 class="dd-card-title">AI 계정</h3>
					<Switch
						checked={!!m.json?.isAI}
						label="AI 계정입니다 — 메신저로 온 메시지에 AI 가 답해요."
						onchange={(v) => {
							const j = { ...(m.json || {}) };
							if (v) j.isAI = true;
							else delete j.isAI;
							m = { ...m, json: j };
						}}
					/>
					{#if m.json?.isAI}
						<TextField
							value={m.json?.aiCharacter || ''}
							multiline
							minRows={4}
							fullWidth
							size="small"
							label="캐릭터 설정"
							maxLength={4000}
							placeholder="성격·말투·역할 예) 자신만만하고 솔직하다. 짧게 반말로 말하지만 일은 정확하다."
							oninput={(v) => setJson({ aiCharacter: v })}
							helperText="AI 가 답할 때 이 설정대로 말해요. 비워 두면 기본(즈비드 CTO 즈클) 성격."
						/>
					{/if}
				</div>
			{/if}

			<div class="dd-card">
				<h3 class="dd-card-title">메모</h3>
				<TextField
					label="새 메모"
					value={memo}
					multiline
					minRows={3}
					fullWidth
					size="small"
					maxLength={100}
					placeholder="필요한 사항을 메모하세요 — 저장할 때 기록에 더해져요"
					oninput={(v) => (memo = v)}
					helperText={`${number(memo.length) || 0}/100`}
					id="member_memo"
				/>
				<MemoHistory memos={m?.json?.memos} />
			</div>

			{#if diffs.length && admin.profile?.isDeveloper}<div class={cx('MuiGrid-root', toClass({ maxWidth: 1000, overflow: 'hidden' }))}><DiffPreview items={diffs} /></div>{/if}

			{#snippet footer()}
				{#if m.adminId > 0}<Button color="error" onclick={remove}>관리자 삭제</Button>{/if}
				<span class={toClass({ flex: 1 })}></span>
				{#if admin.profile?.isDeveloper}
					<IconButton aria-label="JSON 편집" onclick={() => (jsonOpen = true)}><Icon name="data_object" /></IconButton>
				{/if}
				<Button color="inherit" onclick={close}>취소</Button>
				<AuthButton variant="contained" sx={{ minWidth: 120 }} disabled={!edited} onclick={save}>저장</AuthButton>
			{/snippet}
		</DrawerFrame>
		<JsonEditDialog bind:open={jsonOpen} value={m} onChange={(o) => (m = o)} title="관리자 JSON" />
	{/if}
</Drawer>

<AuthSettingDrawer
	group={authGroup}
	{menus}
	{hostingId}
	member={m}
	onclose={() => (authGroup = null)}
	onChange={(items, groupItems) => {
		const flat = (arr: any[]) => arr.flatMap((g) => [g, ...(g.items || [])]);
		const gi = flat(groupItems);
		const authMenus = items.filter((o) => o.auth && gi.find((g) => g.menuId == o.menuId)?.auth != o.auth);
		setJson({ authMenus });
	}}
/>

<style>
	.mm-profile {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.mm-who {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}
	.mm-who strong {
		font-size: 17px;
		line-height: 24px;
	}
	.mm-who strong span {
		font-weight: 400;
		color: var(--gt-600, #777);
	}
	.mm-who > span {
		font-size: 13px;
		color: var(--gt-600, #777);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.mm-who .mm-hint {
		font-size: 12px;
		color: var(--gt-500, #999);
	}
	.mm-group {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.mm-group > :global(.MuiButton-root) {
		align-self: flex-start;
	}
</style>
