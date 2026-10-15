package admincontent

// 이 패키지 안에서만 쓰는 작은 도우미 — 지금 서비스(Node)의 값 모양을 맞추기 위한 것들.

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/web"
)

// dbq — 연결 하나로 조회·실행 둘 다
type dbq interface {
	db.Querier
	db.Execer
}

// conn — 요청 하나에 연결 하나(지금 서비스의 getConnecion + release)
func conn(c *web.Ctx) (*sql.Conn, error) { return c.App.DB.Conn(c.Context()) }

// hostingID — token.hostingId ? token.hostingId : req.query.hostingId || ""
func hostingID(c *web.Ctx) string {
	if c.Token.HostingID != "" {
		return c.Token.HostingID
	}
	return jsx.StrOrEmpty(c.QV("hostingId"))
}

// scbOne — 한 건 조회 결과(dbQueryOne 은 없으면 null)
func scbOne(c *web.Ctx, m *jsx.OrderedMap) {
	if m == nil {
		c.Scb(nil)
		return
	}
	c.Scb(m)
}

// scbCode — scb(data, "0000", meta) — meta 가 비면 넣지 않는다
func scbMeta(c *web.Ctx, data any, meta *jsx.OrderedMap) {
	if meta == nil {
		c.Scb(data)
		return
	}
	c.Scb(data, meta)
}

// body — req.body. 지금 서비스는 JSON 형식으로 온 빈 본문을 {} 로, 형식 없는 빈 본문을 "" 로 받는다.
func body(c *web.Ctx) any {
	b := c.Body()
	if b != nil {
		return b
	}
	ct := strings.ToLower(c.Header("Content-Type"))
	if strings.Contains(ct, "json") {
		return jsx.NewOrderedMap()
	}
	return ""
}

// bodyMap — req.body 가 객체면 그것, 아니면 nil
func bodyMap(c *web.Ctx) *jsx.OrderedMap {
	m, _ := body(c).(*jsx.OrderedMap)
	return m
}

// rows — dbQuery 결과를 화면에 내보낼 배열로(없으면 [])
func rowsAny(rs []*jsx.OrderedMap) []any {
	out := make([]any, 0, len(rs))
	for _, r := range rs {
		out = append(out, r)
	}
	return out
}

// nodeJoin — Node 의 path.join(s): 정리하되 끝의 / 는 남긴다
func nodeJoin(parts ...string) string {
	s := strings.Join(parts, "/")
	if s == "" {
		return "."
	}
	trail := strings.HasSuffix(s, "/")
	p := filepath.Clean(s)
	if trail && p != "/" {
		p += "/"
	}
	return p
}

// uploadDir — 업로드 폴더(지금 서비스의 UPLOAD_DIR = path.join(CDN_UPLOAD_DIR)). 늘 「/」 로 끝난다(부르는 쪽이 바로 회사 폴더를 붙인다)
func uploadDir(c *web.Ctx) string {
	d := c.App.Cfg.Get("CDN_UPLOAD_DIR")
	if d == "" {
		wd, _ := os.Getwd()
		return nodeJoin(wd + "/data" + "../public/upload/")
	}
	d = nodeJoin(d)
	if !strings.HasSuffix(d, "/") {
		d += "/"
	}
	return d
}

// fileExists — utils.checkFile
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// jsNum — Number(v) || 0
func jsNum(v any) float64 {
	f := jsx.Number(v)
	if math.IsNaN(f) {
		return 0
	}
	return f
}

