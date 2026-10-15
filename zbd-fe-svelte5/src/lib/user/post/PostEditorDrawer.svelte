<script lang="ts">
	/**
	 * 글쓰기·글 수정 서랍 — 예전 pages/user/post/components/UserPostManageDrawer.tsx 의 창 부분.
	 * 사용자 화면 묶음 레이아웃에 하나만 둔다(열기는 editor.svelte.ts 의 openPostEditor).
	 *
	 * 칸: [분류(개발자·GNB 글쓰기)] 제목 · 썸네일 · 슬라이드 이미지 · 별점 → 하위분류 → 연락처(푸터 문의) → 본문(편집기) → 상태(개발자)
	 *     아래: 약관 동의 · 작성 안내 / 화면 아래 고정 「저장」
	 * 바뀐 것이 없으면 저장을 막고, 바뀐 채 닫으면 묻는다.
	 * (예전 창 아래의 JSON 바뀐 곳 미리보기(DiffPreview) 는 옮기지 않았다 — 개발용 표시였다)
	 */
	import { goto } from '$app/navigation';
	import { api } from '#lib/shared/api.ts';
	import { HOSTING_ID, hostingIdOf, imagePrefix } from '#lib/shared/config.ts';
	import TinyMceEditor from '#lib/shared/editor/TinyMceEditor.svelte';
	import HtmlView from '#lib/shared/html/HtmlView.svelte';
	import ImagePicker from '#lib/shared/image/ImagePicker.svelte';
	import { viewport } from '#lib/shared/media.svelte.ts';
	import Drawer from '#lib/shared/mui/Drawer.svelte';
	import Icon from '#lib/shared/mui/Icon.svelte';
	import IconButton from '#lib/shared/mui/IconButton.svelte';
	import MuiButton from '#lib/shared/mui/MuiButton.svelte';
	import Select from '#lib/shared/mui/Select.svelte';
	import Stack from '#lib/shared/mui/Stack.svelte';
	import TextField from '#lib/shared/mui/TextField.svelte';
	import { cx, sx as toClass, up } from '#lib/shared/mui/sx.ts';
	import { site } from '#lib/shared/site.svelte.ts';
	import { alert, confirm, toast } from '#lib/shared/ui/dialog.svelte.ts';
	import Agreement from '../Agreement.svelte';
	import { session } from '../session.svelte.ts';
	import { closePostEditor, normalize, postEditor, userPostStatusMap } from './editor.svelte.ts';
	import { getPostGroupAuth } from './postAuth.ts';

	const isMobile = $derived(viewport.isMobile);
	const profile = $derived(session.profile as any);
	const userId = $derived(profile?.userId || null);
	const post = $derived(postEditor.post);
	const userGroups = $derived<any[]>(site.base?.postGroupsMeta?.userPostGroups || []);
	const postGroup = $derived(userGroups.find((x) => x.groupKey == post?.groupKey) || null);
	const groupAuth = $derived(profile?.userId && post?.groupKey ? getPostGroupAuth(profile, post.groupKey) : { AUTH: 'R', groupSubKeys: [] as string[] });
	const imageId = $derived(`u_${userId}_post_${post?.postId}`);

	let html = $state('');
	let editor = $state<any>();
	let bodyPicker = $state<any>();
	let multiPicker = $state<any>();
	let showSave = $state(false);

	// 다른 글을 열면 본문을 그 글로
	$effect(() => {
		void postEditor.seq;
		html = postEditor.post?.jsonUser?.body || '';
		showSave = false;
		const t = setTimeout(() => (showSave = true), 500);
		return () => clearTimeout(t);
	});

	const changed = $derived(!!post && (JSON.stringify(post) != postEditor.orig || html != (post?.jsonUser?.body || '')));

	let maxEditorHeight = $derived.by(() => {
		let h = viewport.height > 1000 ? viewport.height - 500 - ((postGroup?.feature?.agreementGroupSubKeys || []).length > 0 ? 100 : 0) : 500;
		return Math.min(h, 800);
	});

	function setJsonUser(patch: Record<string, any>) {
		postEditor.post.jsonUser = { ...(postEditor.post.jsonUser || {}), ...patch };
	}
	const col = () => postGroup?.dataColumn || 'json';
	function setCol(patch: Record<string, any> | null, del = '') {
		const x = { ...(postEditor.post[col()] || {}) };
		if (patch) Object.assign(x, patch);
		if (del) delete x[del];
		postEditor.post[col()] = x;
	}

	const imgUrl = (img: any, fallbackFolder = imageId) => `${imagePrefix(hostingIdOf(img) || HOSTING_ID)}/${img?.imageId || fallbackFolder}/${img?.name}`;
	const imgTag = (name: string) => `<img src="${imagePrefix()}/${imageId}/${name}" style="max-width:100%;height:auto" />`;

	function onThumb(img: any) {
		if (!img) return;
		setJsonUser({ image: { name: img.name, hostingId: HOSTING_ID, imageId } });
		// 썸네일을 고르면 본문 커서 자리에도 같은 그림을 넣는다(예전과 같게 — 통째로 바꾸지 않는다)
		if (editor) editor.insert(imgTag(img.name));
		else html = html + imgTag(img.name);
	}

	function onBodyImages(imgs: any) {
		const list: any[] = Array.isArray(imgs) ? imgs : imgs ? [imgs] : [];
		let tags = '';
		for (const img of list) {
			if (img?.posterImgge) {
				tags += `<video controls src="${imagePrefix()}/${imageId}/${img.name}" poster="${imagePrefix()}/${imageId}/${img.posterImgge}" width="100%" height="auto"></video>`;
			} else {
				tags += `<img src="${imagePrefix()}/${imageId}/${img?.name}"/>`;
			}
		}
		const first = list[0];
		if (postGroup?.feature?.useThumbnail && !post?.jsonUser?.image?.name && first) {
			setJsonUser({ image: { name: first.name, hostingId: HOSTING_ID, imageId } });
		}
		if (editor) editor.insert(tags);
		else html = (html || '') + '<br/>' + tags;
		bodyPicker?.close();
	}

	const phone = (v: string) => {
		const d = String(v || '')
			.replace(/\D/g, '')
			.slice(0, 11);
		if (d.length < 4) return d;
		if (d.length < 8) return `${d.slice(0, 3)}-${d.slice(3)}`;
		if (d.length == 10) return `${d.slice(0, 3)}-${d.slice(3, 6)}-${d.slice(6)}`;
		return `${d.slice(0, 3)}-${d.slice(3, 7)}-${d.slice(7)}`;
	};

	async function save() {
		const data: any = $state.snapshot(post);
		if (postGroup?.items?.length > 0 && !data.groupSubKey) return alert('하위분류를 선택하세요');
		if (postGroup?.feature?.useCreatedAtEdit && !data.json?.createdAtReal) data.json = { ...(data.json || {}), createdAtReal: Date.now() };
		const jsonUser = { ...(data.jsonUser || {}) };
		jsonUser.name = (profile?.nickname || '').trim() || (profile?.name || 'G').substring(0, 1) + '**';
		jsonUser.memo = data.memo;
		jsonUser.body = html;
		if (!data.groupKey) return alert('분류를 선택하세요');
		if (!String(data.title || '').trim()) return alert('제목을 입력하세요');
		if (!String(jsonUser.body || '').trim()) return alert('본문을 입력하세요');
		if ((postGroup?.feature?.agreementGroupSubKeys || []).length > 0 && !jsonUser.agreement) return alert('약관동의는 필수입니다.');
		if (data.isFooterInquiry && (!jsonUser.contactName || !jsonUser.contactMobileNo || !jsonUser.contactEmail)) return alert('연락받으실 정보를 입력하세요');

		const guest = !!data.isFooterInquiry;
		data.jsonUser = JSON.stringify(jsonUser);
		data.json = JSON.stringify(data.json || {});
		delete data.memo;
		delete data.isNew;
		const r = await api.post(`/service/user/post/${guest ? 'guest/' : ''}${data.postId}`, data);
		if (!r.ok) return alert('처리 도중 오류가 발생하였습니다.');
		const saved = normalize(r.data || { ...data });
		postEditor.onSave?.(saved);
		const wasGnb = postEditor.isGnbPostWriteWidget;
		const groupKey = data.groupKey;
		postEditor.orig = JSON.stringify(postEditor.post);
		if (guest) {
			await alert('정상적으로 접수되었습니다. 빠른 시일 내에 연락 드리겠습니다.');
		} else {
			toast('정상적으로 저장 되었습니다.');
		}
		closePostEditor();
		if (wasGnb) setTimeout(() => goto(`/board/${String(groupKey).toLowerCase()}`), 800);
	}

	async function remove() {
		if (!(await confirm('글을 삭제하시겠습니까?'))) return;
		const r = await api.delete(`/service/user/post/${post.postId}`);
		if (!r.ok) return alert('삭제 도중 오류가 발생하였습니다.');
		toast('정상적으로 삭제되었습니다.');
		postEditor.onSave?.({ ...$state.snapshot(post), DELETED: true });
		postEditor.orig = JSON.stringify(post);
		html = post?.jsonUser?.body || '';
		closePostEditor();
	}

	async function handleClose() {
		if (changed && !(await confirm('작성된 내용이 있습니다. 취소하시겠어요?'))) return;
		closePostEditor();
	}

	const groupOptions = $derived([
		{ value: '', label: '분류를 선택하세요', disabled: true },
		...userGroups.filter((o) => !postEditor.isGnbPostWriteWidget || getPostGroupAuth(profile, o.groupKey).AUTH == 'RW').map((o) => ({ value: o.groupKey, label: o.name, caption: o.gubun || '' }))
	]);
	const subOptions = $derived([
		{ value: '', label: '하위분류를 선택하세요', disabled: true },
		...(postGroup?.items || []).map((s: any) => ({ value: s.groupSubKey, label: s.name || s.groupSubKey, disabled: !groupAuth.groupSubKeys.includes(s.groupSubKey) }))
	]);
	const statusOptions = Object.keys(userPostStatusMap).map((k) => ({ value: k, label: userPostStatusMap[k] }));
	const images = $derived<any[]>(post?.[postGroup?.dataColumn || 'json']?.images || []);
	const rating = $derived(Number(post?.[postGroup?.dataColumn || 'json']?.rating) || 0);

	function rate(e: MouseEvent, i: number) {
		const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const v = e.clientX - r.left < r.width / 2 ? i - 0.5 : i;
		setCol({ rating: v });
	}

	const label = toClass({ width: 130, marginBottom: 8, fontSize: 14, [up('md')]: { fontSize: 16 } });
	const row = toClass({ flexGrow: 1, marginBottom: 16, [up('md')]: { marginBottom: 32 } });
