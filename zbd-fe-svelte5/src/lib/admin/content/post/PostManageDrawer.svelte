<script lang="ts">
	/**
	 * 게시글 편집(React post/components/PostManageDrawer + PostManageForm) — 아래에서 올라오는 전체 화면 서랍.
	 *   <PostManageDrawer bind:this={pm} groupKey="TERMS" useCreateButton onSave={(p) => …} />
	 *   pm.openDrawer(post)   — 저장된 글이면 `GET /service/admin/post/{postId}` 로 새로 읽어 연다
	 *   <PostManageDrawer inline bind:this={pm} … /> + pm.openDrawer(post) — 서랍 없이 화면 안에(글 편집 페이지)
	 * 저장 `POST /service/admin/post/{postId}`(json·jsonUser 는 문자열) · 삭제 `DELETE /service/admin/post/{postId}`.
	 *
	 * 옮긴 칸: 분류(그룹·하위·상단고정·댓글/별점/좋아요·베스트) · 제목·보조제목·작성자·썸네일 · 시행일(TERMS)·처리상태(PROCESS) ·
	 *   부가정보 · 다중 답변 · 본문(TinyMCE + 이미지 첨부) · 관리자 답변 · 파일첨부 · 외부링크 · SEO 설명 · 상태 · 관리자 메모·삭제.
	 * 안 옮긴 칸(즈비드 게시판이 안 씀): 일정·가격·다국어·라벨·고정코드·연관글·견적·CSS/JS·디자인 편집기·JSON 편집기·바뀐 곳 미리보기.
	 */
	import { customHostingIdOf, hostingIdOf } from '#lib/shared/config.ts';
	import { api } from '#lib/shared/api.ts';
	import TinyMceEditor from '#lib/shared/editor/TinyMceEditor.svelte';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import Select from '#lib/shared/mui/Select.svelte';
	import Switch from '#lib/shared/mui/Switch.svelte';
	import SwitchField from '../ui/SwitchField.svelte';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { cx, sx as toClass } from '#lib/shared/mui/sx.ts';
	import { adminImagePrefix } from '../../meta.svelte.ts';
	import { media } from '../../media.svelte.ts';
	import { activeNav } from '../../nav.svelte.ts';
	import { admin } from '../../profile.svelte.ts';
	import Loading from '../../layout/Loading.svelte';
	import Button from '../../ui/Button.svelte';
	import Icon from '../../ui/Icon.svelte';
	import IconButton from '../../ui/IconButton.svelte';
	import { alert, confirm, toast } from '../../ui/modal.svelte.ts';
	import { cloneDeep, getNextId, getYMD } from '../dates.ts';
	import { secureDelete, secureDownload, secureUpload } from '../files.ts';
	import DateField from '../ui/DateField.svelte';
	import OutlinedInput from '../ui/OutlinedInput.svelte';
	import JsonEditDialog from '../ui/JsonEditDialog.svelte';
	import type { Post, PostGroup, PostGroupSub } from '#lib/shared/types/post.ts';
	import PostReplyThread from './PostReplyThread.svelte';
	import { adminPostGroups, dataColumnOf, getPostGroup, postDefault, statusMapDefault, userPostGroups } from './postGroup.ts';
	import { getAdminPost } from './postStore.svelte.ts';
	import { SK } from '#lib/shared/storage-keys.ts';

	let {
		onSave,
		useCreateButton = false,
		buttonText = '신규 글 추가',
		groupKey = '',
		groupSubKey = '',
		readonly = false,
		subGroups = null,
		statusMap = statusMapDefault,
		customPostGroup = null,
		masterGroupIds = null,
		inline = false,
		oncancel
	}: {
		onSave?: (p: Post) => void;
		useCreateButton?: boolean;
		buttonText?: string;
		groupKey?: string;
		groupSubKey?: string;
		readonly?: boolean;
		subGroups?: PostGroupSub[] | null;
		statusMap?: Record<string, string> | null;
		customPostGroup?: PostGroup | null;
		masterGroupIds?: number[] | null;
		/** 서랍 없이 화면 안에 그린다(글 편집 페이지 board/{groupKey}/editor/{postId}) */
		inline?: boolean;
		/** 닫기(inline 이면 화면을 떠날 때) */
		oncancel?: () => void;
	} = $props();

	/** React 관리자 편집기 도구 막대(배경색 포함). 「HTML 정리」(cleanHtmlButton) 단추는 공용 편집기에 없어 빠짐 */
	const TOOLBAR = `imagePickerButton linkAttachButton | undo redo | styles styleselect | fontselect fontsize | bold italic backcolor forecolor  | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent |  code  codesample | link emoticons  fullscreen pagebreak visualblocks visualchars  table`;
	const IGNORE_ETC = ['body', 'memo', 'history', 'image', 'images', 'productId', 'name', 'rating'];
	const etcInfoMap: Record<string, string> = { contactName: '접수자 성명', contactMobileNo: '접수자 휴대폰', contactEmail: '접수자 이메일' };

	let show = $state(false);
	let loading = $state(false);
	let post = $state<Post | null>(null);
	let orig = $state('');
	let html = $state('');
	let htmlOrig = $state('');
	let pg = $state<PostGroup>({ groupKey: '', dataColumn: 'json' });
	let editorKey = $state(0);
	let picker = $state<any>();
	let editor = $state<any>();
	let collapsed = $state<Record<string, boolean>>({});
	let jsonOpen = $state(false);

	const profile = $derived(admin.profile);
	const dc = $derived(dataColumnOf(pg));
	const smap = $derived(statusMap || statusMapDefault);
	const isMaster = $derived(!masterGroupIds || masterGroupIds.indexOf(profile?.groupId as number) !== -1);
	const dirty = $derived(!!post && (JSON.stringify(post) !== orig || html != htmlOrig));
	const folder = $derived(post ? `${post.userId ? `u_${post.userId}_` : ''}post_${post.postId}` : '');

	function groupOf(p: Post | null) {
		return customPostGroup || getPostGroup(p?.groupKey || (groupKey || '').toUpperCase());
	}

	function apply(oo: any) {
		oo.json = oo.json ? (typeof oo.json == 'string' ? JSON.parse(oo.json || '{}') : oo.json) : { body: '', name: '' };
		oo.jsonUser = oo.jsonUser ? (typeof oo.jsonUser == 'string' ? JSON.parse(oo.jsonUser || '{}') : oo.jsonUser) : { body: '', name: '' };
		const col = dataColumnOf(pg);
		if (typeof oo[col]?.sortNo !== 'undefined') oo.sortNo = oo[col].sortNo;
		post = oo;
		orig = JSON.stringify(oo);
		html = oo[col]?.body || oo.json.body || '';
		htmlOrig = html;
		editorKey++;
	}

	/** 목록·상세가 부른다 — 새 글(createdAt 없음)이면 바로, 아니면 서버에서 다시 읽어 연다 */
	export async function openDrawer(target: Post) {
		pg = groupOf(target);
		try {
			collapsed = JSON.parse(localStorage.getItem(SK.admin.postCollapse(pg.groupKey)) || '{}') || {};
		} catch {}
		if (!target?.createdAt) {
			apply(cloneDeep(target));
			show = true;
			return;
		}
		loading = true;
		const oo = await getAdminPost(target.postId);
		loading = false;
		if (!oo) {
			if (target.status === 99) {
				apply(cloneDeep(target));
				show = true;
			} else alert('데이터 조회 도중 오류가 발생하였습니다.');
			return;
		}
		apply(oo);
		show = true;
	}

	/** 이미 읽어 온 글로 바로 채운다(React formRef.init) */
	export function init(target: Post, group: PostGroup | null = null) {
		pg = group || groupOf(target);
		apply(cloneDeep(target));
		show = true;
	}

	function newPost() {
		const g = customPostGroup || getPostGroup((groupKey || '').toUpperCase());
		const m: Post = { ...postDefault(), postId: Date.now().toString(36), groupKey: (groupKey || '').toUpperCase(), groupSubKey: groupSubKey || '' };
		m[dataColumnOf(g)].name = profile?.nickname || profile?.name || '';
		if (g?.isAdminReplyEmailNoti) {
			m.json.email = profile?.email;
			m.json.adminHost = location.hostname;
		}
		if (!g?.isBoardStore || g?.isAdminReply) m.status = 99;
		openDrawer(m);
	}

	async function close() {
		if (dirty && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		show = false;
		post = null;
		html = htmlOrig = '';
		oncancel?.();
	}

	function setField(k: string, v: any) {
		if (!post) return;
		if (k === 'groupKey') post.groupSubKey = '';
		(post as any)[k] = v;
	}
	function setJson(k: string, v: any, col = '') {
		if (!post) return;
		const key = (col || dc) as 'json' | 'jsonUser';
		post[key] = { ...(post[key] || {}), [k]: v };
	}

	async function doSave(override?: Post) {
		const data: any = $state.snapshot(override?.postId ? override : post);
		if (!data) return;
		if (!data.createdAt) data.createdAt = Date.now();
		if (pg?.feature?.useCreatedAtEdit && !data.json?.createdAtReal) data.json.createdAtReal = data.createdAt || Date.now();
		if (((subGroups?.length || 0) > 0 || (pg?.items?.length || 0) > 0) && !data.groupSubKey && !data.json?.isFixTop) return alert('하위 분류를 선택하세요');
		if (!String(data.title || '').trim()) return alert('제목을 입력하세요');
		if (!override) data[dc].body = html;
		if (!data.customHostingId && typeof customHostingIdOf(pg) == 'string') data.customHostingId = customHostingIdOf(pg);
		data.json = JSON.stringify(data.json || {});
		data.jsonUser = JSON.stringify(data.jsonUser || {});
		loading = true;
		const o = await api.post(`/service/admin/post/${data.postId}`, data);
		loading = false;
		if (!o.ok) return alert('처리 도중 오류가 발생하였습니다.');
		toast('정상적으로 저장 되었습니다.');
		apply(o.data);
		onSave?.(cloneDeep(o.data));
	}

	async function del() {
		if (!post || !(await confirm(`글을 삭제하시겠습니까?`))) return;
		const o = await api.delete(`/service/admin/post/${post.postId}`);
		if (!o.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		toast('정상적으로 삭제되었습니다.');
		onSave?.({ ...$state.snapshot(post), DELETED: true } as Post);
		orig = JSON.stringify(post);
		htmlOrig = html;
		close();
	}

	function toggle(key: string) {
		collapsed = { ...collapsed, [key]: !collapsed[key] };
		try {
			localStorage.setItem(SK.admin.postCollapse(pg.groupKey), JSON.stringify(collapsed));
		} catch {}
	}

	/** React: 읽기 전용·특정 그룹·하위분류가 넘어온 화면이면 그룹을 못 바꾼다 */
	const groupLocked = $derived(readonly || ['QNA', 'C_PRODUCT_INQUIRY', 'SURVEY', 'EVENT'].indexOf(post?.groupKey || '') !== -1 || subGroups !== null);
	const groups = $derived(customPostGroup ? [customPostGroup] : dc == 'jsonUser' ? userPostGroups() : adminPostGroups());
	const etcKeys = $derived(Object.keys(post?.jsonUser || {}).filter((k) => IGNORE_ETC.indexOf(k) == -1));
	const subItems = $derived(pg?.items || subGroups || []);

	async function attachFile(e: Event) {
		const f = (e.currentTarget as HTMLInputElement).files?.[0];
		(e.currentTarget as HTMLInputElement).value = '';
		if (!f || !post) return;
		const cid = typeof customHostingIdOf(pg) != 'undefined' ? customHostingIdOf(pg) || 'OCTOPUS' : profile?.hostingId || 'OCTOPUS';
		const o = await secureUpload(f, folder, cid);
		if (!o.ok) return alert('업로드 도중 오류가 발생하였습니다.');
		const files = [...(post[dc]?.files || [])];
		if (!files.find((x) => x.fileName == o.data.fileName)) files.push(o.data);
		setJson('files', files);
	}
	async function removeFile(f: any) {
		if (!post || !(await confirm(`${f.fileName} 파일을 삭제하시겠습니까?`))) return;
		const o = await secureDelete(`${hostingIdOf(f) ? `${hostingIdOf(f)}/` : ''}${folder}`, f.fileName);
		if (!o.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		setJson(
			'files',
			(post[dc]?.files || []).filter((x) => x.fileName != f.fileName)
		);
		toast('정상적으로 삭제되었습니다.');
	}

	function onBodyImages(imgs: any) {
		const list = Array.isArray(imgs) ? imgs : [imgs];
		let tags = list.map((img) => `<img src="${adminImagePrefix()}/${folder}/${img?.name}"/>`).join('');
		if (list.length) tags = '<br />' + tags + '<br />';
		if (editor) editor.insert(tags);
		else html = html + tags;
		picker?.close();
	}

	const linkSx = toClass({ position: 'relative', paddingRight: 48, gap: 4 });
	const sectionHead = toClass({ flexDirection: 'row', alignItems: 'center', cursor: 'pointer', marginBottom: 8, width: 'fit-content' });
	const gray = toClass({ color: 'var(--gt-700, #909090)' });
</script>

{#snippet section(key: string, title: string, useCollapse: boolean, body: any)}
	<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', alignItems: 'start', width: '100%' }))}>
		<div class={cx('MuiStack-root', toClass({ flexGrow: 1, marginBottom: 24 }))}>
			{#if useCollapse}
				<div class={cx('MuiStack-root', sectionHead)} role="button" tabindex="0" onclick={() => toggle(key)} onkeydown={(e) => e.key == 'Enter' && toggle(key)}>
					{#if title}<h6 class="MuiTypography-root MuiTypography-subtitle2">{title}</h6>{/if}
					<Icon name={collapsed[key] ? 'expand_more' : 'expand_less'} sx={{ fontSize: 20, marginLeft: 4 }} />
				</div>
				{#if !collapsed[key]}{@render body()}{/if}
			{:else}
				{#if title}<h6 class={cx('MuiTypography-root MuiTypography-subtitle2', toClass({ marginBottom: 8 }))}>{title}</h6>{/if}
				{@render body()}
			{/if}
		</div>
	</div>
{/snippet}

{#if useCreateButton && !inline}
	<Button variant="contained" disabled={readonly} sx={{ position: 'fixed', bottom: 20, right: 20, zIndex: 10 }} onclick={newPost}>{buttonText}</Button>
{/if}

{#snippet fields()}
	{#if post}
		{@const p = post}
		<div class="MuiStack-root">
			{#if !p.jsonUser?.requestForm}
				{#snippet groupBody()}
					<div class={cx('MuiStack-root', toClass({ flexDirection: media.mdDown ? 'column' : 'row', gap: 8, flexWrap: 'wrap', alignItems: media.mdDown ? 'stretch' : 'center' }))}>
						<Select
							value={p.groupKey || ''}
							class={groupLocked ? 'Mui-disabled' : ''}
							sx={{
								minWidth: 300,
								...(groupLocked
									? {
											pointerEvents: 'none',
											'& .MuiSelect-select': { color: 'var(--adm-text-disabled)', WebkitTextFillColor: 'var(--adm-text-disabled)' },
											'& .MuiSvgIcon-root': { color: 'var(--adm-text-disabled)' }
										}
									: {})
							}}
							options={[{ value: '', label: '분류를 선택하세요', disabled: true }, ...groups.map((g) => ({ value: g.groupKey, label: g.name || g.groupKey }))]}
							onchange={(v) => {
								pg = getPostGroup(v);
								setField('groupKey', v);
							}}
						/>
						{#if !subItems.length}
							<div class={toClass({ minWidth: 300 })}>
								<TextField size="small" placeholder="서브그룹키" value={p.groupSubKey || ''} disabled={readonly} oninput={(v) => setField('groupSubKey', v.substring(0, 50))} />
							</div>
						{:else}
							<Select
								value={p.groupSubKey || ''}
								sx={{ minWidth: 300 }}
								options={[{ value: '', label: '하위분류 선택' }, ...subItems.map((x) => ({ value: x.groupSubKey, label: x.name || x.groupSubKey }))]}
								onchange={(v) => setField('groupSubKey', v)}
							/>
						{/if}
						{#if p.groupKey == 'NOTICE' || pg?.feature?.useFixTop}
							<SwitchField variant="body2" checked={!!p.json?.isFixTop} label="상단고정" onchange={(v) => setJson('isFixTop', v, 'json')} />
							{#if pg?.feature?.useMoreContent}<SwitchField variant="body2" checked={!!p.json?.useMoreContent} label="더 보기" onchange={(v) => setJson('useMoreContent', v, 'json')} />{/if}
							{#if p.json?.isFixTop && pg?.feature?.useComment}
								<SwitchField variant="body2" checked={!!p.json?.useComment} label="댓글 사용" onchange={(v) => setJson('useComment', v, 'json')} />
								<SwitchField variant="body2" checked={!!p.json?.useStarRating} label="별점 사용" onchange={(v) => setJson('useStarRating', v, 'json')} />
								<SwitchField variant="body2" checked={!!p.json?.useLike} label="좋아요 사용" onchange={(v) => setJson('useLike', v, 'json')} />
							{/if}
						{/if}
						{#if pg?.feature?.useBest}<SwitchField variant="body2" checked={!!(p as any)[dc]?.isBest} label="베스트" onchange={(v) => setJson('isBest', v)} />{/if}
					</div>
				{/snippet}
				{@render section('groupKey', '포스트 그룹', false, groupBody)}
			{/if}

			{#snippet titleBody()}
				<!-- React 는 이 줄만 휴대폰에서도 가로(부르는 쪽이 isMobile 을 넘기지 않음) — 같게 -->
				<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 16, justifyContent: 'space-between', alignItems: 'end' }))}>
					<div class={cx('MuiStack-root', toClass({ gap: 8, width: `calc(100%${pg?.feature?.useThumbnail ? ' - 220px' : ''})` }))}>
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 4, alignItems: 'center' }))}>
							<p class="MuiTypography-root MuiTypography-body2">제목</p>
							<p class={cx('MuiTypography-root MuiTypography-body1', toClass({ color: 'var(--adm-error-main)' }))}>*</p>
						</div>
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8 }))}>
							<TextField size="small" value={p.title} disabled={readonly} oninput={(v) => setField('title', v.substring(0, 100))} />
							{#if pg?.feature?.useSubTitle}<TextField
									size="small"
									value={(p as any)[dc]?.subTitle || ''}
									placeholder="보조타이틀을 입력하세요"
									disabled={readonly}
									oninput={(v) => setJson('subTitle', v.substring(0, 100))}
								/>{/if}
							<OutlinedInput sx={{ minWidth: 180 }} value={(p as any)[dc]?.name || ''} disabled={readonly} oninput={(v) => setJson('name', v.substring(0, 20))}>
								{#snippet start()}<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ minWidth: 'fit-content', marginRight: 8, color: 'var(--gt-700, #909090)' }))}>작성자</span
									>{/snippet}
							</OutlinedInput>
						</div>
					</div>
					{#if pg?.feature?.useThumbnail}
						<div class={cx('MuiStack-root', toClass({ minWidth: 210 }))}>
							<ImagePicker
								title="썸네일을 선택하세요"
								size={{ width: 'all', height: 'auto' }}
								image={(p as any)[dc]?.image || null}
								folder={`post_${p.postId}`}
								autoCrop
								maxWidth={1200}
								useOriginalFileName
								userId={p.userId || null}
								onSelect={(img: any) => setJson('image', { imageId: img?.imageId || `${p.userId ? `u_${p.userId}_` : ''}post_${p.postId}`, name: img?.name, hostingId: img?.hostingId })}
							/>
						</div>
					{/if}
				</div>
			{/snippet}
			{@render section('title', '', false, titleBody)}

			{#if pg?.component && (!pg.component.condition || p.json?.[pg.component.condition]) && isMaster && (pg.component.name == 'TERMS' || pg.component.name == 'PROCESS')}
				{#snippet compBody()}
					{#if pg.component?.name == 'TERMS'}
						<DateField width={140} value={p.json?.revisionAt || getYMD()} onchange={(v) => setJson('revisionAt', v)} />
					{:else}
						{@const pmap = pg.component?.processMap || { 0: '신규 접수', 1: '답변 완료' }}
						{@const cur = (dc == 'jsonUser' ? p.jsonUser : p.json)?.processStatus || 0}
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', flexWrap: 'wrap' }))} role="radiogroup">
							{#each Object.keys(pmap) as step (step)}
								<label class={cx('MuiFormControlLabel-root', toClass({ display: 'inline-flex', alignItems: 'center', marginRight: 16, cursor: 'pointer', gap: 4, fontSize: 14 }))}>
									<input type="radio" name="processStatus" checked={String(cur) == step} onchange={() => setJson('processStatus', Number(step))} />
									{pmap[step]}
								</label>
							{/each}
						</div>
					{/if}
				{/snippet}
				{@render section('additionalForm', pg.component?.title || '', true, compBody)}
			{/if}

			{#if !p.jsonUser?.requestForm && etcKeys.length > 0}
				{#snippet etcBody()}
					<div class="MuiStack-root">
						{#each etcKeys as k (k)}
							{@const v = (p.jsonUser as any)[k]}
							{#if typeof v == 'boolean'}
								<div class={cx('MuiStack-root', toClass({ fontSize: 13 }))}>{etcInfoMap[k] || k} : {v ? 'Y' : 'N'}</div>
							{:else if typeof v == 'object' && v !== null && !Array.isArray(v)}
								<div class={cx('MuiStack-root', toClass({ marginBottom: 12, padding: 8, border: '1px solid var(--gt-300, #e5e5e5)', borderRadius: 4, gap: 6 }))}>
									<p class={cx('MuiTypography-root MuiTypography-body2', toClass({ fontWeight: 600 }))}>{etcInfoMap[k] || k}</p>
									{#if Object.keys(v).length === 0}<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>(비어있음)</span>{/if}
									{#each Object.entries(v) as [sk, sv] (sk)}
										{#if typeof sv === 'boolean'}
											<Switch checked={sv} label={etcInfoMap[sk] || sk} onchange={(x) => setJson(k, { ...v, [sk]: x }, 'jsonUser')} />
										{:else if typeof sv === 'string' || typeof sv === 'number' || sv === null}
											<div class="MuiStack-root">
												<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{etcInfoMap[sk] || sk}</span>
												<TextField
													size="small"
													type={typeof sv === 'number' ? 'number' : 'text'}
													value={String(sv ?? '')}
													disabled={readonly}
													oninput={(raw) => setJson(k, { ...v, [sk]: typeof sv === 'number' ? (raw === '' ? 0 : Number(raw)) : raw }, 'jsonUser')}
												/>
											</div>
										{:else}
											<div class={cx('MuiStack-root', toClass({ fontSize: 12 }))}>
												<span class={cx('MuiTypography-root MuiTypography-caption', gray)}>{etcInfoMap[sk] || sk}</span>
												<span class={cx('MuiTypography-root MuiTypography-caption', toClass({ fontFamily: 'monospace', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }))}>{JSON.stringify(sv)}</span>
											</div>
										{/if}
									{/each}
								</div>
							{:else}
								<div class={cx('MuiStack-root', toClass({ fontSize: 13 }))}>{etcInfoMap[k] || k}: {v}</div>
							{/if}
						{/each}
					</div>
				{/snippet}
				{@render section('etcInfo', '부가정보', false, etcBody)}
			{/if}

			{#if pg?.isAdminMultipleReply && !!p.adminId && p.adminId != profile?.adminId}
				{#snippet replyBody()}
					{#if p.createdAt && profile}
						<PostReplyThread
							post={p}
							{profile}
							readonly={activeNav()?.auth != 'RW'}
							onSave={async (patch) => {
								const cur = $state.snapshot(p) as Post;
								const next = { ...cur, json: { ...cur.json, ...patch } };
								post = next;
								await doSave(next);
							}}
						/>
					{:else}
						<p class={cx('MuiTypography-root MuiTypography-body2', gray)}>글을 저장한 뒤 답변을 등록할 수 있습니다.</p>
					{/if}
				{/snippet}
				{@render section('replies', '답변', true, replyBody)}
			{/if}

			{#if !p.jsonUser?.requestForm}
				{#snippet bodyBody()}
					{#if pg.isAdminReply && (masterGroupIds || []).indexOf(profile?.groupId as number) !== -1}
						<div class={cx('MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root', toClass({ padding: 16 }))}>{@html (p as any)[dc]?.body || ''}</div>
					{:else}
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, flexGrow: 1 }))}>
							<div style="display:none">
								<ImagePicker
									bind:this={picker}
									title="본문에 사용할 이미지를 선택하세요"
									size={{ width: 'all', height: 'auto' }}
									image={null}
									autoCrop
									maxWidth={1200}
									folder={`post_${p.postId}`}
									userId={p.userId || null}
									isMultiple
									multiple
									useOriginalFileName
									useDelete
									useUnsplash
									onSelect={onBodyImages}
								/>
							</div>
							{#key editorKey}
								<div class={toClass({ width: '100%' })}>
									<TinyMceEditor
										bind:this={editor}
										customToolbar={TOOLBAR}
										body={html}
										height={innerHeight - 600 > 1000 ? innerHeight - 1000 : 700}
										placeholder={pg?.feature?.placeholder || ''}
										imageId={folder}
										userId={p.userId || null}
										onImageUI={() => picker?.open()}
										onchange={(h) => (html = h)}
									/>
								</div>
							{/key}
						</div>
					{/if}
				{/snippet}
				{@render section('body', pg?.isAdminReply ? '문의 내용' : '본문', false, bodyBody)}
			{/if}

			{#if pg?.isAdminReply && !pg?.isAdminMultipleReply && isMaster}
				{#snippet adminReplyBody()}
					<TextField
						size="small"
						multiline
						minRows={3}
						disabled={readonly}
						value={p.json?.adminReply || ''}
						oninput={(v) => {
							const json: any = { ...(p.json || {}), adminReply: v };
							if (!json.adminReplyAdminId) {
								json.adminReplyTS = Date.now();
								json.adminReplyAdminId = profile?.adminId;
								json.adminReplyAdminName = profile?.name || '';
							}
							p.json = json;
						}}
					/>
				{/snippet}
				{@render section('adminReply', '관리자 답변', false, adminReplyBody)}
			{/if}

			{#if pg?.useAttachFile || pg?.feature?.useAttachFile}
				{#snippet fileBody()}
					<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap', padding: '8px 0' }))}>
						<label
							class={cx(
								'MuiButtonBase-root MuiButton-root MuiButton-contained MuiButton-sizeMedium',
								toClass({ height: 38, color: '#667085', borderRadius: 2, background: '#F2F4F7', fontWeight: 700, boxShadow: 'none', cursor: 'pointer', width: 'fit-content' })
							)}
						>
							파일첨부
							<input type="file" hidden accept={pg?.fileAttachAccept || '.jpg,.png,.webp,.jpeg,.pdf,.xlsx,.zip,.pptx,.doc,.hwp,.doc,.docx'} onchange={attachFile} />
						</label>
						{#each (p as any)[dc]?.files || [] as f (f.fileName)}
							<Button size="small" onclick={() => secureDownload(`${hostingIdOf(f) ? `${hostingIdOf(f)}/` : ''}${folder}`, f.fileName)}>
								<Icon name="download" />{f.fileName}
								<Icon name="close" size="Small" role="button" tabindex={0} onclick={(e: MouseEvent) => (e.preventDefault(), e.stopPropagation(), removeFile(f))} />
							</Button>
						{/each}
					</div>
				{/snippet}
				{@render section('useAttachFile', '파일첨부', false, fileBody)}
			{/if}

			{#if pg?.feature?.useLink}
				{#snippet linkBody()}
					<div class={cx('MuiStack-root', toClass({ gap: 8 }))}>
						{#each (p as any)[dc]?.links || [] as link (link.id)}
							<div class={cx('MuiStack-root', linkSx)}>
								<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, alignItems: 'center' }))}>
									<TextField
										size="small"
										value={link.url || ''}
										placeholder="https://"
										oninput={(v) =>
											setJson(
												'links',
												((p as any)[dc]?.links || []).map((o: any) => (o.id == link.id ? { ...o, url: v } : o))
											)}
									/>
									<SwitchField
										variant="body2"
										checked={link.target == '_blank'}
										label="새창"
										onchange={(v) =>
											setJson(
												'links',
												((p as any)[dc]?.links || []).map((o: any) => (o.id == link.id ? { ...o, target: v ? '_blank' : '_self' } : o))
											)}
									/>
								</div>
								<IconButton
									aria-label="링크 빼기"
									sx={{ position: 'absolute', right: 0, top: 0 }}
									onclick={() =>
										setJson(
											'links',
											((p as any)[dc]?.links || []).filter((o: any) => o.id != link.id)
										)}><Icon name="close" /></IconButton
								>
							</div>
						{/each}
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', gap: 8, marginTop: 8 }))}>
							<Button onclick={() => setJson('links', [...((p as any)[dc]?.links || []), { id: getNextId((p as any)[dc]?.links), url: '', target: '_blank' }])}>링크 추가</Button>
						</div>
					</div>
				{/snippet}
				{@render section('links', '외부링크', false, linkBody)}
			{/if}

			{#if pg?.feature?.useSeoDescription}
				{#snippet seoBody()}<TextField multiline minRows={2} value={p.json?.seoDescription || ''} oninput={(v) => (p.json = { ...(p.json || {}), seoDescription: v })} />{/snippet}
				{@render section('seoDescription', 'SEO 설명', false, seoBody)}
			{/if}

			{#if !pg?.disableOpenStatus}
				{#snippet useBody()}
					<Select
						value={p.status}
						sx={{ minWidth: 300, width: 'fit-content' }}
						options={[{ value: '', label: '상태값을 선택하세요', disabled: true }, ...Object.keys(smap).map((k) => ({ value: Number(k), label: smap[k] }))]}
						onchange={(v) => !readonly && setField('status', parseInt(String(v), 10))}
					/>
				{/snippet}
				{@render section('use', '상태', false, useBody)}
			{/if}

			{#if isMaster}
				{#snippet memoBody()}
					<TextField size="small" multiline minRows={3} disabled={readonly} value={p.json?.memo || ''} oninput={(v) => (p.json = { ...(p.json || {}), memo: v })} />
					{#if p.createdAt && activeNav()?.auth == 'RW'}
						<div class={cx('MuiStack-root', toClass({ flexDirection: 'row', justifyContent: 'space-between', marginTop: 8 }))}>
							<Button variant="contained" disabled={dirty} onclick={del}>글 삭제</Button>
						</div>
					{/if}
				{/snippet}
				{@render section('memo', '관리자 메모', false, memoBody)}
			{/if}
		</div>
	{/if}
{/snippet}

{#snippet saveBar(fixed: boolean)}
	<div class={cx('MuiStack-root', toClass(fixed ? { zIndex: 100, background: 'var(--zc-fff, #fff)', bottom: 0, left: 0, right: 0, position: 'fixed' } : { position: 'relative' }))}>
		<Button variant="contained" fullWidth sx={{ height: 48, borderRadius: 0 }} disabled={!readonly && !dirty} onclick={() => doSave()}>저장</Button>
		{#if profile?.isDeveloper}
			<IconButton aria-label="JSON 편집" sx={{ position: fixed ? 'fixed' : 'absolute', right: 10, bottom: 5, color: '#fff' }} onclick={() => (jsonOpen = true)}><Icon name="data_object" /></IconButton>
		{/if}
	</div>
{/snippet}

{#if inline}
	{@render fields()}
	{#if post}{@render saveBar(false)}{/if}
{:else}
	<Drawer open={show} anchor="bottom" onclose={close} paperSx={{ width: '100%', height: '100%' }}>
		<IconButton aria-label="닫기" onclick={close} sx={{ position: 'fixed', zIndex: 10000, ...(media.mdDown ? { right: 0, top: 0 } : { right: 20, top: 10 }) }}
			><Icon name="close" sx={{ fontSize: 30 }} /></IconButton
		>
		<div class={cx('MuiStack-root', toClass({ padding: media.mdDown ? 0 : 24 }))}>
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root',
					toClass({ display: 'block', padding: media.mdDown ? 16 : 24, ...(media.mdDown ? { paddingTop: 40 } : {}) })
				)}
			>
				{@render fields()}
			</div>
			<div class={toClass({ height: 50 })}></div>
		</div>
		{@render saveBar(true)}
	</Drawer>
{/if}

<JsonEditDialog open={jsonOpen} value={post} onclose={() => (jsonOpen = false)} onchange={(v) => (post = v)} />

{#if loading}
	<div class={toClass({ position: 'fixed', inset: 0, zIndex: 1400 })}><Loading /></div>
{/if}
