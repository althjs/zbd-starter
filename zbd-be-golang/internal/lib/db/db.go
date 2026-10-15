// Package db — MariaDB 연결과, 조회 결과를 **지금 서비스(Node BE)의 mysql 드라이버와 같은 값 모양**으로 돌려주기.
//
// 이관 중에는 화면이 응답을 그대로 쓰므로 값 모양이 같아야 한다:
//
//	· 정수·실수(TINYINT~BIGINT, FLOAT, DOUBLE) → number
//	· DECIMAL(합계·평균 등)                    → number (지금 서비스 드라이버가 supportBigNumbers 없이 쓴다 — 큰 수는 JS 처럼 정밀도를 잃는다)
//	· DATE / DATETIME / TIMESTAMP             → UTC ISO 문자열(서버 시간대 KST 로 읽어 바꾼다)
//	· 그 외(문자열·TEXT·JSON)                  → string
//	· NULL                                    → null
//
// ⚠ interpolateParams=true — 값을 클라이언트에서 끼워 넣어 **텍스트 프로토콜**로 보낸다
//
//	(서버 측 prepared statement 의 바이너리 프로토콜이면 값 타입이 달라진다).
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/config"
)

// KST — DB 의 DATE/DATETIME 은 한국 시각으로 저장돼 있다.
var KST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return time.FixedZone("KST", 9*3600)
	}
	return loc
}()

func Open(c config.DBConfig) (*sql.DB, error) {
	mc := mysql.NewConfig()
	mc.User = c.User
	mc.Passwd = c.Password
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", c.Host, c.Port)
	mc.DBName = c.Database
	mc.InterpolateParams = true
	mc.ClientFoundRows = true // UPDATE 의 affectedRows = 조건에 맞은 행 수(지금 서비스의 mysql 드라이버 기본값과 같다)
	mc.ParseTime = false
	mc.Params = map[string]string{"charset": firstNonEmpty(c.Charset, "utf8mb4")}
	d, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		return nil, err
	}
	// 프로세스 하나가 모든 요청을 받으므로 풀도 하나다(프로세스마다 따로 잡던 구조가 아니다)
	limit := c.ConnLimit
	if limit <= 0 {
		limit = 20
	}
	d.SetMaxOpenConns(limit)
	d.SetMaxIdleConns(limit)
	d.SetConnMaxLifetime(10 * time.Minute)
	return d, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Querier — *sql.DB 와 *sql.Conn 공용.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Query — 결과를 순서 보존 객체 배열로(파일 머리 주석의 값 모양).
func Query(ctx context.Context, q Querier, query string, args ...any) ([]*jsx.OrderedMap, error) {
	query, args = expand(query, args)
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
			m.Set(names[i], convert(kinds[i], b))
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// convert — 컬럼 타입별 값 변환(파일 머리 주석의 규칙).
func convert(kind string, b sql.RawBytes) any {
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
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return s
		}
		return f
	case "DATE", "DATETIME", "TIMESTAMP":
		return jsDate(k, s)
	default:
		return s
	}
}

// jsDate — DATE/DATETIME → UTC ISO(밀리초까지). 잘못된 날짜면 null.
func jsDate(kind, s string) any {
	layout := "2006-01-02 15:04:05"
	v := s
	if kind == "DATE" {
		layout = "2006-01-02"
	} else if i := strings.IndexByte(v, '.'); i > 0 {
		layout = "2006-01-02 15:04:05." + strings.Repeat("0", len(v)-i-1)
	}
	t, err := time.ParseInLocation(layout, v, KST)
	if err != nil {
		return nil
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// Placeholders — "?,?,?"(n 개).
func Placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}
