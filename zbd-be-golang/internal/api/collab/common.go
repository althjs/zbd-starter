package collab

// 이 패키지 안에서 같이 쓰는 작은 도우미 — JS 의 응답 모양·날짜 계산을 맞춘다.

// mysqlCodes — 지금 서비스 드라이버가 쓰는 오류 이름(자주 나는 것만)
var mysqlCodes = map[uint16]string{
	1054: "ER_BAD_FIELD_ERROR", 1146: "ER_NO_SUCH_TABLE", 1064: "ER_PARSE_ERROR", 1062: "ER_DUP_ENTRY",
	1048: "ER_BAD_NULL_ERROR", 1366: "ER_TRUNCATED_WRONG_VALUE_FOR_FIELD", 1406: "ER_DATA_TOO_LONG",
	1292: "ER_TRUNCATED_WRONG_VALUE", 1452: "ER_NO_REFERENCED_ROW_2", 1451: "ER_ROW_IS_REFERENCED_2",
}
