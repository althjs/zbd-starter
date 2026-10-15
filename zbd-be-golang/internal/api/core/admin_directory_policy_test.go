package core

import (
	"strings"
	"testing"

	"zbd.kr/be/internal/jsx"
)

// 관리자 회원 권한이 없는 관리자에게 주는 관리자 목록 — 이름·사진·조직만(E2E K5 뒤)
func TestAdminDirectoryRowPolicy(t *testing.T) {
	m := jsx.ParseMap(`{"adminId":7,"groupId":1,"name":"홍길동","email":"a@b.c","mobile":"010","status":0,"hostingId":"ZBD","groupName":"최고 관리자",
		"json":"{\"image\":{\"name\":\"p.webp\"},\"nickname\":\"길동\",\"org\":{\"unitId\":3},\"isDeveloper\":true,\"isSuperAdmin\":true,\"authMenus\":[1]}"}`)
	o := adminDirectoryRow(m)
	if o.V("email") != "" || o.V("mobile") != "" {
		t.Errorf("이메일·전화가 남음: %v %v", o.V("email"), o.V("mobile"))
	}
	if jsx.Str(o.V("name")) != "홍길동" || jsx.Number(o.V("adminId")) != 7 {
		t.Errorf("이름·번호가 빠짐: %s", jsx.Stringify(o))
	}
	j := jsx.Str(o.V("json"))
	for _, bad := range []string{"isDeveloper", "isSuperAdmin", "authMenus"} {
		if strings.Contains(j, bad) {
			t.Errorf("json 에 %s 가 남음: %s", bad, j)
		}
	}
	for _, good := range []string{"p.webp", "길동", "unitId"} {
		if !strings.Contains(j, good) {
			t.Errorf("json 에 %s 가 빠짐: %s", good, j)
		}
	}
}
