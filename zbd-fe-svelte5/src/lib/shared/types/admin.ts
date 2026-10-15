/** 관리자 — 로그인 프로필·권한 메뉴·관리자 명단·권한 그룹·알림·대시보드·푸시 기기(서버 응답 모양) */

/** 로그인한 관리자(메뉴 API `admin/common/menus` 의 profile) */
export type AdminProfile = {
	adminId: number;
	/** 권한 그룹(admin_roles) — 1 은 최고 그룹 */
	groupId: number;
	name: string;
	nickname?: string;
	mobile?: string;
	email?: string;
	/** 프로필 사진(폴더 admin_profile_{adminId}) */
	image?: { imageId?: string; name?: string };
	/** 빠른 메뉴(즐겨찾기) 메뉴 번호들 */
	/** 호스팅(ZBD) — 공통 "" 이면 본사 */
	hostingId: string;
	/** 위 호스팅(하위 호스팅일 때) */
	parentHostingId?: string;
	/** 마스터가 다른 호스팅을 볼 때 원래 호스팅 */
	_hostingId?: string;
	/** 개발자 — 개발 도구·JSON 편집·로그인(DEV) 등이 열린다 */
	isDeveloper?: boolean;
	/** 최고 관리자(옛 isDxAdmin) — 관리자 설정 칸들이 보인다 */
	isSuperAdmin?: boolean;
	/** 사무실 IP 에서 접속 중 */
	isOffice?: boolean;
	/** 관리자 개인 설정(json.settings) */
	settings?: Record<string, any>;
	/** 토큰이 묶인 브라우저 id */
	browserUid?: string;
	[k: string]: any;
};

/** 권한 메뉴(API 원본) — 왼쪽 메뉴와 권한 검사(주소가 메뉴 안이어야 열림)의 바탕 */
export type AdminMenu = {
	menuId: number;
	name: string;
	/** 메뉴 설명 — 화면 제목 아래 부제로 쓴다 */
	description?: string;
	/** 머티리얼 아이콘 이름 또는 200x200_… 그림 */
	icon?: string;
	/** 주소(/admin/…) — 하위 메뉴가 있는 대메뉴는 비어 있다 */
	path?: string;
	/**
	 * 특수 메뉴 — 하위 메뉴를 화면이 채운다.
	 * mail 웹메일 폴더 · adminPosts 관리자 게시판 목록 · userPosts 사용자 게시판 목록(POST_GROUPS, hideInMenu 는 뺀다)
	 */
	special?: string;
	/** 이 메뉴 권한 — R 읽기 · RW 쓰기(대메뉴는 보통 R, 특수 메뉴는 실제 권한) */
	auth?: '' | 'R' | 'RW';
	hostingId?: string;
	/** 하위 메뉴 */
	items?: AdminMenu[];
};

/** 관리자 알림(오른쪽 위 종) — items 는 알림 글, checkedNotiIds 는 읽은 것 */
export type AdminNotification = { items: any[]; checkedNotiIds: (string | number)[] };

/** 대시보드 */
export type DashboardData = {
	/** 오늘·전체 가입, 탈퇴·탈퇴 신청 */
	memberSummary: { todayCount: number; leaveCount: number; leaveRequestCount: number; totalCount: number };
	recentPosts: { postId: string; groupKey: string; title: string; createdAt: string; jsonUser?: { name?: string } }[];
	recentMembers: { userId: number; nickname: string; status: number; createdAt: number }[];
	/** 게시판별 공개 글 수(전체·최근 7일) */
	boards: { groupKey: string; name: string; kind: 'admin' | 'user'; count: number; week: number }[];
	work: { byStatus: { status: number; count: number }[]; recent: { workId: number; title: string; status: number; updatedAt: number }[] };
};

/** 관리자 명단 한 줄(관리자 회원 관리) */
export type AdminMember = {
	adminId: number;
	name: string;
	groupId?: number;
	groupName?: string;
	hostingId?: string;
	email?: string;
	/** 부가 정보(DB 문자열) — 조직(org)·AI 계정(isAI·aiCharacter)·개발자 등 */
	json?: any;
	/** json 을 풀어 둔 객체(화면에 따라 둘 중 하나) */
	Json?: any;
	[k: string]: any;
};

/** 권한 그룹(admin_roles) */
export type AdminGroup = { groupId: number; name: string; description?: string; hostingId: string; [k: string]: any };

/** 내 웹 푸시 구독(기기·앱) — `GET /service/pwa/push?action=mine` */
export type PushDevice = {
	notiUserId: number;
	/** 0 받음 · 99 이 기기 알림 끔 */
	status: number;
	/** 등록한 앱 주소(backyard·chat·mail …) — 요청의 Origin */
	origin: string;
	/** 기기(User-Agent 요약) */
	ua: string;
	/** 종류별 끄기 — messenger:false 면 메신저 알림을 이 기기로 안 보낸다 */
	notiMap: any;
	/** 푸시 주소(브라우저가 준 것) — 이 기기인지 맞춰 볼 때 쓴다 */
	endpoint: string;
	/** 등록 시각(ms) */
	createdAt: number;
	updatedAt?: number;
};
