<script lang="ts">
	/**
	 * 게시판 검색 칸(React board/components/BoardSearchForm) — 작성일 기간(빠른 단추) + 검색어 + 공개상태 + 진행상태.
	 * 검색 엔진(useES) 게시판은 엔터·「검색」 때만 읽고, 아니면 글자를 칠 때마다 화면이 거른다.
	 */
	import Select from '#lib/shared/mui/Select.svelte';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { cx, sx as toClass, type Sx } from '#lib/shared/mui/sx.ts';
	import Button from '../../ui/Button.svelte';
	import { getThisMonthYMD, getThisWeekYMD, getYesterdayYMD, getYMD } from '../dates.ts';
	import DateField from '../ui/DateField.svelte';
	import { getPostGroup } from '../post/postGroup.ts';
	import type { BoardSearch } from '../types.ts';

	let {
		search,
		getData,
		useES = false,
		isUserPost = false,
		isPoll = false,
		isCommentSearch = false,
		isReport = false,
		isLoginDevice = false,
		sx = null
	}: {
		search: BoardSearch;
		getData: (s: BoardSearch) => void;
		useES?: boolean;
		isUserPost?: boolean;
		isPoll?: boolean;
		isCommentSearch?: boolean;
		isReport?: boolean;
		isLoginDevice?: boolean;
		sx?: Sx | null;
	} = $props();

	let s = $state<BoardSearch>({ ...search });
	$effect(() => {
		const next = { ...search };
		if (JSON.stringify(next) !== JSON.stringify($state.snapshot(s))) s = next;
	});
	const postGroup = $derived(getPostGroup(search.groupKey));

	const today = getYMD();
	const ranges = [
		{ label: '전체', r: { from: '2024-01-01', to: today } },
		{ label: '오늘', r: { from: today, to: today } },
		{ label: '이번주', r: getThisWeekYMD() },
		{ label: '이번달', r: getThisMonthYMD() },
		{ label: '지난 7일', r: { from: getYesterdayYMD(null, 7), to: today } },
		{ label: '지난 30일', r: { from: getYesterdayYMD(null, 30), to: today } }
	];
	function setRange(r: Partial<BoardSearch>) {
		s = { ...s, ...r, pageNo: 1 };
		getData({ ...s });
	}
	const enter = (e: KeyboardEvent) => {
		if (useES && e.key === 'Enter' && !e.isComposing) {
			e.preventDefault();
			getData({ ...s, pageNo: 1 });
		}
	};
	const statusOpts = $derived(postGroup?.statusMap ? [{ value: '', label: '전체' }, ...Object.keys(postGroup.statusMap).map((k) => ({ value: parseInt(k, 10), label: postGroup.statusMap![k] }))] : []);
	const row = toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' });
	const cap = toClass({ minWidth: 0, textAlign: 'center', '@media (min-width:600px)': { minWidth: 100 } });
	const kw = $derived(toClass({ minWidth: '100%', '@media (min-width:600px)': { minWidth: isCommentSearch ? 433 : 326 } }));
</script>

<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16, ...(sx || {}) }))}>
	<div class="MuiStack-root">
		<div class={cx('MuiStack-root', row)}>
			<span class={cx('MuiTypography-root MuiTypography-caption', cap)}>작성일</span>
			<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
				<DateField value={s.from} onchange={(v) => setRange({ from: v })} />
				~
				<DateField value={s.to} onchange={(v) => setRange({ to: v })} />
			</div>
			{#each ranges as b (b.label)}
				<Button size="small" variant={s.from === b.r.from && s.to === b.r.to ? 'outlined' : 'text'} onclick={() => setRange(b.r)}>{b.label}</Button>
			{/each}
		</div>
	</div>

	{#if !isLoginDevice}
		<div class={cx('MuiStack-root', row, toClass({ marginTop: 8 }))}>
			{#if useES && !isCommentSearch && !isReport}
				<Select
					value={s.keywordType || ''}
					sx={{ minWidth: 100 }}
					options={[
						{ value: '', label: '게시글' },
						{ value: 'COMMENT', label: '댓글' }
					]}
					onchange={(v) => (s = { ...s, keywordType: v, pageNo: 1 })}
				/>
			{/if}
			<div class={kw} role="presentation" onkeydown={enter}>
				<TextField
					size="small"
					value={s.keyword || ''}
					placeholder={s.keywordType == 'COMMENT' ? '댓글내용을 입력하세요' : '제목, 본문 검색 키워드를 입력하세요'}
					oninput={(v) => {
						s = { ...s, keyword: v.substring(0, 20), pageNo: 1 };
						if (!useES) getData({ ...s });
					}}
				/>
			</div>
			{#if useES && isUserPost}
				<div class={toClass({ minWidth: 150 })} role="presentation" onkeydown={enter}>
					<TextField size="small" value={s.nickname || ''} placeholder="작성자 닉네임" oninput={(v) => (s = { ...s, nickname: v.substring(0, 20), pageNo: 1 })} />
				</div>
			{/if}
			{#if useES}<Button variant="contained" onclick={() => getData({ ...s, pageNo: 1 })}>검색</Button>{/if}
			{#if statusOpts.length}
				<div class="MuiStack-root">
					<Select value={typeof s.status == 'number' ? s.status : ''} sx={{ minWidth: 100 }} options={statusOpts} onchange={(v) => getData({ ...s, status: v === '' ? null : v, pageNo: 1 })} />
				</div>
			{/if}
			{#if isPoll}
				<Select
					value={s.scheduleStatus || ''}
					sx={{ minWidth: 100 }}
					options={[
						{ value: '', label: '진행상태 전체' },
						{ value: '진행중', label: '진행중' },
						{ value: '대기', label: '대기' },
						{ value: '종료', label: '종료' }
					]}
					onchange={(v) => {
						s = { ...s, scheduleStatus: v, pageNo: 1 };
						if (!useES) getData({ ...s });
					}}
				/>
			{/if}
			{#if (postGroup?.items?.length || 0) > 10}
				<Select
					value={s.groupSubKey || ''}
					sx={{ minWidth: 200 }}
					options={[{ value: '', label: '하위분류 선택' }, ...postGroup.items!.map((x) => ({ value: x.groupSubKey, label: x.name || x.groupSubKey }))]}
					onchange={(v) => getData({ ...s, groupSubKey: v, pageNo: 1 })}
				/>
			{/if}
		</div>
	{/if}
</div>
