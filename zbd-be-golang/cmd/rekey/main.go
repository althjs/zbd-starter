// rekey — 회원 개인정보 키를 바꾼다(지금 설정의 키 → 새 키). 배포가 서버 키를 개발 컴퓨터 키와 맞출 때 쓴다(deploy/server/install.sh).
//
//	NEW_PII_ENC_KEY_V1=… NEW_PII_INDEX_KEY=… ./rekey [-dry]
//	지금 키 = 설정(.env → .env.local → 환경변수)의 PII_ENC_KEY_V1·PII_INDEX_KEY. 새 키는 환경변수로만(명령줄은 ps 에 남는다).
//
// 하는 일(여러 번 돌려도 같다 — 이미 새 키로 풀리는 값은 그대로 둔다):
//   - members name·email·mobile: 옛 키로 풀리는 값만 새 키로 다시 암호화 + name_hash·email_hash·mobile_hash 다시 계산
//   - admins.otp: 옛 키로 풀리는 값만 새 키로
//   - login_devices.email(찾기 값): 회원·관리자 이메일로 새 찾기 값 — 같은 줄이 이미 있으면 옛 줄은 지운다(기억한 기기일 뿐)
//
// 두 키 어느 쪽으로도 안 풀리는 값은 건드리지 않고 개수만 알린다. 같은 이메일이 두 키로 따로 가입해 있으면(찾기 값 겹침) 그 회원은 남기고 알린다.
// 끝 줄: 「rekey ok …」 — 치명적인 오류(DB 접속·키 모양)면 종료 코드 1.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"zbd.kr/be/internal/lib/config"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/pii"
)

