// create-admin — 첫 관리자(최고 관리자·개발자)를 만든다. 이미 있으면 비밀번호만 바꾼다.
//
//	ADMIN_PASSWORD=… go run ./cmd/create-admin -email admin@example.com -name 관리자
//	비밀번호는 환경변수 ADMIN_PASSWORD 로만 받는다(명령줄 인자는 ps·셸 기록에 남는다). 화면·로그에 찍지 않는다.
//	DB 설정은 서버와 같다(.env → .env.local → 환경변수).
//	-image 파일 을 주면 프로필 사진으로 — 업로드 폴더 {호스팅}/admin_profile_{번호}/ 에 복사하고 json.image 에 이름(이미 사진이 있으면 그대로).
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"zbd.kr/be/internal/lib/config"
	"zbd.kr/be/internal/lib/db"
	"zbd.kr/be/internal/lib/util"
)

func main() {
	email := flag.String("email", "", "로그인 이메일")
	name := flag.String("name", "관리자", "이름")
	hosting := flag.String("hosting", "ZBD", "호스팅 id")
	image := flag.String("image", "", "프로필 사진 파일(선택)")
	flag.Parse()
	pw := os.Getenv("ADMIN_PASSWORD")
	if *email == "" || len(pw) < 8 {
		log.Fatal("-email 과 ADMIN_PASSWORD(8자 이상) 가 필요하다")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
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
	// groupId 1 = 최고 관리자(isMaster) · isSuperAdmin = 최고 관리자 설정. isDeveloper(JSON 직접 고치기·저장 전 차이 보기 등 개발 도구)는 끈다 —
	// 필요하면 DB 에서 admins.json 의 isDeveloper 를 true 로(화면에서는 개발자만 켤 수 있다)
	_, err = d.ExecContext(context.Background(), `INSERT INTO admins (groupId, name, email, password, json, hostingId, updatedAt)
  VALUES (1, ?, ?, ?, '{"isDeveloper":false,"isSuperAdmin":true}', ?, UTC_TIMESTAMP())
  ON DUPLICATE KEY UPDATE password = VALUES(password), updatedAt = UTC_TIMESTAMP()`, *name, *email, hash, *hosting)
	if err != nil {
		log.Fatal(err)
	}
	if *image != "" {
		if err := setAdminImage(d, cfg.Get("CDN_UPLOAD_DIR"), *hosting, *email, *image); err != nil {
			log.Fatal("프로필 사진: ", err)
		}
	}
	fmt.Printf("관리자 준비됨: %s (%s)\n", *email, *hosting)
}

// setAdminImage — 사진이 없는 관리자에게만 넣는다(화면에서 바꾼 사진을 덮지 않게)
func setAdminImage(d *sql.DB, uploadDir, hosting, email, src string) error {
	if uploadDir == "" {
		return fmt.Errorf("CDN_UPLOAD_DIR 이 없다")
	}
	ctx := context.Background()
	var id int64
	var has bool
	if err := d.QueryRowContext(ctx, `SELECT adminId, COALESCE(JSON_EXTRACT(json, '$.image.name') IS NOT NULL, 0) FROM admins WHERE email = ? AND hostingId = ? LIMIT 1`, email, hosting).Scan(&id, &has); err != nil {
		return err
	}
	if has {
		return nil
	}
	name := filepath.Base(src)
	if err := copyFile(src, filepath.Join(uploadDir, hosting, fmt.Sprintf("admin_profile_%d", id), name)); err != nil {
		return err
	}
	_, err := d.ExecContext(ctx, `UPDATE admins SET json = JSON_SET(COALESCE(NULLIF(json, ''), '{}'), '$.image', JSON_OBJECT('imageId', ?, 'name', ?)) WHERE adminId = ?`, fmt.Sprintf("admin_profile_%d", id), name, id)
	return err
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
