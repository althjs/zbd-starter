package app

import (
	"os"
	"strings"

	"zbd.kr/be/internal/jsx"
)

// TalkBotIDs — 친구 채팅의 Claude 회원 번호. 설정 TALK_BOT_USER_IDS(없으면 "11", 빈 값이면 끔).
// 이 회원에게 온 친구 신청은 서버가 바로 수락하고(미니홈피 친구), 채팅은 다른 회원처럼 친구끼리만(2026-10-03 대표).
func (a *App) TalkBotIDs() []int64 {
	v, ok := os.LookupEnv("TALK_BOT_USER_IDS")
	if !ok {
		v = a.Cfg.GetOr("TALK_BOT_USER_IDS", "11")
	}
	var out []int64
	for _, s := range strings.Split(v, ",") {
		if n := jsx.Number(strings.TrimSpace(s)); n != 0 && n == n {
			out = append(out, int64(n))
		}
	}
	return out
}

// IsTalkBot — 그 회원이 친구 채팅 Claude 인지
func (a *App) IsTalkBot(id int64) bool {
	for _, x := range a.TalkBotIDs() {
		if x == id {
			return true
		}
	}
	return false
}
