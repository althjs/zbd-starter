package db

import (
	"context"
	"database/sql"
	"strings"

	"zbd.kr/be/internal/jsx"
)

// Execer — *sql.DB · *sql.Tx · *sql.Conn 공용
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// QueryOne — 첫 행(없으면 nil). 지금 서비스의 dbQueryOne.
func QueryOne(ctx context.Context, q Querier, query string, args ...any) (*jsx.OrderedMap, error) {
	rows, err := Query(ctx, q, query, args...)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// Exec — INSERT·UPDATE·DELETE. 돌려주는 값은 지금 서비스 mysql 드라이버의 OkPacket 모양이다
//
//	{fieldCount, affectedRows, insertId, serverStatus, warningCount, message, protocol41, changedRows}
//	⚠ changedRows 는 알 수 없어 affectedRows 와 같게 둔다(값이 같아도 바꾼 것으로 센다) — 화면이 이 칸을 읽는 곳은 없었다.
//	affectedRows 는 「조건에 맞은 행 수」(Open 이 ClientFoundRows 를 켠다 — 지금 서비스와 같다).
func Exec(ctx context.Context, q Execer, query string, args ...any) (*jsx.OrderedMap, error) {
	query, args = expand(query, args)
	res, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	id, _ := res.LastInsertId()
	m := jsx.NewOrderedMap()
	m.Set("fieldCount", float64(0))
	m.Set("affectedRows", float64(aff))
	m.Set("insertId", float64(id))
	m.Set("serverStatus", float64(2))
	m.Set("warningCount", float64(0))
	msg := ""
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "UPDATE") {
		msg = "(Rows matched: " + jsx.NumStr(float64(aff)) + "  Changed: " + jsx.NumStr(float64(aff)) + "  Warnings: 0"
	}
	m.Set("message", msg)
	m.Set("protocol41", true)
	m.Set("changedRows", float64(0))
	if msg != "" {
		m.Set("changedRows", float64(aff))
	}
	return m, nil
}

// InsertID · Affected — OkPacket 에서 꺼내기
func InsertID(ok *jsx.OrderedMap) int64 { return jsx.Int(ok.V("insertId")) }
func Affected(ok *jsx.OrderedMap) int64 { return jsx.Int(ok.V("affectedRows")) }

// Tx — 트랜잭션 안에서 fn. fn 이 오류를 내거나 panic 하면 되돌린다.
func Tx(ctx context.Context, d *sql.DB, fn func(tx *sql.Tx) error) (err error) {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
