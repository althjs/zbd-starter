#!/bin/sh
# 기본 데이터 SQL 조립 — db/setup.sh(개발)·deploy/server/install.sh(서버)가 mariadb 에 흘려 넣는다.
#   sh build.sh <seed 폴더>   → 표준 출력
# 파일마다 넣기 전 모습을 떠서(snap) 넣고(seed) 비교한다(post) — 처음 보는 파일은 지운 메뉴 기억을 무시한다(pre.sql 설명).
set -eu
G="$(cd "$(dirname "$0")" && pwd)"
S="${1:?seed 폴더}"
names=""
cat "$G/pre.sql"
for f in "$S"/*.sql; do
	[ -f "$f" ] || continue
	n="$(basename "$f" .sql | tr -d "'\\\\")"
	names="$names${names:+,}'$n'"
	# 이 장치가 처음 도는 사이트 = 지금 있는 파일은 전부 넣어 본 것으로(예전처럼 지운 메뉴는 그대로 지운 채)
	echo "INSERT IGNORE INTO seed_files (name) SELECT '$n' FROM DUAL WHERE @seed_files_first;"
	echo "SET @seed_file := '$n';"
	cat "$G/snap.sql"
	cat "$f"
	echo
	cat "$G/post.sql"
done
# 없어진 파일(뺀 플러그인)은 잊는다 — 다시 붙이면 처음 보는 파일
[ -n "$names" ] && echo "DELETE FROM seed_files WHERE name NOT IN ($names);"
echo "DROP TEMPORARY TABLE IF EXISTS seed_had_menus, seed_had_roles;"
