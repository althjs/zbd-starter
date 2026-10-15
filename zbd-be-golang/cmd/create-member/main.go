// create-member — 시험 회원 하나를 만든다(이메일 가입 회원). 이미 있으면 비밀번호만 바꾼다. 자유게시판 예시 글 2개도 이 회원 이름으로.
//
//	MEMBER_PASSWORD=… go run ./cmd/create-member -email user@example.com -name 시험회원 -nickname 테스터
//	비밀번호는 환경변수 MEMBER_PASSWORD 로만 받는다(명령줄 인자는 ps·셸 기록에 남는다). 이름·이메일은 서버와 같은 키로 암호화한다.
//	DB·암호화 키 설정은 서버와 같다(.env → .env.local → 환경변수).
//	-image 파일 을 주면 프로필 사진으로 — 업로드 폴더 {호스팅}/u_{번호}_my/profile.webp(회원 사진 규칙) + members.image(이미 사진이 있으면 그대로).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"zbd.kr/be/internal/lib/config"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/pii"
	"zbd.kr/be/internal/lib/util"
)

func main() {
	email := flag.String("email", "", "로그인 이메일")
	name := flag.String("name", "시험회원", "이름")
	nickname := flag.String("nickname", "테스터", "닉네임")
	hosting := flag.String("hosting", "ZBD", "호스팅 id")
	image := flag.String("image", "", "프로필 사진 파일(선택, webp)")
	flag.Parse()
	pw := os.Getenv("MEMBER_PASSWORD")
	if *email == "" || len(pw) < 8 {
		log.Fatal("-email 과 MEMBER_PASSWORD(8자 이상) 가 필요하다")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	keys, err := pii.New(cfg.Get("PII_ENC_KEY_V1"), cfg.Get("PII_INDEX_KEY"))
	if err != nil {
		log.Fatal("개인정보 키(PII_ENC_KEY_V1·PII_INDEX_KEY)가 없다: ", err)
	}
	d, err := db.Open(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	hash, err := util.BcryptHash(pw)
	if err != nil {
		log.Fatal(err)
	}
	encName, err := keys.Enc(*name)
	if err != nil {
		log.Fatal(err)
	}
	encEmail, err := keys.Enc(*email)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	emailHash := keys.Hash(pii.Email, *email)
	var id int64
	err = d.QueryRowContext(ctx, `SELECT userId FROM members WHERE email_hash = ? AND hostingId = ? LIMIT 1`, emailHash, *hosting).Scan(&id)
	if err == nil {
		if _, err := d.ExecContext(ctx, `UPDATE members SET password = ?, updatedAt = UTC_TIMESTAMP() WHERE userId = ?`, hash, id); err != nil {
			log.Fatal(err)
		}
	} else {
		// groupId 0 = 이메일 가입 회원. 가입 약관 동의 표시는 시험 회원이라 미리 넣는다
		res, err := d.ExecContext(ctx, `INSERT INTO members (groupId, name, email, password, mobile, json, nickname, image, description, socialId, hostingId, name_hash, email_hash, mobile_hash)
  VALUES (0, ?, ?, ?, '', '{"agree":true}', ?, '', '', '', ?, ?, ?, '')`, encName, encEmail, hash, *nickname, *hosting, keys.Hash(pii.Name, *name), emailHash)
		if err != nil {
			log.Fatal(err)
		}
		id, _ = res.LastInsertId()
	}
	// 자유게시판 예시 글 2개 — 이 회원이 쓴 것으로(이미 있으면 넣지 않는다). 그림은 화면 정적 파일
	samples := []struct{ id, title, body string }{
		{"free0001", "자유게시판 첫 글이에요", `<p>안녕하세요, 시험 회원 테스터예요. 회원은 <b>자유게시판</b>에 글을 쓰고 고칠 수 있어요.</p><p style="text-align:center"><img src="/assets/starter/sample.svg" alt="예시 그림" style="max-width:100%"></p><p>본문에 사진을 넣으려면 글쓰기 도구의 그림 버튼을 누르세요.</p>`},
		{"free0002", "글 목록·검색 시험용 글", `<p>목록·검색·쪽 나누기를 시험하려고 넣어 둔 글이에요. 확인한 뒤 지워도 됩니다.</p><ul><li>검색어 예: <code>시험</code></li><li>관리자는 「사용자 게시판 › 자유게시판」 에서 이 글을 숨기거나 지울 수 있어요</li></ul>`},
	}
	for i, p := range samples {
		b, _ := json.Marshal(map[string]string{"name": *nickname, "body": p.body})
		ju := string(b)
		if _, err := d.ExecContext(ctx, `INSERT IGNORE INTO posts (postId, adminId, userId, groupKey, groupSubKey, title, json, jsonUser, status, hostingId, createdAt, updatedAt)
  VALUES (?, 0, ?, 'FREE', '', ?, '{}', ?, 0, ?, CURRENT_TIMESTAMP() - INTERVAL ? MINUTE, CURRENT_TIMESTAMP() - INTERVAL ? MINUTE)`, p.id, id, p.title, ju, *hosting, i, i); err != nil {
			log.Fatal(err)
		}
	}
	if *image != "" {
		var cur string
		_ = d.QueryRowContext(ctx, `SELECT COALESCE(image, '') FROM members WHERE userId = ?`, id).Scan(&cur)
		if cur == "" {
			up := cfg.Get("CDN_UPLOAD_DIR")
			if up == "" {
				log.Fatal("프로필 사진: CDN_UPLOAD_DIR 이 없다")
			}
			b, err := os.ReadFile(*image)
			if err != nil {
				log.Fatal("프로필 사진: ", err)
			}
			dir := filepath.Join(up, *hosting, fmt.Sprintf("u_%d_my", id))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "profile.webp"), b, 0o644); err != nil {
				log.Fatal(err)
			}
			if _, err := d.ExecContext(ctx, `UPDATE members SET image = 'profile.webp' WHERE userId = ?`, id); err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("시험 회원 준비됨: %s (%s)\n", *email, *hosting)
}
