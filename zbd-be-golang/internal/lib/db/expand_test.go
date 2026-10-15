package db

import (
	"testing"

	"zbd.kr/be/internal/jsx"
)

func TestExpand(t *testing.T) {
	q, a := expand("SELECT * FROM t WHERE id IN (?) AND c = ? AND ?? = 1", []any{[]any{float64(1), float64(2)}, "x", "col"})
	if q != "SELECT * FROM t WHERE id IN (?, ?) AND c = ? AND `col` = 1" || len(a) != 3 || a[0] != int64(1) {
		t.Fatal(q, a)
	}
	q, a = expand("UPDATE t SET ? WHERE id = ?", []any{jsx.Obj("a", "x", "b", float64(2)), float64(5)})
	if q != "UPDATE t SET `a` = ?, `b` = ? WHERE id = ?" || len(a) != 3 {
		t.Fatal(q, a)
	}
	q, _ = expand("INSERT INTO t (a, b) VALUES ?", []any{[]any{[]any{float64(1), "x"}, []any{float64(2), "y"}}})
	if q != "INSERT INTO t (a, b) VALUES (?, ?), (?, ?)" {
		t.Fatal(q)
	}
}
