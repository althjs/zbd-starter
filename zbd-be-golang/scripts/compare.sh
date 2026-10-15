#!/usr/bin/env bash
# 지금 서비스(Node BE)와 Go 서버에 같은 요청을 보내 응답을 **글자 단위로** 비교한다 — API 를 옮길 때 확인용.
#
#   TOKEN_FILE=/path/to/token scripts/compare.sh /user/typing/stats
#   METHOD=POST BODY_FILE=body.json TOKEN_FILE=… scripts/compare.sh /user/…
#
#   NODE_BASE (기본 http://localhost:4501/api)   GO_BASE (기본 http://localhost:5001/api/v1)
#   토큰은 파일로만 받는다 — 명령줄·로그에 남기지 않는다. 두 서버에 같은 x-request-id 를 보낸다(응답 meta.requestId 가 같아야 비교된다).
set -euo pipefail
path="${1:?API 경로 (예: /user/typing/stats)}"
NODE_BASE="${NODE_BASE:-http://localhost:4501/api}"
GO_BASE="${GO_BASE:-http://localhost:5001/api/v1}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
args=(-s -X "${METHOD:-GET}" -H 'Content-Type: application/json' -H 'x-request-id: compare')
if [[ -n "${TOKEN_FILE:-}" ]]; then
  args+=(-H "Authorization: Bearer $(tr -d '\n' < "$TOKEN_FILE")")
fi
[[ -n "${BODY_FILE:-}" ]] && args+=(--data-binary "@$BODY_FILE")
nc=$(curl "${args[@]}" -o "$tmp/node" -w '%{http_code} %{time_total}' "$NODE_BASE$path")
gc=$(curl "${args[@]}" -o "$tmp/go" -w '%{http_code} %{time_total}' "$GO_BASE$path")
echo "node: $nc s · $(wc -c < "$tmp/node") bytes"
echo "go  : $gc s · $(wc -c < "$tmp/go") bytes"
if cmp -s "$tmp/node" "$tmp/go"; then
  echo "✅ 동일"
else
  echo "❌ 다름 — 첫 차이:"
  cmp "$tmp/node" "$tmp/go" || true
  off=$(cmp "$tmp/node" "$tmp/go" | awk '{print $5}' | tr -d ,)
  s=$(( off > 80 ? off - 80 : 0 ))
  echo "node …$(tail -c +$s "$tmp/node" | head -c 200)"
  echo "go   …$(tail -c +$s "$tmp/go" | head -c 200)"
  exit 1
fi
