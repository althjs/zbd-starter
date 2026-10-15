package content

// 이 패키지 안에서만 쓰는 작은 도우미.

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/web"
)

// ecbCode — res.send(ecb(msg, code)) — 실패 응답에 code 를 따로 줄 때
func ecbCode(c *web.Ctx, msg any, code string) {
	m := web.Ecb(msg)
	m.Set("code", code)
	c.Send(m)
}

// withConn — 연결 하나를 잡아 fn 을 돌린다(오류는 ecb 로)
func withConn(c *web.Ctx, fn func(ctx context.Context, conn *sql.Conn)) {
	ctx := c.Context()
	conn, err := c.App.DB.Conn(ctx)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer conn.Close()
	fn(ctx, conn)
}

// tokenHosting — token.hostingId ? token.hostingId : req.query.hostingId || ""
func tokenHosting(c *web.Ctx) string {
	if c.Token != nil && c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return jsx.StrOrEmpty(c.QV("hostingId"))
}

// tokenKey — token.key (없으면 "")
func tokenKey(c *web.Ctx) string {
	if c.Token == nil {
		return ""
	}
	return c.Token.Key
}

// rowsAny — []*OrderedMap → []any (응답 배열)
func rowsAny(rows []*jsx.OrderedMap) []any {
	out := make([]any, len(rows))
	for i, r := range rows {
		out[i] = r
	}
	return out
}

// mysql 오류 번호 → 지금 서비스 드라이버의 코드 이름(자주 나오는 것만)
var mysqlCodes = map[uint16]string{
	1048: "ER_BAD_NULL_ERROR", 1054: "ER_BAD_FIELD_ERROR", 1062: "ER_DUP_ENTRY", 1064: "ER_PARSE_ERROR",
	1146: "ER_NO_SUCH_TABLE", 1264: "ER_WARN_DATA_OUT_OF_RANGE", 1364: "ER_NO_DEFAULT_FOR_FIELD",
	1366: "ER_TRUNCATED_WRONG_VALUE_FOR_FIELD", 1406: "ER_DATA_TOO_LONG", 1292: "ER_TRUNCATED_WRONG_VALUE",
	1452: "ER_NO_REFERENCED_ROW_2", 3140: "ER_INVALID_JSON_TEXT",
}

// errText — ecb(e.stack) 의 첫 줄과 같은 문구. DB 오류는 "Error: ER_NO_SUCH_TABLE: Table … doesn't exist".
func errText(err error) string {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		code := mysqlCodes[me.Number]
		if code == "" {
			code = "ER_" + strconv.Itoa(int(me.Number))
		}
		return "Error: " + code + ": " + me.Message
	}
	return err.Error()
}

// ecbErr — res.send(ecb(e.stack || e.message || e))
func ecbErr(c *web.Ctx, err error) { c.Ecb(errText(err)) }

// nowMs — Date.now()
func nowMs() float64 { return float64(time.Now().UnixMilli()) }
