<script lang="ts">
	/**
	 * 회원 정보 입력(React components/common/user/FormUserMemberDefault) — 이름·휴대폰·닉네임·한줄 소개·이메일·암호(이메일 가입만)
	 * · 생년월일/성별(useMemberDemographic) · 선택 약관 · 탈퇴 처리(정상 복귀·콘텐츠 숨김·전체 삭제·완전 삭제).
	 * 값은 onChange(키, 값, isJson) 로 올린다.
	 * (주소 입력(useMemberAddress)·권한 그룹(COMMON_CODES AUTH_GROUP) 은 즈비드가 안 써 옮기지 않았다)
	 */
	import { api } from '#lib/shared/api.ts';
	import { date } from '#lib/shared/format.ts';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import { alert, confirm } from '../../ui/modal.svelte.ts';
	import { adminImagePrefix, getMeta } from '../../meta.svelte.ts';
	import { getNumber, getPhoneNumberFormat, getTermPostGroup, socialUserGroupIdMap } from '../common.ts';
	import TextField from '../../ui/TextField.svelte';
	import PersonAvatar from '../ui/PersonAvatar.svelte';
	import OptionalTermsFilterForm from './OptionalTermsFilterForm.svelte';

	let {
		user,
		onChange,
		onDelete,
		isDrawer = false
	}: { user: any; onChange: (k: string, v: any, isJson?: boolean) => void; onDelete?: ((userId: number, isUnDelete: boolean | number) => void) | null; isDrawer?: boolean } = $props();

	const site = getMeta('SITE');
	const hasOptionalTerms = (getTermPostGroup()?.items || []).some((x: any) => x.agreementOptional);
	const social = $derived(socialUserGroupIdMap[user?.groupId]);

	async function hideOrDelete(actionType: string) {
		const o = await api.delete(`/service/admin/user/remove-user-data?userId=${user.userId}&actionType=${actionType}`);
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		alert('처리가 완료되었습니다.');
	}
	async function withdraw() {
		if (!(await confirm({ title: `${user?.name} 회원을 탈퇴처리`, body: `${user.name} 회원을 탈퇴처리 하시겠습니까?` }))) return;
		const o = await api.delete(`/service/admin/user?userId=${user?.userId}`);
		if (!o.ok) return alert('회원정보 삭제 도중 오류가 발생하였습니다.');
		await alert('정상적으로 탈퇴처리 되었습니다.');
		onDelete?.(user?.userId, false);
	}
	async function restore() {
		if (!(await confirm({ title: '정상회원으로 변경합니다.', body: `탈퇴처리된 ${user?.name} 회원정보를 정상회원으로 변경합니다.\n정상회원으로 변경하시겠습니까?` }))) return;
		const o = await api.delete(`/service/admin/user?userId=${user?.userId}&status=DO_NORMAL`);
		if (!o.ok) return alert('회원정보 삭제 도중 오류가 발생하였습니다.');
		await alert('회원정보가 복구구되었습니다.');
		onDelete?.(user?.userId, true);
	}
	async function hardDelete() {
		if (!(await confirm({ title: `${user?.name} 회원 완전 삭제`, body: `${user.name} 회원을 DB에서 완전히 삭제하시겠습니까?` }))) return;
		const o = await api.delete(`/service/admin/user?userId=${user?.userId}`);
		if (!o.ok) return alert('회원정보 삭제 도중 오류가 발생하였습니다.');
		await alert('정상적으로 삭제처리 되었습니다.');
		onDelete?.(user?.userId, -1000);
	}
</script>

