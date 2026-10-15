// mint-token — 비교 시험용 로그인 토큰을 만든다(시험 DB 의 관리자·회원 정보로, 지금 서비스가 발급하는 것과 같은 모양).
//
//	go run ./cmd/mint-token -type A -id 340 -uid TESTUID -out ~/zbd-port/tokens/admin340
//	토큰은 파일로만 쓴다(600) — 화면·로그에 찍지 않는다. 운영 DB 를 가리키는 설정에서는 쓰지 말 것.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"zbd.kr/be/internal/app"
	"zbd.kr/be/internal/jsx"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
)

func main() {
	typ := flag.String("type", "A", "A 관리자 / U 회원")
	id := flag.String("id", "", "adminId / userId")
	uid := flag.String("uid", "TESTUID", "브라우저 id(회원 토큰은 요청 헤더 zbduid 와 같아야 한다)")
	out := flag.String("out", "", "토큰을 쓸 파일")
	flag.Parse()
	if *id == "" || *out == "" {
		log.Fatal("-id 와 -out 이 필요하다")
	}
	ctx := context.Background()
	a, err := app.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(a.Cfg.DB.Host, "127.0.0.1") && !strings.Contains(a.Cfg.DB.Host, "localhost") {
		log.Fatal("시험 DB(로컬)에서만 쓴다")
	}
	var email, name, group, hosting, nick string
	if *typ == "A" {
		r, err := db.QueryOne(ctx, a.DB, "SELECT email, name, groupId, hostingId, json FROM admins WHERE adminId = ?", *id)
		if err != nil || r == nil {
			log.Fatal("관리자 없음")
		}
		email, name, group, hosting = r.S("email"), r.S("name"), r.S("groupId"), r.S("hostingId")
		nick = jsx.ParseMap(r.S("json")).S("nickname")
	} else {
		r, err := db.QueryOne(ctx, a.DB, "SELECT email, name, nickname, groupId, hostingId FROM members WHERE userId = ?", *id)
		if err != nil || r == nil {
			log.Fatal("회원 없음")
		}
		email, name = a.PII.Dec(r.V("email")), a.PII.Dec(r.V("name"))
		group, hosting, nick = r.S("groupId"), r.S("hostingId"), r.S("nickname")
	}
	now := fmt.Sprint(time.Now().UnixMilli())
	fields := []string{*id, email, now, *typ, *uid, hosting, name, group, "0", "0", nick, now, "0"}
	tok := util.AES{Key: a.Cfg.AESKey, IV: a.Cfg.AESIV}.Encrypt(strings.Join(fields, "|"))
	if err := os.WriteFile(*out, []byte(tok), 0o600); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s 토큰 → %s\n", *typ, *id, *out)
}
