/**
 * 관리자 실시간 — 스타터 기본판: **웹소켓 연결이 없다**(모든 화면은 새로고침해야 반영된다).
 * 웹소켓 연동 플러그인을 붙이면 이 파일이 실제로 붙는 판으로 바뀐다 — 쓰는 쪽 코드는 그대로.
 *
 *   startAdminSocket(isDeveloper)   레이아웃이 로그인 확인 뒤 한 번(여기선 아무 일 없음)
 *   adminSocket()                   소켓(여기선 늘 null) — `adminSocket()?.emit(…)` 처럼 쓴다
 *   onSocketDataUpdate(groupKey, fn, { self })   groupKey 신호 구독 → 정리 함수
 *   onAdminSocket(event, fn)        그 밖의 이벤트 구독 → 정리 함수
 *   socketState.connected           연결 여부($state, 여기선 늘 false)
 */
export const socketState = $state({ connected: false });

type DataFn = (data: { groupKey: string; key?: any; zbduid?: string; [k: string]: any }, isSelf: boolean) => void;

// 플러그인 판은 ZbdSocket 을 돌려준다 — 쓰는 쪽은 `?.` 로만 부른다
export function adminSocket(): any {
	return null;
}

export function startAdminSocket(_isDeveloper = false): any {
	return null;
}

export function stopAdminSocket() {}

export function onSocketDataUpdate(_groupKey: string, _fn: DataFn, _opts: { self?: boolean } = {}) {
	return () => {};
}

export function onAdminSocket(_event: string, _fn: (data: any) => void) {
	return () => {};
}