<form class="fu" onsubmit={(e) => e.preventDefault()}>
	<div class="dd-card fu-card">
		<div class="fu-profile">
			<PersonAvatar size={64} src={`${adminImagePrefix(user.hostingId)}/u_${user.userId}_my/profile.webp`} />
			<div class="fu-who">
				{#if user.userId && user.name}
					<strong
						>{user.name}{#if user?.nickname && user.nickname != user.name}<span> · {user.nickname}</span>{/if}</strong
					>
				{:else}
					<TextField label="이름" value={user.name} size="small" fullWidth placeholder="이름을 입력하세요" oninput={(v) => onChange('name', v)} />
				{/if}
				<span class="fu-meta">
					{#if social}<span class="fu-chip">{social} 계정</span>{/if}
					{#if user?.status === -1}<span class="fu-chip fu-off">탈퇴</span>{/if}
					{#if user?.createdAt}가입 {date(user.createdAt)}{/if}
					{#if user?.updatedAt > 0 || (user?.updatedAt && typeof user.updatedAt == 'string')}
						· 마지막 로그인 {date(user.updatedAt)}{/if}
				</span>
			</div>
		</div>
	</div>

	<div class="dd-card fu-card">
		<h3 class="dd-card-title">기본 정보</h3>
		<div class="dd-grid">
			<TextField label="닉네임" value={user?.nickname || ''} size="small" fullWidth placeholder="닉네임" autocomplete="one-time-code" oninput={(v) => onChange('nickname', v)} />
			<TextField label="휴대폰" value={user.mobile} size="small" fullWidth placeholder="010-0000-0000" oninput={(v) => onChange('mobile', getPhoneNumberFormat(v))} />
			<TextField label="이메일 주소" value={user.email} size="small" fullWidth placeholder="이메일 주소" oninput={(v) => onChange('email', v)} />
			{#if [1, 2, 3].indexOf(user?.groupId) == -1}
				<TextField label="암호" type="password" value="" size="small" fullWidth placeholder="바꿀 때만 입력" autocomplete="one-time-code" oninput={(v) => onChange('password', v)} />
			{/if}
		</div>
		<TextField
			label="한줄 소개"
			value={user?.description || ''}
			size="small"
			fullWidth
			multiline
			minRows={2}
			placeholder="한줄 소개"
			autocomplete="one-time-code"
			oninput={(v) => onChange('description', v)}
		/>
		{#if site?.feature?.useMemberDemographic}
			<div class="fu-demo">
				<TextField label="생년월일" value={user?.Json?.birthDate || ''} size="small" fullWidth placeholder="예) 19900101" oninput={(v) => onChange('birthDate', getNumber(v), true)} />
				<div class="fu-gender">
					<Button variant={user?.Json?.gender == 'M' ? 'contained' : 'outlined'} onclick={() => onChange('gender', 'M', true)}>남자</Button>
					<Button variant={user?.Json?.gender == 'F' ? 'contained' : 'outlined'} onclick={() => onChange('gender', 'F', true)}>여자</Button>
				</div>
			</div>
		{/if}
	</div>

	{#if hasOptionalTerms}
		<div class="dd-card fu-card">
			<h3 class="dd-card-title">선택 약관 동의</h3>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row' }))}><OptionalTermsFilterForm agreeMap={user.Json?.agreeMap || {}} onChange={(am) => onChange('agreeMap', am, true)} /></div>
		</div>
	{/if}

	{#if onDelete}
		<div class="dd-card fu-card dd-danger">
			<h3 class="dd-card-title">회원 탈퇴 처리</h3>
			<p class="dd-help">탈퇴·콘텐츠 정리는 바로 반영되고, 콘텐츠 숨김·삭제는 되돌릴 수 없어요.</p>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, flexWrap: 'wrap' }))}>
				{#if user?.status !== -1}<Button variant="outlined" color="error" onclick={withdraw}>회원 탈퇴처리</Button>{/if}
				{#if user?.status === -1 || user?.status === 99}
					<Button variant="contained" onclick={restore}>정상회원 복귀</Button>
					<Button
						variant="outlined"
						color="error"
						onclick={async () =>
							(await confirm({
								title: `${user?.name}님이 작성한 콘텐츠의 프로필 정보를 "탈퇴회원"으로 변경합니다.`,
								body: '회원이 작성한 콘텐츠의 닉네임, 프로필이미지 정보가 "탈퇴회원" 으로 변경됩니다.\n처리 이후 데이터는 복구하실 수 없습니다.\n정말로 진행하시겠습니까?'
							})) && hideOrDelete('HIDE_USER_PROFILE')}>콘텐츠 개인정보 숨김</Button
					>
					<Button
						variant="outlined"
						color="error"
						onclick={async () =>
							(await confirm({
								title: `${user?.name}님이 작성한 콘텐츠를 모두 삭제합니다.`,
								body: '회원이 작성한 모든 콘텐츠(게시물 및 댓글)를 삭제처리합니다.\n처리 이후 데이터는 복구하실 수 없습니다.\n정말로 진행하시겠습니까?'
							})) && hideOrDelete('DELETE_ALL')}>콘텐츠 전체 삭제</Button
					>
					{#if user?.status == -1}<Button variant="contained" color="error" onclick={hardDelete}>회원 완전 삭제</Button>{/if}
				{/if}
			</div>
		</div>
	{/if}
</form>

<style>
	.fu {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	/* 서랍 밖(회원 상세 페이지)에서도 같은 카드 모양 — 서랍 안에서는 DrawerFrame 규칙이 같은 값을 준다 */
	.fu-card {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 20px;
		background: var(--zc-fff, #fff);
		border: 1px solid var(--adm-divider, #eaeaea);
		border-radius: 14px;
	}
	.fu-card :global(.dd-card-title) {
		margin: 0;
		font-size: 15px;
		font-weight: 700;
		line-height: 22px;
	}
	.fu-card.dd-danger {
		border-color: rgba(229, 57, 53, 0.35);
	}
	.fu-card.dd-danger :global(.dd-card-title) {
		color: #e53935;
	}
	.fu-card :global(.dd-grid) {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 16px;
	}
	.fu-card :global(.dd-help) {
		margin: 0;
		font-size: 12px;
		line-height: 17px;
		color: var(--gt-500, #999);
	}
	.fu-profile {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.fu-who {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}
	.fu-who strong {
		font-size: 17px;
		line-height: 24px;
	}
	.fu-who strong span {
		font-weight: 400;
		color: var(--gt-600, #777);
	}
	.fu-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: var(--gt-600, #777);
	}
	.fu-chip {
		padding: 1px 8px;
		border-radius: 10px;
		font-size: 12px;
		background: var(--gt-200, #f2f2f2);
		color: var(--gt-700, #555);
	}
	.fu-chip.fu-off {
		background: rgba(229, 57, 53, 0.1);
		color: #e53935;
	}
	.fu-demo {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.fu-gender {
		display: flex;
		flex-shrink: 0;
		gap: 4px;
	}
</style>
