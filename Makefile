APP_NAME := shortener
CMD_DIR := .
VERSION := 1.0.0
BUILD_DATE := $(shell date +%Y-%m-%d)
export CGO_ENABLED := 1

COMPOSE_LOCAL := docker compose -f docker-compose.local.yml

GOLANGCI_IMAGE := ${APP_NAME}-golangci

.PHONY: all build run test lint tidy clean sqlc audit mocks docker-local docker-local-rebuild lint-docker vegeta-shorten vegeta-redirect vegeta-shorten-sh vegeta-redirect-sh

sqlc:
	sqlc generate -f sqlc/sqlc.yaml
mocks:
	go generate ./internal/service/audit/...
	go generate ./internal/service/shortifier/...
build:
	go build -ldflags "-X main.BuildVersion=${VERSION} -X main.BuildDate=${BUILD_DATE}" -o $(APP_NAME) ./cmd/$(APP_NAME)
run:
	go run -ldflags "-X main.BuildVersion=${VERSION} -X main.BuildDate=${BUILD_DATE}" ./cmd/$(APP_NAME)
audit:
	./cmd/audit/audit_windows_amd64.exe
lint:
	go vet ./...

# golangci-lint in Docker (mounts repo; uses docker/golangci-lint/Dockerfile).
# Do not pass -w /src: Git Bash on Windows rewrites it to C:/Program Files/Git/src.
lint-docker:
	docker build -f docker/golangci-lint/Dockerfile -t $(GOLANGCI_IMAGE) docker/golangci-lint
	docker run --rm -v "$(CURDIR):/src" $(GOLANGCI_IMAGE)
tidy:
	go mod tidy
# Start stack; rebuild app images when Dockerfiles / context change (uses layer cache).
docker-local:
	$(COMPOSE_LOCAL) up -d --build

# Recompile Go inside Docker from scratch (no cache), then start — use when you want a clean image every time.
docker-local-rebuild:
	$(COMPOSE_LOCAL) build --no-cache ${APP_NAME}
	$(COMPOSE_LOCAL) up -d

tests:
	go test ./...

tests-coverage:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -func=coverage.out

show-runtime-pprof:
	go tool pprof -http=":9091" -seconds=30 http://localhost:8080/debug/pprof/profile

store-pprof:
	curl http://127.0.0.1:8080/debug/pprof/heap > ./profiles/base.pprof

show-base-pprof:
	go tool pprof -http=":9091" -seconds=30 ./profiles/base.pprof

store-result-pprof:
	curl http://127.0.0.1:8080/debug/pprof/heap > ./profiles/result.pprof

show-result-pprof:
	go tool pprof -http=":9092" -seconds=30 ./profiles/result.pprof

show-diff:
	pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

VEGETA_PS = powershell -NoProfile -ExecutionPolicy Bypass -File

# Windows (PowerShell) — use on machines without bash/WSL
vegeta-shorten:
	$(VEGETA_PS) scripts/vegeta/shorten.ps1

vegeta-redirect:
	$(VEGETA_PS) scripts/vegeta/redirect.ps1

# Linux / macOS / Git Bash / WSL
vegeta-shorten-sh:
	bash scripts/vegeta/shorten.sh

vegeta-redirect-sh:
	bash scripts/vegeta/redirect.sh

hey-shorten-windows:
	1..1 | ForEach-Object {
	$id = [guid]::NewGuid().ToString("n").Substring(0, 8)
	hey -n 100 -c 10 -m POST `
		-H "Content-Type: text/plain" `
		-d "http://ofdafnyylfqe.biz/page/$id" `
		http://localhost:8080/
	}
run-benchmarks:
	go test -bench="." -benchmem ./internal/handler/tests/...