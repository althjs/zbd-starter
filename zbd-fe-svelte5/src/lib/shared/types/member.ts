/** 회원 — 로그인 프로필·MCP 토큰(서버 응답 모양). 회원 이름·이메일·전화는 DB 에 암호문으로 있고 서버가 풀어 준다 */

/** 로그인한 회원(사용자 화면 session.profile) */
export type UserProfile = {
	userId: number;
	/** 닉네임 — 공개 화면에는 이것만 쓴다(본명·이메일은 내보내지 않는다) */
	nickname?: string;
	/** 본명(본인에게만) */
	name?: string;
	/** 계정 프로필 사진 — 파일 이름(문자열, 폴더 u_{userId}_my). 예전 화면은 { name } 모양이었다 */
	image?: any;
	/** 소셜 가입 뒤 닉네임 등 추가 정보를 아직 안 넣었으면 true(추가정보 화면으로 보낸다) */
	isAdditionalInfoRequired?: boolean;
	[k: string]: any;
};

/** 내 MCP 토큰 — 앱 만들기(`/service/mcp`)·미니홈피 「AI 연동」(`/service/mcp/mini`) 이 같은 목록을 쓴다 */
export type McpToken = {
	id: string;
	/** 사람이 붙인 이름(예 「Claude Desktop」) */
	name: string;
	/** 권한 — widget:read·widget:write·mini:read·mini:write */
	scopes: string[];
	/** 만든 시각(ISO) */
	createdAt: string;
	/** 마지막 사용 시각(ISO) */
	lastUsedAt?: string;
	/** 토큰 원문 — 만든 직후 응답에만 한 번 온다(다시 볼 수 없다) */
	token?: string;
};

/** 주소(카카오 우편번호로 찾은 값) — 회원 json.address · 사이트 회사 주소 SITE.companyAddressInfo 가 같은 모양 */
export interface Address {
	zipcode?: string;
	/** 도로명 주소(검색으로만 채움) */
	address1?: string;
	/** 상세주소(직접 입력) */
	address2?: string;
}
