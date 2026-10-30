/**
 * 브라우저 저장소(localStorage·sessionStorage) 키 — 전부 여기 한 곳에 둔다(2026-10-03 대표 결정).
 *
 *   규칙: `zbd:{영역}:{이름}` · 번호가 붙는 키는 `zbd:{영역}:{이름}:{번호}` (예 `zbd:freecell:history:617`)
 *   화면 코드에 문자열 키를 직접 쓰지 말고 SK.* 를 import 해 쓴다. 새 키도 여기에 더한다.
 *
 * 대상 아님: 앱 격리 실행(iframe) 안의 localStorage 폴리필(widget-sandbox.ts — 앱이 자기 키를 쓴다),
 *            운영자가 손으로 만든 백업 키(`__*_backup_*` 등).
 *
 * 옛 키(React 판 시절 이름)는 migrateStorage() 가 새 키로 옮겨 담는다 — 이 모듈을 import 하는 순간 한 번 돈다
 * (토큰을 모듈 로드 때 읽는 코드보다 먼저 돌게). 루트 +layout.svelte 도 부르지만 두 번째부터는 아무 일도 안 한다.
 */

export const SK = {
	/** 로그인 — 토큰·저장한 아이디·돌아갈 주소 */
	auth: {
		admin: 'zbd:auth:admin',
		user: 'zbd:auth:user',
		adminLoginPath: 'zbd:auth:adminLoginPath',
		adminLogin: 'zbd:auth:adminLogin',
		userLogin: 'zbd:auth:userLogin',
		adminGroupId: 'zbd:auth:adminGroupId'
	},
	/** 관리자 화면 */
	admin: {
		/** 로그인한 호스팅 — 공통 관리자는 "" (빈 값도 값이다) */
		hostingId: 'zbd:admin:hostingId',
		meta: 'zbd:admin:meta',
		themeOpts: 'zbd:admin:themeOpts',
		darkMode: 'zbd:admin:darkMode',
		settingGroupFolded: 'zbd:admin:settingGroupFolded',
		sideNavMode: 'zbd:admin:sideNavMode',
		pageSize: 'zbd:admin:pageSize',
		pageBorder: 'zbd:admin:pageBorder',
		/** 로컬 개발자 모드 — 콘솔에서 손으로 '1' 을 넣는다 */
		isDeveloper: 'zbd:admin:isDeveloper',
		monitorConfig: 'zbd:admin:monitorConfig',
		userManagePostPageSize: 'zbd:admin:userManagePostPageSize',
		userManageCommentPageSize: 'zbd:admin:userManageCommentPageSize',
		/** sessionStorage */
		userMemberManageState: 'zbd:admin:userMemberManageState',
		/** sessionStorage */
		userPostSearch: 'zbd:admin:userPostSearch',
		/** sessionStorage */
		userCommentSearch: 'zbd:admin:userCommentSearch',
		/** sessionStorage */
		reportSearch: 'zbd:admin:reportSearch',
		readySocketAutoReload: 'zbd:admin:readySocketAutoReload',
		readySocketMinimize: 'zbd:admin:readySocketMinimize',
		readySocketShowJson: 'zbd:admin:readySocketShowJson',
		videoPageSize: 'zbd:admin:videoPageSize',
		boardPageSize: 'zbd:admin:boardPageSize',
		youtubeVideoSearch: 'zbd:admin:youtubeVideoSearch',
		youtubeCommentSearch: 'zbd:admin:youtubeCommentSearch',
		hostingTab: 'zbd:admin:hostingTab',
		apiReferenceTab: 'zbd:admin:apiReferenceTab',
		jsonLintDemo: 'zbd:admin:jsonLintDemo',
		esEditorOpts: 'zbd:admin:esEditorOpts',
		esEditorJsOpts: 'zbd:admin:esEditorJsOpts',
		showMinimizedDock: 'zbd:admin:showMinimizedDock',
		templateWidgetSearch: 'zbd:admin:templateWidgetSearch',
		templatePageSearch: 'zbd:admin:templatePageSearch',
		widgetTemplateHomepage: 'zbd:admin:widgetTemplateHomepage',
		/** 목록 칸 고르기 — optionKey 마다 */
		tableColumn: (optionKey: string) => `zbd:admin:tableColumn:${optionKey}`,
		/** 떠 있는 창 자리·크기 — 창 id 마다 */
		window: (id: string) => `zbd:admin:window:${id}`,
		/** 글 편집 서랍의 접힌 칸 — 게시판마다 */
		postCollapse: (groupKey: string) => `zbd:admin:postCollapse:${groupKey}`
	},
	/** 디자인 관리 */
	design: {
		lastDesignId: 'zbd:design:lastDesignId',
		recentPages: 'zbd:design:recentPages',
		lastGnbDesignId: 'zbd:design:lastGnbDesignId',
		gnbManageTab: 'zbd:design:gnbManageTab',
		gnbBannerOpen: 'zbd:design:gnbBannerOpen',
		gnbWidgetFold: 'zbd:design:gnbWidgetFold',
		widgetFold: (foldKey: string | number | null) => `zbd:design:widgetFold:${foldKey}`,
		layoutFold: (imageId: string | number, rowId: string | number, itemId: string | number) => `zbd:design:layoutFold:${imageId}_${rowId}_${itemId}`,
		aiMessages: (id: string | number) => `zbd:design:aiMessages:${id}`
	},
	/** 앱 만들기 편집기(관리자·사용자 공용) */
	builder: {
		editorTab: 'zbd:builder:editorTab',
		editorHeightRatios: 'zbd:builder:editorHeightRatios',
		workbenchLeftWidth: 'zbd:builder:workbenchLeftWidth',
		editorTabView: 'zbd:builder:editorTabView',
		preview: 'zbd:builder:preview',
		chat: (postId: string) => `zbd:builder:chat:${postId}`
	},
	/** 이미지 편집 도구(그림판 — 업무관리 카드 그림에 쓴다) */
	canvas: {
		currentColor: 'zbd:canvas:currentColor',
		strokeWidth: 'zbd:canvas:strokeWidth'
	},
	/** 업무관리(work) */
	work: {
		search: (groupKey: string) => `zbd:work:search:${groupKey}`,
		collapsed: (groupKey: string) => `zbd:work:collapsed:${groupKey}`,
		shapeOpen: (groupKey: string, workId: string | number, id: string) => `zbd:work:shapeOpen:${groupKey}_${workId}_${id}`
	},
	/** 웹하드 */
	explorer: {
		mdViewMode: 'zbd:explorer:mdViewMode',
		lastPath: (rootFolder: string) => `zbd:explorer:lastPath:${rootFolder}`,
		mdSplitRatio: (resizeKey: string, dir: string) => `zbd:explorer:mdSplitRatio:${resizeKey}_${dir}`
	},
	/** 웹오피스 공유 링크 손님(/o/{키}) — 웹오피스 플러그인 */
	office: {
		/** sessionStorage — 비밀번호를 넣어 받은 입장권(새로고침해도 다시 묻지 않게) */
		sharePass: (key: string) => `zbd:office:sharePass:${key}`,
		/** 함께 편집할 때 남에게 보일 내 이름 */
		guestName: 'zbd:office:guestName'
	},
	claude: {
		fontSize: 'zbd:claude:fontSize',
		/** sessionStorage — 지금 보는 대화 */
		current: (owner: string | number, win = false) => `zbd:claude:current:${owner}${win ? '-win' : ''}`,
		/** 대화마다 마지막으로 쓴 중계 번호 {sessionId: owner} — 같은 폴더 중계끼리 대화 목록을 나눠 써서 목록만으로는 어느 중계인지 모른다 */
		sessionOwner: 'zbd:claude:sessionOwner'
	},
	messenger: {
		claudeSeen: 'zbd:messenger:claudeSeen',
		/** 반응 고르기 창 「최근 사용」 이모지(최대 24) */
		recentEmoji: 'zbd:messenger:recentEmoji'
	},
	talk: {
		/** sessionStorage */
		handoffAt: 'zbd:talk:handoffAt'
	},
	/** 웹 푸시 — 관리자·사용자 공용 */
	push: {
		hostingMap: 'zbd:push:hostingMap',
		test: 'zbd:push:test'
	},
	/** 사용자 사이트 공통 */
	site: {
		bannerClicks: 'zbd:site:bannerClicks',
		agencyAdminId: 'zbd:site:agencyAdminId'
	},
	nav: {
		/** sessionStorage — 뒤로 가기 스크롤 자리 */
		scroll: 'zbd:nav:scroll'
	},
	freecell: {
		gameState: 'zbd:freecell:gameState',
		replaySpeed: 'zbd:freecell:replaySpeed',
		range: 'zbd:freecell:range',
		challenge: 'zbd:freecell:challenge',
		history: (gameNumber: string | number) => `zbd:freecell:history:${gameNumber}`,
		currentReplay: (gameNumber: string | number) => `zbd:freecell:currentReplay:${gameNumber}`
	},
	life: {
		lastTab: 'zbd:life:lastTab',
		voiceShowHint: 'zbd:life:voiceShowHint',
		voiceAutoDetect: 'zbd:life:voiceAutoDetect',
		writeGuide: 'zbd:life:writeGuide',
		appShortcuts: 'zbd:life:appShortcuts',
		weatherArea: 'zbd:life:weatherArea',
		financeShowIncome: 'zbd:life:financeShowIncome',
		todoHideDone: 'zbd:life:todoHideDone',
		labelWidgetTab: 'zbd:life:labelWidgetTab',
		desktopWindows: 'zbd:life:desktopWindows',
		/** sessionStorage — 휴대폰에서 보던 앱 */
		desktopMobileOpen: 'zbd:life:desktopMobileOpen'
	},
	mini: {
		editFold: 'zbd:mini:editFold',
		editTab: 'zbd:mini:editTab'
	}
} as const;

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기 (몇 주 뒤 이 아래 표·migrateStorage 와 함께 지운다)
/** React 판 창 사본(`__windows` = {창 id: {...}}) — 미리보기 창이 자리를 한 번 읽어 온다(읽기만) */
export const LEGACY_WINDOWS_KEY = '__windows';

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
/** 로그인 토큰 — BE 소셜 로그인 콜백(kakao·naver·google)·관리자 사이트 로그인 화면이 아직 옛 키에 쓴다 → 옛 키에 값이 있으면 늘 새 키를 덮는다 */
const TOKEN_RENAMES: [string, string][] = [
	['__admin_token', SK.auth.admin],
	['__user_token', SK.auth.user]
];

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
/** localStorage 이름 바꾸기(옛 → 새). 새 키가 이미 있으면 옛 값은 버린다 */
const LOCAL_RENAMES: [string, string][] = [
	['__admin_login_path', SK.auth.adminLoginPath],
	['__admin_login', SK.auth.adminLogin],
	['__user_login', SK.auth.userLogin],
	['companyId', SK.admin.hostingId], // 값 "" (공통 관리자) 도 그대로 옮긴다
	['zbd:admin:companyId', SK.admin.hostingId], // 옛 이름 companyId — 2026-10 옮기는 기간
	['__admin_groupId', SK.auth.adminGroupId],
	['__admin_metal_global', SK.admin.meta],
	['__theme_opts', SK.admin.themeOpts],
	['__dark_mode_preference', SK.admin.darkMode],
	['SIDE_NAV_MODE', SK.admin.sideNavMode],
	['__admin_pageSize', SK.admin.pageSize],
	['__admin_is_border', SK.admin.pageBorder],
	['isDeveloper', SK.admin.isDeveloper],
	['__montor_config', SK.admin.monitorConfig],
	['userManagePostPageSize', SK.admin.userManagePostPageSize],
	['userManageCommentPageSize', SK.admin.userManageCommentPageSize],
	['admin_ready_socket_auto_reload', SK.admin.readySocketAutoReload],
	['admin_ready_socket_minimize', SK.admin.readySocketMinimize],
	['ready_socket_show_json', SK.admin.readySocketShowJson],
	['__admin_video_page_size', SK.admin.videoPageSize],
	['boardPageSize', SK.admin.boardPageSize],
	['youtube-video-search', SK.admin.youtubeVideoSearch],
	['youtube-comment-search', SK.admin.youtubeCommentSearch],
	['__admin_company_tab', SK.admin.hostingTab],
	['zbd:admin:companyTab', SK.admin.hostingTab], // 옛 이름 companyTab — 2026-10 옮기는 기간
	['__api_tab_key', SK.admin.apiReferenceTab],
	['__jsonlint_demo_json_str', SK.admin.jsonLintDemo],
	['__es_editor_opts', SK.admin.esEditorOpts],
	['__es_editor_js_opts', SK.admin.esEditorJsOpts],
	['showMinimizedDock', SK.admin.showMinimizedDock],
	['__last_designId', SK.design.lastDesignId],
	['__last_dnb_designId', SK.design.lastGnbDesignId],
	['__gnb_manage_tab', SK.design.gnbManageTab],
	['__gnb_banner_open', SK.design.gnbBannerOpen],
	['__gnb_widget_fold_map', SK.design.gnbWidgetFold],
	['widgetMakerSelectedTab', SK.builder.editorTab],
	['WIDGET_MAKER_EDITOR_HEIGHT_RATIOS', SK.builder.editorHeightRatios],
	['__widget_workbench_left_width', SK.builder.workbenchLeftWidth],
	['WIDGET_EDITOR_TAB_VIEW', SK.builder.editorTabView],
	['canvas_current_color', SK.canvas.currentColor],
	['canvas_stroke_width', SK.canvas.strokeWidth],
	['__explorer_md_view_mode', SK.explorer.mdViewMode],
	['__cc_chat_font_size', SK.claude.fontSize],
	['__messenger_claude_seen', SK.messenger.claudeSeen],
	['__pwa_push_companyId_map', SK.push.hostingMap],
	['zbd:push:companyMap', SK.push.hostingMap], // 옛 이름 companyMap — 2026-10 옮기는 기간
	['__pwa_test', SK.push.test],
	['__conditional_banner_clicks', SK.site.bannerClicks],
	['agencyAdminId', SK.site.agencyAdminId],
	['freecell-game-state', SK.freecell.gameState],
	['zbd.freecell.replaySpeed', SK.freecell.replaySpeed],
	['zbd.freecell.range', SK.freecell.range],
	['zbd.freecell.challenge', SK.freecell.challenge],
	// 라이프 — React 판 이름(예전엔 lsGet 이 그때그때 옮겼다)
	['__life_last_tab', SK.life.lastTab],
	['__auto_voice_detect_hint', SK.life.voiceShowHint],
	['__auto_voice_detect', SK.life.voiceAutoDetect],
	['__lifeWriteGuide', SK.life.writeGuide],
	['__life_app_shortcuts', SK.life.appShortcuts],
	['__widget_label_tab', SK.life.labelWidgetTab],
	['weather_area', SK.life.weatherArea],
	['__finance_show_income', SK.life.financeShowIncome],
	['lifeTodoHideDone', SK.life.todoHideDone],
	// 이 앱 첫 판의 이름(규칙과 글자 모양이 달랐다)
	['zbd:life:label-widget-tab', SK.life.labelWidgetTab],
	['zbd:life:desktop-windows', SK.life.desktopWindows],
	['zbd:mini:edit-fold', SK.mini.editFold],
	['zbd:mini:edit-tab', SK.mini.editTab]
];

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
/** 읽는 곳이 없는 옛 키 — 지우기만 (serviceType = 옛 BE SSR 이 쓰던 것) */
const LOCAL_DROPS = ['serviceType'];

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
/** localStorage 번호 붙은 키(옛 앞부분 → 새 앞부분). 위 이름 바꾸기 뒤에 돈다 */
const LOCAL_PREFIXES: [string, string][] = [
	['table_column_', 'zbd:admin:tableColumn:'],
	['win_', 'zbd:admin:window:'],
	['__bd_collapse_', 'zbd:admin:postCollapse:'],
	['__design_widget_fold_', 'zbd:design:widgetFold:'],
	['__design_widget_layout_fold_map_', 'zbd:design:layoutFold:'],
	['ai_design_messages_', 'zbd:design:aiMessages:'],
	['canvas_search_', 'zbd:work:search:'],
	['__canvas_collapsed_', 'zbd:work:collapsed:'],
	['shape_', 'zbd:work:shapeOpen:'],
	['explorer_last_path_', 'zbd:explorer:lastPath:'],
	['__md_split_ratio_', 'zbd:explorer:mdSplitRatio:'],
	['freecell-history-', 'zbd:freecell:history:'],
	['freecell-current-replay-', 'zbd:freecell:currentReplay:']
];

// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
/** sessionStorage 이름 바꾸기 */
const SESSION_RENAMES: [string, string][] = [
	['admin_user_member_manage_state', SK.admin.userMemberManageState],
	['admin_user_post_search', SK.admin.userPostSearch],
	['admin_user_comment_search', SK.admin.userCommentSearch],
	['admin_report_search', SK.admin.reportSearch],
	['__talk_handoff_at', SK.talk.handoffAt],
	['zbd:life:desktop-mobile-open', SK.life.desktopMobileOpen],
	['__life_desktop_mobile_open', SK.life.desktopMobileOpen]
];
/** sessionStorage 번호 붙은 키 */
const SESSION_PREFIXES: [string, string][] = [['claude-chat-current-', 'zbd:claude:current:']];

function renameAll(s: Storage, renames: [string, string][], prefixes: [string, string][], overwrite = false) {
	for (const [from, to] of renames) {
		const v = s.getItem(from);
		if (v === null) continue; // 빈 문자열 "" 은 값이다 — 그대로 옮긴다
		if (overwrite || s.getItem(to) === null) s.setItem(to, v);
		s.removeItem(from);
	}
	if (!prefixes.length) return;
	const keys: string[] = [];
	for (let i = 0; i < s.length; i++) {
		const k = s.key(i);
		if (k) keys.push(k);
	}
	for (const k of keys) {
		const p = prefixes.find(([from]) => k.startsWith(from));
		if (!p) continue;
		const to = p[1] + k.slice(p[0].length);
		const v = s.getItem(k);
		if (v !== null && s.getItem(to) === null) s.setItem(to, v);
		s.removeItem(k);
	}
}