</script>

{#snippet field(name: string, required: boolean, body: import('svelte').Snippet)}
	<Stack direction="row" sx={{ alignItems: 'start', width: '100%' }}>
		<div class={cx('MuiStack-root', row)}>
			<h6 class={cx('MuiTypography-root MuiTypography-h6', label)}>
				{#if required}<span style="display:inline-flex;gap:4px;align-items:center">{name} <span class="MuiTypography-root" style="color:#f04438">*</span></span>{:else}{name}{/if}
			</h6>
			{@render body()}
		</div>
	</Stack>
{/snippet}

{#snippet groupField()}
	<Stack direction="row" sx={{ gap: 8 }}>
		<Select sx={{ minWidth: 300 }} value={post.groupKey || ''} options={groupOptions} onchange={(v) => ((postEditor.post.groupKey = v), (postEditor.post.groupSubKey = ''))} />
		{#if !postEditor.isGnbPostWriteWidget}
			<TextField size="small" fullWidth={false} placeholder="서브 그룹 키" maxLength={20} bind:value={postEditor.post.groupSubKey} />
		{/if}
	</Stack>
{/snippet}

{#snippet titleField()}
	<Stack direction={isMobile ? 'column' : 'row'} sx={{ gap: 16, justifyContent: 'space-between', alignItems: isMobile ? 'start' : 'end', position: 'relative' }}>
		<Stack sx={{ flexGrow: 1, width: '100%' }}>
			<Stack direction="row" sx={{ gap: 4, alignItems: 'center' }}>
				<h6 class={cx('MuiTypography-root MuiTypography-h6', toClass({ fontSize: 14, [up('md')]: { fontSize: 16 } }))}>제목</h6>
				<span class="MuiTypography-root" style="color:#f04438">*</span>
			</Stack>
			<TextField size="small" fullWidth={isMobile} maxLength={100} bind:value={postEditor.post.title} />
		</Stack>

		{#if postGroup?.feature?.useThumbnail}
			<Stack sx={{ minWidth: 210 }}>
				{#if isMobile}<span class="MuiTypography-root MuiTypography-caption">썸네일</span>{/if}
				<ImagePicker
					title="썸네일을 선택하세요"
					size={{ width: 'all', height: 'auto' }}
					image={post.jsonUser?.image || null}
					disableCrop
					folder={`post_${post.postId}`}
					useOriginalFileName
					useDelete
					useUnsplash
					{userId}
					onSelect={onThumb}
				>
					{#snippet before()}
						<div class="thumb empty" class:float={!isMobile}><span class="MuiTypography-root MuiTypography-caption">썸네일 이미지를 선택하세요</span></div>
					{/snippet}
					{#snippet after()}
						<img class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img thumb" class:float={!isMobile} src={imgUrl(post.jsonUser?.image)} alt="" />
					{/snippet}
				</ImagePicker>
			</Stack>
		{/if}

		{#if postGroup?.feature?.useMultipleImages}
			<Stack sx={{ minWidth: 210, flexGrow: 1 }}>
				<span class="MuiTypography-root MuiTypography-caption" style="color:#6c737f;margin-bottom:8px">슬라이드 이미지 ({images.length})</span>
				{#if images.length > 0}
					<Stack direction="row" sx={{ flexWrap: 'wrap', gap: 4, marginBottom: 8 }}>
						{#each images as img, idx (img?.name + '_' + idx)}
							<div class="MuiBox-root" style="position:relative;width:70px;height:70px">
								<img class="MuiCardMedia-root MuiCardMedia-media MuiCardMedia-img slide" src={imgUrl(img, post.userId ? `u_${post.userId}_post_${post.postId}` : `post_${post.postId}`)} alt="" />
								<button type="button" class="slide-x" aria-label="삭제" onclick={() => setCol({ images: images.filter((_, i) => i != idx) })}><Icon name="close" sx={{ fontSize: 14 }} /></button>
							</div>
						{/each}
					</Stack>
				{/if}
				<Stack direction="row" sx={{ gap: 8, alignSelf: 'flex-start' }}>
					<MuiButton size="small" variant="outlined" onclick={() => multiPicker?.open()}><Icon name="add_photo_alternate" size="Small" sx={{ marginRight: 4 }} />이미지 추가</MuiButton>
				</Stack>
				<div style="display:none">
					<ImagePicker
						bind:this={multiPicker}
						folder={`post_${post.postId}`}
						size={{ width: 'all', height: 'auto' }}
						maxWidth={1500}
						isMultiple
						multiple
						useUnsplash
						useDelete
						autoCrop
						userId={post.userId || userId}
						onSelect={(imgs: any) => {
							const added = (Array.isArray(imgs) ? imgs : [imgs]).filter(Boolean).map((o: any) => ({ name: o.name, hostingId: hostingIdOf(o) || HOSTING_ID, imageId }));
							setCol({ images: [...images, ...added] });
							multiPicker?.close();
						}}
					/>
				</div>
			</Stack>
		{/if}

		{#if postGroup?.feature?.useReviewStarRating}
			<Stack sx={{ minWidth: 210 }}>
				<span class="MuiTypography-root MuiTypography-caption" style="color:#6c737f;margin-bottom:4px">별점 {rating > 0 ? `(${rating.toFixed(1)})` : ''}</span>
				<Stack direction="row" sx={{ alignItems: 'center', gap: 8 }}>
					<span class="MuiRating-root" role="radiogroup">
						{#each [1, 2, 3, 4, 5] as i (i)}
							<button type="button" class="star" aria-label="{i}점" onclick={(e) => rate(e, i)}>
								<span class="material-icons">{rating >= i ? 'star' : rating >= i - 0.5 ? 'star_half' : 'star_border'}</span>
							</button>
						{/each}
					</span>
					{#if rating > 0}
						<IconButton size="Small" title="별점 삭제" onclick={() => setCol(null, 'rating')}><Icon name="close" size="Small" /></IconButton>
					{/if}
				</Stack>
			</Stack>
		{/if}
	</Stack>
{/snippet}

{#snippet subField()}
	<Select value={post.groupSubKey || ''} options={subOptions} onchange={(v) => (postEditor.post.groupSubKey = v)} />
{/snippet}

{#snippet contactField()}
	<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root" style="padding:16px">
		<Stack direction="row" sx={{ gap: 8 }}>
			<Stack>
				<span class="MuiTypography-root MuiTypography-subtitle2">문의자 성명</span>
				<TextField size="small" placeholder="성함을 입력하세요" maxLength={20} value={post.jsonUser?.contactName || ''} oninput={(v: string) => setJsonUser({ contactName: v })} />
			</Stack>
			<Stack>
				<span class="MuiTypography-root MuiTypography-subtitle2">문의자 휴대폰</span>
				<TextField size="small" placeholder="휴대폰 번호를 입력하세요" value={post.jsonUser?.contactMobileNo || ''} oninput={(v: string) => setJsonUser({ contactMobileNo: phone(v) })} />
			</Stack>
			<Stack>
				<span class="MuiTypography-root MuiTypography-subtitle2">문의자 이메일</span>
				<TextField size="small" placeholder="이메일 주소를 입력하세요" maxLength={60} value={post.jsonUser?.contactEmail || ''} oninput={(v: string) => setJsonUser({ contactEmail: v })} />
			</Stack>
		</Stack>
	</div>
{/snippet}

{#snippet bodyField()}
	{#if !post.isFooterInquiry}
		<div style="display:none">
			<ImagePicker
				bind:this={bodyPicker}
				title="본문에 사용할 이미지를 선택하세요"
				size={{ width: 'all', height: 'auto' }}
				autoCrop
				maxWidth={1200}
				folder={`post_${post.postId}`}
				isMultiple
				useOriginalFileName
				useDelete
				useUnsplash
				{userId}
				onSelect={onBodyImages}
			/>
		</div>
	{/if}
	{#key postEditor.seq}
		<TinyMceEditor
			bind:this={editor}
			body={post.jsonUser?.body || ''}
			onImageUI={post.isFooterInquiry ? null : () => bodyPicker?.open()}
			height={maxEditorHeight}
			onchange={(h) => (html = h)}
			placeholder={postEditor.placeholder || postGroup?.feature?.placeholder || ''}
			{imageId}
			{userId}
			{isMobile}
			customToolbar={isMobile
				? 'imagePickerButton linkAttachButton | undo redo | fontselect fontsize | bold italic forecolor  | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent'
				: ''}
		/>
	{/key}
	{#if post.createdAt && userId == post.userId}
		<Stack direction="row" sx={{ marginTop: 13 }}>
			<MuiButton variant="contained" disabled={changed} onclick={remove}>글 삭제</MuiButton>
		</Stack>
	{/if}
{/snippet}

{#snippet statusField()}
	<Select sx={{ minWidth: 300 }} value={String(post.status ?? '')} options={statusOptions} onchange={(v) => (postEditor.post.status = Number(v) || 0)} />
{/snippet}

<Drawer open={postEditor.open} anchor={isMobile ? 'bottom' : 'right'} onclose={handleClose} paperSx={{ width: isMobile ? '100%' : '50%' }}>
	<IconButton sx={{ position: 'fixed', right: 20, top: 10, zIndex: 10000 }} onclick={handleClose}><Icon name="close" sx={{ fontSize: 24 }} /></IconButton>
	{#if post}
		<Stack sx={{ padding: isMobile ? 16 : 24 }}>
			<h3 class="MuiTypography-root MuiTypography-{isMobile ? 'h5' : 'h3'}" style="text-align:center;margin:16px 0">
				{postEditor.headTitle || userGroups.find((x) => x.groupKey == post.groupKey)?.name || ''}
			</h3>
			<div
				class={cx(
					'MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root post-card',
					toClass(isMobile ? { padding: 0, boxShadow: 'none !important' } : { padding: 16, overflow: 'visible', [up('md')]: { padding: 24 } })
				)}
			>
				{#if profile?.isDeveloper || postEditor.isGnbPostWriteWidget}{@render field('포스트 그룹', false, groupField)}{/if}
				{@render field('', false, titleField)}
				{#if postGroup?.items?.length > 0}{@render field('하위분류', false, subField)}{/if}
				{#if post.isFooterInquiry}{@render field('연락받으실 정보', false, contactField)}{/if}
				{@render field('본문', true, bodyField)}
				{#if profile?.isDeveloper}{@render field('상태', false, statusField)}{/if}
			</div>

			<Stack direction={isMobile ? 'column' : 'row'} sx={{ gap: 16, marginTop: 16 }}>
				{#if (postGroup?.feature?.agreementGroupSubKeys || []).length > 0}
					<div style="min-width:40%;flex-grow:1">
						<p class="MuiTypography-root MuiTypography-body2 agreement-title" style="display:flex;gap:4px">약관 동의 <span style="color:#f04438">*</span></p>
						<Agreement hiddenTitle groupSubKeys={postGroup.feature.agreementGroupSubKeys} onchange={(all) => setJsonUser({ agreement: all })} />
					</div>
				{/if}
				{#if postGroup?.feature?.tipHTML}
					<Stack class="tip">
						<p class="MuiTypography-root MuiTypography-body2 tip-title" style="display:flex;gap:4px">{postGroup.feature.tipTitle || '작성안내'}</p>
						<div class="MuiPaper-root MuiPaper-elevation MuiPaper-rounded MuiPaper-elevation1 MuiCard-root" style="padding:8px 16px 8px 0">
							<HtmlView html={postGroup.feature.tipHTML} sx={{ fontSize: 13 }} />
						</div>
					</Stack>
				{/if}
			</Stack>

			<div style="height:{isMobile ? 50 : 100}px"></div>

			{#if showSave}
				<div class="save-bar" class:mobile={isMobile}>
					<MuiButton variant="contained" sx={{ height: 48, borderRadius: 0, width: '100%' }} disabled={!changed} onclick={save}>저장</MuiButton>
				</div>
			{/if}
		</Stack>
	{/if}
</Drawer>

<style>
	.thumb {
		width: 200px;
		height: 200px;
		object-fit: cover;
		border-radius: 8px;
		cursor: pointer;
	}
	.thumb.empty {
		display: flex;
		align-items: center;
		justify-content: center;
		background: #e5e7eb;
	}
	.thumb.float {
		position: absolute;
		top: -80px;
		right: 0;
	}
	.slide {
		width: 70px;
		height: 70px;
		object-fit: cover;
		border-radius: 8px;
		border: 1px solid #e5e7eb;
	}
	.slide-x {
		position: absolute;
		top: -6px;
		right: -6px;
		display: flex;
		padding: 2px;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.6);
		color: #fff;
		cursor: pointer;
	}
	.slide-x:hover {
		background: rgba(0, 0, 0, 0.8);
	}
	.star {
		cursor: pointer;
		color: #faaf00;
		display: inline-flex;
	}
	.star .material-icons {
		font-size: 24px;
		width: 24px;
		height: 24px;
	}
	.save-bar {
		position: fixed;
		z-index: 1;
		background: #fff;
		bottom: 0;
		left: 50%;
		width: 50%;
		animation: fadein 225ms ease;
	}
	.save-bar.mobile {
		left: 0;
		width: 100%;
	}
	@keyframes fadein {
		from {
			opacity: 0;
		}
	}
</style>
