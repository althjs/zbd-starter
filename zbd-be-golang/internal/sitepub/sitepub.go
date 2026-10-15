// Package sitepub — 사용자 화면에 내보내는 사이트 공통 데이터(base)의 허용 목록.
//
// base 는 누구나 받는 공개 주소(/service/design/{회사}/base)로 나간다. 예전엔 관리용 설정(관리자 공통코드·회사 내부 값 …)
// 까지 통째로, 그것도 모든 HTML 소스에 실렸다(2026-10-03 대표 — 「소스보기에 안 나오게」). 여기서 사용자 화면이 쓰는 칸만 고른다.
// 화면이 새 칸을 쓰게 되면 여기에 더한다 — 빠지면 그 값이 조용히 비어 보인다.
package sitepub

import (
	"crypto/sha1"
	"encoding/hex"

	"zbd.kr/be/internal/jsx"
)

// 사이트 메타 중 화면이 쓰는 칸 — 회사 정보(푸터 HTML 의 {키} 치환)·디자인·로그인 공개 키·기능 스위치
var siteMetaKeys = []string{
	"title", "logo", "favicon", "ceoName", "companyName", "companyTel", "companyFax", "companyEmail", "companyNo",
	"companyType", "companyAddress", "companyAddressOldType", "hostingId", "theme", "disablePageBorder", "headerHeight",
	"homepage", "homepagePath", "cookieHost", "serviceType", "design", "design_outbound", "feature", "auth", "apiKeys",
}

// 사이트 디자인 중 내보내지 않는 칸 — 서버가 HTML 에 따로 넣거나(globalHTML·robots) 관리용(cssMemo)
var designDrop = []string{"globalHTML", "globalHTMLBodyTop", "cssMemo", "robots"}

// 공통코드 중 화면이 쓰는 묶음(group 또는 component)
var commonCodeGroups = map[string]bool{"POST_ISSUE_REPORT": true, "SEO": true, "AUTH_GROUP": true}

// PublicBase — base 에서 사용자 화면이 쓰는 칸만 새로 담는다(원본은 건드리지 않는다)
func PublicBase(base *jsx.OrderedMap) *jsx.OrderedMap {
	out := jsx.NewOrderedMap()
	if base == nil {
		return out
	}
	// GNB 는 한 벌만 — 화면은 gnb 를 gnbMap[designId] 에도 넣는다(예전엔 gnb·gnbMap·activeGNB 로 세 번 실렸다)
	if v := base.V("gnb"); v != nil {
		out.Set("gnb", jsx.Clone(v))
	}
	if sm := base.Map("siteMeta"); sm != nil {
		m := jsx.NewOrderedMap()
		for _, k := range siteMetaKeys {
			if sm.Has(k) {
				m.Set(k, jsx.Clone(sm.V(k)))
			}
		}
		for _, dk := range []string{"design", "design_outbound"} {
			if d := m.Map(dk); d != nil {
				for _, k := range designDrop {
					d.Delete(k)
				}
			}
		}
		out.Set("siteMeta", m)
	}
	if v := base.V("postGroupsMeta"); v != nil {
		out.Set("postGroupsMeta", jsx.Clone(v))
	}
	codes := []any{}
	for _, x := range base.Arr("commonCodes") {
		if c, ok := x.(*jsx.OrderedMap); ok && (commonCodeGroups[jsx.Str(c.V("group"))] || commonCodeGroups[jsx.Str(c.V("component"))]) {
			codes = append(codes, jsx.Clone(c))
		}
	}
	out.Set("commonCodes", codes)
	for _, k := range []string{"holydays", "label", "CSS"} {
		if base.Has(k) {
			out.Set(k, jsx.Clone(base.V(k)))
		}
	}
	return out
}

// Version — 공개 base 의 지문(캐시 주소 ?v= 에 쓴다 — 내용이 바뀌면 주소가 바뀐다)
func Version(pub *jsx.OrderedMap) string {
	h := sha1.Sum([]byte(jsx.Stringify(pub)))
	return hex.EncodeToString(h[:])[:12]
}
