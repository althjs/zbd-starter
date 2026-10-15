package v1x

import (
	"encoding/base64"
	"strings"
	"testing"

	"zbd.kr/be/internal/lib/pii"
)

func keys(t *testing.T) *pii.Keys {
	k32 := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	i32 := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("i", 32)))
	k, err := pii.New(k32, i32)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestMemberSearch(t *testing.T) {
	k := keys(t)
	cases := []struct {
		q, sql string
		args   int
	}{
		{"a@b.com", "m.email_hash = ?", 1},
		{"010-1234-5678", "(m.mobile_hash = ? OR m.userId = ?)", 2},
		{"존슨", "(m.nickname LIKE ? OR m.name_hash = ?)", 2},
		{"", "", 0},
	}
	for _, c := range cases {
		w := &Where{}
		MemberSearch(w, k, "m.", c.q)
		if got := strings.TrimPrefix(w.SQL(), " WHERE "); got != c.sql || len(w.Args) != c.args {
			t.Errorf("%q → %q %v", c.q, got, w.Args)
		}
	}
	w := &Where{}
	MemberSearch(w, k, "", "50%_")
	if w.Args[0] != `50\%\_%` {
		t.Errorf("LIKE 이스케이프: %v", w.Args[0])
	}
	w = &Where{}
	MemberSearch(w, k, "", "010-1234")
	if w.Args[1] != "0101234" {
		t.Errorf("숫자만: %v", w.Args[1])
	}
}

func TestInts(t *testing.T) {
	if got := Ints("1, 2,x,-3"); len(got) != 3 || got[2] != int64(-3) {
		t.Errorf("%v", got)
	}
	if got := Strs(" a,,b "); len(got) != 2 {
		t.Errorf("%v", got)
	}
}
