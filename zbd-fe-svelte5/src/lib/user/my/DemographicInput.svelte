<script module lang="ts">
	export function isValidBirthDate(v: string) {
		if (!/^\d{8}$/.test(v || '')) return false;
		const y = +v.slice(0, 4),
			m = +v.slice(4, 6),
			d = +v.slice(6, 8);
		const dt = new Date(y, m - 1, d);
		return y >= 1900 && dt.getFullYear() == y && dt.getMonth() == m - 1 && dt.getDate() == d && dt <= new Date();
	}
	export const demographicOk = (j: any) => isValidBirthDate(String(j?.birthDate || '')) && (j?.gender == 'M' || j?.gender == 'F');
</script>

<script lang="ts">
	/**
	 * 생년월일(8자리 숫자) + 성별(남/여) — 사이트 설정 「생년월일/성별」(useMemberDemographic) 일 때 가입·내 정보에.
	 * 값은 회원 json 의 birthDate·gender('M'|'F'). 서버는 birthDate 가 없으면 부가정보 필수 화면으로 보낸다(Go member.checkRequiredUserInfo).
	 */
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import TextField from '#lib/shared/mui/TextField.svelte';

	let { birthDate = '', gender = '', onChange }: { birthDate?: string; gender?: string; onChange: (v: { birthDate: string; gender: string }) => void } = $props();
	const help = $derived(birthDate && !isValidBirthDate(birthDate) ? '생년월일은 8자리 숫자로 입력해주세요.' : '');
</script>

<Stack>
	<Stack direction="row" sx={{ gap: 8, alignItems: 'center' }}>
		<div role="presentation" style="flex:1;display:flex" oninput={(e) => onChange({ birthDate: (e.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 8), gender })}>
			<TextField value={birthDate} size="small" placeholder="생년월일 (예: 19900101)" sx={{ flex: 1 }} />
		</div>
		<MuiButton variant={gender == 'M' ? 'contained' : 'outlined'} sx={{ minWidth: 56, height: 40 }} onclick={() => onChange({ birthDate, gender: 'M' })}>남자</MuiButton>
		<MuiButton variant={gender == 'F' ? 'contained' : 'outlined'} sx={{ minWidth: 56, height: 40 }} onclick={() => onChange({ birthDate, gender: 'F' })}>여자</MuiButton>
	</Stack>
	{#if help}<p class="MuiTypography-root MuiTypography-body2" style="color:#f04438;margin-top:4px">{help}</p>{/if}
</Stack>
