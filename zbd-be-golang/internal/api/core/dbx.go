package core

// 사본: internal/lib/db 의 Query·expand — 전환 뒤 합칠 것.
//   다른 점: 못 읽는 날짜(0000-00-00 00:00:00)를 null 이 아니라 원래 글자로 — 지금 서비스 드라이버와 같다.

import (
	"context"
	"database/sql"
	"math"
	"strconv"
	"strings"
	"time"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
)

// query — 결과를 순서 보존 객체 배열로(파일 머리 주석의 값 모양).
func query(ctx context.Context, q db.Querier, query string, args ...any) ([]*jsx.OrderedMap, error) {
	query, args = dbExpand(query, args)
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(cols))
	kinds := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name()
		kinds[i] = strings.ToUpper(c.DatabaseTypeName())
	}
	var out []*jsx.OrderedMap
	for rows.Next() {
		raw := make([]sql.RawBytes, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := jsx.NewOrderedMap()
		for i, b := range raw {
			m.Set(names[i], dbConvert(kinds[i], b))
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// dbConvert — 컬럼 타입별 값 변환(파일 머리 주석의 규칙).
func dbConvert(kind string, b sql.RawBytes) any {
	if b == nil {
		return nil
	}
	s := string(b)
	k := strings.TrimPrefix(kind, "UNSIGNED ")
	switch k {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT", "YEAR", "FLOAT", "DOUBLE", "REAL":
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return s
		}
		return f
	case "DECIMAL", "NEWDECIMAL":
		return s
	case "DATE", "DATETIME", "TIMESTAMP":
		return dbJSDate(k, s)
	default:
		return s
	}
}

// dbJSDate — DATE/DATETIME → UTC ISO(밀리초까지). 잘못된 날짜면 null.
func dbJSDate(kind, s string) any {
	layout := "2006-01-02 15:04:05"
	v := s
	if kind == "DATE" {
		layout = "2006-01-02"
	} else if i := strings.IndexByte(v, '.'); i > 0 {
		layout = "2006-01-02 15:04:05." + strings.Repeat("0", len(v)-i-1)
	}
	t, err := time.ParseInLocation(layout, v, db.KST)
	if err != nil {
		return s // 드라이버처럼 0000-00-00 등 못 읽는 날짜는 원래 글자 그대로
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// queryOne — 첫 행(없으면 nil)
func queryOne(ctx context.Context, q db.Querier, sqlText string, args ...any) (*jsx.OrderedMap, error) {
	rows, err := query(ctx, q, sqlText, args...)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// dbExpand — 사본(internal/lib/db expand)
func dbExpand(query string, args []any) (string, []any) {
	if len(args) == 0 {
		return query, args
	}
	var b strings.Builder
	out := make([]any, 0, len(args))
	ai := 0
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch != '?' {
			b.WriteByte(ch)
			continue
		}
		if ai >= len(args) {
			b.WriteByte(ch)
			continue
		}
		if i+1 < len(query) && query[i+1] == '?' { // ??
			i++
			b.WriteString(dbIdent(jsx.Str(args[ai])))
			ai++
			continue
		}
		dbWriteValue(&b, &out, args[ai], true)
		ai++
	}
	return b.String(), out
}

func dbIdent(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = "`" + strings.ReplaceAll(p, "`", "``") + "`"
	}
	return strings.Join(parts, ".")
}

func dbWriteValue(b *strings.Builder, out *[]any, v any, top bool) {
	switch x := v.(type) {
	case []any:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			if inner, ok := e.([]any); ok {
				b.WriteByte('(')
				dbWriteValue(b, out, inner, false)
				b.WriteByte(')')
				continue
			}
			dbWriteValue(b, out, e, false)
		}
	case []int64:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
			*out = append(*out, e)
		}
	case []string:
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
			*out = append(*out, e)
		}
	case *jsx.OrderedMap:
		if x == nil {
			b.WriteString("NULL")
			return
		}
		for i, k := range x.Keys() {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(dbIdent(k))
			b.WriteString(" = ")
			dbWriteValue(b, out, x.V(k), false)
		}
	default:
		b.WriteByte('?')
		*out = append(*out, dbScalar(v))
	}
}

func dbScalar(v any) any {
	switch x := v.(type) {
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1<<53 {
			return int64(x)
		}
		return x
	case *jsx.OrderedMap, []any:
		return jsx.Stringify(x) // 드라이버에 객체를 넘기면 안 된다 — JSON 문자열로
	}
	return v
}