func main() {
	dry := flag.Bool("dry", false, "바꾸지 않고 개수만")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	oldK, err := pii.New(cfg.Get("PII_ENC_KEY_V1"), cfg.Get("PII_INDEX_KEY"))
	if err != nil {
		log.Fatal("지금 키: ", err)
	}
	newK, err := pii.New(os.Getenv("NEW_PII_ENC_KEY_V1"), os.Getenv("NEW_PII_INDEX_KEY"))
	if err != nil {
		log.Fatal("새 키(NEW_…): ", err)
	}
	if cfg.Get("PII_ENC_KEY_V1") == os.Getenv("NEW_PII_ENC_KEY_V1") && cfg.Get("PII_INDEX_KEY") == os.Getenv("NEW_PII_INDEX_KEY") {
		fmt.Println("rekey ok — 키가 같다(할 일 없음)")
		return
	}
	d, err := db.Open(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	ctx := context.Background()

	// 값 하나: 새 키로 풀리면 그대로, 옛 키로 풀리면 새 키로. 평문·결과·바뀌었는지·못 풂
	conv := func(v string) (plain, out string, changed, bad bool) {
		if v == "" {
			return "", "", false, false
		}
		if p := newK.Dec(v); p != "" {
			return p, v, false, false
		}
		if p := oldK.Dec(v); p != "" {
			return p, newK.MustEnc(p), true, false
		}
		return "", v, false, true
	}

	var mOK, mSame, mBad, mDup, oOK, oBad, dOK, dDel int
	var dups []int64

	rows, err := d.QueryContext(ctx, `SELECT userId, name, email, mobile, name_hash, email_hash, mobile_hash FROM members`)
	if err != nil {
		log.Fatal(err)
	}
	type mem struct {
		id                  int64
		name, email, mobile string
		nameH, emailH, mobH string
	}
	var ms []mem
	for rows.Next() {
		var m mem
		var n, e, mo, nh, eh, mh sql.NullString
		if err := rows.Scan(&m.id, &n, &e, &mo, &nh, &eh, &mh); err != nil {
			log.Fatal(err)
		}
		m.name, m.email, m.mobile, m.nameH, m.emailH, m.mobH = n.String, e.String, mo.String, nh.String, eh.String, mh.String
		ms = append(ms, m)
	}
	rows.Close()
	emailOf := map[int64]string{} // 회원 이메일(평문) — 기기 찾기 값 다시 계산용
	for _, m := range ms {
		pn, on, cn, bn := conv(m.name)
		pe, oe, ce, be := conv(m.email)
		pm, om, cm, bm := conv(m.mobile)
		if pe != "" {
			emailOf[m.id] = pe
		}
		if bn || be || bm {
			mBad++
		}
		// 찾기 값은 풀린 칸만 새로(못 푼 칸은 옛 값 그대로)
		nh, eh, mh := m.nameH, m.emailH, m.mobH
		if pn != "" {
			nh = newK.Hash(pii.Name, pn)
		}
		if pe != "" {
			eh = newK.Hash(pii.Email, pe)
		}
		if pm != "" {
			mh = newK.Hash(pii.Mobile, pm)
		}
		if !cn && !ce && !cm && nh == m.nameH && eh == m.emailH && mh == m.mobH {
			mSame++
			continue
		}
		if *dry {
			mOK++
			continue
		}
		if _, err := d.ExecContext(ctx, `UPDATE members SET name = ?, email = ?, mobile = ?, name_hash = ?, email_hash = ?, mobile_hash = ? WHERE userId = ?`,
			on, oe, om, nh, eh, mh, m.id); err != nil {
			mDup++ // 같은 이메일·소셜이 이미 새 키로 있다(uq_member_email_social)
			dups = append(dups, m.id)
			continue
		}
		mOK++
	}

	// 관리자 2단계 인증 비밀값
	arows, err := d.QueryContext(ctx, `SELECT adminId, otp, email FROM admins`)
	if err != nil {
		log.Fatal(err)
	}
	adminEmail := map[int64]string{}
	type otpRow struct {
		id  int64
		otp string
	}
	var otps []otpRow
	for arows.Next() {
		var id int64
		var otp, em sql.NullString
		if err := arows.Scan(&id, &otp, &em); err != nil {
			log.Fatal(err)
		}
		adminEmail[id] = em.String // 관리자 이메일은 평문
		otps = append(otps, otpRow{id, otp.String})
	}
	arows.Close()
	for _, a := range otps {
		_, out, changed, bad := conv(a.otp)
		if bad {
			oBad++
		}
		if !changed {
			continue
		}
		oOK++
		if !*dry {
			if _, err := d.ExecContext(ctx, `UPDATE admins SET otp = ? WHERE adminId = ?`, out, a.id); err != nil {
				log.Fatal(err)
			}
		}
	}

	// 기억한 기기 — 찾기 값을 새 키로(같은 줄이 이미 있으면 옛 줄은 지운다)
	drows, err := d.QueryContext(ctx, `SELECT memberType, email, browserUid, hostingId, userId FROM login_devices`)
	if err != nil {
		log.Fatal(err)
	}
	type dev struct {
		mt, email, uid, host string
		userID               int64
	}
	var devs []dev
	for drows.Next() {
		var x dev
		if err := drows.Scan(&x.mt, &x.email, &x.uid, &x.host, &x.userID); err != nil {
			log.Fatal(err)
		}
		devs = append(devs, x)
	}
	drows.Close()
	for _, x := range devs {
		plain := ""
		if x.mt == "A" {
			plain = adminEmail[x.userID]
		} else {
			plain = emailOf[x.userID]
		}
		if plain == "" {
			continue
		}
		h := newK.Hash(pii.Email, plain)
		if h == x.email {
			continue
		}
		if *dry {
			dOK++
			continue
		}
		key := []any{x.mt, x.email, x.uid, x.host, x.userID}
		ok, err := d.ExecContext(ctx, `UPDATE IGNORE login_devices SET email = ? WHERE memberType = ? AND email = ? AND browserUid = ? AND hostingId = ? AND userId = ?`, append([]any{h}, key...)...)
		if err != nil {
			log.Fatal(err)
		}
		if n, _ := ok.RowsAffected(); n == 1 {
			dOK++
			continue
		}
		if _, err := d.ExecContext(ctx, `DELETE FROM login_devices WHERE memberType = ? AND email = ? AND browserUid = ? AND hostingId = ? AND userId = ?`, key...); err != nil {
			log.Fatal(err)
		}
		dDel++
	}

	mode := ""
	if *dry {
		mode = "(미리 보기) "
	}
	fmt.Printf("rekey ok %s회원 %d명 바꿈 · %d명 그대로 · 못 푼 회원 %d · 겹쳐서 못 바꾼 회원 %d · 관리자 2단계 %d(못 풂 %d) · 기기 %d(겹쳐 지움 %d)\n",
		mode, mOK, mSame, mBad, mDup, oOK, oBad, dOK, dDel)
	if len(dups) > 0 {
		fmt.Printf("  ⚠ 같은 이메일이 이미 있어 못 바꾼 회원 번호: %v — 관리자 화면에서 한쪽을 정리하세요\n", dups)
	}
}