// parseIntJS — parseInt(s, 10) (실패면 NaN 대신 0, ok=false)
func parseIntJS(v any) (int64, bool) {
	s := strings.TrimSpace(jsx.StrOrEmpty(v))
	if v == nil {
		s = ""
	} else if _, isStr := v.(string); !isStr {
		s = strings.TrimSpace(jsx.Str(v))
	}
	end := 0
	for end < len(s) && (s[end] >= '0' && s[end] <= '9' || (end == 0 && (s[end] == '-' || s[end] == '+'))) {
		end++
	}
	n, err := strconv.ParseInt(s[:end], 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// looseEq — JS 의 a == b (숫자·문자열·null 만)
func looseEq(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	as, aStr := a.(string)
	bs, bStr := b.(string)
	if aStr && bStr {
		return as == bs
	}
	if ab, ok := a.(bool); ok {
		if bb, ok2 := b.(bool); ok2 {
			return ab == bb
		}
	}
	fa, fb := jsx.Number(a), jsx.Number(b)
	if math.IsNaN(fa) || math.IsNaN(fb) {
		return false
	}
	return fa == fb
}

// strictEq — a === b (원시값)
func strictEq(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case string:
		y, ok := b.(string)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case float64:
		y, ok := b.(float64)
		return ok && x == y
	}
	return false
}

// jsDateMs — new Date(v).getTime() (흔한 모양만: 숫자 ms · ISO · "YYYY-MM-DD"(UTC) · "YYYY-MM-DD HH:mm(:ss)"(한국 시각)). 못 읽으면 NaN.
func jsDateMs(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case float64:
		return x
	case bool:
		return math.NaN()
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return math.NaN()
		}
		utcLayouts := []string{time.RFC3339Nano, "2006-01-02T15:04:05.000Z", "2006-01-02", "2006-01", "2006"}
		for _, l := range utcLayouts {
			if t, err := time.Parse(l, s); err == nil {
				return float64(t.UnixMilli())
			}
		}
		localLayouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006/01/02 15:04:05", "2006/01/02", "2006-01-02T15:04:05.000"}
		for _, l := range localLayouts {
			if t, err := time.ParseInLocation(l, s, db.KST); err == nil {
				return float64(t.UnixMilli())
			}
		}
		return math.NaN()
	}
	return math.NaN()
}

// dbqConn — 요청이 잡은 연결
type dbqConn = *sql.Conn

// withConn — 핸들러 몸통을 연결 하나와 함께(오류면 ecb)
func withConn(c *web.Ctx, fn func(ctx context.Context, q dbqConn) error) {
	q, err := conn(c)
	if err != nil {
		c.Ecb(err)
		return
	}
	defer q.Close()
	if err := fn(c.Context(), q); err != nil && !c.Wrote() {
		c.Ecb(jsErrText(err))
	}
}

// mysqlCodes — 지금 서비스 드라이버가 오류 앞에 붙이는 코드 이름(자주 나는 것만)
var mysqlCodes = map[uint16]string{
	1046: "ER_NO_DB_ERROR", 1048: "ER_BAD_NULL_ERROR", 1054: "ER_BAD_FIELD_ERROR", 1062: "ER_DUP_ENTRY",
	1064: "ER_PARSE_ERROR", 1146: "ER_NO_SUCH_TABLE", 1264: "ER_WARN_DATA_OUT_OF_RANGE", 1292: "ER_TRUNCATED_WRONG_VALUE",
	1364: "ER_NO_DEFAULT_FOR_FIELD", 1366: "ER_TRUNCATED_WRONG_VALUE_FOR_FIELD", 1406: "ER_DATA_TOO_LONG",
	4025: "ER_CONSTRAINT_FAILED",
}

// jsErrText — 오류 응답 문구를 지금 서비스(스택 첫 줄)와 같은 모양으로: "Error: …" · "TypeError: …" · "Error: ER_…: …"
func jsErrText(err error) string {
	var e errMsg
	if errors.As(err, &e) {
		return string(e)
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		code := mysqlCodes[me.Number]
		if code == "" {
			code = "ER_" + strconv.Itoa(int(me.Number))
		}
		return "Error: " + code + ": " + me.Message
	}
	return "Error: " + err.Error()
}

// errMsg — 오류 문구(지금 서비스의 throw new Error 와 같은 문구로 돌려주려고)
type errMsg string

func (e errMsg) Error() string { return string(e) }

// parseJSONStr — JSON.parse(s) (실패하면 오류, DB 의 NULL 은 null)
func parseJSONStr(v any) (any, error) {
	s, ok := v.(string)
	if !ok {
		if v == nil {
			return nil, nil // JSON.parse(null) = null
		}
		s = jsx.Str(v)
	}
	return jsx.ParseString(s)
}

// asMap — 값이 객체면 그것(아니면 nil)
func asMap(v any) *jsx.OrderedMap {
	m, _ := v.(*jsx.OrderedMap)
	return m
}

// encodeURIComponent — JS 와 같은 규칙(A-Z a-z 0-9 - _ . ! ~ * ' ( ) 만 그대로)
func encodeURIComponent(s string) string {
	const keep = "-_.!~*'()"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || strings.IndexByte(keep, ch) >= 0 {
			b.WriteByte(ch)
			continue
		}
		b.WriteString("%" + strings.ToUpper(strconv.FormatInt(int64(ch)>>4, 16)) + strings.ToUpper(strconv.FormatInt(int64(ch)&15, 16)))
	}
	return b.String()
}

// uniqAny — _.uniq
func uniqAny(a []any) []any {
	var out []any
	for _, x := range a {
		dup := false
		for _, y := range out {
			if strictEq(x, y) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, x)
		}
	}
	return out
}
