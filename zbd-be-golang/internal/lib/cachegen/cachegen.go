// Package cachegen — 「설정·권한·디자인이 바뀌었다」 세대 번호.
//
// 시간으로 버티는 작은 캐시(SEO 메뉴·막은 미니홈피 주소·관리자 이름 …) 가 값을 담을 때 Gen() 을 같이 적어 두고,
// 꺼낼 때 지금 Gen() 과 다르면 버린다. 관리자 화면 저장·소켓 신호(onUpdateCodes) 때 app.Changed 가 Bump 한다.
// → 시간 제한은 「아무도 안 바꿨을 때 DB 를 덜 읽는」 용도로만 남고, 바꾸면 바로 반영된다(2026-10-04).
package cachegen

import "sync/atomic"

var gen atomic.Int64

// Gen — 지금 세대
func Gen() int64 { return gen.Load() }

// Bump — 세대를 올린다(모든 작은 캐시가 다음 조회 때 새로 읽는다)
func Bump() { gen.Add(1) }
