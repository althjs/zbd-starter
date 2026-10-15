# 즈비드 BE (Go)
.PHONY: dev run build vet test

# air — PATH 에 없으면 GOPATH/bin 에서 찾는다(go install 기본 위치가 PATH 에 없는 경우가 많다).
AIR := $(shell command -v air 2>/dev/null || echo $(shell go env GOPATH)/bin/air)

dev: ## 저장하면 다시 빌드·재시작(air 필요: go install github.com/air-verse/air@latest)
	@test -x "$(AIR)" || { echo "air 가 없습니다 — go install github.com/air-verse/air@latest"; exit 1; }
	$(AIR)

run: ## 한 번 실행
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

vet:
	go vet ./...

test:
	go test ./...