let migrated = false;
/**
 * 옛 키 → 새 키 옮겨 담기(새 키가 없을 때만 복사, 옛 키는 지운다). 로그인 토큰만 옛 키 값이 이긴다(위 TOKEN_RENAMES).
 * 페이지를 열 때마다 한 번 — 소셜 로그인 콜백이 옛 키로 쓴 토큰도 여기서 넘어온다.
 */
export function migrateStorage() {
	if (migrated || typeof window == 'undefined') return;
	migrated = true;
	// 옛 localStorage 키 — 2026-10 이전 저장값 옮겨 담기
	try {
		const ls = window.localStorage;
		renameAll(ls, TOKEN_RENAMES, [], true);
		// 왼쪽 메뉴: 아주 옛 SIDE_NAV_OPEN('0' = 숨김) 은 모드로 바꿔 담는다
		const oldOpen = ls.getItem('SIDE_NAV_OPEN');
		if (oldOpen !== null) {
			if (oldOpen === '0' && ls.getItem('SIDE_NAV_MODE') === null && ls.getItem(SK.admin.sideNavMode) === null) ls.setItem(SK.admin.sideNavMode, 'hidden');
			ls.removeItem('SIDE_NAV_OPEN');
		}
		renameAll(ls, LOCAL_RENAMES, LOCAL_PREFIXES);
		for (const k of LOCAL_DROPS) ls.removeItem(k);
	} catch {
		// 사생활 보호 창 등 — 옮기지 못해도 화면은 돈다
	}
	try {
		renameAll(window.sessionStorage, SESSION_RENAMES, SESSION_PREFIXES);
	} catch {}
}

// import 하는 순간 한 번 — 토큰·설정을 모듈 로드 때 읽는 코드(api.ts·esOpts·nav 등)보다 먼저
migrateStorage();
